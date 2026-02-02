package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadIncludeFilter_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	includePath := filepath.Join(tempDir, "empty.txt")
	if err := os.WriteFile(includePath, []byte("# comment only\n\n"), 0600); err != nil {
		t.Fatalf("write include file: %v", err)
	}

	_, err := loadIncludeFilter(includePath)
	if err == nil || !strings.Contains(err.Error(), "no patterns") {
		t.Fatalf("expected empty patterns error, got: %v", err)
	}
}

func TestLoadIncludeFilter_AbsolutePattern(t *testing.T) {
	tempDir := t.TempDir()
	includePath := filepath.Join(tempDir, "abs.txt")
	if err := os.WriteFile(includePath, []byte("/abs/path\n"), 0600); err != nil {
		t.Fatalf("write include file: %v", err)
	}

	_, err := loadIncludeFilter(includePath)
	if err == nil || !strings.Contains(err.Error(), "relative to the sync source root") {
		t.Fatalf("expected absolute pattern error, got: %v", err)
	}
}

func TestLoadIncludeFilter_Matching(t *testing.T) {
	tempDir := t.TempDir()
	includePath := filepath.Join(tempDir, "patterns.txt")
	content := strings.Join([]string{
		"# comment",
		"logs/",
		"data/*.json",
		"nested/**/file.txt",
		"./root/*.txt",
		"",
	}, "\n")
	if err := os.WriteFile(includePath, []byte(content), 0600); err != nil {
		t.Fatalf("write include file: %v", err)
	}

	filter, err := loadIncludeFilter(includePath)
	if err != nil {
		t.Fatalf("load include filter: %v", err)
	}

	testCases := []struct {
		path string
		want bool
	}{
		{filepath.Join("logs", "app.log"), true},
		{filepath.Join("data", "test.json"), true},
		{filepath.Join("data", "test.csv"), false},
		{filepath.Join("nested", "a", "file.txt"), true},
		{filepath.Join("root", "note.txt"), true},
		{filepath.Join("root", "note.md"), false},
	}

	for _, tc := range testCases {
		if got := filter(tc.path); got != tc.want {
			t.Errorf("filter(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}
