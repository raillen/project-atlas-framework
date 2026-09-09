#!/bin/sh
set -eu
echo "=== Verifying Java LTS Modern Engineering & Concurrency Discipline ==="
atlas tool check-escape-hatches .
exit 0
