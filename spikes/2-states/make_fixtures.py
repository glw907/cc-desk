"""Build fixtures.json from the saved live samples and the recorded texts.

Usage: python3 -I make_fixtures.py   (run from the repo root)

Live samples: proof 1's saved rows and screens (working, lost connection) and this
proof's own run (idle, permission prompt). Recorded texts: the four reason texts in
the plan's mechanism table, copied character for character into `state_json.detail`,
the field they were read from. Their rows are synthetic, with the state and status
shape proof 1 saw for a blocked session.
"""

import json

P1 = "spikes/1-wake/out/"
P2 = "spikes/2-states/out/"


def rows(path):
    return [json.loads(x) for x in open(path) if x.strip()]


def one(path):
    return json.loads(open(path).read())


def text(path):
    return open(path).read().strip()


p1 = rows(P1 + "rows.txt")  # cut1..cut5, blocked, end
cases = [
    {"name": "working, proof 1 cut1", "row": p1[0], "expect": "working"},
    {"name": "working, proof 1 cut2 (60 s later)", "row": p1[1], "expect": "working"},
    {"name": "idle, own run save 1", "row": one(P2 + "row-idle-1.json"),
     "logs": text(P2 + "logs-idle-1.txt"), "expect": "idle"},
    {"name": "idle, own run save 2 (60 s later)", "row": one(P2 + "row-idle-2.json"),
     "logs": text(P2 + "logs-idle-2.txt"), "expect": "idle"},
    {"name": "permission prompt, save 1", "row": one(P2 + "row-prompt-1.json"),
     "logs": text(P2 + "logs-prompt-1.txt"), "expect": "permission prompt"},
    {"name": "permission prompt, save 2 (60 s later)", "row": one(P2 + "row-prompt-2.json"),
     "logs": text(P2 + "logs-prompt-2.txt"), "expect": "permission prompt"},
    {"name": "lost connection, proof 1 at the block", "row": p1[5],
     "logs": text(P1 + "logs-blocked.txt"), "expect": "lost connection"},
    {"name": "lost connection, proof 1 ten minutes later", "row": p1[6],
     "logs": text(P1 + "logs-end.txt"), "expect": "lost connection"},
]
blocked_row = {"id": "fixture", "kind": "background", "state": "blocked", "status": "idle"}
for name, detail, expect in [
    ("EAI_AGAIN, 2026-11-14", "API Error: Can't reach the API server — check your internet or DNS (EAI_AGAIN)",
     "lost connection"),
    ("expired sign-in, 2026-11-17", "Login expired", "expired sign-in"),
    ("weekly limit", "You've hit your weekly limit · resets Nov 6, 11pm (America/Anchorage)", "usage limit"),
    ("Fable limit", "You've hit your Fable limit · resets Nov 20, 11pm (America/Anchorage)", "usage limit"),
]:
    cases.append({"name": f"recorded text: {name}", "row": blocked_row,
                  "state_json": {"detail": detail}, "expect": expect})
    cases.append({"name": f"same row, no state.json: {name}", "row": blocked_row,
                  "expect": "can't tell"})

with open("spikes/2-states/fixtures.json", "w") as f:
    json.dump(cases, f, indent=1, ensure_ascii=False)
    f.write("\n")
print(len(cases), "cases")
