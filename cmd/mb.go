package cmd

import (
	"fmt"
	"log"

	"github.com/erdos-ai/r2/pkg"

	"github.com/spf13/cobra"
)

// mbCmd represents the mb command
var mbCmd = &cobra.Command{
	Use:   "mb",
	Short: "Create an R2 bucket",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Get profile client
		profileName, err := cmd.Flags().GetString("profile")
		if err != nil {
			log.Fatal(err)
		}
		c := pkg.Client(getProfile(profileName))

		// If a bucket name is provided, create the bucket
		if len(args) > 0 {
			bucketName := args[0]
			c.MakeBucket(bucketName)
		} else {
			fmt.Println("Please provide a bucket name")
		}
	},
}

func init() {
	// Add the mb command to the root command
	rootCmd.AddCommand(mbCmd)
}
