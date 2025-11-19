---
title: Commands
---

## Commands

Top-level commands provided by the `taco` CLI (implemented under `internal/cli`):

- `taco init [name]` — Create a project scaffold and (optionally) a remote repo.

### `init` flags

- `--private` — make the created repository private
- `--remote` — `ssh` or `https` (remote URL type)
- `--description` — repository description

Examples:

```bash
./taco init myproject
./taco init myproject --private --remote=https --description="My project"
```

For interactive prompts and parameter gathering see `internal/cli/root.go` and `gatherInitParams`.
