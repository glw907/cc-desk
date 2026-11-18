# Status

## Current state

- The design (`docs/specs/2026-11-17-cc-desk-design.md`) and the route
  (`docs/plans/2026-11-17-route.md`, seven rounds) are approved.
- The Prove round is merged as pull request 1. Its report is
  `docs/research/2026-11-17-prove-findings.md`: five proofs held, and two fall back to a restart
  from the files. Open tests are in `docs/follow-ups/2026-11-17-rerun-prove-proofs.md`.
- The State and rules round has landed (pull requests 3 to 6, 2026-11-18). Its design is
  `docs/specs/2026-11-17-state-and-rules-design.md`, and its report is
  `~/Projects/initiatives/cc-desk/reports/state-rules-report.md`.
- The Sessions round and the Page round are designed and planned, and nothing of them is
  built. The designs are `docs/specs/2026-11-17-sessions-design.md` and
  `docs/specs/2026-11-17-page-design.md`. Each has a round plan and five planner prompts in
  `docs/plans/`. Geoff has read neither design yet.
- No `BACKLOG.md`. Follow-ups are files in `docs/follow-ups/`.
- The usage stop is 100 percent of the week. The figure read 91 percent on 2026-11-17, and
  the reset is Nov 20, 10:59pm Alaska time.

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

Run the Sessions round and the Page round side by side, one pass of each at a time. Geoff
starts a fresh Opus session in his terminal, from `~/Projects/cc-desk`, with the line below.
Starting it is his approval of both designs. A change he wants goes in his first message.

```text
claude --model claude-opus-5-5 --effort high "$(cat docs/plans/2026-11-17-sessions-page-conductor-prompt.txt)"
```

Two steps are his, and neither holds the first eight passes. The conductor shows him an
auto-mode entry to approve (its text is in `docs/plans/2026-11-17-sessions-round.md`). He
accepts Claude Code's trust prompt once in `~/Projects/cc-desk-scratch`, which is cloned.

## Open decisions

- The drafts of the stopped dubplate design sessions are kept in
  `~/.local/state/desk/kept-drafts/`. The Retire round removes that folder, so the Move round must
  give them a home in the dubplate repo first.
