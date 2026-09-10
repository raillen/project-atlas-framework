#!/bin/sh
set -eu

echo "=== Verifying x86-64 Assembly Safety & Directives ==="

echo "[1/2] Checking for .note.GNU-stack in assembly files..."
find . -type f \( -name "*.s" -o -name "*.asm" \) | while read -r f; do
    if ! grep -q ".note.GNU-stack" "$f"; then
        echo "Warning: Missing .note.GNU-stack directive in $f"
    fi
done

echo "[2/2] Assembly security checks complete."
exit 0
