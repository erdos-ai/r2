package pkg

import (
	"bytes"
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
