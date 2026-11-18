# A session blocked by a failing hook is stopped for good

In the first live run of pass 5 (2026-11-18), both fixture sessions launched with a service `PATH`
that lacked `~/.local/bin`. Every Bash call failed in the `claude-launch-guard` hook with
`command not found`, and each session answered with a message and went idle. `claude agents --json`
listed them as `idle` with `state` `blocked`. At 14:56:32Z the check recorded both passes as
stopped with the reason `unreadable_stop`, and from then on it decided nothing for them. Nothing
told Geoff, and no later check would have resumed them.

Pass 5 fixed the `PATH` (commit `ba50c71`) and resumed the passes by hand with `pass start`. The
open question is what the check should do with a session that is idle and blocked for a cause it
can't read. Two choices: leave it stopped and put it in "Needs you now" with the transcript's last
message as the cause, or relaunch it once and stop only if it blocks again.
