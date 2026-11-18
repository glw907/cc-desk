"""Read the weekly usage figure and its reset time from `claude -p "/usage"` JSON output.

Run as: python3 -I read.py <usage.json>
Run with --fixture to check the parser on a recorded shape.
"""

import json
import re
import sys

LINE = re.compile(r"Current week \(all models\): (\d+)% used · resets (.+)")

FIXTURE = {
    "total_cost_usd": 0,
    "usage": {"input_tokens": 0, "output_tokens": 0},
    "result": (
        "You are currently using your subscription to power your Claude Code usage\n\n"
        "Current week (all models): 87% used · resets Nov 20, 11pm (America/Anchorage)\n"
        "Current week (Fable): 81% used · resets Nov 20, 11pm (America/Anchorage)\n"
    ),
}


def read(doc):
    """Return (percent used, reset text, model tokens, cost) from a /usage result."""
    m = LINE.search(doc["result"])
    if not m:
        raise ValueError("no 'Current week (all models)' line")
    usage = doc.get("usage", {})
    tokens = sum(usage.get(k, 0) for k in ("input_tokens", "output_tokens", "cache_creation_input_tokens", "cache_read_input_tokens"))
    return int(m.group(1)), m.group(2).strip(), tokens, doc.get("total_cost_usd", 0)


def main(argv):
    doc = FIXTURE if argv[1:] == ["--fixture"] else json.load(open(argv[1], encoding="utf-8"))
    used, resets, tokens, cost = read(doc)
    print(f"weekly used: {used}%  resets: {resets}  model tokens: {tokens}  cost: {cost}")
    if argv[1:] == ["--fixture"]:
        assert (used, resets, tokens, cost) == (87, "Nov 20, 11pm (America/Anchorage)", 0, 0)
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
