
Title: [Enhancement] Expand `gh` and `git` packages for CI/CD and workflow automation

Type: Feature

Description:
Current `internal/gh` and `internal/git` provide minimal operations. Enhance these packages to support repository automation such as creating workflows, managing secrets, and branch protections to enable better CI/CD integration during scaffolding.

Location:
- Files: `internal/gh/gh.go`, `internal/git/git.go`

Steps to reproduce (current behavior):
1. The scaffolder generates project files but doesn't create CI workflows or manage GitHub settings.

Expected result (after change):
- Helpers to create common GitHub Actions workflows, set repository secrets, and configure branch protections as opt-in steps during `taco init`.

Environment:
- Requires GitHub API access (GITHUB_TOKEN) and proper permissions.

Checks:
- [ ] I searched existing issues
- [ ] I considered RBAC/permissions needed for GitHub API calls

Suggested labels: enhancement, infra, git
