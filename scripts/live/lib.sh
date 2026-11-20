#!/usr/bin/env bash
# lib.sh: helpers the live scripts share. Source it; never run it.
#
# The caller sets REPO and RUN (the run folder) and defines say. Nothing here runs kill, and
# nothing here starts, stops, or restarts a unit.

# LISTING_SCRIPT is the script whose `listing-loop` phase is the background reader.
LISTING_SCRIPT=${LISTING_SCRIPT:-$REPO/scripts/live-repairs.sh}
# SCRATCH_ROOT is fixed, never read from the environment: kill_target's folder guards rest on it.
SCRATCH_ROOT=$HOME/Projects/cc-desk-scratch

# listing_reads counts the reads the reader has written to listing.out.
listing_reads() {
  [ -f "$RUN/listing.out" ] || {
    echo 0
    return 0
  }
  grep -Ec '^(pass|fail) minute' "$RUN/listing.out" || true
}

# start_listing runs the reader in the background from the merge, so the conductor phase
# runs beside it. The reader's pid is recorded; abort ends it by that pid alone. A reader
# that already finished its ten reads is never started again, and listing.out is appended to,
# never truncated, since the reads taken from the merge are the evidence and a restart after
# the conductor phase would erase them.
start_listing() {
  [ -f "$RUN/merge-ids" ] || return 0
  [ -f "$RUN/listing.pid" ] && kill -0 "$(cat "$RUN/listing.pid")" 2>/dev/null && return 0
  if [ "$(listing_reads)" -ge 10 ]; then
    say "listing: the reader already holds its 10 reads; not started again"
    return 0
  fi
  setsid nohup bash "$LISTING_SCRIPT" listing-loop >>"$RUN/listing.out" 2>&1 </dev/null &
  echo $! >"$RUN/listing.pid"
  say "listing: started in the background, pid $(cat "$RUN/listing.pid"), one read a minute for 10 minutes"
}

# save_agents writes `claude agents --json` to the file named by $1. The list goes to a file,
# never through a pipe, since a reader that closes early can cut the list short. It fails
# when the list is unreadable or is not a JSON array.
save_agents() {
  claude agents --json >"$1" 2>"$1.err" && jq -e 'type == "array"' "$1" >/dev/null 2>&1
}

# kill_target prints the process id of a record's session, given its session id ($1) and a
# list saved by save_agents ($2), and only when the row is one of this suite's own claude
# processes. The name and cwd checks keep it off a real session. A background session runs as
# a claimed `claude bg-spare` process under a `bg-pty-host`, and every one of them, a real
# session's included, reads `claude` too, so the process's own working folder must also sit
# under the scratch checkout, and a pty host is refused. It prints nothing and fails
# otherwise. It never runs kill; the caller's executor does, alone.
kill_target() {
  local sid=$1 list=$2 name cwd pid
  name=$(jq -r --arg s "$sid" '[.[]? | select(.sessionId == $s)][0].name // empty' "$list")
  cwd=$(jq -r --arg s "$sid" '[.[]? | select(.sessionId == $s)][0].cwd // empty' "$list")
  pid=$(jq -r --arg s "$sid" '[.[]? | select(.sessionId == $s)][0].pid // empty' "$list")
  if [ -z "$name" ] || [ -z "$pid" ]; then
    say "kill_target: no row with session id $sid, or it holds no pid" >&2
    return 1
  fi
  case $name in
    cc-desk-scratch-*) ;;
    *)
      say "kill_target: row name '$name' does not start cc-desk-scratch-" >&2
      return 1
      ;;
  esac
  case $cwd in
    "$SCRATCH_ROOT" | "$SCRATCH_ROOT"/*) ;;
    *)
      say "kill_target: cwd '$cwd' is not under $SCRATCH_ROOT" >&2
      return 1
      ;;
  esac
  case $pid in
    '' | *[!0-9]*)
      say "kill_target: pid '$pid' is not a number" >&2
      return 1
      ;;
  esac
  if [ "$(cat "/proc/$pid/comm" 2>/dev/null || true)" != claude ]; then
    say "kill_target: /proc/$pid/comm does not read claude" >&2
    return 1
  fi
  if tr '\0' ' ' <"/proc/$pid/cmdline" 2>/dev/null | grep -q -e '--bg-pty-host'; then
    say "kill_target: /proc/$pid/cmdline holds --bg-pty-host" >&2
    return 1
  fi
  local pcwd
  pcwd=$(readlink "/proc/$pid/cwd" 2>/dev/null || true)
  case $pcwd in
    "$SCRATCH_ROOT" | "$SCRATCH_ROOT"/*) ;;
    *)
      say "kill_target: /proc/$pid/cwd '${pcwd:-unreadable}' is not under $SCRATCH_ROOT" >&2
      return 1
      ;;
  esac
  printf '%s\n' "$pid"
}

# unit_env_path prints the PATH the installed check service sets, with %h filled in. It
# reads the unit's properties and starts nothing.
unit_env_path() {
  systemctl --user show cc-desk-check.service -p Environment --value |
    tr ' ' '\n' | sed -n 's/^PATH=//p' | head -n 1 | sed "s|%h|$HOME|g"
}

# check_unit_path resolves each command the check needs under the unit's own PATH and prints
# one line each. It fails when a command does not resolve, or resolves outside the place the
# unit's PATH puts it. $1 is the PATH to test; the default is the installed unit's.
check_unit_path() {
  local upath=${1:-$(unit_env_path)} cmd want got rc=0
  if [ -z "$upath" ]; then
    say "unit PATH: the service sets no PATH"
    return 1
  fi
  say "unit PATH: $upath"
  for cmd in claude gh git go systemd-run cc-desk; do
    case $cmd in
      systemd-run) want=/usr/bin/systemd-run ;;
      cc-desk) want=$HOME/.local/share/cc-desk/bin/cc-desk ;;
      *) want=/home/linuxbrew/.linuxbrew/bin/$cmd ;;
    esac
    got=$(env -i PATH="$upath" bash -c "command -v $cmd" 2>/dev/null || true)
    if [ "$got" = "$want" ]; then
      say "unit PATH: $cmd -> $got"
    else
      say "unit PATH: $cmd resolves to '${got:-nothing}', want $want"
      rc=1
    fi
  done
  return $rc
}
