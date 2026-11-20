# The live script's abort waits on a session the check leaves alone

`scripts/live-repairs.sh abort` stops the passes of the live run and then waits for the check to
end their sessions. The check leaves a stopped pass's session alone when the stop carries a reason:
`early` in `internal/sessions/decide.go` returns `none` for it. A pass the script stopped with
`--reason scope` therefore keeps its session for good. The wait prints "still live after the wait"
and gives up, and the next `prep` refuses while any `cc-desk-scratch-` session is live.

A run on 2026-11-19 left two sessions this way, the blocked refuse probe and the working hold pass,
after the script ended. Someone has to stop them by hand before the script can run again.

Two ways out: stop the pass without a reason in the script, so the check ends the session, or add
a `cc-desk` verb that ends a named session, which the script and a person can both call.
