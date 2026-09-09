#!/bin/sh
set -eu

echo "=== Verifying Dlang Safety & Scope Pointers ==="

echo "[1/2] Scanning for unregistered escape hatches..."
atlas tool check-escape-hatches .

echo "[2/2] Dlang verification complete."
exit 0
