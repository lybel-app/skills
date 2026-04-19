// lybel-docs converts extended markdown (with Confluence macros) into ADF
// JSON for use with the Atlassian Confluence REST API, and edits existing
// ADF documents by section without losing macros.
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

const helpText = `lybel-docs — convert extended markdown to ADF, and edit existing ADF docs.

USAGE:
  lybel-docs adf  [--file PATH] [--pretty]
  lybel-docs edit [--input PATH] OPERATION [--pretty]
  lybel-docs --version
  lybel-docs --help

COMMANDS:
  adf     Convert markdown (stdin or --file) to an ADF JSON document.
  edit    Apply a section-level edit to an ADF JSON document. Reads the
          current ADF from --input or stdin; writes the edited ADF to stdout.

EDIT OPERATIONS (exactly one required):
  --append FRAGMENT.md                   Append the fragment's blocks to the end.
  --insert-after  "Heading" FRAGMENT.md  Insert blocks right after the section.
  --insert-before "Heading" FRAGMENT.md  Insert blocks right before the section.
  --replace-section "Heading" FRAGMENT.md   Replace the section with the fragment.
  --delete-section  "Heading"            Remove the heading and its body.

SECTION SEMANTICS:
  A "section" is the matched heading plus all following top-level nodes up to
  (but not including) the next heading of equal or higher level. Headings are
  matched by exact case-sensitive text (trimmed). First match wins.

FLAGS:
  -f, --file  PATH   (adf)  Read markdown from PATH instead of stdin.
  -i, --input PATH   (edit) Read ADF from PATH instead of stdin. Use - for stdin.
      --pretty       Pretty-print the JSON output.
  -v, --version      Print version and exit.
  -h, --help         Show this help and exit.

MARKDOWN EXTENSIONS (adf & edit fragments):
  [TOC]                              Confluence Table of Contents macro.
  [TOC maxLevel=3 minLevel=1]        With explicit min/max levels.
  :::expand Title                    Expand block; close with :::
  :::warning Title                   Panel of type warning/info/note/success/error.

EXAMPLES:
  # Convert markdown to ADF
  lybel-docs adf < page.md > page.adf.json

  # Append a new section to an existing page (preserves all macros)
  lybel-docs edit --input page.json --append new-section.md > updated.json

  # Replace a section by heading text
  lybel-docs edit --input page.json \
    --replace-section "📇 Page ID Index" new-index.md > updated.json

  # Delete a stale section
  lybel-docs edit --input page.json --delete-section "TODO antigo" > updated.json

EXIT CODES:
  0  success
  1  parse error (markdown -> ADF or ADF unmarshal)
  2  invalid input (missing file, bad flags, section not found)
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
	case "edit":
		return runEdit(args[1:], stdin, stdout, stderr)
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

	return writeJSON(doc, pretty, stdout, stderr)
}

// editOp identifies which operation the edit subcommand will apply.
type editOp int

const (
	opNone editOp = iota
	opAppend
	opInsertAfter
	opInsertBefore
	opReplaceSection
	opDeleteSection
)

// runEdit parses edit-subcommand flags and applies one section-level operation
// to the ADF doc read from stdin or --input.
func runEdit(args []string, stdin io.Reader, stdout, stderr io.Writer) (int, error) {
	var (
		input        string
		pretty       bool
		op           editOp
		heading      string
		fragmentPath string
	)

	setOp := func(newOp editOp, name string) error {
		if op != opNone {
			return fmt.Errorf("multiple operations specified; use only one of --append, --insert-after, --insert-before, --replace-section, --delete-section")
		}
		op = newOp
		_ = name
		return nil
	}

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "-i", "--input":
			if i+1 >= len(args) {
				fmt.Fprintln(stderr, "flag", a, "requires a value")
				return exitInputErr, errInvalidUsage
			}
			input = args[i+1]
			i++
		case "--pretty":
			pretty = true
		case "-h", "--help":
			fmt.Fprint(stdout, helpText)
			return exitOK, nil

		case "--append":
			if err := setOp(opAppend, a); err != nil {
				fmt.Fprintln(stderr, err)
				return exitInputErr, errInvalidUsage
			}
			if i+1 >= len(args) {
				fmt.Fprintln(stderr, a, "requires FRAGMENT.md")
				return exitInputErr, errInvalidUsage
			}
			fragmentPath = args[i+1]
			i++

		case "--insert-after", "--insert-before", "--replace-section":
			var newOp editOp
			switch a {
			case "--insert-after":
				newOp = opInsertAfter
			case "--insert-before":
				newOp = opInsertBefore
			case "--replace-section":
				newOp = opReplaceSection
			}
			if err := setOp(newOp, a); err != nil {
				fmt.Fprintln(stderr, err)
				return exitInputErr, errInvalidUsage
			}
			if i+2 >= len(args) {
				fmt.Fprintln(stderr, a, `requires "Heading" FRAGMENT.md`)
				return exitInputErr, errInvalidUsage
			}
			heading = args[i+1]
			fragmentPath = args[i+2]
			i += 2

		case "--delete-section":
			if err := setOp(opDeleteSection, a); err != nil {
				fmt.Fprintln(stderr, err)
				return exitInputErr, errInvalidUsage
			}
			if i+1 >= len(args) {
				fmt.Fprintln(stderr, a, `requires "Heading"`)
				return exitInputErr, errInvalidUsage
			}
			heading = args[i+1]
			i++

		default:
			fmt.Fprintln(stderr, "unknown flag:", a)
			return exitInputErr, errInvalidUsage
		}
	}

	if op == opNone {
		fmt.Fprintln(stderr, "edit: no operation specified")
		fmt.Fprint(stderr, helpText)
		return exitInputErr, errInvalidUsage
	}

	adfBytes, err := readADFInput(input, stdin)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitInputErr, err
	}

	doc, err := adf.UnmarshalDoc(adfBytes)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitParseErr, err
	}

	var fragment []adf.Node
	if fragmentPath != "" {
		src, err := os.ReadFile(fragmentPath)
		if err != nil {
			fmt.Fprintln(stderr, "reading fragment:", err)
			return exitInputErr, err
		}
		nodes, err := adf.ConvertFragment(src)
		if err != nil {
			fmt.Fprintln(stderr, "parse fragment:", err)
			return exitParseErr, err
		}
		fragment = nodes
	}

	var result adf.Node
	switch op {
	case opAppend:
		result = adf.Append(doc, fragment)
	case opInsertAfter:
		result, err = adf.InsertAfter(doc, heading, fragment)
	case opInsertBefore:
		result, err = adf.InsertBefore(doc, heading, fragment)
	case opReplaceSection:
		result, err = adf.ReplaceSection(doc, heading, fragment)
	case opDeleteSection:
		result, err = adf.DeleteSection(doc, heading)
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitInputErr, err
	}

	return writeJSON(result, pretty, stdout, stderr)
}

// writeJSON marshals n and writes it to stdout.
func writeJSON(n adf.Node, pretty bool, stdout, stderr io.Writer) (int, error) {
	out, err := adf.Marshal(n, pretty)
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

// readADFInput returns ADF JSON bytes from file or stdin. "-" or empty means stdin.
func readADFInput(path string, stdin io.Reader) ([]byte, error) {
	if path == "" || path == "-" {
		return io.ReadAll(stdin)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return b, nil
}
