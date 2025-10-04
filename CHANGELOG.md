# Changelog

This changelog goes through all the changes that have been made in each release.

## [0.3.0](https://github.com/erdos-one/r2/compare/v0.2.1...v0.3.0) (2025-10-04)


### Features

* implement pipe command for streaming uploads ([13f9760](https://github.com/erdos-one/r2/commit/13f9760f6198cb59716caa59dc349aad1c21277a))
* implement pipe command for streaming uploads ([3ca7847](https://github.com/erdos-one/r2/commit/3ca7847d52eb9f68ac6f4b808aefdca8d28a64d4)), closes [#1](https://github.com/erdos-one/r2/issues/1)
* upgrade all dependencies to latest versions and fix R2 endpoint configuration ([63edb6a](https://github.com/erdos-one/r2/commit/63edb6a008e3a8911586cbb70bc478f5fe3527e1))


### Bug Fixes

* correct error message in RemoveBucket function ([e5b3887](https://github.com/erdos-one/r2/commit/e5b3887ed973b5f2ef4e37e46210641048080a14))
* handle uppercase uname output ([766f168](https://github.com/erdos-one/r2/commit/766f168b8c8e5dae8786525194785c588f228ab6))
* implement pagination for GetObjects to handle buckets with &gt;1000 objects ([8a1bd1a](https://github.com/erdos-one/r2/commit/8a1bd1a8542b8d7a628987833827081fbe167096)), closes [#3](https://github.com/erdos-one/r2/issues/3)
* implement pagination for GetObjects to handle large buckets ([b56c9c8](https://github.com/erdos-one/r2/commit/b56c9c841ddfad3eda5d8db95d8233d5a5518eb6))
* improve configuration parsing regex to be more restrictive ([feed758](https://github.com/erdos-one/r2/commit/feed75895cd81e05db5593762edc36da4eb76d59))
* improve credential validation in writeConfig ([abd8e81](https://github.com/erdos-one/r2/commit/abd8e81efa785fead9bde73b749d16919216da4c))
* make install-script job handle unchanged install.sh gracefully ([3458186](https://github.com/erdos-one/r2/commit/34581861330c395e40721ebda327ac965b5d01e9))
* require bucket argument for ls command ([7b4b1d2](https://github.com/erdos-one/r2/commit/7b4b1d2efc5a0e4ad098086f67e75847929dfaad))
* require bucket argument for ls command ([6745659](https://github.com/erdos-one/r2/commit/67456591e9a9ceeefc64c42770d57bda8967da3f)), closes [#2](https://github.com/erdos-one/r2/issues/2)
* resolve SignatureDoesNotMatch error in config handling ([9c35b7f](https://github.com/erdos-one/r2/commit/9c35b7f3a40e159ac4f7fdd075d44cc72a6d0e23))
* resolve SignatureDoesNotMatch error in config handling ([75f43f2](https://github.com/erdos-one/r2/commit/75f43f286662046e03f8468a6acc9a5e88221010)), closes [#13](https://github.com/erdos-one/r2/issues/13)
* resolve sync command authentication errors and add subdirectory support ([68e0dbe](https://github.com/erdos-one/r2/commit/68e0dbe6ce657727e73d576aa5a84bd378c57b35))
* update GoReleaser flag for v2 compatibility ([f643f2c](https://github.com/erdos-one/r2/commit/f643f2c982acccd900fa82b2978d8b22fffa8ba5))


### Performance Improvements

* move regex compilation outside getConfig function ([6652121](https://github.com/erdos-one/r2/commit/6652121c2fc4abbed9ec8ca203479ef495f57ffa))


### Documentation

* add repository guidelines for contributing ([0085bf7](https://github.com/erdos-one/r2/commit/0085bf7a60cb39a405a49a02d24819eb059f70fa))
* update CHANGELOG.md with recent releases ([fc82d8b](https://github.com/erdos-one/r2/commit/fc82d8b00e5c49222bdfb33dc19d5dc77d940e85))

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
