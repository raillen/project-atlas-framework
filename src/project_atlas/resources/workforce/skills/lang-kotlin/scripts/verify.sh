#!/bin/sh
set -eu
echo "=== Verifying Kotlin Coroutines & Sound Null-Safety ==="
atlas tool check-escape-hatches .
exit 0
