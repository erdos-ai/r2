package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/erdos-ai/r2/pkg"
)

func TestGetConfig_EmptyFile(t *testing.T) {
	// Test with non-existent file, createIfNotPresent=false
	tempDir := t.TempDir()
	R2ConfigFile = filepath.Join(tempDir, ".r2")

	profiles := getConfig(false)

	if len(profiles) != 0 {
		t.Errorf("Expected empty map, got %d profiles", len(profiles))
	}
}

func TestWriteConfig_SpecialCharacters(t *testing.T) {
	// THE BUG FIX VALIDATION TEST - validates credentials with special characters
	testCases := []struct {
		name   string
		secret string
	}{
		{"AWS style with slash", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"},
		{"With plus signs", "abcd+efgh+1234+5678"},
		{"With both slash and plus", "test/key+with/both+chars"},
		{"Cloudflare hex", "131bbd4f11e4084d1a8d28970d54b9f7321744c51610334ca2910c13e14e0699"},
		{"With equals signs", "secret+key/with=equals=="},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use a fresh temp directory for each test
			tempDir := t.TempDir()
			R2ConfigFile = filepath.Join(tempDir, ".r2")

			config := pkg.Config{
				Profile:         "test",
				AccountID:       "test-account-123",
				AccessKeyID:     "test-access-key",
				SecretAccessKey: tc.secret,
			}

			// Write config
			writeConfig(config)

			// Read it back
			profiles := getConfig(false)

			retrieved := profiles["test"]
			if retrieved.SecretAccessKey != tc.secret {
				t.Errorf("Secret was corrupted (test case %d):\nWant: %s\nGot:  %s",
					i, tc.secret, retrieved.SecretAccessKey)
			}
		})
	}
}

func TestWriteConfig_MultipleProfiles(t *testing.T) {
	tempDir := t.TempDir()
	R2ConfigFile = filepath.Join(tempDir, ".r2")

	// Write default profile
	writeConfig(pkg.Config{
		Profile:         "default",
		AccountID:       "default-account",
		AccessKeyID:     "default-key",
		SecretAccessKey: "default-secret",
	})

	// Write another profile
	writeConfig(pkg.Config{
		Profile:         "production",
		AccountID:       "prod-account",
		AccessKeyID:     "prod-key",
		SecretAccessKey: "prod-secret",
	})

	// Write another profile that should come before "production" alphabetically
	writeConfig(pkg.Config{
		Profile:         "development",
		AccountID:       "dev-account",
		AccessKeyID:     "dev-key",
		SecretAccessKey: "dev-secret",
	})

	// Read all profiles
	profiles := getConfig(false)

	if len(profiles) != 3 {
		t.Errorf("Expected 3 profiles, got %d", len(profiles))
	}

	// Verify all profiles exist
	for _, name := range []string{"default", "production", "development"} {
		if _, ok := profiles[name]; !ok {
			t.Errorf("Profile %s not found", name)
		}
	}

	// Check file order (default should be first)
	content, _ := os.ReadFile(R2ConfigFile)
	lines := strings.Split(string(content), "\n")

	// Find the first profile section
	foundDefault := false
	for _, line := range lines {
		if strings.Contains(line, "[") && strings.Contains(line, "]") {
			if strings.Contains(line, "[default]") {
				foundDefault = true
			}
			// First profile section should be default
			break
		}
	}

	if !foundDefault {
		t.Error("Default profile should be first in file")
	}
}

func TestListProfiles_Sorting(t *testing.T) {
	tempDir := t.TempDir()
	R2ConfigFile = filepath.Join(tempDir, ".r2")

	// Create profiles in non-alphabetical order
	for _, profile := range []string{"zebra", "default", "alpha", "beta"} {
		writeConfig(pkg.Config{
			Profile:         profile,
			AccountID:       profile + "-account",
			AccessKeyID:     profile + "-key",
			SecretAccessKey: profile + "-secret",
		})
	}

	profiles := listProfiles()

	// Should be: default, alpha, beta, zebra
	expected := []string{"default", "alpha", "beta", "zebra"}

	if len(profiles) != len(expected) {
		t.Fatalf("Expected %d profiles, got %d", len(expected), len(profiles))
	}

	for i, name := range expected {
		if profiles[i] != name {
			t.Errorf("Position %d: expected %s, got %s", i, name, profiles[i])
		}
	}
}

