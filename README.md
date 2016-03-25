# tracenest

**tracenest** is an agent-run trace explorer. Point it at the structured
spans your agent emits — tool calls, LLM turns, retrieval steps, retries —
and it rebuilds the nesting, finds the chain of work responsible for your
latency, flags retry storms, and renders a single self-contained report
you can open in any browser or feed to downstream tooling as JSON.

It ships as a matched pair: a Python 3.11 package (`tracenest`) and a Go
module (`nest/`). Both read the *same* trace files and produce *identical*
summaries, so you can drop whichever runtime is already in your pipeline
without re-teaching your tooling a new format. There are no third-party
dependencies in either language — only the standard libraries.

![tracenest pipeline](docs/assets/pipeline.svg)

```
ingest  →  nest  →  analyze  →  report
(parse)   (tree)   (crit path)  (html/json)
```

---

## Why tracenest exists

Agent frameworks emit a firehose of spans, but the raw log rarely answers
the questions that actually matter when a run is slow or flaky:

- Which chain of nested calls is *actually* driving the wall-clock time?
  Not the longest single span — the longest **path** of self time.
- Where did the agent retry, how many times, and did it eventually
  recover, or is it silently burning attempts?
- How is time split across LLM turns versus tool execution versus
  retrieval, once you stop double-counting nested work?

