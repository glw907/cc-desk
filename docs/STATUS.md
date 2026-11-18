# Status

## Current state

- The design (`docs/specs/2026-11-17-cc-desk-design.md`) and the route
  (`docs/plans/2026-11-17-route.md`, seven rounds) are approved.
- The Prove pass's first run landed as pull request 1 on 2026-11-17. It built the Go scaffold,
  `scripts/gate.sh`, the CI workflow, and `spikes/own.sh`, the helper that launches and controls
  proof sessions. A second review found three blocking bugs in `own.sh`, fixed before the merge.
- That run stopped on the plan's old 85 percent usage limit, with usage at 87 percent, so it
  launched no proof session. Proof 7 held through `claude -p "/usage"`. Proofs 1 to 6 close as
  not attempted in `docs/research/2026-11-17-prove-findings.md` until the second run replaces them.
- Geoff raised this pass's weekly usage ceiling to 100 percent on 2026-11-17, in his terminal
  session. At the limit the pass pauses and never closes. He resets the allowance, and his
  terminal session wakes the executor. The plan's "Stops and their answers" row says the same.
- No `BACKLOG.md`. Follow-ups are files in `docs/follow-ups/`.

## Next action

The second run, proofs 1 to 6, runs as the background session `cc-desk-prove`, launched from the
main checkout with `docs/plans/2026-11-17-prove-rerun-executor-prompt.txt`. It closes the pass
itself and rewrites this file at close. If it pauses at the usage limit, wake it after Geoff's
reset. Then design the State and rules round from the findings.

## Open decisions

- The drafts of the stopped dubplate design sessions are kept in
  `~/.local/state/desk/kept-drafts/`. The Retire round removes that folder, so the Move round must
  give them a home in the dubplate repo first.