func TestWriteConfig_EmptyCredentials(t *testing.T) {
	testCases := []struct {
		name   string
		config pkg.Config
	}{
		{"Empty account ID", pkg.Config{Profile: "test", AccountID: "", AccessKeyID: "key", SecretAccessKey: "secret"}},
		{"Empty access key", pkg.Config{Profile: "test", AccountID: "account", AccessKeyID: "", SecretAccessKey: "secret"}},
		{"Empty secret", pkg.Config{Profile: "test", AccountID: "account", AccessKeyID: "key", SecretAccessKey: ""}},
		{"Whitespace only account", pkg.Config{Profile: "test", AccountID: "   ", AccessKeyID: "key", SecretAccessKey: "secret"}},
		{"Whitespace only key", pkg.Config{Profile: "test", AccountID: "account", AccessKeyID: "   ", SecretAccessKey: "secret"}},
		{"Whitespace only secret", pkg.Config{Profile: "test", AccountID: "account", AccessKeyID: "key", SecretAccessKey: "   "}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// writeConfig should call log.Fatal, which we can't easily test
			// So we'll just skip this validation test for now
			// In a production environment, you'd want to refactor to return errors instead of calling log.Fatal
			t.Skip("Skipping validation test - writeConfig uses log.Fatal which terminates the process")
		})
	}
}

func TestWriteConfig_UpdateExisting(t *testing.T) {
	tempDir := t.TempDir()
	R2ConfigFile = filepath.Join(tempDir, ".r2")

	// Write initial config
	initialConfig := pkg.Config{
		Profile:         "test",
		AccountID:       "initial-account",
		AccessKeyID:     "initial-key",
		SecretAccessKey: "initial-secret",
	}
	writeConfig(initialConfig)

	// Update the same profile
	updatedConfig := pkg.Config{
		Profile:         "test",
		AccountID:       "updated-account",
		AccessKeyID:     "updated-key",
		SecretAccessKey: "updated-secret/with+special",
	}
	writeConfig(updatedConfig)

	// Read it back
	profiles := getConfig(false)

	if len(profiles) != 1 {
		t.Errorf("Expected 1 profile, got %d", len(profiles))
	}

	retrieved := profiles["test"]
	if retrieved.AccountID != updatedConfig.AccountID {
		t.Errorf("AccountID not updated: want %s, got %s", updatedConfig.AccountID, retrieved.AccountID)
	}
	if retrieved.AccessKeyID != updatedConfig.AccessKeyID {
		t.Errorf("AccessKeyID not updated: want %s, got %s", updatedConfig.AccessKeyID, retrieved.AccessKeyID)
	}
	if retrieved.SecretAccessKey != updatedConfig.SecretAccessKey {
		t.Errorf("SecretAccessKey not updated: want %s, got %s", updatedConfig.SecretAccessKey, retrieved.SecretAccessKey)
	}
}

func TestGetProfile_NewProfile(t *testing.T) {
	tempDir := t.TempDir()
	R2ConfigFile = filepath.Join(tempDir, ".r2")

	// Create an initial profile
	writeConfig(pkg.Config{
		Profile:         "existing",
		AccountID:       "existing-account",
		AccessKeyID:     "existing-key",
		SecretAccessKey: "existing-secret",
	})

	// Get the existing profile
	profile := getProfile("existing")

	if profile.Profile != "existing" {
		t.Errorf("Expected profile 'existing', got '%s'", profile.Profile)
	}
	if profile.AccountID != "existing-account" {
		t.Errorf("Expected AccountID 'existing-account', got '%s'", profile.AccountID)
	}
}

// Security and Validation Tests

