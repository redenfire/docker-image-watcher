#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd -- "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

# shellcheck source=scripts/load-opencode-env.sh
. "$SCRIPT_DIR/load-opencode-env.sh" ".env"

echo
echo "Checking OpenCode resolved config..."
opencode debug config

echo
echo "Starting OpenCode..."
opencode
