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

Finish the Prove plan's pre-flight, then launch the executor from the main checkout:

```
claude --bg --model sonnet --effort medium --advisor fable --name cc-desk-prove \
  "$(cat docs/plans/2026-11-17-prove-executor-prompt.txt)"
```

Pre-flight has two steps only Geoff can do, both in the plan under "Steps for Geoff": accept the
workspace trust prompt in this folder, and approve the text of the `cc-desk-prove-owned:` auto mode
entry. The launching session then runs the plan's checks 3 to 7, which include creating the
private scratch repo `glw907/cc-desk-scratch`.

The later rounds are designed after the Prove round's findings report lands.

## Open decisions

- The drafts of the stopped dubplate design sessions are kept in
  `~/.local/state/desk/kept-drafts/`. The Retire round removes that folder, so the Move round must
  give them a home in the dubplate repo first.
