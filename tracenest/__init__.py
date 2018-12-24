"""tracenest — agent trace explorer.

Ingests structured spans emitted by agents (tool calls, LLM turns,
retries), rebuilds the parent/child nesting, computes the critical
path and timing summaries, detects retry clusters, and renders a
self-contained HTML/JSON report.
