
Title: [Feature] Add optional GitHub repository creation in `init`

Type: Feature

Description:
The `init` command scaffolds local project files; the GitHub creation and push flow is currently commented out. Add an opt-in flag (e.g. `--create-remote`) to create a remote repository and push initial commit using `internal/gh` and `internal/git`.

Location:
- File: `internal/cli/root.go`
- Approx lines: 200-220

Steps to reproduce (current behavior):
1. Run `taco init` and complete scaffold steps.
2. No remote repository is created; the GitHub flow is commented out.

Expected result (after change):
- New flag `--create-remote` creates the GitHub repo and pushes the initial commit.
- If push fails, provide clean rollback or clear error guidance.

Environment:
- CLI interactive flow with a GITHUB_TOKEN environment variable present

Checks:
- [ ] I searched existing issues
- [ ] I included expected behavior and rollback considerations

Suggested labels: enhancement, feature, cli
