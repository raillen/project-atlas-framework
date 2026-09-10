#!/bin/sh
set -eu
echo "=== Verifying Lua Scope Protection & Sandboxing ==="
prumo tool check-escape-hatches .
exit 0
