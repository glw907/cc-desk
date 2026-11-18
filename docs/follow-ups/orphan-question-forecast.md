# A question no pass rests on leaves a null forecast time

A written, unanswered question that no pass rests on has no round. The snapshot treats it as needed
now and the next session reads `ready_now` with a time of now.

`snapshot.Render` writes the forecast's names only when the setter has a round. A pass event in
that state appends a line with `"time": null` while `snapshot.json` shows a time. The estimate
pass's acceptance says the two must match.

`TestNextSessionQuestionWithoutPass` covers the snapshot side. Two fixes are open: leave a question
with no pass out of the session's items, or let a forecast line carry a time with no round, which
changes the strict reader's "project, round, and time together or none" rule in
`internal/state/read.go`. Geoff's call, since the first changes what "Ready now" means.

Found by the whole-branch review of pull request 4.
