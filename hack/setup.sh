#!/usr/bin/env bash
set -euo pipefail

echo "==> Verifying Go installation..."
if ! command -v go >/dev/null 2>&1; then
    echo "Go is not installed. Install Go 1.XX+ first."
    exit 1
fi

echo "==> Installing Go tools..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

echo "==> Installing Lefthook..."
echo "Tools will be installed in: $(go env GOPATH)/bin"
if ! command -v lefthook >/dev/null 2>&1; then
    go install github.com/evilmartians/lefthook/cmd/lefthook@latest
fi
lefthook install

echo "==> Setup complete!"
echo "Try running 'make check' to confirm everything works."

