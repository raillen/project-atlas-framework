#!/bin/sh
set -eu
echo "=== Verifying C# & .NET Engineering and Memory Efficiency ==="
atlas tool check-escape-hatches .
exit 0
