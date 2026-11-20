#!/bin/sh
# The limits phase's git stand-in, written to <fixture>/bin/git with the two words below filled
# in. A `git fetch` or `git pull` aimed at the fixture's own checkout, when the control file
# sleep-fetch or sleep-pull is in the fixture folder, starts a child `sleep 120` that holds
# stdout, writes its process id to child.pid, and waits. Every other call goes to the real git
# by absolute path, so the gh stand-in's own clone and the worktree's commits are untouched.
fixture=@FIXTURE@
real=@REAL_GIT@
sub=
target=$(pwd)
skip=
for a in "$@"; do
  if [ -n "$skip" ]; then
    [ "$skip" = C ] && target=$a
    skip=
    continue
  fi
  case $a in
    -C) skip=C ;;
    -c) skip=c ;;
    -*) ;;
    *)
      sub=$a
      break
      ;;
  esac
done
case $sub in
  fetch | pull)
    case $target in
      "$fixture"/project | "$fixture"/project/*)
        if [ -e "$fixture/sleep-$sub" ]; then
          echo "$(date +%s.%N) $sub" >>"$fixture/sleep-calls.log"
          sleep 120 &
          echo $! >"$fixture/child.pid"
          wait
          exit 1
        fi
        ;;
    esac
    ;;
esac
exec "$real" "$@"
