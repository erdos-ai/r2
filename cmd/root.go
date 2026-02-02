package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Store version information
var version string = "unset"

// rootCmd represents the base command when called without any commands
var rootCmd = &cobra.Command{
	Use:   "r2",
	Short: "Command Line Interface for Cloudflare R2 Storage",
	Long: `r2 is a command line interface for working with Cloudflare's R2 Storage.

Cloudflare's R2 implements the S3 API, attempting to allow users and their
applications to migrate easily, but importantly lacks the key, simple-to-use
features provided by the AWS CLI's s3 subcommand, as opposed to the more complex
and verbose API calls of the s3api subcommand. This CLI fills that gap.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If the version flag is set, print version information and quit
		if v, _ := cmd.Flags().GetBool("version"); v {
			cmd.Println(version)
			return
		}

		// If no subcommand is provided, print help and quit
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Enable profile flag for all commands
	rootCmd.PersistentFlags().StringP("profile", "p", "default", "R2 profile to use")
	rootCmd.PersistentFlags().String("config", "", "Path to R2 config file (default \"~/.r2\")")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		configPath, err := cmd.Root().PersistentFlags().GetString("config")
		if err != nil {
			return err
		}
		return applyConfigFlag(configPath)
	}

	// Add version flag
	rootCmd.Flags().BoolP("version", "v", false, "Print version information and quit")
}
