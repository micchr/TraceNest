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
