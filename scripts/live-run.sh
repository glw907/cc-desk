#!/usr/bin/env bash
# shellcheck disable=SC2016  # jq programs in single quotes take their variables from --arg
# live-run.sh: the rehearsal of a whole cc-desk release in the scratch repo, one phase per call.
#
#   bash scripts/live-run.sh                run every phase in order and end with the result
#   bash scripts/live-run.sh <phase>        run one phase (the list is in `help`)
#
# One typed answer on a terminal the script makes starts a pass; the pass opens a pull request;
# the timed check merges it, cleans up, and starts the conductor, which records the round's
# report and then the release's. Each phase prints one line per check, `pass <check>` or
# `fail <check>`, appends all of it to $CLAUDE_JOB_DIR/tmp/live-run/run.log, and a run ends with
# `result: pass` or `result: fail`. A failed check never stops the phases after it, except that a
# phase whose input is missing records one failed check and returns. It reads CC_DESK_STATE_DIR
# (default ~/.local/state/cc-desk), SCRATCH_CHECKOUT (default ~/Projects/cc-desk-scratch), and
# CC_DESK_PAGE (default http://127.0.0.1:7908). Every session and tmux name starts with
# cc-desk-scratch-. The plan is docs/plans/2026-11-18-desk-run-plan.md, "Task 6".
#
# The script never starts, stops, or restarts a unit: the timed check does each step, so each
# phase waits on it. It never runs `claude stop` or `gh pr merge`. The only kill is `tmux
# kill-session` on the desk's recorded name, and the listing reader's recorded pid.
set -euo pipefail

