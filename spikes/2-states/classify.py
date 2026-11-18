"""Label a background session with one of six states, and name the field the label rests on.

Usage: python3 -I classify.py fixtures.json

The input is a JSON list of cases. A case holds a row (one entry of
`claude agents --json --all`), and may hold `logs` (the screen text from
`claude logs <id>`, owned sessions only) and `state_json` (a session's job
`state.json`, an undocumented file, owned sessions only). Each case also holds the
expected label. The script prints one line per case and exits 1 on any wrong label.

States: working, idle, lost connection, permission prompt, expired sign-in,
usage limit. Anything else is "can't tell".
"""

import json
import re
import sys

REASONS = [
    ("expired sign-in", re.compile(r"Login expired")),
    ("usage limit", re.compile(r"You've hit your .{0,30}limit")),
    ("lost connection", re.compile(
        r"Can't reach the API server|EAI_AGAIN|Couldn't connect through your proxy|ERR_PROXY_TUNNEL|proxy refused the tunnel")),
]


def label(case):
    row = case["row"]
    state = row.get("state")
    if state == "working":
        return "working", "row.state"
    if state == "done" and row.get("status") == "idle":
        return "idle", "row.state and row.status"
    if row.get("waitingFor") == "permission prompt":
        return "permission prompt", "row.waitingFor"
    sj = case.get("state_json") or {}
    for field in ("detail", "needs"):
        text = sj.get(field) or ""
        for name, pattern in REASONS:
            if pattern.search(text):
                return name, f"state.json {field} (undocumented file)"
    logs = case.get("logs") or ""
    for name, pattern in REASONS:
        if pattern.search(logs):
            return name, "claude logs text"
    return "can't tell", "none"


def main(path):
    cases = json.load(open(path))
    wrong = 0
    for case in cases:
        got, source = label(case)
        ok = got == case["expect"]
        wrong += not ok
        print(f"{'ok  ' if ok else 'FAIL'} {case['name']}: {got} (expected {case['expect']}) <- {source}")
    sys.exit(1 if wrong else 0)


if __name__ == "__main__":
    main(sys.argv[1])
