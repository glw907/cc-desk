# A relaunch after a per-model limit can loop

A session stopped by a per-model limit, such as "You've hit your Fable limit", is stopped with the
reason `usage`. The design (rows 4 and 9) relaunches it as soon as the usage reading is under the
stop line, and that relaunch counts nothing against the limit of relaunches. The reading is the
account-wide percentage, so it can be well under 90 while the one model is still at its limit. The
relaunched session then hits the same limit, stops again, and the check relaunches it again, with
nothing in the count to end the loop.

`TestCheckStops/a_usage-limit_relaunch` proves the designed behavior: at 100 percent the decision is
`wait_usage`, and at 40 percent it is an uncounted `relaunch`.

Three ways out:

1. Build it as designed and accept the loop until the limit's reset.
2. Hold the relaunch until the reset time in the stop's text ("resets Nov 6, 11pm"), which the
   classifier already keeps in `stop_detail`.
3. Count every usage relaunch after the first, so the relaunch limit ends the loop.

Recommended: hold until the reset time. It stops the loop at its cause, and the reset text is
already recorded. The cost is that a stop with no readable reset time needs a fallback, and the
natural one is option 3.

Geoff's call, since it sets when a stopped pass runs again. The pull request for the check carries
it as its one call, and the design stands until he rules.
