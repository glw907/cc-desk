#!/usr/bin/env bash
# own_test.sh: tests own.sh against a stub claude and fake /proc and log files.
set -uo pipefail

HERE=$(cd "$(dirname "$0")" && pwd)
T=$(mktemp -d "${CLAUDE_JOB_DIR:-/tmp}/own-test.XXXXXX")
trap 'rm -rf "$T"' EXIT
fail=0
check() { # check <label> <command...>
  local label=$1
  shift
  if "$@"; then echo "ok   $label"; else echo "FAIL $label"; fail=1; fi
}

mkdir -p "$T/bin" "$T/state" "$T/proc" "$T/wt/sub"
cat >"$T/bin/claude" <<'E'
#!/usr/bin/env bash
# Stub claude: prints FAKE_LIST for the session list, and records every other call.
# With FAKE_NEW_ID set, a --bg launch adds that row to the list under its --name.
# With FAKE_NOTE_ID set, a --resume prints a note: line and adds that row as the copy.
if [ "$1" = agents ]; then cat "$FAKE_LIST"; exit 0; fi
echo "$*" >>"$FAKE_CALLS"
addrow() {
  python3 -I -c '
import json, sys
rows = json.load(open(sys.argv[1]))
rows.append({"id": sys.argv[2], "sessionId": "s-" + sys.argv[2], "name": sys.argv[3],
             "pid": 0, "cwd": "/x", "state": "working", "startedAt": "t9"})
json.dump(rows, open(sys.argv[1], "w"))' "$FAKE_LIST" "$1" "$2"
}
if [ "$1" = --bg ] && [ "$2" = --name ] && [ -n "${FAKE_NEW_ID:-}" ]; then addrow "$FAKE_NEW_ID" "$3"; fi
if [ "$1" = --resume ] && [ -n "${FAKE_NOTE_ID:-}" ]; then
  addrow "$FAKE_NOTE_ID" copy
  echo "note: started a copy as $FAKE_NOTE_ID"
fi
E
chmod +x "$T/bin/claude"

export CLAUDE_BIN=$T/bin/claude FAKE_LIST=$T/list.json FAKE_CALLS=$T/calls
export STATE_DIR=$T/state PROC_ROOT=$T/proc DAEMON_LOG=$T/daemon.log WORKTREE_ROOT=$T/wt
: >"$FAKE_CALLS"

# fakeproc <pid> <ppid> <arg0> <arg1> [arg2]
fakeproc() {
  mkdir -p "$T/proc/$1"
  printf 'Name:\tclaude\nPPid:\t%s\n' "$2" >"$T/proc/$1/status"
  printf '%s\0%s\0%s\0' "$3" "$4" "${5:-}" >"$T/proc/$1/cmdline"
}
fakeproc 100 1 /home/u/.local/bin/claude daemon run
fakeproc 200 100 claude bg-pty-host
fakeproc 300 200 claude bg-spare
fakeproc 301 200 claude other-thing
fakeproc 302 200 claude bg-spare
fakeproc 303 200 claude bg-spare

cat >"$FAKE_LIST" <<E
[
 {"id":"aaaa1111","sessionId":"s-a","name":"ccd-spike-x","pid":300,"cwd":"$T/wt/sub","state":"working","startedAt":"t1"},
 {"id":"bbbb2222","sessionId":"s-b","name":"ccd-spike-y","pid":301,"cwd":"$T/wt/sub","state":"working","startedAt":"t1"},
 {"id":"cccc3333","sessionId":"s-c","name":"ccd-spike-z","pid":302,"cwd":"$T/wt/sub","state":"working","startedAt":"t1"},
 {"id":"dddd4444","sessionId":"s-d","name":"ccd-spike-w","pid":302,"cwd":"$T/wt/sub","state":"working","startedAt":"t1"},
 {"id":"eeee5555","sessionId":"s-e","name":"ccd-spike-v","pid":303,"cwd":"/elsewhere","state":"working","startedAt":"t1"},
 {"id":"ffff6666","sessionId":"s-f","name":"stranger","pid":999,"cwd":"/elsewhere","state":"idle","startedAt":"t0"},
 {"pid":2811137,"kind":"interactive","name":"someone-else","cwd":"/elsewhere"}
]
E

now=$(date -u +%Y-%m-%dT%H:%M:%SZ)
for r in aaaa1111 bbbb2222 cccc3333 dddd4444 eeee5555; do
  printf '%s\ts-x\tn\t%s\tl\t%s\n' "$r" "$T/wt/sub" "$now" >>"$T/state/owned.tsv"
done

# The audit on an empty log.
rm -f "$T/state/owned.tsv"
out=$(bash "$HERE/own.sh" audit)
check "audit on an empty log prints strangers: 0" grep -qx 'strangers: 0' <<<"$out"
for r in aaaa1111 bbbb2222 cccc3333 dddd4444 eeee5555; do
  printf '%s\ts-x\tn\t%s\tl\t%s\n' "$r" "$T/wt/sub" "$now" >>"$T/state/owned.tsv"
done

