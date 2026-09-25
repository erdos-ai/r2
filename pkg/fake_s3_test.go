package pkg

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// fakeS3 is a minimal in-memory, path-style S3 endpoint for a single bucket. It implements only
// the calls the bucket helpers make (ListObjectsV2, PutObject, GetObject), so tests can exercise
// them through the real AWS SDK without live R2 calls.
type fakeS3 struct {
	bucket  string
	mu      sync.Mutex
	objects map[string][]byte
}

// newFakeS3 starts a fake S3 server for bucket and returns it with an R2Bucket pointed at it.
func newFakeS3(t *testing.T, bucket string) (*fakeS3, R2Bucket) {
	t.Helper()

	fake := &fakeS3{bucket: bucket, objects: map[string][]byte{}}
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

	switch {
	case r.Method == http.MethodGet && key == "" && r.URL.Query().Get("list-type") == "2":
		prefix := r.URL.Query().Get("prefix")
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
		w.Header().Set("Content-Type", "application/xml")
		_ = xml.NewEncoder(w).Encode(result)
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

func md5Hex(body []byte) string {
	sum := md5.Sum(body)
	return hex.EncodeToString(sum[:])
}
