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
- The usage stop is 100 percent of the week. The figure read 93 percent on 2026-11-18, and
  the reset is Nov 20, 10:59pm Alaska time.

<!-- round-report sessions -->
**The Sessions round's first pass waits for Geoff's approve.** Pull request 8 teaches the
`cc-desk` command to record sessions and decide what each needs. Its review accepted it, and its
check passed. It waits on one question: what a session should do after a one-model limit. The
second pass has started from its branch. Weekly usage read 93 percent.

Open documents: none. Pull request 8 waits under "Open decisions".
Resume from: `docs/plans/2026-11-17-sessions-page-conductor-prompt.txt`
<!-- /round-report -->

<!-- round-report page -->
**The Page round has started.** Its first pass builds the status page's app, its sidebar, and
its front page. Nothing has landed yet. Weekly usage read 92 percent before the first launch.
Nothing waits for Geoff: the auto-mode entry its third pass needs is in place.

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

- Pull request 8 (`sessions-check`) waits for Geoff's approve. Its question: a session stopped
  by a one-model limit relaunches once the account's usage reads under the stop line, uncounted,
  so it can loop until the reset. Keep it, hold the relaunch until the reset time in the stop's
  text (recommended), or count each relaunch after the first. The detail is
  `docs/follow-ups/usage-relaunch-loop-on-a-model-limit.md` on the branch.
- The drafts of the stopped dubplate design sessions are kept in
  `~/.local/state/desk/kept-drafts/`. The Retire round removes that folder, so the Move round must
  give them a home in the dubplate repo first.
