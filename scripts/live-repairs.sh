#!/usr/bin/env bash
# live-repairs.sh: the live run of the desk-check pass, one phase per call.
#
#   bash scripts/live-repairs.sh <prep|refuse|help>
#
# Each phase prints the commands it ran with their output, and appends all of it to
# $CLAUDE_JOB_DIR/tmp/live-repairs/run.log. Each live check prints `pass <check>` or
# `fail <check>`. It reads CC_DESK_STATE_DIR (default ~/.local/state/cc-desk) and
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

# live_rows prints the id, name, and state of each scratch session the list shows as
# working or blocked. The list goes to a file first, never through a pipe.
live_rows() {
  claude agents --json --all >"$RUN/agents.json" 2>/dev/null || true
  jq -r '.[]? | select((.name // "" | startswith("cc-desk-scratch-")) and (.state == "working" or .state == "blocked"))
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
  show cp "$REPO/scripts/live/desk/hold-plan.md" "$SCRATCH/docs/plans/hold-$STAMP-plan.md"
  show git -C "$SCRATCH" add "docs/plans/refuse-$STAMP-plan.md" "docs/plans/hold-$STAMP-plan.md"
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
export -f add_round add_round_to_release
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
  say "session id: $sid"
  say "the session's deny list: $(jq -c '.permissions.deny' "$STATE/launch/$PROJECT-$pass.json")"
  say "the session's hooks: $(jq -c '.hooks' "$STATE/launch/$PROJECT-$pass.json")"

  local i found=""
  for ((i = 0; i < 120; i++)); do
    found=$(hook_input_for "$sid")
    [ -n "$found" ] && break
    sleep 5
  done
  if [ -n "$found" ]; then
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

# ---- main ------------------------------------------------------------------

usage() {
  cat <<'EOF'
usage: bash scripts/live-repairs.sh <phase>
  prep    refuse on a dirty scratch, another release, or a live scratch session; build the binary; join or add the release
  refuse  run the probe: a pass whose session runs a command auto mode refuses; keep the hook input
  help    print this list
EOF
}

main() {
  local phase=${1:-help}
  case $phase in
    help | -h | --help)
      usage
      return 0
      ;;
    prep | refuse) ;;
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
