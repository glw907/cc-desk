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
**All four passes are built, and pass 4 waits on its pull request.** Passes 1 to 3 have landed as
pull requests 3, 4, and 5. Pass 4 is pull request 6. The command keeps the state files, forecasts
the next session with Geoff, starts, holds, or refuses a pass by the rules, and records a ruling
from a registered desk session. Weekly usage read 91 percent on the live run, under the 100 percent
stop. Nothing waits for Geoff.

Open documents: none.
Resume from: `docs/plans/2026-11-17-state-rules-conductor-prompt.txt`
<!-- /round-report -->

## State and rules passes

Passes 1 to 3 have landed. The snapshot holds each release's pace and range, each pass's and
round's needed time, and the next session with `waits_for_geoff`. Each pass event appends a line
to `forecasts.jsonl`.

Pass 3 (`state-rules-order`) landed as pull request 5. A start now begins, waits on the two-pass
limit or on usage, or is refused with every broken rule named. `cc-desk usage` and `cc-desk free`
are new. Geoff kept the resume rule as built: a resumed stopped pass skips both limits.

Pass 4 (`state-rules-rulings`) is pull request 6. `cc-desk ruling record` appends one line holding
Geoff's words, and refuses with the failing test named unless the caller is under a registered desk
process that has a terminal. A ruling answers its question, closes any guess on it as held or
missed, and frees what waited. Two follow-ups are filed: the orphan ruling after a crash, and a desk
that lives through a clock step.

## Next action

Run the State and rules round. Geoff starts the conductor in a fresh Opus session in his
terminal, from `~/Projects/cc-desk`, with this prompt:

```text
You are the conductor of the cc-desk State and rules round. Read
docs/plans/2026-11-17-state-rules-conductor-prompt.txt and follow it from its first line.
```

## Open decisions

- The drafts of the stopped dubplate design sessions are kept in
  `~/.local/state/desk/kept-drafts/`. The Retire round removes that folder, so the Move round must
  give them a home in the dubplate repo first.
