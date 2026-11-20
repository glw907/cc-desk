#!/usr/bin/env bash
# live-skills.sh: the live run of the desk skill, one phase per call.
#
#   bash scripts/live-skills.sh run
#   bash scripts/live-skills.sh reader <file>
#
# run builds the binary into the job folder, prints each skills fixture's brief
# from a copy, and diffs it against scripts/live/skills/. It then starts one
# claude -p desk session per fixture from the scratch checkout and grades each
# opening. reader grades the no-context reader's reply, saved to <file>. Every
# check prints "pass <case>: <check>" or "fail <case>: <check>: <found>", and
# each phase ends with "result: pass" or "result: fail". The log is
# $CLAUDE_JOB_DIR/tmp/live-skills/run.log. The plan is
# docs/plans/2026-11-18-desk-skills-plan.md, "The live script".
set -euo pipefail

REPO=$(cd "$(dirname "$0")/.." && pwd)
SCRATCH=${SCRATCH_CHECKOUT:-$HOME/Projects/cc-desk-scratch}
JOB=${CLAUDE_JOB_DIR:-${TMPDIR:-/tmp}/cc-desk-live-skills}/tmp
RUN=$JOB/live-skills
BIN=$JOB/cc-desk-live-skills
NOW=2026-11-20T23:00:00Z
SKILL=$REPO/skills/cc-desk/SKILL.md
CASES=(none one two)

mkdir -p "$RUN"
exec > >(tee -a "$RUN/run.log") 2>&1

show() {
  printf '$ %s\n' "$*" >&2
  "$@"
}

# grade runs the checks table of the plan's "The live script" over the saved
# openings, or over the reader's reply.
grade() {
  python3 -I - "$@" <<'PY'
import json, re, sys

results = []

def check(case, name, ok, found=""):
    results.append(ok)
    print(f"pass {case}: {name}" if ok else f"fail {case}: {name}: {found}")

IDS = {
    "none": ["q-tide-1"],
    "one": ["q-juneau-7", "a7c2e9d1"],
    "two": ["merge-dock-2", "q-ferry-3", "9e41b07"],
}
OFFERS = ["Would you like", "Shall I", "Do you want", "Should I"]

def lines_of(text):
    return [l.strip().strip("*_").strip() for l in text.splitlines()]

def first_tool(path):
    for raw in open(path):
        try:
            ev = json.loads(raw)
        except ValueError:
            continue
        if ev.get("type") != "assistant":
            continue
        for part in ev.get("message", {}).get("content", []):
            if part.get("type") == "tool_use":
                return part
    return None

def every(case, text, jsonl, skill):
    tool = first_tool(jsonl)
    got = f"{tool.get('name')} {json.dumps(tool.get('input'))}" if tool else "no tool call"
    check(case, "the first tool call reads the skill",
          bool(tool) and tool.get("name") == "Read" and tool.get("input", {}).get("file_path") == skill, got)
    body = "\n".join(l for l in lines_of(text) if not l.endswith("?"))
    offers = [o for o in OFFERS if o.lower() in body.lower()]
    check(case, "opens without being asked", bool(text.strip()) and not offers,
          "empty opening" if not text.strip() else ", ".join(offers))
    bad = [i for i in IDS[case] + ["r-0042"] if re.search(r"\b" + re.escape(i) + r"\b", text)]
    bad += re.findall(r"\br-[0-9]{4}\b", text)
    check(case, "holds no ruling or question id", not bad, ", ".join(sorted(set(bad))))

def has(text, phrase):
    return phrase.lower() in text.lower()

def none(text):
    t = text.strip()
    ends = re.findall(r"[.?!](?=\s|$)", t)
    check("none", "the opening is one sentence", len(ends) == 1, f"{len(ends)} sentence ends")
    ok = re.search(r"\b10:00 ?pm\b", t, re.I) and not has(t, "tomorrow")
    check("none", "holds the forecast time, 10:00 pm, and no tomorrow", bool(ok), t)

def one(text):
    want = ["30 GB", "15 minutes longer", "40 minutes a night", "the Sitka library"]
    miss = [w for w in want if not has(text, w)]
    if not re.search(r"\b85\b", text):
        miss.append("85")
    check("one", "holds a marked phrase from each part", not miss, "missing " + ", ".join(miss))
    ls = lines_of(text)
    ok = False
    for i, l in enumerate(ls):
        if l.endswith("?") and has(l, "the Juneau archive"):
            above = i == 0 or ls[i - 1] == ""
            below = i == len(ls) - 1 or ls[i + 1] == ""
            ok = ok or (above and below)
    check("one", "puts the plain question on a separate line", ok, "no such line")
    bare = [re.sub(r"[*_]", "", l).strip().lower() for l in ls]
    labels = [a for a in ("yes", "no") if any(re.fullmatch(a + r"( \(recommended\))?", b) for b in bare)]
    check("one", "puts each answer's label on its own line", labels == ["yes", "no"],
          "own-line labels: " + (", ".join(labels) or "none"))

def two(text):
    low = text.lower()
    m = low.find("one searchable file")
    plain = [low.find(p) for p in ["the family calendar", "about 60 photos", "sunday morning",
                                   "your mail only", "crowd the shared view"] if p in low]
    check("two", "the question that holds more planned minutes comes first",
          m >= 0 and all(m < p for p in plain), f"merge at {m}, plain at {plain}")
    miss = [w for w in ["merge", "fix again", "close"] if w not in low]
    num = re.search(r"\b4817\b", text)
    ok = m >= 0 and not num and not miss
    found = []
    if m < 0:
        found.append("no title phrase")
    if num:
        found.append("holds 4817")
    if miss:
        found.append("missing " + ", ".join(miss))
    check("two", "names its pull request by what it does and gives the three answers", ok, "; ".join(found))

mode = sys.argv[1]
if mode == "run":
    run, skill = sys.argv[2], sys.argv[3]
    for case, fn in (("none", none), ("one", one), ("two", two)):
        text = open(f"{run}/{case}.txt").read()
        every(case, text, f"{run}/{case}.jsonl", skill)
        fn(text)
else:
    text = open(sys.argv[2]).read()
    miss = [w for w in ["the Juneau archive", "15 minutes longer", "40 minutes a night"] if not has(text, w)]
    check("reader", "states the question and what each answer causes", not miss, "missing " + ", ".join(miss))

print("result: pass" if all(results) else "result: fail")
sys.exit(0 if all(results) else 1)
PY
}

