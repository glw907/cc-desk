#!/bin/sh
# The limits phase's systemd-run stand-in: it checks the three flags a launch passes, then runs
# the rest of the command line, so no scope is made for the claude stand-in.
if [ "$1 $2 $3" != "--user --scope -q" ]; then
  echo "systemd-run stand-in: unexpected flags: $*" >&2
  exit 64
fi
shift 3
exec "$@"
