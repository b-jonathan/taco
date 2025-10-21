
Title: [Task] Harden and document `internal/gh` wrapper

Type: Task

Description:
`internal/gh` is a small wrapper around the GitHub client. Improve resiliency, add tests, and document its usage and lifecycle.

Location:
- File: `internal/gh/gh.go`
- Approx line: 11

Steps to reproduce (current behavior):
1. Call `FromContext` on a context without a client -> returns generic error.
2. There are no tests validating behavior.

Expected result (after change):
- Clear error types/messages, unit tests for context helpers, and documentation on how to attach a client to a context.

Environment:
- Codebase + unit test environment

Checks:
- [ ] I searched existing issues
- [ ] I added a test plan for context helper behavior

Suggested labels: maintenance, tests, docs
