# Status

## Current state

- The design (`docs/specs/2026-11-17-cc-desk-design.md`) and the route
  (`docs/plans/2026-11-17-route.md`, seven rounds) are approved.
- Prove landed as pull request 1, with open tests in `2026-11-17-rerun-prove-proofs.md`. State
  and rules landed as 3 to 6; two of its follow-ups need Geoff's choice when taken up:
  `orphan-question-forecast.md` and `reading-errors-and-guess-on-resume.md`.
- The Sessions and Page rounds run side by side, one pass of each at a time, from the designs
  `docs/specs/2026-11-17-sessions-design.md` and `docs/specs/2026-11-17-page-design.md`. Each
  round plan sits beside its planner prompts in `docs/plans/`. Follow-ups are files in
  `docs/follow-ups/`, with no `BACKLOG.md`.
- The usage stop is 100 percent of the week. The figure read 99 percent on 2026-11-18, and
  the reset is Nov 20, 11pm Alaska time.

<!-- round-report sessions -->
**The Sessions round is built, and its five pull requests wait for Geoff's approval.** Pull
requests 8 to 14 record, start, answer, and merge sessions; 16 runs the check on a timer, live.
Pull requests 8 and 16 ask calls, and the rest land in order behind them.

Open documents: none. Pull requests 8 and 16 wait under "Open decisions".
Resume from: `docs/plans/2026-11-17-sessions-page-conductor-prompt.txt`
<!-- /round-report -->

<!-- round-report page -->
**The Page round is built, and its five pull requests wait for Geoff's approval.** Pull requests
7, 9, 11, and 13 build the page, its item pages, its service at http://127.0.0.1:7908/, and its
link, and each asks a choice. Pull request 15 polishes the page, asks nothing, and lands after 13.

Open documents: none. Pull requests 7, 9, 11, and 13 wait under "Open decisions".
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
- Pull requests 9 (`page-items`) and 11 (`page-surface`) wait for Geoff's approve and one call
  each (recommended: approve): closed rounds list their passes as muted links on the front page,
  and the page reads "No state yet" until `~/.local/state/cc-desk` exists.
- Pull request 13 (`page-link`, carries 11) waits for Geoff's approve, three calls (recommended:
  all), and his click in `docs/follow-ups/page-link-chromium-click.md`.
- Pull request 16 (`sessions-timer`, carries 14) waits for Geoff's approve and two calls: accept
  item 6 without its hook test and its 10-minute session check (recommended: both).
- The drafts of the stopped dubplate design sessions are kept in
  `~/.local/state/desk/kept-drafts/`. The Retire round removes that folder, so the Move round must
  give them a home in the dubplate repo first.
