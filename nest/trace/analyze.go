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
	if r := total - covered; r > 0 {
		return r
	}
	return 0
}

func criticalPath(t *Tree, selfCache map[string]float64) ([]string, float64) {
	bestCost := map[string]float64{}
	bestNext := map[string]string{}

	var order []*Node
	var collect func(*Node)
	collect = func(n *Node) {
		for _, c := range n.Children {
			collect(c)
		}
		order = append(order, n)
	}
	for _, r := range t.Roots {
		collect(r)
	}

	for _, n := range order {
		own := selfCache[n.Span.ID]
		if len(n.Children) == 0 {
			bestCost[n.Span.ID] = own
			bestNext[n.Span.ID] = ""
			continue
		}
		var best *Node
		for _, c := range n.Children {
			if best == nil || bestCost[c.Span.ID] > bestCost[best.Span.ID] {
				best = c
			}
		}
		bestCost[n.Span.ID] = own + bestCost[best.Span.ID]
		bestNext[n.Span.ID] = best.Span.ID
	}

	if len(t.Roots) == 0 {
		return nil, 0
	}
	start := t.Roots[0]
	for _, r := range t.Roots {
		if bestCost[r.Span.ID] > bestCost[start.Span.ID] {
			start = r
		}
	}
	var path []string
	cursor := start.Span.ID
	total := bestCost[start.Span.ID]
	for cursor != "" {
		path = append(path, cursor)
		cursor = bestNext[cursor]
	}
	return path, total
}

func isMarkedRetry(s Span) bool {
	for _, key := range []string{"retry", "retries", "attempt", "retry_count"} {
		v, ok := s.Attributes[key]
		if !ok {
			continue
		}
		switch t := v.(type) {
		case bool:
			if t {
				return true
			}
		case float64:
			if t > 0 {
				return true
			}
		case string:
			switch normalize(t) {
			case "true", "yes", "retry":
				return true
			}
		}
	}
	return false
}

func detectRetries(t *Tree) []RetryCluster {
	var clusters []RetryCluster
	t.Walk(func(parent *Node) {
		if len(parent.Children) == 0 {
			return
		}
		buckets := map[string][]*Node{}
		var names []string
		for _, c := range parent.Children {
			if _, ok := buckets[c.Span.Name]; !ok {
				names = append(names, c.Span.Name)
			}
			buckets[c.Span.Name] = append(buckets[c.Span.Name], c)
		}
		for _, name := range names {
			group := buckets[name]
			marked := false
			for _, c := range group {
				if isMarkedRetry(c.Span) {
					marked = true
					break
				}
			}
			if len(group) < 2 && !marked {
				continue
			}
			sort.Slice(group, func(i, j int) bool {
				return group[i].Span.StartMS < group[j].Span.StartMS
			})
			errors := 0
			for _, c := range group {
				if c.Span.Status == StatusError {
					errors++
				}
			}
			recovered := errors > 0 && group[len(group)-1].Span.Status == StatusOK
			ids := make([]string, len(group))
			for i, c := range group {
				ids[i] = c.Span.ID
			}
			clusters = append(clusters, RetryCluster{
				Name: name, ParentID: parent.Span.ID, Attempts: len(group),
				Errors: errors, SpanIDs: ids, Recovered: recovered,
			})
		}
	})
	sort.SliceStable(clusters, func(i, j int) bool {
		if clusters[i].Attempts != clusters[j].Attempts {
			return clusters[i].Attempts > clusters[j].Attempts
		}
		return clusters[i].Name < clusters[j].Name
	})
	return clusters
}

// Analyze computes the full analysis for a span tree.
func Analyze(t *Tree) Analysis {
	kindMS := map[string]float64{}
	kindCounts := map[string]int{}
	statusCounts := map[string]int{}
	selfMS := map[string]float64{}
	selfCache := map[string]float64{}
	var errorIDs []string

	t.Walk(func(n *Node) { selfCache[n.Span.ID] = selfTime(n) })

	t.Walk(func(n *Node) {
		s := n.Span
		if s.ID == t.SyntheticRoot {
			return
		}
		k := string(s.Kind)
		kindMS[k] += s.DurationMS()
		kindCounts[k]++
		statusCounts[string(s.Status)]++
		selfMS[k] += selfCache[s.ID]
		if s.Status == StatusError {
			errorIDs = append(errorIDs, s.ID)
		}
	})

	var total float64
	if len(t.Roots) > 0 {
		minS, maxE := t.Roots[0].Span.StartMS, t.Roots[0].Span.EndMS
		for _, r := range t.Roots {
			if r.Span.StartMS < minS {
				minS = r.Span.StartMS
			}
			if r.Span.EndMS > maxE {
				maxE = r.Span.EndMS
			}
		}
		total = maxE - minS
	}
	if total < 0 {
		total = 0
	}

	path, pathMS := criticalPath(t, selfCache)
	filtered := path[:0:0]
	for _, id := range path {
		if id != t.SyntheticRoot {
			filtered = append(filtered, id)
		}
	}

	count := 0
	t.Walk(func(n *Node) {
		if n.Span.ID != t.SyntheticRoot {
			count++
		}
	})

	return Analysis{
		TotalMS: total, SpanCount: count, MaxDepth: t.MaxDepth(),
		KindMS: kindMS, KindCounts: kindCounts, StatusCounts: statusCounts,
		SelfMS: selfMS, CriticalPath: filtered, CriticalPathMS: pathMS,
		Retries: detectRetries(t), ErrorSpanIDs: errorIDs,
	}
}
