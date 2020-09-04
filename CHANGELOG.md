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
