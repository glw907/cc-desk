#!/usr/bin/env bash
# gate.sh: the cc-desk gate. Runs gofmt, go vet, go test, and vale over docs/,
# then the skills step (each skills/*/SKILL.md under 160 lines, clean under vale
# and under tellgrader's agent register), then the page's steps when web/package.json is tracked: npm ci, type check,
# build, and the Playwright suite. It fails when any Go test reports a skip, and
# when the Playwright run skips a test, finds no browser, or runs fewer tests than
# the suite lists. The page mock-ups under docs/specs/2026-11-17-cc-desk-mockups/
# are left out of the vale run.
set -euo pipefail
cd "$(dirname "$0")/.."

echo "gate: gofmt"
unformatted=$(git ls-files "*.go" | xargs -r gofmt -l)
if [ -n "$unformatted" ]; then
  echo "gofmt: these files need formatting:" >&2
  echo "$unformatted" >&2
  exit 1
fi

echo "gate: go vet"
go vet ./...

echo "gate: go test"
if ! command -v jq >/dev/null 2>&1; then
  echo "gate: jq is not installed; install it, do not skip the gate" >&2
  exit 1
fi
events=$(mktemp)
trap 'rm -f "$events"' EXIT
# -count=1 because the e2e package builds the binary itself, and go test
# does not see the binary's source as an input of the e2e package, so a
# cached e2e result can pass on a stale build.
status=0
go test -count=1 -json ./... >"$events" || status=$?
jq -rj 'select(.Action == "build-output") | .Output' "$events"
jq -r 'select(.Action == "pass" and .Test == null and .Package != null) | "ok  \t\(.Package)\t\(.Elapsed)s"' "$events"
failed=$(jq -sc '[.[] | select(.Action == "fail" and .Package != null) | .Package] | unique' "$events")
jq -rj --argjson failed "$failed" \
  'select(.Action == "output" and (.Package as $p | $failed | index($p))) | .Output' "$events"
if [ "$status" -ne 0 ]; then
  echo "gate: go test failed (exit $status)" >&2
  exit 1
fi
# A package with no test files reports a skip with no Test key; only a
# skipped test counts.
skipped=$(jq -r 'select(.Action == "skip" and .Test != null) | "\(.Package) \(.Test)"' "$events")
if [ -n "$skipped" ]; then
  echo "gate: these tests were skipped; no test may skip:" >&2
  echo "$skipped" >&2
  exit 1
fi

echo "gate: vale"
if ! command -v vale >/dev/null 2>&1; then
  echo "gate: vale is not installed; install it, do not skip the gate" >&2
  exit 1
fi
mapfile -t docs < <(git ls-files 'docs/*.md' | grep -v '^docs/specs/2026-11-17-cc-desk-mockups/')
vale --minAlertLevel=error --output=line "${docs[@]}"

echo "gate: skills"
if ! command -v tellgrader >/dev/null 2>&1; then
  echo "gate: tellgrader is not installed; install it, do not skip the gate" >&2
  exit 1
fi
mapfile -t skills < <(git ls-files 'skills/*/SKILL.md')
if [ "${#skills[@]}" -eq 0 ]; then
  echo "gate: no skills/*/SKILL.md is tracked" >&2
  exit 1
fi
for skill in "${skills[@]}"; do
  lines=$(wc -l <"$skill")
  if [ "$lines" -ge 160 ]; then
    echo "gate: $skill has $lines lines; a skill stays under 160" >&2
    exit 1
  fi
done
vale --minAlertLevel=error --output=line "${skills[@]}"
# tellgrader exits 2 when a finding gates, and gates only under this profile.
for skill in "${skills[@]}"; do
  tg=0
  tellgrader --register agent --profile docs-register "$skill" >/dev/null || tg=$?
  if [ "$tg" -ne 0 ]; then
    tellgrader --register agent --profile docs-register "$skill" >&2 || true
    echo "gate: tellgrader failed on $skill (exit $tg)" >&2
    exit 1
  fi
done

if git ls-files --error-unmatch web/package.json >/dev/null 2>&1; then
  echo "gate: web install"
  (cd web && npm ci)

  echo "gate: web check"
  (cd web && npm run check)

  echo "gate: web build"
  (cd web && npm run build)

  echo "gate: web tests"
  # --reporter=list matters: a plain --list rewrites the JSON report with every
  # test skipped.
  listing=$(cd web && npx playwright test --list --reporter=list)
  count=$(printf '%s\n' "$listing" | sed -n 's/^Total: \([0-9]\{1,\}\) tests\{0,1\} in .*/\1/p')
  (cd web && npx playwright test)
  bash web/scripts/check-playwright-report.sh web/test-results/report.json "$count"
fi

echo "gate: ok"
