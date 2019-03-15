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
