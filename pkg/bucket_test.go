package pkg

import (
	"bytes"
	"net/url"
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
