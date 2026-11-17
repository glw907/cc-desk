#!/usr/bin/env bash
# own.sh: launches the Prove pass's sessions, records them, acts only on them,
# and audits the pass's effect on every other session.
#
# Usage: own.sh <command> [args]
#   start                          snapshot every other session before the first launch
#   finish                         snapshot every other session at close
#   launch <name> <folder> <flags...>
#   stop|respawn|rm|logs <id>
#   resume <id> [prompt]
#   kill <id>
#   daemonpid                      print the supervisor's process id
#   audit                          print five counts and write spikes/audit.txt
#
# Test overrides: CLAUDE_BIN, PROC_ROOT, DAEMON_LOG, STATE_DIR, WORKTREE_ROOT.
set -euo pipefail

HERE=$(cd "$(dirname "$0")" && pwd)
STATE_DIR=${STATE_DIR:-$HERE}
CLAUDE_BIN=${CLAUDE_BIN:-claude}
PROC_ROOT=${PROC_ROOT:-/proc}
DAEMON_LOG=${DAEMON_LOG:-$HOME/.claude/daemon.log}
WORKTREE_ROOT=${WORKTREE_ROOT:-$(cd "$HERE/.." && pwd)}

OWNED=$STATE_DIR/owned.tsv
CONTROL=$STATE_DIR/control.log
BEFORE=$STATE_DIR/others-before.tsv
AFTER=$STATE_DIR/others-after.tsv
AUDIT=$STATE_DIR/audit.txt

now() { date -u +%Y-%m-%dT%H:%M:%SZ; }
die() { echo "own.sh: $*" >&2; exit 1; }

# save_list <file>: write claude agents --json --all to <file>.
save_list() {
  "$CLAUDE_BIN" agents --json --all >"$1"
}

# rows <list.json> <tsv.out>: id, state, startedAt of every row.
rows() {
  python3 -I -c '
import json, sys
for r in json.load(open(sys.argv[1])):
    print("\t".join(str(r.get(k, "")) for k in ("id", "state", "startedAt")))
' "$1" >"$2"
}

# field <list.json> <id> <key>: one field of the row with that id.
field() {
  python3 -I -c '
import json, sys
for r in json.load(open(sys.argv[1])):
    if r.get("id") == sys.argv[2]:
        print(r.get(sys.argv[3], ""))
        break
' "$1" "$2" "$3"
}

# require_owned <cmd> <id>: log the attempt, and die unless <id> is in owned.tsv.
require_owned() {
  local cmd=$1 id=$2
  if [ -z "$id" ] || ! [ -f "$OWNED" ] || ! cut -f1 "$OWNED" | grep -qxF -- "$id"; then
    printf '%s\t%s\t%s\trefused\n' "$(now)" "$cmd" "$id" >>"$CONTROL"
    die "$cmd refused: $id is not in owned.tsv"
  fi
}

# refuse_bypass <cmd> <args...>: die when an argument turns permission checks off,
# or names a settings file that does.
refuse_bypass() {
  local cmd=$1 arg prev=""
  shift
  for arg in "$@"; do
    case $arg in
      *bypass* | *dangerously*) die "$cmd refused: $arg turns permission checks off" ;;
    esac
    if [ "$prev" = --settings ] && [ -f "$arg" ] && grep -qiE 'bypass|dangerously' "$arg"; then
      die "$cmd refused: settings file $arg turns permission checks off"
    fi
    prev=$arg
  done
}

log_ran() { printf '%s\t%s\t%s\tran\n' "$(now)" "$1" "$2" >>"$CONTROL"; }

# argv_of <pid>: the process's arguments, one per line.
argv_of() {
  tr '\0' '\n' <"$PROC_ROOT/$1/cmdline" 2>/dev/null || true
}

# runs <pid> <word1> <word2>: true when the process is `claude <word1> <word2>`.
runs() {
  local a0 a1 a2
  mapfile -t a < <(argv_of "$1")
  a0=${a[0]:-}
  a1=${a[1]:-}
  a2=${a[2]:-}
  [ "$(basename -- "$a0")" = claude ] || return 1
  [ "$a1" = "$2" ] || return 1
  [ -z "$3" ] || [ "$a2" = "$3" ]
}

