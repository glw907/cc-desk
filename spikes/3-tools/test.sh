#!/usr/bin/env bash
# Test tokens.py on a fixture transcript built from the documented usage fields.
set -euo pipefail
HERE=$(cd "$(dirname "$0")" && pwd)
T=$(mktemp -d)
trap 'rm -rf "$T"' EXIT
mkdir -p "$T/p/proj"
printf 'id1\tsid-full\tfull\n' >"$T/owned.tsv"
printf 'id2\tsid-lean\tlean\n' >>"$T/owned.tsv"
cat >"$T/p/proj/sid-full.jsonl" <<'EOF'
{"type":"attachment","attachment":{"type":"deferred_tools_delta","addedNames":["Artifact","Workflow"]}}
{"type":"assistant","message":{"usage":{"input_tokens":10,"cache_creation_input_tokens":20000,"cache_read_input_tokens":5}}}
EOF
cat >"$T/p/proj/sid-lean.jsonl" <<'EOF'
{"type":"assistant","message":{"usage":{"input_tokens":10,"cache_creation_input_tokens":8000,"cache_read_input_tokens":5}}}
EOF
OUT=$(python3 -I "$HERE/tokens.py" "$T/owned.tsv" full lean "$T/p")
echo "$OUT" | python3 -I -c '
import json, sys
d = json.load(sys.stdin)
assert d["full"]["total"] == 20015, d
assert d["lean"]["total"] == 8015, d
assert d["saved"] == 12000, d
assert d["deferred_in_full"]["Artifact"] and d["deferred_in_full"]["Workflow"], d
assert not d["deferred_in_full"]["SendFeedback"], d
'
echo ok
