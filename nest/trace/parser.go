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
}

func extractAttributes(obj map[string]interface{}) map[string]interface{} {
	attrs := map[string]interface{}{}
	for _, bag := range []string{"attributes", "attrs", "meta", "metadata", "tags"} {
		if b, ok := obj[bag].(map[string]interface{}); ok {
			for k, v := range b {
				attrs[k] = v
			}
		}
	}
	for k, v := range obj {
		if !reserved[k] {
			if _, exists := attrs[k]; !exists {
				attrs[k] = v
			}
		}
	}
	return attrs
}

func spanFromObj(obj map[string]interface{}, index int) (Span, error) {
	var s Span
	if id, ok := firstString(obj, idKeys); ok {
		s.ID = id
	} else {
		s.ID = fmt.Sprintf("span-%d", index)
	}
	if p, ok := firstString(obj, parentKeys); ok {
		s.ParentID = p
	}
	if n, ok := firstString(obj, nameKeys); ok {
		s.Name = n
	} else {
		s.Name = s.ID
	}

	start, hasStart := toMS(first(obj, startKeys))
	if !hasStart {
		return s, &ParseError{Msg: fmt.Sprintf("span '%s' is missing a start timestamp", s.ID)}
	}
	s.StartMS = start

	if end, ok := toMS(first(obj, endKeys)); ok {
		s.EndMS = end
	} else if dur, ok := toMS(first(obj, durKeys)); ok {
		s.EndMS = start + dur
	} else {
		s.EndMS = start
	}
	if s.EndMS < s.StartMS {
		s.EndMS = s.StartMS
	}

	kindStr, _ := firstString(obj, kindKeys)
	s.Kind = CoerceKind(kindStr)
	statusStr, _ := firstString(obj, statusKeys)
	s.Status = CoerceStatus(statusStr)
	s.Attributes = extractAttributes(obj)
	return s, nil
}

func rebase(spans []Span) []Span {
	if len(spans) == 0 {
		return spans
	}
	origin := spans[0].StartMS
	for _, s := range spans {
		if s.StartMS < origin {
			origin = s.StartMS
		}
	}
	if origin == 0 {
