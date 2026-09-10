#!/bin/sh
set -eu
echo "=== Verifying Swift 6 Strict Concurrency & Value Semantics ==="
prumo tool check-escape-hatches .
exit 0
