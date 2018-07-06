package trace

import "sort"

// Node is a span plus its ordered children and computed depth.
type Node struct {
	Span     Span
	Depth    int
	Children []*Node
}

// Tree is a forest of spans with lookup and traversal helpers.
type Tree struct {
	Roots         []*Node
	Index         map[string]*Node
	SyntheticRoot string
}

const syntheticID = "__tracenest_root__"

// Walk visits nodes depth-first, pre-order.
func (t *Tree) Walk(fn func(*Node)) {
	var visit func(*Node)
	visit = func(n *Node) {
		fn(n)
		for _, c := range n.Children {
			visit(c)
		}
	}
	for _, r := range t.Roots {
		visit(r)
	}
}

// Len returns the number of indexed spans.
func (t *Tree) Len() int { return len(t.Index) }

// MaxDepth returns the deepest depth in the tree.
func (t *Tree) MaxDepth() int {
	max := 0
	t.Walk(func(n *Node) {
		if n.Depth > max {
			max = n.Depth
		}
	})
	return max
}

// BuildTree assembles a Tree from a flat span slice, reattaching orphans
// and breaking cycles so the result is always a DAG with a single entry.
func BuildTree(spans []Span) *Tree {
	nodes := make(map[string]*Node, len(spans))
	order := make([]string, 0, len(spans))
	for _, s := range spans {
		if _, exists := nodes[s.ID]; !exists {
			order = append(order, s.ID)
		}
		nodes[s.ID] = &Node{Span: s}
	}

	var roots []*Node
	for _, id := range order {
		n := nodes[id]
		p := n.Span.ParentID
		if p != "" && p != n.Span.ID {
			if parent, ok := nodes[p]; ok {
				parent.Children = append(parent.Children, n)
				continue
			}
		}
		roots = append(roots, n)
	}

	roots = breakCycles(nodes, roots)

	for _, n := range nodes {
		sort.SliceStable(n.Children, func(i, j int) bool {
			if n.Children[i].Span.StartMS != n.Children[j].Span.StartMS {
				return n.Children[i].Span.StartMS < n.Children[j].Span.StartMS
			}
			return n.Children[i].Span.ID < n.Children[j].Span.ID
		})
	}
