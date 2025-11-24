#!/usr/bin/env bash
set -euo pipefail

echo "==> Checking development dependencies..."

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

check_cmd() {
    if ! command -v "$1" >/dev/null 2>&1; then
        fail "$1 is not installed or not in PATH"
    else
        ok "$1 is installed"
    fi
}

echo ""
echo "-- Checking required tools --"

check_cmd go
check_cmd lefthook
check_cmd golangci-lint

echo ""
ok "All required dependencies are installed."
