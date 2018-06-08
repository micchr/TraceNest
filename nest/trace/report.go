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
func RenderHTML(t *Tree, a Analysis) (string, error) {
	b, err := json.Marshal(buildDoc(t, a))
	if err != nil {
		return "", err
	}
	payload := strings.ReplaceAll(string(b), "</", "<\\/")
	page := htmlTemplate
	page = strings.ReplaceAll(page, "{{CARDS}}", cards(a))
	page = strings.ReplaceAll(page, "{{LEGEND}}", legend())
	page = strings.ReplaceAll(page, "{{BARS}}", bars(t, a))
	page = strings.ReplaceAll(page, "{{RETRIES}}", retryRows(a))
	page = strings.ReplaceAll(page, "{{PAYLOAD}}", payload)
	return page, nil
}

const htmlTemplate = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>tracenest report</title>
<style>
:root{--bg:#0e1116;--panel:#161b22;--panel2:#1c2230;--ink:#d7e0ea;--muted:#8a94a6;--neon:#39ff14;--line:#242c38;}
*{box-sizing:border-box;}
body{margin:0;background:var(--bg);color:var(--ink);font:14px/1.5 ui-monospace,Menlo,Consolas,monospace;}
header{padding:20px 24px;border-bottom:1px solid var(--line);background:linear-gradient(180deg,#12161d,#0e1116);}
header h1{margin:0;font-size:18px;letter-spacing:.5px;}
header h1 span{color:var(--neon);}
.wrap{max-width:1080px;margin:0 auto;padding:20px 24px 60px;}
.cards{display:grid;grid-template-columns:repeat(6,1fr);gap:10px;margin:18px 0;}
.card{background:var(--panel);border:1px solid var(--line);border-radius:10px;padding:12px 14px;}
.card .k{color:var(--muted);font-size:11px;text-transform:uppercase;letter-spacing:.8px;}
.card .v{color:var(--neon);font-size:20px;margin-top:4px;}
h2{font-size:13px;text-transform:uppercase;letter-spacing:1px;color:var(--muted);margin:26px 0 10px;}
.legend{display:flex;gap:14px;flex-wrap:wrap;margin-bottom:10px;}
.chip{display:inline-flex;align-items:center;gap:6px;color:var(--muted);font-size:12px;}
.chip i{width:10px;height:10px;border-radius:2px;display:inline-block;}
.waterfall{background:var(--panel);border:1px solid var(--line);border-radius:10px;padding:8px;}
.row{display:grid;grid-template-columns:280px 1fr;gap:12px;align-items:center;padding:3px 6px;border-radius:6px;}
.row:hover{background:var(--panel2);}
.row.crit{box-shadow:inset 3px 0 0 var(--neon);}
.row.err .lbl{color:#ff6b6b;}
.lbl{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;}
.lbl .sub{color:var(--muted);font-size:11px;margin-left:8px;}
.track{position:relative;height:16px;background:#0c0f14;border-radius:4px;}
.bar{position:absolute;top:2px;height:12px;border-radius:3px;opacity:.9;}
.row.crit .bar{box-shadow:0 0 8px var(--neon);}
.row:hover .bar{filter:brightness(1.25);}
