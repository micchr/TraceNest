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
	statusKeys = []string{"status", "state", "outcome", "result"}
)

var reserved = func() map[string]bool {
	m := map[string]bool{"attributes": true, "attrs": true, "meta": true,
		"metadata": true, "tags": true}
	for _, group := range [][]string{idKeys, parentKeys, nameKeys, startKeys,
		endKeys, durKeys, kindKeys, statusKeys} {
		for _, k := range group {
			m[k] = true
		}
	}
	return m
}()

func first(obj map[string]interface{}, keys []string) interface{} {
	for _, k := range keys {
		if v, ok := obj[k]; ok && v != nil {
			return v
		}
	}
	return nil
}

func firstString(obj map[string]interface{}, keys []string) (string, bool) {
	v := first(obj, keys)
	if v == nil {
		return "", false
	}
	switch t := v.(type) {
	case string:
		return t, true
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64), true
	default:
		return fmt.Sprintf("%v", t), true
	}
}

// toMS reduces a timestamp value to milliseconds. Returns ok=false when
// the value is absent or unparseable.
func toMS(v interface{}) (float64, bool) {
	if v == nil {
		return 0, false
	}
	switch t := v.(type) {
	case float64:
		return t, true
	case int:
		return float64(t), true
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return 0, false
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f, true
		}
		if strings.HasSuffix(s, "Z") {
			s = strings.TrimSuffix(s, "Z") + "+00:00"
		}
		for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
			if ts, err := time.Parse(layout, s); err == nil {
				return float64(ts.UnixNano()) / 1e6, true
			}
		}
	}
	return 0, false
