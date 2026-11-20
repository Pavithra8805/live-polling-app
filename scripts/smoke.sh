#!/usr/bin/env sh
set -eu

BASE_URL="${1:-}"
if [ -z "$BASE_URL" ]; then
  echo "Usage: ./scripts/smoke.sh https://your-live-api.example [poll-slug]"
  exit 2
fi
POLL_SLUG="${2:-}"

echo "Checking health..."
curl --fail --silent --show-error "$BASE_URL/healthz"
echo

if [ -n "$POLL_SLUG" ]; then
  echo "Fetching live poll..."
  curl --fail --silent --show-error "$BASE_URL/api/polls/$POLL_SLUG/results"
  echo
fi
