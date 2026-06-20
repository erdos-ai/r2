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

func TestLoadExcludeFilter_Matching(t *testing.T) {
	tempDir := t.TempDir()
	excludePath := filepath.Join(tempDir, "exclude.txt")
	content := strings.Join([]string{
		"# comment",
		"logs/",
		"*.tmp",
		"secret/**/*.key",
		"",
	}, "\n")
	if err := os.WriteFile(excludePath, []byte(content), 0600); err != nil {
		t.Fatalf("write exclude file: %v", err)
	}

	filter, err := loadExcludeFilter(excludePath, nil)
	if err != nil {
		t.Fatalf("load exclude filter: %v", err)
	}
	if filter == nil {
		t.Fatal("expected non-nil exclude filter")
	}

	// The exclude matcher reports true for paths that match an exclude pattern.
	// (combineFilters inverts this into a skip decision.) Note that a bare "*"
	// does not cross directory boundaries, so "*.tmp" only matches at the root.
	testCases := []struct {
		path string
		want bool
	}{
		{filepath.Join("logs", "app.log"), true},
		{"out.tmp", true},
		{filepath.Join("secret", "a", "b", "k.key"), true},
		{filepath.Join("src", "main.go"), false},
	}
	for _, tc := range testCases {
		if got := filter(tc.path); got != tc.want {
			t.Errorf("filter(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestLoadExcludeFilter_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	excludePath := filepath.Join(tempDir, "empty.txt")
	if err := os.WriteFile(excludePath, []byte("# comment only\n\n"), 0600); err != nil {
		t.Fatalf("write exclude file: %v", err)
	}

	// Unlike include, an empty exclude file is a no-op rather than an error: it
	// means "exclude nothing", which is already the default behavior.
	filter, err := loadExcludeFilter(excludePath, nil)
	if err != nil {
		t.Fatalf("expected no error for empty exclude file, got: %v", err)
	}
	if filter != nil {
		t.Fatal("expected nil exclude filter for empty exclude file")
	}
}

func TestLoadExcludeFilter_AbsolutePattern(t *testing.T) {
	tempDir := t.TempDir()
	excludePath := filepath.Join(tempDir, "abs.txt")
	if err := os.WriteFile(excludePath, []byte("/abs/path\n"), 0600); err != nil {
		t.Fatalf("write exclude file: %v", err)
	}

	_, err := loadExcludeFilter(excludePath, nil)
	if err == nil || !strings.Contains(err.Error(), "relative to the sync source root") {
		t.Fatalf("expected absolute pattern error, got: %v", err)
	}
}

func TestLoadExcludeFilter_Inline(t *testing.T) {
	filter, err := loadExcludeFilter("", []string{"*.tmp", "cache/"})
	if err != nil {
		t.Fatalf("load exclude filter: %v", err)
	}
	if filter == nil {
		t.Fatal("expected non-nil exclude filter")
	}

	testCases := []struct {
		path string
		want bool
	}{
		{"build.tmp", true},
		{filepath.Join("cache", "x", "y.bin"), true},
		{filepath.Join("src", "main.go"), false},
	}
	for _, tc := range testCases {
		if got := filter(tc.path); got != tc.want {
			t.Errorf("filter(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}

	// Empty/whitespace inline patterns are skipped; if nothing remains, the
	// result is a no-op (nil) filter rather than an error.
	empty, err := loadExcludeFilter("", []string{"", "   "})
	if err != nil {
		t.Fatalf("expected no error for empty inline patterns, got: %v", err)
	}
	if empty != nil {
		t.Fatal("expected nil exclude filter when all inline patterns are empty")
	}
}

func TestLoadExcludeFilter_BracePattern(t *testing.T) {
	// Brace alternation must survive flag parsing intact, which is why --exclude
	// is registered as a StringArray: a StringSlice would split "{foo,bar}.txt"
	// on the comma. Feeding the pattern directly confirms the matcher honors it.
	filter, err := loadExcludeFilter("", []string{"{foo,bar}.txt"})
	if err != nil {
		t.Fatalf("load exclude filter: %v", err)
	}
	if filter == nil {
		t.Fatal("expected non-nil exclude filter")
	}

	testCases := []struct {
		path string
		want bool
	}{
		{"foo.txt", true},
		{"bar.txt", true},
		{"baz.txt", false},
	}
	for _, tc := range testCases {
		if got := filter(tc.path); got != tc.want {
			t.Errorf("filter(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestCombineFilters_Interaction(t *testing.T) {
	include := compileMatcher([]string{"data/**"})
	exclude := compileMatcher([]string{"data/secret/**", "**/*.tmp"})

	t.Run("include and exclude", func(t *testing.T) {
		filter := combineFilters(include, exclude)
		if filter == nil {
			t.Fatal("expected non-nil filter")
		}
		testCases := []struct {
			path string
			want bool
		}{
			{filepath.Join("data", "a.json"), true},           // included, not excluded
			{filepath.Join("data", "secret", "k.key"), false}, // included but excluded
			{filepath.Join("data", "x.tmp"), false},           // included but excluded
			{filepath.Join("other", "a.json"), false},         // not included
		}
		for _, tc := range testCases {
			if got := filter(tc.path); got != tc.want {
				t.Errorf("filter(%q) = %v, want %v", tc.path, got, tc.want)
			}
		}
	})

	t.Run("exclude only", func(t *testing.T) {
		filter := combineFilters(nil, compileMatcher([]string{"*.log"}))
		if got := filter("app.log"); got != false {
			t.Errorf("filter(app.log) = %v, want false", got)
		}
		if got := filter("app.txt"); got != true {
			t.Errorf("filter(app.txt) = %v, want true", got)
		}
	})

	t.Run("include only", func(t *testing.T) {
		filter := combineFilters(compileMatcher([]string{"data/**"}), nil)
		if got := filter(filepath.Join("data", "a")); got != true {
			t.Errorf("filter(data/a) = %v, want true", got)
		}
		if got := filter(filepath.Join("other", "a")); got != false {
			t.Errorf("filter(other/a) = %v, want false", got)
		}
	})

	t.Run("both nil", func(t *testing.T) {
		if filter := combineFilters(nil, nil); filter != nil {
			t.Error("expected nil filter when both include and exclude are nil")
		}
	})
}
