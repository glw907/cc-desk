#!/usr/bin/env bash
# live-desk.sh: the live run of the desk-open pass, one phase per call.
#
#   bash scripts/live-desk.sh <prep|open|round|rule|start|refuse|replace|merge|close|abort|evidence|help>
#
# Each phase prints the commands it ran with their output and appends all of it to
# $CLAUDE_JOB_DIR/tmp/live-desk/run.log. Each live clause prints `pass <check>` or
# `fail <check>`, and the evidence phase prints `result: pass` or `result: fail`.
# It reads CC_DESK_STATE_DIR (default ~/.local/state/cc-desk) and SCRATCH_CHECKOUT
# (default ~/Projects/cc-desk-scratch). The plan is docs/plans/2026-11-18-desk-open-plan.md,
# "Task 6" and "Clause table".
set -euo pipefail

REPO=$(cd "$(dirname "$0")/.." && pwd)
STATE=${CC_DESK_STATE_DIR:-$HOME/.local/state/cc-desk}
SCRATCH=${SCRATCH_CHECKOUT:-$HOME/Projects/cc-desk-scratch}
RUN=${CLAUDE_JOB_DIR:-${TMPDIR:-/tmp}/cc-desk-live}/tmp/live-desk
BIN=$RUN/cc-desk
UNIT_DIR=$HOME/.config/systemd/user
PROJECT=cc-desk-scratch
# BASE carries the prep time, so a second run gets its own branch; prep writes it
# and every later phase reads it back.
BASE=$(cat "$RUN/base" 2>/dev/null || true)
RELEASE=$STATE/releases/$PROJECT.json
RESULTS=$RUN/results.tsv
SKILL=$REPO/skills/cc-desk/SKILL.md
# The words the rule phase types: the four characters, and a second line.
WORDS_1="keep the scale, it's \"fine\" at \$5"
WORDS_2="second line"
EXPECTED=$WORDS_1$'\n'$WORDS_2

# Every clause a full run records; the result line fails when one is missing.
EXPECTED_CLAUSES="prep: claude is on the unit's PATH
prep: gh is on the unit's PATH
prep: git is on the unit's PATH
prep: systemd-run is on the unit's PATH
3 desk registered
3 listed interactive under the registered id
3 arguments name Opus and high effort
3 the screen holds the first line
3 the screen holds no brief line and no instruction
3 no permission prompt, no trust question
3 the round form's screen holds that round's question
4 the typed answer adds one line, byte for byte
4 the snapshot drops the question within 5 seconds
4 the pass starts within 5 minutes and the log holds the start
4 the pass session has a row
4 the desk ran no start command
4 refused from a background session
4 refused from a shell with no terminal, naming it
5 the session was ended and a new one started
5 the new session's first prompt holds the words
5 the count is unchanged
5 never two sessions at once
5 merge asks again when the head moved
5 merge merges at the commit asked about
5 close closes the pull request
5 close records the pass closed
5 close empties Needs you now for it"

say() { printf '%s\n' "$*"; }
show() {
  printf '$ %s\n' "$*"
  "$@"
}
try() {
  printf '$ %s\n' "$*"
  "$@" || printf '(exit %s)\n' "$?"
}

cc() { "$BIN" --state-dir "$STATE" "$@"; }

# check prints the verdict of one live clause and records it for the result line.
check() {
  local verdict=$1 name=$2
  say "$verdict $name"
  printf '%s\t%s\n' "$verdict" "$name" >>"$RESULTS"
}
# expect runs a test command and prints the clause's verdict.
expect() {
  local name=$1
  shift
  if "$@"; then check pass "$name"; else check fail "$name"; fi
}

unit_path() {
  grep '^Environment=PATH=' "$UNIT_DIR/cc-desk-check.service" | sed -e 's/^Environment=PATH=//' -e "s|%h|$HOME|g"
}

