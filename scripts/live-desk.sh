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
BASE=live-desk-2026-11-18
RELEASE=$STATE/releases/$PROJECT.json
RESULTS=$RUN/results.tsv
SKILL=$REPO/skills/cc-desk/SKILL.md
# The words the rule phase types: the four characters, and a second line.
WORDS_1="keep the scale, it's \"fine\" at \$5"
WORDS_2="second line"
EXPECTED=$WORDS_1$'\n'$WORDS_2

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
  : >"$RESULTS"
  say "prep: the scratch checkout is clean on main, and the state folder holds no other release"

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

  show git -C "$SCRATCH" fetch origin
  show git -C "$SCRATCH" switch -c "$BASE" origin/main
  mkdir -p "$SCRATCH/docs/plans"
  show cp "$REPO/scripts/live/desk/desk-live-a-plan.md" "$SCRATCH/docs/plans/desk-live-a-plan.md"
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
  show tmux new-session -d -s "$s" -x 200 -y 50 -c "$SCRATCH" "$BIN --state-dir $STATE --skill-file $SKILL"
  local id
  if ! wait_for 60 desk_id_new; then
    check fail "3 desk registered"
  fi
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

# desk_id_new is true once desks.json lists a desk this run did not start before.
desk_id_new() { [ -n "$(desk_id)" ]; }

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
  jq -r '.. | objects | select(.type? == "tool_use" and .name? == "Bash") | .input.command // empty' \
    "$HOME"/.claude/projects/*/"$(cat "$RUN/desk-id")".jsonl 2>/dev/null >"$RUN/desk-bash.txt" || true
  say "desk Bash calls recorded in $RUN/desk-bash.txt"
}

# ---- start -----------------------------------------------------------------

phase_start() {
  local before_t
  before_t=$(date -u +%Y-%m-%dT%H:%M:%SZ)
  echo "$before_t" >"$RUN/start-time"
  started() { grep -q '"decision":"start"' "$STATE/check-log.jsonl" 2>/dev/null && [ -n "$(pass_field desk-live-a '.session.session_id // empty')" ]; }
  local i
  for ((i = 0; i < 60; i++)); do
    save_list
    started && break
    sleep 5
  done
  if started; then check pass "4 the pass starts within 5 minutes and the log holds the start"; else check fail "4 the pass starts within 5 minutes and the log holds the start"; fi
  expect "4 the desk ran no start command" bash -c "! grep -q 'pass start' '$RUN/desk-bash.txt'"
  local sid
  sid=$(pass_field desk-live-a '.session.session_id // empty')
  echo "$sid" >"$RUN/pass-session-1"
  expect "4 the pass session has a row" row_for "$sid"
  save_list
}

# ---- refuse ----------------------------------------------------------------

phase_refuse() {
  local copy=$RUN/state-copy
  rm -rf "$copy"
  show cp -r "$STATE" "$copy"
  # The executor's own session: a background session, so no terminal.
  local own=${CLAUDE_SESSION_ID:-executor}
  local out
  out=$("$BIN" --state-dir "$copy" ruling record --project "$PROJECT" --round desk-live --question dq1 --option keep "from the executor" 2>&1) || true
  say "$out"
  if printf '%s' "$out" | grep -qi 'desk'; then check pass "4 refused from a background session"; else check fail "4 refused from a background session"; fi
  say "(executor session $own)"

  # Real claude under setsid, registered in the copy, records against the copy.
  local id s pid
  id=$(cat /proc/sys/kernel/random/uuid)
  s=$PROJECT-nosetty-$(date -u +%H%M%S)
  tmux new-session -d -s "$s" -x 200 -y 50 -c "$SCRATCH" \
    "setsid -w claude --session-id $id --model sonnet --effort low --permission-mode default --name $PROJECT-nodesk 'Run this command and print what it says: $BIN --state-dir $copy ruling record --project $PROJECT --round desk-live --question dq1 --option keep words'"
  wait_for 60 bash -c "pgrep -f -- '--session-id $id' >/dev/null" || true
  pid=$(pgrep -f -- "--session-id $id" | head -n1 || true)
  say "setsid claude pid: ${pid:-none}"
  if [ -n "$pid" ]; then
    jq --arg id "$id" --argjson pid "$pid" --arg t "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
      '.desks += [{session_id: $id, pid: $pid, started: $t}]' "$copy/desks.json" >"$copy/desks.json.new"
    mv "$copy/desks.json.new" "$copy/desks.json"
  fi
  wait_for 120 bash -c "tmux capture-pane -p -t '$s' | grep -qi 'terminal'" || true
  pane "$s" | tee "$RUN/refuse.pane"
  expect "4 refused from a shell with no terminal, naming it" grep -qi 'terminal' "$RUN/refuse.pane"
  try tmux kill-session -t "$s"
  [ -n "$pid" ] && try kill "$pid"
}

# ---- replace ---------------------------------------------------------------

phase_replace() {
  local s old
  s=$(desk_session)
  old=$(cat "$RUN/pass-session-1" 2>/dev/null || true)
  type_in "$s" "I recorded the wrong thing. Replace my last ruling: I want to change the scale. Record it as my ruling with the replaces flag."
  wait_for 300 bash -c "[ \"\$(pass_field desk-live-a '.session.session_id')\" != '$old' ]" || true
  save_list
  local new relaunch
  new=$(pass_field desk-live-a '.session.session_id // empty')
  relaunch=$(pass_field desk-live-a '.session.relaunches // [] | length')
  say "pass session was $old, is now $new; relaunches recorded: $relaunch"
  expect "5 the session was ended and a new one started" test -n "$new" -a "$new" != "$old"
  expect "5 the new session's first prompt holds the words" \
    bash -c "grep -l 'change the scale' \"\$HOME\"/.claude/projects/*/$new.jsonl >/dev/null 2>&1"
  expect "5 the count is unchanged" test "$(pass_field desk-live-a '.session.launch_tries // 0')" -le 2
  expect "5 never two sessions at once" bash -c "[ \$(jq -s '[.[] | [.. | objects | select(.name? == \"$PROJECT-desk-live-a\")] | length] | max' '$RUN'/lists/*.json) -le 1 ]"
}

