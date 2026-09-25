# Installing r2

> [!WARNING]
> `r2` is no longer maintained, and v0.4.2 is the final release. These instructions still work. See
> [Migrating off r2](README.md#migrating-off-r2) for maintained alternatives.

## CLI

The zero-dependency CLI is available for Linux and macOS and can be installed with the following
command:

```bash
sh <(curl -fsSL https://raw.githubusercontent.com/erdos-ai/r2/main/install.sh)
```

Run as root, the script installs `r2` to `/usr/local/bin`; otherwise it prints the commands to finish
the installation. Releases before v0.4.2 installed to `/usr/bin`, and the script removes an older copy
there when run as root. It never replaces a different program named `r2`, such as radare2.

To uninstall, run:

```bash
curl -fsSL https://raw.githubusercontent.com/erdos-ai/r2/main/uninstall.sh | sudo sh
```

This removes the CLI from `/usr/local/bin` and `/usr/bin`, and leaves your credentials in `~/.r2`.

## Library

The library is available as a Go module and can be installed with the following command:

```bash
go get github.com/erdos-ai/r2/pkg
```