pass_field() {
  jq -r --arg p "$1" ".passes[] | select(.name == \$p) | $2" "$RELEASE"
}

# wait_for polls a command every 5 seconds until it succeeds or the seconds run out.
wait_for() {
  local secs=$1 i
  shift
  for ((i = 0; i * 5 < secs; i++)); do
    "$@" && return 0
    sleep 5
  done
  return 1
}

# wait_saving is wait_for that saves the session list before each try.
wait_saving() {
  local secs=$1 i
  shift
  for ((i = 0; i * 5 < secs; i++)); do
    save_list
    "$@" && return 0
    sleep 5
  done
  return 1
}

desk_session() { cat "$RUN/desk-session" 2>/dev/null; }
pane() { tmux capture-pane -p -t "$1" -S -200; }
pane_has() { pane "$1" | grep -q -- "$2"; }

# The polls below run in bash -c, so the helpers they call are exported.
export RELEASE
export -f pass_field

# type_in sends text literally to a pane, then Enter.
type_in() {
  tmux send-keys -t "$1" -l -- "$2"
  sleep 1
  tmux send-keys -t "$1" Enter
}

# save_list writes the current session list to a numbered file.
save_list() {
  mkdir -p "$RUN/lists"
  claude agents --json --all >"$RUN/lists/$(date -u +%H%M%S).json" 2>/dev/null || true
}

# desk_row prints the desk's registered session id when the saved list holds its row.
desk_count() { jq -r '.desks | length' "$STATE/desks.json" 2>/dev/null || echo 0; }
desk_id() { jq -r '.desks[-1].session_id // empty' "$STATE/desks.json" 2>/dev/null; }
row_for() {
  claude agents --json --all 2>/dev/null |
    jq -e --arg id "$1" '[.. | objects | select((.sessionId? // .session_id? // .id?) == $id)] | length > 0' >/dev/null
}

# ---- prep ------------------------------------------------------------------

prep_refuse() {
  if [ ! -e "$SCRATCH/.git" ]; then
    say "refused: $SCRATCH is not a git checkout"
    exit 1
  fi
  local dirty branch other
  dirty=$(git -C "$SCRATCH" status --porcelain)
  if [ -n "$dirty" ]; then
    say "refused: the scratch checkout $SCRATCH is dirty:"
    say "$dirty"
    exit 1
  fi
  branch=$(git -C "$SCRATCH" branch --show-current)
  if [ "$branch" != main ]; then
    say "refused: the scratch checkout is on '$branch', not main"
    exit 1
  fi
  other=$(find "$STATE/releases" -maxdepth 1 -name '*.json' ! -name "$PROJECT.json" 2>/dev/null || true)
  if [ -n "$other" ]; then
    say "refused: the state folder $STATE holds a release other than $PROJECT's:"
    say "$other"
    exit 1
  fi
}

