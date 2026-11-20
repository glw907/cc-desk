# Status

## Current state

- The design (`docs/specs/2026-11-17-cc-desk-design.md`) and the route
  (`docs/plans/2026-11-17-route.md`, seven rounds) are approved.
- Prove landed as pull request 1, with open tests in `2026-11-17-rerun-prove-proofs.md`. State
  and rules landed as 3 to 6; two of its follow-ups need Geoff's choice when taken up:
  `orphan-question-forecast.md` and `reading-errors-and-guess-on-resume.md`.
- The Sessions and Page rounds landed on 2026-11-18 as pull requests 7 to 16, from the designs
  `docs/specs/2026-11-17-sessions-design.md` and `docs/specs/2026-11-17-page-design.md`. The
  page serves main's build at http://127.0.0.1:7908/, and the check's timer runs main's binary,
  `~/.local/share/cc-desk/bin/cc-desk`, every five minutes. Follow-ups are files in
  `docs/follow-ups/`, with no `BACKLOG.md`; three carry Geoff's rulings of 2026-11-18.
- The usage stop is 100 percent of the week; `cc-desk usage` prints the figure.
- Geoff approved the Desk and skills round on 2026-11-18, in his terminal: the design
  (`docs/specs/2026-11-18-desk-skills-design.md`), its round plan
  (`docs/plans/2026-11-18-desk-skills-round.md`), and the auto mode entry `cc-desk-desk-round`,
  which is in the settings. The round is six passes under the id `desk`, run by the background
  session `conductor-cc-desk`. It ends with one real run that waits for Geoff at the desk, from
  `docs/specs/2026-11-18-stuck-sessions-design.md`.
- Pass 1 of that round (`desk-records`, pull request 17) gives questions, rulings, reports, rounds,
  and stops their new records, a check log, and the `brief` verb. It waits for the conductor to
  merge it; passes 2 to 4 read its names from main.
- Pass 3 of that round (`desk-open`, pull request 20) opens the desk from a bare `cc-desk`, starts
  the pass a ruling frees, replaces a wrong record before the merge, and acts on a merge
  question's three answers. It waits for the conductor to merge it.

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

## Stop

Pass 5 (desk-check, pull request 21) stops with its live clause unmet, and nothing is marked
ready. The relaunch ran `bash scripts/live-repairs.sh` twice more, which spent the two reruns it
allowed. Run 3, on `9271b50`, passed every check but two, and the unit of both is the leftover
scratch state of my own failed runs, not a defect in the check:

- "the next check started cc-desk-scratch-conductor": the rounds of runs 1 and 2 were done and
  had no report, and the check names the first such round, so it never reached this run's.
  I recorded reports `rep-0002` and `rep-0003` for them with `cc report add`, and the conductor
  launched on the next tick (`ceb3e1c9`).
- "no second conductor started": the script counted `7870b697`, the finished conductor of an
  earlier run, still in the `--all` list. The script now counts only rows started since the run's
  merge.

Two product defects surfaced and are fixed: a done conductor row that stays listed after a clean
`claude stop` blocked the end (`9271b50`), and a run on main never let the check pull the
checkout, so the conductor's branch stayed and the next launch failed (`405130e`, the script now
switches the checkout to the base branch for the merge and conductor phases). The refuse probe's
plan also says the scratch stash is nobody's work, because the probe session declined the step
once without a refusal.

No cc-desk-scratch session is live: I stopped `b2818683`, `9f9efe3b`, `a14bac73`, and `aea1e689`,
each launched by my own runs. Recommendation: relaunch `cc-desk-desk-check` for one run of the
script on the last commit. The state is clean, so it should end `result: pass`; then it replaces
this section with the evidence and marks the pull request ready.

## Next action

The conductor runs the Desk and skills round from its prompt,
`docs/plans/2026-11-18-desk-conductor-prompt.txt`. If no session named `conductor-cc-desk` is
running and the round isn't finished, launch it again from `~/Projects/cc-desk`, and it resumes
from this file and the round's report:

```text
claude --bg --model claude-opus-5-5 --effort medium --name conductor-cc-desk "$(cat /var/home/glw907/Projects/cc-desk/docs/plans/2026-11-18-desk-conductor-prompt.txt)"
```

When the round reports, the Move round is designed next. It also reshapes the cairn and dubplate
initiatives so that cc-desk can take them over (Geoff, 2026-11-18).

## Open decisions

- Geoff's click in Chromium on the page's terminal link is still owed:
  `docs/follow-ups/page-link-chromium-click.md`.
- The drafts of the stopped dubplate design sessions are kept in
  `~/.local/state/desk/kept-drafts/`. The Retire round removes that folder, so the Move round must
  give them a home in the dubplate repo first.
