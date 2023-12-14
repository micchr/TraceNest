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
- [x] **Critical path & retry analysis** — self-time computation,
  critical-path selection, per-kind timing, and retry clustering
  (`tracenest/analyze.py`, `nest/trace/analyze.go`).
- [x] **Self-contained reporting** — HTML waterfall plus normalized JSON,
  wired to both CLIs with `analyze` / `tree` / `report` commands
  (`tracenest/report.py`, `nest/trace/report.go`, `nest/main.go`).

## Planned

- [ ] Span diffing between two traces to surface latency regressions.
- [ ] Percentile aggregation across many traces of the same workflow.
- [ ] OpenTelemetry OTLP/JSON span import mapping.
- [ ] Collapsible subtrees and search in the HTML report.
- [ ] Token-cost overlay driven by per-span attribute annotations.

<!-- draft note 230 -->
