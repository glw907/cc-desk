# A session blocked by a failing hook is stopped for good

In the first live run of pass 5 (2026-11-18), both fixture sessions launched with a service `PATH`
that lacked `~/.local/bin`. Every Bash call failed in the `claude-launch-guard` hook with
`command not found`, and each session answered with a message and went idle. `claude agents --json`
listed them as `idle` with `state` `blocked`. At 14:56:32Z the check recorded both passes as
stopped with the reason `unreadable_stop`, and from then on it decided nothing for them. Nothing
told Geoff, and no later check would have resumed them.

Pass 5 fixed the `PATH` (commit `ba50c71`) and resumed the passes by hand with `pass start`.

The desk-page pass settled part of this. The stop stays, and the page now says it in words: a
pass's page gives the stop's time and a sentence for its reason, and "Needs you now" lists it. What
stays open is the cause. The page can't say why the session blocked, because the check doesn't read
the transcript. Show the transcript's last message as the cause, so a failing hook reads as a
failing hook and not as an unreadable stop.
