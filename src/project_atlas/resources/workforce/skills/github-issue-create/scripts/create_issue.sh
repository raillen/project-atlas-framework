#!/bin/bash
# scripts/create_issue.sh
# Creates a GitHub issue using the `gh` CLI.

set -e

if [ "$#" -lt 2 ]; then
    echo "Usage: $0 <title> <body_file> [label1,label2,...] [assignee]"
    exit 1
fi

TITLE="$1"
BODY_FILE="$2"
LABELS="${3:-}"
ASSIGNEE="${4:-}"

if [ ! -f "$BODY_FILE" ]; then
    echo "Error: Body file '$BODY_FILE' not found."
    exit 1
fi

CMD="gh issue create --title \"$TITLE\" --body-file \"$BODY_FILE\""

if [ -n "$LABELS" ]; then
    CMD="$CMD --label \"$LABELS\""
fi

if [ -n "$ASSIGNEE" ]; then
    CMD="$CMD --assignee \"$ASSIGNEE\""
fi

echo "Creating issue: $TITLE"
eval $CMD
echo "Issue created successfully."
