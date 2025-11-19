---
title: gh (GitHub client wrapper)
---

Purpose
-------
`internal/gh` is a thin wrapper that manages a `github.Client` stored in a `context.Context`.

Key APIs
--------
- `NewClient(ctx context.Context, token string) *github.Client` — create a client using an OAuth token.
- `WithContext(ctx context.Context, c *github.Client) context.Context` — store client in context.
- `FromContext(ctx context.Context) (*github.Client, error)` — retrieve client or error.
- `HasClient(ctx context.Context) bool` — check presence.
- `MustFromContext(ctx context.Context) *github.Client` — retrieve client or panic if missing.

Functions (implementation details)
----------------------------------
- `NewClient(ctx context.Context, token string) *github.Client`
  - Purpose: Build a `github.Client` using oauth2 with a static token.
  - Notes: Uses `oauth2.StaticTokenSource` and `oauth2.NewClient` so the returned client will include authentication on requests.

- `WithContext(ctx context.Context, c *github.Client) context.Context`
  - Purpose: Store the client value in a derived context so downstream code can fetch it with `FromContext`.

- `FromContext(ctx context.Context) (*github.Client, error)`
  - Purpose: Retrieve the stored client; returns an error when missing (good for user-facing flows that should handle absent auth).

- `HasClient(ctx context.Context) bool`
  - Purpose: Quick presence check useful for code paths that optionally act on GitHub when a client is available.

- `MustFromContext(ctx context.Context) *github.Client`
  - Purpose: Convenience accessor that panics if the client is not present. Use in tests or places where absence is a programming error.

Usage pattern
-------------
- Typical pattern: `client := gh.NewClient(ctx, token); ctx = gh.WithContext(ctx, client);` then downstream functions call `gh.FromContext(ctx)`.

Notes
-----
- Improve error messages and add unit tests to ensure missing-client cases are handled gracefully.
