
Title: [Feature] Refactor stack selection into dependency-aware flow

Type: Feature

Description:
Currently the CLI prompts for frontend, backend, and database stacks using static lists. Some stack combinations may be incompatible. Refactor selection so options are dependency-aware (e.g., after selecting a backend, only compatible databases are shown).

Location:
- File: `internal/cli/root.go`
- Approx line: 140

Steps to reproduce (current behavior):
1. Run `taco init` in interactive mode.
2. Select frontend and backend from static lists.
3. Observe database options are not filtered by backend capabilities.

Expected result (after change):
- Stack selection UI presents only compatible options (backend -> allowed DBs, etc.).
- Less chance of generating incompatible scaffolding.

Environment:
- CLI interactive flow (survey prompts)

Checks:
- [ ] I searched existing issues
- [ ] I included a description and proposed changes

Suggested labels: enhancement, refactor, cli
