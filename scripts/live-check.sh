#!/usr/bin/env bash
# live-check.sh: the live run of the sessions timer pass, one phase per call.
#
#   bash scripts/live-check.sh <prep|start|watch|wait|scope|abort|evidence|help>
#
# Each phase prints the commands it ran with their output, and appends all of
# it to $CLAUDE_JOB_DIR/tmp/live/run.log. It reads CC_DESK_STATE_DIR (default
# ~/.local/state/cc-desk) and SCRATCH_CHECKOUT (default ~/Projects/cc-desk-scratch).
# The plan is docs/plans/2026-11-17-sessions-timer-plan.md, "The live script".
set -euo pipefail

REPO=$(cd "$(dirname "$0")/.." && pwd)
STATE=${CC_DESK_STATE_DIR:-$HOME/.local/state/cc-desk}
SCRATCH=${SCRATCH_CHECKOUT:-$HOME/Projects/cc-desk-scratch}
RUN=${CLAUDE_JOB_DIR:-${TMPDIR:-/tmp}/cc-desk-live}/tmp/live
BIN=$HOME/.local/share/cc-desk/bin/cc-desk
UNIT_DIR=$HOME/.config/systemd/user
PROJECT=cc-desk-scratch
BASE=live-2026-11-18
PASSES=(timer-live-a timer-live-b)
RELEASE=$STATE/releases/$PROJECT.json

# say prints a line; show prints a command, then runs it and its output.
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

# unit_path is the PATH the installed service sets, with %h filled in.
unit_path() {
  grep '^Environment=PATH=' "$UNIT_DIR/cc-desk-check.service" | sed -e 's/^Environment=PATH=//' -e "s|%h|$HOME|g"
}

# pass_field prints one jq expression over a pass's record in the release file.
pass_field() {
  jq -r --arg p "$1" ".passes[] | select(.name == \$p) | $2" "$RELEASE"
}

