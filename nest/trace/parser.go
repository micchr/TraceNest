package trace

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseError describes a trace that cannot be interpreted.
type ParseError struct{ Msg string }

func (e *ParseError) Error() string { return "tracenest: " + e.Msg }

func normalize(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

var (
	idKeys     = []string{"span_id", "spanId", "id"}
	parentKeys = []string{"parent_id", "parentSpanId", "parentId", "parent"}
	nameKeys   = []string{"name", "label", "operation", "op"}
	startKeys  = []string{"start_ms", "startTime", "start", "startedAt", "ts"}
	endKeys    = []string{"end_ms", "endTime", "end", "endedAt"}
	durKeys    = []string{"duration_ms", "duration", "elapsed_ms", "elapsed"}
	kindKeys   = []string{"kind", "type", "category"}
