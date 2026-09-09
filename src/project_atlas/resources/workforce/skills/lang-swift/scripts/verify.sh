#!/bin/sh
set -eu
echo "=== Verifying Swift 6 Strict Concurrency & Value Semantics ==="
atlas tool check-escape-hatches .
exit 0
