#!/usr/bin/env python3
"""A stand-in for the gh commands the check runs, with `pr close`.

This copies check-merge/gh.py and adds `pr close <n> -R <repo>`, which
rewrites the answer as CLOSED.

The rig copies this file into its own folder, which holds the stand-in's
files:

  answers/<branch>.json  the JSON list `pr list --head <branch>` prints
  remote                 the path of the bare remote
  down                   when present, every call exits 1 after logging
  gh-calls.log           one JSON line per call: its argument list
  work/                  the stand-in's own clone of the bare remote

`pr merge <n> ... --match-head-commit <sha>` exits 1 when the sha is not the
answer's head. Otherwise it merges the branch into main on the bare remote
with a real `git merge --no-ff`, pushes, and rewrites the answer as MERGED.
"""
import json
import os
import subprocess
import sys

d = os.path.dirname(os.path.abspath(__file__))
a = sys.argv[1:]

with open(os.path.join(d, "gh-calls.log"), "a") as f:
    f.write(json.dumps(a) + "\n")

if os.path.exists(os.path.join(d, "down")):
    sys.stderr.write("gh: could not reach github.com\n")
    sys.exit(1)


def flag(name):
    return a[a.index(name) + 1] if name in a else ""


def answers():
    out = {}
    folder = os.path.join(d, "answers")
    for name in sorted(os.listdir(folder)):
        with open(os.path.join(folder, name)) as f:
            out[name[: -len(".json")]] = json.load(f)
    return out


def git(*args, cwd):
    # The binary's GIT_TRACE is for its own calls only.
    env = {k: v for k, v in os.environ.items() if k != "GIT_TRACE"}
    env.update(GIT_CONFIG_GLOBAL="/dev/null", GIT_CONFIG_SYSTEM="/dev/null")
    cmd = ["git", "-c", "user.name=Stand-in", "-c", "user.email=gh@example.com",
           "-c", "commit.gpgsign=false"] + list(args)
    p = subprocess.run(cmd, cwd=cwd, env=env, capture_output=True, text=True)
    if p.returncode != 0:
        sys.stderr.write("gh stand-in: git %s: %s\n" % (" ".join(args), p.stderr))
        sys.exit(1)
    return p.stdout.strip()


if a[:2] == ["pr", "list"]:
    branch = flag("--head")
    path = os.path.join(d, "answers", branch + ".json")
    if os.path.exists(path):
        with open(path) as f:
            sys.stdout.write(f.read())
    else:
        sys.stdout.write("[]")
elif a[:2] == ["pr", "merge"]:
    number = int(a[2])
    sha = flag("--match-head-commit")
    for branch, pulls in answers().items():
        for pull in pulls:
            if pull["number"] != number:
                continue
            if pull["headRefOid"] != sha:
                sys.stderr.write("gh: pull request head changed\n")
                sys.exit(1)
            work = os.path.join(d, "work")
            if not os.path.isdir(work):
                with open(os.path.join(d, "remote")) as f:
                    git("clone", f.read().strip(), work, cwd=d)
            git("fetch", "origin", cwd=work)
            git("checkout", "-B", "main", "origin/main", cwd=work)
            git("merge", "--no-ff", "origin/" + branch, "-m", "Merge pull request %d" % number, cwd=work)
            git("push", "origin", "main", cwd=work)
            pull["state"] = "MERGED"
            pull["mergedAt"] = os.environ.get("GH_MERGED_AT", "2026-11-20T23:00:00Z")
            pull["mergeCommit"] = {"oid": git("rev-parse", "HEAD", cwd=work)}
            with open(os.path.join(d, "answers", branch + ".json"), "w") as f:
                json.dump(pulls, f)
            sys.exit(0)
    sys.stderr.write("gh: no pull request %d\n" % number)
    sys.exit(1)
elif a[:2] == ["pr", "close"]:
    number = int(a[2])
    for branch, pulls in answers().items():
        for pull in pulls:
            if pull["number"] == number:
                pull["state"] = "CLOSED"
                with open(os.path.join(d, "answers", branch + ".json"), "w") as f:
                    json.dump(pulls, f)
                sys.exit(0)
    sys.stderr.write("gh: no pull request %d\n" % number)
    sys.exit(1)
else:
    sys.stderr.write("gh stand-in: unexpected call: %s\n" % " ".join(a))
    sys.exit(64)
