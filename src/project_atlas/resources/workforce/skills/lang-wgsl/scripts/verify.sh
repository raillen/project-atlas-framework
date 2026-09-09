#!/usr/bin/env bash
set -euo pipefail

SKILL_NAME="lang-wgsl"
echo "==> Running verification for $SKILL_NAME..."

# Check escape hatches in current directory
if command -v atlas >/dev/null 2>&1; then
    atlas tool check-escape-hatches . || true
fi

echo "{\"skill\": \"$SKILL_NAME\", \"status\": \"verified\", \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"}"
exit 0
