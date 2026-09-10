#!/bin/sh
set -eu

echo "=== Verifying Zig Safety & Allocations ==="

echo "[1/2] Scanning for unregistered escape hatches..."
prumo tool check-escape-hatches .

echo "[2/2] Zig safety checks complete."
exit 0
