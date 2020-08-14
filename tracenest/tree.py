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


_SYNTHETIC_ID = "__tracenest_root__"


def build_tree(spans: Iterable[Span]) -> SpanTree:
    """Assemble a :class:`SpanTree` from a flat span iterable.

    Spans are attached to their declared parent when present. Missing
    parents cause the span to be treated as a root. If more than one
    root results, a synthetic root is inserted so the report always has
    a single entry point spanning the whole trace.
    """
    span_list = list(spans)
    nodes: Dict[str, TreeNode] = {}
    for span in span_list:
        # Later duplicates overwrite earlier ones deterministically.
        nodes[span.span_id] = TreeNode(span=span)

    roots: List[TreeNode] = []
    for node in nodes.values():
        parent_id = node.span.parent_id
        if parent_id is not None and parent_id in nodes and parent_id != node.span_id:
            nodes[parent_id].children.append(node)
        else:
            roots.append(node)

    _break_cycles(nodes, roots)

    # Order children and roots by start time for stable, readable output.
    for node in nodes.values():
        node.children.sort(key=lambda n: (n.span.start_ms, n.span.span_id))
    roots.sort(key=lambda n: (n.span.start_ms, n.span.span_id))

    synthetic_id: Optional[str] = None
    if len(roots) > 1:
        synthetic_id = _SYNTHETIC_ID
        start = min(n.span.start_ms for n in roots)
        end = max(n.span.end_ms for n in roots)
        synthetic = TreeNode(
            span=Span(
                span_id=synthetic_id,
                name="trace",
                start_ms=start,
                end_ms=end,
                parent_id=None,
                kind=SpanKind.ROOT,
                status=SpanStatus.OK,
                attributes={"synthetic": True, "root_count": len(roots)},
            ),
            children=list(roots),
        )
        nodes[synthetic_id] = synthetic
        roots = [synthetic]

    _assign_depth(roots, 0)
    return SpanTree(roots=roots, index=nodes, synthetic_root_id=synthetic_id)


