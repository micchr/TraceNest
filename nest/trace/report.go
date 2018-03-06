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

