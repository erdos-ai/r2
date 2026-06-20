package cmd

import (
	"log"

	"github.com/erdos-ai/r2/pkg"

	"github.com/spf13/cobra"
)

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove an object from an R2 bucket",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Get profile client
		profile, err := cmd.Flags().GetString("profile")
		if err != nil {
			log.Fatal(err)
		}
		c := pkg.Client(getProfile(profile))

		// If a bucket name is provided, create the bucket
		for _, arg := range args {
			if pkg.IsR2URI(arg) {
				uri := pkg.ParseR2URISafe(arg)
				b := c.Bucket(uri.Bucket)
				b.Delete(uri.Path)
			} else {
				log.Fatalf("Path %s is not a valid R2 URI", arg)
			}
		}
	},
}

func init() {
	// Add the rm command to the root command
	rootCmd.AddCommand(rmCmd)
}
