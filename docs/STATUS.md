# Status

## Current state

- The design was approved on 2026-11-17 and amended the same day for three later rulings: the old
  system is removed in full, no background session may need a permission granted by hand, and the
  page gets a link that opens a terminal on the desk. It is
  `docs/specs/2026-11-17-cc-desk-design.md`.
- The route is `docs/plans/2026-11-17-route.md`. It has seven rounds: Prove, State and rules,
  Sessions, Page, Desk and skills, Move, and Retire.
- The first round's plan is `docs/plans/2026-11-17-prove-plan.md`, reviewed once and folded. Its
  path estimate is 530 minutes.
- The repo holds the design, the page mockup, and the plans. No code exists.
- Until cc-desk runs, `pass-core` "Until cc-desk runs" says how a question, a report, or a pull
  request reaches Geoff. Nothing uses the Approval Queue's documents.

## Next action

The Prove pass is running as the background session `cc-desk-prove`, launched on 2026-11-17 from
the main checkout with `docs/plans/2026-11-17-prove-executor-prompt.txt`. It closes the pass
itself and rewrites this file at close.

Pre-flight finished on 2026-11-17. Geoff accepted the workspace trust prompt and approved the
`cc-desk-prove-owned:` auto mode entry, which landed in the dotfiles as `0a41854`. Checks 3 to 7
passed, and the private scratch repo `glw907/cc-desk-scratch` exists with its gate workflow.

The later rounds are designed after the Prove round's findings report lands.

## Open decisions

- The drafts of the stopped dubplate design sessions are kept in
  `~/.local/state/desk/kept-drafts/`. The Retire round removes that folder, so the Move round must
  give them a home in the dubplate repo first.
