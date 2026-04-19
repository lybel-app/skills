// lybel-docs converts extended markdown (with Confluence macros) into ADF
// JSON for use with the Atlassian Confluence REST API.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/lybel-app/skills/cli/lybel-docs/adf"
)

// version is injected at build time via -ldflags "-X main.version=..."
var version = "dev"

const helpText = `lybel-docs — convert extended markdown to Atlassian Document Format (ADF) JSON.

USAGE:
  lybel-docs adf [--file PATH] [--pretty]
  lybel-docs --version
  lybel-docs --help

COMMANDS:
  adf     Convert markdown (stdin or --file) to ADF JSON on stdout.

FLAGS:
  -f, --file PATH    Read markdown from PATH instead of stdin.
      --pretty       Pretty-print the JSON output.
  -v, --version      Print version and exit.
  -h, --help         Show this help and exit.

EXTENSIONS (in addition to CommonMark + GFM tables):
  [TOC]                              Confluence Table of Contents macro.
  [TOC maxLevel=3 minLevel=1]        With explicit min/max levels.
  :::expand Title                    Expand block; close with :::
  :::warning Title                   Panel of type warning/info/note/success/error.

EXAMPLES:
  lybel-docs adf < page.md > page.adf.json
  lybel-docs adf -f page.md --pretty
  echo "# Hello" | lybel-docs adf

EXIT CODES:
  0  success
  1  parse error (markdown -> ADF)
  2  invalid input (missing file, bad flags, etc.)
  3  unknown error
`

const (
	exitOK         = 0
	exitParseErr   = 1
	exitInputErr   = 2
	exitUnknownErr = 3
)

// errInvalidUsage is returned when CLI flag parsing detects bad input.
var errInvalidUsage = errors.New("invalid usage")

func main() {
	code, err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "lybel-docs:", err)
	}
	os.Exit(code)
}

// run is the testable entry point: it returns an exit code so unit tests can
// drive the CLI without calling os.Exit.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) (int, error) {
	// Top-level flags (--help / --version) take precedence over subcommands.
	if len(args) == 0 {
		fmt.Fprint(stderr, helpText)
		return exitInputErr, errInvalidUsage
	}

	switch args[0] {
	case "-h", "--help":
		fmt.Fprint(stdout, helpText)
		return exitOK, nil
	case "-v", "--version":
		fmt.Fprintln(stdout, "lybel-docs", version)
		return exitOK, nil
	case "adf":
		return runADF(args[1:], stdin, stdout, stderr)
	}

	fmt.Fprintln(stderr, "unknown command:", args[0])
	fmt.Fprint(stderr, helpText)
	return exitInputErr, errInvalidUsage
}

// runADF parses adf-subcommand flags and performs the conversion.
func runADF(args []string, stdin io.Reader, stdout, stderr io.Writer) (int, error) {
	var (
		file   string
		pretty bool
	)

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "-f", "--file":
			if i+1 >= len(args) {
				fmt.Fprintln(stderr, "flag", a, "requires a value")
				return exitInputErr, errInvalidUsage
			}
			file = args[i+1]
			i++
		case "--pretty":
			pretty = true
		case "-h", "--help":
			fmt.Fprint(stdout, helpText)
			return exitOK, nil
		default:
			// Allow --file=path style.
			if len(a) > 7 && a[:7] == "--file=" {
				file = a[7:]
				continue
			}
			fmt.Fprintln(stderr, "unknown flag:", a)
			return exitInputErr, errInvalidUsage
		}
	}

	src, err := readInput(file, stdin)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitInputErr, err
	}

	doc, err := adf.Convert(src)
	if err != nil {
		fmt.Fprintln(stderr, "parse error:", err)
		return exitParseErr, err
	}

	out, err := adf.Marshal(doc, pretty)
	if err != nil {
		fmt.Fprintln(stderr, "marshal error:", err)
		return exitUnknownErr, err
	}

	if _, err := stdout.Write(out); err != nil {
		return exitUnknownErr, err
	}
	if pretty {
		fmt.Fprintln(stdout)
	}
	return exitOK, nil
}

// readInput returns markdown bytes from file (if provided) or from stdin.
func readInput(file string, stdin io.Reader) ([]byte, error) {
	if file != "" {
		b, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", file, err)
		}
		return b, nil
	}
	return io.ReadAll(stdin)
}
