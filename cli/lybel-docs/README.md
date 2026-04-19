# lybel-docs

Convert extended Markdown (with Confluence macro extensions) to Atlassian
Document Format (ADF) JSON. Designed for piping into the Confluence REST API.

## Install

```bash
make install   # builds and copies binary to ../../skills/lybel-docs/bin/lybel-docs
make build     # builds bin/lybel-docs only
```

Requires Go 1.21+.

## Usage

```bash
# stdin -> stdout
lybel-docs adf < page.md > page.adf.json
echo "# Hello" | lybel-docs adf

# file input
lybel-docs adf --file page.md
lybel-docs adf -f page.md

# pretty-printed JSON
lybel-docs adf --pretty -f page.md

# version / help
lybel-docs --version
lybel-docs --help
```

Exit codes: `0` success, `1` parse error, `2` invalid input, `3` unknown error.

## Supported Markdown

CommonMark plus GitHub-Flavoured tables, strikethrough, and these Confluence
extensions:

### Table of Contents

```
[TOC]
[TOC maxLevel=3 minLevel=1]
```

Renders as the Confluence `toc` extension macro.

### Expand block

```
:::expand Click to reveal
Anything markdown-ish goes here.

- including lists
- tables, code blocks, etc.
:::
```

### Panel blocks

```
:::warning Heads up
Be careful.
:::

:::info
Just FYI.
:::
```

Supported panel types: `info`, `warning`, `note`, `success`, `error`. The title
after the panel keyword is optional and renders as a bold first paragraph
inside the panel.

## Development

```bash
make test            # go test -v ./...
make build           # native build
make build-all       # darwin/linux/windows × amd64/arm64 into dist/
make fmt             # go fmt ./...
make lint            # go vet ./...
```

The version string is injected at build time via `-ldflags "-X main.version=..."`.
By default `make` reads `git describe --tags --always --dirty`; override with
`make build VERSION=1.2.3`.

## Project layout

```
adf/builder.go    ADF node + mark types and constructor helpers
adf/converter.go  goldmark AST -> ADF walker
adf/macros.go     Pre-processing for [TOC] and ::: container blocks
adf/converter_test.go  Table-driven unit tests
main.go           CLI entry, flag parsing, IO plumbing
```

## Contributing

Keep dependencies minimal — currently `github.com/yuin/goldmark` plus stdlib.
Add a test in `adf/converter_test.go` for any new markdown construct or macro.
Run `make test` and `make lint` before opening a PR.
