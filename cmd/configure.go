package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/erdos-ai/r2/pkg"

	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

// getConfigPath returns the path to the ~/.r2 configuration file, accounting for different
// operating systems' conventions for naming the home directory.
func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(homeDir, ".r2")
}

// R2ConfigFile globally defines the path to the ~/.r2 configuration file.
var R2ConfigFile = getConfigPath()

// profileNameRegex defines the valid characters for profile names
var profileNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+$`)

// normalizeProfileName converts a profile name to lowercase and trims whitespace.
// This ensures case-insensitive profile handling.
func normalizeProfileName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// validateProfileName checks if a profile name is valid.
// Valid profile names contain only alphanumeric characters, underscores, hyphens, and dots.
func validateProfileName(name string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}
	if !profileNameRegex.MatchString(name) {
		return fmt.Errorf("profile name can only contain alphanumeric characters, underscores, hyphens, and dots")
	}
	return nil
}

// ensureSecurePermissions sets secure file permissions (0600) on the config file.
// This ensures that only the owner can read/write the file containing credentials.
func ensureSecurePermissions() {
	if _, err := os.Stat(R2ConfigFile); err == nil {
		if err := os.Chmod(R2ConfigFile, 0600); err != nil {
			log.Printf("Warning: Failed to set secure permissions on config file: %v", err)
		}
	}
}

// getProfile returns the Cloudflare R2 credentials for the specified profile. If the profile does
// not exist, it is created interactively and saved to the ~/.r2 configuration file.
func getProfile(profileName string) pkg.Config {
	// Normalize profile name for case-insensitive lookup
	profileName = normalizeProfileName(profileName)

	// Get profiles
	profiles := getConfig(false)

	// If profile exists, return it
	for _, profile := range profiles {
		if profile.Profile == profileName {
			return profile
		}
	}

	// Profile doesn't exist, create new one and save to ~/.r2 config file
	profile := getCredentials(profileName)
	writeConfig(profile)

	return profile
}

// getCredentials prompts the user to enter the Cloudflare R2 credentials for a specified profile.
// If no profile is specified, the user is prompted to enter a profile name.
func getCredentials(profile string) pkg.Config {
	var c pkg.Config

	// Get profile
	if profile == "" {
		// Get profile name
		fmt.Print("Profile [default]: ")
		fmt.Scanln(&profile)
		if profile == "" {
			profile = "default"
		}
	}
	c.Profile = normalizeProfileName(profile)

	// Get account ID
	fmt.Print("Account ID: ")
	fmt.Scanln(&c.AccountID)

	// Get access key ID
	fmt.Print("Access Key ID: ")
	fmt.Scanln(&c.AccessKeyID)

	// Get secret access key
	fmt.Print("Secret Access Key: ")
	fmt.Scanln(&c.SecretAccessKey)

	return c
}

// Parse configuration file and return profiles
func getConfig(createIfNotPresent bool) map[string]pkg.Config {
	// Ensure secure permissions on existing config file
	ensureSecurePermissions()

	// Create configuration file if it doesn't exist
	if _, err := os.Stat(R2ConfigFile); os.IsNotExist(err) {
		// If not creating configuration file, return empty map
		if !createIfNotPresent {
			return make(map[string]pkg.Config)
		}

		// Create file with secure permissions (0600 - owner read/write only)
		f, err := os.OpenFile(R2ConfigFile, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatal(err)
		}
		f.Close()

		// Get credentials interactively and write to configuration file
		writeConfig(getCredentials(""))
	}

	// Load INI file
	cfg, err := ini.Load(R2ConfigFile)
	if err != nil {
		log.Fatalf("Failed to load config file: %v", err)
	}

	// Parse sections into profiles
	profiles := make(map[string]pkg.Config)

	for _, section := range cfg.Sections() {
		// Skip default section (unnamed)
		if section.Name() == ini.DEFAULT_SECTION {
			continue
		}

		profile := pkg.Config{
			Profile:         normalizeProfileName(section.Name()),
			AccountID:       section.Key("account_id").String(),
			AccessKeyID:     section.Key("access_key_id").String(),
			SecretAccessKey: section.Key("secret_access_key").String(),
		}

		// Only add if has ALL credentials (complete profile)
		if profile.AccountID != "" && profile.AccessKeyID != "" && profile.SecretAccessKey != "" {
			profiles[normalizeProfileName(section.Name())] = profile
		}
	}

	return profiles
}

// listProfiles returns a list of all profiles in the ~/.r2 configuration file. Profile names are
// sorted alphabetically, irrespective of case, with the default profile always first.
func listProfiles() []string {
	// Get profiles
	profiles := getConfig(false)

	// Get profile names and sort alphabetically (default profile is always first)
	var profileNames []string
	for _, p := range profiles {
		if p.Profile != "default" {
			profileNames = append(profileNames, p.Profile)
		}
	}

	sort.Slice(profileNames, func(i, j int) bool {
		return strings.ToLower(profileNames[i]) < strings.ToLower(profileNames[j])
	})

	if _, ok := profiles["default"]; ok {
		profileNames = append([]string{"default"}, profileNames...)
	}

	return profileNames
}

// sortConfig ensures profiles are sorted with 'default' first, then alphabetically (case-insensitive).
// This maintains consistency in the config file format.
func sortConfig(cfg *ini.File) *ini.File {
	sorted := ini.Empty()
	sections := cfg.Sections()

	// Collect section names
	var names []string
	var defaultSection *ini.Section

	for _, sec := range sections {
		if sec.Name() == ini.DEFAULT_SECTION {
			continue
		}
		if sec.Name() == "default" {
			defaultSection = sec
		} else {
			names = append(names, sec.Name())
		}
	}

	// Sort non-default sections alphabetically (case-insensitive)
	sort.Slice(names, func(i, j int) bool {
		return strings.ToLower(names[i]) < strings.ToLower(names[j])
	})

	// Add default first
	if defaultSection != nil {
		copySection(sorted, defaultSection)
	}

	// Add others in sorted order
	for _, name := range names {
		copySection(sorted, cfg.Section(name))
	}

	return sorted
}

// copySection copies a section and all its keys from source to destination INI file
func copySection(dst *ini.File, src *ini.Section) {
	newSec, _ := dst.NewSection(src.Name())
	for _, key := range src.Keys() {
		newSec.Key(key.Name()).SetValue(key.String())
	}
}

// writeConfig writes the provided profiles to the ~/.r2 configuration file. If a profile already
// exists, it is overwritten. If all credentials are not provided, the function fails. Profiles are
// sorted alphabetically, irrespective of case, with the default profile always first.
func writeConfig(c pkg.Config) {
	// Normalize and validate profile name
	c.Profile = normalizeProfileName(c.Profile)
	if err := validateProfileName(c.Profile); err != nil {
		log.Fatalf("Invalid profile name: %v", err)
	}

	// Trim whitespace from credentials
	c.AccountID = strings.TrimSpace(c.AccountID)
	c.AccessKeyID = strings.TrimSpace(c.AccessKeyID)
	c.SecretAccessKey = strings.TrimSpace(c.SecretAccessKey)

	// Validate credentials
	if c.AccountID == "" || c.AccessKeyID == "" || c.SecretAccessKey == "" {
		log.Fatal("All credentials must be provided and cannot be empty or contain only whitespace")
	}

	// Ensure secure permissions on existing config file
	ensureSecurePermissions()

	// Load or create INI file
	var cfg *ini.File
	var err error

	if _, statErr := os.Stat(R2ConfigFile); os.IsNotExist(statErr) {
		cfg = ini.Empty()
	} else {
		cfg, err = ini.Load(R2ConfigFile)
		if err != nil {
			log.Fatalf("Failed to load config file: %v", err)
		}
	}

	// Delete any existing sections that normalize to the same name (for migration)
	// This handles cases where old config has [PRODUCTION] and we're writing [production]
	for _, section := range cfg.Sections() {
		if section.Name() == ini.DEFAULT_SECTION {
			continue
		}
		if normalizeProfileName(section.Name()) == c.Profile {
			cfg.DeleteSection(section.Name())
		}
	}

	// Create new section with normalized name
	section, err := cfg.NewSection(c.Profile)
	if err != nil {
		log.Fatalf("Failed to create section: %v", err)
	}

	// Set values
	section.Key("account_id").SetValue(c.AccountID)
	section.Key("access_key_id").SetValue(c.AccessKeyID)
	section.Key("secret_access_key").SetValue(c.SecretAccessKey)

	// Sort config to maintain consistent ordering
	sortedCfg := sortConfig(cfg)

	// Atomic write: write to temp file first, then rename
	tempFile := R2ConfigFile + ".tmp"
	err = sortedCfg.SaveTo(tempFile)
	if err != nil {
		log.Fatalf("Failed to save config file: %v", err)
	}

	// Set secure permissions on temp file before rename
	if err := os.Chmod(tempFile, 0600); err != nil {
		os.Remove(tempFile) // Clean up temp file
		log.Fatalf("Failed to set permissions on config file: %v", err)
	}

	// Atomic rename (overwrites existing file on most filesystems)
	if err := os.Rename(tempFile, R2ConfigFile); err != nil {
		os.Remove(tempFile) // Clean up temp file
		log.Fatalf("Failed to write config file: %v", err)
	}

	// Ensure final file has secure permissions
	ensureSecurePermissions()
}

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure R2 access",
	Long: `Configure R2 access by providing Cloudflare R2 API Token credentials.

