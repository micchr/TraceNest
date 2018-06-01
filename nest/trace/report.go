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
		if color == "" {
			color = kindColors["unknown"]
		}
		classes := "row"
		if crit[s.ID] {
			classes += " crit"
		}
		if s.Status == StatusError {
			classes += " err"
		}
		label := html.EscapeString(s.Name)
		meta := html.EscapeString(fmt.Sprintf("%s · %s", s.Kind, fmtMS(s.DurationMS())))
		fmt.Fprintf(&sb,
			`<div class="%s" style="padding-left:%dpx"><div class="lbl" title="%s">%s`+
				`<span class="sub">%s</span></div><div class="track"><div class="bar" `+
				`style="left:%.3f%%;width:%.3f%%;background:%s"></div></div></div>`+"\n",
			classes, n.Depth*16, label, label, meta, left, width, color)
	})
	return sb.String()
}

func legend() string {
	order := []string{"root", "llm", "tool", "retrieval", "chain", "event"}
	var sb strings.Builder
	for _, k := range order {
		fmt.Fprintf(&sb, `<span class="chip"><i style="background:%s"></i>%s</span>`, kindColors[k], k)
	}
	return sb.String()
}

func cards(a Analysis) string {
	pairs := [][2]string{
		{"total", fmtMS(a.TotalMS)},
		{"spans", fmt.Sprintf("%d", a.SpanCount)},
		{"depth", fmt.Sprintf("%d", a.MaxDepth)},
		{"critical path", fmtMS(a.CriticalPathMS)},
		{"errors", fmt.Sprintf("%d", len(a.ErrorSpanIDs))},
		{"retry groups", fmt.Sprintf("%d", len(a.Retries))},
	}
	var sb strings.Builder
	for _, p := range pairs {
		fmt.Fprintf(&sb, `<div class="card"><div class="k">%s</div><div class="v">%s</div></div>`,
			html.EscapeString(p[0]), html.EscapeString(p[1]))
	}
	return sb.String()
}

func retryRows(a Analysis) string {
	if len(a.Retries) == 0 {
		return `<tr><td colspan="4" class="muted">no retry clusters detected</td></tr>`
	}
	var sb strings.Builder
	for _, r := range a.Retries {
		state := "repeated"
		if r.Recovered {
			state = "recovered"
		} else if r.Errors > 0 {
			state = "failing"
		}
		fmt.Fprintf(&sb, `<tr><td>%s</td><td>%d</td><td>%d</td><td>%s</td></tr>`,
			html.EscapeString(r.Name), r.Attempts, r.Errors, state)
	}
	return sb.String()
}

// kindOrder gives a stable listing used by callers that print summaries.
func kindOrder(m map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return m[keys[i]] > m[keys[j]] })
	return keys
}

// RenderHTML returns a single self-contained HTML report.
