#!/bin/sh
set -eu
echo "=== Verifying Modern JavaScript ES2024+ & Security Discipline ==="
atlas tool check-escape-hatches .
exit 0
