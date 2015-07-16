# tracenest — build and demo targets (standard toolchains only)
#
#   make compile   compile the Python package and build the Go module
#   make demo       run both CLIs against the bundled fixtures
#   make report     render an HTML report from the sample trace
#   make clean      remove generated artifacts

PY ?= python
GO ?= go
TRACE ?= examples/agent_run.json

