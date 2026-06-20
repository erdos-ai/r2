# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

R2 is a CLI and Go library for working with Cloudflare's R2 Storage service. It implements AWS S3-compatible commands to provide simple-to-use features similar to the AWS CLI's s3 subcommand. The project is written in Go and uses the Cobra framework for CLI commands.

## Development Commands

### Building the Project
```bash
# Build the CLI
go build -o r2 main.go

# Build with version information
go build -ldflags "-s -w -X github.com/erdos-ai/r2/cmd.version=v1.0.0" -o r2 main.go
```

### Installing Dependencies
```bash
go mod tidy
go mod download
```

### Running the CLI
```bash
# Run directly without building
go run main.go [command] [flags]

# After building
./r2 [command] [flags]
```

### Release Process
The project uses release-please for automated versioning and GoReleaser for building releases.

**Automated Process:**
1. Push commits to main using Conventional Commits format:
   - `feat:` for new features (minor version bump)
   - `fix:` for bug fixes (patch version bump)
   - `feat!:` or `BREAKING CHANGE:` for breaking changes (major version bump)
2. release-please automatically creates/updates a Release PR
3. Review and merge the Release PR
4. release-please creates a tag, triggering GoReleaser to build and publish

**Conventional Commit Examples:**
```bash
git commit -m "feat: add bucket encryption support"
git commit -m "fix: resolve timeout issue in sync command"
git commit -m "feat!: change configuration file format"
```

**Manual release (not recommended):**
```bash
# Only use if automated process fails
git tag v0.1.0
git push origin v0.1.0
```

## Architecture

### Package Structure
- **`/cmd`**: CLI command implementations using Cobra framework
  - Each command (configure, cp, ls, mb, mv, presign, rb, rm, sync) has its own file
  - `root.go` contains the base command and global flags
  
- **`/pkg`**: Core library for R2 operations
  - `client.go`: R2 client configuration and bucket-level operations
  - `bucket.go`: Object-level operations within buckets
  - `helpers.go`: Utility functions

### Key Dependencies
- `github.com/spf13/cobra`: CLI framework
- `github.com/aws/aws-sdk-go-v2`: AWS SDK for S3-compatible operations
- Go 1.24+ required

### Command Pattern
All CLI commands follow the same pattern:
1. Parse flags and arguments
2. Create R2 client using profile configuration
3. Execute the operation using the pkg library
4. Handle errors and output results

### Configuration
The CLI uses profiles stored locally for authentication:
- Default profile: "default"
- Profile flag: `-p, --profile` available globally
- Configuration includes: AccountID, AccessKeyID, SecretAccessKey

## Important Notes

- The project implements S3-compatible operations for Cloudflare R2
- Version information is injected at build time via ldflags
- GoReleaser handles cross-platform builds for Linux and Darwin
- The install script is automatically updated after each release