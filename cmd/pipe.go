package cmd

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/erdos-ai/r2/pkg"
	"github.com/spf13/cobra"
)

// pipeCmd represents the pipe command
var pipeCmd = &cobra.Command{
	Use:   "pipe TARGET",
	Short: "Stream data from stdin to an R2 object",
	Long: `Stream data from stdin directly to an R2 object without intermediate storage.

The pipe command reads from standard input and uploads the stream directly to 
the specified R2 location. This is useful for backup scripts, data pipelines,
and situations where you want to avoid creating temporary files.

Examples:
  # Stream text to R2
  echo "Hello World" | r2 pipe r2://bucket/hello.txt

  # Backup a database
  mysqldump mydb | r2 pipe r2://backups/db-backup.sql

  # Compress and upload a directory
  tar czf - /path/to/dir | r2 pipe r2://bucket/archive.tar.gz

  # Stream from a file
  cat large-file.bin | r2 pipe r2://bucket/large-file.bin`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get the target path
		target := args[0]

		// Validate that target is an R2 URI
		if !pkg.IsR2URI(target) {
			log.Fatal("Target must be an R2 URI (e.g., r2://bucket/path)")
		}

		// Parse the R2 URI
		uri := pkg.ParseR2URISafe(target)

		// Check if stdin is a terminal (no piped input)
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			fmt.Println("Error: No data provided on stdin")
			fmt.Println("Usage: <command> | r2 pipe r2://bucket/path")
			os.Exit(1)
		}

		// Get profile client
		profileName, err := cmd.Flags().GetString("profile")
		if err != nil {
			log.Fatal(err)
		}
		c := pkg.Client(getProfile(profileName))
		b := c.Bucket(uri.Bucket)

		// Get optional flags
		partSize, err := cmd.Flags().GetInt64("part-size")
		if err != nil {
			log.Fatal(err)
		}
		concurrency, err := cmd.Flags().GetInt("concurrency")
		if err != nil {
			log.Fatal(err)
		}
		quiet, err := cmd.Flags().GetBool("quiet")
		if err != nil {
			log.Fatal(err)
		}

		// Create a reader from stdin
		reader := io.Reader(os.Stdin)

		// Upload the stream
		if !quiet {
			fmt.Printf("Streaming to r2://%s/%s...\n", uri.Bucket, uri.Path)
		}

		err = b.PutStream(reader, uri.Path, partSize, concurrency)
		if err != nil {
			log.Fatalf("Failed to stream to r2://%s/%s: %v\n", uri.Bucket, uri.Path, err)
		}

		if !quiet {
			fmt.Printf("Successfully streamed to r2://%s/%s\n", uri.Bucket, uri.Path)
		}
	},
}

func init() {
	// Add the pipe command to the root command
	rootCmd.AddCommand(pipeCmd)

	// Add optional flags
	pipeCmd.Flags().Int64("part-size", 5*1024*1024, "Part size for multipart upload in bytes (minimum 5MB, default 5MB)")
	pipeCmd.Flags().Int("concurrency", 5, "Number of concurrent upload threads")
	pipeCmd.Flags().BoolP("quiet", "q", false, "Suppress progress output")
}
