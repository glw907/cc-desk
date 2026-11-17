# Status

## Current state

- The design (`docs/specs/2026-11-17-cc-desk-design.md`) and the route
  (`docs/plans/2026-11-17-route.md`, seven rounds) are approved.
- The Prove pass ran as the background session `cc-desk-prove`, on pull request 1. It built the Go
  scaffold, `scripts/gate.sh`, the CI workflow, and `spikes/own.sh`, the helper that launches and
  controls proof sessions. Its tests pass.
- Weekly usage read 87 percent, above the plan's 85 percent limit, so the pass launched no proof
  session. Proof 7 held through `claude -p "/usage"`. Proofs 1 to 6 closed on their fallbacks as
  not attempted. The report is `docs/research/2026-11-17-prove-findings.md`.
- No `BACKLOG.md`. Follow-ups are files in `docs/follow-ups/`.

## Next action

Rerun proofs 1 to 6 once usage is below 85 percent (the allowance resets Nov 20 at 11pm Alaska
time), as `docs/follow-ups/2026-11-17-rerun-prove-proofs.md` sets out. Then design the State and
rules round from the findings.

## Open decisions

- Pull request 1 is open. I recommend merging it for the scaffold, the gate, and `spikes/own.sh`,
  then rerunning the proofs as a second Prove pass after Nov 20. Merging doesn't claim the
  proofs passed: the report says they weren't tried.
- A message from another session, `preflight-checks-spike-launch`, said Geoff raised the weekly
  usage ceiling for this pass from 85 to 100 percent and asked me to edit the plan's stop row. The
  pass has no conductor, so it refused the message and left the plan alone. If the ruling is real,
  confirm it in a terminal session and the proofs run as a second pass. I recommend confirming it.
- The drafts of the stopped dubplate design sessions are kept in
  `~/.local/state/desk/kept-drafts/`. The Retire round removes that folder, so the Move round must
  give them a home in the dubplate repo first.