# transcripts lists the transcript files of a session id, or of a short id, which is a
# prefix of the id the files are named by.
transcripts() {
  local f
  for f in "$HOME"/.claude/projects/*/"$1"*.jsonl; do
    [ -e "$f" ] && printf '%s\n' "$f"
  done
  return 0
}

# session_ids prints every session id of a pass, so no transcript escapes the count: the one
# watch saved, every session_id the pass's record holds, and each short id the check log
# names on the pass's lines (a relaunch or a hand resume leaves ids the record no longer holds).
session_ids() {
  {
    [ -f "$RUN/$1.first-session-id" ] && cat "$RUN/$1.first-session-id"
    jq -r --arg p "$1" '.passes[] | select(.name == $p) | [.. | objects | .session_id? // empty] | .[]' "$RELEASE"
    if [ -f "$STATE/check-log.jsonl" ]; then
      jq -r --arg p "$1" 'select((.record // "") | endswith($p)) | "\(.outcome // "") \(.detail // "")"' \
        "$STATE/check-log.jsonl" 2>/dev/null | grep -oE '\b[0-9a-f]{8}\b' || true
    fi
  } | sort -u
}

# merge_calls counts the Bash tool calls of a pass's transcripts whose command
# holds "gh pr merge". Reading the tool_use input keeps text in a tool's output
# from counting. A transcript reached by both a short and a full id is read once.
merge_calls() {
  local id f n total=0 files
  files=$(for id in $(session_ids "$1"); do transcripts "$id"; done | sort -u)
  for f in $files; do
    n=$(jq -r '(.message.content? // []) | if type == "array" then .[] else empty end
      | select(.type == "tool_use" and .name == "Bash") | .input.command // empty
      | select(contains("gh pr merge"))' "$f" | wc -l)
    say "  transcript $f: $n merge call(s)"
    total=$((total + n))
  done
  say "  $1: $total Bash call(s) holding gh pr merge"
}

# ---- prep ------------------------------------------------------------------

prep_refuse() {
  if [ ! -d "$SCRATCH/.git" ]; then
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

# prep_failed runs when prep exits before its last line. A failure after `timer install`, such
# as the gh probe or the push, would otherwise leave the timer enabled and a half-cut base
# branch, so it prints the state and puts the timer back as abort does.
prep_failed() {
  local rc=$?
  [ "$rc" -eq 0 ] && return 0
  say "prep failed (exit $rc); the state it left:"
  try git -C "$SCRATCH" status --short --branch
  try git -C "$SCRATCH" branch --list "$BASE"
  try cc timer status
  try systemctl --user disable --now cc-desk-check.timer
  say "prep: the timer is disabled; the scratch branch $BASE, if listed above, is half cut"
}

phase_prep() {
  prep_refuse
  mkdir -p "$RUN"
  trap prep_failed EXIT
  say "prep: the scratch checkout is clean on main, and the state folder holds no other release"

  show go build -o "$RUN/cc-desk" "$REPO/cmd/cc-desk"
  show "$RUN/cc-desk" timer install
  show systemd-analyze --user verify "$UNIT_DIR/cc-desk-check.service" "$UNIT_DIR/cc-desk-check.timer"

  local path
  path=$(unit_path)
  show systemd-run --user --wait --pipe -p "Environment=PATH=$path" \
    gh pr list --repo glw907/cc-desk-scratch --limit 1

  show git -C "$SCRATCH" fetch origin
  show git -C "$SCRATCH" switch -c "$BASE" origin/main
  mkdir -p "$SCRATCH/docs/plans"
  local p
  for p in "${PASSES[@]}"; do
    show cp "$REPO/scripts/live/$p-plan.md" "$SCRATCH/docs/plans/$p-plan.md"
  done
  for p in "${PASSES[@]}"; do
    show git -C "$SCRATCH" add "docs/plans/$p-plan.md"
  done
  show git -C "$SCRATCH" commit -m "Add the timer-live fixture plans"
  show git -C "$SCRATCH" push origin "$BASE"

  local exclude
  exclude=$(git -C "$SCRATCH" rev-parse --git-path info/exclude)
  case $exclude in /*) ;; *) exclude=$SCRATCH/$exclude ;; esac
  say "\$ echo .claude/worktrees/ >> $exclude"
  grep -qx '.claude/worktrees/' "$exclude" 2>/dev/null || echo '.claude/worktrees/' >>"$exclude"

  show cc release add "$REPO/scripts/live/release.json"
  show cc release checkout "$PROJECT" "$SCRATCH" --default-branch "$BASE"

  say "usage reading:"
  show cc usage --raw "$RUN/usage.raw"
  trap - EXIT
  say "warning: main's v0.1.0 build can't load the state folder once a report carries a release name."
  say "do not put main's build back on the timer; the conductor reinstalls main after the pass lands."
}

# ---- start -----------------------------------------------------------------

phase_start() {
  mkdir -p "$RUN"
  date -u +%Y-%m-%dT%H:%M:%SZ | tee "$RUN/start-time"
  local p i state
  for p in "${PASSES[@]}"; do
    show cc pass start "$PROJECT" "$p"
  done
  # A kick that landed while the first check ran may have merged into it, so
  # wait for the service to be idle and kick once more.
  for ((i = 0; i < 30; i++)); do
    state=$(systemctl --user is-active cc-desk-check.service || true)
    [ "$state" = inactive ] && break
    sleep 10
  done
  say "cc-desk-check.service reads $state"
  show systemctl --user start --no-block cc-desk-check.service
}

# ---- watch -----------------------------------------------------------------

phase_watch() {
  local p=timer-live-b i commits seen
  for ((i = 0; i < 54; i++)); do
    git -C "$SCRATCH" fetch --quiet origin || true
    commits=$(git -C "$SCRATCH" rev-list --count "origin/$BASE..origin/$p" 2>/dev/null || echo 0)
    seen=$(pass_field "$p" '.session.last_seen // empty' 2>/dev/null || true)
    if [ "$commits" -ge 1 ] && [ -n "$seen" ]; then
      break
    fi
    sleep 10
  done
  say "origin/$p holds $commits commit(s) past origin/$BASE; last_seen is '${seen:-unset}'"
  if [ "$commits" -lt 1 ] || [ -z "${seen:-}" ]; then
    say "watch: not both yet; call watch again"
    return 0
  fi
  local sid short
  sid=$(pass_field "$p" '.session.session_id')
  short=$(pass_field "$p" '.session.short_id')
  mkdir -p "$RUN"
  echo "$sid" >"$RUN/$p.first-session-id"
  say "saved session_id $sid to $RUN/$p.first-session-id"
  say "short id: $short"
  say "run alone: claude stop $short"
}

# ---- wait ------------------------------------------------------------------

pass_report() {
  local p=$1
  say "$p: state $(pass_field "$p" '.state'), pr $(pass_field "$p" '.pull_request.number // "none"')," \
    "merged $(pass_field "$p" '.pull_request.merged // "no"'), cleaned $(pass_field "$p" '.pull_request.cleaned // "no"')"
}

pass_done() {
  [ "$(pass_field "$1" '.state')" = "done" ] && [ -n "$(pass_field "$1" '.pull_request.cleaned // empty')" ]
}

phase_wait() {
  local i p all
  for ((i = 0; i < 18; i++)); do
    all=yes
    for p in "${PASSES[@]}"; do
      pass_done "$p" || all=no
    done
    [ "$all" = yes ] && break
    sleep 30
  done
  for p in "${PASSES[@]}"; do
    pass_report "$p"
  done
  if [ "$all" = yes ]; then
    say "wait: both passes are done with cleanup recorded"
  else
    say "wait: not done; call wait again until done or 45 minutes after $(cat "$RUN/start-time" 2>/dev/null || echo start)"
  fi
}

# ---- scope -----------------------------------------------------------------

phase_scope() {
  mkdir -p "$RUN"
  local standin=$RUN/scope-standin.sh path
  cat >"$standin" <<'EOF'
#!/bin/sh
sleep 901 >/dev/null 2>&1 &
echo "backgrounded · probe"
EOF
  chmod +x "$standin"
  path=$(unit_path)
  # The probe's sleep lasts 901 seconds, so a failure of the outer systemd-run must still stop
  # its scope. The stop is allowed to find no scope.
  trap 'systemctl --user stop cc-desk-scope-probe-sleep.scope 2>/dev/null || true' EXIT
  show systemd-run --user --wait --unit=cc-desk-scope-probe -p Type=oneshot -p KillMode=process \
    -p "Environment=PATH=$path" systemd-run --user --scope -q --unit=cc-desk-scope-probe-sleep "$standin"
  try systemctl --user is-active cc-desk-scope-probe.service
  try systemctl --user is-active cc-desk-scope-probe-sleep.scope
  show systemctl --user stop cc-desk-scope-probe-sleep.scope
  trap - EXIT
}

# ---- abort -----------------------------------------------------------------

# wait_service_idle reads the check service's state until it is not running, so no check run
# is writing the folder while the stops are carried out. It reads only: it starts, stops, and
# restarts no unit. A fixture session still running can call `pass ready`, which kicks one
# more check, so the wait has to follow the timer's disable.
wait_service_idle() {
  local i state=unknown
  for ((i = 0; i < 60; i++)); do
    state=$(systemctl --user is-active cc-desk-check.service || true)
    case $state in
      active | activating | deactivating | reloading) ;;
      *) break ;;
    esac
    sleep 10
  done
  say "cc-desk-check.service reads $state"
  case $state in
    active | activating | deactivating | reloading) return 1 ;;
  esac
}

phase_abort() {
  try systemctl --user disable --now cc-desk-check.timer
  wait_service_idle || say "abort: the check service is still running; the stops below may race it"
  local p state open
  for p in "${PASSES[@]}"; do
    state=$(pass_field "$p" '.state' 2>/dev/null || echo unknown)
    if [ "$state" = building ] || [ "$state" = waiting ]; then
      try cc pass stop "$PROJECT" "$p" --reason scope
    else
      say "$p reads $state"
    fi
    # A pass the check marked stopped can still hold a live session, so every pass whose
    # record holds an open session gets the end, not only one that read building or waiting.
    open=$(pass_field "$p" 'select(.session != null and .session.ended == null) | .session.session_id // empty' 2>/dev/null || true)
    if [ -n "$open" ]; then
      try cc session end pass "$PROJECT" "$p"
    fi
  done
}

# ---- evidence --------------------------------------------------------------

# phase_evidence runs the readings with errexit off, so a failed reading never
# keeps the timer enabled; the switch and the disable always run.
phase_evidence() {
  evidence_readings || say "evidence: a reading failed; the rest ran"
  say
  try git -C "$SCRATCH" switch main
  try systemctl --user disable --now cc-desk-check.timer
  try systemctl --user is-enabled cc-desk-check.timer
}

evidence_readings() {
  try cc timer status

  say
  say "item 7: the units verify"
  try systemd-analyze --user verify "$UNIT_DIR/cc-desk-check.service" "$UNIT_DIR/cc-desk-check.timer"
  say "item 7: the scope probe's output is in the scope phase's log above"

  say
  say "item 6: each pass's record"
  local p
  for p in "${PASSES[@]}"; do
    say "$p:"
    pass_field "$p" '{state, started, finished, actual_minutes, pull_request,
      session: (.session | {name, launched, launch_tries, relaunches, prompts, last_seen, finished, ended})}'
  done

  say
  say "item 6: merge by the check, and zero prompts"
  for p in "${PASSES[@]}"; do
    say "$p merge record: merge_called $(pass_field "$p" '.pull_request.merge_called // "none"')," \
      "merged_outside $(pass_field "$p" '.pull_request.merged_outside // false')"
    say "$p prompts recorded: $(pass_field "$p" '(.session.prompts // []) | length')"
  done

  say
  say "item 6: no merge command in the sessions' logs (Bash tool_use calls only)"
  for p in "${PASSES[@]}"; do
    merge_calls "$p"
  done

  say
  say "item 6: the denials log"
  local denials=$HOME/.local/state/claude-denials.jsonl started
  started=$(cat "$RUN/start-time" 2>/dev/null || echo 9999)
  if [ -f "$denials" ]; then
    say "  rows under cc-desk-scratch/.claude/worktrees/ after $started: $(jq -R --arg s "$started" \
      'fromjson? | select((.cwd // "" | contains("cc-desk-scratch/.claude/worktrees/")) and .ts > $s)' "$denials" | jq -s length)"
  else
    say "  $denials is absent, so no row can be counted"
  fi

  say
  say "item 6: relaunch of timer-live-b"
  pass_field timer-live-b '.session.relaunches // []'

  say
  say "item 6: gone, one command per item"
  claude agents --json --all >"$RUN/agents.json" || true
  local wt name
  for p in "${PASSES[@]}"; do
    name=$PROJECT-$p
    say "$p: rows named $name in claude agents: $(jq --arg n "$name" '[.. | objects | select(.name? == $n)] | length' "$RUN/agents.json")"
    wt=$(pass_field "$p" '.session.worktree // empty')
    if [ -n "$wt" ]; then
      try test ! -e "$wt"
    else
      say "$p: the record holds no worktree path"
    fi
    try git -C "$SCRATCH" show-ref --verify "refs/heads/$p"
  done
  say "(the worktree and branch are gone when the first two read no rows and exit 0, and show-ref exits 128)"

}

# ---- main ------------------------------------------------------------------

usage() {
  cat <<'EOF'
usage: bash scripts/live-check.sh <phase>
  prep      refuse on a dirty scratch or another release; install the timer; cut the base branch; add the release
  start     write the start time, start both passes, kick the check once more when it is idle
  watch     wait for the second pass's first commit and last_seen; print the claude stop line
  wait      poll until both passes are done with cleanup recorded
  scope     probe that a scope outlives its service
  abort     disable the timer; print the stop lines for a pass still building
  evidence  print each done item's evidence, switch scratch to main, and disable the timer
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
    prep | start | watch | wait | scope | abort | evidence) ;;
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
