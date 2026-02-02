# Changelog

This changelog goes through all the changes that have been made in each release.

## [0.3.0](https://github.com/erdos-ai/r2/compare/v0.2.1...v0.3.0) (2026-02-02)


### Features

* add --include-from and --config flags for sync command ([288c6b8](https://github.com/erdos-ai/r2/commit/288c6b875ed4bce0661a28e21db2bbde391c62d8)), closes [#17](https://github.com/erdos-ai/r2/issues/17)


### Bug Fixes

* handle uppercase uname output ([766f168](https://github.com/erdos-ai/r2/commit/766f168b8c8e5dae8786525194785c588f228ab6))

## v0.1.3-alpha

- FIXED
  - [`GetObjects` function](pkg/bucket.go) — now properly paginates through S3 API responses to retrieve all objects from buckets with more than 1000 objects
  - `sync` command — fixed for large buckets (>1000 objects) that previously would only sync the first 1000 objects
  - `ls` command — now shows complete listings for buckets with more than 1000 objects

## v0.1.2-alpha

- FIXED
  - [`ls` command](cmd/ls.go) — now requires at least one bucket argument and displays usage information when run without arguments instead of attempting to list all buckets

## v0.1.1-alpha

- FIXED
  - GoReleaser configuration — updated for v2 compatibility
  - Install script — improved handling of unchanged script updates

## v0.1.0-alpha

First release of `r2`! This release includes all the commands of the AWS CLI's `s3` subcommand, but
not all the options.

- ADDED
  - [`configure` command](cmd/configure.go) — configure `r2` to use your R2 credentials
  - [`cp` command](cmd/cp.go) — copy objects and directories
  - [`ls` command](cmd/ls.go) — list objects and directories
  - [`mb` command](cmd/mb.go) — make buckets
  - [`mv` command](cmd/mv.go) — move objects and directories
  - [`presign` command](cmd/presign.go) — generate pre-signed URLs
  - [`rb` command](cmd/rb.go) — remove buckets
  - [`rm` command](cmd/rm.go) — remove objects and directories
  - [`sync` command](cmd/sync.go) — synchronize objects and directories
