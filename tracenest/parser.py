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
