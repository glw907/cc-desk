# Status

## Current state

- The design (`docs/specs/2026-11-17-cc-desk-design.md`) and the route
  (`docs/plans/2026-11-17-route.md`, seven rounds) are approved.
- The Prove round is merged as pull request 1. Its report is
  `docs/research/2026-11-17-prove-findings.md`: five proofs held, and two fall back to a restart
  from the files. Open tests are in `docs/follow-ups/2026-11-17-rerun-prove-proofs.md`.
- The State and rules round has landed as pull requests 3 to 6. Its report is
  `~/Projects/initiatives/cc-desk/reports/state-rules-report.md`. Two of its follow-ups need
  Geoff's choice when taken up: `orphan-question-forecast.md` and
  `reading-errors-and-guess-on-resume.md`.
- The Sessions round and the Page round run side by side, one pass of each at a time. Geoff
  approved both designs by starting the conductor on 2026-11-18. The designs are
  `docs/specs/2026-11-17-sessions-design.md` and `docs/specs/2026-11-17-page-design.md`, and
  each round plan sits beside its planner prompts in `docs/plans/`.
- No `BACKLOG.md`. Follow-ups are files in `docs/follow-ups/`.
- The usage stop is 100 percent of the week. The figure read 92 percent on 2026-11-18, and
  the reset is Nov 20, 10:59pm Alaska time.

<!-- round-report sessions -->
**The Sessions round has started.** Its first pass teaches the `cc-desk` command to read the
session list and decide what each session needs. Nothing has landed yet. Weekly usage read 92
percent before the first launch. One thing waits for Geoff: his approve of the auto-mode entry
in the round plan, which the fifth pass needs.

Open documents: none.
Resume from: `docs/plans/2026-11-17-sessions-page-conductor-prompt.txt`
<!-- /round-report -->

<!-- round-report page -->
**The Page round has started.** Its first pass builds the status page's app, its sidebar, and
its front page. Nothing has landed yet. Weekly usage read 92 percent before the first launch.
One thing waits for Geoff: the same auto-mode entry, which the third pass needs.

Open documents: none.
Resume from: `docs/plans/2026-11-17-sessions-page-conductor-prompt.txt`
<!-- /round-report -->

## Next action

The conductor runs in Geoff's terminal. After a crash, any session can resume it from
`~/Projects/cc-desk` with this line:

```text
claude --model claude-opus-5-5 --effort high "$(cat docs/plans/2026-11-17-sessions-page-conductor-prompt.txt)"
```

## Open decisions

- The drafts of the stopped dubplate design sessions are kept in
  `~/.local/state/desk/kept-drafts/`. The Retire round removes that folder, so the Move round must
  give them a home in the dubplate repo first.
