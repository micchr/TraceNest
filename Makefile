# tracenest — build and demo targets (standard toolchains only)
#
#   make compile   compile the Python package and build the Go module
#   make demo       run both CLIs against the bundled fixtures
#   make report     render an HTML report from the sample trace
#   make clean      remove generated artifacts

PY ?= python
GO ?= go
TRACE ?= examples/agent_run.json

.PHONY: compile py go demo report tree json clean

compile: py go

py:
	$(PY) -m compileall -q tracenest

go:
	cd nest && $(GO) build ./...
	cd nest && $(GO) vet ./...

demo: compile
	@echo "== python analyze =="
	$(PY) -m tracenest analyze $(TRACE)
	@echo "== go analyze =="
	cd nest && $(GO) run . analyze ../$(TRACE)

tree: compile
	$(PY) -m tracenest tree $(TRACE)

json: compile
	$(PY) -m tracenest analyze $(TRACE) --json

report: compile
