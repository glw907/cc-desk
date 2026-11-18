"""PermissionRequest hook for proof 4's answer test.

It logs every request it sees to hook.log, waits up to 120 seconds for a file
under decisions/ that names the tool and command of the request, and prints an
allow decision if one appears. In every other case it prints a deny decision.

The decision file is decisions/<sha1 of "<tool>\n<command>">.txt. Set
DECIDE_DIR and DECIDE_WAIT to override the folder and the wait in a test.
"""

import hashlib
import json
import os
import sys
import time

HERE = os.path.dirname(os.path.abspath(__file__))
DIR = os.environ.get("DECIDE_DIR", HERE)
WAIT = float(os.environ.get("DECIDE_WAIT", "120"))


def key(tool, command):
    return hashlib.sha1(f"{tool}\n{command}".encode()).hexdigest()


def decide(req):
    tool = req.get("tool_name", "")
    command = (req.get("tool_input") or {}).get("command", "")
    with open(os.path.join(DIR, "hook.log"), "a") as log:
        log.write(f"{time.strftime('%H:%M:%S', time.gmtime())} request {tool} {command!r}\n")
    path = os.path.join(DIR, "decisions", key(tool, command) + ".txt")
    deadline = time.time() + WAIT
    while time.time() < deadline:
        if os.path.exists(path):
            return "allow"
        time.sleep(1)
    return "deny"


def main():
    req = json.load(sys.stdin)
    behavior = decide(req)
    print(json.dumps({"hookSpecificOutput": {
        "hookEventName": "PermissionRequest",
        "decision": {"behavior": behavior},
    }}))


if __name__ == "__main__":
    main()
