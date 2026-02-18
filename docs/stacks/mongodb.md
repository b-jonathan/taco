---
title: MongoDB stack
---

## MongoDB stack

What it generates:
- Adds a `db/client.ts` template to `backend/src/db` and injects a `/seed` route in `backend/src/index.ts` to wire seeding endpoints.

Compatibility
-------------
- Backend stacks: `express` (recommended) — the MongoDB stack injects code into `backend/src/index.ts` that expects a Node/Express backend.

Key implementation points
-------------------------
- See `internal/stacks/mongodb/mongodb.go`.
- `Init` sets `opts.DatabaseURI`, `Generate` writes the client template and injects a seed route, `Seed` performs a simple insert to verify connectivity, and `Post` appends env entries.

Init(), Generate(), Post(), Seed() details
-----------------------------------------
Init()
- Interactive flow that asks whether to use a local MongoDB (`mongodb://localhost:27017`) or provide an authenticated URI (Atlas/custom).
- If the user supplies a custom URI the code validates the format and offers an "undo" flow to return to the Local choice.
- Stores the chosen URI in `opts.DatabaseURI` (printed to the console for confirmation).

Generate()
- Installs `mongodb` and dev types (`@types/mongodb`) in the backend via npm.
- Writes `backend/src/db/client.ts` from template (`internal/stacks/templates/mongodb/db/client.ts.tmpl`) which contains the DB connection helper (connect/disconnect helpers).
- Reads `backend/src/index.ts` and injects:
	- A DB import line: `import { connectDB } from "./db/client"` (if not present).
	- A `/seed` route from `internal/stacks/templates/mongodb/seed.tmpl` (if not present) that calls `connectDB` and performs a simple insert when triggered.

Post()
- Appends a `MONGODB_URI` entry to `backend/.env` using `fsutil.AppendUniqueLines` in the form:
	- `MONGODB_URI=<your-uri>/<appName>`
- This is idempotent and will not duplicate lines if run multiple times.

Seed()
- Signature: `Seed(ctx context.Context, opts *Options) error`
- Purpose: Run a sanity seeding operation against the configured MongoDB URI to verify connectivity and demonstrate a simple write flow.
- Preconditions:
	- `opts.DatabaseURI` must be set (typically by `Init()`); the function returns an error if it's empty.
- Behavior:
	- Connects to MongoDB using `opts.DatabaseURI` and uses `opts.AppName` as the database name.
	- Pings the server with a 5s timeout to verify connectivity.
	- Creates/uses a collection named `seed_test` and inserts a small document (for example `{ "value": 1 }`).
	- Prints a confirmation message with the database name, collection, and inserted `_id` on success.
- Errors & edge-cases:
	- Returns errors if connection/ping/insert fail. Uses the MongoDB driver errors for diagnostics.
	- Respects the provided `ctx` for cancellation/timeouts.

Validation
- After generation you should see `backend/src/db/client.ts` and `backend/src/index.ts` contains the DB import and a `/seed` route. Run the stack's `Seed()` to verify connectivity to the configured URI; successful runs print the inserted `_id`.

Notes
- Do not commit real credentials. The stack appends an env line; in production use secure secrets storage.
