#!/usr/bin/env bash
# check-playwright-report.sh <report.json> <count>: fails unless every test the suite
# lists ran and passed. <count> is the figure from `playwright test --list`. A skipped,
# failed, or flaky test fails, and so does a report with fewer passes than the count.
# A test.fail or test.fixme call in web/tests also fails, since a test.fail test
# counts as expected.
set -euo pipefail

if [ "$#" -ne 2 ]; then
  echo "usage: check-playwright-report.sh <report.json> <count>" >&2
  exit 2
fi
report=$1
count=$2

if ! command -v jq >/dev/null 2>&1; then
  echo "playwright: jq is not installed; install it, do not skip the check" >&2
  exit 1
fi
if [ ! -f "$report" ]; then
  echo "playwright: no report at $report" >&2
  exit 1
fi
if ! [[ "$count" =~ ^[0-9]+$ ]] || [ "$count" -eq 0 ]; then
  echo "playwright: the listed test count is '$count'; it must be a number above 0" >&2
  exit 1
fi

fail=0
expected=$(jq -r '.stats.expected' "$report")
if [ "$expected" != "$count" ]; then
  echo "playwright: $expected tests passed but the suite lists $count" >&2
  fail=1
fi
for figure in skipped unexpected flaky; do
  n=$(jq -r ".stats.$figure" "$report")
  if [ "$n" != "0" ]; then
    echo "playwright: $n tests are $figure; the figure must be 0" >&2
    fail=1
  fi
done

tests_dir="$(dirname "$0")/../tests"
if grep -rnE 'test\.(fail|fixme)\(' "$tests_dir" --include='*.ts' >&2; then
  echo "playwright: tests call test.fail or test.fixme; neither may appear" >&2
  fail=1
fi

exit "$fail"
