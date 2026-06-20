package cmd

import (
	"log"

	"github.com/erdos-ai/r2/pkg"

	"github.com/spf13/cobra"
)

// cpCmd represents the cp command
var cpCmd = &cobra.Command{
	Use:   "cp",
	Short: "Copy an object from one R2 path to another",
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
				// Copy local file to R2
				destURI := pkg.ParseR2URISafe(destinationPath)
				b := c.Bucket(destURI.Bucket)
				b.Upload(sourcePath, destURI.Path)
			} else if pkg.IsR2URI(sourcePath) && !pkg.IsR2URI(destinationPath) {
				// Copy R2 object to local file
				sourceURI := pkg.ParseR2URISafe(sourcePath)
				b := c.Bucket(sourceURI.Bucket)
				b.Download(sourceURI.Path, destinationPath)
			} else if pkg.IsR2URI(sourcePath) && pkg.IsR2URI(destinationPath) {
				// Copy R2 object to R2 object
				sourceURI := pkg.ParseR2URISafe(sourcePath)
				destURI := pkg.ParseR2URISafe(destinationPath)
				b := c.Bucket(sourceURI.Bucket)
				b.Copy(sourceURI.Path, destURI)
			}
		} else {
			log.Fatal("Please provide both a source and destination path.")
		}
	},
}

func init() {
	// Add the cp command to the root command
	rootCmd.AddCommand(cpCmd)
}
