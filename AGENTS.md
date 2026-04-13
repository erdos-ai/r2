# Repository Guidelines

This repository hosts `r2`, a Go (1.24) CLI and library for Cloudflare R2 built on Cobra and the AWS S3 SDK. Use this guide to contribute changes consistently and safely.

## Project Structure & Module Organization
- `cmd/`: Cobra commands (one file per subcommand, e.g. `cp.go`, `ls.go`, `sync.go`; root in `root.go`).
- `pkg/`: Library for R2 operations (`client.go`, `bucket.go`, `helpers.go`). Prefer adding logic here first.
- `main.go`: CLI entry point.
- `assets/`: Static assets (e.g., `bucket.svg`).
- `.github/workflows/`: Release/automation.
- Docs: `README.md`, `USAGE.md`, `ARCHITECTURE.md`.

## Build, Test, and Development Commands
- Build: `go build -o bin/r2 .` (module root), or `go build ./...` to verify packages.
- Run locally: `go run .` or after install `r2 help`.
- Tests: `go test ./...` (add package tests under `pkg/`).
- Format/Vet: `go fmt ./... && go vet ./...`.
- Release (dry run): `goreleaser release --snapshot --skip-publish --clean`.

## Coding Style & Naming Conventions
- Formatting: enforce `go fmt`; keep imports tidy (`go mod tidy` before PRs).
- Structure: business logic in `pkg`, thin Cobra layers in `cmd` for flags/UX.
- Naming: exported types/methods in `pkg` use Go conventions (e.g., `R2Client`, `R2Bucket.Sync...`); `cmd/<name>.go` for subcommands.
- Errors: return errors from `pkg`; CLI may `log.Fatal` on unrecoverable user errors.

## Testing Guidelines
- Framework: standard `testing` with table‑driven tests.
- Location: co‑locate `*_test.go` with sources (focus on `pkg`).
- Run/Coverage: `go test -v -cover ./pkg/...`.
- Isolation: avoid real R2 calls in tests; mock/stub S3 interactions or factor logic for unit testing.

## Commit & Pull Request Guidelines
- Commits: follow Conventional Commits (`feat:`, `fix:`, `docs:`, `refactor:`), as in this repo’s history.
- PRs must include: clear description, linked issue, CLI examples (e.g., `r2 cp r2://bucket/a b`), and doc updates when UX changes.
- Checks: ensure `go fmt`, `go vet`, `go test`, and `go mod tidy` pass; avoid diff in `go.sum` unless necessary.

## Security & Configuration Tips
- Never commit credentials; `r2 configure` stores profiles in `~/.r2`.
- Redact secrets in logs/output; validate bucket names and paths (see `pkg/helpers.go`).
