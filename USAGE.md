# Usage

`r2` is a library and command line interface for working with Cloudflare's R2 Storage.

## CLI

```bash
r2 [command] [flags]
```

### Available Commands

- `configure` — Configure R2 access
- `cp` — Copy an object from one R2 path to another
- `help` — Help about any command
- `ls` — List either all buckets or all objects in a bucket
- `mb` — Create an R2 bucket
- `mv` — Moves a local file or R2 object to another location locally or in R2.
- `pipe` — Stream data from stdin to an R2 object
- `presign` — Generate a pre-signed URL for a Cloudflare R2 object
- `rb` — Remove an R2 bucket
- `rm` — Remove an object from an R2 bucket
- `sync` — Syncs directories and R2 prefixes.

### Global Flags

- `-p, --profile` — R2 profile to use (default "default")
- `--config` — Path to R2 config file (default "~/.r2")
- `-h, --help` — Help for any command

### Help

Help for any command can be obtained by running `r2 help [command]`. For example:

```bash
# Help for the configure command
r2 help configure
```

### Sync Command

The `sync` command copies changed files between a local directory and an R2 prefix.

#### Flags

- `--include-from` — Read include patterns from a file (patterns are relative to the sync source root)
- `--exclude-from` — Read exclude patterns from a file (same format as `--include-from`)
- `--exclude` — Exclude files matching a glob pattern; may be repeated

#### Pattern File Format

The include and exclude files share one format. Lines are glob patterns. A `#` starts a comment,
either on its own line or inline after whitespace (e.g. `*.sql # database dumps`); write `\#` to
match a literal `#`. `**` matches across directories (a bare `*` does not). A trailing `/` means
"everything under this directory".

Patterns are relative to the sync source root. For example, use `db-dump/` not `/d/db-dump/`.

Example `patterns.txt`:

```text
# Only include db dumps and SQL files
db-dump/
**/*.sql   # all SQL files, at any depth
```

#### Filter Precedence

A file is synced only if it matches the include filter (when one is given) **and** does not
match the exclude filter (when one is given). Exclude wins on conflict. With no include flag,
everything is included; with no exclude flag, nothing is excluded.

An empty include file is an error (it would sync nothing), while an empty exclude file or an
empty `--exclude` value is a harmless no-op.

Example usage:

```bash
# Include-only sync
r2 sync --include-from patterns.txt /d/ r2://backup/2026-02-02/

# Exclude build artifacts and secrets, both inline and from a file
r2 sync --exclude 'node_modules/**' --exclude '**/*.tmp' --exclude-from .syncignore /d/ r2://backup/2026-02-02/
```

### Pipe Command

The `pipe` command allows you to stream data from stdin directly to R2 without creating temporary files. This is useful for backup scripts, data pipelines, and real-time data processing. Note: Data is buffered in memory during upload.

#### Basic Usage

```bash
<command> | r2 pipe r2://bucket/path
```

#### Examples

```bash
# Stream text to R2
echo "Hello World" | r2 pipe r2://bucket/hello.txt

# Backup a database directly to R2
mysqldump mydb | r2 pipe r2://backups/db-backup.sql

# Compress and upload a directory
tar czf - /path/to/dir | r2 pipe r2://bucket/archive.tar.gz

# Stream from a file
cat large-file.bin | r2 pipe r2://bucket/large-file.bin

# Use with quiet mode
echo "data" | r2 pipe r2://bucket/file.txt --quiet
```

#### Flags

- `--part-size` — Part size for multipart upload in bytes (minimum 5MB, default 5MB)
- `--concurrency` — Number of concurrent upload threads (default 5)
- `-q, --quiet` — Suppress progress output

## Library

The `r2` library can be used to interact with R2 from within your Go application. All library code
exists in the [pkg](pkg) directory and is well documented.

Documentation may be found [here](https://pkg.go.dev/github.com/erdos-one/r2/pkg).

### Example

Uploading a file to a bucket:

```go
package main

import (
  r2 "github.com/erdos-one/r2/pkg"
)

func main() {
  // Create client
  config := r2.Config{
    Profile:         "default",
    AccountID:       "<ACCOUNT ID>",
    AccessKeyID:     "<ACCESS KEY ID>",
    SecretAccessKey: "<SECRET ACCESS KEY>"
  }
  client := r2.Client(config)

  // Connect to bucket
  bucket := client.Bucket("my-bucket")

  // Upload a file to the bucket
  bucket.Upload("my-local-file.txt", "my-remote-file.txt")
}
```
