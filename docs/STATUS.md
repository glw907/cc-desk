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

## Next action

Design the State and rules round in a fresh session in Geoff's terminal. The Prove pass closed with
pull request 2 on 2026-11-17. Launch prompt, run from `~/Projects/cc-desk`:

```text
Read docs/STATUS.md, then design the State and rules round of the cc-desk route with me, using the
standards:brainstorming skill. Start from the round's row in docs/plans/2026-11-17-route.md, the
design in docs/specs/2026-11-17-cc-desk-design.md, and the findings in
docs/research/2026-11-17-prove-findings.md, with the open tests in
docs/follow-ups/2026-11-17-rerun-prove-proofs.md. Proof 7 decides how the usage stop reads its figure.
```

## Open decisions

- The drafts of the stopped dubplate design sessions are kept in
  `~/.local/state/desk/kept-drafts/`. The Retire round removes that folder, so the Move round must
  give them a home in the dubplate repo first.
