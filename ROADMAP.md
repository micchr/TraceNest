# Roadmap

tracenest ships a working dual-language pipeline today. The milestones
below marked complete map directly to code in this repository.

## Completed

- [x] **Flexible trace ingestion** — JSON array, object envelope, and
  JSONL parsers with field-alias matching and timestamp normalization
  (`tracenest/parser.py`, `nest/trace/parser.go`).
- [x] **Nesting tree builder** — orphan reattachment, synthetic root,
  cycle breaking, and depth assignment
  (`tracenest/tree.py`, `nest/trace/tree.go`).
