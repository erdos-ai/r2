<p align="center">
  <a href="https://github.com/erdos-one/r2">
    <img alt="R2 CLI" src="assets/bucket.svg" width="150"/>
  </a>
</p>

<h1 align="center">
  Cloudflare R2 object storage made easy
</h1>

<p align="center">
  <a href="https://github.com/erdos-one/r2/releases/latest" title="GitHub release">
    <img src="https://img.shields.io/github/release/erdos-one/r2.svg">
  </a>
  <a href="https://opensource.org/licenses/Apache-2.0" title="License: Apache-2.0">
    <img src="https://img.shields.io/badge/License-Apache%202-blue.svg">
  </a>
</p>

## Purpose

`r2` is a library and command line interface for working with Cloudflare's
[R2 Storage](https://www.cloudflare.com/products/r2/).

Cloudflare's R2 implements the [S3
API](https://docs.aws.amazon.com/AmazonS3/latest/API/Welcome.html), attempting to allow users and
their applications to migrate easily, but importantly lacks the key, simple-to-use features provided
by the AWS CLI's [s3 subcommand](https://docs.aws.amazon.com/cli/latest/reference/s3/), as opposed
to the more complex and verbose API calls of the [s3api
subcommand](https://docs.aws.amazon.com/cli/latest/reference/s3api/index.html). This CLI fills that
gap.

## Installation

To install the `r2` CLI, simply run the following command:

```bash
go install github.com/erdos-one/r2@latest
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
- `r2 ls` — List either all buckets or all objects in a bucket
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

# Use a custom config file
r2 --config /path/to/r2.ini sync /d/ r2://backup/2026-02-02/
```

Example `patterns.txt`:

```text
# Only include db dumps
db-dump/
**/*.sql
```

For more usage information — including library usage — see [USAGE.md](USAGE.md).

## Progress

We're working to implement all the functionality of the AWS CLI's s3 subcommand. As of
[v0.1.0-alpha](https://github.com/erdos-one/r2/tree/v0.1.0-alpha), we have all the commands
implemented, but not all the options. We're working on it, but if you'd like to lend a helping hand,
we'd much appreciate your help!

To view the latest changes, see [CHANGELOG.md](CHANGELOG.md).

## Contributing

Our expected workflow is: Fork → Patch → Push → Pull Request.

Another helpful way to contribute is to report bugs or request features by opening an issue. We
appreciate contributions of all kinds!

To understand the codebase, we recommend reading the [ARCHITECTURE.md](ARCHITECTURE.md) file.