phase_prep() {
  prep_refuse
  mkdir -p "$RUN"
  rm -rf "$RUN/lists"
  # Sessions from an earlier run keep their rows in the list, so the clauses
  # that count sessions count only those started after this moment.
  date +%s%3N >"$RUN/epoch"
  : >"$RESULTS"
  BASE=live-desk-$(date -u +%Y%m%d-%H%M%S)
  echo "$BASE" >"$RUN/base"
  say "prep: the scratch checkout is clean on main, and the state folder holds no other release; base branch $BASE"

  show go build -o "$BIN" "$REPO/cmd/cc-desk"
  show "$BIN" timer install
  show systemd-analyze --user verify "$UNIT_DIR/cc-desk-check.service" "$UNIT_DIR/cc-desk-check.timer"

  local path tool
  path=$(unit_path)
  for tool in claude gh git systemd-run; do
    if PATH=$path command -v "$tool" >/dev/null; then
      check pass "prep: $tool is on the unit's PATH"
    else
      check fail "prep: $tool is on the unit's PATH"
    fi
  done

  # A rerun starts clean: an earlier run's pass worktrees and branches would
  # make a launch refuse, as it refuses a worktree no record names.
  local leftover
  for leftover in desk-live-a desk-merge desk-close; do
    if [ -d "$SCRATCH/.claude/worktrees/$leftover" ]; then
      show git -C "$SCRATCH" worktree remove --force ".claude/worktrees/$leftover"
    fi
    git -C "$SCRATCH" branch -D "$leftover" >/dev/null 2>&1 || true
  done
  show git -C "$SCRATCH" fetch origin
  show git -C "$SCRATCH" switch -c "$BASE" origin/main
  mkdir -p "$SCRATCH/docs/plans"
  # desk-live-a must still be building when replace runs, so its plan only waits.
  show cp "$REPO/scripts/live/desk/hold-plan.md" "$SCRATCH/docs/plans/desk-live-a-plan.md"
  show cp "$REPO/scripts/live/desk/hold-plan.md" "$SCRATCH/docs/plans/desk-merge-plan.md"
  show cp "$REPO/scripts/live/desk/hold-plan.md" "$SCRATCH/docs/plans/desk-close-plan.md"
  show git -C "$SCRATCH" add docs/plans
  show git -C "$SCRATCH" commit -m "Add the desk-live fixture plans"
  show git -C "$SCRATCH" push origin "$BASE"

  local exclude
  exclude=$(git -C "$SCRATCH" rev-parse --git-path info/exclude)
  case $exclude in /*) ;; *) exclude=$SCRATCH/$exclude ;; esac
  grep -qx '.claude/worktrees/' "$exclude" 2>/dev/null || echo '.claude/worktrees/' >>"$exclude"

  local rel=$RUN/release.json
  jq --arg n "$PROJECT-$(date -u +%Y%m%dT%H%M%SZ)" '.name = $n' "$REPO/scripts/live/desk/release.json" >"$rel"
  show "$BIN" --state-dir "$STATE" release add "$rel"
  show "$BIN" --state-dir "$STATE" release checkout "$PROJECT" "$SCRATCH" --default-branch "$BASE"
  say "usage reading:"
  show "$BIN" --state-dir "$STATE" usage --raw "$RUN/usage.raw"
  say "the timer now runs this branch's build; the conductor reinstalls from main once the pass lands"
}

# ---- open ------------------------------------------------------------------

phase_open() {
  local s
  s=$PROJECT-desk-$(date -u +%H%M%S)
  echo "$s" >"$RUN/desk-session"
  DESKS_BEFORE=$(desk_count)
  show tmux new-session -d -s "$s" -x 200 -y 50 -c "$SCRATCH" "$BIN --state-dir $STATE --skill-file $SKILL"
  local id
  if wait_for 60 desk_id_new; then check pass "3 desk registered"; else check fail "3 desk registered"; fi
  id=$(desk_id)
  echo "$id" >"$RUN/desk-id"
  say "desk session id: $id"
  wait_for 120 row_for "$id" || true
  save_list
  # Claude Code opens by itself; wait for the opening to paint.
  wait_for 90 pane_has "$s" 'waiting for you' || true
  pane "$s" | tee "$RUN/open.pane"

  expect "3 listed interactive under the registered id" \
    bash -c "claude agents --json --all | jq -e --arg id '$id' '[.. | objects | select((.sessionId? // .session_id? // .id?) == \$id) | tostring | contains(\"interactive\")] | any' >/dev/null"
  local pid
  pid=$(jq -r '.desks[-1].pid' "$STATE/desks.json")
  expect "3 arguments name Opus and high effort" \
    bash -c "tr '\\0' ' ' </proc/$pid/cmdline | grep -q 'opus' && tr '\\0' ' ' </proc/$pid/cmdline | grep -q -- '--effort high'"
  expect "3 the screen holds the first line" grep -q 'What is waiting for me' "$RUN/open.pane"
  expect "3 the screen holds no brief line and no instruction" \
    bash -c "! grep -q 'Brief at' '$RUN/open.pane' && ! grep -q 'Type each cc-desk verb as this command' '$RUN/open.pane'"
  expect "3 no permission prompt, no trust question" \
    bash -c "! grep -qi 'do you want to\\|trust the files\\|trust this folder' '$RUN/open.pane'"
}

# desk_id_new is true once desks.json holds more desks than before the open.
desk_id_new() { [ "$(desk_count)" -gt "$DESKS_BEFORE" ]; }

# ---- round -----------------------------------------------------------------

phase_round() {
  local s
  s=$PROJECT-desk-round-$(date -u +%H%M%S)
  show tmux new-session -d -s "$s" -x 200 -y 50 -c "$SCRATCH" "$BIN --state-dir $STATE --skill-file $SKILL $PROJECT desk-live"
  wait_for 120 pane_has "$s" 'harbor chart' || true
  pane "$s" | tee "$RUN/round.pane"
  expect "3 the round form's screen holds that round's question" grep -q 'harbor chart' "$RUN/round.pane"
  try tmux kill-session -t "$s"
}

# ---- rule ------------------------------------------------------------------

phase_rule() {
  local s lines before after t0
  s=$(desk_session)
  before=$(wc -l <"$STATE/rulings.jsonl" 2>/dev/null || echo 0)
  date -u +%Y-%m-%dT%H:%M:%SZ >"$RUN/answer-time"
  tmux send-keys -t "$s" -l -- "$WORDS_1"
  # Alt+Enter is the newline key; the run records whether it worked.
  tmux send-keys -t "$s" M-Enter
  tmux send-keys -t "$s" -l -- "$WORDS_2"
  sleep 1
  tmux send-keys -t "$s" Enter
  t0=$(date +%s)
  wait_for 180 bash -c "[ \"\$(wc -l <'$STATE/rulings.jsonl' 2>/dev/null || echo 0)\" -gt $before ]" || true
  lines=$(wc -l <"$STATE/rulings.jsonl" 2>/dev/null || echo 0)
  after=$lines
  say "rulings.jsonl: $before line(s) before, $after after; newline key: M-Enter"
  expect "4 the typed answer adds one line, byte for byte" \
    bash -c "[ $after -eq $((before + 1)) ] && [ \"\$(tail -n1 '$STATE/rulings.jsonl' | jq -r .words)\" = \"\$1\" ]" _ "$EXPECTED"
  expect "4 the snapshot drops the question within 5 seconds" \
    bash -c "for i in 1 2 3 4 5; do cc() { '$BIN' --state-dir '$STATE' \"\$@\"; }; if ! cc snapshot | jq -e '.waits_for_geoff | map(select(.project == \"$PROJECT\")) | length > 0' >/dev/null; then exit 0; fi; sleep 1; done; exit 1"
  say "rule: $(($(date +%s) - t0)) seconds from typing to the last check"
  pane "$s" | tee "$RUN/rule.pane" >/dev/null
}

# ---- start -----------------------------------------------------------------

# started is true once the check logged a start of desk-live-a after the answer
# and the pass has a session.
started() {
  local since
  since=$(cat "$RUN/answer-time" 2>/dev/null || cat "$RUN/start-time")
  jq -e -s --arg p "$PROJECT" --arg since "$since" '
    def secs: sub("\\.[0-9]+"; "") | fromdateiso8601;
    any(.[]; .project == $p and .record == "pass:desk-live-a" and .decision == "start" and (.time | secs) >= ($since | secs))' \
    "$STATE/check-log.jsonl" >/dev/null 2>&1 &&
    [ -n "$(pass_field desk-live-a '.session.session_id // empty')" ]
}

phase_start() {
  date -u +%Y-%m-%dT%H:%M:%SZ >"$RUN/start-time"
  wait_saving 300 started || true
  if started; then check pass "4 the pass starts within 5 minutes and the log holds the start"; else check fail "4 the pass starts within 5 minutes and the log holds the start"; fi
  local sid
  sid=$(pass_field desk-live-a '.session.session_id // empty')
  echo "$sid" >"$RUN/pass-session-1"
  expect "4 the pass session has a row" row_for "$sid"
  save_list
  # The desk's transcript, read now that the start window has passed.
  local t tr=""
  for t in "$HOME"/.claude/projects/*/"$(cat "$RUN/desk-id")".jsonl; do
    [ -f "$t" ] && tr=$t && break
  done
  if [ -n "$tr" ] && jq -r '.. | objects | select(.type? == "tool_use" and .name? == "Bash") | .input.command // empty' "$tr" >"$RUN/desk-bash.txt"; then
    say "desk Bash calls recorded in $RUN/desk-bash.txt from $tr"
    expect "4 the desk ran no start command" bash -c "! grep -q 'pass start' '$RUN/desk-bash.txt'"
  else
    say "no desk transcript found"
    check fail "4 the desk ran no start command"
  fi
}

