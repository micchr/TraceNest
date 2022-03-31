# Contributing to tracenest

Thanks for taking the time to contribute. tracenest is deliberately small
and dependency-free, and contributions keep it that way.

## Ground rules

- No third-party dependencies in either language. Python uses only the
  standard library; Go uses only the standard library. A patch that adds a
  dependency will not be merged.
- Both implementations stay in lockstep. A behaviour change in
  `tracenest/` must land with the matching change in `nest/`, and both
  must produce the same summary on the bundled fixtures.
- Every behavioural change comes with a fixture or a test-free repro
  explained in the pull request description (the project keeps its test
  surface in `examples/`).

## Getting started

```bash
git clone https://github.com/micchr/tracenest.git
cd tracenest

# Python side
python -m compileall -q tracenest
python -m tracenest analyze examples/agent_run.json

# Go side
cd nest
go build ./...
go vet ./...
go run . analyze ../examples/agent_run.json
```

Or `make compile && make demo` from the repository root.

## Pull requests

1. Keep the change scoped: one behaviour or one document per pull request.
2. Run `make compile` (or the explicit commands above) before pushing.
3. Match the existing style: standard library only, deterministic output,
   no logging on the happy path.
4. Describe the *why* in the description — what trace shape or report
   problem the change addresses.

## Reporting bugs

Open an issue with the `bug` label. Include:

- the command you ran (Python or Go CLI),
