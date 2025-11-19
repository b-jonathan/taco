---
title: Helper packages
---

This page documents the primary helper packages used across the project and when to use each.

1. `internal/execx`
   - Purpose: run external commands, capture stdout/stderr, and control working directory and timeouts.
   - Key functions: `RunCmd(ctx, dir, cmd string)` (runs a command line string), `RunShell(ctx, dir, cmdline string)` for explicit shell invocation.
   - Notes: Prefer passing explicit arguments where possible for cross-platform safety; use shell for pasted command lines.

2. `internal/fsutil`
   - Purpose: filesystem helpers and template rendering.
   - Key functions: `EnsureFile`, `WriteFile`, `WriteMultipleFiles`, `AppendUniqueLines`, `RenderTemplate`.
   - Notes: helpers try to be idempotent (e.g., `AppendUniqueLines`) — add unit tests when modifying behavior.

3. `internal/gh`
   - Purpose: thin wrapper for a `github.Client` stored in `context.Context`.
   - Key functions: `NewClient`, `WithContext`, `FromContext`, `HasClient`, `MustFromContext`.
   - Notes: improve error messages and add tests around missing client behavior.

4. `internal/git`
   - Purpose: wrapper helpers to run `git` commands for initializing, committing, adding remotes, and pushing.
   - Key functions: `InitAndPush` (consider splitting into smaller helpers: `InitRepo`, `CommitAll`, `ConfigureRemote`, `Push`).

5. `internal/nodepkg`
   - Purpose: convenience for creating `package.json` content and `npm` scripts programmatically.

6. `internal/prompt`
   - Purpose: centralized prompt helpers (survey wrappers) so non-interactive flows are supported.

Where to read the code:
- `internal/execx/run.go`
- `internal/fsutil/fsutil.go`
- `internal/gh/gh.go`
- `internal/git/git.go`

Recommended docs additions:
- Add a short README in `internal/execx/README.md` with examples of when to use `RunCmd` vs `RunShell`.
- Add unit tests around `fsutil` behaviors (AppendUniqueLines, WithFileLock).