# ---- refuse ----------------------------------------------------------------

# proc_started prints a process's real start as a UTC timestamp, the way the
# desk's guard reads it: field 22 of /proc/<pid>/stat over the clock tick, on
# the boot time from /proc/stat.
proc_started() {
  local ticks btime hz
  ticks=$(sed -e 's/^.*) //' "/proc/$1/stat" | cut -d' ' -f20)
  btime=$(awk '$1 == "btime" {print $2}' /proc/stat)
  hz=$(getconf CLK_TCK)
  date -u -d "@$((btime + ticks / hz))" +%Y-%m-%dT%H:%M:%SZ
}

# claude_pid prints the pid of the process named claude whose command line
# holds the session id, not the setsid or shell parent.
claude_pid() {
  local p
  for p in $(pgrep -f -- "--session-id $1" || true); do
    if [ "$(cat "/proc/$p/comm" 2>/dev/null)" = claude ]; then
      echo "$p"
      return 0
    fi
  done
  return 1
}

phase_refuse() {
  local copy=$RUN/state-copy
  rm -rf "$copy"
  show cp -r "$STATE" "$copy"
  # The executor's own session: a background session, so no terminal and no desk.
  local own=${CLAUDE_SESSION_ID:-executor}
  local out
  out=$("$BIN" --state-dir "$copy" ruling record --project "$PROJECT" --round desk-live --question dq1 --option keep "from the executor" 2>&1) || true
  say "$out"
  if printf '%s' "$out" | grep -q "isn.t a registered desk"; then check pass "4 refused from a background session"; else check fail "4 refused from a background session"; fi
  say "(executor session $own)"
  save_list

  # Real claude under setsid, registered in the copy before it can act, records against the copy.
  local id s pid=""
  id=$(cat /proc/sys/kernel/random/uuid)
  s=$PROJECT-nosetty-$(date -u +%H%M%S)
  local settings=$RUN/nodesk-settings.json
  jq -n --arg r "Bash($BIN --state-dir $copy ruling record *)" '{permissions: {allow: [$r]}}' >"$settings"
  tmux new-session -d -s "$s" -x 200 -y 50 -c "$SCRATCH" \
    "setsid -w claude --session-id $id --model sonnet --effort low --permission-mode default --settings $settings --name $PROJECT-nodesk"
  wait_saving 60 claude_pid "$id" || true
  pid=$(claude_pid "$id" || true)
  say "setsid claude pid: ${pid:-none}"
  if [ -n "$pid" ]; then
    jq --arg id "$id" --argjson pid "$pid" --arg t "$(proc_started "$pid")" \
      '.desks += [{session_id: $id, pid: $pid, started: $t}]' "$copy/desks.json" >"$copy/desks.json.new"
    mv "$copy/desks.json.new" "$copy/desks.json"
    # Registered; now let the session act. Wait for its input box, then type the prompt.
    wait_saving 60 pane_has "$s" 'for shortcuts' || true
    sleep 3
    type_in "$s" "Run this command and print what it says: $BIN --state-dir $copy ruling record --project $PROJECT --round desk-live --question dq1 --option keep words"
  fi
  wait_saving 120 bash -c "tmux capture-pane -p -t '$s' -S -200 | tr -s '[:space:]' ' ' | grep -q \"desk.s process has no terminal\"" || true
  pane "$s" | tee "$RUN/refuse.pane"
  expect "4 refused from a shell with no terminal, naming it" \
    bash -c "tr -s '[:space:]' ' ' <'$RUN/refuse.pane' | grep -q \"desk.s process has no terminal\""
  try tmux kill-session -t "$s"
  if [ -n "$pid" ]; then try kill "$pid"; fi
  save_list
}

