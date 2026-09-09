#!/bin/sh
set -eu
echo "=== Verifying Ruby Gradual Typing & Defensive Architecture ==="
atlas tool check-escape-hatches .
exit 0
