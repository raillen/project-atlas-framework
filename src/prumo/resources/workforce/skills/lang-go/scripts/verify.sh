#!/bin/sh
set -eu
echo "=== Verifying Go Idiomatic Reliability & Concurrency Hygiene ==="
prumo tool check-escape-hatches .
exit 0
