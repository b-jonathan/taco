
Title: [Task] Clean up MongoDB `.env` handling in `mongodb.Post`

Type: Task

Description:
`mongodb.Post` appends `MONGODB_URI` to `backend/.env` using ad-hoc formatting. Make this idempotent, well-tested, and readable.

Location:
- File: `internal/stacks/mongodb/mongodb.go`
- Approx line: 200

Steps to reproduce (current behavior):
1. Run `taco init` with MongoDB selected.
2. `Post` appends a formatted `MONGODB_URI` to `backend/.env`.

Expected result (after change):
- A helper writes env entries idempotently, avoids duplicates, and preserves file formatting.

Environment:
- Local filesystem where `taco init` writes scaffolding

Checks:
- [ ] I searched existing issues
- [ ] I added a test plan for env writing behavior

Suggested labels: maintenance, tests, refactor
