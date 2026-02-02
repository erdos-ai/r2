package cmd

import (
	"log"

	"github.com/erdos-one/r2/pkg"

	"github.com/spf13/cobra"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Syncs directories and R2 prefixes.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		includePath, err := cmd.Flags().GetString("include-from")
		if err != nil {
			log.Fatal(err)
		}
		var includeFilter func(string) bool
		if includePath != "" {
			includeFilter, err = loadIncludeFilter(includePath)
			if err != nil {
				log.Fatal(err)
			}
		}

		// Get profile client
		profileName, err := cmd.Flags().GetString("profile")
		if err != nil {
			log.Fatal(err)
		}
		c := pkg.Client(getProfile(profileName))

		// If a bucket name is provided, create the bucket
		if len(args) == 2 {
			sourcePath := args[0]
			destinationPath := args[1]
			if !pkg.IsR2URI(sourcePath) && pkg.IsR2URI(destinationPath) {
				// Sync local directory to R2 bucket
				destURI := pkg.ParseR2URISafe(destinationPath)
				b := c.Bucket(destURI.Bucket)
				b.SyncLocalToR2WithPrefix(sourcePath, destURI.Path, includeFilter)
			} else if pkg.IsR2URI(sourcePath) && !pkg.IsR2URI(destinationPath) {
				// Sync R2 bucket to local directory
				sourceURI := pkg.ParseR2URISafe(sourcePath)
				b := c.Bucket(sourceURI.Bucket)
				b.SyncR2ToLocalWithPrefix(destinationPath, sourceURI.Path, includeFilter)
			} else if pkg.IsR2URI(sourcePath) && pkg.IsR2URI(destinationPath) {
				// Sync R2 bucket to R2 bucket
				sourceURI := pkg.ParseR2URISafe(sourcePath)
				destURI := pkg.ParseR2URISafe(destinationPath)
				b := c.Bucket(sourceURI.Bucket)
				destBucket := c.Bucket(destURI.Bucket)
				b.SyncR2ToR2WithPrefix(destBucket, sourceURI.Path, destURI.Path, includeFilter)
			} else if !pkg.IsR2URI(sourcePath) && !pkg.IsR2URI(destinationPath) {
				// Both paths are local - not supported
				log.Fatal("Local-to-local sync is not supported. At least one path must be an R2 URI (r2://bucket/path).")
			}
		} else {
			log.Fatal("Please provide both a source and destination path.")
		}
	},
}

func init() {
	// Add the sync command to the root command
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().String("include-from", "", "Read include patterns from file (relative to source root)")
}
