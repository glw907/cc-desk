#!/usr/bin/env bash
# setup-release.sh: record the real cc-desk release in a state folder.
#
#   bash scripts/setup-release.sh --state-dir <dir>
#
# It adds the release of project cc-desk with one round, designs that round's one pass on
# the question the design asks Geoff, and records the checkout. It uses the cc-desk on PATH
# unless CC_DESK_BINARY names another. It refuses a folder that already holds the release,
# since a second add would not replace it.
set -euo pipefail

REPO=$(cd "$(dirname "$0")/.." && pwd)
BIN=${CC_DESK_BINARY:-cc-desk}
PROJECT=cc-desk
ROUND=stuck-sessions
GH_REPO=glw907/cc-desk
RULES=$REPO/scripts/live/run/allow-rules.json

STATE=""
while [ $# -gt 0 ]; do
  case $1 in
    --state-dir)
      [ $# -ge 2 ] || {
        echo "setup-release: --state-dir needs a folder" >&2
        exit 2
      }
      STATE=$2
      shift 2
      ;;
    *)
      echo "usage: setup-release.sh --state-dir <dir>" >&2
      exit 2
      ;;
  esac
done
[ -n "$STATE" ] || {
  echo "usage: setup-release.sh --state-dir <dir>" >&2
  exit 2
}

if [ -e "$STATE/releases/$PROJECT.json" ]; then
  echo "setup-release: refused: $STATE already holds a $PROJECT release" >&2
  exit 1
fi

# A background job's own tmp folder keeps parallel jobs off each other's files in /tmp.
if [ -n "${CLAUDE_JOB_DIR:-}" ]; then
  mkdir -p "$CLAUDE_JOB_DIR/tmp"
  tmp=$(mktemp -d "$CLAUDE_JOB_DIR/tmp/setup-release.XXXXXX")
else
  tmp=$(mktemp -d)
fi
trap 'rm -rf "$tmp"' EXIT

jq --arg repo "$GH_REPO" --arg branch "$ROUND" \
  'map(gsub("\\{\\{repo\\}\\}"; $repo) | gsub("\\{\\{branch\\}\\}"; $branch))' "$RULES" >"$tmp/rules.json"

jq -n --arg repo "$GH_REPO" --arg round "$ROUND" --slurpfile rules "$tmp/rules.json" '{
  format: 1,
  project: "cc-desk",
  name: "cc-desk-stuck-sessions",
  repo: $repo,
  rounds: [{
    name: $round,
    summary: "Tell a stopped pass how to start again, and notice a fixed sign-in without being told.",
    needs_geoff: true,
    design: "docs/specs/2026-11-18-stuck-sessions-design.md"
  }],
  allow_rules: $rules[0]
}' >"$tmp/release.json"

# The question's text is the design's section "The question", taken as written.
jq -n --arg round "$ROUND" '{
  format: 1,
  passes: [{
    name: $round,
    summary: "Tell a stopped pass how to start again, and notice a fixed sign-in without being told.",
    round: $round,
    planned_minutes: 60,
    questions: ["ss1"],
    state: "not_started"
  }],
  questions: [{
    id: "ss1",
    text: "Should cc-desk notice by itself that you fixed the problem?",
    options: ["yes", "no"],
    recommended: "yes",
    never_guess: true,
    background: [
      "The check can tell when one of the four causes is gone: `claude auth status` reports whether the sign-in is good, with no model call. For the other three the cause is something Geoff read and judged, and the check has nothing to look at."
    ],
    causes: {
      yes: "After a sign-in stop, the check reads the sign-in every five minutes. Once it is good, the check removes the line and resumes the work. Geoff signs in and walks away. The other three lines name the command, because the check can\u0027t see those fixes.",
      no: "All four lines name the command. The work resumes when that command is run at the desk, so nothing restarts without a deliberate act."
    },
    reason: "Geoff shouldn\u0027t have to tell the system about a fix it can see. A wrong resume costs little, since a session that meets the expired sign-in again stops again."
  }]
}' >"$tmp/design.json"

run() {
  printf '$ %s\n' "$*"
  "$@"
}

run "$BIN" --state-dir "$STATE" release add "$tmp/release.json"
run "$BIN" --state-dir "$STATE" round design "$PROJECT" "$ROUND" "$tmp/design.json"
run "$BIN" --state-dir "$STATE" release checkout "$PROJECT" "$HOME/Projects/cc-desk" --default-branch main
