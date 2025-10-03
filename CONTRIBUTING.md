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

## Adding a New Stack

Taco’s stack system is **registry-based**, meaning each stack is a self-contained module registered into the CLI.

To add a new stack:

1. Implement the `Stack` interface under `internal/stacks/`.
2. Add template files and generation logic for the new stack.
3. Update or add tests and documentation.
4. Submit a pull request describing the new stack and its intended use case.
