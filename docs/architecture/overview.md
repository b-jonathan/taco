---
title: Overview
---

## System overview

`taco` is a CLI scaffolding tool that:

1. Gathers project parameters from flags and interactive prompts.
2. Selects the requested stacks (frontend, backend, database).
3. For each stack: runs `Init`, `Generate`, and optional `Post` steps.
4. Optionally creates a GitHub repository and pushes the initial commit.

Key implementation points:
- Stacks implement the `stacks.Stack` interface and are registered via a factory.
- Templates live under `internal/stacks/templates` and are rendered with `fsutil.RenderTemplate`.
- External commands are executed through `internal/execx` which captures stdout/stderr.

Where to look in code:
- `internal/cli/root.go` — CLI wiring and main `init` command.
- `internal/stacks/*` — stack implementations.
- `internal/execx` and `internal/fsutil` — helpers for running commands and writing files.
