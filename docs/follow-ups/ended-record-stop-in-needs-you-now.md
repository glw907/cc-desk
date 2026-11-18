# An ended record with a stop reason still reads "Needs you now"

In `internal/snapshot/sessions.go`, the `add` function gives a `needs_you_now` entry to any
conductor or design writer record that carries a `stop_reason`, whether or not the record also
carries `ended`. Nothing in the check sets `ended`, since ending a session is pass 2's work, so no
entry like this exists yet.

Once pass 2 ends sessions, a record that was stopped and then ended on purpose would keep asking
Geoff for something he has already settled. Pass 2 should decide whether `ended` clears the entry,
and add a case to `TestNeedsYouNowStoppedConductor`'s neighbors in `internal/snapshot/sessions_test.go` for it.

Recommended: an ended record never reads "Needs you now", because the end is the answer.
