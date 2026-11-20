#!/usr/bin/env bash
# live-recover.sh: the live run of the recovery checks (acceptance item 8), one phase per call.
#
#   bash scripts/live-recover.sh <prep|conductor|ask|pass|evidence|close|help>
#   bash scripts/live-recover.sh conductor --after
#
# The phases run in that order. One release holds all three rounds from prep, and the two-pass
# limit orders the passes: the merge and work passes start first together, and the ask pass
# starts once the merge frees a slot, so the ask phase runs right after the conductor phase.
# Each prints the commands it ran with their output and appends
# all of it to $CLAUDE_JOB_DIR/tmp/live-recover/run.log. Each live check prints `pass <check>` or
# `fail <check>`; close ends the run with `result: pass` or `result: fail`. A failed check never
# stops the phases after it. It reads CC_DESK_STATE_DIR (default ~/.local/state/cc-desk) and
# SCRATCH_CHECKOUT (default ~/Projects/cc-desk-scratch). Every session and tmux name starts
# cc-desk-scratch-. The plan is docs/plans/2026-11-18-desk-run-plan.md, "Task 7" and the three
# recovery rows of the mechanism table; docs/research/2026-11-20-kill-probe.md settles how a
# death is drawn.
#
# A death is drawn in the form standing answer 70 settled. A pass dies through
# `cc-desk session end pass`, which the script runs itself, and the next check counts it as a
# `relaunch` with reason `missing`. A lone `kill` of a pass comes back under a new pid, so it
# proves nothing. The conductor dies by a `kill`: the conductor phase prints the pid and
# stops, the executor runs `kill <pid>` alone, and then runs `conductor --after`.
#
# No phase starts, stops, or restarts a systemd unit. Each waits on the timed check, which runs
# at :x4:41 and :x9:41 UTC. No phase runs `claude stop`.
set -euo pipefail

REPO=$(cd "$(dirname "$0")/.." && pwd)
STATE=${CC_DESK_STATE_DIR:-$HOME/.local/state/cc-desk}
SCRATCH=${SCRATCH_CHECKOUT:-$HOME/Projects/cc-desk-scratch}
RUN=${CLAUDE_JOB_DIR:-${TMPDIR:-/tmp}/cc-desk-live}/tmp/live-recover
BIN=$RUN/cc-desk
UNIT_DIR=$HOME/.config/systemd/user
PROJECT=cc-desk-scratch
RELEASE=$STATE/releases/$PROJECT.json
RESULTS=$RUN/results.tsv
PAGE=http://127.0.0.1:7908
CONDUCTOR=$PROJECT-conductor
# STAMP names this run's rounds and passes; prep writes it and every later phase reads it back.
STAMP=$(cat "$RUN/stamp" 2>/dev/null || true)
MERGE=merge-$STAMP
WORK=work-$STAMP
ASK=ask-$STAMP

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

# lib.sh holds save_agents and the kill-target guards, shared with the other live scripts.
# shellcheck source=live/lib.sh
. "$REPO/scripts/live/lib.sh"

# check prints the verdict of one live check and records it for the result line.
check() {
  local verdict=$1 name=$2
  say "$verdict $name"
  printf '%s\t%s\n' "$verdict" "$name" >>"$RESULTS"
}

pass_field() {
  jq -r --arg p "$1" ".passes[] | select(.name == \$p) | $2" "$RELEASE" 2>/dev/null || true
}

