---
title: git (git helpers)
---

Purpose
-------
`internal/git` wraps `git` CLI commands used during scaffolding (init, commit, remote add, push).

Key APIs
--------
- `InitAndPush(ctx, projectRoot, remoteURL, commitMsg) error` — convenience helper that initializes a repo, commits, configures remote and pushes.

Functions (implementation details)
----------------------------------
- `InitAndPush(ctx context.Context, projectRoot, remoteURL, commitMsg string) error`
	- Purpose: High-level convenience helper used during scaffolding to initialize a git repo, set branch to `main`, stage and commit code, configure remote origin, and push the initial commit.
	- Steps performed (in order):
		1. If `.git` is missing, run `git init`.
		2. Set branch: `git checkout -B main`.
		3. Stage and commit: `git add .` and `git commit -m <commitMsg>`.
		4. Remove existing `origin` (best-effort) and `git remote add origin <remoteURL>`.
		5. Push: `git push -u origin main`.
	- Error / edge-case notes:
		- The function currently calls `execx.RunCmd` with single-string commands in some places (e.g., `"git init"`). With the current `RunCmd` implementation (expects executable + args), those call sites will fail; update callers to pass `"git", "init"` or update `RunCmd` to accept shell lines.
		- Commit message handling should avoid naive string concatenation — pass the message as an argument to avoid quoting issues (`"git", "commit", "-m", commitMsg`).
		- Authentication or network failures during push should be surfaced to the user; the caller should provide guidance when push fails.
	- Suggested refactor:
		- Break `InitAndPush` into smaller functions (`InitRepo`, `SetBranch`, `CommitAll`, `ConfigureRemote`, `Push`) for testability and clearer error contexts.

When to use
-----------
- Use these helpers during scaffolding to create and push initial commits when the user opts into remote creation.

Notes
-----
- Consider splitting `InitAndPush` into smaller functions (`InitRepo`, `CommitAll`, `ConfigureRemote`, `Push`) for testability and reuse.
