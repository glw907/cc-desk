"""Group the workstation's denials log by class and tool.

Usage: python3 -I denials.py <denials.jsonl>

Prints one line per class: its row count and a count per tool. The class is
the bracketed text at the start of the reason, or "no reason" when there is
none.
"""

import collections
import json
import re
import sys


def classes(path):
    out = collections.defaultdict(collections.Counter)
    sessions = collections.defaultdict(set)
    with open(path) as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            row = json.loads(line)
            m = re.match(r"\[([^\]]+)\]", row.get("reason") or "")
            cls = m.group(1) if m else "no reason"
            out[cls][row.get("tool", "?")] += 1
            sessions[cls].add(row.get("session"))
    return out, sessions


def main():
    out, sessions = classes(sys.argv[1])
    total = 0
    for cls, tools in sorted(out.items(), key=lambda kv: -sum(kv[1].values())):
        n = sum(tools.values())
        total += n
        detail = ", ".join(f"{c} {t}" for t, c in tools.most_common())
        print(f"{cls}: {n} ({detail}); sessions {len(sessions[cls])}")
    print(f"total: {total}")


if __name__ == "__main__":
    main()
