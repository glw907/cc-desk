# Status

## Current state

- The design (`docs/specs/2026-11-17-cc-desk-design.md`) and the route
  (`docs/plans/2026-11-17-route.md`, seven rounds) are approved.
- The Prove pass's first run is merged as pull request 1. It holds the Go scaffold, `scripts/gate.sh`,
  the CI workflow, and `spikes/own.sh`, the helper that launches and controls proof sessions.
- The first run stopped on the plan's old 85 percent usage limit and launched no proof session.
  The ceiling is 100 percent since 2026-11-17, and the second run ran all six proofs.
- The report is `docs/research/2026-11-17-prove-findings.md`. Proofs 5, 4, 2, 3, and 7 held. Proofs 1
  and 6 fall back: the check restarts a blocked or killed session from the files, and a fresh
  session starts from the plan and branch. A pass runs with zero prompts under five allow rules.
- Open tests are in `docs/follow-ups/2026-11-17-rerun-prove-proofs.md`.
- No `BACKLOG.md`. Follow-ups are files in `docs/follow-ups/`.
- The State and rules round is designed and approved (Geoff, 2026-11-17):
  `docs/specs/2026-11-17-state-and-rules-design.md`. Its round plan is
  `docs/plans/2026-11-17-state-rules-round.md`: four passes in one chain, 315 planned minutes,
  round id `state-rules`. No pass has started.
- The usage stop is 100 percent of the week. The figure read 88 percent on 2026-11-17.

<!-- round-report state-rules -->
**The start rules are built and wait on one question for Geoff, and the last pass is starting.**
Passes 1 and 2 have landed as pull requests 3 and 4. Pass 3 is ready as pull request 5 and waits
for Geoff's answer under "Open decisions". Pass 4, which records rulings, is starting from pass
3's branch. Weekly usage read 90 percent, under the 100 percent stop.

Open documents: none. Pull request 5 waits for Geoff's answer in his terminal session.
Resume from: `docs/plans/2026-11-17-state-rules-conductor-prompt.txt`
<!-- /round-report -->

## State and rules passes

Passes 1 and 2 have landed. The snapshot holds each release's pace and range, each pass's and
round's needed time, and the next session with `waits_for_geoff`. Each pass event appends a line
to `forecasts.jsonl`. `settle` in `internal/state/commit.go` is still empty for pass 3, which also
adds the unreadable usage figure to "Needs you now".

## Next action

Run the State and rules round. Geoff starts the conductor in a fresh Opus session in his
terminal, from `~/Projects/cc-desk`, with this prompt:

```text
You are the conductor of the cc-desk State and rules round. Read
docs/plans/2026-11-17-state-rules-conductor-prompt.txt and follow it from its first line.
```

## Open decisions

- Pull request 5 (the start rules) waits for Geoff's answer: should a stopped pass that resumes
  skip the limit of two passes building at once? Recommended: yes, as built. See
  `docs/follow-ups/resume-skips-the-limits.md` on its branch. Pass 4 builds on its branch and
  lands after it.
- The drafts of the stopped dubplate design sessions are kept in
  `~/.local/state/desk/kept-drafts/`. The Retire round removes that folder, so the Move round must
  give them a home in the dubplate repo first.
