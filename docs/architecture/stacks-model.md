---
title: Stacks model
---

## Stacks model

Stacks are the modular generators used by `taco`. Each stack implements the `stacks.Stack` interface and may optionally implement `stacks.Seeder`.

Key methods:

- `Type()` — returns stack type (frontend/backend/database)
- `Name()` — canonical stack name
- `Init(ctx, opts)` — initialize (install tooling, scaffold files)
- `Generate(ctx, opts)` — generate source files and templates
- `Post(ctx, opts)` — optional finalization (writing env files, gitignore updates)

See `internal/stacks/express/express.go`, `internal/stacks/nextjs/nextjs.go`, and `internal/stacks/mongodb/mongodb.go` for examples.
