package pkg

import (
	"bytes"
	"io"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestPutStreamRejectsSubMinimumMultipartPartSize(t *testing.T) {
	var bucket R2Bucket

	err := bucket.PutStream(bytes.NewReader([]byte("ab")), "object", 1, 1)
	if err == nil {
		t.Fatal("expected invalid part size error")
	}

	want := "part size must be at least 5242880 bytes"
	if err.Error() != want {
		t.Fatalf("expected %q, got %q", want, err.Error())
	}
}

func TestPutStreamUploadsFromPipe(t *testing.T) {
	fake, bucket := newFakeS3(t, "test-bucket")

	// A pipe is what `r2 pipe` receives on stdin. *os.File implements io.Seeker, but seeking a
	// pipe fails, which previously aborted the upload with "illegal seek".
	reader := pipeWith(t, []byte("hello from stdin\n"))

	if err := bucket.PutStream(reader, "streamed.txt", minMultipartPartSize, 5); err != nil {
		t.Fatalf("PutStream from pipe: %v", err)
	}

	body, ok := fake.object("streamed.txt")
	if !ok {
		t.Fatalf("object was not uploaded; bucket has %v", fake.keys())
	}
	if string(body) != "hello from stdin\n" {
		t.Fatalf("uploaded body = %q, want %q", body, "hello from stdin\n")
	}
}

func TestPutStreamUploadsLargePipeInParts(t *testing.T) {
	fake, bucket := newFakeS3(t, "test-bucket")

	// 12 MiB of non-repeating bytes, so misordered or dropped parts change the result. With 5 MiB
	// parts this streams as three parts of unknown total size, like a piped backup.
	data := make([]byte, 12*1024*1024)
	for i := range data {
		data[i] = byte(i % 251)
	}
	reader := pipeWith(t, data)

	if err := bucket.PutStream(reader, "backup.tar", minMultipartPartSize, 5); err != nil {
		t.Fatalf("PutStream from pipe: %v", err)
	}

	body, ok := fake.object("backup.tar")
	if !ok {
		t.Fatalf("object was not uploaded; bucket has %v", fake.keys())
	}
	if !bytes.Equal(body, data) {
		t.Fatalf("uploaded %d bytes that differ from the %d bytes piped in", len(body), len(data))
	}
	if parts := fake.multipartParts("backup.tar"); parts != 3 {
		t.Fatalf("uploaded in %d parts, want 3", parts)
	}
}

// pipeWith returns the read end of an os.Pipe that yields data and then EOF, like piped stdin.
func pipeWith(t *testing.T, data []byte) *os.File {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	t.Cleanup(func() { reader.Close() })
	go func() {
		writer.Write(data)
		writer.Close()
	}()
	return reader
}

func TestStreamBodyHidesSeekerOnlyWhenSeekFails(t *testing.T) {
	pipeReader, pipeWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer pipeReader.Close()
	defer pipeWriter.Close()

	file, err := os.CreateTemp(t.TempDir(), "body")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer file.Close()

	tests := []struct {
		name         string
		reader       io.Reader
		wantSeekable bool
	}{
		{"pipe", pipeReader, false},
		{"regular file", file, true},
		{"in-memory reader", bytes.NewReader([]byte("abc")), true},
		{"non-seeking reader", struct{ io.Reader }{strings.NewReader("abc")}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, seekable := streamBody(tc.reader).(io.Seeker)
			if seekable != tc.wantSeekable {
				t.Fatalf("streamBody seekable = %v, want %v", seekable, tc.wantSeekable)
			}
		})
	}
}

func TestPresignedGetURLDoesNotRequireChecksumHeader(t *testing.T) {
	client := testPresignClient()
	assertNoChecksumSignedHeaders(t, client.GetURL(R2URI{Bucket: "test-bucket", Path: "object.txt"}))
}

func TestPresignedPutURLDoesNotRequireChecksumHeader(t *testing.T) {
	client := testPresignClient()
	assertNoChecksumSignedHeaders(t, client.PutURL(R2URI{Bucket: "test-bucket", Path: "object.txt"}))
}

func testPresignClient() R2PresignClient {
	return PresignClient(Config{
		AccountID:       "test-account",
		AccessKeyID:     "test-access-key",
		SecretAccessKey: "test-secret-key",
	})
}

func assertNoChecksumSignedHeaders(t *testing.T, rawURL string) {
	t.Helper()

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse presigned URL: %v", err)
	}

	var signedHeaders string
	for key, values := range parsedURL.Query() {
		if strings.EqualFold(key, "X-Amz-SignedHeaders") && len(values) > 0 {
			signedHeaders = values[0]
			break
		}
	}
	if signedHeaders == "" {
		t.Fatal("presigned URL is missing X-Amz-SignedHeaders")
	}

	for _, header := range strings.Split(strings.ToLower(signedHeaders), ";") {
		if strings.HasPrefix(header, "x-amz-checksum-") {
			t.Fatalf("presigned URL unexpectedly requires checksum header %q: %q", header, signedHeaders)
		}
	}
}
