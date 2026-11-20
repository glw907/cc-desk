#!/usr/bin/env bash
# live-repairs.sh: the live run of the desk-check pass, one phase per call.
#
#   bash scripts/live-repairs.sh                run every phase in order and end with the result
#   bash scripts/live-repairs.sh <phase>        run one phase (the list is in `help`)
#
# Each phase prints the commands it ran with their output, and appends all of it to
# $CLAUDE_JOB_DIR/tmp/live-repairs/run.log. Each live check prints `pass <check>` or
# `fail <check>`, and a run ends with `result: pass` or `result: fail`. A failed check never
# stops the phases after it (standing answer 60). It reads CC_DESK_STATE_DIR (default ~/.local/state/cc-desk) and
# SCRATCH_CHECKOUT (default ~/Projects/cc-desk-scratch). All live work is in the scratch
# checkout on main, and every session and tmux name starts with cc-desk-scratch-. The plan is
# docs/plans/2026-11-18-desk-check-plan.md, "Task 1" and "The live release".
set -euo pipefail

REPO=$(cd "$(dirname "$0")/.." && pwd)
STATE=${CC_DESK_STATE_DIR:-$HOME/.local/state/cc-desk}
SCRATCH=${SCRATCH_CHECKOUT:-$HOME/Projects/cc-desk-scratch}
RUN=${CLAUDE_JOB_DIR:-${TMPDIR:-/tmp}/cc-desk-live}/tmp/live-repairs
BIN=$RUN/cc-desk
UNIT_DIR=$HOME/.config/systemd/user
PROJECT=cc-desk-scratch
RELEASE=$STATE/releases/$PROJECT.json
RESULTS=$RUN/results.tsv
DENIALS=$HOME/.local/state/claude-denials.jsonl
FIXTURE=$REPO/e2e/testdata/refusal/hook-input.json
# STAMP names this run's round and passes; prep writes it and every later phase reads it back.
STAMP=$(cat "$RUN/stamp" 2>/dev/null || true)

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

# lib.sh holds start_listing and the kill-target guards, shared with the other live scripts.
# shellcheck source=live/lib.sh
. "$REPO/scripts/live/lib.sh"

