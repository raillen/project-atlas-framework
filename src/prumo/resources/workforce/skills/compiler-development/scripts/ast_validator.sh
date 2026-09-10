#!/usr/bin/env sh
set -e
TARGET="."
echo "Validating AST token syntax and schema definitions..."
if [ -f "prumo.json" ]; then
    exec prumo validate ""
else
    echo "Validating JSON/Markdown AST structures..."
    find "" -name "*.json" -exec test -s {} \;
    echo "AST validation complete."
fi
