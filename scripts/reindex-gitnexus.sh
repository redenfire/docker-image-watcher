#!/usr/bin/env bash
set -euo pipefail

skip_git="false"

if [ "${1:-}" = "--skip-git" ]; then
  skip_git="true"
fi

if [ "$skip_git" != "true" ] && [ ! -d .git ]; then
  echo "No .git directory found. Initialize Git first or rerun with --skip-git for early setup." >&2
  exit 1
fi

if [ "$skip_git" = "true" ]; then
  gitnexus analyze --skip-git
else
  gitnexus analyze
fi

gitnexus status
