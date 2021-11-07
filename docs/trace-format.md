# Trace Format Reference

tracenest reads a *trace* — a set of spans describing one agent run. The
reader is deliberately permissive so traces emitted by different agent
frameworks parse without a translation layer.

## Container shapes

A trace file may be any one of:

| Shape          | Detection            | Example                              |
|----------------|----------------------|--------------------------------------|
| JSON array     | first char is `[`    | `[{...}, {...}]`                     |
| Object envelope| first char is `{`    | `{"spans": [{...}]}`                 |
| JSONL / NDJSON | one object per line  | `{...}\n{...}\n`                     |

For the object envelope, the first present key among `spans`, `trace`,
`events` is used as the span list. If an object has none of those, it is
treated as a single-span trace.

## Span fields

Every span is a JSON object. Field names are matched flexibly; the first
present alias wins.

| Canonical    | Accepted aliases                                   | Required |
|--------------|----------------------------------------------------|----------|
| `span_id`    | `span_id`, `spanId`, `id`                           | no*      |
| `parent_id`  | `parent_id`, `parentSpanId`, `parentId`, `parent`  | no       |
| `name`       | `name`, `label`, `operation`, `op`                 | no*      |
| `start_ms`   | `start_ms`, `startTime`, `start`, `startedAt`, `ts`| **yes**  |
| `end_ms`     | `end_ms`, `endTime`, `end`, `endedAt`              | no       |
| `duration_ms`| `duration_ms`, `duration`, `elapsed_ms`, `elapsed` | no       |
| `kind`       | `kind`, `type`, `category`                         | no       |
| `status`     | `status`, `state`, `outcome`, `result`             | no       |

\* If `span_id` is absent it is synthesized as `span-N` by ordinal
position; if `name` is absent it defaults to the span id.

Only `start` is strictly required. If `end` is missing, tracenest uses
`start + duration`; if both are missing the span is zero-length.

## Timestamps

A timestamp may be:

- a number (milliseconds), or
- a numeric string, or
- an ISO-8601 datetime string (a trailing `Z` is accepted).

All timestamps are converted to a millisecond float and then **rebased**
so the earliest span in the trace starts at zero. Reports therefore show
offsets relative to trace start, not wall-clock epoch.

## Kinds

`kind` is normalized to one of: `root`, `llm`, `tool`, `retrieval`,
`chain`, `event`, `unknown`. Common aliases are folded in:

- `model`, `completion`, `generation` → `llm`
- `function`, `action` → `tool`
- `search`, `embed` → `retrieval`
- `span` → `event`

## Status

`status` normalizes to `ok`, `error`, `cancelled`, or `unknown`. Aliases
such as `success`, `succeeded`, `completed` map to `ok`; `failed`,
