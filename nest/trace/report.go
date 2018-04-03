package trace

import (
	"encoding/json"
	"fmt"
	"html"
	"sort"
	"strings"
)

// kindColors mirrors the Python palette so both renderers match.
var kindColors = map[string]string{
	"root":      "#8a94a6",
	"llm":       "#39ff14",
	"tool":      "#00e0a4",
	"retrieval": "#5bd6ff",
	"chain":     "#c8ff5b",
	"event":     "#a0f0c0",
	"unknown":   "#6b7280",
}

type reportDoc struct {
	Version int      `json:"version"`
	Summary Analysis `json:"summary"`
	Spans   []Span   `json:"spans"`
}

func buildDoc(t *Tree, a Analysis) reportDoc {
	var spans []Span
	t.Walk(func(n *Node) {
		if n.Span.ID != t.SyntheticRoot {
			spans = append(spans, n.Span)
		}
	})
	return reportDoc{Version: 1, Summary: a, Spans: spans}
}

// RenderJSON returns a normalized JSON report.
func RenderJSON(t *Tree, a Analysis) (string, error) {
	b, err := json.MarshalIndent(buildDoc(t, a), "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func fmtMS(ms float64) string {
	if ms >= 1000 {
		return fmt.Sprintf("%.2fs", ms/1000)
	}
	return fmt.Sprintf("%.1fms", ms)
}

func bars(t *Tree, a Analysis) string {
	total := a.TotalMS
	if total == 0 {
		total = 1
	}
	var origin float64
	if len(t.Roots) > 0 {
		origin = t.Roots[0].Span.StartMS
		for _, r := range t.Roots {
			if r.Span.StartMS < origin {
				origin = r.Span.StartMS
			}
		}
	}
	crit := map[string]bool{}
	for _, id := range a.CriticalPath {
		crit[id] = true
	}
	var sb strings.Builder
	t.Walk(func(n *Node) {
		s := n.Span
		if s.ID == t.SyntheticRoot {
			return
		}
		left := (s.StartMS - origin) / total * 100
		width := s.DurationMS() / total * 100
		if width < 0.6 {
			width = 0.6
		}
		color := kindColors[string(s.Kind)]
