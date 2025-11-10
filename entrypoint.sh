#!/bin/sh
set -e

# Run planguard and capture output
OUTPUT=$(/usr/local/bin/planguard "$@")

# If format is SARIF, write to file
if echo "$@" | grep -q "\-format sarif"; then
    echo "$OUTPUT" > planguard-results.sarif
fi

# Always print output to stdout
echo "$OUTPUT"

# Exit with planguard's exit code
exit $?
