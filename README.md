<p align="center">
  <a href="https://github.com/erdos-ai/r2">
    <img alt="R2 CLI" src="assets/bucket.svg" width="150"/>
  </a>
</p>

<h1 align="center">
  Cloudflare R2 object storage made easy
</h1>

<p align="center">
  <a href="https://github.com/erdos-ai/r2/releases/latest" title="GitHub release">
    <img src="https://img.shields.io/github/release/erdos-ai/r2.svg">
  </a>
  <a href="https://opensource.org/licenses/Apache-2.0" title="License: Apache-2.0">
    <img src="https://img.shields.io/badge/License-Apache%202-blue.svg">
  </a>
</p>

> [!WARNING]
> **`r2` is no longer maintained, and this repository is archived.** v0.4.2 is the final release. It
> still installs and works, but it won't get fixes, including for security issues or changes to R2.
> Use [rclone](https://developers.cloudflare.com/r2/examples/rclone/) or the
> [AWS CLI](https://developers.cloudflare.com/r2/examples/aws/aws-cli/) instead. See
> [Migrating off r2](#migrating-off-r2) for equivalent commands.

## Purpose

`r2` is a library and command line interface for working with Cloudflare's
[R2 Storage](https://www.cloudflare.com/products/r2/).

It was started in January 2023, when R2 had no simple command line tool for everyday object
operations like the AWS CLI's [s3 subcommand](https://docs.aws.amazon.com/cli/latest/reference/s3/).
General-purpose S3 tools now work well with R2 and do more than `r2` does, including parallel
multipart transfers, files larger than 5 GiB, dry runs, and buckets in a jurisdiction. Cloudflare's
[Wrangler](https://developers.cloudflare.com/r2/get-started/cli/) covers R2-specific bucket settings.

## Migrating off r2

Your existing credentials are in `~/.r2`, one section per profile.

### rclone

Add a remote to `~/.config/rclone/rclone.conf`, or create one with `rclone config`:

```ini
[r2]
type = s3
provider = Cloudflare
access_key_id = <ACCESS_KEY_ID>
secret_access_key = <SECRET_ACCESS_KEY>
endpoint = https://<ACCOUNT_ID>.r2.cloudflarestorage.com
acl = private
```

If your API token only has object-level permissions, also add `no_check_bucket = true`.

### AWS CLI

Run `aws configure --profile r2` with your access key ID, secret access key, and a region of `auto`.
Then add the endpoint to that profile in `~/.aws/config`:

```ini
[profile r2]
region = auto
endpoint_url = https://<ACCOUNT_ID>.r2.cloudflarestorage.com
```

Pass `--profile r2` to each command, or set `AWS_PROFILE=r2`.

### Equivalent commands

| `r2` | rclone | AWS CLI |
| --- | --- | --- |
| `r2 ls my-bucket` | `rclone ls r2:my-bucket` | `aws s3 ls s3://my-bucket --recursive` |
| `r2 cp file.txt r2://my-bucket/file.txt` | `rclone copyto file.txt r2:my-bucket/file.txt` | `aws s3 cp file.txt s3://my-bucket/file.txt` |
| `r2 mv r2://my-bucket/a.txt a.txt` | `rclone moveto r2:my-bucket/a.txt a.txt` | `aws s3 mv s3://my-bucket/a.txt a.txt` |
| `r2 rm r2://my-bucket/file.txt` | `rclone deletefile r2:my-bucket/file.txt` | `aws s3 rm s3://my-bucket/file.txt` |
| `r2 mb my-bucket` | `rclone mkdir r2:my-bucket` | `aws s3 mb s3://my-bucket` |
| `r2 rb my-bucket` | `rclone rmdir r2:my-bucket` | `aws s3 rb s3://my-bucket` |
| `r2 sync ./dir r2://my-bucket/backup` | `rclone copy ./dir r2:my-bucket/backup` | `aws s3 sync ./dir s3://my-bucket/backup` |
| `cmd \| r2 pipe r2://my-bucket/out.gz` | `cmd \| rclone rcat r2:my-bucket/out.gz` | `cmd \| aws s3 cp - s3://my-bucket/out.gz` |
| `r2 presign r2://my-bucket/file.txt` | `rclone link --expire 15m r2:my-bucket/file.txt` | `aws s3 presign s3://my-bucket/file.txt --expires-in 900` |

A few differences to watch for:

- **Sync never deleted.** `r2 sync` never removed anything at the destination. `rclone sync` does, so
  use `rclone copy` to keep the same behavior. `aws s3 sync` only deletes when you pass `--delete`.
- **Sync filters.** Each tool has its own pattern syntax. Check
  [rclone's filtering rules](https://rclone.org/filtering/) or the AWS CLI's
  [`--exclude` and `--include` filters](https://docs.aws.amazon.com/cli/latest/reference/s3/index.html#use-of-exclude-and-include-filters)
  before you migrate a filtered sync.
- **Upload URLs.** For a key that didn't exist yet, `r2 presign` printed a presigned upload (PUT)
  URL. `rclone link` and `aws s3 presign` only create download (GET) URLs. Generate upload URLs with
  an SDK, as shown in [Cloudflare's presigned URL docs](https://developers.cloudflare.com/r2/api/s3/presigned-urls/).
- **Library.** The `pkg` library is a thin wrapper around the AWS SDK for Go v2. Use the SDK directly,
  as in Cloudflare's [Go example](https://developers.cloudflare.com/r2/examples/aws/aws-sdk-go/).

## Known limitations of v0.4.2

- `cp`, `mv`, and `sync` upload each file in a single request, so files over R2's single-upload limit
  of about 5 GiB fail. Only `pipe` uses multipart uploads.
- `sync` compares local MD5 hashes to object ETags. Objects uploaded in parts, for example by
  `r2 pipe`, never match, so `sync` transfers them again every time.
- The endpoint is always `https://<ACCOUNT_ID>.r2.cloudflarestorage.com`, so buckets in a
  jurisdiction, such as the EU, aren't reachable.
- `ls` lists every object in a bucket and doesn't take a prefix.
- Presigned URLs expire after 15 minutes.

## Installation

To install the final release of the `r2` CLI, run:

```bash
go install github.com/erdos-ai/r2@latest
```

For more installation options, see [INSTALL.md](INSTALL.md).

## Usage

To view the CLI's help message, run:

```bash
r2 help
```

### Available Commands

- `r2 configure` — Configure R2 access
- `r2 cp` — Copy an object from one R2 path to another
- `r2 help` — Help about any command
- `r2 ls` — List all objects in a bucket
- `r2 mb` — Create an R2 bucket
- `r2 mv` — Moves a local file or R2 object to another location locally or in R2.
- `r2 pipe` — Stream data from stdin to an R2 object
- `r2 presign` — Generate a pre-signed URL for a Cloudflare R2 object
- `r2 rb` — Remove an R2 bucket
- `r2 rm` — Remove an object from an R2 bucket
- `r2 sync` — Syncs directories and R2 prefixes.

To view the help message for a specific command, run:

```bash
r2 help <command>
```

### Sync Examples

```bash
# Include-only sync from local to R2
r2 sync --include-from patterns.txt /d/ r2://backup/2026-02-02/

# Exclude files matching a glob pattern (repeatable)
r2 sync --exclude 'node_modules/**' --exclude '**/*.tmp' /d/ r2://backup/2026-02-02/

# Read exclude patterns from a file
r2 sync --exclude-from .syncignore /d/ r2://backup/2026-02-02/

# Use a custom config file
r2 --config /path/to/r2.ini sync /d/ r2://backup/2026-02-02/
```

Example `patterns.txt`:

```text
# Only include db dumps
db-dump/
**/*.sql   # all SQL files, at any depth
```

For more usage information — including library usage — see [USAGE.md](USAGE.md).

## Changelog

To view the history of changes, see [CHANGELOG.md](CHANGELOG.md). To understand the codebase, see
[ARCHITECTURE.md](ARCHITECTURE.md).

## Forking

This repository is archived, so it no longer accepts issues or pull requests. `r2` is licensed under
[Apache-2.0](LICENSE), and you're welcome to fork it. Please publish a fork under your own module
path.

Thank you to everyone who starred the project, filed issues, and asked for features.
