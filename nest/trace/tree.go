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
