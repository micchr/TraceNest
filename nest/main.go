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

