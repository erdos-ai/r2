package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// loadIncludeFilter reads include patterns from a file and returns a matcher that
// reports whether a path matches any of them. An include file with no usable
// patterns is an error, since it would cause nothing to be synced.
func loadIncludeFilter(path string) (func(string) bool, error) {
	raws, err := readPatternLines(path)
	if err != nil {
		return nil, err
	}
	patterns, err := normalizePatterns(raws, "include")
	if err != nil {
		return nil, err
	}
	if len(patterns) == 0 {
		return nil, fmt.Errorf("include file contains no patterns; nothing would be synced")
	}
	return compileMatcher(patterns), nil
}

// loadExcludeFilter builds a matcher from exclude patterns drawn from an optional
// file (excludePath, "" if unset) and inline patterns (from repeated --exclude
// flags). The two sources are unioned. Unlike include, an empty exclude set is a
// no-op rather than an error: it returns (nil, nil), meaning "exclude nothing".
func loadExcludeFilter(excludePath string, inline []string) (func(string) bool, error) {
	var raws []string
	if excludePath != "" {
		lines, err := readPatternLines(excludePath)
		if err != nil {
			return nil, err
		}
		raws = append(raws, lines...)
	}
	for _, p := range inline {
		if strings.TrimSpace(p) == "" {
			continue
		}
		raws = append(raws, p)
	}
	if len(raws) == 0 {
		return nil, nil
	}
	patterns, err := normalizePatterns(raws, "exclude")
	if err != nil {
		return nil, err
	}
	return compileMatcher(patterns), nil
}

// combineFilters merges an include matcher and an exclude matcher into a single
// keep/skip predicate (true = sync the file). A file is kept when it matches the
// include filter (if any) and does not match the exclude filter (if any); exclude
// wins on conflict. Returns nil when both matchers are nil, preserving the
// "sync everything" default.
func combineFilters(include, exclude func(string) bool) func(string) bool {
	if include == nil && exclude == nil {
		return nil
	}
	return func(relPath string) bool {
		if include != nil && !include(relPath) {
			return false
		}
		if exclude != nil && exclude(relPath) {
			return false
		}
		return true
	}
}

// readPatternLines opens a pattern file and returns its meaningful glob patterns:
// blank lines and comments are dropped, inline comments are stripped, and each
// returned pattern is trimmed of surrounding whitespace.
func readPatternLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read pattern file %q: %w", path, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		pattern, ok := stripComment(scanner.Text())
		if !ok {
			continue
		}
		lines = append(lines, pattern)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read pattern file %q: %w", path, err)
	}
	return lines, nil
}

// stripComment removes comments from a pattern-file line and returns the remaining
// glob pattern, reporting false when nothing usable remains. A line whose first
// non-whitespace character is "#" is a full-line comment. Otherwise an unescaped
// "#" that follows whitespace starts an inline comment and is dropped along with
// the rest of the line. A "#" with no preceding space stays literal, and "\#"
// escapes a literal "#" anywhere — including the start of a name — emitting a bare
// "#" so the backslash never reaches the matcher (a leading "\" looks absolute).
func stripComment(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return "", false
	}

	var b strings.Builder
	prevSpace := false
	for i := 0; i < len(trimmed); i++ {
		c := trimmed[i]
		switch {
		case c == '\\' && i+1 < len(trimmed) && trimmed[i+1] == '#':
			// Escaped "#": emit a literal '#' and drop the backslash.
			b.WriteByte('#')
			i++
			prevSpace = false
		case c == '#' && prevSpace:
			// Unescaped '#' after whitespace begins an inline comment.
			pattern := strings.TrimRight(b.String(), " \t")
			return pattern, pattern != ""
		default:
			b.WriteByte(c)
			prevSpace = c == ' ' || c == '\t'
		}
	}
	pattern := strings.TrimRight(b.String(), " \t")
	return pattern, pattern != ""
}

// normalizePatterns normalizes and validates each raw pattern, returning the first
// error encountered. kind ("include" or "exclude") is used only in error messages.
func normalizePatterns(raws []string, kind string) ([]string, error) {
	var patterns []string
	for _, raw := range raws {
		normalized, err := normalizePattern(raw, kind)
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, normalized)
	}
	return patterns, nil
}

// normalizePattern validates a single raw glob pattern and returns its normalized
// form (forward slashes, no leading "./" or "/", trailing "/" expanded to "/**").
// Absolute patterns are rejected because patterns are relative to the sync source
// root. kind ("include" or "exclude") is used only in error messages.
func normalizePattern(raw, kind string) (string, error) {
	if isAbsolutePattern(raw) {
		return "", fmt.Errorf("%s pattern %q is absolute; patterns must be relative to the sync source root (e.g. \"db-dump/\" not \"/d/db-dump/\")", kind, raw)
	}

	normalized := filepath.ToSlash(raw)
	normalized = strings.TrimPrefix(normalized, "./")
	normalized = strings.TrimPrefix(normalized, "/")
	if strings.HasSuffix(normalized, "/") {
		normalized = strings.TrimSuffix(normalized, "/") + "/**"
	}

	if _, err := doublestar.Match(normalized, ""); err != nil {
		return "", fmt.Errorf("invalid %s pattern %q: %w", kind, raw, err)
	}
	return normalized, nil
}

// compileMatcher returns a function reporting whether a relative path matches any
// of the normalized patterns. It returns nil when patterns is empty.
func compileMatcher(patterns []string) func(string) bool {
	if len(patterns) == 0 {
		return nil
	}
	return func(relPath string) bool {
		normalized := filepath.ToSlash(relPath)
		normalized = strings.TrimPrefix(normalized, "./")
		normalized = strings.TrimPrefix(normalized, "/")
		for _, pattern := range patterns {
			matched, err := doublestar.Match(pattern, normalized)
			if err != nil {
				continue
			}
			if matched {
				return true
			}
		}
		return false
	}
}

func isAbsolutePattern(pattern string) bool {
	if pattern == "" {
		return false
	}
	if strings.HasPrefix(pattern, "/") || strings.HasPrefix(pattern, "\\") {
		return true
	}
	if len(pattern) >= 2 && pattern[1] == ':' {
		c := pattern[0]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			return true
		}
	}
	return false
}