tracenest answers those three questions directly. It computes **self
time** (a span's duration minus the merged span of its children), walks
the tree to find the **critical path** by cumulative self time, groups
same-named sibling attempts into **retry clusters**, and lays it all out
in a graphite/neon-green waterfall where the critical path glows.

---

## Install / run

Nothing to install beyond a Python 3.11 interpreter and a Go toolchain.
Clone the repository and run either CLI in place.

**Python**

```bash
python -m compileall -q tracenest          # compile check
python -m tracenest analyze examples/agent_run.json
```

**Go**

```bash
cd nest
go build ./...                              # build check
go run . analyze ../examples/agent_run.json
```

Or use the Makefile to do both at once:

```bash
make compile     # compile Python + build/vet Go
make demo        # run both CLIs on the sample trace
make report      # render tracenest.report.html
```

---

## The trace format in 60 seconds

A trace is a set of spans. tracenest accepts three container shapes and
auto-detects which one you handed it:

- a **JSON array** of span objects,
- a **JSON object** with a `spans` (or `trace` / `events`) array,
- **JSONL / NDJSON**, one span object per line.

Each span needs only a start time; everything else has a sensible
default. Field names are matched flexibly, so traces from different
frameworks parse as-is:

```jsonc
{
  "id": "reason",            // span_id | spanId | id
  "parent": "root",          // parent_id | parentSpanId | parentId | parent
  "name": "reasoner.llm",    // name | label | operation | op
  "start": 1340,             // start_ms | startTime | start | startedAt | ts
  "end": 3200,               // end_ms | endTime | end | endedAt
  "kind": "llm",             // kind | type | category
  "status": "ok",            // status | state | outcome | result
  "attributes": { "model": "gpt-large", "completion_tokens": 540 }
}
```

Timestamps may be milliseconds (numbers or numeric strings) or ISO-8601
datetimes. tracenest normalizes them to a millisecond float and **rebases**
the whole trace so the earliest span starts at zero. Any top-level key
that is not a recognized field alias is folded into the span's attribute
bag, so you never lose framework-specific metadata.

`kind` normalizes to `root | llm | tool | retrieval | chain | event |
unknown` (with aliases such as `model → llm`, `function → tool`,
`search → retrieval`). `status` normalizes to `ok | error | cancelled |
unknown` (`success → ok`, `failure → error`, and so on). The full field
and alias tables live in [docs/trace-format.md](docs/trace-format.md).

---

## Event lifecycle

A span moves through four stages inside tracenest. Understanding them
makes the report anatomy obvious.

1. **Parse.** The reader detects the container shape, extracts each span
   object, resolves field aliases, coerces `kind`/`status`, converts
   timestamps to milliseconds, and derives `end` from `start + duration`
   when only a duration is present. Missing ids are synthesized as
   `span-N`. The whole set is then rebased to a zero origin.

2. **Nest.** Spans are linked to their declared parent. A span whose
   `parent_id` points at a missing span is treated as a root rather than
   being dropped. Children are ordered by start time for stable output.
   If the trace ends up with more than one root, a synthetic `trace` root
   is inserted so the report always has one entry point spanning the
   whole run. Accidental cycles (a span reachable from itself) are broken
   by re-rooting the offending child — analysis always runs on a DAG.

3. **Analyze.** For every span the engine computes self time by merging
   its children's intervals and subtracting their union from the span's
   own duration (concurrent children never double-count). A single
   post-order pass then selects the critical path — the root-to-leaf
   chain with the greatest cumulative self time. Per-kind durations,
   per-kind self time, status counts, and error span ids are tallied.
   Finally, same-named sibling spans are grouped into retry clusters and
   the recovered-after-error pattern is flagged.

4. **Report.** The normalized model is rendered either as JSON (for
   downstream tooling) or as a single self-contained HTML page: summary
   cards, a nested waterfall, and a retry table, with the full JSON
   payload embedded in a `<script type="application/json">` tag for
   programmatic consumers.

![span waterfall](docs/assets/waterfall.svg)

---

## Architecture

Both implementations share the same four-stage shape and the same data
model, which is why their output matches byte-for-byte on the same input.

```
                         ┌───────────────────────────────────────────┐
   trace file            │                 tracenest                  │
  (json / jsonl) ──────► │                                            │
                         │   parser ──► tree ──► analyze ──► report    │
                         │     │          │         │           │      │
                         │  aliases    reattach   self-time    html    │
                         │  ts norm    synth root crit path    json    │
                         │  rebase     de-cycle   retries      cards   │
                         └───────────────────────────────────────────┘
                                  │                        │
                        Python: tracenest/*.py     Go: nest/trace/*.go
                        CLI: python -m tracenest    CLI: nest (main.go)
```

### Python package (`tracenest/`)

| Module        | Responsibility                                            |
|---------------|-----------------------------------------------------------|
| `model.py`    | `Span`, `SpanKind`, `SpanStatus` with alias coercion      |
| `parser.py`   | Container detection, field aliasing, timestamp/rebase     |
| `tree.py`     | `SpanTree`, orphan reattach, synthetic root, de-cycle     |
| `analyze.py`  | self time, critical path, retry clusters, `TraceAnalysis` |
| `report.py`   | HTML + JSON rendering, shared color palette               |
| `cli.py`      | `analyze` / `tree` / `report` subcommands                 |

### Go module (`nest/`)

| File                   | Responsibility                                   |
|------------------------|--------------------------------------------------|
| `trace/model.go`       | `Span`, `Kind`, `Status`, coercion               |
| `trace/parser.go`      | container detection, aliasing, timestamps        |
| `trace/tree.go`        | `Tree`, reattach, synthetic root, de-cycle       |
| `trace/analyze.go`     | self time, critical path, retries                |
| `trace/report.go`      | HTML + JSON rendering                            |
| `main.go`              | `nest` CLI                                        |

The two share nothing at runtime — they are independent programs — but
they implement the same algorithms so a trace analyzed by one matches the
other. The bundled fixtures let you confirm this yourself in seconds.

---

## CLI recipes

Every command takes a trace path, or `-` to read from standard input.

**Timing + retry summary (text):**

```bash
python -m tracenest analyze examples/agent_run.json
```

```
spans        : 11
total        : 4200.0 ms
max depth    : 2
critical path: 1340.0 ms (3 spans)

by kind:
  chain          4200.0 ms  (1)
  llm            3240.0 ms  (3)
  tool           1460.0 ms  (5)
  retrieval       500.0 ms  (2)

status:
  error      2
  ok         9
