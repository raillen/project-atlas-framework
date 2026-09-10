#!/bin/sh
set -eu
echo "=== Verifying Python Strict Typing & Ruff Quality Architecture ==="
prumo tool check-escape-hatches .
exit 0
