# What the desk round's stops should change

Geoff's rule for this round (2026-11-19): the system being built evolves from each mistake. Every
stop in the desk round has a row here, with its cause and the change it calls for. A 1.1 round,
after the first release and before Move, answers each row with a design change or a reason to
reject it, for cc-desk and `pass-core` together (standing answer 34). Where a detail file says the
Move design settles a point, the 1.1 design settles it.

| Stop | Waited | Cause | cc-desk change | `pass-core` change | Detail |
| --- | --- | --- | --- | --- | --- |
| Pass 2 parks on its estimate | about 58 min | The twice-the-estimate stop compares two forecasts | Two figures, a scope flag, a budget on the release | Drop the stop from `round.md` | `round-estimates-priced-from-evidence.md` |
| Pass 2's review escalates | about 131 min | The landing test reads the review's word; the round plan held pass 3 behind the merge | The third verdict in `pass ready` (built in pass 5); build past a merge question | Land on findings and calls; no plan may undo run-ahead | Fable review, sections 2 and 3 |
| Pass 3 parks on its estimate | about 120 min | Same as pass 2 | Same as pass 2 | Same as pass 2 | `round-estimates-priced-from-evidence.md` |
| Pass 5 parks on its estimate past the clock | about 10.5 h | The conductor asked for a standing answer without checking its provisos against the forecast | Test each standing answer's conditions against the forecast; a standing-answer record on the release | The conductor checks provisos when it asks | `round-estimates-priced-from-evidence.md`, "The fourth stop" |
| Pass 5 parks on a leftover live session | about 85 min unseen, then a question | Nothing ends a stopped pass's session; the conductor watched only for a ready PR | A way to end a session cc-desk launched; sessions nobody owns shown on the page | The conductor's watch also ends on a blocked or done executor | `leftover-live-session-parks-a-pass.md` |
| Pass 5 parks on its own live run's sessions | about 20 min | The live script left the sessions it launched running after a failed check, and the relaunch prompt told the executor to end on any live scratch session, its own included | The live script stops every session it launched before it exits, pass or fail | A relaunch prompt lets an executor stop the sessions its own run launched; only a session it can't trace to its run stops it | `live-abort-waits-on-stopped-sessions.md` (on the desk-check branch) |
| Two PR bodies sent back | minutes each | Body rules live in four places and nothing checks them | Calls and misses taken as data by `pass ready` | One body template and a check script | Fable review, sections 2 and 3 |

The Fable review is `~/Projects/initiatives/cc-desk/reports/desk-failure-review.md`. A stop
later in this round adds a row here before the conductor goes on.
