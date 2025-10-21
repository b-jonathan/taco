
Title: [Task] Refactor `fsutil` for consistency and testability

Type: Task

Description:
`internal/fsutil` contains several utilities that were implemented quickly. Refactor for consistent APIs, better error handling, and add unit tests to prevent regressions.

Location:
- File: `internal/fsutil/fsutil.go`
- Approx line: 12

Steps to reproduce (current behavior):
1. Use functions like `EnsureFile` and `AppendUniqueLines` in stack generators.
2. Observe inconsistent behavior or lack of tests.

Expected result (after change):
- Clean, documented APIs with unit tests covering edge cases (missing files, concurrent writes).

Environment:
- Local development and CI tests

Checks:
- [ ] I searched existing issues
- [ ] I included a test plan for key functions

Suggested labels: maintenance, tests, refactor
