// Command nest is the Go CLI for tracenest. It parses a structured trace,
// rebuilds the span nesting, computes the critical path and timing/retry
// summaries, and renders text, JSON, or a self-contained HTML report.
//
// Usage:
//
//	nest analyze TRACE [--json]
//	nest tree    TRACE
//	nest report  TRACE [--out FILE] [--format html|json]
//
// TRACE may be a file path or "-" for standard input.
package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"tracenest/nest/trace"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) < 1 {
		usage()
		return 2
	}
	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "analyze":
		return cmdAnalyze(rest)
	case "tree":
		return cmdTree(rest)
	case "report":
		return cmdReport(rest)
	case "-h", "--help", "help":
		usage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "nest: unknown command %q\n", cmd)
		usage()
		return 2
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `nest — agent trace explorer

  nest analyze TRACE [--json]     print timing and retry summaries
  nest tree    TRACE              print the nested span tree
  nest report  TRACE [--out F] [--format html|json]

TRACE may be a file path or - for stdin.
`)
}

func readTree(source string) (*trace.Tree, error) {
	var data []byte
	var err error
	if source == "-" {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(source)
	}
	if err != nil {
		return nil, err
	}
	spans, err := trace.ParseSpans(data)
