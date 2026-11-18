"""Count a proof session's prompts and denials.

Usage: python3 -I tally.py <rows.jsonl> <session-id> <denials.jsonl>

rows.jsonl holds one saved session row per line (the output of
claude agents --json --all, one row each). A prompt is a saved row for the
session whose state is "blocked". A denial is a row in the denials log whose
session id matches.
"""

import json
import sys


def tally(rows_path, session, denials_path):
    prompts = []
    with open(rows_path) as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            row = json.loads(line)
            if row.get("sessionId") == session or row.get("id") == session[:8]:
                if row.get("state") == "blocked":
                    prompts.append(row)
    denials = []
    try:
        with open(denials_path) as f:
            for line in f:
                line = line.strip()
                if line and json.loads(line).get("session") == session:
                    denials.append(json.loads(line))
    except FileNotFoundError:
        pass
    return prompts, denials


def main():
    rows_path, session, denials_path = sys.argv[1:4]
    prompts, denials = tally(rows_path, session, denials_path)
    print(f"prompts: {len(prompts)}")
    print(f"denials: {len(denials)}")
    for d in denials:
        print(f"  denial {d.get('tool')} {d.get('reason')} {str(d.get('target'))[:80]!r}")


if __name__ == "__main__":
    main()
