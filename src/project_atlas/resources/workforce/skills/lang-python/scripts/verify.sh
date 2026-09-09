#!/bin/sh
set -eu
echo "=== Verifying Python Strict Typing & Ruff Quality Architecture ==="
atlas tool check-escape-hatches .
exit 0
