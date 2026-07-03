#!/usr/bin/env bash
set -euo pipefail

required=(
  "README.md"
  "AGENTS.md"
  "docs/TOOLING_MODEL.md"
  "docs/HOW_TO_USE.md"
  "docs/IMPLEMENTATION_WORKFLOW.md"
  "docs/OPENCODE.md"
  "docs/GITNEXUS.md"
  "docs/CAVEMAN.md"
  "docs/CAVEMAN_GITNEXUS.md"
  "docs/CODEX.md"
  "docs/FREELLMAPI.md"
  "docs/STATUS.md"
  "docs/TASKS.md"
  "memory/PROJECT_BRIEF.md"
  "memory/CONSTRAINTS.md"
  "scripts/start-opencode.sh"
  "scripts/start-opencode.ps1"
  "opencode.json"
  ".cave/settings.json"
  "src/.gitkeep"
)

for f in "${required[@]}"; do
  if [ ! -f "$f" ]; then
    echo "Missing required file: $f" >&2
    exit 1
  fi
done


python3 -m json.tool opencode.json >/dev/null

echo "Scaffold check passed."
