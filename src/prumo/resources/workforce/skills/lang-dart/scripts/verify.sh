#!/bin/sh
set -eu
echo "=== Verifying Dart Sound Null-Safety & Isolate Discipline ==="
prumo tool check-escape-hatches .
exit 0