# Refusals run no claude command.
bash "$HERE/own.sh" stop ffff6666 2>/dev/null
check "stop of an unowned id is refused" test $? -ne 0
bash "$HERE/own.sh" launch x "$T/wt" --permission-mode bypassPermissions 2>/dev/null
check "launch with bypassPermissions is refused" test $? -ne 0
bash "$HERE/own.sh" launch x "$T/wt" --dangerously-skip-permissions 2>/dev/null
check "launch with --dangerously-skip-permissions is refused" test $? -ne 0
bash "$HERE/own.sh" kill ffff6666 2>/dev/null
check "kill of an unowned id is refused" test $? -ne 0
bash "$HERE/own.sh" kill bbbb2222 2>/dev/null
check "kill of a process that is not bg-spare is refused" test $? -ne 0
bash "$HERE/own.sh" kill cccc3333 2>/dev/null
check "kill where two rows share a pid is refused" test $? -ne 0
bash "$HERE/own.sh" kill eeee5555 2>/dev/null
check "kill outside the worktree is refused" test $? -ne 0
printf '%s\ts\tn\t%s\tl\t%s\n' ssss0001 "$T/wt/sub" "$now" >>"$T/state/owned.tsv"
python3 -I -c "
import json
rows = json.load(open('$FAKE_LIST'))
rows.append({'id': 'ssss0001', 'sessionId': 's', 'name': 'sup', 'pid': 100, 'cwd': '$T/wt/sub', 'state': 'working', 'startedAt': 't'})
json.dump(rows, open('$FAKE_LIST', 'w'))"
bash "$HERE/own.sh" kill ssss0001 2>/dev/null
check "kill of the supervisor is refused" test $? -ne 0
bash "$HERE/own.sh" resume aaaa1111 --dangerously-skip-permissions 2>/dev/null
check "resume with a bypass flag is refused" test $? -ne 0
printf '{"permissions": {"defaultMode": "bypassPermissions"}}\n' >"$T/bad-settings.json"
bash "$HERE/own.sh" launch x "$T/wt" --settings "$T/bad-settings.json" 2>/dev/null
check "launch with a bypass settings file is refused" test $? -ne 0
bash "$HERE/own.sh" launch x "$T/wt" "--settings=$T/bad-settings.json" 2>/dev/null
check "launch with a bypass settings file in --settings= form is refused" test $? -ne 0
check "no claude command ran for any refusal" test ! -s "$FAKE_CALLS"

# A launch adopts only a row that is new, never a pre-existing row of the same name.
bash "$HERE/own.sh" launch stranger "$T/wt" >/dev/null 2>&1
check "launch does not adopt a pre-existing row of the same name" test $? -ne 0
check "the pre-existing row is not in owned.tsv" bash -c "! grep -q ffff6666 '$T/state/owned.tsv'"
: >"$FAKE_CALLS"

# A launch adopts exactly the one new row, with an id-less interactive row in the list.
FAKE_NEW_ID=nnnn7777 bash "$HERE/own.sh" launch ccd-spike-new "$T/wt" >/dev/null 2>&1
check "a launch with an interactive row in the list succeeds" test $? -eq 0
check "the launch adds exactly its new row to owned.tsv" test "$(grep -c '^nnnn7777	s-nnnn7777	ccd-spike-new	' "$T/state/owned.tsv")" = 1

# A resume that starts a copy records the copy as owned.
printf '%s\ts-a\tn\t%s\tl\t%s\n' aaaa1111 "$T/wt/sub" "$now" >"$T/state/owned.tsv"
FAKE_NOTE_ID=rrrr8888 bash "$HERE/own.sh" resume aaaa1111 >/dev/null 2>&1
check "a resume with an interactive row in the list succeeds" test $? -eq 0
check "the resume's copy is in owned.tsv" grep -q '^rrrr8888	' "$T/state/owned.tsv"
: >"$FAKE_CALLS"

# A good stop is logged and runs.
bash "$HERE/own.sh" stop aaaa1111 >/dev/null
check "stop of an owned id runs claude stop" grep -qx 'stop aaaa1111' "$FAKE_CALLS"

# The audit counts.
cat >"$T/state/control.log" <<E
$now	stop	aaaa1111	ran
$now	stop	zzzz9999	ran
$now	stop	ffff6666	refused
E
printf '# saved\nf1\tworking\tT1\nf2\tidle\tT1\nf3\tidle\tT1\n' >"$T/state/others-before.tsv"
printf '# saved\nf1\tidle\tT1\nf2\tidle\tT2\nf3\tidle\tT1\n' >"$T/state/others-after.tsv"
later=$(date -u -d '+3 seconds' +%Y-%m-%dT%H:%M:%S.000Z)
cat >"$DAEMON_LOG" <<E
[2000-01-01T00:00:00.000Z] [bg] bg settled ffff6666 (killed)
[$later] [bg] bg settled ffff6666 (killed)
[$later] [bg] bg settled aaaa1111 (killed)
E
out=$(bash "$HERE/own.sh" audit)
check "audit counts one stranger" grep -qx 'strangers: 1' <<<"$out"
check "audit counts one foreign kill" grep -qx 'foreign kills: 1' <<<"$out"
check "audit counts one foreign restart" grep -qx 'foreign restarts: 1' <<<"$out"
check "audit counts no bypass" grep -qx 'bypass: 0' <<<"$out"
printf '{"x": "bypassPermissions"}\n' >"$T/state/settings.json"
out=$(bash "$HERE/own.sh" audit)
check "audit counts a bypass in a settings file" grep -qx 'bypass: 1' <<<"$out"

# closing.py.
printf '## Proof 1\n\nHeld.\nHeld.\n' >"$T/report.md"
python3 -I "$HERE/closing.py" "$T/report.md" >/dev/null
check "closing.py rejects a section with two closing lines" test $? -ne 0
rm -f "$T/report.md"
for n in 1 2 3 4 5 6 7; do printf '## Proof %s\n\nHeld.\n\n' "$n" >>"$T/report.md"; done
out=$(python3 -I "$HERE/closing.py" "$T/report.md")
check "closing.py accepts seven good sections" test "$out" = "7 sections, one closing line each"

exit $fail