# ---- replace ---------------------------------------------------------------

# relaunch_counts prints the pass's relaunch entries as "counted uncounted".
relaunch_counts() {
  pass_field desk-live-a '[.session.relaunches // [] | .[] | .counted] | "\(map(select(. == true)) | length) \(map(select(. == false)) | length)"'
}

phase_replace() {
  local s old
  s=$(desk_session)
  old=$(cat "$RUN/pass-session-1" 2>/dev/null || true)
  local c0 u0
  read -r c0 u0 <<<"$(relaunch_counts)"
  save_list
  type_in "$s" "I recorded the wrong thing. Replace my last ruling: I want to change the scale. Record it as my ruling with the replaces flag."
  wait_saving 300 bash -c "[ \"\$(pass_field desk-live-a '.session.session_id')\" != '$old' ]" || true
  save_list
  local new c1 u1
  new=$(pass_field desk-live-a '.session.session_id // empty')
  read -r c1 u1 <<<"$(relaunch_counts)"
  say "pass session was $old, is now $new; counted relaunches $c0 -> $c1, uncounted $u0 -> $u1"
  expect "5 the session was ended and a new one started" test -n "$new" -a "$new" != "$old"
  expect "5 the new session's first prompt holds the words" \
    bash -c "grep -l 'change the scale' \"\$HOME\"/.claude/projects/*/$new.jsonl >/dev/null 2>&1"
  expect "5 the count is unchanged" test "$c1" = "$c0" -a "$((u1 - u0))" -eq 1
  expect "5 never two sessions at once" bash -c "[ \$(jq -s --argjson t \$(cat '$RUN/epoch') '[.[] | [.. | objects | select(.name? == \"$PROJECT-desk-live-a\" and .startedAt? >= \$t and (.state? == \"working\" or .state? == \"blocked\"))] | length] | max' '$RUN'/lists/*.json) -le 1 ]"
}