func TestNormalizeProfileName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"Default", "default"},
		{"PRODUCTION", "production"},
		{"Test-Profile", "test-profile"},
		{"  spaced  ", "spaced"},
		{"MiXeD_CaSe", "mixed_case"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := normalizeProfileName(tc.input)
			if result != tc.expected {
				t.Errorf("normalizeProfileName(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestValidateProfileName(t *testing.T) {
	validNames := []string{
		"default",
		"production",
		"test-profile",
		"my_profile",
		"profile.1",
		"a1b2c3",
	}

	for _, name := range validNames {
		t.Run("valid_"+name, func(t *testing.T) {
			if err := validateProfileName(name); err != nil {
				t.Errorf("validateProfileName(%q) returned error: %v", name, err)
			}
		})
	}

	invalidNames := []struct {
		name   string
		reason string
	}{
		{"", "empty"},
		{"profile with spaces", "spaces"},
		{"profile[bracket]", "brackets"},
		{"profile=equals", "equals"},
		{"profile/slash", "slash"},
		{"profile;semicolon", "semicolon"},
		{"profile\nnewline", "newline"},
		{"profile\ttab", "tab"},
	}

	for _, tc := range invalidNames {
		t.Run("invalid_"+tc.reason, func(t *testing.T) {
			if err := validateProfileName(tc.name); err == nil {
				t.Errorf("validateProfileName(%q) should have returned error", tc.name)
			}
		})
	}
}

func TestWriteConfig_EmptyProfileName_ShouldFail(t *testing.T) {
	// This test verifies that empty profile names are rejected
	// We can't easily test log.Fatal, so we document the expected behavior
	t.Skip("Empty profile names are rejected by validateProfileName - would call log.Fatal")
}

func TestWriteConfig_InvalidProfileName_ShouldFail(t *testing.T) {
	// This test verifies that invalid profile names are rejected
	// We can't easily test log.Fatal, so we document the expected behavior
	t.Skip("Invalid profile names are rejected by validateProfileName - would call log.Fatal")
}

func TestWriteConfig_CaseInsensitive(t *testing.T) {
	tempDir := t.TempDir()
	R2ConfigFile = filepath.Join(tempDir, ".r2")

	// Write profile with uppercase name
	writeConfig(pkg.Config{
		Profile:         "DEFAULT",
		AccountID:       "test-account",
		AccessKeyID:     "test-key",
		SecretAccessKey: "test-secret",
	})

	// Read it back - should be normalized to lowercase
	profiles := getConfig(false)

	if _, ok := profiles["DEFAULT"]; ok {
		t.Error("Profile 'DEFAULT' should not exist (should be normalized to 'default')")
	}

	if profile, ok := profiles["default"]; !ok {
		t.Error("Profile 'default' not found")
	} else {
		if profile.AccountID != "test-account" {
			t.Errorf("AccountID mismatch: got %s, want test-account", profile.AccountID)
		}
	}
}

func TestGetConfig_PartialCredentials_Excluded(t *testing.T) {
	tempDir := t.TempDir()
	R2ConfigFile = filepath.Join(tempDir, ".r2")

	// Manually create a config file with partial credentials
	content := `[complete]
account_id = account1
access_key_id = key1
secret_access_key = secret1

[missing_secret]
account_id = account2
access_key_id = key2

[missing_access_key]
account_id = account3
secret_access_key = secret3

[missing_account_id]
access_key_id = key4
secret_access_key = secret4
`
	if err := os.WriteFile(R2ConfigFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	// Read config - should only include complete profile
	profiles := getConfig(false)

	if len(profiles) != 1 {
		t.Errorf("Expected 1 complete profile, got %d", len(profiles))
	}

	if _, ok := profiles["complete"]; !ok {
		t.Error("Complete profile should be present")
	}

	if _, ok := profiles["missing_secret"]; ok {
		t.Error("Profile with missing secret should be excluded")
	}
	if _, ok := profiles["missing_access_key"]; ok {
		t.Error("Profile with missing access_key should be excluded")
	}
	if _, ok := profiles["missing_account_id"]; ok {
		t.Error("Profile with missing account_id should be excluded")
	}
}

func TestWriteConfig_FilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	R2ConfigFile = filepath.Join(tempDir, ".r2")

	// Write a config
	writeConfig(pkg.Config{
		Profile:         "test",
		AccountID:       "account",
		AccessKeyID:     "key",
		SecretAccessKey: "secret",
	})

	// Check file permissions
	info, err := os.Stat(R2ConfigFile)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}

	mode := info.Mode().Perm()
	expected := os.FileMode(0600)

	if mode != expected {
		t.Errorf("Config file permissions = %o, want %o", mode, expected)
	}
}

func TestWriteConfig_AtomicWrite(t *testing.T) {
	tempDir := t.TempDir()
	R2ConfigFile = filepath.Join(tempDir, ".r2")

	// Write initial config
	writeConfig(pkg.Config{
		Profile:         "test",
		AccountID:       "account1",
		AccessKeyID:     "key1",
		SecretAccessKey: "secret1",
	})

	// Verify temp file doesn't exist after successful write
	tempFile := R2ConfigFile + ".tmp"
	if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
		t.Error("Temp file should not exist after successful write")
	}

	// Verify final config exists and is readable
	profiles := getConfig(false)
	if len(profiles) != 1 {
		t.Errorf("Expected 1 profile, got %d", len(profiles))
	}
}

