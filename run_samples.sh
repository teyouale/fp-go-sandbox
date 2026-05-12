#!/bin/bash
set -euo pipefail

# Run fp-go samples against the live sandbox
# Usage: ./run_samples.sh [sample_file]
#   Run all:    ./run_samples.sh
#   Run one:    ./run_samples.sh samples/04_currying.go

ENDPOINT="${SANDBOX_URL:-https://fpgo-sandbox.fly.dev}"

run_sample() {
    local file="$1"
    local name=$(basename "$file" .go)

    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "▶ Running: $name"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    # Read the file and escape for JSON
    local code
    code=$(cat "$file")

    # Use python to safely JSON-encode the code
    local json_payload
    json_payload=$(python3 -c "
import json, sys
code = sys.stdin.read()
print(json.dumps({
    'sandbox': 'go',
    'command': 'run',
    'files': {'': code}
}))
" <<< "$code")

    local response
    response=$(curl -s -X POST "$ENDPOINT/v1/exec" \
        -H "Content-Type: application/json" \
        -d "$json_payload")

    # Parse response
    local ok stdout stderr duration
    ok=$(echo "$response" | python3 -c "import json,sys; print(json.load(sys.stdin).get('ok', False))")
    stdout=$(echo "$response" | python3 -c "import json,sys; print(json.load(sys.stdin).get('stdout', ''))")
    stderr=$(echo "$response" | python3 -c "import json,sys; print(json.load(sys.stdin).get('stderr', ''))")
    duration=$(echo "$response" | python3 -c "import json,sys; print(json.load(sys.stdin).get('duration', 0))")

    if [ "$ok" = "True" ]; then
        echo "$stdout"
        echo ""
        echo "✅ OK (${duration}ms)"
    else
        echo "❌ FAILED (${duration}ms)"
        if [ -n "$stderr" ]; then
            echo "$stderr"
        fi
        if [ -n "$stdout" ]; then
            echo "$stdout"
        fi
    fi
    echo ""
}

# Run specified file or all samples
if [ $# -gt 0 ]; then
    run_sample "$1"
else
    for file in samples/*.go; do
        run_sample "$file"
    done
fi