# ---- merge and close -------------------------------------------------------

# pr_worktree is where a pass's launch made its worktree. The check reads the
# pull requests whose head is the pass's own branch, so the fixture pushes that
# branch, as a pass's executor does.
pr_worktree() { echo "$SCRATCH/.claude/worktrees/$1"; }

# scratch_pr commits one note on the pass's branch in its worktree, pushes the
# branch, opens a pull request from it, and prints the number.
scratch_pr() {
  local pass=$1 wt
  wt=$(pr_worktree "$pass")
  echo "$pass ran." >"$wt/notes-$pass.md"
  git -C "$wt" add "notes-$pass.md"
  git -C "$wt" commit --quiet -m "Add $pass notes"
  git -C "$wt" push --quiet --force origin "$pass"
  gh pr create --repo glw907/cc-desk-scratch --base "$BASE" --head "$pass" --title "Add $pass notes" --body "Fixture pass $pass." |
    sed -e 's|.*/pull/||'
}

# add_pass appends one of extra-passes.json to the run's release file.
add_pass() {
  jq --arg p "$1" --slurpfile x "$REPO/scripts/live/desk/extra-passes.json" \
    '.passes += [$x[0][] | select(.name == $p)]' "$RELEASE" >"$RELEASE.new"
  mv "$RELEASE.new" "$RELEASE"
}

