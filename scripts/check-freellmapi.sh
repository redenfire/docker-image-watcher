#!/usr/bin/env bash
set -euo pipefail

if [ -f .env ]; then
  set -a
  # shellcheck disable=SC1091
  . ./.env
  set +a
fi

base_url="${FREELLMAPI_BASE_URL:-http://127.0.0.1:3001/v1}"
api_key="${FREELLMAPI_API_KEY:-}"

if [ -z "$api_key" ]; then
  echo "FREELLMAPI_API_KEY is not set. Add it to .env or export it in your shell."
  exit 1
fi

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required."
  exit 1
fi

echo "Checking FreeLLMAPI models endpoint: $base_url/models"
curl -fsS "$base_url/models" \
  -H "Authorization: Bearer $api_key" \
  -H "Content-Type: application/json" >/tmp/freellmapi-models.json

echo "Checking FreeLLMAPI chat endpoint with model=auto"
curl -fsS "$base_url/chat/completions" \
  -H "Authorization: Bearer $api_key" \
  -H "Content-Type: application/json" \
  -d '{"model":"auto","messages":[{"role":"user","content":"Reply with OK only."}],"stream":false}' >/tmp/freellmapi-chat.json

echo "FreeLLMAPI check passed."
