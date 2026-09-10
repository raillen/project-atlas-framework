#!/bin/sh
set -eu
echo "=== Verifying Elixir OTP Fault Tolerance & Property Testing ==="
prumo tool check-escape-hatches .
exit 0
