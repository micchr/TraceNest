# Changelog

All notable changes to tracenest are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project
uses [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Planned
- Span diffing between two traces of the same workflow.
- Percentile aggregation across many runs.

## [1.0.1] - 2026-09-15

### Fixed
- Retry table now keeps clusters stable when two siblings start at the
  same millisecond (tie broken by span id, matching the tree order).
- Waterfall rows no longer overlap by a pixel when a span's duration is
  shorter than the minimum bar width.

## [1.0.0] - 2026-09-01

### Added
- Flexible trace parser accepting JSON arrays, object envelopes
  (`spans` / `trace` / `events`), and JSONL/NDJSON, with alias matching
  for `id`/`parent`/`start`/`end`/`duration`/`kind`/`status` fields.
- Millisecond and ISO-8601 timestamp normalization, with automatic
  rebasing so the earliest span starts at zero.
- Span tree builder that reattaches orphan spans, inserts a synthetic
  root when multiple roots exist, and breaks accidental cycles.
- Analysis engine computing total time, per-kind and self-time
  breakdowns, status counts, and a critical path maximizing cumulative
  self time.
- Retry-cluster detection over same-named sibling spans and explicit
  retry attribute markers, flagging the recovered pattern.
- Self-contained HTML report (graphite/neon-green theme) with an inline
  nested waterfall, summary cards, retry table, and embedded JSON, plus
  a normalized JSON report.
- Python CLI (`python -m tracenest`) and a Go CLI (`nest`) with matching
  `analyze`, `tree`, and `report` subcommands producing identical
  summaries from the same trace file.
- Two animated SVG assets (pipeline, waterfall) and bundled example
  fixtures.

## [0.9.9] - 2025-12-09

### Changed
- Report ordering is fully deterministic: roots and children sort by
  start time, then id, so repeated runs render byte-identical HTML.
- The embedded JSON payload is emitted with stable key order.

### Fixed
- Empty traces (zero spans) render an empty state instead of a blank
  page with a missing summary.

## [0.9.7] - 2025-06-17

### Added
- `examples/rag_session.jsonl` — a retrieval-augmented session fixture
  covering the JSONL path, `id`/`parent` aliases and `start`+`duration`
  spans.
- Full alias tables in `docs/trace-format.md`, including the attribute
  keys read by retry detection.

## [0.9.4] - 2024-12-10

### Fixed
- Numeric timestamp strings ("1340") are now accepted, not only JSON
  numbers.
- ISO-8601 timestamps with a trailing `Z` parse on both CLIs.

### Changed
- `kind` aliases expanded (`completion`, `generation` -> `llm`;
  `embed` -> `retrieval`).

## [0.9.0] - 2023-11-28

### Added
- Critical-path rail: rows on the critical path carry a glowing neon
  edge in the waterfall.
- Retry table states (`recovered`, `failing`, `repeated`).

## [0.8.1] - 2022-10-18

### Fixed
- Accidental cycles in the parent graph are broken by re-rooting the
  offending child instead of recursing forever.
- Orphan spans reattach as roots even when their declared parent appears
  later in the file.

## [0.7.0] - 2021-11-09

### Added
- Summary cards (total time, span count, max depth, critical-path time,
  error count, retry-group count).
- The complete normalized report embedded in the HTML output under
  `window.tracenest`.

## [0.6.0] - 2020-12-15

### Added
- JSONL / NDJSON input on both CLIs.
- Automatic rebasing of all timestamps to a zero origin.

## [0.5.0] - 2019-10-29

### Added
- Go implementation of the full pipeline (`nest/`), with an
  `analyze` / `tree` / `report` CLI matching the Python one.

### Changed
- Shared test fixtures introduced so the two implementations are
  checked against the same traces.

## [0.4.0] - 2018-11-06

### Added
- Self-contained HTML report with the nested span waterfall, colored by
  kind.
- `--format json` normalized report on both CLIs.

## [0.3.0] - 2017-10-24

### Added
- Retry-cluster detection: same-named sibling attempts grouped with
  attempt/error counts and the recovered pattern flagged.

## [0.2.0] - 2016-09-13

### Added
- Analysis engine: total time, per-kind breakdowns, and the critical
  path maximizing cumulative self time.
- Self time computed over the merged union of child intervals so
  concurrent children never double-count.

## [0.1.0] - 2015-10-14

### Added
- First release: permissive trace reader (JSON array and object
  envelope), span nesting into a tree, and a plain-text timing summary.
- `python -m tracenest analyze <trace>` entry point.

[Unreleased]: https://github.com/micchr/tracenest/compare/v1.0.1...HEAD
[1.0.1]: https://github.com/micchr/tracenest/releases/tag/v1.0.1
[1.0.0]: https://github.com/micchr/tracenest/releases/tag/v1.0.0
[0.9.9]: https://github.com/micchr/tracenest/releases/tag/v0.9.9
[0.9.7]: https://github.com/micchr/tracenest/releases/tag/v0.9.7
[0.9.4]: https://github.com/micchr/tracenest/releases/tag/v0.9.4
[0.9.0]: https://github.com/micchr/tracenest/releases/tag/v0.9.0
[0.8.1]: https://github.com/micchr/tracenest/releases/tag/v0.8.1
[0.7.0]: https://github.com/micchr/tracenest/releases/tag/v0.7.0
[0.6.0]: https://github.com/micchr/tracenest/releases/tag/v0.6.0
[0.5.0]: https://github.com/micchr/tracenest/releases/tag/v0.5.0
[0.4.0]: https://github.com/micchr/tracenest/releases/tag/v0.4.0
[0.3.0]: https://github.com/micchr/tracenest/releases/tag/v0.3.0
[0.2.0]: https://github.com/micchr/tracenest/releases/tag/v0.2.0
[0.1.0]: https://github.com/micchr/tracenest/releases/tag/v0.1.0

<!-- draft note 201 -->
