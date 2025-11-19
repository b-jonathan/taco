---
title: How to add a stack
---

This guide walks through adding a new stack implementation.

1. Create a new package under `internal/stacks/<yourstack>` implementing `stacks.Stack`.
2. Add templates under `internal/stacks/templates/<yourstack>` and render them with `fsutil.RenderTemplate`.
3. Register the stack in the factory (see `internal/stacks/registry.go` or similar).
4. Add unit tests that run `Generate` into a temp dir and assert files exist.
5. Update `docs/stacks/<yourstack>.md` with a summary of generated artifacts.
