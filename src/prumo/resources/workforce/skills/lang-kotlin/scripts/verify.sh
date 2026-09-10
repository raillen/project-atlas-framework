#!/bin/sh
set -eu
echo "=== Verifying Kotlin Coroutines & Sound Null-Safety ==="
prumo tool check-escape-hatches .
exit 0
