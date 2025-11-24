#!/usr/bin/env bash
set -euo pipefail

echo "==> Applying lint fixes..."

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

fail() {
    echo -e "${RED}✗ $1${NC}"
    exit 1
}

ok() {
    echo -e "${GREEN}✓ $1${NC}"
}

echo ""
echo "-- Running go fmt (auto-fix formatting) --"
if go fmt ./... >/dev/null 2>&1; then
    ok "go fmt applied"
else
    fail "go fmt encountered errors"
fi

echo ""
echo "-- Running go vet (no auto-fix, but must pass) --"
if go vet ./... >/dev/null 2>&1; then
    ok "go vet passed"
else
    fail "go vet found issues"
fi

echo ""
echo "-- Running golangci-lint in fix mode --"
if golangci-lint run --fix --new-from-rev=HEAD~ --timeout=5m; then
    ok "golangci-lint auto-fix completed"
else
    fail "golangci-lint still reports issues after --fix"
fi

echo ""
ok "Lint fix complete."
