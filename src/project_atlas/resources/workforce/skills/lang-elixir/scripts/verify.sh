#!/bin/sh
set -eu
echo "=== Verifying Elixir OTP Fault Tolerance & Property Testing ==="
atlas tool check-escape-hatches .
exit 0