run_phase() {
  show go build -C "$REPO" -o "$BIN" ./cmd/cc-desk
  local c state brief system
  for c in "${CASES[@]}"; do
    state=$RUN/$c
    rm -rf "$state"
    cp -r "$REPO/e2e/testdata/skills-$c" "$state"
    brief=$RUN/brief-$c.txt
    show "$BIN" --state-dir "$state" --now "$NOW" brief >"$brief"
    show diff -u "$REPO/scripts/live/skills/brief-$c.txt" "$brief"
  done
  for c in "${CASES[@]}"; do
    system=$(sed -e "s|@SKILL_PATH@|$SKILL|" "$REPO/scripts/live/skills/system.txt" |
      awk -v f="$RUN/brief-$c.txt" '$0 == "@BRIEF@" { while ((getline l < f) > 0) print l; next } { print }')
    printf '$ (cd %s && timeout 120 claude -p ... "What is waiting for me?") > %s\n' "$SCRATCH" "$RUN/$c.jsonl"
    (cd "$SCRATCH" && timeout 120 claude -p --model claude-opus-5-5 --effort high --tools Read \
      --add-dir "$REPO/skills/cc-desk" --output-format stream-json --verbose \
      --append-system-prompt "$system" "What is waiting for me?") >"$RUN/$c.jsonl" || true
    jq -r 'select(.type == "result") | .result // empty' "$RUN/$c.jsonl" >"$RUN/$c.txt"
    printf -- '--- opening (%s) ---\n' "$c"
    cat "$RUN/$c.txt"
    printf -- '--- end ---\n'
  done
  grade run "$RUN" "$SKILL"
}

case "${1:-}" in
  run) run_phase ;;
  reader)
    [ -n "${2:-}" ] || { echo "usage: live-skills.sh reader <file>" >&2; exit 2; }
    printf -- '--- reader reply ---\n'
    cat "$2"
    printf -- '--- end ---\n'
    grade reader "$2"
    ;;
  *) echo "usage: live-skills.sh <run|reader <file>>" >&2; exit 2 ;;
esac
