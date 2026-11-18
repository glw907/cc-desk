"""What the supervisor log shows next for each session in the denials log's Classifier-unavailable rows.

Usage: python3 -I fate.py <denials.jsonl> <daemon.log>

Reads only those two files. For every session id in a row with no reason, it prints
the count of daemon.log lines naming the session's first 8 characters and the last
such line, cut to 160 characters.
"""

import collections
import json
import sys


def main(denials_path, log_path):
    rows = [json.loads(line) for line in open(denials_path) if line.strip()]
    empty = [r for r in rows if r.get("reason", "").startswith("Classifier unavailable")]
    ids = sorted({r["session"] for r in empty})
    log = open(log_path).read().splitlines()
    print(f"Classifier-unavailable rows: {len(empty)}; sessions: {len(ids)}; daemon.log lines: {len(log)}")
    seen = collections.Counter()
    for sid in ids:
        lines = [x for x in log if sid[:8] in x]
        seen[bool(lines)] += 1
        print(sid[:8], len(lines), lines[-1][:160] if lines else "-")
    print("sessions named in daemon.log:", seen[True], "not named:", seen[False])


if __name__ == "__main__":
    main(sys.argv[1], sys.argv[2])
