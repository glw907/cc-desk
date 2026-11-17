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
if [ "$1" = agents ]; then cat "$FAKE_LIST"; exit 0; fi
echo "$*" >>"$FAKE_CALLS"
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
 {"id":"ffff6666","sessionId":"s-f","name":"stranger","pid":999,"cwd":"/elsewhere","state":"idle","startedAt":"t0"}
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
check "no claude command ran for any refusal" test ! -s "$FAKE_CALLS"

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
