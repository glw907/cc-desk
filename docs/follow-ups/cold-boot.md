# Check the timer after a cold boot

The timer's check starts at login with `OnStartupSec=1min`. On a cold boot it may be the first thing
to start Claude Code's supervisor, the daemon that `claude --bg` sessions need. Pass 5 ran the
check on a workstation where the daemon was already up, because Geoff's sessions had started
it. A boot where the check comes first is untested, and the round can't test it without ending
every live session. The Move round's first restart is the chance to watch one.

Two things to check on that boot:

1. The check's first `claude --bg` launch starts the daemon and succeeds, or fails in a way the
   check records as a launch failure and retries.
2. The check's `gh` calls and its pushes need the keyring, which they reach over D-Bus. A user unit
   that runs before the keyring is unlocked may get no token. Pass 5's `prep` probed `gh` through
   `systemd-run` on a running desktop, and that probe says nothing about a cold boot.

If either fails, the first check after login stops with a recorded reason, and the next one five
minutes later is the retry. Decide then whether the unit needs an `After=` ordering or the check
needs to treat a missing token as a wait.
