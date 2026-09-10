#!/bin/sh
set -eu
echo "=== Verifying Ruby Gradual Typing & Defensive Architecture ==="
prumo tool check-escape-hatches .
exit 0