# hold_pr adds the pass, starts it, and reports its pull request ready outside
# the design, so the check asks a merge question.
hold_pr() {
  local pass=$1 n sha
  add_pass "$pass"
  show "$BIN" --state-dir "$STATE" pass start "$PROJECT" "$pass"
  wait_for 300 bash -c "[ -n \"\$(jq -r --arg p '$pass' '.passes[] | select(.name == \$p) | .session.session_id // empty' '$RELEASE')\" ]" || true
  wait_for 120 test -d "$(pr_worktree "$pass")" || true
  n=$(scratch_pr "$pass")
  sha=$(git -C "$(pr_worktree "$pass")" rev-parse HEAD)
  show "$BIN" --state-dir "$STATE" pass ready "$PROJECT" "$pass" --pr "$n" --commit "$sha" --verdict accept --design outside
  echo "$n" >"$RUN/$pass.pr"
  echo "$sha" >"$RUN/$pass.sha"
  say "$pass: pull request $n at $sha"
}

question_of() { pass_field "$1" '.pull_request.question // empty'; }

phase_merge() {
  local s p=desk-merge n sha q q2
  s=$(desk_session)
  hold_pr "$p"
  n=$(cat "$RUN/$p.pr")
  show systemctl --user start --no-block cc-desk-check.service
  wait_for 300 bash -c "[ -n \"\$(jq -r '.passes[] | select(.name == \"$p\") | .pull_request.question // empty' '$RELEASE')\" ]" || true
  q=$(question_of "$p")
  say "merge question: ${q:-none}"
  # Move the head before the answer, so the first answer asks again.
  echo "second push" >>"$(pr_worktree "$p")/notes-$p.md"
  git -C "$(pr_worktree "$p")" commit --quiet -am "Move the head"
  git -C "$(pr_worktree "$p")" push --quiet origin "$p"
  type_in "$s" "Run a fresh brief. For the merge question about the $p pull request, my answer is merge."
  wait_for 300 bash -c "[ \"\$(jq -r '.passes[] | select(.name == \"$p\") | .pull_request.question // empty' '$RELEASE')\" != '$q' ]" || true
  q2=$(question_of "$p")
  expect "5 merge asks again when the head moved" test -n "$q2" -a "$q2" != "$q"
  type_in "$s" "Run a fresh brief. For the new merge question about the $p pull request, my answer is merge."
  wait_for 300 bash -c "[ -n \"\$(gh pr view $n --repo glw907/cc-desk-scratch --json mergedAt -q .mergedAt)\" ]" || true
  sha=$(gh pr view "$n" --repo glw907/cc-desk-scratch --json headRefOid -q .headRefOid)
  expect "5 merge merges at the commit asked about" \
    bash -c "gh pr view $n --repo glw907/cc-desk-scratch --json state,mergeCommit -q '.state' | grep -q MERGED"
  say "head at the merge: $sha"
}

phase_close() {
  local s p=desk-close n
  s=$(desk_session)
  hold_pr "$p"
  n=$(cat "$RUN/$p.pr")
  show systemctl --user start --no-block cc-desk-check.service
  wait_for 300 bash -c "[ -n \"\$(jq -r '.passes[] | select(.name == \"$p\") | .pull_request.question // empty' '$RELEASE')\" ]" || true
  type_in "$s" "Run a fresh brief. For the merge question about the $p pull request, my answer is close."
  wait_for 300 bash -c "[ \"\$(gh pr view $n --repo glw907/cc-desk-scratch --json state -q .state)\" = CLOSED ]" || true
  expect "5 close closes the pull request" bash -c "[ \"\$(gh pr view $n --repo glw907/cc-desk-scratch --json state -q .state)\" = CLOSED ]"
  expect "5 close records the pass closed" bash -c "[ -n \"\$(jq -r '.passes[] | select(.name == \"$p\") | .closed // empty' '$RELEASE')\" ]"
  expect "5 close empties Needs you now for it" \
    bash -c "! '$BIN' --state-dir '$STATE' snapshot | jq -e '.needs_you_now | map(select(.session == \"$PROJECT-$p\")) | length > 0' >/dev/null"
}