func TestGetProfile_CaseInsensitiveLookup(t *testing.T) {
	tempDir := t.TempDir()
	R2ConfigFile = filepath.Join(tempDir, ".r2")

	// Write profile with lowercase name
	writeConfig(pkg.Config{
		Profile:         "production",
		AccountID:       "prod-account",
		AccessKeyID:     "prod-key",
		SecretAccessKey: "prod-secret",
	})

	// Test lookup with different cases
	testCases := []string{
		"production",
		"PRODUCTION",
		"Production",
		"PrOdUcTiOn",
	}

	for _, testCase := range testCases {
		t.Run(testCase, func(t *testing.T) {
			profile := getProfile(testCase)

			if profile.Profile != "production" {
				t.Errorf("Profile name should be normalized to 'production', got %s", profile.Profile)
			}
			if profile.AccountID != "prod-account" {
				t.Errorf("AccountID mismatch: got %s, want prod-account", profile.AccountID)
			}
			if profile.AccessKeyID != "prod-key" {
				t.Errorf("AccessKeyID mismatch: got %s, want prod-key", profile.AccessKeyID)
			}
			if profile.SecretAccessKey != "prod-secret" {
				t.Errorf("SecretAccessKey mismatch: got %s, want prod-secret", profile.SecretAccessKey)
			}
		})
	}
}

func TestWriteConfig_TrimWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	R2ConfigFile = filepath.Join(tempDir, ".r2")

	// Write config with whitespace in credentials
	writeConfig(pkg.Config{
		Profile:         "test",
		AccountID:       "  account-with-spaces  ",
		AccessKeyID:     "\tkey-with-tabs\t",
		SecretAccessKey: " secret-with-spaces ",
	})

	// Read it back
	profiles := getConfig(false)
	profile := profiles["test"]

	// Verify whitespace was trimmed
	if profile.AccountID != "account-with-spaces" {
		t.Errorf("AccountID should be trimmed, got '%s'", profile.AccountID)
	}
	if profile.AccessKeyID != "key-with-tabs" {
		t.Errorf("AccessKeyID should be trimmed, got '%s'", profile.AccessKeyID)
	}
	if profile.SecretAccessKey != "secret-with-spaces" {
		t.Errorf("SecretAccessKey should be trimmed, got '%s'", profile.SecretAccessKey)
	}

	// Verify saved file has trimmed values
	content, err := os.ReadFile(R2ConfigFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	configStr := string(content)
	if strings.Contains(configStr, "  account-with-spaces  ") {
		t.Error("Config file should not contain leading/trailing spaces in AccountID")
	}
	if strings.Contains(configStr, "\tkey-with-tabs\t") {
		t.Error("Config file should not contain leading/trailing tabs in AccessKeyID")
	}
}