REPO=$(cd "$(dirname "$0")/.." && pwd)
STATE=${CC_DESK_STATE_DIR:-$HOME/.local/state/cc-desk}
SCRATCH=${SCRATCH_CHECKOUT:-$HOME/Projects/cc-desk-scratch}
PAGE=${CC_DESK_PAGE:-http://127.0.0.1:7908}
RUN=${CLAUDE_JOB_DIR:-${TMPDIR:-/tmp}/cc-desk-live}/tmp/live-run
# The binary is the installed one. Prep builds this branch, installs it with `timer install`,
# and checks the installed file is that build.
BIN=${CC_DESK_BINARY:-$HOME/.local/share/cc-desk/bin/cc-desk}
BUILD=$RUN/cc-desk-build
PROJECT=cc-desk-scratch
GH_REPO=glw907/cc-desk-scratch
RULES=$REPO/scripts/live/run/allow-rules.json
RESULTS=$RUN/results.tsv
DENIALS=$HOME/.local/state/claude-denials.jsonl
# STAMP names this run's release, round, pass, and branches; prep writes it and every later
# phase reads it back.
STAMP=$(cat "$RUN/stamp" 2>/dev/null || true)
ROUND=run-$STAMP
PASS=run-$STAMP
NAME=$PROJECT-run-$STAMP
BASE=$(cat "$RUN/base" 2>/dev/null || true)
# The one answer the run types, and the option the desk must record for it.
WORDS="yes, add Double"
OPTION=yes
QUESTION=rq1

say() { printf '%s\n' "$*"; }
show() {
  printf '$ %s\n' "$*"
  "$@"
}
# try is show for a command whose failure is part of the reading.
try() {
  printf '$ %s\n' "$*"
  "$@" || printf '(exit %s)\n' "$?"
}

cc() { "$BIN" --state-dir "$STATE" "$@"; }

# lib.sh holds start_listing, save_agents, and the unit PATH check, shared with the other live
# scripts. Its reader is this script's `listing-loop` phase.
export LISTING_SCRIPT=$REPO/scripts/live-run.sh
# shellcheck source=scripts/live/lib.sh
. "$REPO/scripts/live/lib.sh"

# check prints the verdict of one live check and records it for the result line.
check() {
  local verdict=$1 name=$2
  say "$verdict $name"
  printf '%s\t%s\n' "$verdict" "$name" >>"$RESULTS"
}
# expect runs a test command and prints its check's verdict.
expect() {
  local name=$1
  shift
  if "$@"; then check pass "$name"; else check fail "$name"; fi
}

need_prep() {
  if [ -z "$STAMP" ]; then
    say "run prep first (no stamp in $RUN)"
    exit 1
  fi
}

# rel prints this run's release file: the live one, or else the newest in releases/done/.
rel() {
  local live=$STATE/releases/$PROJECT.json f
  if [ -f "$live" ] && [ "$(jq -r .name "$live")" = "$NAME" ]; then
    echo "$live"
    return 0
  fi
  # shellcheck disable=SC2012  # file names are this run's own, with no odd characters
  f=$(ls -t "$STATE"/releases/done/"$NAME"-*.json 2>/dev/null | head -n 1 || true)
  [ -n "$f" ] && echo "$f"
  return 0
}

pass_field() {
  local f
  f=$(rel)
  [ -n "$f" ] || return 0
  jq -r --arg p "$PASS" ".passes[] | select(.name == \$p) | $1" "$f"
}

# note_ids adds every session id and short id the release's records hold to the run's list, so a
# relaunch's new session is counted whichever poll sees it.
note_ids() {
  local f
  f=$(rel)
  [ -n "$f" ] || return 0
  jq -r '[.passes[]?.session?, .conductor?] | .[] | select(. != null) | .session_id, .short_id' "$f" >>"$RUN/ids" 2>/dev/null || true
}

# wait_for polls a command every 5 seconds until it succeeds or the seconds run out. Every 30
# seconds it saves the session list and notes the ids the records hold.
wait_for() {
  local secs=$1 i
  shift
  for ((i = 0; i * 5 < secs; i++)); do
    "$@" && return 0
    if [ $((i % 6)) -eq 5 ]; then
      mkdir -p "$RUN/lists"
      save_agents "$RUN/lists/$(date -u +%H%M%S).json" || true
      note_ids
    fi
    sleep 5
  done
  return 1
}

secs_def='def secs: sub("\\.[0-9]+"; "") | fromdateiso8601;'

# log_lines prints the check-log lines (JSON) of one record, since a time when $2 is given.
log_lines() {
  [ -f "$STATE/check-log.jsonl" ] || return 0
  jq -c --arg p "$PROJECT" --arg r "$1" --arg s "${2:-1970-01-01T00:00:00Z}" "$secs_def"'
    select(.project == $p and .record == $r and (.time | secs) >= ($s | secs))' "$STATE/check-log.jsonl" 2>/dev/null || true
}

# ---- the desk's screen -----------------------------------------------------

desk_session() { cat "$RUN/desk-session" 2>/dev/null || true; }
pane() { tmux capture-pane -p -t "$1" -S -200; }
pane_has() { pane "$1" | grep -q -- "$2"; }
desk_count() { jq -r '.desks | length' "$STATE/desks.json" 2>/dev/null || echo 0; }
desk_id() { jq -r '.desks[-1].session_id // empty' "$STATE/desks.json" 2>/dev/null || true; }
desk_new() { [ "$(desk_count)" -gt "$(cat "$RUN/desks-before" 2>/dev/null || echo 0)" ]; }

# row_for succeeds when the session list holds a row for the id.
row_for() {
  local f=$RUN/row-for.json
  claude agents --json --all >"$f" 2>/dev/null || return 1
  jq -e --arg id "$1" '[.. | objects | select((.sessionId? // .session_id? // .id?) == $id)] | length > 0' "$f" >/dev/null
}

# pane_stable succeeds when two reads of the pane, 10 seconds apart, match.
pane_stable() {
  local a b
  a=$(pane "$1" | cksum)
  sleep 10
  b=$(pane "$1" | cksum)
  [ "$a" = "$b" ]
}

# type_in sends text literally to a pane, then Enter, and records the entry. Only the rule
# phase calls it, once.
type_in() {
  tmux send-keys -t "$1" -l -- "$2"
  sleep 1
  tmux send-keys -t "$1" Enter
  printf '%s\t%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$2" >>"$RUN/typed.log"
}

# ---- prep ------------------------------------------------------------------

# scratch_rows prints the id, name, and state of each live scratch session in the list, the
# list written to a file first, never through a pipe. A row reading stopped, done, or failed
# is not live.
scratch_rows() {
  local f=$RUN/agents.json
  save_agents "$f" || {
    say "refused: the session list could not be read"
    exit 1
  }
  jq -r '.[]? | select((.name // "") | startswith("cc-desk-scratch-"))
    | select((.state // "") as $s | ["stopped", "done", "failed"] | index($s) == null)
    | "\(.id // .sessionId) \(.name) \(.state)"' "$f"
}

prep_refuse() {
  if [ ! -e "$SCRATCH/.git" ]; then
    say "refused: $SCRATCH is not a git checkout"
    exit 1
  fi
  local dirty branch rows tm left
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
  rows=$(scratch_rows)
  if [ -n "$rows" ]; then
    say "refused: a cc-desk-scratch- session is live:"
    say "$rows"
    exit 1
  fi
  tm=$(tmux list-sessions -F '#{session_name}' 2>/dev/null | grep '^cc-desk-scratch-' || true)
  if [ -n "$tm" ]; then
    say "refused: a cc-desk-scratch- tmux session is live:"
    say "$tm"
    exit 1
  fi
  left=
  git -C "$SCRATCH" show-ref --verify --quiet refs/heads/conductor && left="the branch conductor"
  [ -e "$SCRATCH/.claude/worktrees/conductor" ] && left="$left the worktree .claude/worktrees/conductor"
  if [ -n "$left" ]; then
    say "refused: the scratch checkout holds a left conductor:$left"
    say "a left conductor branch makes the next conductor's launch fail three times"
    exit 1
  fi
  if [ -f "$STATE/releases/$PROJECT.json" ]; then
    say "refused: $STATE holds a live $PROJECT release ($(jq -r .name "$STATE/releases/$PROJECT.json")), and release add takes no second one"
    say "end it with: bash scripts/live-run.sh close (this run's), or cc-desk release done $PROJECT"
    exit 1
  fi
}

# fill puts a repo and a branch into the shared rules.
fill() {
  jq --arg repo "$1" --arg branch "$2" \
    'map(gsub("\\{\\{repo\\}\\}"; $repo) | gsub("\\{\\{branch\\}\\}"; $branch))' "$RULES"
}

PREP_DONE=0
prep_exit() {
  if [ "$PREP_DONE" != 1 ]; then
    say "prep: stopped before its end. The scratch checkout is on '$(git -C "$SCRATCH" branch --show-current)'."
    say "prep: bash scripts/live-run.sh close puts it back on main and ends anything this run made"
  fi
}

phase_prep() {
  prep_refuse
  mkdir -p "$RUN"
  rm -rf "$RUN/lists"
  rm -f "$RUN"/{ids,typed.log,pane.log,ruling-time,desk-session,desk-id,merge-ids,merge-time,listing.out,listing.pid,pass-ids}
  rm -f "$RUN"/listing-*.json
  : >"$RESULTS"
  STAMP=$(date -u +%Y%m%dt%H%M%Sz)
  ROUND=run-$STAMP
  PASS=run-$STAMP
  NAME=$PROJECT-run-$STAMP
  BASE=run-base-$STAMP
  echo "$STAMP" >"$RUN/stamp"
  echo "$BASE" >"$RUN/base"
  date -u +%Y-%m-%dT%H:%M:%SZ >"$RUN/epoch"
  wc -l <"$STATE/rulings.jsonl" 2>/dev/null >"$RUN/rulings-before" || echo 0 >"$RUN/rulings-before"
  trap prep_exit EXIT
  check pass "prep: the scratch checkout is clean on main, no cc-desk-scratch- session is live, no conductor is left, and the state folder holds no live $PROJECT release; run $STAMP"

  show go build -o "$BUILD" "$REPO/cmd/cc-desk"
  show "$BUILD" --state-dir "$STATE" timer install
  expect "prep: the installed binary is this branch's build" cmp -s "$BUILD" "$HOME/.local/share/cc-desk/bin/cc-desk"
  show systemd-analyze --user verify "$HOME/.config/systemd/user/cc-desk-check.service" "$HOME/.config/systemd/user/cc-desk-check.timer"
  if check_unit_path; then
    check pass "prep: claude, gh, git, go, systemd-run, and cc-desk resolve under the unit's PATH"
  else
    check fail "prep: claude, gh, git, go, systemd-run, and cc-desk resolve under the unit's PATH"
  fi

  # The base is cut from the scratch main, carries the toy design and plan, and is the
  # release's default branch, with the checkout on it so the cleanup's pull runs there.
  local exclude
  show git -C "$SCRATCH" fetch origin
  show git -C "$SCRATCH" switch -c "$BASE" origin/main
  mkdir -p "$SCRATCH/docs/specs" "$SCRATCH/docs/plans"
  show cp "$REPO/scripts/live/run/toy-design.md" "$SCRATCH/docs/specs/$ROUND-design.md"
  sed -e "s/{{pass}}/$PASS/g" -e "s/{{base}}/$BASE/g" "$REPO/scripts/live/run/toy-plan.md" >"$SCRATCH/docs/plans/$PASS-plan.md"
  say "wrote $SCRATCH/docs/plans/$PASS-plan.md"
  show git -C "$SCRATCH" add "docs/specs/$ROUND-design.md" "docs/plans/$PASS-plan.md"
  show git -C "$SCRATCH" commit -m "Add the run-$STAMP toy design and plan"
  show git -C "$SCRATCH" push origin "$BASE"
  exclude=$(git -C "$SCRATCH" rev-parse --git-path info/exclude)
  case $exclude in /*) ;; *) exclude=$SCRATCH/$exclude ;; esac
  grep -qx '.claude/worktrees/' "$exclude" 2>/dev/null || echo '.claude/worktrees/' >>"$exclude"

  # The release: the shared allow rules filled with the scratch repo and this pass's branch.
  fill "$GH_REPO" "$PASS" >"$RUN/rules.json"
  sed "s/{{stamp}}/$STAMP/g" "$REPO/scripts/live/run/release.json" |
    jq --slurpfile r "$RUN/rules.json" '.allow_rules = $r[0]' >"$RUN/release.json"
  sed "s/{{stamp}}/$STAMP/g" "$REPO/scripts/live/run/design.json" >"$RUN/design.json"
  show "$BIN" --state-dir "$STATE" release add "$RUN/release.json"
  show "$BIN" --state-dir "$STATE" round design "$PROJECT" "$ROUND" "$RUN/design.json"
  show "$BIN" --state-dir "$STATE" release checkout "$PROJECT" "$SCRATCH" --default-branch "$BASE"
  expect "prep: the snapshot holds one question waiting for $PROJECT" \
    bash -c "'$BIN' --state-dir '$STATE' snapshot | jq -e '[.waits_for_geoff[] | select(.project == \"$PROJECT\" and .round == \"$ROUND\")] | length == 1' >/dev/null"
  PREP_DONE=1
  say "prep: the timer now runs this branch's build, and the page runs the installed one"
  restore_line
}

# ---- open ------------------------------------------------------------------

phase_open() {
  need_prep
  local s id pid
  s=$PROJECT-desk-$(date -u +%H%M%S)
  echo "$s" >"$RUN/desk-session"
  desk_count >"$RUN/desks-before"
  # The desk opens in the run folder, never the scratch checkout: the cleanup's pull skips while
  # a live row's cwd sits in the checkout outside its worktrees.
  show tmux new-session -d -s "$s" -x 200 -y 50 -c "$RUN" "$BIN --state-dir $STATE"
  show tmux pipe-pane -o -t "$s" "cat >> $RUN/pane.log"
  if wait_for 60 desk_new; then check pass "9 the desk registered"; else check fail "9 the desk registered"; fi
  id=$(desk_id)
  echo "$id" >"$RUN/desk-id"
  say "desk session id: ${id:-none}"
  if [ -n "$id" ]; then
    expect "9 the desk has a row in the session list" wait_for 120 row_for "$id"
  else
    check fail "9 the desk has a row in the session list"
  fi
  # Claude Code opens by itself: wait for the opening to name the question, then to settle.
  wait_for 240 pane_has "$s" 'toy package' || true
  wait_for 120 pane_stable "$s" || true
  pane "$s" | tee "$RUN/open.pane" >/dev/null
  say "the desk's opening screen is saved in $RUN/open.pane"
  expect "9 the opening screen holds the waiting question" grep -q 'toy package' "$RUN/open.pane"
  expect "9 the opening screen holds no permission prompt and no trust question" \
    bash -c "! grep -qi 'do you want to\\|trust the files\\|trust this folder' '$RUN/open.pane'"
  pid=$(jq -r '.desks[-1].pid // empty' "$STATE/desks.json" 2>/dev/null || true)
  say "desk process: ${pid:-none}"
}

# ---- rule ------------------------------------------------------------------

phase_rule() {
  need_prep
  local s before after t0 line
  s=$(desk_session)
  if [ -z "$s" ] || ! tmux has-session -t "=$s" 2>/dev/null; then
    check fail "9 the desk is open to type the answer (run open first)"
    return 0
  fi
  before=$(wc -l <"$STATE/rulings.jsonl" 2>/dev/null || echo 0)
  type_in "$s" "$WORDS"
  t0=$(date +%s)
  wait_for 180 ruling_recorded "$before" || true
  after=$(wc -l <"$STATE/rulings.jsonl" 2>/dev/null || echo 0)
  line=$(tail -n 1 "$STATE/rulings.jsonl" 2>/dev/null || true)
  say "rulings.jsonl: $before line(s) before, $after after; $(($(date +%s) - t0)) seconds from typing"
  # The desk ends right after the ruling line, alone, by the name the run recorded.
  try tmux kill-session -t "=$s"
  expect "9 the desk session is gone" bash -c "! tmux has-session -t '=$s' 2>/dev/null"
  if [ "$after" -eq $((before + 1)) ]; then
    check pass "9 the typed answer adds one ruling line"
    jq -r .time <<<"$line" >"$RUN/ruling-time"
  else
    check fail "9 the typed answer adds one ruling line ($before before, $after after)"
    return 0
  fi
  say "the ruling: $line"
  expect "9 the ruling's words are the typed answer, byte for byte" \
    test "$(jq -r .words <<<"$line")" = "$WORDS"
  expect "9 the ruling names $PROJECT, $ROUND, $QUESTION, and option $OPTION" \
    jq -e --arg p "$PROJECT" --arg r "$ROUND" --arg q "$QUESTION" --arg o "$OPTION" \
    '.project == $p and .round == $r and .question == $q and .option == $o' <<<"$line"
  expect "9 the snapshot drops the question" \
    bash -c "! '$BIN' --state-dir '$STATE' snapshot | jq -e '.waits_for_geoff | map(select(.project == \"$PROJECT\")) | length > 0' >/dev/null"

  # The desk's own transcript: it recorded the ruling and started nothing.
  local id t tr=""
  id=$(cat "$RUN/desk-id" 2>/dev/null || true)
  for t in "$HOME"/.claude/projects/*/"$id".jsonl; do
    [ -f "$t" ] && tr=$t && break
  done
  if [ -n "$id" ] && [ -n "$tr" ] &&
    jq -r '.. | objects | select(.type? == "tool_use" and .name? == "Bash") | .input.command // empty' "$tr" >"$RUN/desk-bash.txt"; then
    say "desk Bash calls saved in $RUN/desk-bash.txt from $tr"
    expect "9 the desk ran no start command" bash -c "! grep -q 'pass start' '$RUN/desk-bash.txt'"
  else
    say "no desk transcript found for ${id:-no desk id}"
    check fail "9 the desk ran no start command"
  fi
}

# ruling_recorded succeeds when rulings.jsonl holds more lines than $1.
ruling_recorded() { [ "$(wc -l <"$STATE/rulings.jsonl" 2>/dev/null || echo 0)" -gt "$1" ]; }

# ---- pass ------------------------------------------------------------------

# started succeeds when the check logged a start of the pass after the ruling and the pass has
# a session.
started() {
  local since
  since=$(cat "$RUN/ruling-time")
  log_lines "pass:$PASS" "$since" | jq -e -s 'any(.[]; .decision == "start")' >/dev/null 2>&1 &&
    [ -n "$(pass_field '.session.session_id // empty')" ]
}
pr_ready() { [ -n "$(pass_field '.pull_request.ready // empty')" ]; }

phase_pass() {
  need_prep
  if [ ! -s "$RUN/ruling-time" ]; then
    check fail "9 the pass starts (no ruling was recorded; run rule first)"
    return 0
  fi
  if wait_for 900 started; then
    check pass "9 the check logged a start for the pass after the ruling, and the pass has a session"
  else
    check fail "9 the check logged a start for the pass after the ruling, and the pass has a session"
    return 0
  fi
  say "the check log's lines for the pass:"
  log_lines "pass:$PASS" | tee "$RUN/pass-log.txt"
  local sid
  sid=$(pass_field '.session.session_id // empty')
  note_ids
  echo "$sid" >"$RUN/pass-session-id"
  say "pass session id: $sid"
  expect "9 the pass session has a row" row_for "$sid"

  # The pass plans, builds, gates, opens its pull request, and reports it ready. Its minutes
  # are unmeasured, so the wait is long.
  if wait_for 2400 pr_ready; then
    check pass "9 the pass opened a pull request and reported it ready"
  else
    check fail "9 the pass opened a pull request and reported it ready"
    return 0
  fi
  local n
  n=$(pass_field '.pull_request.number')
  echo "$n" >"$RUN/pr"
  say "pull request $n at $(pass_field '.pull_request.commit // empty'), verdict $(pass_field '.pull_request.verdict // empty')"
  gh pr view "$n" --repo "$GH_REPO" --json baseRefName,headRefName,files >"$RUN/pr.json" || true
  say "the pull request: $(jq -c '{base: .baseRefName, head: .headRefName, files: [.files[]?.path]}' "$RUN/pr.json" 2>/dev/null || echo unreadable)"
  expect "9 the pull request is $PASS into $BASE and changes toy.go and toy_test.go" \
    jq -e --arg b "$BASE" --arg h "$PASS" '.baseRefName == $b and .headRefName == $h
      and ([.files[]?.path] | index("toy.go") != null and index("toy_test.go") != null)' "$RUN/pr.json"
  note_ids
}

# ---- merge -----------------------------------------------------------------

# merge_line prints the check-log line of the pass whose outcome reads "merged pull request".
merge_line() {
  log_lines "pass:$PASS" "$(cat "$RUN/epoch")" |
    jq -c 'select((.outcome // "") | startswith("merged pull request"))' | head -n 1
}
merged_logged() { [ -n "$(merge_line)" ]; }

phase_merge() {
  need_prep
  if [ ! -s "$RUN/pr" ]; then
    check fail "9 the timed check merges the pull request (no pull request; run pass first)"
    return 0
  fi
  local n line sid short
  n=$(cat "$RUN/pr")
  # The timed check does the merge. This phase never starts the unit, so it waits for the timer.
  if wait_for 1500 merged_logged; then
    line=$(merge_line)
    say "the check log's merge line: $line"
    check pass "9 a run of the timed check logged '$(jq -r .outcome <<<"$line")'"
    jq -r .time <<<"$line" >"$RUN/merge-time"
  else
    check fail "9 a run of the timed check logged a merged pull request"
    return 0
  fi
  expect "9 pull request $n reads MERGED on GitHub, merged by the check and not by this script" \
    bash -c "[ \"\$(gh pr view $n --repo $GH_REPO --json state -q .state)\" = MERGED ]"
  say "head at the merge: $(gh pr view "$n" --repo "$GH_REPO" --json headRefOid -q .headRefOid)"
  note_ids
  # The reader starts from the merge and records the pass's session, worktree, and branch.
  sid=$(pass_field '.session.session_id // empty')
  short=$(pass_field '.session.short_id // empty')
  printf '%s\n%s\n' "$sid" "$short" >"$RUN/merge-ids"
  say "merge session: id $short, session id $sid"
  start_listing
}

# ---- listing ---------------------------------------------------------------

# session_rows counts the rows of a saved list that name the session by either id and are not
# stopped, done, or failed.
session_rows() {
  jq --arg a "$1" --arg b "$2" '[.[]? | select(((.id? // "") == $a or (.sessionId? // "") == $a
    or (.id? // "") == $b or (.sessionId? // "") == $b)
    and ((.state? // "") as $s | ["stopped", "done", "failed"] | index($s) == null))] | length' "$3"
}

# phase_listing_loop is the background reader: once a minute for 10 minutes it saves the
# default list to a file and prints one line saying which of the session, the worktree, and
# the branch remain. Its output is only its own lines; the listing phase reads them.
phase_listing_loop() {
  local sid short i n left wt
  sid=$(sed -n 1p "$RUN/merge-ids")
  short=$(sed -n 2p "$RUN/merge-ids")
  wt=$SCRATCH/.claude/worktrees/$PASS
  for ((i = 1; i <= 10; i++)); do
    sleep 60
    if save_agents "$RUN/listing-$i.json"; then
      left=
      n=$(session_rows "$sid" "$short" "$RUN/listing-$i.json")
      [ "$n" = 0 ] || left="session($n row)"
      git -C "$SCRATCH" worktree list --porcelain | grep -qxF "worktree $wt" && left="$left worktree"
      [ -z "$(git -C "$SCRATCH" branch --list "$PASS")" ] || left="$left branch"
      if [ -z "$left" ]; then
        echo "pass minute $i $(date -u +%H:%M:%SZ): session, worktree, and branch are gone"
      else
        echo "fail minute $i $(date -u +%H:%M:%SZ): still here:$left"
      fi
    else
      echo "fail minute $i $(date -u +%H:%M:%SZ): the list could not be read"
    fi
  done
  echo "done"
}

reader_done() { ! kill -0 "$(cat "$RUN/listing.pid")" 2>/dev/null; }

phase_listing() {
  need_prep
  if [ ! -f "$RUN/merge-ids" ]; then
    check fail "9 the listing has a merged session to read (run merge first)"
    return 0
  fi
  start_listing
  if [ -f "$RUN/listing.pid" ]; then
    say "listing: waiting for the reader (pid $(cat "$RUN/listing.pid")) to finish its 10 reads"
    wait_for 900 reader_done || true
  fi
  local line n=0 first=""
  while IFS= read -r line; do
    case $line in
      pass\ minute* | fail\ minute*)
        n=$((n + 1))
        say "listing: ${line#* }"
        [ -z "$first" ] && [ "${line%% *}" = pass ] && first=$n
        ;;
    esac
  done <"$RUN/listing.out"
  if [ "$n" -lt 10 ]; then
    check fail "9 the listing holds 10 reads ($n)"
  else
    check pass "9 the listing holds 10 reads, one a minute from the merge"
  fi
  # Gone from some minute on, and never back: every read from the first clean one reads clean.
  if [ -n "$first" ] && ! tail -n "+$first" <(grep -E '^(pass|fail) minute' "$RUN/listing.out") | grep -q '^fail'; then
    check pass "9 the session, worktree, and branch are gone within 10 minutes of the merge (clean from read $first)"
  else
    check fail "9 the session, worktree, and branch are gone within 10 minutes of the merge"
  fi
}

# ---- report ----------------------------------------------------------------

# conductor_logged succeeds when a log line since the merge launches the conductor with detail
# $1 in its words.
conductor_logged() {
  log_lines conductor "$(cat "$RUN/merge-time")" |
    jq -e -s --arg d "$1" 'any(.[]; .decision == "launch" and .reason == "report_due" and ((.detail // "") | contains($d)))' >/dev/null 2>&1
}
round_launched() { conductor_logged "the report for round $ROUND is due"; }
release_launched() { conductor_logged "the release report is due"; }

# reports_of prints the ids of this release's reports, the round's or the release's.
reports_of() {
  local expr=$1
  [ -f "$STATE/reports.json" ] || return 0
  jq -r --arg p "$PROJECT" --arg n "$NAME" --arg r "$ROUND" "[.reports[]? | select(.project == \$p and .release == \$n and $expr)] | .[].id" "$STATE/reports.json"
}
round_report() { reports_of '.round == $r'; }
release_report() { reports_of '(.round // "") == ""'; }
has_round_report() { [ -n "$(round_report)" ]; }
has_release_report() { [ -n "$(release_report)" ]; }

PAGE_DOWN=0
# listed succeeds when the page lists the report; when the page does not answer, the snapshot
# stands in (standing answer 78) and the first time says so.
listed() {
  local body
  if body=$(curl -fsS --max-time 10 "$PAGE/" 2>/dev/null); then
    grep -q "reports/$1" <<<"$body"
    return
  fi
  if [ "$PAGE_DOWN" = 0 ]; then
    say "the page at $PAGE does not answer; the snapshot stands in for the page"
    PAGE_DOWN=1
  fi
  cc snapshot | jq -e --arg id "$1" '[.reports[]? | select(.id == $id)] | length == 1' >/dev/null
}

# mark_read sends the request the page's button sends, and falls back to the binary when the
# page does not answer.
mark_read() {
  local code
  code=$(curl -sS -o /dev/null -w '%{http_code}' --max-time 15 -X POST -H "Origin: $PAGE" \
    --data '' "$PAGE/reports/$1?/read" 2>/dev/null || echo 000)
  say "mark $1 read: POST $PAGE/reports/$1?/read gave HTTP $code"
  if ! report_is_read "$1"; then
    say "mark $1 read: not read after the request; the binary stands in (standing answer 78)"
    try cc report read "$1"
  fi
}
report_is_read() {
  jq -e --arg id "$1" '[.reports[]? | select(.id == $id and .read != null)] | length == 1' "$STATE/reports.json" >/dev/null 2>&1
}

# conductor_gone succeeds when no conductor worktree or branch is left in the scratch checkout.
conductor_gone() {
  [ ! -e "$SCRATCH/.claude/worktrees/conductor" ] && ! git -C "$SCRATCH" show-ref --verify --quiet refs/heads/conductor
}

phase_report() {
  need_prep
  if [ ! -s "$RUN/merge-time" ]; then
    check fail "9 the conductor starts by itself (no merge; run merge first)"
    return 0
  fi
  local id
  # The reader may have held this phase back, so a conductor that already started is read
  # from the check log and the state, and the waits return at once.
  if wait_for 1500 round_launched; then
    check pass "9 the check launched the conductor for round $ROUND's report (a launch, report_due)"
  else
    check fail "9 the check launched the conductor for round $ROUND's report (a launch, report_due)"
  fi
  note_ids
  wait_for 1200 has_round_report || true
  say "reports of round $ROUND: $(round_report | tr '\n' ' ')"
  id=$(round_report | head -n 1)
  if [ "$(round_report | grep -c .)" = 1 ]; then
    check pass "9 the conductor recorded exactly one report for round $ROUND ($id)"
  else
    check fail "9 the conductor recorded exactly one report for round $ROUND ($(round_report | grep -c .))"
  fi
  [ -n "$id" ] && expect "9 the page lists report $id" listed "$id"

  # The round's report is the first of two: the release's falls due when its last round does.
  if wait_for 1500 release_launched; then
    check pass "9 the check launched the conductor again for the release report (a launch, report_due)"
  else
    check fail "9 the check launched the conductor again for the release report (a launch, report_due)"
  fi
  note_ids
  wait_for 1200 has_release_report || true
  say "release reports: $(release_report | tr '\n' ' ')"
  local rid
  rid=$(release_report | head -n 1)
  if [ "$(release_report | grep -c .)" = 1 ]; then
    check pass "9 the conductor recorded exactly one release report ($rid)"
  else
    check fail "9 the conductor recorded exactly one release report ($(release_report | grep -c .))"
  fi
  [ -n "$rid" ] && expect "9 the page lists report $rid" listed "$rid"

  local launches
  launches=$(log_lines conductor "$(cat "$RUN/merge-time")" | jq -s '[.[] | select(.decision == "launch" and .reason == "report_due")] | length')
  say "conductor launches since the merge: $launches"
  expect "9 two conductor launches, one for each report" test "$launches" -eq 2
  if wait_for 300 conductor_gone; then
    check pass "9 the conductor's worktree and branch are gone"
  else
    check fail "9 the conductor's worktree and branch are gone"
  fi

  # The run ends with nothing unread: the page's own request marks both read.
  for id in $id $rid; do
    [ -n "$id" ] || continue
    mark_read "$id"
    expect "9 report $id is marked read" report_is_read "$id"
  done
  note_ids
}

# ---- evidence --------------------------------------------------------------

# transcripts lists the transcript files of a session id, or of a short id, which is a prefix
# of the id the files are named by.
transcripts() {
  local f
  for f in "$HOME"/.claude/projects/*/"$1"*.jsonl; do
    [ -e "$f" ] && printf '%s\n' "$f"
  done
  return 0
}

# all_ids prints every session id and short id the run's records name: those the polls noted,
# those the release file holds now, those the check log names on a launch, and the desk's.
all_ids() {
  {
    [ -f "$RUN/ids" ] && cat "$RUN/ids"
    local f
    f=$(rel)
    [ -n "$f" ] && jq -r '[.. | objects | .session_id? // empty, .short_id? // empty] | .[]' "$f"
    if [ -f "$STATE/check-log.jsonl" ]; then
      jq -r --arg p "$PROJECT" --arg n "$PASS" --arg s "$(cat "$RUN/epoch")" "$secs_def"'
        select(.project == $p and (.record == "pass:" + $n or .record == "conductor") and (.time | secs) >= ($s | secs))
        | "\(.outcome // "") \(.detail // "")"' "$STATE/check-log.jsonl" 2>/dev/null | grep -oE '\b[0-9a-f]{8}\b' || true
    fi
    [ -f "$RUN/desk-id" ] && cat "$RUN/desk-id"
  } | grep -E '.' | sort -u
}

evidence_readings() {
  local f
  f=$(rel)
  say "the release file: ${f:-none}"
  say "the pass's record:"
  pass_field '{state, started, finished, pull_request, session: (.session | {name, short_id, launched, ended, relaunches, prompts, refusals})}' || true
  say "the check log's lines for the pass:"
  log_lines "pass:$PASS" "$(cat "$RUN/epoch")"
  say "the check log's conductor lines:"
  log_lines conductor "$(cat "$RUN/epoch")"
  say "the listing (default claude agents --json, once a minute from the merge, $(cat "$RUN/merge-time" 2>/dev/null || echo unknown)):"
  cat "$RUN/listing.out" 2>/dev/null || say "(none)"
  try cc timer status
}

phase_evidence() {
  need_prep
  evidence_readings || say "evidence: a reading failed; the rest ran"
  local f ids want got
  f=$(rel)
  if [ -z "$f" ]; then
    check fail "10 the run's release file is readable"
    return 0
  fi

  # The rules the release carried, against the shared source filled with the scratch values.
  want=$(fill "$GH_REPO" "$PASS")
  got=$(jq -S -c '.allow_rules' "$f")
  say "rules: the release carried $(jq -r 'length' <<<"$got") rule(s): $got"
  expect "10 the release's allow rules equal the shared source filled with $GH_REPO and $PASS" \
    test "$(jq -S -c . <<<"$want")" = "$got"
  say "the setup check: bash scripts/setup-release-check.sh"

  # Prompts and refusals, on every record of the run: the pass's session and the conductor's.
  local prompts refusals denied
  prompts=$(jq '[.passes[]?.session?, .conductor?] | map(select(. != null) | (.prompts // []) | length) | add // 0' "$f")
  refusals=$(jq '[.passes[]?.session?, .conductor?] | map(select(. != null) | (.refusals // []) | length) | add // 0' "$f")
  say "prompts recorded: $prompts"
  say "refusals recorded: $refusals"
  jq -r '[.passes[]?.session?, .conductor?] | .[] | select(. != null) | .name as $n | (.refusals // [])[]
    | "  \($n): \(.tool) \(.target): \(.reason)"' "$f"
  ids=$(all_ids)
  denied=0
  if [ -f "$DENIALS" ]; then
    printf '%s\n' "$ids" >"$RUN/ids.txt"
    denied=$(jq -s --rawfile ids "$RUN/ids.txt" '($ids | split("\n") | map(select(. != ""))) as $i
      | [.[] | select(.session? as $x | $x != null and ($i | any(. as $y | $x == $y or ($x | startswith($y)))))] | length' "$DENIALS" 2>/dev/null || echo 0)
  fi
  say "denials log rows for the run's ids: $denied"
  expect "9 no permission prompt on any record of the run (prompts $prompts)" test "$prompts" -eq 0
  expect "9 the refusals recorded ($refusals) equal the denials log's rows for the run's ids ($denied)" test "$refusals" -eq "$denied"

  # No merge command in any transcript of the run's sessions.
  local files t n total=0 read=0
  files=$(for id in $ids; do transcripts "$id"; done | sort -u)
  say "session ids read: $(tr '\n' ' ' <<<"$ids")"
  for t in $files; do
    n=$(jq -r '(.message.content? // []) | if type == "array" then .[] else empty end
      | select(.type == "tool_use" and .name == "Bash") | .input.command // empty
      | select(contains("gh pr merge"))' "$t" | wc -l)
    say "  transcript $t: $n merge call(s)"
    total=$((total + n))
    read=$((read + 1))
  done
  if [ "$read" -gt 0 ] && [ "$total" -eq 0 ]; then
    check pass "9 the transcripts hold no merge command ($read transcripts read, 0 matches)"
  else
    check fail "9 the transcripts hold no merge command ($read transcripts read, $total matches)"
  fi

  # One typed entry, and one ruling, for the run.
  local typed rulings words
  typed=$(wc -l <"$RUN/typed.log" 2>/dev/null || echo 0)
  rulings=$(jq -s --arg p "$PROJECT" --arg r "$ROUND" '[.[] | select(.project == $p and .round == $r)] | length' "$STATE/rulings.jsonl" 2>/dev/null || echo 0)
  words=$(grep -cF -- "$WORDS" "$RUN/pane.log" 2>/dev/null || true)
  say "typed entries: $typed; rulings for $ROUND: $rulings; the pane log ($RUN/pane.log) holds the words on $words line(s)"
  expect "9 one typed entry in the run, and one ruling for the round; nothing was typed after the answer" \
    bash -c "[ '$typed' = 1 ] && [ '$rulings' = 1 ]"
  restore_line
  if [ "${LIVE_RUN_ALL:-}" != 1 ]; then
    result_line
  fi
}

# ---- close -----------------------------------------------------------------

# open_session prints the open session's id on a record, or nothing.
open_session() {
  local f
  f=$(rel)
  [ -n "$f" ] && [ "$f" = "$STATE/releases/$PROJECT.json" ] || return 0
  case $1 in
    pass) jq -r --arg p "$PASS" '.passes[] | select(.name == $p) | select(.session != null and .session.ended == null) | .session.session_id // empty' "$f" ;;
    conductor) jq -r 'select(.conductor != null and .conductor.ended == null) | .conductor.session_id // empty' "$f" ;;
  esac
}

# release_moved succeeds when no live release is left for the project.
release_moved() { [ ! -f "$STATE/releases/$PROJECT.json" ]; }

# snapshot_line prints the counts the run must end on, and succeeds when each is 0.
snapshot_line() {
  local s needs waits waiting sessions unread
  s=$(cc snapshot)
  needs=$(jq '.needs_you_now | length' <<<"$s")
  waits=$(jq '.waits_for_geoff | length' <<<"$s")
  waiting=$(jq '.waiting | length' <<<"$s")
  sessions=$(jq '.sessions | length' <<<"$s")
  unread=$(jq '[.reports[]? | select(.read == null)] | length' <<<"$s")
  say "snapshot: needs $needs, waits $waits, waiting $waiting, sessions $sessions, unread $unread"
  [ "$needs" = 0 ] && [ "$waits" = 0 ] && [ "$waiting" = 0 ] && [ "$sessions" = 0 ] && [ "$unread" = 0 ]
}

# teardown_end ends an open session through the binary. `session end` waits only 10 seconds for
# check.lock and a check run holds it for about 40, so a refusal waits 60 seconds and tries once
# more; a second refusal is a stop (standing answer 72).
teardown_end() {
  printf '$ cc-desk session end %s\n' "$*"
  if cc session end "$@"; then
    return 0
  fi
  say "session end was refused; waiting 60 seconds and trying once more"
  sleep 60
  if ! cc session end "$@"; then
    check fail "close: session end $* was refused twice; this is a stop (standing answer 72): the conductor ends the session"
  fi
}

phase_close() {
  need_prep
  local live=$STATE/releases/$PROJECT.json state s id
  if [ -f "$live" ] && [ "$(jq -r .name "$live")" != "$NAME" ]; then
    say "close: the live release is $(jq -r .name "$live"), not this run's ($NAME); it is left alone"
    check fail "close: the live release is this run's"
    return 0
  fi
  if [ -f "$live" ]; then
    # A pass still building or waiting is stopped first, so the check does not relaunch the
    # session that session end is about to stop.
    state=$(pass_field '.state')
    case $state in
      building | waiting) try cc pass stop "$PROJECT" "$PASS" --reason scope ;;
      *) say "the pass reads $state" ;;
    esac
    note_ids
    id=$(open_session pass)
    if [ -n "$id" ]; then
      teardown_end pass "$PROJECT" "$PASS"
    fi
    id=$(open_session conductor)
    if [ -n "$id" ]; then
      try cc session finish conductor "$PROJECT"
      teardown_end conductor "$PROJECT"
    fi
  else
    say "close: no live $PROJECT release; nothing to end"
  fi
  # The desk, if the run left it open: alone, by the recorded name.
  s=$(desk_session)
  case $s in
    cc-desk-scratch-*)
      if tmux has-session -t "=$s" 2>/dev/null; then try tmux kill-session -t "=$s"; fi
      ;;
  esac
  if [ -f "$RUN/listing.pid" ] && kill -0 "$(cat "$RUN/listing.pid")" 2>/dev/null &&
    tr '\0' ' ' <"/proc/$(cat "$RUN/listing.pid")/cmdline" 2>/dev/null | grep -q 'live-run.sh listing-loop'; then
    say "close: the listing reader (pid $(cat "$RUN/listing.pid")) is still running"
    # shellcheck disable=SC2046  # one recorded pid
    kill $(cat "$RUN/listing.pid")
  fi

  # Every report of the project is marked read, the run's two with the page's own request.
  for id in $(jq -r --arg p "$PROJECT" '.reports[]? | select(.project == $p and .read == null) | .id' "$STATE/reports.json" 2>/dev/null || true); do
    mark_read "$id"
  done
  expect "close: no report of $PROJECT is unread" \
    bash -c "[ \"\$(jq --arg p '$PROJECT' '[.reports[]? | select(.project == \$p and .read == null)] | length' '$STATE/reports.json')\" = 0 ]"

  # release done refuses while a session of the release is live in a fresh list, so it waits
  # for the stopped rows to leave the list.
  if [ -f "$live" ]; then
    local i
    for ((i = 0; i < 6; i++)); do
      if cc release "done" "$PROJECT"; then break; fi
      say "close: release done refused; waiting 15 seconds for the list to clear"
      sleep 15
    done
  fi
  if release_moved; then
    # shellcheck disable=SC2012  # the names are this run's own
    check pass "close: the release moved to $(ls -t "$STATE"/releases/done/"$NAME"-*.json 2>/dev/null | head -n 1 || true)"
  else
    check fail "close: the release $NAME is still live"
  fi

  if [ "$(git -C "$SCRATCH" branch --show-current)" = main ]; then
    check pass "close: the scratch checkout is on main"
  elif [ -z "$(git -C "$SCRATCH" status --porcelain)" ]; then
    try git -C "$SCRATCH" switch main
    expect "close: the scratch checkout is on main" test "$(git -C "$SCRATCH" branch --show-current)" = main
  else
    say "close: the scratch checkout is dirty and not on main; it is left as it is:"
    git -C "$SCRATCH" status --porcelain
    check fail "close: the scratch checkout is on main"
  fi

  expect "close: no question waiting, no report unread, and no session running" snapshot_line
  restore_line
  if [ "${LIVE_RUN_ALL:-}" != 1 ]; then
    result_line
  fi
}

# ---- result ----------------------------------------------------------------

# restore_line prints the warning that stands where the line putting main's build back was.
# Once a run writes a report that carries a release name, main's v0.1.0 build cannot load the
# state folder, so reinstalling it before the pass lands would break the timer and the page.
restore_line() {
  say "warning: main's v0.1.0 build can't load the state folder once a report carries a release name."
  say "do not put main's build back on the timer; the conductor reinstalls main after the pass lands."
}

result_line() {
  if grep -q '^fail' "$RESULTS" 2>/dev/null; then
    say "result: fail"
  elif [ -s "$RESULTS" ]; then
    say "result: pass"
  else
    say "result: fail"
  fi
}

# ---- main ------------------------------------------------------------------

PHASES=(prep open rule pass merge listing report evidence close)

usage() {
  cat <<'EOF'
usage: bash scripts/live-run.sh [phase]
  (none)    run every phase in the order below and end with `result: pass` or `result: fail`
  prep      refuse on a dirty scratch, a live cc-desk-scratch- session, or a left conductor; install this build; add the release
  open      open the desk in a tmux session in the run folder and check its screen
  rule      type the one answer, check the ruling, and end the desk right after the ruling line
  pass      wait for the timed check to start the pass, and for its pull request to be ready
  merge     wait for the timed check to merge the pull request; start the listing reader
  listing   wait for the reader's 10 reads and check the session, worktree, and branch are gone
  report    wait for the conductor's round report and release report, check the page lists each, mark both read
  evidence  check the rules, the prompts and refusals, the transcripts, and the typed entries
  close     end the run's sessions, mark every report read, run release done, put the checkout on main, print the snapshot
  help      print this list
Each wait is for the timed check; no phase starts or restarts a unit.
EOF
}

run_all() {
  : >"$RESULTS"
  local p code
  for p in "${PHASES[@]}"; do
    code=0
    LIVE_RUN_NOTEE=1 LIVE_RUN_ALL=1 bash "$REPO/scripts/live-run.sh" "$p" || code=$?
    if [ "$code" -ne 0 ]; then
      say "phase $p ended early (exit $code)"
      check fail "phase $p ran to its end"
      # A refused prep made nothing, so close would touch what is not this run's.
      [ "$p" = prep ] && break
    fi
  done
  result_line
}

main() {
  local phase=${1:-all}
  case $phase in
    help | -h | --help)
      usage
      return 0
      ;;
    all | prep | open | rule | pass | merge | listing | listing-loop | report | evidence | close) ;;
    run) phase=all ;;
    *)
      usage >&2
      exit 2
      ;;
  esac
  mkdir -p "$RUN"
  if [ "$phase" = listing-loop ]; then
    phase_listing_loop
    return 0
  fi
  # A phase the no-argument run calls writes to that run's stream, which is already logged.
  if [ "${LIVE_RUN_NOTEE:-}" != 1 ]; then
    exec > >(tee -a "$RUN/run.log") 2>&1
  fi
  say "== $phase $(date -u +%Y-%m-%dT%H:%M:%SZ) =="
  if [ "$phase" = all ]; then
    run_all
  else
    "phase_$phase"
  fi
}

main "$@"
