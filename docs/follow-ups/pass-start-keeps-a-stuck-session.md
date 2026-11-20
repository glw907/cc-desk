# pass start keeps a stuck session

`cc-desk pass start` on a stopped pass sets it to building and closes the stop, but keeps the live session record (`internal/state/events.go`, `start`). Nothing relaunches that record: `Plan` launches only a building pass with no session (`internal/sessions/decide.go`). So the next check reads the same row.

What that means for each stop line that now names the command:

- A row still blocked on a screen the check can't read is stopped again as `unreadable_stop` at the next check. The command works only once someone has cleared the screen in the session itself.
- A pass stopped at `relaunch_limit` keeps its counted relaunches, so the next idle or missing row stops it at the limit again.
- A sign-in stop whose resumes are spent is stopped again while the row still shows "Login expired".

A fix could have `pass start` on a stopped pass end the live record, so that the next check launches a fresh session, and could reset the counted relaunches on a `relaunch_limit` stop. That changes what `pass start` does, so it needs its own design.

Found by the plan review of the stuck-sessions pass.
