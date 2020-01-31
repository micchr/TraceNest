"""Trace parsers.

Supports three on-disk shapes, auto-detected from content:

* **JSON array** — a top-level list of span objects.
* **JSON object** — an object with a ``"spans"`` (or ``"trace"``) array.
* **JSONL / NDJSON** — one span object per line.

Each span object is normalized into a :class:`~tracenest.model.Span`.
Field names are matched flexibly so traces from different agent
frameworks parse without a translation layer:

    id / span_id / spanId               -> span_id
    parent / parent_id / parentSpanId   -> parent_id
    name / label / operation            -> name
    start / start_ms / startTime        -> start_ms
    end / end_ms / endTime              -> end_ms
    duration / duration_ms              -> used with start to derive end
    kind / type / category              -> kind
    status / state / outcome            -> status

Timestamps may be given in milliseconds (numbers) or ISO-8601 strings;
both are reduced to a millisecond float relative to the earliest span.
"""

from __future__ import annotations

import json
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Dict, List, Optional, Union

from .model import Span, SpanKind, SpanStatus

PathLike = Union[str, Path]


class ParseError(ValueError):
    """Raised when a trace cannot be interpreted."""


_ID_KEYS = ("span_id", "spanId", "id")
_PARENT_KEYS = ("parent_id", "parentSpanId", "parentId", "parent")
_NAME_KEYS = ("name", "label", "operation", "op")
_START_KEYS = ("start_ms", "startTime", "start", "startedAt", "ts")
_END_KEYS = ("end_ms", "endTime", "end", "endedAt")
_DUR_KEYS = ("duration_ms", "duration", "elapsed_ms", "elapsed")
_KIND_KEYS = ("kind", "type", "category")
_STATUS_KEYS = ("status", "state", "outcome", "result")

_RESERVED = set(
    _ID_KEYS + _PARENT_KEYS + _NAME_KEYS + _START_KEYS
    + _END_KEYS + _DUR_KEYS + _KIND_KEYS + _STATUS_KEYS
) | {"attributes", "attrs", "meta", "metadata", "tags"}


def _first(obj: Dict[str, Any], keys) -> Any:
    for key in keys:
        if key in obj and obj[key] is not None:
            return obj[key]
    return None


def _to_ms(value: Any) -> Optional[float]:
    """Reduce a timestamp to milliseconds.

    Accepts numeric epoch/relative milliseconds or ISO-8601 datetime
    strings. Returns ``None`` if the value is absent or unparseable.
    """
    if value is None:
        return None
    if isinstance(value, bool):
        return None
    if isinstance(value, (int, float)):
        return float(value)
    text = str(value).strip()
    if not text:
        return None
    # Numeric string.
    try:
        return float(text)
    except ValueError:
        pass
    # ISO-8601, tolerate a trailing Z.
    try:
        if text.endswith("Z"):
            text = text[:-1] + "+00:00"
        dt = datetime.fromisoformat(text)
        if dt.tzinfo is None:
            dt = dt.replace(tzinfo=timezone.utc)
        return dt.timestamp() * 1000.0
    except ValueError:
        return None


def _extract_attributes(obj: Dict[str, Any]) -> Dict[str, Any]:
    attrs: Dict[str, Any] = {}
    for bag_key in ("attributes", "attrs", "meta", "metadata", "tags"):
        bag = obj.get(bag_key)
        if isinstance(bag, dict):
            attrs.update(bag)
    # Fold any unreserved top-level keys into attributes.
    for key, value in obj.items():
        if key not in _RESERVED:
            attrs.setdefault(key, value)
    return attrs


def _span_from_obj(obj: Dict[str, Any], index: int) -> Span:
    if not isinstance(obj, dict):
        raise ParseError(f"span #{index} is not an object: {type(obj).__name__}")

    raw_id = _first(obj, _ID_KEYS)
    span_id = str(raw_id) if raw_id is not None else f"span-{index}"

    raw_parent = _first(obj, _PARENT_KEYS)
    parent_id = str(raw_parent) if raw_parent is not None else None

    name = _first(obj, _NAME_KEYS)
    name = str(name) if name is not None else span_id

