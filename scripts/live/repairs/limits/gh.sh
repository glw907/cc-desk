#!/bin/sh
# The limits phase's gh stand-in. It sits beside gh.py, a copy of e2e/testdata/check-merge/gh.py.
# With the control file sleep-merge in this folder, a `pr merge` call starts a child `sleep 120`
# that holds stdout, writes the child's process id to child.pid, and waits, so only a group
# kill ends it. Every other call, and a `pr merge` with no control, goes to gh.py.
d=$(cd "$(dirname "$0")" && pwd)
if [ "$1 $2" = "pr merge" ] && [ -e "$d/sleep-merge" ]; then
  echo "$(date +%s.%N) merge" >>"$d/sleep-calls.log"
  sleep 120 &
  echo $! >"$d/child.pid"
  wait
  exit 1
fi
exec python3 "$d/gh.py" "$@"