Configuration can be done interactively or by passing flags. If you pass flags,
you must provide both the access key ID and secret access key, otherwise the
command will fail.

To configure interactively, run:
  r2 configure

To configure with flags, run:
  r2 configure --access-key-id <access-key-id> \
    --secret-access-key <secret-access-key>

If you have multiple R2 tokens, you can configure a named profile by passing
the --profile flag.

  Interactively:
    r2 configure --profile my-profile

  With flags:
    r2 configure --profile my-profile --access-key-id <access-key-id> \
      --secret-access-key <secret-access-key>

Profiles are stored in ~/.r2 and can be used by passing the --profile flag to
any command.

To list available profiles, run:
  r2 configure --list

To generate an API Token, follow Cloudflare's guide at:
  https://developers.cloudflare.com/r2/data-access/s3-api/tokens/

Be careful not to share your API Token credentials with anyone.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Handle list flag
		list, err := cmd.Flags().GetBool("list")
		if err != nil {
			log.Fatal(err)
		}
		if list {
			// List profiles
			fmt.Println(strings.Join(listProfiles(), "\n"))
		} else {
			// Parse configuration
			var c pkg.Config
			var err error

			// Get profile name
			c.Profile, err = cmd.Flags().GetString("profile")
			if err != nil {
				log.Fatal(err)
			}

			// Get account ID
			c.AccountID, err = cmd.Flags().GetString("account-id")
			if err != nil {
				log.Fatal(err)
			}

			// Get access key ID
			c.AccessKeyID, err = cmd.Flags().GetString("access-key-id")
			if err != nil {
				log.Fatal(err)
			}

			// Get secret access key
			c.SecretAccessKey, err = cmd.Flags().GetString("secret-access-key")
			if err != nil {
				log.Fatal(err)
			}

			// Either access key ID or secret access key not passed but not both
			if (c.AccessKeyID == "" && c.SecretAccessKey != "") || (c.AccessKeyID != "" && c.SecretAccessKey == "") {
				log.Fatal(`Error: You must either provide both the access key ID and secret access key or
	neither to configure interactively.

	For more information, run:
		r2 help configure`)
			} else {
				// Check if configuration provided
				if c.AccountID != "" && c.AccessKeyID != "" && c.SecretAccessKey != "" {
					writeConfig(c)
				} else {
					// If no configuration provided, get configuration interactively
					writeConfig(getCredentials(""))
				}
			}
		}
	},
}

// init adds the configure command to the root command and adds flags to the configure command
func init() {
	// Add the configure command to the root command
	rootCmd.AddCommand(configureCmd)

	// Add flags to the configure command
	configureCmd.Flags().BoolP("list", "l", false, "List all named profiles")
	configureCmd.Flags().String("profile", "", "Configure a named profile")
	configureCmd.Flags().String("account-id", "", "R2 Account ID")
	configureCmd.Flags().String("access-key-id", "", "R2 Access Key ID")
	configureCmd.Flags().String("secret-access-key", "", "R2 Secret Access Key")
}
