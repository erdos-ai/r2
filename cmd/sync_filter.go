package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

func loadIncludeFilter(path string) (func(string) bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read include file %q: %w", path, err)
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" || strings.HasPrefix(raw, "#") {
			continue
		}
		if isAbsolutePattern(raw) {
			return nil, fmt.Errorf("include pattern %q is absolute; patterns must be relative to the sync source root (e.g. \"db-dump/\" not \"/d/db-dump/\")", raw)
		}

		normalized := filepath.ToSlash(raw)
		normalized = strings.TrimPrefix(normalized, "./")
		normalized = strings.TrimPrefix(normalized, "/")
		if strings.HasSuffix(normalized, "/") {
			normalized = strings.TrimSuffix(normalized, "/") + "/**"
		}

		if _, err := doublestar.PathMatch(normalized, ""); err != nil {
			return nil, fmt.Errorf("invalid include pattern %q: %w", raw, err)
		}

		patterns = append(patterns, normalized)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read include file %q: %w", path, err)
	}
	if len(patterns) == 0 {
		return nil, fmt.Errorf("include file contains no patterns; nothing would be synced")
	}

	return func(relPath string) bool {
		normalized := filepath.ToSlash(relPath)
		normalized = strings.TrimPrefix(normalized, "./")
		normalized = strings.TrimPrefix(normalized, "/")
		for _, pattern := range patterns {
			matched, err := doublestar.PathMatch(pattern, normalized)
			if err != nil {
				continue
			}
			if matched {
				return true
			}
		}
		return false
	}, nil
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
