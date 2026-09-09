#!/bin/sh
set -eu
echo "=== Verifying TypeScript Strict Soundness & Boundary Schemas ==="
atlas tool check-escape-hatches .
exit 0
