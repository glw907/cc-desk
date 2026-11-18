# Status

## Current state

- The design (`docs/specs/2026-11-17-cc-desk-design.md`) and the route
  (`docs/plans/2026-11-17-route.md`, seven rounds) are approved.
- The Prove round is merged as pull request 1 (`docs/research/2026-11-17-prove-findings.md`);
  its open tests are in `docs/follow-ups/2026-11-17-rerun-prove-proofs.md`.
- The State and rules round has landed as pull requests 3 to 6. Its report is
  `~/Projects/initiatives/cc-desk/reports/state-rules-report.md`. Two of its follow-ups need
  Geoff's choice when taken up: `orphan-question-forecast.md` and
  `reading-errors-and-guess-on-resume.md`.
- The Sessions and Page rounds run side by side, one pass of each at a time, from the designs
  `docs/specs/2026-11-17-sessions-design.md` and `docs/specs/2026-11-17-page-design.md`. Each
  round plan sits beside its planner prompts in `docs/plans/`.
- No `BACKLOG.md`. Follow-ups are files in `docs/follow-ups/`.
- The usage stop is 100 percent of the week. The figure read 96 percent on 2026-11-18, and
  the reset is Nov 20, 10:59pm Alaska time.

<!-- round-report sessions -->
**The Sessions round's first pass waits for Geoff's approval, and the second is ready behind
it.** Pull request 8 records sessions and asks one call; pull request 10 starts them, asks
nothing, and lands right after 8. The third pass has started. Usage read 95 percent.

Open documents: none. Pull request 8 waits under "Open decisions".
Resume from: `docs/plans/2026-11-17-sessions-page-conductor-prompt.txt`
<!-- /round-report -->

<!-- round-report page -->
**The Page round's first three passes wait for Geoff's approval.** Pull requests 7, 9, and 11
build the page, its item pages, and its service at http://127.0.0.1:7908/. Each passed its
review and check and asks a choice of Geoff. The fourth pass has started. Usage read 96 percent.

Open documents: none. Pull requests 7, 9, and 11 wait under "Open decisions".
Resume from: `docs/plans/2026-11-17-sessions-page-conductor-prompt.txt`
<!-- /round-report -->

## Next action

The conductor runs in Geoff's terminal. After a crash, any session can resume it from
`~/Projects/cc-desk` with this line:

```text
claude --model claude-opus-5-5 --effort high "$(cat docs/plans/2026-11-17-sessions-page-conductor-prompt.txt)"
```

## Open decisions

- Pull request 8 (`sessions-check`) waits for Geoff's approve and one call: when a session
  stopped by a one-model limit restarts. Recommended: at the reset time in its stop text
  (`docs/follow-ups/usage-relaunch-loop-on-a-model-limit.md` on the branch).
- Pull request 7 (`page-front`) waits for Geoff's approve on two calls (recommended: both):
  `trustedOrigins: ['*']` so a write gets 405, and a missing `snapshot.json` clears the page.
- Pull request 9 (`page-items`, carries 7) waits for Geoff's approve and one call: closed rounds
  list their passes as muted links on the front page (recommended: keep).
- Pull request 11 (`page-surface`, carries 9) waits for Geoff's approve and one call: the page
  reads "No state yet" until `~/.local/state/cc-desk` exists (recommended: approve).
- The drafts of the stopped dubplate design sessions are kept in
  `~/.local/state/desk/kept-drafts/`. The Retire round removes that folder, so the Move round must
  give them a home in the dubplate repo first.
