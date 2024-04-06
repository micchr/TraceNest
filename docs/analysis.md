# Analysis Reference

This note explains exactly what the numbers in a tracenest report mean.

## Self time

A span's **self time** is its own duration minus the union of its
children's intervals:

```
self(s) = duration(s) - measure( union of [child.start, child.end] )
```

Because the child intervals are merged before subtracting, overlapping
children (concurrent tool calls) are never double-counted, and self time
is clamped at zero. Self time answers "how much wall-clock time is spent
in *this* span rather than in the work it delegates?"

## Critical path

The **critical path** is the root-to-leaf chain that maximizes the sum of
self times along it. It is computed with a single post-order pass:

```
best(leaf)   = self(leaf)
best(parent) = self(parent) + max(best(child) for child in children)
```

The reported `critical_path` is the sequence of span ids from the highest
root down through the best child at each level; `critical_path_ms` is the
accumulated self time. This is the chain most responsible for total
latency — the place to optimize first. In the HTML report these spans get
a neon rail on their left edge.

## Timing breakdowns

- `total_ms` — span of the whole trace: max end minus min start across
  roots.
- `kind_ms` / `kind_counts` — summed duration and count per kind.
- `self_ms` — summed self time per kind, useful for attributing latency.
- `status_counts` — how many spans ended `ok` / `error` / etc.

## Retry clusters

Sibling spans (same parent) that share a `name` are grouped. A group is
reported as a retry cluster when it has two or more members, or when any
member carries a retry attribute marker. For each cluster tracenest
reports:

- `attempts` — number of same-named siblings,
- `errors` — how many ended in `error`,
- `recovered` — `true` when at least one attempt errored **and** the last
  attempt (by start time) ended `ok`,
- `span_ids` — the attempts in start order.

Clusters are sorted by attempt count descending.

## Tree normalization guarantees

- Orphan spans (parent id points at a missing span) become roots.
- When more than one root exists, a synthetic `trace` root spans them so
  every report has a single entry point. The synthetic root is excluded
  from counts, kind sums, and the reported critical path.
- Cycles (a span reachable from itself) are broken by re-rooting the
  offending child, so analysis always runs on a DAG.

<!-- draft note 291 -->
