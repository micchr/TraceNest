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

// CoerceKind maps a raw string (including common aliases) to a Kind.
func CoerceKind(s string) Kind {
	switch normalize(s) {
	case "root":
		return KindRoot
	case "llm", "model", "completion", "generation":
		return KindLLM
	case "tool", "function", "action":
		return KindTool
	case "retrieval", "search", "embed":
		return KindRetrieval
	case "chain":
		return KindChain
	case "event", "span":
		return KindEvent
	default:
		return KindUnknown
	}
}

// Status is the terminal disposition of a span.
type Status string

const (
	StatusOK        Status = "ok"
	StatusError     Status = "error"
	StatusCancelled Status = "cancelled"
	StatusUnknown   Status = "unknown"
)

// CoerceStatus maps a raw string (including common aliases) to a Status.
func CoerceStatus(s string) Status {
	switch normalize(s) {
	case "ok", "success", "succeeded", "completed":
		return StatusOK
	case "error", "failed", "failure", "exception":
		return StatusError
	case "cancelled", "canceled", "aborted":
		return StatusCancelled
	default:
		return StatusUnknown
	}
}

// Span is a single timed unit of work.
type Span struct {
	ID         string                 `json:"span_id"`
	Name       string                 `json:"name"`
	StartMS    float64                `json:"start_ms"`
	EndMS      float64                `json:"end_ms"`
	ParentID   string                 `json:"parent_id,omitempty"`
	Kind       Kind                   `json:"kind"`
	Status     Status                 `json:"status"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
