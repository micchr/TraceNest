// Package trace holds the span model and the parse/analyze pipeline for
// the tracenest Go command. It mirrors the Python package on-disk trace
// formats so a single trace file works with either implementation.
package trace

// Kind is the category of work a span represents.
type Kind string

const (
	KindRoot      Kind = "root"
	KindLLM       Kind = "llm"
	KindTool      Kind = "tool"
	KindRetrieval Kind = "retrieval"
	KindChain     Kind = "chain"
	KindEvent     Kind = "event"
	KindUnknown   Kind = "unknown"
)