# check prints the verdict of one live check and records it for the result line.
check() {
  local verdict=$1 name=$2
  say "$verdict $name"
  printf '%s\t%s\n' "$verdict" "$name" >>"$RESULTS"
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

# known_ids is the file of session ids this run's phases learned, so that prep names them as
# this run's own and not as a stranger's.
known_ids() { echo "$RUN/known-ids"; }

# live_rows prints the id, name, and state of each scratch session the list shows as
# working or blocked and that no record of this run names. The list goes to a file first,
# never through a pipe.
live_rows() {
  mkdir -p "$RUN"
  touch "$(known_ids)"
  claude agents --json --all >"$RUN/agents.json" 2>/dev/null || true
  jq -r --rawfile known "$(known_ids)" '($known | split("\n")) as $k
    | .[]? | select((.name // "" | startswith("cc-desk-scratch-")) and (.state == "working" or .state == "blocked")
      and ((.id // "") as $i | ($k | index($i)) == null))
    | "\(.id) \(.name) \(.state)"' "$RUN/agents.json"
}

# ---- prep ------------------------------------------------------------------

prep_refuse() {
  if [ ! -e "$SCRATCH/.git" ]; then
    say "refused: $SCRATCH is not a git checkout"
    exit 1
  fi
  local dirty branch other rows
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
  rows=$(live_rows)
  if [ -n "$rows" ]; then
    say "refused: a live scratch session this run doesn't name is working or blocked:"
    say "$rows"
    say "a leftover of an earlier probe is ended by: bash scripts/live-repairs.sh abort"
    exit 1
  fi
}

# add_round appends this run's round, its probe pass, and its held pass to the release
# JSON on stdin and prints the result. The held pass only waits, so the scratch project
# never has every pass done and records no release report.
add_round() {
  jq --arg r "repairs-$STAMP" --arg p "refuse-$STAMP" --arg h "hold-$STAMP" '
    .rounds += [{name: $r, summary: "Fixture passes for the repairs live run.", needs_geoff: false,
      no_need_reason: "A fixture run in the scratch repo, with nothing to decide.",
      design_written: true, design: ("docs/plans/" + $p + "-plan.md")}]
    | .passes += [
      {name: $p, summary: "Run one command auto mode refuses.", round: $r, planned_minutes: 10, model: "sonnet", effort: "medium", state: "not_started"},
      {name: $h, summary: "Holds the round open.", round: $r, planned_minutes: 10, model: "sonnet", effort: "medium", state: "not_started"}]'
}

# add_merge_round appends the merge phase's round and its one pass to the release JSON on
# stdin. The round has no other pass, so its report comes due once that pass is done; the
# held pass of repairs-<stamp> keeps the release report from ever coming due.
add_merge_round() {
  jq --arg r "merge-$STAMP" '
    .rounds += [{name: $r, summary: "One fixture pass the timed check merges.", needs_geoff: false,
      no_need_reason: "A fixture run in the scratch repo, with nothing to decide.",
      design_written: true, design: ("docs/plans/" + $r + "-plan.md")}]
    | .passes += [{name: $r, summary: "Waits while the script opens its pull request.", round: $r,
      planned_minutes: 10, model: "sonnet", effort: "medium", state: "not_started"}]'
}

# live_default_branch prints the live scratch release's default branch, or nothing.
live_default_branch() {
  [ -f "$RELEASE" ] && jq -r 'select(.done != true) | .default_branch // empty' "$RELEASE"
  return 0
}

phase_prep() {
  prep_refuse
  mkdir -p "$RUN"
  : >"$RESULTS"
  STAMP=$(date -u +%Y%m%dt%H%M%Sz)
  echo "$STAMP" >"$RUN/stamp"
  say "prep: the scratch checkout is clean on main, the state folder holds no other release, and no scratch session is live; run $STAMP"

  show go build -o "$BIN" "$REPO/cmd/cc-desk"
  show "$BIN" timer install
  show systemd-analyze --user verify "$UNIT_DIR/cc-desk-check.service" "$UNIT_DIR/cc-desk-check.timer"

  local base
  base=$(live_default_branch)
  show git -C "$SCRATCH" fetch origin
  if [ -z "$base" ]; then
    base=live-repairs-$STAMP
    show git -C "$SCRATCH" push origin "origin/main:refs/heads/$base"
    show git -C "$SCRATCH" fetch origin
  fi
  echo "$base" >"$RUN/base"

  # A pass's worktree is cut from origin/<base>, so the plans go there. The checkout
  # returns to main after the commit.
  local exclude
  show git -C "$SCRATCH" switch --detach "origin/$base"
  mkdir -p "$SCRATCH/docs/plans"
  show cp "$REPO/scripts/live/repairs/refuse-plan.md" "$SCRATCH/docs/plans/refuse-$STAMP-plan.md"
  show cp "$REPO/scripts/live/repairs/hold-plan.md" "$SCRATCH/docs/plans/hold-$STAMP-plan.md"
  show cp "$REPO/scripts/live/repairs/merge-plan.md" "$SCRATCH/docs/plans/merge-$STAMP-plan.md"
  show git -C "$SCRATCH" add "docs/plans/refuse-$STAMP-plan.md" "docs/plans/hold-$STAMP-plan.md" "docs/plans/merge-$STAMP-plan.md"
  show git -C "$SCRATCH" commit -m "Add the repairs-$STAMP fixture plans"
  show git -C "$SCRATCH" push origin "HEAD:$base"
  show git -C "$SCRATCH" switch main
  exclude=$(git -C "$SCRATCH" rev-parse --git-path info/exclude)
  case $exclude in /*) ;; *) exclude=$SCRATCH/$exclude ;; esac
  grep -qx '.claude/worktrees/' "$exclude" 2>/dev/null || echo '.claude/worktrees/' >>"$exclude"

  if [ -n "$(live_default_branch)" ]; then
    # One live release per project: this run's round joins it, under the state lock.
    say "prep: a live $PROJECT release exists, so round repairs-$STAMP joins it"
    # shellcheck disable=SC2016  # $1 expands in the inner shell
    show flock "$STATE/lock" bash -c 'add_round_to_release "$1"' _ "$RELEASE"
  else
    local rel=$RUN/release.json
    add_round <"$REPO/scripts/live/repairs/release.json" |
      jq --arg n "$PROJECT-repairs-$STAMP" '.name = $n' >"$rel"
    show cc release add "$rel"
    show cc release checkout "$PROJECT" "$SCRATCH" --default-branch "$base"
  fi
  say "prep: the timer now runs this branch's build; the conductor reinstalls from main once the pass lands"
}

# add_round_to_release edits the release file in place. It runs under flock, in a
# subshell, so the helpers it needs are exported.
add_round_to_release() {
  local file=$1
  add_round <"$file" >"$file.new"
  mv "$file.new" "$file"
}
add_merge_to_release() {
  local file=$1
  add_merge_round <"$file" >"$file.new"
  mv "$file.new" "$file"
}
export -f add_round add_round_to_release add_merge_round add_merge_to_release
export STAMP

# ---- refuse ----------------------------------------------------------------

# hook_input_for prints the kept hook inputs that name the session id.
hook_input_for() {
  grep -lF -- "\"$1\"" "$STATE"/hook-unmatched/*-refusal-*.json 2>/dev/null || true
}

# denial_row prints the workstation's log rows for a session id.
denial_row() {
  [ -f "$DENIALS" ] || return 0
  jq -c --arg s "$1" 'select(.session == $s)' "$DENIALS" 2>/dev/null || true
}

phase_refuse() {
  [ -n "$STAMP" ] || {
    say "refuse: run prep first"
    exit 1
  }
  local pass=refuse-$STAMP

  say "the deny lists the session's settings could hold (a deny rule refuses without the classifier, and the hook doesn't fire for it):"
  local f
  for f in "$SCRATCH"/.claude/settings*.json "$HOME"/.claude/settings.json; do
    if [ -f "$f" ]; then
      say "$f: $(jq -c '.permissions.deny // []' "$f")"
    else
      say "$f: absent"
    fi
  done

  # The check starts a free pass on its own, so a pass already building is expected.
  try cc pass start "$PROJECT" "$pass"
  wait_for 300 bash -c "[ -n \"\$(jq -r --arg p '$pass' '.passes[] | select(.name == \$p) | .session.session_id // empty' '$RELEASE')\" ]" || true
  local sid
  sid=$(pass_field "$pass" '.session.session_id // empty')
  if [ -z "$sid" ]; then
    check fail "7 the probe session started"
    return 0
  fi
  check pass "7 the probe session started"
  echo "$sid" >"$RUN/refuse-session-id"
  printf '%s\n%s\n' "$sid" "$(pass_field "$pass" '.session.short_id // empty')" >>"$(known_ids)"
  say "session id: $sid"
  say "the session's deny list: $(jq -c '.permissions.deny' "$STATE/launch/$PROJECT-$pass.json")"
  say "the session's hooks: $(jq -c '.hooks' "$STATE/launch/$PROJECT-$pass.json")"

  # Matched input lands on the session's record; only input that matched no record is kept
  # under hook-unmatched, so the check reads the record first and the folder second.
  local i found="" recorded=""
  for ((i = 0; i < 120; i++)); do
    recorded=$(pass_field "$pass" '.session.refusals // [] | map(select((.reason // "") != "")) | length')
    found=$(hook_input_for "$sid")
    [ "${recorded:-0}" != 0 ] || [ -n "$found" ] && break
    sleep 5
  done
  if [ "${recorded:-0}" != 0 ]; then
    say "the session's record holds the refusal:"
    pass_field "$pass" '.session.refusals'
    check pass "7 the refusal hook fired with the session id"
    check pass "7 the hook input holds a reason"
  elif [ -n "$found" ]; then
    show cp "$(echo "$found" | head -n1)" "$FIXTURE"
    say "the hook input:"
    jq . "$FIXTURE"
    check pass "7 the refusal hook fired with the session id"
    if [ -n "$(jq -r '.reason // empty' "$FIXTURE")" ]; then
      check pass "7 the hook input holds a reason"
    else
      check fail "7 the hook input holds a reason"
    fi
  else
    check fail "7 the refusal hook fired with the session id"
  fi

  local rows
  rows=$(denial_row "$sid")
  say "the workstation's denials log, rows for $sid:"
  say "${rows:-(none)}"
  if [ -n "$rows" ]; then
    check pass "7 the denials log has a row for the session"
  else
    check fail "7 the denials log has a row for the session"
  fi

  # The pass's own stop goes through the binary; the check's carry-out ends the session.
  try cc pass stop "$PROJECT" "$pass" --reason scope
  say "refuse: the probe pass is stopped; the next check ends session $sid"
}

# ---- merge -----------------------------------------------------------------

# scratch_pr commits one note on the pass's branch in its worktree, pushes the branch, opens
# a pull request from it against the run's base branch, and prints the number.
scratch_pr() {
  local pass=$1 wt base
  wt=$SCRATCH/.claude/worktrees/$pass
  base=$(cat "$RUN/base")
  echo "$pass ran." >"$wt/notes-$pass.md"
  git -C "$wt" add "notes-$pass.md"
  git -C "$wt" commit --quiet -m "Add $pass notes"
  git -C "$wt" push --quiet origin "$pass"
  gh pr create --repo glw907/cc-desk-scratch --base "$base" --head "$pass" --title "Add $pass notes" --body "Fixture pass $pass." |
    sed -e 's|.*/pull/||'
}

# merge_log_lines prints the check-log lines (JSON) that name the pass.
merge_log_lines() {
  [ -f "$STATE/check-log.jsonl" ] || return 0
  jq -c --arg p "$1" 'select((.record // "") | endswith($p))' "$STATE/check-log.jsonl" 2>/dev/null || true
}

# stop_step_logged succeeds when a log line for the pass names the stop step: the merge's
# outcome reads "stopped session <id>" or "session <id> already gone".
stop_step_logged() {
  merge_log_lines "$1" | jq -e 'select((.outcome // "") | test("stopped session|already gone"))' >/dev/null 2>&1
}

merged_recorded() { [ -n "$(pass_field "$1" '.pull_request.merged // empty')" ]; }

# checkout_on_base puts the scratch checkout on the run's base branch, which the check pulls
# after a merge only when the checkout is on the release's default branch. checkout_on_main
# puts it back, as prep requires.
checkout_on_base() {
  local base
  base=$(cat "$RUN/base")
  show git -C "$SCRATCH" switch -C "$base" "origin/$base"
}
checkout_on_main() {
  try git -C "$SCRATCH" switch main
}

phase_merge() {
  [ -n "$STAMP" ] || {
    say "merge: run prep first"
    exit 1
  }
  local pass=merge-$STAMP
  checkout_on_base
  # The release is edited under the state lock, as the mechanism table says.
  # shellcheck disable=SC2016  # $1 expands in the inner shell
  show flock "$STATE/lock" bash -c 'add_merge_to_release "$1"' _ "$RELEASE"
  try cc pass start "$PROJECT" "$pass"
  wait_for 300 bash -c "[ -n \"\$(jq -r --arg p '$pass' '.passes[] | select(.name == \$p) | .session.session_id // empty' '$RELEASE')\" ]" || true
  local sid short
  sid=$(pass_field "$pass" '.session.session_id // empty')
  short=$(pass_field "$pass" '.session.short_id // empty')
  if [ -z "$sid" ]; then
    check fail "7 the merge pass started a session"
    return 0
  fi
  check pass "7 the merge pass started a session"
  # A new merge starts a new listing, so an earlier merge's reads are not taken for this one's.
  rm -f "$RUN/listing.out" "$RUN/listing.pid"
  printf '%s\n%s\n' "$sid" "$short" >"$RUN/merge-ids"
  printf '%s\n%s\n' "$sid" "$short" >>"$(known_ids)"
  say "merge session: id $short, session id $sid"

  wait_for 180 test -d "$SCRATCH/.claude/worktrees/$pass" || true
  local n sha
  n=$(scratch_pr "$pass")
  sha=$(git -C "$SCRATCH/.claude/worktrees/$pass" rev-parse HEAD)
  say "$pass: pull request $n at $sha"
  show cc pass ready "$PROJECT" "$pass" --pr "$n" --commit "$sha" --verdict accept --design inside

  # The timer's own check merges it. The kick only saves the wait for the next tick.
  date -u +%Y-%m-%dT%H:%M:%SZ >"$RUN/merge-start"
  try systemctl --user start --no-block cc-desk-check.service
  if wait_for 900 merged_recorded "$pass"; then
    check pass "7 the timed check merged the pull request"
  else
    check fail "7 the timed check merged the pull request"
  fi
  date -u +%Y-%m-%dT%H:%M:%SZ >"$RUN/merge-time"
  start_listing

  # The cleanup outcome lands in the same run's log; wait for the line that names the step.
  wait_for 120 stop_step_logged "$pass" || true
  say "the check log's lines for $pass:"
  merge_log_lines "$pass"
  if stop_step_logged "$pass"; then
    check pass "7 the check's log shows the stop step"
  else
    check fail "7 the check's log shows the stop step"
  fi
}

# ---- listing ---------------------------------------------------------------

# session_rows counts the rows of the default list that name the session, by either id.
session_rows() {
  jq --arg a "$1" --arg b "$2" '[.. | objects | select((.id? // "") == $a or (.sessionId? // "") == $a
    or (.id? // "") == $b or (.sessionId? // "") == $b)] | length' "$3"
}

# phase_listing_loop is the background reader. Its output is only its own lines; the
# listing phase turns them into checks. It reads the default list, `claude agents --json`
# with no flag, into a file in the run folder, never through a pipe.
phase_listing_loop() {
  local sid short i n
  sid=$(sed -n 1p "$RUN/merge-ids")
  short=$(sed -n 2p "$RUN/merge-ids")
  for ((i = 1; i <= 10; i++)); do
    sleep 60
    if claude agents --json >"$RUN/listing-$i.json" 2>"$RUN/listing-$i.err" && jq -e 'type == "array"' "$RUN/listing-$i.json" >/dev/null 2>&1; then
      n=$(session_rows "$sid" "$short" "$RUN/listing-$i.json")
      if [ "$n" = 0 ]; then
        echo "pass minute $i $(date -u +%H:%M:%SZ): no row for $short"
      else
        echo "fail minute $i $(date -u +%H:%M:%SZ): $n row(s) for $short"
      fi
    else
      echo "fail minute $i $(date -u +%H:%M:%SZ): the list could not be read"
    fi
  done
  echo "done"
}

phase_listing() {
  if [ ! -f "$RUN/merge-ids" ]; then
    check fail "7 the listing has a merged session to read (run merge first)"
    return 0
  fi
  start_listing
  local pid
  pid=$(cat "$RUN/listing.pid")
  say "listing: waiting for the reader (pid $pid) to finish its 10 reads"
  wait_for 900 bash -c "! kill -0 $pid 2>/dev/null" || true
  local line n=0
  while IFS= read -r line; do
    case $line in
      pass\ minute* | fail\ minute*)
        n=$((n + 1))
        check "${line%% *}" "7 listing: ${line#* }"
        ;;
    esac
  done <"$RUN/listing.out"
  if [ "$n" -lt 10 ]; then
    check fail "7 the listing holds 10 reads ($n)"
  fi
}

# ---- limits ----------------------------------------------------------------

# The fixture state folder, the stand-ins, and the throwaway clone are made under
# $RUN/limits/<case>. The real state folder is never read or copied.
LIMITS_PAST=2026-11-01T00:00:00Z
L=

# fx runs the binary on the fixture state folder in $L, with the stand-in usage command.
fx() {
  (cd "$L" && env PATH="$L/bin:$(dirname "$(command -v python3)"):/usr/bin:/bin" \
    "$BIN" --state-dir "$L/state" --usage-command "$L/usage.sh" "$@")
}

# fx_check runs the check with the stand-ins and the real 60-second limit (the flag is left
# at its default). Setup runs pass --now in the past; the timed run takes the real clock.
fx_check() {
  fx "$@" check --claude-command "$L/claude.py" --systemd-run-command "$L/systemd-run.sh" \
    --jobs-dir "$L/jobs" --second-read-wait 1ms --launch-row-wait 50ms --gh-command "$L/gh.sh"
}

limits_setup() {
  local c=$1 pass=lim-$1 real_git
  L=$RUN/limits/$c
  rm -rf "$L"
  mkdir -p "$L/bin" "$L/state" "$L/answers" "$L/lists" "$L/jobs"
  real_git=$(command -v git)
  local f
  for f in claude.py gh.sh usage.sh systemd-run.sh; do
    cp "$REPO/scripts/live/repairs/limits/$f" "$L/$f"
    chmod +x "$L/$f"
  done
  cp "$REPO/e2e/testdata/check-merge/gh.py" "$L/gh.py"
  sed -e "s|@FIXTURE@|$L|" -e "s|@REAL_GIT@|$real_git|" "$REPO/scripts/live/repairs/limits/git.sh" >"$L/bin/git"
  chmod +x "$L/bin/git"
  echo '[]' >"$L/rows.json"
  git clone --quiet --bare "$SCRATCH" "$L/remote.git"
  git clone --quiet "$L/remote.git" "$L/project"
  echo "$L/remote.git" >"$L/remote"
  jq --arg p "$pass" '.project = "proj" | .name = "limits" | .repo = "glw907/proj"
    | .rounds = [{name: "r1", summary: "A round.", needs_geoff: false, no_need_reason: "A fixture.", design_written: true}]
    | .passes = [{name: $p, summary: "A pass.", round: "r1", planned_minutes: 30, state: "not_started"}]' \
    "$REPO/scripts/live/repairs/release.json" >"$L/release.json"
  fx release add "$L/release.json"
  fx release checkout proj "$L/project" --default-branch main
  # The pass is building since the past, so the timed check, an hour on in the clock's sense,
  # reads it as worked.
  jq --arg p "$pass" --arg t "$LIMITS_PAST" '(.passes[] | select(.name == $p)) |= (.state = "building" | .started = $t)' \
    "$L/state/releases/proj.json" >"$L/release.new"
  mv "$L/release.new" "$L/state/releases/proj.json"
  fx_check --now "$LIMITS_PAST" >"$L/setup-check.out" 2>&1
  local wt=$L/project/.claude/worktrees/$pass sha
  if [ ! -d "$wt" ]; then
    say "limits $c: the setup check made no worktree; its output:"
    cat "$L/setup-check.out"
    return 1
  fi
  echo "$pass" >"$wt/notes.md"
  git -C "$wt" add notes.md
  git -C "$wt" -c user.name=Limits -c user.email=limits@example.com -c commit.gpgsign=false commit --quiet -m "Add notes"
  git -C "$wt" push --quiet origin "$pass"
  sha=$(git -C "$wt" rev-parse HEAD)
  jq -n --argjson n 1 --arg s "$sha" '[{number: $n, state: "OPEN", headRefOid: $s, mergedAt: null, mergeCommit: null,
    statusCheckRollup: [{name: "gate", status: "COMPLETED", conclusion: "SUCCESS"}]}]' >"$L/answers/$pass.json"
  fx --now "$LIMITS_PAST" pass ready proj "$pass" --pr 1 --commit "$sha" --verdict accept --design inside
  jq 'map(.state = "done" | .status = "idle")' "$L/rows.json" >"$L/rows.new"
  mv "$L/rows.new" "$L/rows.json"
}

# limits_run runs one case: the control file puts the named call to sleep, and the check
# is timed from the call's start in the stand-in's own log to the line in the fixture
# state folder's check log. Where that log line is stamped at the run's start, the wall
# time of the binary stands in, and the line says which one it used.
limits_run() {
  local c=$1 control=$2
  say "limits $c: the $control call sleeps 120 seconds with a child that holds stdout"
  if ! limits_setup "$c"; then
    check fail "7 limits $c: the fixture set up"
    return 0
  fi
  : >"$L/sleep-$control"
  local t0 t1 code=0 wall
  t0=$(date +%s.%N)
  fx_check >"$L/check.out" 2>&1 || code=$?
  t1=$(date +%s.%N)
  wall=$(awk -v a="$t0" -v b="$t1" 'BEGIN { printf "%.1f", b - a }')
  say "limits $c: the check exited $code after $wall seconds; its output:"
  cat "$L/check.out"
  say "limits $c: the fixture check log:"
  cat "$L/state/check-log.jsonl" 2>/dev/null || say "(no check log)"

  local started logt via elapsed=$wall
  started=$(sed -n 1p "$L/sleep-calls.log" 2>/dev/null | cut -d' ' -f1)
  logt=$(jq -r 'select(.outcome != null) | .time' "$L/state/check-log.jsonl" 2>/dev/null | tail -n1)
  via="the binary's wall time"
  if [ -n "$started" ] && [ -n "$logt" ]; then
    local lg
    lg=$(awk -v a="$started" -v b="$(date -d "$logt" +%s.%N)" 'BEGIN { printf "%.1f", b - a }')
    if awk -v x="$lg" 'BEGIN { exit !(x >= 55 && x <= 90) }'; then
      elapsed=$lg
      via="the check log's line"
    fi
    say "limits $c: call start to log line $lg s; wall $wall s"
  fi
  if [ -z "$started" ]; then
    check fail "7 limits $c: the stand-in's sleeping call ran"
  elif awk -v x="$elapsed" 'BEGIN { exit !(x >= 55 && x <= 90) }'; then
    check pass "7 limits $c: the call ended at the 60-second limit ($elapsed s, by $via)"
  else
    check fail "7 limits $c: the call ended at the 60-second limit ($elapsed s, by $via)"
  fi
  if grep -hqE 'stopped at the .* limit' "$L/check.out" "$L/state/check-log.jsonl" 2>/dev/null; then
    check pass "7 limits $c: the output or log names the limit"
  else
    check fail "7 limits $c: the output or log names the limit"
  fi
  local pid
  pid=$(cat "$L/child.pid" 2>/dev/null || true)
  if [ -n "$pid" ] && [ ! -d "/proc/$pid" ]; then
    check pass "7 limits $c: the stand-in's child $pid is gone"
  else
    check fail "7 limits $c: the stand-in's child ${pid:-none} is gone"
  fi
}

phase_limits() {
  [ -x "$BIN" ] || {
    say "limits: run prep first (no built binary at $BIN)"
    exit 1
  }
  limits_run merge merge
  limits_run fetch fetch
  limits_run pull pull
}

# ---- conductor -------------------------------------------------------------

# conductor_launches counts the check-log lines since the merge that launch a conductor.
conductor_launches() {
  local since
  since=$(cat "$RUN/merge-start" 2>/dev/null || echo 1970-01-01T00:00:00Z)
  [ -f "$STATE/check-log.jsonl" ] || {
    echo 0
    return 0
  }
  jq -s --arg s "$since" '[.[] | select(.project == "cc-desk-scratch" and ((.record // "") | test("conductor"))
    and .decision == "launch" and .time > $s)] | length' "$STATE/check-log.jsonl"
}

# conductor_seen adds the ids of the rows named <project>-conductor to the run's list and
# prints how many distinct ids the list holds.
conductor_seen() {
  claude agents --json --all >"$RUN/agents.json" 2>/dev/null || true
  # A finished conductor of an earlier run stays in the --all list, so only rows started
  # since this run's merge count.
  local since=0
  [ -f "$RUN/merge-start" ] && since=$(date -d "$(cat "$RUN/merge-start")" +%s)
  jq -r --arg n "$PROJECT-conductor" --argjson s "$since" '.[]? | select(.name == $n and (.startedAt // 0) >= $s * 1000) | .id' "$RUN/agents.json" >>"$RUN/conductor-ids"
  sort -u "$RUN/conductor-ids" | grep -c . || true
}

round_reports() {
  jq --arg p "$PROJECT" --arg r "$1" '[.reports[]? | select(.project == $p and .round == $r)] | length' "$STATE/reports.json" 2>/dev/null || echo 0
}

conductor_started() { [ "$(conductor_launches)" -ge 1 ]; }

phase_conductor() {
  [ -n "$STAMP" ] || {
    say "conductor: run prep first"
    exit 1
  }
  local round=merge-$STAMP
  [ -f "$RUN/merge-start" ] || date -u +%Y-%m-%dT%H:%M:%SZ >"$RUN/merge-start"
  : >>"$RUN/conductor-ids"

  # The pass must be done and cleaned for the round's report to come due; the timer's next
  # check then launches the conductor.
  if wait_for 1200 conductor_started; then
    check pass "7 the next check started $PROJECT-conductor"
  else
    check fail "7 the next check started $PROJECT-conductor"
  fi
  conductor_seen >/dev/null
  cat "$RUN/conductor-ids" >>"$(known_ids)"

  local i reports=0
  for ((i = 0; i < 240; i++)); do
    reports=$(round_reports "$round")
    [ "$reports" -ge 1 ] && break
    conductor_seen >/dev/null
    sleep 5
  done
  say "conductor: $reports report(s) recorded for round $round"
  if [ "$reports" = 1 ]; then
    check pass "7 one report recorded for round $round"
  else
    check fail "7 one report recorded for round $round ($reports)"
  fi

  # The worktree goes when the conductor ends with its report recorded.
  local gone=no
  for ((i = 0; i < 60; i++)); do
    if [ ! -e "$SCRATCH/.claude/worktrees/conductor" ] && ! git -C "$SCRATCH" show-ref --verify --quiet refs/heads/conductor; then
      gone=yes
      break
    fi
    conductor_seen >/dev/null
    sleep 5
  done
  if [ "$gone" = yes ]; then
    check pass "7 the conductor's worktree and branch are gone"
  else
    check fail "7 the conductor's worktree and branch are gone"
  fi

  # Two more ticks of the check with the report recorded: no second conductor starts.
  for ((i = 0; i < 24; i++)); do
    conductor_seen >/dev/null
    sleep 5
  done
  local ids launches
  ids=$(sort -u "$RUN/conductor-ids" | grep -c . || true)
  launches=$(conductor_launches)
  say "conductor: $ids distinct session id(s) named $PROJECT-conductor, $launches conductor launch line(s) in the check log since the merge"
  if [ "$ids" -le 1 ] && [ "$launches" -le 1 ]; then
    check pass "7 no second conductor started"
  else
    check fail "7 no second conductor started"
  fi
  cat "$RUN/conductor-ids" >>"$(known_ids)"
  checkout_on_main
}

# ---- abort -----------------------------------------------------------------

# this_runs_rows prints the live scratch rows (id and name) that belong to a probe, hold, or
# merge pass of any run of this script, or to the scratch conductor of this run.
this_runs_rows() {
  claude agents --json --all >"$RUN/agents.json" 2>/dev/null || true
  jq -r --arg c "$PROJECT-conductor" '.[]? | select((.state == "working" or .state == "blocked")
    and ((.name // "") | test("^cc-desk-scratch-(refuse|hold|merge)-[0-9]{8}t[0-9]{6}z$") or . == $c))
    | "\(.id) \(.name) \(.state)"' "$RUN/agents.json"
}

no_live_rows() { [ -z "$(this_runs_rows)" ]; }

# Abort stops each pass still building or waiting, so the check does not relaunch it, then
# ends every open session through `session end`, which stops it whatever reason the stop
# carried (the check leaves a stopped pass's session alone). The only kill is of the listing
# reader's recorded pid.
phase_abort() {
  [ -x "$BIN" ] || show go build -o "$BIN" "$REPO/cmd/cc-desk"
  mkdir -p "$RUN"
  [ "$(git -C "$SCRATCH" branch --show-current)" = main ] || checkout_on_main
  local p state open
  if [ -f "$RELEASE" ]; then
    for p in $(jq -r '.passes[] | select(.name | test("^(refuse|hold|merge)-[0-9]{8}t[0-9]{6}z$")) | .name' "$RELEASE"); do
      state=$(pass_field "$p" '.state')
      case $state in
        building | waiting)
          try cc pass stop "$PROJECT" "$p" --reason scope
          ;;
        *) say "$p reads $state" ;;
      esac
      open=$(pass_field "$p" 'select(.session != null and .session.ended == null) | .session.session_id // empty')
      if [ -n "$open" ]; then
        try cc session end pass "$PROJECT" "$p"
      fi
    done
  else
    say "abort: no $PROJECT release in the state folder"
  fi
  if [ -s "$RUN/conductor-ids" ] && [ -n "$(this_runs_rows | grep -F "$PROJECT-conductor" || true)" ]; then
    try cc session finish conductor "$PROJECT"
    try cc session end conductor "$PROJECT"
  fi

  if [ -f "$RUN/listing.pid" ] && kill -0 "$(cat "$RUN/listing.pid")" 2>/dev/null; then
    say "abort: the listing reader (pid $(cat "$RUN/listing.pid")) is still running"
    # shellcheck disable=SC2046  # one recorded pid
    kill $(cat "$RUN/listing.pid")
  fi

  local rows
  wait_for 420 no_live_rows || true
  rows=$(this_runs_rows)
  if [ -z "$rows" ]; then
    say "abort: no live row of a probe, hold, merge, or conductor session remains"
  else
    say "abort: still live after the wait, which only the check's carry-out can end:"
    say "$rows"
  fi
}

# ---- evidence --------------------------------------------------------------

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

# evidence_readings runs with errexit off, so one failed reading never hides the rest.
evidence_readings() {
  say "the refuse probe's session: $(cat "$RUN/refuse-session-id" 2>/dev/null || echo none)"
  local p
  for p in refuse-$STAMP hold-$STAMP merge-$STAMP; do
    say "$p:"
    pass_field "$p" '{state, started, finished, pull_request, session: (.session | {name, short_id, launched, ended, finished, relaunches})}' 2>/dev/null || say "(no record)"
  done
  say
  say "the check log's lines for this run's passes:"
  for p in refuse-$STAMP hold-$STAMP merge-$STAMP; do
    merge_log_lines "$p"
  done
  say
  say "conductor launches since the merge: $(conductor_launches); reports for round merge-$STAMP: $(round_reports "merge-$STAMP")"
  say "scratch conductor row ids seen: $(sort -u "$RUN/conductor-ids" 2>/dev/null | tr '\n' ' ')"
  say
  say "the listing (default claude agents --json, once a minute from the merge, $(cat "$RUN/merge-time" 2>/dev/null || echo unknown)):"
  cat "$RUN/listing.out" 2>/dev/null || say "(none)"
  local f
  for f in "$RUN"/listing-*.json; do
    [ -e "$f" ] && say "$(basename "$f"): $(jq -c '[.[]? | select((.name // "") | startswith("cc-desk-scratch-")) | {id, name, state}]' "$f" 2>/dev/null)"
  done
  say
  try cc timer status
}

phase_evidence() {
  evidence_readings || say "evidence: a reading failed; the rest ran"
  say
  say "the checks of this run:"
  cat "$RESULTS" 2>/dev/null || true
  say
  restore_line
  if [ "${LIVE_REPAIRS_ALL:-}" != 1 ]; then
    result_line
  fi
}

# ---- main ------------------------------------------------------------------

PHASES=(prep refuse merge conductor limits listing abort evidence)

usage() {
  cat <<'EOF'
usage: bash scripts/live-repairs.sh [phase]
  (none)    run every phase in the order below and end with `result: pass` or `result: fail`
  prep      refuse on a dirty scratch, another release, or a live scratch session; build the binary; join or add the release
  refuse    run the probe: a pass whose session runs a command auto mode refuses; keep the hook input
  merge     add a one-pass round, report its pull request ready, and wait for the timed check to merge it; start the listing
  conductor wait for the next check to start the scratch conductor and for its one report; check no second one, worktree gone
  limits    run the three sleeping stand-ins at the real 60 seconds against a fixture state folder
  listing   wait for the background reader (10 reads of the default list, one a minute from the merge) and check each
  abort     stop this run's sessions and any leftover probe's, through the binary only
  evidence  print each clause's evidence and the line that puts main's build back
  help      print this list
The listing starts inside merge and runs beside conductor and limits, so the no-argument run
calls the phases in the order shown above, which collects the listing after them.
EOF
}

run_all() {
  : >"$RESULTS"
  local p code
  for p in "${PHASES[@]}"; do
    code=0
    LIVE_REPAIRS_NOTEE=1 LIVE_REPAIRS_ALL=1 bash "$REPO/scripts/live-repairs.sh" "$p" || code=$?
    if [ "$code" -ne 0 ]; then
      say "phase $p ended early (exit $code)"
      check fail "phase $p ran to its end"
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
    all | prep | refuse | merge | conductor | limits | listing | listing-loop | abort | evidence) ;;
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
  if [ "${LIVE_REPAIRS_NOTEE:-}" != 1 ]; then
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
