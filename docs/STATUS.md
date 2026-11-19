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

Pass 5 (desk-check, pull request 21) stops before its live evidence. The relaunch ran
`bash scripts/live-repairs.sh` on `e5017f2`. Two checks failed, and the run left two scratch
sessions live that block the next `prep`: `757f73c4` (`cc-desk-scratch-refuse-20261119t221850z`,
blocked) and `cedbf0db` (`cc-desk-scratch-hold-20261119t221850z`, working). Stop both, then
relaunch `cc-desk-desk-check`; it runs the script again on the last commit and marks the pull
request ready.

The first failure was the script's own: it looked for the hook's input only under
`hook-unmatched/`, and matched input lands on the session record. The second was real: a
finished conductor's session reads `done`, which `claude stop` refuses, so the end failed on
every check and its worktree stayed. Commit `70e3b15` fixes both.

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
