package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyConfigFlag(t *testing.T) {
	original := R2ConfigFile
	defer func() {
		R2ConfigFile = original
	}()

	tempDir := t.TempDir()
	homeBackup := os.Getenv("HOME")
	defer os.Setenv("HOME", homeBackup)
	if err := os.Setenv("HOME", tempDir); err != nil {
		t.Fatalf("set HOME: %v", err)
	}

	envPath := filepath.Join(tempDir, "env.r2")
	if err := os.Setenv("TEST_R2_CONFIG", envPath); err != nil {
		t.Fatalf("set env: %v", err)
	}

	if err := applyConfigFlag("~/config.r2"); err != nil {
		t.Fatalf("applyConfigFlag: %v", err)
	}
	wantHome := filepath.Join(tempDir, "config.r2")
	if R2ConfigFile != wantHome {
		t.Fatalf("expected %q, got %q", wantHome, R2ConfigFile)
	}

	if err := applyConfigFlag("$TEST_R2_CONFIG"); err != nil {
		t.Fatalf("applyConfigFlag env: %v", err)
	}
	if R2ConfigFile != envPath {
		t.Fatalf("expected %q, got %q", envPath, R2ConfigFile)
	}

	if err := applyConfigFlag("relative.r2"); err != nil {
		t.Fatalf("applyConfigFlag relative: %v", err)
	}
	if !filepath.IsAbs(R2ConfigFile) {
		t.Fatalf("expected absolute path, got %q", R2ConfigFile)
	}
}

func TestRootCmd_ConfigFlag_UsesCustomConfig(t *testing.T) {
	original := R2ConfigFile
	defer func() {
		R2ConfigFile = original
	}()

	tempDir := t.TempDir()
	defaultPath := filepath.Join(tempDir, "default.r2")
	customPath := filepath.Join(tempDir, "custom.r2")

	defaultContent := strings.Join([]string{
		"[default]",
		"account_id=default-account",
		"access_key_id=default-key",
		"secret_access_key=default-secret",
		"",
	}, "\n")
	customContent := strings.Join([]string{
		"[custom]",
		"account_id=custom-account",
		"access_key_id=custom-key",
		"secret_access_key=custom-secret",
		"",
	}, "\n")
	if err := os.WriteFile(defaultPath, []byte(defaultContent), 0600); err != nil {
		t.Fatalf("write default config: %v", err)
	}
	if err := os.WriteFile(customPath, []byte(customContent), 0600); err != nil {
		t.Fatalf("write custom config: %v", err)
	}

	R2ConfigFile = defaultPath

	output := captureStdout(t, func() {
		rootCmd.SetArgs([]string{"--config", customPath, "configure", "--list"})
		defer rootCmd.SetArgs([]string{})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("execute: %v", err)
		}
	})

	if strings.TrimSpace(output) != "custom" {
		t.Fatalf("expected output %q, got %q", "custom", strings.TrimSpace(output))
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	os.Stdout = original

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read output: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close reader: %v", err)
	}

	return buf.String()
}