# ---- merge and close -------------------------------------------------------

# scratch_pr makes a branch off the base with one commit, pushes it, opens a
# pull request, and prints its number.
scratch_pr() {
  local pass=$1
  git -C "$SCRATCH" fetch --quiet origin
  git -C "$SCRATCH" switch --quiet -c "$pass" "origin/$BASE"
  echo "$pass ran." >"$SCRATCH/notes-$pass.md"
  git -C "$SCRATCH" add "notes-$pass.md"
  git -C "$SCRATCH" commit --quiet -m "Add $pass notes"
  git -C "$SCRATCH" push --quiet origin "$pass"
  git -C "$SCRATCH" switch --quiet "$BASE"
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
  n=$(scratch_pr "$pass")
  sha=$(git -C "$SCRATCH" rev-parse "origin/$pass")
  add_pass "$pass"
  show "$BIN" --state-dir "$STATE" pass start "$PROJECT" "$pass"
  wait_for 300 bash -c "[ -n \"\$(jq -r --arg p '$pass' '.passes[] | select(.name == \$p) | .session.session_id // empty' '$RELEASE')\" ]" || true
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
  git -C "$SCRATCH" switch --quiet "$p"
  echo "second push" >>"$SCRATCH/notes-$p.md"
  git -C "$SCRATCH" commit --quiet -am "Move the head"
  git -C "$SCRATCH" push --quiet origin "$p"
  git -C "$SCRATCH" switch --quiet "$BASE"
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
    bash -c "! '$BIN' --state-dir '$STATE' snapshot | jq -e '.needs_you_now | map(select(.project == \"$PROJECT\")) | length > 0' >/dev/null"
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
  local fails
  fails=$(grep -c '^fail' "$RESULTS" 2>/dev/null || true)
  say
  if [ "${fails:-0}" -eq 0 ] && [ -s "$RESULTS" ]; then say "result: pass"; else say "result: fail"; fi
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
