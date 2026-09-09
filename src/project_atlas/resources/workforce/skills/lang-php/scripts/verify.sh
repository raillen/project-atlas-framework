#!/bin/sh
set -eu
echo "=== Verifying PHP Strict Types & Maximum Static Analysis ==="
atlas tool check-escape-hatches .
exit 0
