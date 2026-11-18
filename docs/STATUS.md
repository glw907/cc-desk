# Status

## Current state

- The design (`docs/specs/2026-11-17-cc-desk-design.md`) and the route
  (`docs/plans/2026-11-17-route.md`, seven rounds) are approved.
- Prove landed as pull request 1, with open tests in `2026-11-17-rerun-prove-proofs.md`. State
  and rules landed as 3 to 6; two of its follow-ups need Geoff's choice when taken up:
  `orphan-question-forecast.md` and `reading-errors-and-guess-on-resume.md`.
- The Sessions and Page rounds run side by side, one pass of each at a time, from the designs
  `docs/specs/2026-11-17-sessions-design.md` and `docs/specs/2026-11-17-page-design.md`. Each
  round plan sits beside its planner prompts in `docs/plans/`.
- No `BACKLOG.md`. Follow-ups are files in `docs/follow-ups/`.
- The usage stop is 100 percent of the week. The figure read 97 percent on 2026-11-18, and
  the reset is Nov 20, 11pm Alaska time.

<!-- round-report sessions -->
**The Sessions round's first pass waits for Geoff's approval, and two more are ready behind
it.** Pull request 8 records sessions and asks one call; pull requests 10 and 12 start them and
answer their prompts, ask nothing, and land in order after 8. The fourth pass, pull request 14,
is ready for the conductor to land; its session is cc-desk-sessions-merge (job f73e67ad).

Open documents: none. Pull request 8 waits under "Open decisions".
Resume from: `docs/plans/2026-11-17-sessions-page-conductor-prompt.txt`
<!-- /round-report -->

<!-- round-report page -->
**The Page round's first four passes wait for Geoff's approval.** Pull requests 7, 9, 11, and 13
build the page, its item pages, its service at http://127.0.0.1:7908/, and its terminal link.
Each asks a choice of Geoff. The fifth pass has started. Usage read 97 percent.

Open documents: none. Pull requests 7, 9, 11, and 13 wait under "Open decisions".
Resume from: `docs/plans/2026-11-17-sessions-page-conductor-prompt.txt`
<!-- /round-report -->

## Running

- Pass 1 (sessions-check), pull request 8, built `cc-desk check`. Pass 2 (sessions-launch), pull
  request 10, carries out its decisions: a launch in a worktree the command makes, and the check's
  end, confirm, and relaunch. Round items 6 and 7 are not Go tests; they belong to pass 5, and its
  script output goes there.
- Pass 3 (sessions-prompts), pull request 12, records a permission prompt from the hook, answers it
  with `prompt yes` or `prompt no` from a desk session, adds `rules add` and `rules remove`, and
  files each answered prompt and each session's follow-up (`followup`). No verb registers a desk
  yet (`docs/follow-ups/desk-registration-gap.md`). Pass 4 (sessions-merge), pull request 14, has
  the check merge a clean pull request, send back or ask on any other, and clean up after a merge.
  Its follow-ups: `merge-question-answer.md`, `ready-verb-service-start.md`,
  `pull-request-closed-without-merge.md`, `cleanup-of-a-release-marked-done.md`.

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
- Pull request 13 (`page-link`, carries 11) waits for Geoff's approve, three calls (recommended:
  all), and his click in `docs/follow-ups/page-link-chromium-click.md`.
- The drafts of the stopped dubplate design sessions are kept in
  `~/.local/state/desk/kept-drafts/`. The Retire round removes that folder, so the Move round must
  give them a home in the dubplate repo first.
