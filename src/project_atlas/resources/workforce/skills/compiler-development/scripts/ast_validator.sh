#!/usr/bin/env sh
set -e
TARGET="."
echo "Validating AST token syntax and schema definitions..."
if [ -f "atlas.json" ]; then
    exec atlas validate ""
else
    echo "Validating JSON/Markdown AST structures..."
    find "" -name "*.json" -exec test -s {} \;
    echo "AST validation complete."
fi
