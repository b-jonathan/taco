
Title: [Refactor] Split `InitAndPush` into smaller git operations

Type: Task

Description:
`InitAndPush` currently does init, commit, remote config, and push. Break it into smaller functions (`InitRepo`, `SetBranch`, `CommitAll`, `ConfigureRemote`, `Push`) for testability and reusability.

Location:
- File: `internal/git/git.go`
- Approx line: 12

Steps to reproduce (current behavior):
1. `git.InitAndPush(ctx, projectRoot, remoteURL, commitMsg)` runs the full flow.

Expected result (after change):
- Smaller composable functions with unit tests; `InitAndPush` remains as a convenience wrapper.

Environment:
- Requires `git` available in PATH for integration tests; unit tests should mock `execx`.

Checks:
- [ ] I searched existing issues
- [ ] I included a testing strategy for mocking exec calls

Suggested labels: refactor, tests, git
