"""Span tree assembly.

Turns a flat list of spans into a navigable forest. Orphan spans (those
whose ``parent_id`` points at a missing span) are re-attached under a
synthetic root so nothing is silently dropped. Cycles are broken
defensively.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Dict, Iterable, Iterator, List, Optional

from .model import Span, SpanKind, SpanStatus


@dataclass
class TreeNode:
    """A span plus its ordered children and computed depth."""

    span: Span
    depth: int = 0
    children: List["TreeNode"] = field(default_factory=list)

    @property
    def span_id(self) -> str:
        return self.span.span_id

    def walk(self) -> Iterator["TreeNode"]:
        """Depth-first pre-order traversal, self first."""
