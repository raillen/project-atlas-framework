#!/bin/sh
set -eu
echo "=== Verifying Java LTS Modern Engineering & Concurrency Discipline ==="
prumo tool check-escape-hatches .
exit 0
