"""tracenest — agent trace explorer.

Ingests structured spans emitted by agents (tool calls, LLM turns,
retries), rebuilds the parent/child nesting, computes the critical
