#!/bin/sh
set -eu
echo "=== Verifying Go Idiomatic Reliability & Concurrency Hygiene ==="
atlas tool check-escape-hatches .
exit 0
