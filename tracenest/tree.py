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
        yield self
        for child in self.children:
            yield from child.walk()


@dataclass
class SpanTree:
    """A forest of spans with helpers for traversal and lookup."""

    roots: List[TreeNode]
    index: Dict[str, TreeNode]
    synthetic_root_id: Optional[str] = None

    def __iter__(self) -> Iterator[TreeNode]:
        for root in self.roots:
            yield from root.walk()

    def __len__(self) -> int:
        return len(self.index)

    def get(self, span_id: str) -> Optional[TreeNode]:
        return self.index.get(span_id)

    def spans(self) -> List[Span]:
        return [node.span for node in self]

    def max_depth(self) -> int:
        return max((node.depth for node in self), default=0)


