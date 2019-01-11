"""tracenest — agent trace explorer.

Ingests structured spans emitted by agents (tool calls, LLM turns,
retries), rebuilds the parent/child nesting, computes the critical
path and timing summaries, detects retry clusters, and renders a
self-contained HTML/JSON report.

Public surface:
    load_spans        -- parse a trace file (JSON, JSONL, or NDJSON) into Span objects
