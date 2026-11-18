#!/usr/bin/env bash
# Proves the xdg-open chain on this workstation after install-link.sh: each bad address reaches the
# handler and is refused, and the good one runs a stand-in cc-desk once with "alpha r1".
#
# xdg-open exits 0 whatever the handler does, so a refusal is proved by the handler's stderr line,
# and a run by the record the stand-in writes.
set -euo pipefail

root="${CLAUDE_JOB_DIR:+$CLAUDE_JOB_DIR/tmp}"
dir=$(mktemp -d "${root:-${TMPDIR:-/tmp}}/check-link.XXXXXX")
trap 'rm -rf "$dir"' EXIT

record="$dir/record"
mkdir "$dir/bin"
# ptyxis runs the command in its own service, so the record path is written into the stand-in.
printf '#!/bin/sh\necho "$*" >> %q\n' "$record" > "$dir/bin/cc-desk"
chmod +x "$dir/bin/cc-desk"
export PATH="$dir/bin:$PATH"

bad=(
	'cc-desk://alpha/r1;x'
	'cc-desk://alpha/r1%3Bx'
	'cc-desk://Alpha/r1'
	'cc-desk://alpha/r1/x'
	'cc-desk://alpha/../r1'
	'cc-desk://alpha/r1?x=1'
)
n=0
for address in "${bad[@]}"; do
	n=$((n + 1))
	err="$dir/$n.err"
	xdg-open "$address" 2> "$err" || true
	for _ in $(seq 50); do
		grep -q 'cc-desk-open: refused' "$err" 2> /dev/null && break
		sleep 0.1
	done
	if ! grep -q 'cc-desk-open: refused' "$err"; then
		echo "check-link: $address was not refused" >&2
		exit 1
	fi
	echo "$address -> $(grep 'cc-desk-open: refused' "$err" | head -1)"
done
if [[ -e $record ]]; then
	echo "check-link: a refused address ran the command: $(cat "$record")" >&2
	exit 1
fi

xdg-open cc-desk://alpha/r1 || true
for _ in $(seq 100); do
	[[ -s $record ]] && break
	sleep 0.1
done
if [[ ! -s $record ]]; then
	echo "check-link: cc-desk://alpha/r1 ran nothing" >&2
	exit 1
fi
# A second run would show as a second line; wait a moment for one.
sleep 1
if [[ $(cat "$record") != 'alpha r1' ]]; then
	echo "check-link: the record holds: $(cat "$record")" >&2
	exit 1
fi
echo "cc-desk://alpha/r1 -> record: $(cat "$record")"
echo "check-link: ok"
