"""Core data model for spans.

A *span* is a single timed unit of work in an agent run. Spans form a
forest via ``parent_id`` links. Every span carries a monotonic start and
end timestamp (in milliseconds), a kind, a status, and an open attribute
bag used for retry detection and display.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Dict, Optional


class SpanKind(str, Enum):
    """The category of work a span represents."""

    ROOT = "root"
    LLM = "llm"
    TOOL = "tool"
    RETRIEVAL = "retrieval"
    CHAIN = "chain"
    EVENT = "event"
    UNKNOWN = "unknown"

    @classmethod
    def coerce(cls, value: Any) -> "SpanKind":
        if isinstance(value, cls):
            return value
        if value is None:
            return cls.UNKNOWN
        text = str(value).strip().lower()
        for member in cls:
            if member.value == text:
                return member
        # Common aliases emitted by various agent frameworks.
        aliases = {
            "model": cls.LLM,
            "completion": cls.LLM,
            "generation": cls.LLM,
            "function": cls.TOOL,
            "action": cls.TOOL,
            "search": cls.RETRIEVAL,
            "embed": cls.RETRIEVAL,
            "span": cls.EVENT,
        }
        return aliases.get(text, cls.UNKNOWN)


class SpanStatus(str, Enum):
    """Terminal disposition of a span."""

    OK = "ok"
    ERROR = "error"
    CANCELLED = "cancelled"
    UNKNOWN = "unknown"

    @classmethod
    def coerce(cls, value: Any) -> "SpanStatus":
        if isinstance(value, cls):
            return value
        if value is None:
            return cls.UNKNOWN
        text = str(value).strip().lower()
        for member in cls:
            if member.value == text:
                return member
        aliases = {
            "success": cls.OK,
            "succeeded": cls.OK,
            "completed": cls.OK,
            "failed": cls.ERROR,
            "failure": cls.ERROR,
            "exception": cls.ERROR,
            "aborted": cls.CANCELLED,
            "canceled": cls.CANCELLED,
        }
        return aliases.get(text, cls.UNKNOWN)


@dataclass
class Span:
    """A single timed unit of work.

    Attributes
    ----------
    span_id:
        Unique identifier within the trace.
    name:
        Human-readable label (tool name, model name, chain step).
    start_ms / end_ms:
        Wall-clock offsets in milliseconds. ``end_ms >= start_ms``.
    parent_id:
        Identifier of the enclosing span, or ``None`` for a root.
    kind / status:
        Enumerated category and disposition.
    attributes:
        Arbitrary key/value bag (token counts, model name, error text,
        retry markers, etc.).
    """

    span_id: str
    name: str
    start_ms: float
    end_ms: float
    parent_id: Optional[str] = None
    kind: SpanKind = SpanKind.UNKNOWN
    status: SpanStatus = SpanStatus.UNKNOWN
    attributes: Dict[str, Any] = field(default_factory=dict)

    @property
    def duration_ms(self) -> float:
        """Elapsed time of the span, never negative."""
        return max(0.0, self.end_ms - self.start_ms)

    def attr(self, key: str, default: Any = None) -> Any:
        return self.attributes.get(key, default)

    def to_dict(self) -> Dict[str, Any]:
        return {
            "span_id": self.span_id,
            "name": self.name,
            "start_ms": self.start_ms,
            "end_ms": self.end_ms,
            "duration_ms": self.duration_ms,
            "parent_id": self.parent_id,
            "kind": self.kind.value,
            "status": self.status.value,
            "attributes": self.attributes,
        }