# ---- abort -----------------------------------------------------------------

phase_abort() {
  try systemctl --user disable --now cc-desk-check.timer
  local p state
  for p in desk-live-a desk-merge desk-close; do
    state=$(pass_field "$p" '.state' 2>/dev/null || echo unknown)
    if [ "$state" = building ] || [ "$state" = waiting ]; then
      try "$BIN" --state-dir "$STATE" pass stop "$PROJECT" "$p" --reason scope
    else
      say "$p reads $state"
    fi
  done
  local s
  for s in $(tmux list-sessions -F '#{session_name}' 2>/dev/null | grep "^$PROJECT-" || true); do
    try tmux kill-session -t "$s"
  done
}

# ---- evidence --------------------------------------------------------------

phase_evidence() {
  evidence_readings || say "evidence: a reading failed; the rest ran"
  try git -C "$SCRATCH" switch main
  try systemctl --user disable --now cc-desk-check.timer
  local fails missing="" c
  fails=$(grep -c '^fail' "$RESULTS" 2>/dev/null || true)
  while IFS= read -r c; do
    cut -f2 "$RESULTS" 2>/dev/null | grep -Fxq -- "$c" || missing+="$c"$'\n'
  done <<<"$EXPECTED_CLAUSES"
  say
  if [ -n "$missing" ]; then
    say "clauses never recorded:"
    printf '%s' "$missing"
  fi
  if [ "${fails:-0}" -eq 0 ] && [ -z "$missing" ] && [ -s "$RESULTS" ]; then say "result: pass"; else say "result: fail"; fi
}

evidence_readings() {
  try "$BIN" --state-dir "$STATE" timer status
  say "desks.json:"
  jq . "$STATE/desks.json"
  say "check log, last 20 lines:"
  tail -n 20 "$STATE/check-log.jsonl"
  say "rulings, last 4 lines:"
  tail -n 4 "$STATE/rulings.jsonl"
  say "each pass's record:"
  jq '.passes[] | {name, state, pull_request, session: (.session | {session_id, launch_tries, relaunches, ended})}' "$RELEASE"
  say "every clause:"
  cat "$RESULTS"
}

# ---- main ------------------------------------------------------------------

usage() {
  cat <<'EOF'
usage: bash scripts/live-desk.sh <phase>
  prep      refuse on a dirty scratch, a scratch not on main, or another release; build; install the timer; add the release
  open      open the desk in a tmux pane and check the screen, the row, and the arguments
  round     open the round form in a second pane and check its screen
  rule      type an answer with the four characters and check the log line and the snapshot
  start     wait for the check to start the pass, saving the session list every 5 seconds
  refuse    record from the executor's session and from a setsid shell, on a copy of the folder
  replace   type a correction and check the relaunch
  merge     answer "merge" on a scratch pull request, twice
  close     answer "close" on a second scratch pull request
  abort     disable the timer and end every session the run started
  evidence  print the readings, switch scratch to main, disable the timer, print the result
  help      print this list
EOF
}

main() {
  local phase=${1:-help}
  case $phase in
    help | -h | --help)
      usage
      return 0
      ;;
    prep | open | round | rule | start | refuse | replace | merge | close | abort | evidence) ;;
    *)
      usage >&2
      exit 2
      ;;
  esac
  mkdir -p "$RUN"
  exec > >(tee -a "$RUN/run.log") 2>&1
  say "== $phase $(date -u +%Y-%m-%dT%H:%M:%SZ) =="
  "phase_$phase"
}

main "$@"
