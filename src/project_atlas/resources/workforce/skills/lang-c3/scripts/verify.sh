#!/bin/sh
set -eu

echo "=== Verifying C3 Safety & Contracts ==="

echo "[1/2] Scanning for unregistered escape hatches..."
atlas tool check-escape-hatches .

echo "[2/2] C3 verification complete."
exit 0
