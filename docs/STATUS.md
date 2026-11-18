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

Design the Desk and skills round with Geoff. From `~/Projects/cc-desk`, launch the design session:

```text
claude --model claude-fable-5-2 --effort high "$(cat docs/plans/2026-11-18-desk-skills-design-prompt.txt)"
```

## Open decisions

- Geoff's click in Chromium on the page's terminal link is still owed:
  `docs/follow-ups/page-link-chromium-click.md`.
- The drafts of the stopped dubplate design sessions are kept in
  `~/.local/state/desk/kept-drafts/`. The Retire round removes that folder, so the Move round must
  give them a home in the dubplate repo first.
