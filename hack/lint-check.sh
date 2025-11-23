#!/usr/bin/env bash
set -euo pipefail

echo "==> Running lint checks..."

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

prompt_fix() {
    # If CI is running, never prompt interactively
    if [[ -n "${CI:-}" ]]; then
        echo -e "${RED}✗ $1${NC}"
        echo -e "${YELLOW}→ Run './hack/lint-fix.sh' locally to fix issues.${NC}"
        exit 1
    fi

    echo -e "${RED}✗ $1${NC}"
    echo ""
    read -p "Would you like to run './hack/lint-fix.sh' now? (y/N): " answer

    case "$answer" in
        [yY]|[yY][eE][sS])
            echo ""
            echo "==> Running lint-fix..."
            ./hack/lint-fix.sh
            exit $?;;
        *)
            echo -e "${YELLOW}→ Skipping auto-fix. You can run './hack/lint-fix.sh' later.${NC}"
            exit 1;;
    esac
}

ok() {
    echo -e "${GREEN}✓ $1${NC}"
}

echo ""
echo "-- Checking go vet --"
if go vet ./... >/dev/null 2>&1; then
    ok "go vet passed"
else
    prompt_fix "go vet failed"
fi

echo ""
echo "-- Checking go fmt --"

UNFORMATTED=$(gofmt -l .)

if [[ -n "$UNFORMATTED" ]]; then
    echo "$UNFORMATTED"
    prompt_fix "Some files are not gofmt-formatted"
else
    ok "go fmt passed"
fi

echo ""
echo "-- Running golangci-lint --"
if golangci-lint run --new-from-rev=HEAD~ --timeout=5m; then
    ok "golangci-lint passed"
else
    prompt_fix "golangci-lint found issues"
fi

echo ""
ok "All lint checks passed."
