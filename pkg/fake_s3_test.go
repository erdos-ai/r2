package pkg

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// fakeS3 is a minimal in-memory, path-style S3 endpoint for a single bucket. It implements only
// the calls the bucket helpers make (ListObjectsV2, PutObject, GetObject, and multipart uploads),
// so tests can exercise them through the real AWS SDK without live R2 calls.
type fakeS3 struct {
	bucket  string
	mu      sync.Mutex
	objects map[string][]byte

	// Multipart uploads in progress (upload ID -> part number -> body), and the number of parts
	// each completed multipart object was assembled from.
	uploads      map[string]map[int][]byte
	partCounts   map[string]int
	nextUploadID int
}

// newFakeS3 starts a fake S3 server for bucket and returns it with an R2Bucket pointed at it.
func newFakeS3(t *testing.T, bucket string) (*fakeS3, R2Bucket) {
	t.Helper()

	fake := &fakeS3{
		bucket:     bucket,
		objects:    map[string][]byte{},
		uploads:    map[string]map[int][]byte{},
		partCounts: map[string]int{},
	}
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)

	client := R2Client{*s3.New(s3.Options{
		BaseEndpoint:               aws.String(server.URL),
		Region:                     "auto",
		UsePathStyle:               true,
		Credentials:                credentials.NewStaticCredentialsProvider("test-access-key", "test-secret-key", ""),
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	})}
	return fake, client.Bucket(bucket)
}

func (f *fakeS3) put(key string, body []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.objects[key] = body
}

func (f *fakeS3) keys() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	keys := make([]string, 0, len(f.objects))
	for key := range f.objects {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (f *fakeS3) object(key string) ([]byte, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	body, ok := f.objects[key]
	return body, ok
}

// multipartParts returns how many parts key was uploaded in, or 0 if it wasn't a multipart upload.
func (f *fakeS3) multipartParts(key string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.partCounts[key]
}

func (f *fakeS3) createUpload(key string) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextUploadID++
	id := fmt.Sprintf("upload-%d", f.nextUploadID)
	f.uploads[id] = map[int][]byte{}
	return id
}

func (f *fakeS3) putPart(uploadID string, partNumber int, body []byte) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	parts, ok := f.uploads[uploadID]
	if ok {
		parts[partNumber] = body
	}
	return ok
}

// completeUpload assembles the parts of uploadID in part-number order into key.
func (f *fakeS3) completeUpload(key, uploadID string) ([]byte, int, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	parts, ok := f.uploads[uploadID]
	if !ok {
		return nil, 0, false
	}
	numbers := make([]int, 0, len(parts))
	for n := range parts {
		numbers = append(numbers, n)
	}
	sort.Ints(numbers)
	var body []byte
	for _, n := range numbers {
		body = append(body, parts[n]...)
	}
	delete(f.uploads, uploadID)
	f.objects[key] = body
	f.partCounts[key] = len(numbers)
	return body, len(numbers), true
}

func (f *fakeS3) abortUpload(uploadID string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.uploads, uploadID)
}

type fakeListResult struct {
	XMLName     xml.Name        `xml:"ListBucketResult"`
	Xmlns       string          `xml:"xmlns,attr"`
	Name        string          `xml:"Name"`
	Prefix      string          `xml:"Prefix"`
	KeyCount    int             `xml:"KeyCount"`
	MaxKeys     int             `xml:"MaxKeys"`
	IsTruncated bool            `xml:"IsTruncated"`
	Contents    []fakeListEntry `xml:"Contents"`
}

type fakeListEntry struct {
	Key          string `xml:"Key"`
	ETag         string `xml:"ETag"`
	Size         int    `xml:"Size"`
	LastModified string `xml:"LastModified"`
}

func (f *fakeS3) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bucket, key, _ := strings.Cut(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if bucket != f.bucket {
		http.Error(w, "NoSuchBucket", http.StatusNotFound)
		return
	}

	query := r.URL.Query()
	switch {
	case r.Method == http.MethodPost && key != "" && query.Has("uploads"):
		writeXML(w, struct {
			XMLName  xml.Name `xml:"InitiateMultipartUploadResult"`
			Bucket   string
			Key      string
			UploadId string
		}{Bucket: f.bucket, Key: key, UploadId: f.createUpload(key)})
	case r.Method == http.MethodPut && key != "" && query.Has("uploadId"):
		partNumber, err := strconv.Atoi(query.Get("partNumber"))
		if err != nil {
			http.Error(w, "InvalidArgument", http.StatusBadRequest)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if !f.putPart(query.Get("uploadId"), partNumber, body) {
			http.Error(w, "NoSuchUpload", http.StatusNotFound)
			return
		}
		w.Header().Set("ETag", `"`+md5Hex(body)+`"`)
	case r.Method == http.MethodPost && key != "" && query.Has("uploadId"):
		_, _ = io.Copy(io.Discard, r.Body)
		body, parts, ok := f.completeUpload(key, query.Get("uploadId"))
		if !ok {
			http.Error(w, "NoSuchUpload", http.StatusNotFound)
			return
		}
		writeXML(w, struct {
			XMLName xml.Name `xml:"CompleteMultipartUploadResult"`
			Bucket  string
			Key     string
			ETag    string
		}{Bucket: f.bucket, Key: key, ETag: fmt.Sprintf(`"%s-%d"`, md5Hex(body), parts)})
	case r.Method == http.MethodDelete && key != "" && query.Has("uploadId"):
		f.abortUpload(query.Get("uploadId"))
		w.WriteHeader(http.StatusNoContent)
	case r.Method == http.MethodGet && key == "" && query.Get("list-type") == "2":
		prefix := query.Get("prefix")
		result := fakeListResult{
			Xmlns:   "http://s3.amazonaws.com/doc/2006-03-01/",
			Name:    f.bucket,
			Prefix:  prefix,
			MaxKeys: 1000,
		}
		for _, k := range f.keys() {
			if !strings.HasPrefix(k, prefix) {
				continue
			}
			body, _ := f.object(k)
			result.Contents = append(result.Contents, fakeListEntry{
				Key:          k,
				ETag:         `"` + md5Hex(body) + `"`,
				Size:         len(body),
				LastModified: "2026-01-01T00:00:00.000Z",
			})
		}
		result.KeyCount = len(result.Contents)
		writeXML(w, result)
	case r.Method == http.MethodPut && key != "":
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		f.put(key, body)
		w.Header().Set("ETag", `"`+md5Hex(body)+`"`)
	case r.Method == http.MethodGet && key != "":
		body, ok := f.object(key)
		if !ok {
			http.Error(w, "NoSuchKey", http.StatusNotFound)
			return
		}
		w.Header().Set("ETag", `"`+md5Hex(body)+`"`)
		_, _ = w.Write(body)
	default:
		http.Error(w, "fake S3 does not support "+r.Method+" "+r.URL.String(), http.StatusNotImplemented)
	}
}

func writeXML(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/xml")
	_ = xml.NewEncoder(w).Encode(v)
}

func md5Hex(body []byte) string {
	sum := md5.Sum(body)
	return hex.EncodeToString(sum[:])
}
