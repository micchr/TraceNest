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
