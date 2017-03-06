package trace

import "sort"

// RetryCluster is a group of same-named sibling attempts.
type RetryCluster struct {
	Name      string   `json:"name"`
	ParentID  string   `json:"parent_id,omitempty"`
	Attempts  int      `json:"attempts"`
	Errors    int      `json:"errors"`
	SpanIDs   []string `json:"span_ids"`
	Recovered bool     `json:"recovered"`
}

// Analysis is the full result of analyzing a trace.
type Analysis struct {
	TotalMS        float64            `json:"total_ms"`
	SpanCount      int                `json:"span_count"`
	MaxDepth       int                `json:"max_depth"`
	KindMS         map[string]float64 `json:"kind_ms"`
	KindCounts     map[string]int     `json:"kind_counts"`
	StatusCounts   map[string]int     `json:"status_counts"`
	SelfMS         map[string]float64 `json:"self_ms"`
	CriticalPath   []string           `json:"critical_path"`
	CriticalPathMS float64            `json:"critical_path_ms"`
	Retries        []RetryCluster     `json:"retries"`
	ErrorSpanIDs   []string           `json:"error_span_ids"`
}

func selfTime(n *Node) float64 {
	total := n.Span.DurationMS()
	if len(n.Children) == 0 {
		return total
	}
	type iv struct{ s, e float64 }
	ivs := make([]iv, 0, len(n.Children))
	for _, c := range n.Children {
		ivs = append(ivs, iv{c.Span.StartMS, c.Span.EndMS})
	}
	sort.Slice(ivs, func(i, j int) bool { return ivs[i].s < ivs[j].s })
	var covered float64
	var cur iv
	have := false
	for _, x := range ivs {
		if have && x.s <= cur.e {
			if x.e > cur.e {
				cur.e = x.e
			}
		} else {
			if have {
				covered += cur.e - cur.s
			}
			cur = x
			have = true
		}
	}
	if have {
		covered += cur.e - cur.s
	}
