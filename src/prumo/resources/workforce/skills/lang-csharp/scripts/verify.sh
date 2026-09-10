#!/bin/sh
set -eu
echo "=== Verifying C# & .NET Engineering and Memory Efficiency ==="
prumo tool check-escape-hatches .
exit 0
