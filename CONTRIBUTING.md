# Contributing to Taco

We’re excited you’re interested in contributing to Taco! This guide covers installation, local development, and how to contribute.

## Installation

### Prerequisites
- Go 1.25 or higher
- Git installed
- Make (optional but recommended)

### Install Taco via Go
```bash
go install github.com/OWNER/taco/cmd/taco@latest


The binary will be placed in $(go env GOPATH)/bin.

Run Taco from source
git clone https://github.com/OWNER/taco.git
cd taco
```
## Development Setup

### Install dependencies:
```bash
go mod download
```

### Run the linter:
```bash
golangci-lint run
```
or
```bash
make lint
```

### Build the CLI:
```bash
go build -o bin/taco ./cmd/taco
```
or
```bash
make build
```

## Contributing Guidelines

- Create a feature branch from `main` (e.g., `feat/add-postgres-stack`).
- Use [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `chore:`) to keep history clean and readable.
- Include tests where applicable.
- Run linting and tests before submitting a pull request.
- Ensure your changes follow the existing project structure and registry-based stack system.

---

## Issue Labels

When opening issues or reviewing pull requests, please apply the most relevant label(s) to help maintainers triage and prioritize them. Here’s a basic guide:

| Label | When to Use |
|-------|-------------|
| `bug` | A reproducible error, crash, or incorrect behavior in Taco. |
| `feature` | A new capability or enhancement that does not exist yet. |
| `enhancement` | Improvements to existing features, performance, or developer experience. |
| `docs` | Documentation updates, typos, or README/CONTRIBUTING changes. |
| `refactor` | Code cleanup or restructuring without changing behavior. |
| `ci` | Issues related to GitHub Actions, testing, or build pipelines. |
| `good first issue` | Small, well-scoped issues suitable for new contributors. |

**Tips:**
- Most issues should have **one primary label**.
- If an issue touches multiple areas (e.g., a feature that also needs docs), feel free to add more than one.
- PRs should ideally match the label of the issue they close.
  
## Adding a New Stack

Taco’s stack system is **registry-based**, meaning each stack is a self-contained module registered into the CLI.

To add a new stack:

1. Implement the `Stack` interface under `internal/stacks/`.
2. Add template files and generation logic for the new stack.
3. Update or add tests and documentation.
4. Submit a pull request describing the new stack and its intended use case.
