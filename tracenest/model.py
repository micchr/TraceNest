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
