#!/usr/bin/env bash
# setup-release-check.sh: item 10's check, run on a copy of the state folder.
#
#   bash scripts/setup-release-check.sh [rehearsal-pass]
#
# The rehearsal's pass name is the argument, or REHEARSAL_PASS, or `run`. The script reads
# the real folder only to copy it: it waits, reading only, until cc-desk-check.service is
# inactive so no check run is mid-write, copies the folder, and does everything else to the
# copy. It never starts, stops, or restarts a unit. Only `check --dry-run` runs a check on
# the copy, since the copy's checkout is the real ~/Projects/cc-desk.
set -euo pipefail

REPO=$(cd "$(dirname "$0")/.." && pwd)
REAL=${CC_DESK_STATE_DIR:-$HOME/.local/state/cc-desk}
BIN=${CC_DESK_BINARY:-cc-desk}
RULES=$REPO/scripts/live/run/allow-rules.json
PASS=${1:-${REHEARSAL_PASS:-run}}
UNIT=cc-desk-check.service
WAIT_SECONDS=${SETUP_CHECK_WAIT:-1800}

say() { printf '%s\n' "$*"; }
RESULTS=()
check() {
  local verdict=$1 name=$2
  say "$verdict $name"
  RESULTS+=("$verdict")
}

# fill puts a repo and a branch into the shared rules.
fill() {
  jq --arg repo "$1" --arg branch "$2" \
    'map(gsub("\\{\\{repo\\}\\}"; $repo) | gsub("\\{\\{branch\\}\\}"; $branch))' "$RULES"
}
# unfill puts the two placeholders back in a filled list, given the repo and branch it holds.
unfill() {
  jq --arg repo "$1" --arg branch "$2" \
    'map(split($repo) | join("{{repo}}") | split($branch) | join("{{branch}}"))'
}

waited=0
while [ "$(systemctl --user is-active "$UNIT" 2>/dev/null || true)" != inactive ]; do
  if [ "$waited" -ge "$WAIT_SECONDS" ]; then
    say "refused: $UNIT stayed active for $WAIT_SECONDS seconds"
    exit 1
  fi
  [ "$waited" -eq 0 ] && say "waiting for $UNIT to go inactive"
  sleep 5
  waited=$((waited + 5))
done

base=${CLAUDE_JOB_DIR:+$CLAUDE_JOB_DIR/tmp}
if [ -n "$base" ]; then
  mkdir -p "$base"
  COPY=$(mktemp -d "$base/setup-release-check.XXXXXX")
else
  COPY=$(mktemp -d)
fi
rmdir "$COPY"
cp -a "$REAL" "$COPY"
say "copied $REAL to $COPY"

cc() { "$BIN" --state-dir "$COPY" "$@"; }

if [ -e "$COPY/releases/cc-desk-scratch.json" ]; then
  cc release "done" cc-desk-scratch
fi
if [ -f "$COPY/reports.json" ]; then
  while IFS= read -r id; do
    [ -n "$id" ] && cc report read "$id"
  done < <(jq -r '.reports[]? | select(.read == null) | .id' "$COPY/reports.json")
fi

CC_DESK_BINARY=$BIN bash "$REPO/scripts/setup-release.sh" --state-dir "$COPY"

waits=$(cc snapshot | jq -c '.waits_for_geoff')
if [ "$(jq -c 'length' <<<"$waits")" = 1 ] &&
  jq -e '.[0].project == "cc-desk" and .[0].questions == 1' <<<"$waits" >/dev/null; then
  check ok "one entry waits for cc-desk, with one question"
else
  check fail "one entry waits for cc-desk, with one question: $waits"
fi

if jq -e '.questions | length > 0 and all(.[]; .never_guess == true)' "$COPY/releases/cc-desk.json" >/dev/null; then
  check ok "never_guess is true in the copy's releases/cc-desk.json"
else
  check fail "never_guess is not true in the copy's releases/cc-desk.json"
fi

want=$(fill glw907/cc-desk stuck-sessions)
if [ "$(jq -S -c . <<<"$want")" = "$(jq -S -c '.allow_rules' "$COPY/releases/cc-desk.json")" ]; then
  check ok "the release's allow rules equal the shared source filled with glw907/cc-desk and stuck-sessions"
else
  check fail "the release's allow rules differ from the shared source filled with glw907/cc-desk and stuck-sessions"
fi

scratch_back=$(fill glw907/cc-desk-scratch "$PASS" | unfill glw907/cc-desk-scratch "$PASS")
real_back=$(fill glw907/cc-desk stuck-sessions | unfill glw907/cc-desk stuck-sessions)
if diff <(jq -S . <<<"$scratch_back") <(jq -S . <<<"$real_back") >/dev/null; then
  check ok "the rehearsal's rules (glw907/cc-desk-scratch, $PASS) and the release's, values put back, match"
else
  check fail "the rehearsal's rules and the release's differ once the values are put back"
  diff <(jq -S . <<<"$scratch_back") <(jq -S . <<<"$real_back") || true
fi

dry=$(cc check --dry-run 2>&1) && dry_rc=0 || dry_rc=$?
if [ "$dry_rc" -eq 0 ] && [ -z "$dry" ]; then
  check ok "check --dry-run on the copy decides nothing"
else
  check fail "check --dry-run on the copy decides something (exit $dry_rc): $dry"
fi

say
if [[ " ${RESULTS[*]} " == *" fail "* ]]; then say "result: fail"; else say "result: pass"; fi
