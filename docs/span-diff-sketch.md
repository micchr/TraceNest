# Span diffing sketch

Status: proposal. Nothing here is implemented yet.

## Goal

Given two traces of the same workflow (a baseline and a candidate), report where
latency moved: per span name, per depth, and along the critical path.

## Rough shape

```
tracenest diff baseline.json candidate.json
```

- Align spans by (name, parent name) first, then by ordinal when a name repeats.
- Reuse the existing analysis model on each side; diff `self_ms`, `total_ms` and the
  critical path.
- Output the same normalized JSON so a CI gate can fail when a span regresses past a
  threshold.

## Open questions

- Name-based alignment breaks when the agent renames tools between versions; is a
  fallback to position-within-parent good enough?
- Should output live behind the `report` command (`report --diff`) instead of a new
  subcommand?
