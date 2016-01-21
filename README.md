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
