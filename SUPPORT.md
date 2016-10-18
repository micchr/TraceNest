# Support

tracenest is a small, dependency-free tool. Before opening an issue,
please try the steps below — they cover nearly every report.

## Self-serve checklist

1. Confirm the trace parses at all:

   ```bash
   python -m tracenest analyze examples/agent_run.json
   ```

2. Try your own trace on both CLIs and compare:

   ```bash
   python -m tracenest analyze your_trace.json --json
   cd nest && go run . analyze ../your_trace.json --json
   ```

   Both are expected to produce the same summary. If they differ, that
   is a bug worth reporting with the trace attached (trimmed if needed).

3. Check the reference docs:
   - [docs/trace-format.md](../docs/trace-format.md) — accepted fields,
