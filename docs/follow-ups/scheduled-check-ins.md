# Scheduled development check-ins

Geoff's feature for a 1.2 round (2026-11-19), with his spelling corrected: he wants to
schedule development check-in sessions. A likely shape is two scheduled sessions a day, 8 to 12
hours apart, though this may vary. He would schedule them in advance with cc-desk, which would
work to maximize the development work done between sessions.

## What it changes

Today's design runs the other way. `docs/specs/2026-11-17-cc-desk-design.md` "Forecasting
decisions" has cc-desk forecast when work will next need Geoff, and the page shows that time as
"The next decision session". With this feature, Geoff sets the times in advance, and cc-desk plans
the work between them.

## Points for the 1.2 design to settle

These are the conductor's notes, not Geoff's. The brainstorm with Geoff settles them.

- How Geoff sets and changes the schedule: a desk verb, a release field, or the page.
- What "maximize the work between sessions" means for the order of passes. One reading: start
  first the passes whose questions can be asked at the coming check-in, and save for the long
  gap the passes that need nothing from him.
- Asking ahead. Before a check-in, cc-desk could gather every question it can foresee for the
  next gap, so the gap doesn't stop on one. The lessons ledger's fourth row (pass 5 waited about
  10.5 hours overnight) is the case this would have caught.
- What happens when a question arises mid-gap: wait for the next check-in, build on the
  recommended answer (the main design's run-ahead), or both.
- Whether the forecast stays, as a check that the schedule and the work fit, and what the page
  shows when they don't.
- How it fits the usage stop and the other release fields the ledger proposes (the overrun
  budget and standing answers as a record).

Geoff expects this to be larger than the 1.1 fixes, so it is its own 1.2 round, after 1.1. Its
order against Move is his call. The 1.1 design prompt names this file so that the release fields
1.1 adds (the overrun budget, standing answers as a record) leave room for a schedule.
