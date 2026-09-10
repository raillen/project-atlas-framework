#!/bin/sh
set -eu
echo "=== Verifying Bash Strict Defensive Engineering ==="
prumo tool check-escape-hatches .
exit 0