ppid_of() {
  awk '/^PPid:/ {print $2}' "$PROC_ROOT/$1/status" 2>/dev/null
}

cmd_start() {
  local tmp
  tmp=$STATE_DIR/others-before.json
  save_list "$tmp"
  { echo "# saved $(now)"; } >"$BEFORE"
  rows "$tmp" "$BEFORE.body"
  cat "$BEFORE.body" >>"$BEFORE"
  rm -f "$BEFORE.body"
  echo "own.sh: saved $(grep -vc '^#' "$BEFORE") rows to others-before.tsv"
}

cmd_finish() {
  local tmp
  tmp=$STATE_DIR/others-after.json
  save_list "$tmp"
  { echo "# saved $(now)"; } >"$AFTER"
  rows "$tmp" "$AFTER.body"
  cat "$AFTER.body" >>"$AFTER"
  rm -f "$AFTER.body"
  echo "own.sh: saved $(grep -vc '^#' "$AFTER") rows to others-after.tsv"
}

cmd_launch() {
  local name=${1:-} folder=${2:-}
  [ -n "$name" ] && [ -n "$folder" ] || die "usage: launch <name> <folder> <flags...>"
  shift 2
  refuse_bypass launch "$@"
  [ -d "$folder" ] || die "launch refused: $folder is not a folder"
  local before=$STATE_DIR/list-prelaunch.json
  save_list "$before"
  local out=$STATE_DIR/launch-$name-$(date -u +%H%M%S).out
  (cd "$folder" && "$CLAUDE_BIN" --bg --name "$name" "$@") >"$out" 2>&1 || {
    cat "$out" >&2
    die "launch failed; output saved to $out"
  }
  local list=$STATE_DIR/list-launch.json
  save_list "$list"
  touch "$OWNED"
  local row
  row=$(python3 -I -c '
import json, sys
existing = {r["id"] for r in json.load(open(sys.argv[2]))}
new = [r for r in json.load(open(sys.argv[1]))
       if r.get("name") == sys.argv[3] and r["id"] not in existing]
if len(new) == 1:
    print("\t".join([new[0]["id"], new[0].get("sessionId", "")]))
' "$list" "$before" "$name")
  [ -n "$row" ] || die "not exactly one new row named $name in the session list; output saved to $out"
  printf '%s\t%s\t%s\t%s\t%s\t%s\n' "${row%%$'\t'*}" "${row#*$'\t'}" "$name" "$folder" \
    "claude --bg --name $name $*" "$(now)" >>"$OWNED"
  cat "$out"
  echo "own.sh: owned ${row%%$'\t'*} ($name)"
}

cmd_simple() { # stop respawn rm logs
  local cmd=$1 id=${2:-}
  require_owned "$cmd" "$id"
  log_ran "$cmd" "$id"
  "$CLAUDE_BIN" "$cmd" "$id"
}

cmd_resume() {
  local id=${1:-}
  shift || true
  refuse_bypass resume "$@"
  require_owned resume "$id"
  local sid
  sid=$(awk -F'\t' -v id="$id" '$1 == id {print $2}' "$OWNED")
  log_ran resume "$id"
  local out=$STATE_DIR/resume-$id-$(date -u +%H%M%S).out
  "$CLAUDE_BIN" --resume "$sid" --bg "$@" 2>&1 | tee "$out"
  # A note: line means a copy under a new id; record it by finding the new row.
  if grep -q 'note:' "$out"; then
    local list=$STATE_DIR/list-resume.json
    save_list "$list"
    python3 -I -c '
import json, sys, re
owned = {l.split("\t")[0] for l in open(sys.argv[2]) if l.strip()}
text = open(sys.argv[3]).read()
note = " ".join(l for l in text.splitlines() if "note:" in l)
for r in json.load(open(sys.argv[1])):
    sid = r.get("sessionId", "")
    if r["id"] not in owned and (r["id"] in note or (sid and sid in note)):
        print("\t".join([r["id"], r.get("sessionId", "")]))
' "$list" "$OWNED" "$out" | while IFS=$'\t' read -r nid nsid; do
      printf '%s\t%s\t%s\t%s\t%s\t%s\n' "$nid" "$nsid" "copy-of-$id" "?" "resume $sid" "$(now)" >>"$OWNED"
      echo "own.sh: owned the copy $nid"
    done
  fi
}

cmd_kill() {
  local id=${1:-}
  require_owned kill "$id"
  local list=$STATE_DIR/list-kill.json
  save_list "$list"
  local pid cwd
  pid=$(field "$list" "$id" pid)
  cwd=$(field "$list" "$id" cwd)
  refuse() {
    printf '%s\t%s\t%s\trefused\n' "$(now)" kill "$id" >>"$CONTROL"
    die "kill refused: $1"
  }
  [ -n "$pid" ] || refuse "no row or no pid for $id"
  runs "$pid" bg-spare "" || refuse "pid $pid does not run claude bg-spare"
  local parent gp
  parent=$(ppid_of "$pid")
  [ -n "$parent" ] || refuse "no parent for pid $pid"
  gp=$(ppid_of "$parent")
  [ -n "$gp" ] || refuse "no grandparent for pid $pid"
  runs "$gp" daemon run || refuse "grandparent $gp does not run claude daemon run"
  [ "$gp" != "$pid" ] || refuse "the target is the supervisor"
  local same
  same=$(python3 -I -c '
import json, sys
print(sum(1 for r in json.load(open(sys.argv[1])) if str(r.get("pid", "")) == sys.argv[2]))
' "$list" "$pid")
  [ "$same" = 1 ] || refuse "$same rows share pid $pid"
  case $cwd in
    "$WORKTREE_ROOT"/*) ;;
    *) refuse "folder $cwd is not under $WORKTREE_ROOT" ;;
  esac
  log_ran kill "$id"
  kill "$pid"
}

cmd_daemonpid() {
  [ -f "$OWNED" ] || die "no owned sessions"
  local list=$STATE_DIR/list-daemonpid.json id pid parent gp
  save_list "$list"
  while IFS=$'\t' read -r id _; do
    pid=$(field "$list" "$id" pid)
    [ -n "$pid" ] || continue
    runs "$pid" bg-spare "" || continue
    parent=$(ppid_of "$pid")
    gp=$(ppid_of "$parent")
    if [ -n "$gp" ] && runs "$gp" daemon run; then
      echo "$gp"
      return 0
    fi
  done <"$OWNED"
  die "no live owned session leads to a claude daemon run process"
}

cmd_audit() {
  python3 -I - "$STATE_DIR" "$DAEMON_LOG" "$STATE_DIR" <<'PY' | tee "$AUDIT.summary"
import glob, json, os, re, sys
from datetime import datetime, timezone

state, daemon_log, spikes = sys.argv[1:4]

def read_tsv(path):
    if not os.path.exists(path):
        return []
    return [l.rstrip("\n").split("\t") for l in open(path) if l.strip() and not l.startswith("#")]

def ts(s):
    return datetime.strptime(s, "%Y-%m-%dT%H:%M:%SZ").replace(tzinfo=timezone.utc).timestamp()

owned = read_tsv(os.path.join(state, "owned.tsv"))
owned_ids = {r[0] for r in owned}
control = read_tsv(os.path.join(state, "control.log"))
lines = []

# strangers: ran control targets that aren't owned.
strangers = [r for r in control if len(r) >= 4 and r[3] == "ran" and r[2] not in owned_ids]
refused = [r for r in control if len(r) >= 4 and r[3] == "refused"]
lines.append("== strangers (control.log targets not in owned.tsv)")
lines += ["\t".join(r) for r in strangers]
lines.append("== refused calls (not counted)")
lines += ["\t".join(r) for r in refused]

# foreign kills.
first = min((ts(r[5]) for r in owned if len(r) >= 6), default=None)
ctl_times = [ts(r[0]) for r in control if len(r) >= 4 and r[3] == "ran"]
pat = re.compile(r"^\[(\S+)\] \[bg\] bg settled (\S+) \(killed\)")
foreign, counted = [], 0
if os.path.exists(daemon_log) and first is not None:
    for line in open(daemon_log):
        m = pat.match(line)
        if not m:
            continue
        t = datetime.fromisoformat(m.group(1).replace("Z", "+00:00")).timestamp()
        if t < first:
            continue
        sid = m.group(2)
        if any(i == sid or i.startswith(sid) or sid.startswith(i) for i in owned_ids):
            continue
        foreign.append(line.strip())
        if any(0 <= t - c <= 10 for c in ctl_times):
            counted += 1
lines.append("== foreign settled-killed lines after the first launch (all listed)")
lines += foreign
foreign_kills = counted

# foreign restarts.
before = {r[0]: r for r in read_tsv(os.path.join(state, "others-before.tsv"))}
after = {r[0]: r for r in read_tsv(os.path.join(state, "others-after.tsv"))}
restarts, changed = [], []
for i, b in before.items():
    a = after.get(i)
    if a is None:
        continue
    if len(b) > 2 and len(a) > 2 and b[2] != a[2]:
        restarts.append(f"{i}\tstartedAt {b[2]} -> {a[2]}")
    elif len(b) > 1 and len(a) > 1 and b[1] != a[1]:
        changed.append(f"{i}\tstate {b[1]} -> {a[1]}")
lines.append("== foreign restarts (startedAt changed)")
lines += restarts
lines.append("== state changes in other sessions (not counted)")
lines += changed

# bypass.
patterns = re.compile(r"bypassPermissions|dangerously|\"defaultMode\"\s*:\s*\"bypass")
settings = sorted(glob.glob(os.path.join(spikes, "**", "*settings*.json"), recursive=True))
hook_files, scan = set(), list(settings)

def commands(o):
    if isinstance(o, dict):
        for k, v in o.items():
            if k == "command" and isinstance(v, str):
                yield v
            else:
                yield from commands(v)
    elif isinstance(o, list):
        for v in o:
            yield from commands(v)

for s in settings:
    try:
        data = json.load(open(s))
    except ValueError:
        continue
    for c in commands(data):
        for tok in c.split():
            if tok.endswith((".py", ".sh")):
                p = tok if os.path.isabs(tok) else os.path.join(os.path.dirname(s), tok)
                if not os.path.exists(p):
                    p = os.path.join(spikes, "..", tok)
                if os.path.exists(p):
                    hook_files.add(os.path.normpath(p))
for p in glob.glob(os.path.join(spikes, "**", "*hook*"), recursive=True):
    if os.path.isfile(p):
        hook_files.add(os.path.normpath(p))
scan += sorted(hook_files)
scan.append(os.path.join(state, "owned.tsv"))
bypass = []
for f in scan:
    if os.path.isfile(f):
        for n, l in enumerate(open(f, errors="replace"), 1):
            if patterns.search(l):
                bypass.append(f"{f}:{n}:{l.strip()}")
lines.append("== bypass matches")
lines += bypass

allow_files = []
for f in sorted(hook_files):
    if re.search(r"\ballow\b", open(f, errors="replace").read()):
        allow_files.append(os.path.relpath(f, os.path.join(spikes, "..")))

with open(os.path.join(state, "audit.txt"), "w") as out:
    out.write("\n".join(lines) + "\n")
print(f"strangers: {len(strangers)}")
print(f"foreign kills: {foreign_kills}")
print(f"foreign restarts: {len(restarts)}")
print(f"bypass: {len(bypass)}")
print("allow hooks: " + (" ".join(allow_files) if allow_files else "none"))
PY
  rm -f "$AUDIT.summary"
}

cmd=${1:-}
[ $# -gt 0 ] && shift
case $cmd in
  start) cmd_start ;;
  finish) cmd_finish ;;
  launch) cmd_launch "$@" ;;
  stop | respawn | rm | logs) cmd_simple "$cmd" "${1:-}" ;;
  resume) cmd_resume "$@" ;;
  kill) cmd_kill "${1:-}" ;;
  daemonpid) cmd_daemonpid ;;
  audit) cmd_audit ;;
  *) die "usage: own.sh start|finish|launch|stop|respawn|rm|logs|resume|kill|daemonpid|audit" ;;
esac
