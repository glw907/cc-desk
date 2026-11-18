"""Sum the first request's input tokens for the two proof 3 sessions.

Usage: python3 -I tokens.py <owned.tsv> <full-name> <lean-name> [projects-dir]

Finds each session's transcript by its session id in owned.tsv, reads the
`usage` of its first assistant record (input, cache creation, cache read
tokens), and from the full session's transcript any record that names
deferred tools. It opens no other transcript and reads no message text
except a deferred-tool record's tool names.
"""
import glob
import json
import os
import re
import sys

KEYS = ("input_tokens", "cache_creation_input_tokens", "cache_read_input_tokens")
TOOLS = ("Artifact", "SendFeedback", "Workflow", "AskUserQuestion", "ReportFindings")


def session_id(owned, name):
    for line in open(owned):
        cols = line.rstrip("\n").split("\t")
        if len(cols) > 2 and cols[2] == name:
            return cols[1]
    raise SystemExit(f"tokens.py: {name} is not in {owned}")


def transcript(projects, sid):
    hits = glob.glob(os.path.join(projects, "*", sid + ".jsonl"))
    if len(hits) != 1:
        raise SystemExit(f"tokens.py: {len(hits)} transcripts for session {sid}")
    return hits[0]


def first_usage(path):
    """Return the summed token count of the first assistant record that has usage."""
    for line in open(path):
        try:
            rec = json.loads(line)
        except ValueError:
            continue
        usage = (rec.get("message") or {}).get("usage")
        if rec.get("type") == "assistant" and usage:
            return sum(usage.get(k, 0) for k in KEYS), {k: usage.get(k, 0) for k in KEYS}
    raise SystemExit(f"tokens.py: no assistant usage in {path}")


def deferred_present(path):
    """Return, per tool, whether a deferred-tools record names it."""
    seen = {t: False for t in TOOLS}
    for line in open(path):
        if "deferred" not in line:
            continue
        try:
            rec = json.loads(line)
        except ValueError:
            continue
        att = rec.get("attachment") or {}
        if "deferred" not in str(att.get("type", "")):
            continue
        names = json.dumps(att)
        for t in TOOLS:
            if re.search(r'"%s"|\b%s\b' % (t, t), names):
                seen[t] = True
    return seen


def main(argv):
    owned, full, lean = argv[1], argv[2], argv[3]
    projects = argv[4] if len(argv) > 4 else os.path.expanduser("~/.claude/projects")
    out = {}
    for label, name in (("full", full), ("lean", lean)):
        path = transcript(projects, session_id(owned, name))
        total, parts = first_usage(path)
        out[label] = {"total": total, **parts}
        if label == "full":
            out["deferred_in_full"] = deferred_present(path)
    out["saved"] = out["full"]["total"] - out["lean"]["total"]
    print(json.dumps(out, indent=2))


if __name__ == "__main__":
    main(sys.argv)
