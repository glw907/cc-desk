#!/usr/bin/env python3
# The limits phase's stand-in claude: a copy of launchClaudeStandIn in e2e/launch_rig_test.go,
# with the rig folder taken from this file's own folder. Its controls are files in that folder:
# no-id, no-row, needs-prompt, fail-bg, stop-fails, and lists/<n>.json.
import json, os, sys

d = os.path.dirname(os.path.abspath(__file__))
a = sys.argv[1:]


def path(*p):
    return os.path.join(d, *p)


def has(name):
    return os.path.exists(path(name))


with open(path("calls.log"), "a") as f:
    f.write(json.dumps({"args": a, "cwd": os.getcwd()}) + "\n")


def load():
    with open(path("rows.json")) as f:
        return json.load(f)


def save(rows):
    with open(path("rows.json.tmp"), "w") as f:
        json.dump(rows, f)
    os.replace(path("rows.json.tmp"), path("rows.json"))


def bump(name):
    n = 0
    if has(name):
        with open(path(name)) as f:
            n = int(f.read() or 0)
    n += 1
    with open(path(name), "w") as f:
        f.write(str(n))
    return n


cmd = a[0] if a else ""
if cmd == "agents":
    n = bump("reads")
    p = path("lists", "%d.json" % n)
    if not os.path.exists(p):
        p = path("rows.json")
    with open(p) as f:
        sys.stdout.write(f.read())
elif cmd == "--bg":
    if has("fail-bg"):
        sys.stderr.write("claude: could not start the session\n")
        sys.exit(1)
    name = a[a.index("--name") + 1] if "--name" in a else ""
    n = bump("counter")
    sid = "b%07d" % n
    full = sid + "-0000"
    if not has("no-row"):
        row = {"pid": 4000 + n, "id": sid, "sessionId": full, "name": name,
               "cwd": os.getcwd(), "kind": "background",
               "startedAt": 1794970000000, "status": "working", "state": "working"}
        if has("needs-prompt"):
            row["needs"] = "send a prompt to start"
        rows = load()
        rows.append(row)
        save(rows)
    os.makedirs(path("jobs", sid), exist_ok=True)
    with open(path("jobs", sid, "state.json"), "w") as f:
        json.dump({"sessionId": full, "name": name}, f)
    if not has("no-id"):
        sys.stdout.write("backgrounded · %s · %s\n" % (sid, name))
elif cmd == "stop":
    if has("stop-fails"):
        sys.stderr.write("claude: could not stop the session\n")
        sys.exit(1)
    rows = load()
    for r in rows:
        if r["id"] == a[1]:
            r["state"], r["status"] = "stopped", "stopped"
    save(rows)
elif cmd == "logs":
    pass
