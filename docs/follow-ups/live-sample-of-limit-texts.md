# A live sample of the sign-in and limit texts

The classifier reads a blocked row from the job state file's `detail`, then from the whole
`claude logs` screen. Its patterns for "Login expired" and for a usage limit ("You've hit your weekly
limit · resets ...") come from the spike captures in `spikes/2-states/` and from the job file the
fixtures copy. None came from a live session at the moment it stopped.

Proof 2 of the design left this as an open test: capture the screen and the job file of a real
session at an expired sign-in and at a real limit, and run the classifier on them. Until then a
change in Claude Code's wording reads as `unreadable_stop`, which still reaches "Needs you now"
with the last 200 runes of the screen as its cause, so the failure is loud, not silent.

The test needs a session that really hits a limit, so it waits on a day Geoff is near one. Save the
two captures under `spikes/2-states/out/`, add them to `e2e/testdata/check/` as new fixtures beside
the old ones, and add a case to `TestClassifyFixtures`.
