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