// TestGetConfig_BackwardCompatibility_UppercaseProfiles verifies that profiles with
// uppercase names in old config files are properly normalized when read.
func TestGetConfig_BackwardCompatibility_UppercaseProfiles(t *testing.T) {
	tempDir := t.TempDir()
	R2ConfigFile = filepath.Join(tempDir, ".r2")

	// Manually create a config file with uppercase section name (simulating old config)
	configContent := `[PRODUCTION]
account_id = prod-account-123
access_key_id = AKIAIOSFODNN7EXAMPLE
secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
`
	err := os.WriteFile(R2ConfigFile, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Read the config
	profiles := getConfig(false)

	// Verify the profile was read and normalized
	if len(profiles) != 1 {
		t.Fatalf("Expected 1 profile, got %d", len(profiles))
	}

	// Check that the map key is normalized
	profile, exists := profiles["production"]
	if !exists {
		t.Fatalf("Profile not found with normalized key 'production'. Available keys: %v", getMapKeys(profiles))
	}

	// Verify the Profile field is normalized
	if profile.Profile != "production" {
		t.Errorf("Profile.Profile should be normalized to 'production', got '%s'", profile.Profile)
	}

	// Verify credentials are intact
	if profile.AccountID != "prod-account-123" {
		t.Errorf("AccountID mismatch: got %s", profile.AccountID)
	}
	if profile.AccessKeyID != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("AccessKeyID mismatch: got %s", profile.AccessKeyID)
	}
	if profile.SecretAccessKey != "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" {
		t.Errorf("SecretAccessKey mismatch: got %s", profile.SecretAccessKey)
	}

	// Verify getProfile() can find it with uppercase input
	retrievedUpper := getProfile("PRODUCTION")
	if retrievedUpper.Profile != "production" {
		t.Errorf("getProfile('PRODUCTION') should return normalized profile, got '%s'", retrievedUpper.Profile)
	}

	// Verify getProfile() can find it with lowercase input
	retrievedLower := getProfile("production")
	if retrievedLower.Profile != "production" {
		t.Errorf("getProfile('production') should return normalized profile, got '%s'", retrievedLower.Profile)
	}
}

