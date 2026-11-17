"""Check that each of the seven proof sections of the findings report has one closing line.

Run as: python3 -I spikes/closing.py <report.md>
"""

import re
import sys

HEADING = re.compile(r"^## Proof (\d+)\b")
CLOSING = re.compile(r"^(Held\.|Fallback applies: )")
FENCE = re.compile(r"^(```|~~~)")


def sections(text):
    """Map each proof number to the closing lines found in its section."""
    found = {}
    current = None
    in_fence = False
    for line in text.splitlines():
        if FENCE.match(line):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        if line.startswith("## "):
            m = HEADING.match(line)
            current = int(m.group(1)) if m else None
            if current is not None:
                found.setdefault(current, [])
            continue
        if current is not None and CLOSING.match(line):
            found[current].append(line)
    return found


def main(argv):
    if len(argv) != 2:
        print("usage: closing.py <report.md>", file=sys.stderr)
        return 2
    with open(argv[1], encoding="utf-8") as f:
        found = sections(f.read())
    bad = []
    for n in range(1, 8):
        count = len(found.get(n, []))
        if n not in found:
            bad.append(f"Proof {n}: no section")
        elif count != 1:
            bad.append(f"Proof {n}: {count} closing lines")
    if bad:
        print("\n".join(bad))
        return 1
    print("7 sections, one closing line each")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
