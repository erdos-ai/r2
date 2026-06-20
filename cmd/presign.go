package cmd

import (
	"fmt"
	"log"

	"github.com/erdos-ai/r2/pkg"

	"github.com/spf13/cobra"
)

// presignCmd represents the presign command
var presignCmd = &cobra.Command{
	Use:   "presign",
	Short: "Generate a pre-signed URL for a Cloudflare R2 object",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Get profile client
		profileName, err := cmd.Flags().GetString("profile")
		if err != nil {
			log.Fatal(err)
		}
		c := pkg.Client(getProfile(profileName))
		pc := pkg.PresignClient(getProfile(profileName))

		for _, arg := range args {
			// Get R2 URI components from argument
			uri := pkg.ParseR2URISafe(arg)

			// If object exists in bucket, print presigned URL to get object from bucket, otherwise print
			// presigned URL to put object in bucket
			b := c.Bucket(uri.Bucket)
			if pkg.Contains(b.GetObjectPaths(), uri.Path) {
				fmt.Println(pc.GetURL(uri))
			} else {
				fmt.Println(pc.PutURL(uri))
			}
		}
	},
}

func init() {
	// Add the presign command to the root command
	rootCmd.AddCommand(presignCmd)
}
