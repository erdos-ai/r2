package cmd

import (
	"log"

	"github.com/erdos-ai/r2/pkg"

	"github.com/spf13/cobra"
)

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls [bucket-name]",
	Short: "List objects in a bucket",
	Long: `List objects in an R2 bucket.

To list objects in a bucket, provide the bucket name as an argument.

Examples:
  # List all objects in a bucket
  r2 ls example-bucket

  # List objects in multiple buckets
  r2 ls bucket1 bucket2

  # List objects using R2 URI format
  r2 ls r2://example-bucket`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get profile client
		profileName, err := cmd.Flags().GetString("profile")
		if err != nil {
			log.Fatal(err)
		}
		c := pkg.Client(getProfile(profileName))

		// List objects in each bucket passed
		for _, bucketName := range args {
			// Remove URI scheme if present
			bucketName = pkg.RemoveR2URIPrefix(bucketName)

			b := c.Bucket(bucketName)
			b.PrintObjects()
		}
	},
}

func init() {
	// Add the ls command to the root command
	rootCmd.AddCommand(lsCmd)
}
