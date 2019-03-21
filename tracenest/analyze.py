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
