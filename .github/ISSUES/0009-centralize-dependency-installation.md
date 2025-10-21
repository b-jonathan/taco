
Title: [Enhancement] Centralize dependency installation lists for stack generators

Type: Feature

Description:
Stacks currently invoke `npm install` inline with hard-coded lists (see `express`). Centralize dependency lists and an installer helper so stacks declare dependencies and a shared installer performs the work.

Location:
- File: `internal/stacks/express/express.go`
- Approx line: 40

Steps to reproduce (current behavior):
1. Generate an express backend; `Init` runs `npm install` with inline dependency lists.

Expected result (after change):
- Stack implementations return dependency lists; a shared installer handles installs uniformly and can add caching or retry logic.

Environment:
- Node.js/npm environment used by `taco init`

Checks:
- [ ] I searched existing issues
- [ ] I included integration considerations for npm failures

Suggested labels: enhancement, refactor
