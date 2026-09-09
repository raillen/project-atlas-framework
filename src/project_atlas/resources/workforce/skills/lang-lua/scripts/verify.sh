#!/bin/sh
set -eu
echo "=== Verifying Lua Scope Protection & Sandboxing ==="
atlas tool check-escape-hatches .
exit 0
