"""Trace analysis.

Given a :class:`~tracenest.tree.SpanTree`, computes:

* **Critical path** — the chain of nested spans from a root to a leaf
  that maximizes cumulative *self* time, i.e. the sequence most
  responsible for total wall-clock latency.
* **Timing summaries** — total, per-kind, and self-time breakdowns.
* **Retry clusters** — sibling spans sharing a name whose executions
  repeat, with the failed-then-succeeded pattern flagged.

Self time is a span's own duration minus the union of its children's
intervals, so overlapping children never double-count.
"""

from __future__ import annotations

from collections import defaultdict
from dataclasses import dataclass, field
from typing import Dict, List, Optional, Tuple

from .model import Span, SpanKind, SpanStatus
from .tree import SpanTree, TreeNode


@dataclass
class RetryCluster:
    """A group of same-named sibling attempts."""

    name: str
    parent_id: Optional[str]
    attempts: int
    errors: int
    span_ids: List[str]
    recovered: bool  # ended OK after at least one error

    def to_dict(self) -> Dict[str, object]:
        return {
            "name": self.name,
            "parent_id": self.parent_id,
            "attempts": self.attempts,
            "errors": self.errors,
            "span_ids": self.span_ids,
            "recovered": self.recovered,
        }


@dataclass
class TraceAnalysis:
    """The full result of analyzing a trace."""

    total_ms: float
    span_count: int
    max_depth: int
    kind_ms: Dict[str, float]
    kind_counts: Dict[str, int]
    status_counts: Dict[str, int]
    self_ms: Dict[str, float]
    critical_path: List[str]
    critical_path_ms: float
    retries: List[RetryCluster] = field(default_factory=list)
    error_span_ids: List[str] = field(default_factory=list)

    def to_dict(self) -> Dict[str, object]:
        return {
            "total_ms": self.total_ms,
            "span_count": self.span_count,
            "max_depth": self.max_depth,
            "kind_ms": self.kind_ms,
            "kind_counts": self.kind_counts,
            "status_counts": self.status_counts,
            "self_ms": self.self_ms,
            "critical_path": self.critical_path,
            "critical_path_ms": self.critical_path_ms,
            "retries": [r.to_dict() for r in self.retries],
            "error_span_ids": self.error_span_ids,
        }


def _self_time(node: TreeNode) -> float:
    """Duration of a span minus the union of its children's intervals."""
    total = node.span.duration_ms
    if not node.children:
        return total
    intervals = sorted(
        (c.span.start_ms, c.span.end_ms) for c in node.children
    )
    merged: List[Tuple[float, float]] = []
    for start, end in intervals:
        if merged and start <= merged[-1][1]:
            merged[-1] = (merged[-1][0], max(merged[-1][1], end))
        else:
            merged.append((start, end))
    covered = sum(e - s for s, e in merged)
    return max(0.0, total - covered)


def _critical_path(tree: SpanTree) -> Tuple[List[str], float]:
    """Chain from a root to a leaf maximizing cumulative self time."""
    best_cost: Dict[str, float] = {}
    best_next: Dict[str, Optional[str]] = {}

    # Post-order so children are solved before parents.
    order: List[TreeNode] = []

    def collect(node: TreeNode) -> None:
        for child in node.children:
            collect(child)
        order.append(node)

    for root in tree.roots:
        collect(root)

    self_cache = {node.span_id: _self_time(node) for node in tree}

    for node in order:
        own = self_cache[node.span_id]
        if not node.children:
            best_cost[node.span_id] = own
            best_next[node.span_id] = None
            continue
        best_child = max(node.children, key=lambda c: best_cost[c.span_id])
        best_cost[node.span_id] = own + best_cost[best_child.span_id]
        best_next[node.span_id] = best_child.span_id

    if not tree.roots:
        return [], 0.0

    start = max(tree.roots, key=lambda r: best_cost[r.span_id])
    path: List[str] = []
    cursor: Optional[str] = start.span_id
    total = best_cost[start.span_id]
    while cursor is not None:
        path.append(cursor)
        cursor = best_next[cursor]
    return path, total


def _detect_retries(tree: SpanTree) -> List[RetryCluster]:
    """Find sibling attempts sharing a name (retry signature)."""
    clusters: List[RetryCluster] = []
    # Group children by (parent_id, name).
    parents: List[TreeNode] = [n for n in tree if n.children]
    for parent in parents:
