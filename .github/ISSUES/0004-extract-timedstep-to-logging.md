
Title: [Task] Extract `timedStep` to a middleware/logging helper

Type: Task

Description:
`timedStep` measures and logs durations inline in CLI code. Extract it into a reusable helper (e.g., `internal/logging`) with structured logging and optional metrics hooks.

Location:
- File: `internal/cli/root.go`
- Approx line: 250

Steps to reproduce (current behavior):
1. `timedStep` is declared in `root.go` and used to time multiple step functions.

Expected result (after change):
- `timedStep` lives in a shared package, supporting structured logs and optional metric callbacks.

Environment:
- CLI runtime

Checks:
- [ ] I searched existing issues
- [ ] I scoped the helper for metrics/structured logging

Suggested labels: enhancement, refactor, logging
