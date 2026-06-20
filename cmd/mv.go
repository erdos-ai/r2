package cmd

import (
	"log"
	"os"

	"github.com/erdos-ai/r2/pkg"

	"github.com/spf13/cobra"
)

// mvCmd represents the mv command
var mvCmd = &cobra.Command{
	Use:   "mv",
	Short: "Moves a local file or R2 object to another location locally or in R2.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
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
				// Move local file to R2
				destURI := pkg.ParseR2URISafe(destinationPath)
				b := c.Bucket(destURI.Bucket)
				b.Upload(sourcePath, destURI.Path)
				os.Remove(sourcePath)
			} else if pkg.IsR2URI(sourcePath) && !pkg.IsR2URI(destinationPath) {
				// Move R2 object to local file
				sourceURI := pkg.ParseR2URISafe(sourcePath)
				b := c.Bucket(sourceURI.Bucket)
				b.Download(sourceURI.Path, destinationPath)
				b.Delete(sourceURI.Path)
			} else if pkg.IsR2URI(sourcePath) && pkg.IsR2URI(destinationPath) {
				// Move R2 object to R2 object
				sourceURI := pkg.ParseR2URISafe(sourcePath)
				destURI := pkg.ParseR2URISafe(destinationPath)
				b := c.Bucket(sourceURI.Bucket)
				b.Copy(sourceURI.Path, destURI)
				b.Delete(sourceURI.Path)
			}
		} else {
			log.Fatal("Please provide both a source and destination path.")
		}
	},
}

func init() {
	// Add the mv command to the root command
	rootCmd.AddCommand(mvCmd)
}
