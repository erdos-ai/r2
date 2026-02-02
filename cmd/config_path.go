package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func applyConfigFlag(path string) error {
	if path == "" {
		return nil
	}

	expanded := os.ExpandEnv(path)
	expanded, err := expandHome(expanded)
	if err != nil {
		return err
	}

	absPath, err := filepath.Abs(expanded)
	if err != nil {
		return fmt.Errorf("resolve config path: %w", err)
	}

	R2ConfigFile = absPath
	return nil
}

func expandHome(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		if path == "~" {
			return homeDir, nil
		}
		return filepath.Join(homeDir, path[2:]), nil
	}
	return path, nil
}
