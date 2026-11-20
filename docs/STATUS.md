# Status

## Current state

- The design (`docs/specs/2026-11-17-cc-desk-design.md`) and the route
  (`docs/plans/2026-11-17-route.md`, seven rounds) are approved.
- Prove landed as pull request 1, with open tests in `2026-11-17-rerun-prove-proofs.md`. State
  and rules landed as 3 to 6; two of its follow-ups need Geoff's choice when taken up:
  `orphan-question-forecast.md` and `reading-errors-and-guess-on-resume.md`.
- The Sessions and Page rounds landed on 2026-11-18 as pull requests 7 to 16. The page serves
  main's build at http://127.0.0.1:7908/, and the check's timer runs main's binary,
  `~/.local/share/cc-desk/bin/cc-desk`, every five minutes. Follow-ups are files in
  `docs/follow-ups/`, with no `BACKLOG.md`.
- The usage stop is 100 percent of the week; `cc-desk usage` prints the figure.
- The Desk and skills round (`desk`; design `docs/specs/2026-11-18-desk-skills-design.md`, round
  plan `docs/plans/2026-11-18-desk-skills-round.md`) has landed passes 1 to 5 as pull requests 17
  to 21: the records, the skills, the page, the desk, and the check's repairs. A fresh
  conductor, `conductor-cc-desk`, took pass 6 (standing answer 36). It tagged `v0.1.0`, pinned
  both skills on the workstation (workstation pull request 251), and installed `main`'s binary,
  timer, page, and link handler. The round's report is
  `~/Projects/initiatives/cc-desk/reports/desk-report.md`.
- Every stop in the round has a row in `docs/follow-ups/desk-round-lessons.md`. The 1.1 round
  answers each row, for cc-desk and `pass-core` (standing answer 34), beside Move (answer 37).

<!-- round-report sessions -->
**The Sessions round has landed.** Pull requests 8, 10, 12, 14, and 16 record, start, answer,
merge, and time sessions. Geoff approved 8 and 16 with every recommendation.

Open documents: none.
<!-- /round-report -->

<!-- round-report page -->
**The Page round has landed.** Pull requests 7, 9, 11, 13, and 15 build the page, its item pages,
its service, its terminal link, and its polish. Geoff approved every call as recommended.

Open documents: none.
<!-- /round-report -->

## Next action

Pass 6 (`desk-run`, draft pull request 22) resumed at its rehearsal on 2026-11-20 at 17:01 UTC,
after both cc-desk repos were made public so Actions runs. `conductor-cc-desk` lands the pass
and runs the steps after it. The timer runs the branch's build, so `main`'s build isn't
reinstalled before the pass lands. Then Geoff's sitting, then his call on the Move design
(below), and 1.1's design (`docs/plans/2026-11-19-1.1-design-prompt.txt`).

## Open decisions

- The Move design (`docs/specs/2026-11-19-move-design.md`) needs a rewrite before Geoff
  approves it. dubplate's `structure` round and cairn's `checks` round run outside cc-desk,
  under their own conductors, while this week's spare usage lasts (Geoff, 2026-11-20: "We'll
  bring both other [initiatives] under the cc-desk regime once cc-desk is ready"). Move still
  brings both under cc-desk, but its first act and Questions 1 and 4 must match what those
  rounds leave behind. cairn #260 and #259 have landed.
- `conductor-cc-desk`, `conductor-dubplate`, and `conductor-cairn-cms` share one rule (agreed
  2026-11-20): at most four planners or executors on pavilion, no new session past 85 percent
  weekly usage, and a message to the others before each launch.