conductor_field() {
  jq -r ".conductor | $1" "$RELEASE" 2>/dev/null || true
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

# epoch prints a timestamp as seconds since 1970.
epoch() { date -d "$1" +%s; }

need_stamp() {
  [ -n "$STAMP" ] || {
    say "$1: run prep first"
    exit 1
  }
}

# live_rows_jq is the jq test for a live row: listed, and not reading stopped, done, or failed.
# shellcheck disable=SC2016  # a jq program, not a shell expansion
LIVE_JQ='(.state // "") as $s | $s != "stopped" and $s != "done" and $s != "failed"'

# scratch_live_rows prints the id, name, and state of each live scratch row in the list file $1.
scratch_live_rows() {
  jq -r "[.[]? | select((.name // \"\" | startswith(\"cc-desk-scratch-\")) and ($LIVE_JQ))] | .[] | \"\\(.id) \\(.name) \\(.state)\"" "$1"
}

no_live_rows() {
  save_agents "$RUN/now.json" || return 1
  [ -z "$(scratch_live_rows "$RUN/now.json")" ]
}

# ---- the page --------------------------------------------------------------

# page_json writes what the page draws to the file $1 and says which source it read. The page
# is client-rendered, so its data source, GET /api/state, is what a script can read; when it
# does not answer, the snapshot stands in and the line says so (standing answer 78).
page_json() {
  local code
  code=$(curl -s -m 10 -o /dev/null -w '%{http_code}' "$PAGE/" 2>/dev/null || true)
  if [ "$code" = 200 ] && curl -fsS -m 10 "$PAGE/api/state" >"$1" 2>/dev/null && jq -e '.snapshot' "$1" >/dev/null 2>&1; then
    echo page
    return 0
  fi
  if cc snapshot >"$1.snap" 2>/dev/null && jq '{snapshot: .}' "$1.snap" >"$1" 2>/dev/null; then
    echo "snapshot (the page answered ${code:-nothing})"
    return 0
  fi
  echo none
  return 1
}

# ---- prep ------------------------------------------------------------------

prep_refuse() {
  if [ ! -e "$SCRATCH/.git" ]; then
    say "refused: $SCRATCH is not a git checkout"
    exit 1
  fi
  local dirty branch rows
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
  if [ -f "$RELEASE" ] && jq -e '.done != true' "$RELEASE" >/dev/null 2>&1; then
    say "refused: a live $PROJECT release is in $STATE; end it first with: cc-desk release done $PROJECT"
    exit 1
  fi
  if [ -e "$SCRATCH/.claude/worktrees/conductor" ] || git -C "$SCRATCH" show-ref --verify --quiet refs/heads/conductor; then
    say "refused: the scratch checkout holds a conductor branch or worktree, which makes the conductor's launch fail"
    exit 1
  fi
  mkdir -p "$RUN"
  if ! save_agents "$RUN/now.json"; then
    say "refused: the session list could not be read"
    exit 1
  fi
  rows=$(scratch_live_rows "$RUN/now.json")
  if [ -n "$rows" ]; then
    say "refused: a live scratch session is listed:"
    say "$rows"
    exit 1
  fi
}

# add_rounds appends the three rounds and their passes to the release JSON on stdin, so one
# release holds all of them from the start. The two-pass limit orders them: the merge pass and
# the work pass start first together, and the ask round waits on the merge round, so the ask
# pass starts once the merge frees a slot.
add_rounds() {
  jq --arg m "merge-$STAMP" --arg w "work-$STAMP" --arg a "ask-$STAMP" '
    .rounds += [
      {name: $m, summary: "One fixture pass the timed check merges.", needs_geoff: false,
        no_need_reason: "A fixture run in the scratch repo, with nothing to decide.",
        design_written: true, design: ("docs/plans/" + $m + "-plan.md")},
      {name: $w, summary: "One fixture pass that keeps working.", needs_geoff: false,
        no_need_reason: "A fixture run in the scratch repo, with nothing to decide.",
        design_written: true, design: ("docs/plans/" + $w + "-plan.md")},
      {name: $a, summary: "One fixture pass whose command always asks.", needs_geoff: false,
        no_need_reason: "A fixture run in the scratch repo, with nothing to decide.",
        design_written: true, design: ("docs/plans/" + $a + "-plan.md"), after: [$m]}]
    | .passes += [
      {name: $m, summary: "Waits while the script opens its pull request.", round: $m,
        planned_minutes: 10, model: "sonnet", effort: "medium", state: "not_started"},
      {name: $w, summary: "Runs the scratch gate, step after step.", round: $w,
        planned_minutes: 60, model: "sonnet", effort: "medium", state: "not_started"},
      {name: $a, summary: "Runs the one command the release says always asks.", round: $a,
        planned_minutes: 10, model: "sonnet", effort: "medium", state: "not_started"}]'
}

# saver_loop is the background reader: it saves the session list every 30 seconds until the
# stop file appears. It ends itself, so nothing kills it.
phase_saver_loop() {
  mkdir -p "$RUN/lists"
  while [ ! -e "$RUN/saver.stop" ]; do
    local f i
    f=$RUN/lists/$(date -u +%Y%m%dT%H%M%S).json
    if save_agents "$f.tmp"; then
      mv "$f.tmp" "$f"
    else
      rm -f "$f.tmp" "$f.tmp.err"
    fi
    for i in 1 2 3 4 5 6; do
      [ -e "$RUN/saver.stop" ] && break
      sleep 5
    done
  done
}

start_saver() {
  rm -f "$RUN/saver.stop"
  if [ -f "$RUN/saver.pid" ] && kill -0 "$(cat "$RUN/saver.pid")" 2>/dev/null; then
    say "saver: already running, pid $(cat "$RUN/saver.pid")"
    return 0
  fi
  setsid nohup bash "$REPO/scripts/live-recover.sh" saver-loop >>"$RUN/saver.out" 2>&1 </dev/null &
  echo $! >"$RUN/saver.pid"
  say "saver: saving the session list every 30 seconds, pid $(cat "$RUN/saver.pid"); close ends it through $RUN/saver.stop"
}

PREP_DONE=0
prep_exit() {
  if [ "$PREP_DONE" != 1 ]; then
    say "prep: stopped before its end. The scratch checkout is on '$(git -C "$SCRATCH" branch --show-current)'."
    say "prep: bash scripts/live-recover.sh close puts it back on main"
  fi
}

phase_prep() {
  prep_refuse
  trap prep_exit EXIT
  : >"$RESULTS"
  STAMP=$(date -u +%Y%m%dt%H%M%Sz)
  echo "$STAMP" >"$RUN/stamp"
  MERGE=merge-$STAMP WORK=work-$STAMP ASK=ask-$STAMP
  rm -rf "$RUN/lists" "$RUN/conductor"
  rm -f "$RUN/work-ids" "$RUN/deaths.tsv" "$RUN/saver.stop"
  say "prep: the scratch checkout is clean on main, the state folder holds no live $PROJECT release, and no scratch session is listed; run $STAMP"

  show go build -o "$BIN" "$REPO/cmd/cc-desk"
  show "$BIN" timer install
  show systemd-analyze --user verify "$UNIT_DIR/cc-desk-check.service" "$UNIT_DIR/cc-desk-check.timer"
  if ! check_unit_path; then
    say "refused: a command the check needs does not resolve under the unit's PATH"
    exit 1
  fi

  local base=recover-base-$STAMP exclude
  echo "$base" >"$RUN/base"
  show git -C "$SCRATCH" fetch origin
  show git -C "$SCRATCH" push origin "origin/main:refs/heads/$base"
  show git -C "$SCRATCH" fetch origin

  # A pass's worktree is cut from origin/<base>, so the plans and the ask probe's file go
  # there. The checkout returns to main after the commit.
  show git -C "$SCRATCH" switch --detach "origin/$base"
  mkdir -p "$SCRATCH/docs/plans"
  show cp "$REPO/scripts/live/repairs/merge-plan.md" "$SCRATCH/docs/plans/$MERGE-plan.md"
  show cp "$REPO/scripts/live/recover/work-plan.md" "$SCRATCH/docs/plans/$WORK-plan.md"
  show cp "$REPO/scripts/live/recover/ask-plan.md" "$SCRATCH/docs/plans/$ASK-plan.md"
  echo "the ask probe's file" >"$SCRATCH/ask-probe.txt"
  show git -C "$SCRATCH" add "docs/plans/$MERGE-plan.md" "docs/plans/$WORK-plan.md" "docs/plans/$ASK-plan.md" ask-probe.txt
  show git -C "$SCRATCH" commit -m "Add the recover-$STAMP fixture plans"
  show git -C "$SCRATCH" push origin "HEAD:$base"
  show git -C "$SCRATCH" switch main
  exclude=$(git -C "$SCRATCH" rev-parse --git-path info/exclude)
  case $exclude in /*) ;; *) exclude=$SCRATCH/$exclude ;; esac
  grep -qx '.claude/worktrees/' "$exclude" 2>/dev/null || echo '.claude/worktrees/' >>"$exclude"

  local rel=$RUN/release.json
  add_rounds <"$REPO/scripts/live/recover/release.json" |
    jq --arg n "$PROJECT-recover-$STAMP" '.name = $n' >"$rel"
  show "$BIN" --state-dir "$STATE" release add "$rel"
  show "$BIN" --state-dir "$STATE" release checkout "$PROJECT" "$SCRATCH" --default-branch "$base"
  start_saver
  say "prep: the timer now runs this branch's build; the conductor reinstalls from main once the pass lands"
  PREP_DONE=1
}

# ---- conductor -------------------------------------------------------------

# checkout_on_base puts the scratch checkout on the run's base branch, which the check pulls
# after a merge only when the checkout is on the release's default branch. It runs alone, on a
# clean checkout. close puts it back on main.
checkout_on_base() {
  local base
  base=$(cat "$RUN/base")
  if [ -n "$(git -C "$SCRATCH" status --porcelain)" ]; then
    say "refused: the scratch checkout is dirty; no switch"
    exit 1
  fi
  [ "$(git -C "$SCRATCH" branch --show-current)" = "$base" ] || show git -C "$SCRATCH" switch -C "$base" "origin/$base"
}

# scratch_pr commits one note on the pass's branch in its worktree, pushes the branch, opens a
# pull request from it against the run's base branch, and prints the number.
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

have_session() { [ -n "$(pass_field "$1" '.session.session_id // empty')" ]; }
merged_recorded() { [ -n "$(pass_field "$1" '.pull_request.merged // empty')" ]; }
conductor_open() { [ -n "$(conductor_field 'select(.ended == null) | .session_id // empty')" ]; }

# log_lines prints the check-log lines (time, decision, reason, outcome) for a record, from
# the epoch $2 on.
log_lines() {
  local record=$1 since=$2 t d r o
  [ -f "$STATE/check-log.jsonl" ] || return 0
  while IFS=$'\t' read -r t d r o; do
    [ -n "$t" ] || continue
    [ "$(epoch "$t")" -ge "$since" ] && printf '%s  %s  %s  %s\n' "$t" "$d" "${r:--}" "${o:--}"
  done < <(jq -r --arg r "$record" 'select(.project == "cc-desk-scratch" and .record == $r)
    | [.time, .decision, (.reason // ""), (.outcome // "")] | @tsv' "$STATE/check-log.jsonl" 2>/dev/null)
}

phase_conductor_before() {
  need_stamp conductor
  mkdir -p "$RUN/conductor"
  checkout_on_base

  # The check starts the merge pass on its own, since it is free.
  if wait_for 1800 have_session "$MERGE"; then
    check pass "8 the check started the merge pass"
  else
    check fail "8 the check started the merge pass"
    return 0
  fi
  say "merge session: $(pass_field "$MERGE" '.session.short_id') ($(pass_field "$MERGE" '.session.session_id'))"
  wait_for 180 test -d "$SCRATCH/.claude/worktrees/$MERGE" || true
  local n sha
  n=$(scratch_pr "$MERGE")
  sha=$(git -C "$SCRATCH/.claude/worktrees/$MERGE" rev-parse HEAD)
  say "$MERGE: pull request $n at $sha"
  show "$BIN" --state-dir "$STATE" pass ready "$PROJECT" "$MERGE" --pr "$n" --commit "$sha" --verdict accept --design inside

  # The timed check merges it; nothing here starts the unit.
  date -u +%Y-%m-%dT%H:%M:%SZ >"$RUN/conductor/merge-start"
  if wait_for 1800 merged_recorded "$MERGE"; then
    check pass "8 the timed check merged the pull request"
  else
    check fail "8 the timed check merged the pull request"
    return 0
  fi

  # The next check launches the conductor for the round's report.
  if ! wait_for 1500 conductor_open; then
    check fail "8 the next check started $CONDUCTOR"
    return 0
  fi
  check pass "8 the next check started $CONDUCTOR"
  local sid launched now
  sid=$(conductor_field '.session_id')
  launched=$(conductor_field '.launched')
  say "conductor: session id $sid, short id $(conductor_field '.short_id'), launched $launched, for report '$(conductor_field '.for_report // ""')'"

  # The list is read until the conductor's row holds a process id that passes the guards.
  local pid="" i
  for ((i = 0; i < 12; i++)); do
    if save_agents "$RUN/conductor/pre.json"; then
      pid=$(kill_target "$sid" "$RUN/conductor/pre.json" 2>/dev/null || true)
      [ -n "$pid" ] && break
    fi
    sleep 5
  done
  now=$(date +%s)
  if [ -z "$pid" ]; then
    check fail "8 the conductor's process id passed the guards"
    kill_target "$sid" "$RUN/conductor/pre.json" || true
    return 0
  fi
  printf '%s\t%s\t%s\n' "$sid" "$pid" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" >"$RUN/conductor/target"
  if [ $((now - $(epoch "$launched"))) -le 60 ]; then
    check pass "8 the conductor's pid was printed within its first minute ($((now - $(epoch "$launched"))) s)"
  else
    check fail "8 the conductor's pid was printed within its first minute ($((now - $(epoch "$launched"))) s)"
  fi
  say "conductor: the pid is $pid. Run this alone, then the --after step:"
  say "  kill $pid"
  say "  bash scripts/live-recover.sh conductor --after"
}

phase_conductor_after() {
  need_stamp conductor
  if [ ! -s "$RUN/conductor/target" ]; then
    say "conductor --after: no recorded target; run the conductor phase first"
    exit 1
  fi
  local sid pid killed started i first="" mech="" stops=0
  IFS=$'\t' read -r sid pid killed <"$RUN/conductor/target"
  started=$(date +%s)
  say "conductor: the kill of pid $pid (session $sid) was run before $(date -u +%H:%M:%SZ); old pid $( [ -d "/proc/$pid" ] && echo 'still present' || echo gone )"

  local old_pid_gone=no
  for ((i = 1; i <= 10; i++)); do
    local f=$RUN/conductor/min-$i.json src live row rsid rpid
    if save_agents "$f"; then
      row=$(jq -r --arg n "$CONDUCTOR" "[.[]? | select(.name == \$n and ($LIVE_JQ))][0] // empty
        | \"\\(.sessionId // \"none\") \\(.pid // \"none\") \\(.state // \"none\")\"" "$f")
      if [ -n "$row" ]; then
        read -r rsid rpid live <<<"$row"
        # A row the supervisor is still bringing back lists no pid yet; it counts once it has one.
        if [ -z "$first" ] && { [ "$rsid" != "$sid" ] || { [ "$rpid" != none ] && [ "$rpid" != "$pid" ]; }; }; then
          if [ "$rsid" = "$sid" ]; then
            mech="the supervisor: the same session id under a new pid ($pid -> $rpid)"
          else
            mech="the check: a new session id ($sid -> $rsid)"
          fi
          first="minute $i, $(( $(date +%s) - started )) s after the step began"
        fi
        say "conductor read $i: live row $rsid pid $rpid $live"
      else
        say "conductor read $i: no live $CONDUCTOR row"
      fi
    else
      say "conductor read $i: the list could not be read"
    fi
    [ -d "/proc/$pid" ] || old_pid_gone=yes

    local src stopped
    src=$(page_json "$RUN/conductor/page-$i.json" || true)
    stopped=$(jq -r --arg n "$CONDUCTOR" '([.snapshot.sessions[]? | select(.name == $n and .stop_reason != null)] | length)
      + ([.snapshot.needs_you_now[]? | select(.session == $n)] | length)' "$RUN/conductor/page-$i.json" 2>/dev/null || echo "?")
    say "conductor read $i: the $src shows $stopped stop(s) for $CONDUCTOR"
    [ "$stopped" = 0 ] || stops=$((stops + 1))
    [ "$i" -lt 10 ] && sleep 60
  done

  if [ -n "$first" ]; then
    check pass "8 the conductor is running again within 10 minutes: $mech, seen at $first"
  else
    check fail "8 the conductor is running again within 10 minutes (no live row after the kill)"
  fi
  say "conductor: old pid gone: $old_pid_gone"
  say "the check log's conductor lines since the kill:"
  log_lines conductor "$(epoch "$killed")"
  if [ "$stops" = 0 ]; then
    check pass "8 the page shows no stop for the conductor in 10 reads"
  else
    check fail "8 the page shows no stop for the conductor ($stops of 10 reads)"
  fi
}

phase_conductor() {
  if [ "${1:-}" = --after ]; then
    phase_conductor_after
  else
    phase_conductor_before
  fi
}

# ---- ask -------------------------------------------------------------------

# ask_line_ok succeeds when the page reads Needs you now and an entry for the ask round names
# the command.
ask_line_ok() {
  page_json "$RUN/page-ask.json" >/dev/null || return 1
  jq -e --arg r "$ASK" '.snapshot.next_session.state == "needs_you_now"
    and ([.snapshot.needs_you_now[]? | select(.round == $r and (((.cause // "") + " " + (.todo // "")) | contains("cksum ask-probe.txt")))] | length > 0)' \
    "$RUN/page-ask.json" >/dev/null 2>&1
}

# end_run_pass stops a pass still building or waiting, so the check does not relaunch it, and
# ends its open session through the binary. Neither call is fatal.
end_run_pass() {
  local p=$1 state open
  state=$(pass_field "$p" '.state')
  case $state in
    building | waiting) try cc pass stop "$PROJECT" "$p" --reason scope ;;
    *) say "$p reads $state" ;;
  esac
  open=$(pass_field "$p" 'select(.session != null and .session.ended == null) | .session.session_id // empty')
  if [ -n "$open" ]; then
    teardown_end pass "$p"
  fi
}

phase_ask() {
  need_stamp ask
  # The release already holds the ask round; the two-pass limit starts its pass after the merge.
  if ! wait_for 2400 have_session "$ASK"; then
    check fail "8 the check started the ask pass"
    end_run_pass "$ASK"
    return 0
  fi
  check pass "8 the check started the ask pass"
  local launched deadline now
  launched=$(pass_field "$ASK" '.session.launched')
  deadline=$(($(epoch "$launched") + 600))
  say "ask: session $(pass_field "$ASK" '.session.session_id') launched $launched; the page must name the command by $(date -u -d "@$deadline" +%H:%M:%SZ)"
  while :; do
    if ask_line_ok; then
      # The page reads the state on each request, so it has read Needs you now since the check
      # recorded the prompt. The prompt's time, not this read's, is the evidence.
      local asked
      asked=$(pass_field "$ASK" '[.session.prompts[]? | select((.command // "") | contains("cksum ask-probe.txt")) | .time][0] // empty')
      if [ -n "$asked" ] && [ "$(epoch "$asked")" -le "$deadline" ]; then
        check pass "8 the page reads Needs you now and names cksum ask-probe.txt; the check recorded the prompt at $asked, $(($(epoch "$asked") - $(epoch "$launched"))) s after the launch"
      else
        check fail "8 the page names cksum ask-probe.txt, but the prompt on the record (${asked:-none}) is not within 10 minutes of the launch"
      fi
      break
    fi
    now=$(date +%s)
    if [ "$now" -gt "$deadline" ]; then
      check fail "8 the page reads Needs you now and names cksum ask-probe.txt within 10 minutes of the launch"
      break
    fi
    sleep 10
  done
  say "the ask pass's prompts on its record:"
  pass_field "$ASK" '.session.prompts // []' || true
  say "the needs entries on the page:"
  jq -c '.snapshot.needs_you_now' "$RUN/page-ask.json" 2>/dev/null || true
  # The stuck session ends through session end (standing answer 18).
  end_run_pass "$ASK"
}

# ---- pass ------------------------------------------------------------------

# track_ids adds the work pass's current session id and short id to the run's id file. A row
# counts as the pass's when its name, short id, or session id is on this file, or its cwd is
# under the pass's worktree.
track_ids() {
  pass_field "$WORK" '.session | select(. != null) | .session_id, .short_id' >>"$RUN/work-ids"
  sort -u "$RUN/work-ids" -o "$RUN/work-ids"
}

work_seen() {
  local l s
  l=$(pass_field "$WORK" '.session.launched // empty')
  s=$(pass_field "$WORK" '.session.last_seen // empty')
  [ -n "$l" ] && [ -n "$s" ] && [ "$(epoch "$s")" -gt "$(epoch "$l")" ]
}

counted_relaunches() {
  pass_field "$WORK" '[.session.relaunches[]? | select(.counted)] | length'
}

# relaunched_after succeeds when the pass holds $1 counted relaunches and a session launched
# at or after the epoch $2.
relaunched_after() {
  local l
  [ "$(counted_relaunches)" = "$1" ] || return 1
  l=$(pass_field "$WORK" '.session.launched // empty')
  [ -n "$l" ] && [ "$(epoch "$l")" -ge "$2" ]
}

stopped_at_limit() {
  [ "$(pass_field "$WORK" '.state')" = stopped ] && [ "$(pass_field "$WORK" '.stop_reason // empty')" = relaunch_limit ]
}

# end_session ends an open session through the binary: `end_session pass <name>` or
# `end_session conductor`. `session end` waits only 10 seconds for check.lock and a check run
# holds it for about 40, so a refusal waits 60 seconds and retries once (standing answer 72).
end_session() {
  printf '$ cc-desk session end %s %s\n' "$1" "$PROJECT${2:+ $2}"
  if cc session end "$1" "$PROJECT" ${2:+"$2"}; then
    return 0
  fi
  say "session end was refused; waiting 60 seconds and trying once more"
  sleep 60
  cc session end "$1" "$PROJECT" ${2:+"$2"}
}

# teardown_end is end_session for close: a second refusal is a fail line, not a stop of the phase.
# A check run may end the session between the read and the verb; "has no open session" then
# means it is already ended, which is what the teardown wants.
teardown_end() {
  local out
  if out=$(end_session "$@" 2>&1); then
    say "$out"
    return 0
  fi
  say "$out"
  if grep -q 'has no open session' <<<"$out"; then
    say "close: a check run already ended the session"
  else
    check fail "close: session end $* was refused twice; this is a stop (standing answer 72): the conductor ends the session"
  fi
}

# page_has_limit succeeds when the page reads Needs you now and holds an entry for the work
# round with the relaunch limit.
page_has_limit() {
  local src
  src=$(page_json "$RUN/page-limit.json") || return 1
  jq -e --arg r "$WORK" '.snapshot.next_session.state == "needs_you_now"
    and ([.snapshot.needs_you_now[]? | select(.round == $r and .reason == "relaunch_limit")] | length > 0)' "$RUN/page-limit.json" >/dev/null 2>&1
}

# death draws one pass death in turn: it waits for a check to see the session since its launch,
# ends the session, and waits for the next check's decision.
death() {
  local n=$1 sid launched seen dt
  if ! wait_for 1800 work_seen; then
    check fail "8 death $n: a check saw the session after its launch (last_seen later than launched)"
    return 1
  fi
  track_ids
  sid=$(pass_field "$WORK" '.session.session_id')
  launched=$(pass_field "$WORK" '.session.launched')
  seen=$(pass_field "$WORK" '.session.last_seen')
  check pass "8 death $n: a check saw session $sid ($(pass_field "$WORK" '.session.last_state // "?"')) at $seen, after its launch at $launched"
  dt=$(date +%s)
  printf '%s\t%s\t%s\n' "$n" "$sid" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" >>"$RUN/deaths.tsv"
  if ! end_session pass "$WORK"; then
    say "death $n: session end was refused twice. This is a stop (standing answer 72): the conductor ends the session."
    check fail "8 death $n: session end ran"
    return 1
  fi
  say "death $n: session $sid ended through session end at $(date -u +%H:%M:%SZ)"
  if [ "$n" -lt 3 ]; then
    if wait_for 1200 relaunched_after "$n" "$dt"; then
      track_ids
      say "the check log's lines for pass:$WORK since the death:"
      log_lines "pass:$WORK" "$dt"
      if log_lines "pass:$WORK" "$dt" | grep -Eq '  relaunch  missing  '; then
        check pass "8 death $n: the next check gave a counted relaunch, reason missing (count $n, new session $(pass_field "$WORK" '.session.session_id'))"
      else
        check fail "8 death $n: the check log holds no relaunch with reason missing"
      fi
    else
      check fail "8 death $n: the next check relaunched the pass with $n counted relaunch(es)"
      return 1
    fi
  else
    if wait_for 1200 stopped_at_limit; then
      say "the check log's lines for pass:$WORK since the death:"
      log_lines "pass:$WORK" "$dt"
      if log_lines "pass:$WORK" "$dt" | grep -Eq '  stop  relaunch_limit  '; then
        check pass "8 death 3: the next check gave stop, reason relaunch_limit, after 2 counted relaunches"
      else
        check fail "8 death 3: the check log holds no stop with reason relaunch_limit"
      fi
    else
      check fail "8 death 3: the pass reads stopped with relaunch_limit"
      return 1
    fi
    if wait_for 300 page_has_limit; then
      check pass "8 the page reads Needs you now with the round $WORK and the relaunch limit"
    else
      check fail "8 the page reads Needs you now with the round $WORK and the relaunch limit"
    fi
    say "the needs entries on the page:"
    jq -c '.snapshot.needs_you_now' "$RUN/page-limit.json" 2>/dev/null || true
  fi
}

phase_pass() {
  need_stamp pass
  : >>"$RUN/work-ids"
  say "pass: waiting for the check to start $WORK once the merge pass is done"
  if ! wait_for 2400 have_session "$WORK"; then
    check fail "8 the check started the work pass"
    return 0
  fi
  check pass "8 the check started the work pass"
  track_ids
  local n
  for n in 1 2 3; do
    death "$n" || return 0
  done
}

# ---- evidence --------------------------------------------------------------

# pass_rows counts the live rows of the list file $1 that belong to the work pass: by name,
# by a short id or session id on the pass's records, or by a cwd under its worktree.
pass_rows() {
  jq --arg n "$PROJECT-$WORK" --arg wt "$SCRATCH/.claude/worktrees/$WORK" --rawfile ids "$RUN/work-ids" \
    "(\$ids | split(\"\\n\") | map(select(. != \"\"))) as \$k
    | [.[]? | select(($LIVE_JQ)
      and ((.name // \"\") == \$n
        or ((.id // \"\") as \$i | \$k | any(. == \$i))
        or ((.sessionId // \"\") as \$s | \$k | any(. == \$s))
        or ((.cwd // \"\") | startswith(\$wt))))] | length" "$1"
}

check_two_rows() {
  local f n bad=0 reads=0 max=0
  for f in "$RUN"/lists/*.json; do
    [ -e "$f" ] || continue
    n=$(pass_rows "$f" 2>/dev/null || echo 0)
    reads=$((reads + 1))
    [ "$n" -gt "$max" ] && max=$n
    if [ "$n" -gt 1 ]; then
      bad=$((bad + 1))
      say "two live rows: $(basename "$f") holds $n rows for $WORK"
    fi
  done
  say "lists read: $reads; most live rows for $WORK in one list: $max"
  if [ "$reads" -eq 0 ]; then
    check fail "8 the saved lists never show two live rows for the pass (no list was saved)"
  elif [ "$bad" -eq 0 ]; then
    check pass "8 the saved lists never show two live rows for the pass ($reads lists)"
  else
    check fail "8 the saved lists never show two live rows for the pass ($bad of $reads lists)"
  fi
}

# restore_line prints the warning that stands where the line putting main's build back was.
# Once a run writes a report that carries a release name, main's v0.1.0 build cannot load the
# state folder, so reinstalling it before the pass lands would break the timer and the page.
restore_line() {
  say "warning: main's v0.1.0 build can't load the state folder once a report carries a release name."
  say "do not put main's build back on the timer; the conductor reinstalls main after the pass lands."
}

phase_evidence() {
  need_stamp evidence
  check_two_rows
  say
  say "the deaths (number, the session ended, when):"
  cat "$RUN/deaths.tsv" 2>/dev/null || say "(none)"
  say
  say "the work pass:"
  pass_field "$WORK" '{state, stop_reason, started, session: (.session | select(. != null) | {name, short_id, launched, last_seen, last_state, relaunches})}' || say "(no record)"
  say
  say "the check log's lines for pass:$WORK:"
  log_lines "pass:$WORK" 0
  say "the check log's lines for the conductor:"
  log_lines conductor 0
  say "the check log's lines for pass:$ASK:"
  log_lines "pass:$ASK" 0
  say
  say "the checks so far:"
  cat "$RESULTS" 2>/dev/null || true
}

# ---- close -----------------------------------------------------------------

mark_reports_read() {
  local id
  for id in $(jq -r --arg p "$PROJECT" '.reports[]? | select(.project == $p and .read == null) | .id' "$STATE/reports.json" 2>/dev/null); do
    try curl -fsS -m 10 -o /dev/null -X POST -H "Origin: $PAGE" --data '' "$PAGE/reports/$id?/read"
    if jq -e --arg i "$id" '.reports[] | select(.id == $i and .read == null)' "$STATE/reports.json" >/dev/null 2>&1; then
      say "report $id is still unread after the page's request; marking it through the binary"
      try cc report read "$id"
    fi
  done
}

snapshot_line() {
  local snap=$RUN/final-snapshot.json
  if cc snapshot >"$snap" 2>/dev/null; then
    jq -r '"snapshot: \(.needs_you_now | length) needs, \(.waits_for_geoff | length) waits, \(.waiting | length) waiting, \(.sessions | length) sessions, \([.reports[]? | select(.read == null)] | length) unread"' "$snap"
  else
    say "snapshot: unreadable"
  fi
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

# Close ends every session the run's records hold through session end, stops the list saver,
# marks the reports read, ends the release, and puts the scratch checkout on main. It runs the
# binary only; no kill.
phase_close() {
  need_stamp close
  [ -x "$BIN" ] || show go build -o "$BIN" "$REPO/cmd/cc-desk"
  local p live_name="" mine=no
  [ -f "$RELEASE" ] && live_name=$(jq -r '.name // ""' "$RELEASE" 2>/dev/null || true)
  [ "$live_name" = "$PROJECT-recover-$STAMP" ] && mine=yes
  if [ -f "$RELEASE" ] && [ "$mine" = no ]; then
    say "close: the live release is $live_name, not this run's ($PROJECT-recover-$STAMP); it is left alone"
    check fail "close: the live release is this run's"
  elif [ -f "$RELEASE" ]; then
    for p in "$MERGE" "$WORK" "$ASK"; do
      pass_field "$p" '.name' | grep -q . && end_run_pass "$p"
    done
    if conductor_open; then
      try cc session finish conductor "$PROJECT"
      teardown_end conductor
    fi
  else
    say "close: no $PROJECT release in the state folder"
  fi
  if wait_for 420 no_live_rows; then
    say "close: no live scratch row remains"
  else
    say "close: still listed live after the wait:"
    scratch_live_rows "$RUN/now.json"
  fi

  touch "$RUN/saver.stop"
  if [ -f "$RUN/saver.pid" ]; then
    wait_for 60 bash -c "! kill -0 $(cat "$RUN/saver.pid") 2>/dev/null" || say "close: the saver is still running; it ends on its own at its next read"
  fi

  if [ "$mine" = yes ] || [ ! -f "$RELEASE" ]; then
    mark_reports_read
  fi
  if [ "$mine" = yes ]; then
    try cc release 'done' "$PROJECT"
    if [ -f "$RELEASE" ]; then
      say "release done was refused; waiting 60 seconds and trying once more"
      sleep 60
      try cc release 'done' "$PROJECT"
    fi
  fi
  if [ "$mine" = no ] && [ -f "$RELEASE" ]; then
    : # the fail line above names the release found
  elif [ -f "$RELEASE" ]; then
    check fail "close: the release is ended"
  else
    check pass "close: the release is ended"
  fi

  if [ -z "$(git -C "$SCRATCH" status --porcelain)" ]; then
    try git -C "$SCRATCH" switch main
  else
    say "close: the scratch checkout is dirty; it stays on $(git -C "$SCRATCH" branch --show-current)"
  fi
  if [ "$(git -C "$SCRATCH" branch --show-current)" = main ]; then
    check pass "close: the scratch checkout is on main"
  else
    check fail "close: the scratch checkout is on main"
  fi
  snapshot_line
  restore_line
  result_line
}

# ---- main ------------------------------------------------------------------

usage() {
  cat <<'EOF'
usage: bash scripts/live-recover.sh <phase>
  prep            refuse on a dirty scratch, a live release, or a live scratch session; build and install the
                  branch's binary; cut the base branch; add the merge, work, and ask rounds; start the 30-second saver
  conductor       wait for the check to start the merge pass, report its pull request ready, wait for the merge
                  and the conductor's launch, print the conductor's pid, and stop (the executor runs the kill)
  conductor --after
                  read the list and the page each minute for 10 minutes after the kill; name which brought it back
  ask             wait for the ask pass the check starts once the merge frees a slot, and check the page names
                  cksum ask-probe.txt within 10 minutes of its launch; then stop the pass and end its session
  pass            three deaths of the work pass through session end; each waits for a check to see the session;
                  the third gives relaunch_limit and Needs you now with the round's name
  evidence        check the saved lists never show two live rows for the pass; print each death's evidence
  close           end every session the records hold, stop the saver, mark reports read, release done,
                  switch the scratch checkout to main, and print the snapshot line and the result
  help            print this list
Run them in this order, one at a time. No phase starts, stops, or restarts a unit; each waits on the timed check.
EOF
}

main() {
  local phase=${1:-help}
  case $phase in
    help | -h | --help)
      usage
      return 0
      ;;
    prep | conductor | ask | pass | evidence | close | saver-loop) ;;
    *)
      usage >&2
      exit 2
      ;;
  esac
  mkdir -p "$RUN"
  if [ "$phase" = saver-loop ]; then
    phase_saver_loop
    return 0
  fi
  exec > >(tee -a "$RUN/run.log") 2>&1
  say "== $phase ${2:-} $(date -u +%Y-%m-%dT%H:%M:%SZ) =="
  "phase_${phase//-/_}" "${@:2}"
}

main "$@"
