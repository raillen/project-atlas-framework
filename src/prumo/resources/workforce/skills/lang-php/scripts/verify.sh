#!/bin/sh
set -eu
echo "=== Verifying PHP Strict Types & Maximum Static Analysis ==="
prumo tool check-escape-hatches .
exit 0
