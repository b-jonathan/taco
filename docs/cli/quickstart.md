---
title: Quickstart
---

## Quickstart

Prerequisites:
- Go (for building `taco` from source)
- Node.js and npm (for generated stacks)
- Optional: `GITHUB_TOKEN` environment variable if you plan to create remote repos

Minimal example (interactive):

```bash
# build and run from repo root
go build -o taco ./cmd/taco
./taco init
```

Non-interactive example (scripted):

```bash
# example using environment variables and flags
GITHUB_TOKEN=... ./taco init myproject --private --remote=ssh --description="Example"
```

If `create-next-app` or other scaffold commands fail on Windows, see `docs/cli/troubleshooting.md` for tips.
