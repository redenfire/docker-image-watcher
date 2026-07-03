#!/usr/bin/env bash
set -euo pipefail

skip_gitnexus="false"
skip_caveman="false"

while [ $# -gt 0 ]; do
  case "$1" in
    --skip-gitnexus)
      skip_gitnexus="true"
      ;;
    --skip-caveman)
      skip_caveman="true"
      ;;
    -h|--help)
      echo "Usage: $0 [--skip-gitnexus] [--skip-caveman]"
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      exit 1
      ;;
  esac
  shift
done

if [ "$skip_gitnexus" != "true" ]; then
  npm install -g gitnexus@latest
fi

if [ "$skip_caveman" != "true" ]; then
  npm install -g @juliusbrussee/caveman-code
fi

echo "Agent workstation tools installation complete."
echo "OpenCode must be installed separately according to docs/OPENCODE.md."
