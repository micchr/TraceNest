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
