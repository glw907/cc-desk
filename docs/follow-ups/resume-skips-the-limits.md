# A resumed stopped pass can put three passes building

`pass start` on a `stopped` pass resumes it without checking the two-pass limit or the usage stop.
The design says passes already building carry on, and a stopped pass held its worktree and its
session, so the rule lets it back in. Only `building` passes count toward two, though, so a resume
while two others are building leaves three.

The code is `start` in `internal/state/events.go`, which returns the resume before it reads
`s.Limits()`. The `TestSettle` case "a resumed stopped pass ignores both limits" covers the resume.

Two fixes are open. Keep the rule, and accept the extra session now and then. Or make a resume wait
on the two-pass limit like any other start, which would leave a stopped pass holding its worktree
with nothing running in it. Recommended: keep it, since stopping a pass twice costs more than one
extra session.

Geoff's call, since it sets how many sessions run at once. The pull request for the start rules
carries it as its one call.
