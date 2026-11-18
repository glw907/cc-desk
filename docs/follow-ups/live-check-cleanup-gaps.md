# Close the cleanup gaps in the live-check script

The whole-branch review of pass 5 (sessions-timer) accepted the branch and left these gaps in
`scripts/live-check.sh` and `cmd/cc-desk/desk.go`. None changed the live run's result.

1. `session_ids` reads the id that `watch` saved and the record's current id. The first run's two
   blocked sessions, and any session a hand resume replaced, are never read, so the count of
   `gh pr merge` calls covers only part of the transcripts. Collect every session id from the
   transcripts folder or the check's events.
2. `abort` prints `pass stop` and `claude stop` lines only for passes that read `building` or
   `waiting`. A pass the check marked `stopped` with an idle session still alive gets no
   `claude stop` line.
3. `abort` disables the timer but not the service. A fixture session still running can call
   `pass ready`, which kicks one more check before the stop lines are carried out.
4. `prep` has no trap. A failure after `timer install`, such as the `gh` probe or the push, exits
   with the timer enabled and a half-cut base branch.
5. If the outer `systemd-run` of the `scope` phase fails, the `systemctl --user stop` of the
   probe's scope never runs, and the `sleep` stays alive until it ends.
6. `desk register` runs `claude agents --json` with no timeout, where `internal/sessions/read.go`
   uses a context. A hung `claude` hangs the verb.
7. The hidden `--claude-command` flag lets a caller supply a row list. The ancestor-terminal
   check is the real barrier. Weigh it again if `desk register` is ever classed as `auth-data`.

Pass 5's live evidence reads three transcripts, and the first run's two blocked sessions are not
among them. The Acceptance section of PR 16 says so.
