# A merged pass's session was still listed

In the live run of pass 5 (2026-11-18), the check recorded each merge and set the session's `ended`
time. `claude agents --json` still listed both sessions at 15:24:46Z, idle with `state` `done`.
The session of `timer-live-a` (`70cdede6`) was alive 5 minutes 48 seconds after its merge at
15:18:58Z, and that of `timer-live-b` (`ba577031`) 16 seconds after its merge at 15:24:30Z. Pass 5
stopped both with a lone `claude stop` at 15:24:51Z and 15:24:55Z, so the run never showed whether
the sessions would have gone by themselves within the 10 minutes that item 6 allows.

Two readings fit the evidence. The check's cleanup may not stop the session at all, and only
records `ended`. Or a finished session leaves the list after a delay longer than the run watched.
Settle it by merging a fixture pass with the timer on and listing the sessions every minute for 10
minutes without stopping any. If the session stays, the cleanup needs a stop step.

## Geoff's ruling

On 2026-11-18 Geoff accepted item 6 of the Sessions round without showing the sessions gone when
he approved pull request 16. The next pass that merges with the timer on runs the 10-minute
listing above, and adds a stop step to the cleanup if the session stays.
