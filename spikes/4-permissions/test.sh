#!/usr/bin/env bash
# Fixture tests for tally.py, denials.py, and decide-hook.py.
set -uo pipefail
HERE=$(cd "$(dirname "$0")" && pwd)
fail=0
check() { # check <label> <want> <got>
  if [ "$2" = "$3" ]; then echo "ok   $1"; else echo "FAIL $1: want [$2] got [$3]"; fail=1; fi
}
out=$(python3 -I "$HERE/tally.py" "$HERE/fixtures/rows.jsonl" aaaa1111-0000 "$HERE/fixtures/denials.jsonl" | head -2 | tr '\n' ' ')
check "tally counts one prompt and two denials" "prompts: 1 denials: 2 " "$out"
out=$(python3 -I "$HERE/denials.py" "$HERE/fixtures/denials.jsonl" | tail -1)
check "denials totals three rows" "total: 3" "$out"
out=$(python3 -I "$HERE/denials.py" "$HERE/fixtures/denials.jsonl" | grep -c 'no reason: 1')
check "denials counts the empty reason" "1" "$out"
tmp=$(mktemp -d)
mkdir "$tmp/decisions"
req='{"tool_name":"Bash","tool_input":{"command":"echo hi"}}'
out=$(echo "$req" | DECIDE_DIR=$tmp DECIDE_WAIT=2 python3 -I "$HERE/decide-hook.py" | python3 -I -c 'import json,sys;print(json.load(sys.stdin)["hookSpecificOutput"]["decision"]["behavior"])')
check "hook denies with no decision file" "deny" "$out"
key=$(printf 'Bash\necho hi' | sha1sum | cut -d' ' -f1)
: >"$tmp/decisions/$key.txt"
out=$(echo "$req" | DECIDE_DIR=$tmp DECIDE_WAIT=2 python3 -I "$HERE/decide-hook.py" | python3 -I -c 'import json,sys;print(json.load(sys.stdin)["hookSpecificOutput"]["decision"]["behavior"])')
check "hook allows with a decision file" "allow" "$out"
rm -rf "$tmp"
exit $fail
