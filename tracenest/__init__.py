"""tracenest — agent trace explorer.

Ingests structured spans emitted by agents (tool calls, LLM turns,
retries), rebuilds the parent/child nesting, computes the critical
path and timing summaries, detects retry clusters, and renders a
self-contained HTML/JSON report.

Public surface:
    load_spans        -- parse a trace file (JSON, JSONL, or NDJSON) into Span objects
    build_tree        -- assemble a SpanTree from a flat span list
    analyze           -- produce a TraceAnalysis (critical path, timings, retries)
    render_html       -- emit a single self-contained HTML report
    render_json       -- emit a normalized JSON report
"""

from .model import Span, SpanKind, SpanStatus
from .parser import load_spans, parse_spans, ParseError
from .tree import SpanTree, build_tree
from .analyze import TraceAnalysis, analyze
