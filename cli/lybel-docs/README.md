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

Two subcommands:

- **`adf`** — convert markdown to a fresh ADF document (for `createConfluencePage`).
- **`edit`** — apply a section-level edit to an existing ADF document (for `updateConfluencePage` without destroying macros).

### `adf` — markdown to ADF

```bash
# stdin -> stdout
lybel-docs adf < page.md > page.adf.json
echo "# Hello" | lybel-docs adf

# file input
lybel-docs adf --file page.md
lybel-docs adf -f page.md

# pretty-printed JSON
lybel-docs adf --pretty -f page.md
```

### `edit` — section-level edits on an existing ADF doc

Reads current ADF (stdin or `--input`), applies one operation, writes new ADF
to stdout. Macros outside the touched section are preserved verbatim.

```bash
# Append a fragment to the end
lybel-docs edit -i current.json --append fragment.md > new.json

# Replace a section matched by heading text
lybel-docs edit -i current.json --replace-section "📇 Page ID Index" fragment.md > new.json

# Insert right after / before a section
lybel-docs edit -i current.json --insert-after "🔍 Research" fragment.md > new.json
lybel-docs edit -i current.json --insert-before "🤖 Uso com IA" fragment.md > new.json

# Delete a section (heading + its body, no fragment needed)
lybel-docs edit -i current.json --delete-section "TODO antigo" > new.json
```

Section = matched heading + all following top-level nodes up to (but not
including) the next heading of equal-or-higher level. Headings are matched by
exact case-sensitive text (trimmed); first match wins.

```bash
# version / help
lybel-docs --version
lybel-docs --help
```

Exit codes: `0` success, `1` parse error, `2` invalid input (incl. section not
found), `3` unknown error.

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
adf/builder.go         ADF node + mark types and constructor helpers
adf/converter.go       goldmark AST -> ADF walker
adf/macros.go          Pre-processing for [TOC] and ::: container blocks
adf/edit.go            Section-level edit ops (append/insert/replace/delete)
adf/converter_test.go  Tests for markdown -> ADF
adf/edit_test.go       Tests for section-level edits
main.go                CLI entry, flag parsing, IO plumbing
```

## Contributing

Keep dependencies minimal — currently `github.com/yuin/goldmark` plus stdlib.
Add a test in `adf/converter_test.go` for any new markdown construct or macro.
Run `make test` and `make lint` before opening a PR.
