#!/usr/bin/env bash
# gate.sh: the cc-desk gate. Runs gofmt, go vet, go test, and vale over docs/.
# The page mock-ups under docs/specs/2026-11-17-cc-desk-mockups/ are left out.
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
go test ./...

echo "gate: vale"
if ! command -v vale >/dev/null 2>&1; then
  echo "gate: vale is not installed; install it, do not skip the gate" >&2
  exit 1
fi
mapfile -t docs < <(git ls-files 'docs/*.md' | grep -v '^docs/specs/2026-11-17-cc-desk-mockups/')
vale --minAlertLevel=error --output=line "${docs[@]}"

echo "gate: ok"
