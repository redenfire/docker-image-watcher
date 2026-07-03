#!/usr/bin/env bash
set -euo pipefail

PROJECT_NAME="${1:-}"
FORGEJO_REMOTE="${2:-}"

if [ -z "$PROJECT_NAME" ]; then
  echo "Usage: ./scripts/bootstrap.sh <project-name> [forgejo-remote-url]"
  exit 1
fi

find . -type f \
  -not -path './.git/*' \
  -not -path './.gitnexus/*' \
  -exec sed -i "s/TODO_PROJECT/${PROJECT_NAME}/g" {} +

if [ ! -d .git ]; then
  git init
fi

if [ -n "$FORGEJO_REMOTE" ]; then
  if git remote get-url origin >/dev/null 2>&1; then
    git remote set-url origin "$FORGEJO_REMOTE"
  else
    git remote add origin "$FORGEJO_REMOTE"
  fi
fi

echo "Bootstrap complete for ${PROJECT_NAME}"
