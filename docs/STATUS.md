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
- The State and rules round has landed (pull requests 3 to 6, 2026-11-18). Its design is
  `docs/specs/2026-11-17-state-and-rules-design.md`, and its report is
  `~/Projects/initiatives/cc-desk/reports/state-rules-report.md`.
- The usage stop is 100 percent of the week. The figure read 91 percent on 2026-11-18.

<!-- round-report state-rules -->
**The State and rules round has landed all four passes.** They are pull requests 3 to 6. The
command keeps the state files, forecasts the next session with Geoff, starts, holds, or refuses a
pass by the rules, and records a ruling only from a registered desk process with a terminal. The
round took about 191 of its planned 315 minutes. Weekly usage read 91 percent, under the 100
percent stop. Nothing waits for Geoff.

Open documents: none.
<!-- /round-report -->

## State and rules passes

The snapshot holds each release's pace and range, each pass's and round's needed time, and the
next session with `waits_for_geoff`. A start follows the waiting order, the two-pass limit, and
the usage stop; a resumed stopped pass skips both limits (Geoff, "Yes as built"). `cc-desk ruling
record` appends Geoff's words and frees what waited. Eight follow-ups from the round are in
`docs/follow-ups/`. Two need Geoff's choice when taken up: `orphan-question-forecast.md` and
`reading-errors-and-guess-on-resume.md`.

## Next action

Design the Sessions round and the Page round. The route (`docs/plans/2026-11-17-route.md`) lets
them run side by side, since both rest only on the state files. Geoff starts a fresh Opus session
in his terminal, from `~/Projects/cc-desk`, with this prompt:

```text
Design the cc-desk Sessions round and the Page round. Read docs/STATUS.md, the route
(docs/plans/2026-11-17-route.md), and the State and rules design, then follow pass-core
to design each round with me, one question at a time.
```

## Open decisions

- The drafts of the stopped dubplate design sessions are kept in
  `~/.local/state/desk/kept-drafts/`. The Retire round removes that folder, so the Move round must
  give them a home in the dubplate repo first.