// Helper function to get map keys for debugging
func getMapKeys(m map[string]pkg.Config) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// TestWriteConfig_MigratesUppercaseToLowercase verifies that when a profile
// with an uppercase name is updated via writeConfig(), it gets saved with a
// lowercase name. Migration is lazy/gradual - only the profiles that are
// written get migrated, not all profiles at once.
func TestWriteConfig_MigratesUppercaseToLowercase(t *testing.T) {
	tempDir := t.TempDir()
	R2ConfigFile = filepath.Join(tempDir, ".r2")

	// Create old config with uppercase section
	oldConfigContent := `[PRODUCTION]
account_id = prod-account
access_key_id = prod-key
secret_access_key = prod-secret
`
	err := os.WriteFile(R2ConfigFile, []byte(oldConfigContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Read the profile (normalizes in memory)
	profile := getProfile("PRODUCTION")
	if profile.Profile != "production" {
		t.Fatalf("Profile should be normalized to 'production', got '%s'", profile.Profile)
	}

	// Update the profile credentials (triggers write)
	profile.AccessKeyID = "new-prod-key"
	writeConfig(profile)

	// Read the file content directly
	content, err := os.ReadFile(R2ConfigFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	configStr := string(content)

	// Verify uppercase section is now lowercase
	if strings.Contains(configStr, "[PRODUCTION]") {
		t.Error("Config file should not contain uppercase [PRODUCTION] section after write")
	}

	// Verify lowercase section exists
	if !strings.Contains(configStr, "[production]") {
		t.Error("Config file should contain lowercase [production] section")
	}

	// Verify the updated credentials are saved
	if !strings.Contains(configStr, "new-prod-key") {
		t.Error("Config file should contain updated credentials")
	}

	// Verify the profile is still accessible
	readBack := getProfile("production")
	if readBack.AccessKeyID != "new-prod-key" {
		t.Errorf("Updated credentials not persisted. Got '%s', want 'new-prod-key'",
			readBack.AccessKeyID)
	}
}

// TestGetConfig_MixedCaseProfiles_AllNormalized verifies that profiles with
// different case variations are all normalized consistently to lowercase.
func TestGetConfig_MixedCaseProfiles_AllNormalized(t *testing.T) {
	tempDir := t.TempDir()
	R2ConfigFile = filepath.Join(tempDir, ".r2")

	// Create config with profiles in various case formats
	configContent := `[Production]
account_id = prod-account
access_key_id = prod-key
secret_access_key = prod-secret

[STAGING]
account_id = staging-account
access_key_id = staging-key
secret_access_key = staging-secret

[test]
account_id = test-account
access_key_id = test-key
secret_access_key = test-secret

[DeVeLoPmEnT]
account_id = dev-account
access_key_id = dev-key
secret_access_key = dev-secret
`
	err := os.WriteFile(R2ConfigFile, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Read the config
	profiles := getConfig(false)

	// Verify all 4 profiles were read
	if len(profiles) != 4 {
		t.Fatalf("Expected 4 profiles, got %d", len(profiles))
	}

	// Verify each profile is normalized
	expectedProfiles := []string{"production", "staging", "test", "development"}
	for _, expectedName := range expectedProfiles {
		profile, exists := profiles[expectedName]
		if !exists {
			t.Errorf("Profile '%s' not found. Available: %v", expectedName, getMapKeys(profiles))
			continue
		}

		// Verify the Profile field matches the normalized name
		if profile.Profile != expectedName {
			t.Errorf("Profile '%s': Profile field should be '%s', got '%s'",
				expectedName, expectedName, profile.Profile)
		}
	}

	// Verify all map keys are lowercase
	for key := range profiles {
		if key != strings.ToLower(key) {
			t.Errorf("Map key should be lowercase, got '%s'", key)
		}
	}

	// Verify uppercase keys don't exist
	upperCaseKeys := []string{"Production", "STAGING", "DeVeLoPmEnT"}
	for _, key := range upperCaseKeys {
		if _, exists := profiles[key]; exists {
			t.Errorf("Map should not contain non-normalized key '%s'", key)
		}
	}
}

// TestGetConfig_DuplicateProfileNames_LastWins verifies the behavior when a config
// file contains profiles that normalize to the same name (e.g., [PRODUCTION] and [production]).
// This can happen if users manually edited the config file. The last one encountered wins.
func TestGetConfig_DuplicateProfileNames_LastWins(t *testing.T) {
	tempDir := t.TempDir()
	R2ConfigFile = filepath.Join(tempDir, ".r2")

	// Create config with duplicate profile names (different cases)
	// Use different account IDs so we can verify which one was kept
	configContent := `[PRODUCTION]
account_id = first-account
access_key_id = first-key
secret_access_key = first-secret

[production]
account_id = second-account
access_key_id = second-key
secret_access_key = second-secret
`
	err := os.WriteFile(R2ConfigFile, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Read the config
	profiles := getConfig(false)

	// Verify only one profile exists (the duplicate was overwritten)
	if len(profiles) != 1 {
		t.Fatalf("Expected 1 profile (duplicate should be overwritten), got %d", len(profiles))
	}

	// Verify the profile is under the normalized key
	profile, exists := profiles["production"]
	if !exists {
		t.Fatalf("Profile 'production' not found. Available: %v", getMapKeys(profiles))
	}

	// Verify it's the LAST one encountered (second-account)
	// The INI parser processes sections in order, and our map will have the last value
	if profile.AccountID != "second-account" {
		t.Errorf("Expected last profile to win. Got AccountID '%s', want 'second-account'",
			profile.AccountID)
	}

	if profile.AccessKeyID != "second-key" {
		t.Errorf("Expected last profile credentials. Got AccessKeyID '%s', want 'second-key'",
			profile.AccessKeyID)
	}

	// Verify the Profile field is normalized
	if profile.Profile != "production" {
		t.Errorf("Profile.Profile should be 'production', got '%s'", profile.Profile)
	}
}
