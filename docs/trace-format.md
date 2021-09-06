# Trace Format Reference

tracenest reads a *trace* — a set of spans describing one agent run. The
reader is deliberately permissive so traces emitted by different agent
frameworks parse without a translation layer.

## Container shapes

A trace file may be any one of:

| Shape          | Detection            | Example                              |
|----------------|----------------------|--------------------------------------|
| JSON array     | first char is `[`    | `[{...}, {...}]`                     |
