# Status

## Current state

- The design (`docs/specs/2026-11-17-cc-desk-design.md`) and the route
  (`docs/plans/2026-11-17-route.md`, seven rounds) are approved.
- Prove landed as pull request 1, with open tests in `2026-11-17-rerun-prove-proofs.md`. State
  and rules landed as 3 to 6; two of its follow-ups need Geoff's choice when taken up:
  `orphan-question-forecast.md` and `reading-errors-and-guess-on-resume.md`.
- The Sessions and Page rounds landed on 2026-11-18 as pull requests 7 to 16. The page serves
  main's build at http://127.0.0.1:7908/, and the check's timer runs main's binary,
  `~/.local/share/cc-desk/bin/cc-desk`, every five minutes. Follow-ups are files in
  `docs/follow-ups/`, with no `BACKLOG.md`. The usage stop is 100 percent of the week.
- The Desk and skills round (`desk`) landed all six passes as pull requests 17 to 22 (design
  `docs/specs/2026-11-18-desk-skills-design.md`, round plan
  `docs/plans/2026-11-18-desk-skills-round.md`), with the skills pinned at `v0.1.0`. Its report
  is `~/Projects/initiatives/cc-desk/reports/desk-report.md`. Each stop in it has a row in
  `docs/follow-ups/desk-round-lessons.md`, which the 1.1 round answers (standing answer 34).

- Geoff's sitting of 2026-11-20 finished the desk round. One answer launched the
  `stuck-sessions` pass, and the timed check merged its pull request 24 with no approve from
  him, then started a conductor that recorded the round's report, which he has read. The desk
  now lays out every question with each answer on its own line (pull request 25, skill tag
  `v0.1.1`). The terminal link's "Always allow" held in Chromium, and pull request 26 lets its
  handler find `cc-desk` from the desktop's `PATH`.

## Next action

A fresh brainstorm for 1.1 runs in Geoff's terminal. The first design session wrote a draft
(`docs/specs/2026-11-20-1.1-design.md`), and Geoff set it aside on 2026-11-20: the Go already
built holds process that changes with `pass-core` and Claude Code, and the draft tested only
its additions. Nothing in the draft was built. Geoff picked sidebar candidate A
(`docs/design-benchmark/sidebar-headings/`), and that choice stands.

```bash
claude --model claude-fable-5-2 --effort high --name cc-desk-1.1-brainstorm "$(cat /var/home/glw907/Projects/cc-desk/docs/plans/2026-11-20-1.1-brainstorm-prompt.txt)"
```

## Open decisions

- How much of cc-desk stays in Go is the fresh brainstorm's question (Geoff, 2026-11-20: "the
  logic that lives at the Go level needs to not be in our way as best practices evolve").
- The Move design (`docs/specs/2026-11-19-move-design.md`) needs a rewrite before Geoff
  approves it. dubplate's `structure` round and cairn's `checks` round run under their own
  conductors for now (Geoff, 2026-11-20: "We'll bring both other [initiatives] under the cc-desk
  regime once cc-desk is ready"), so Move's first act and Questions 1 and 4 must match what
  those rounds leave behind. cairn #260 and #259 have landed.
- The conductors of dubplate and cairn-cms share one rule with cc-desk's sessions (agreed
  2026-11-20): at most four planners or executors on pavilion, no new session past 85 percent
  weekly usage, and a message to the others before each launch.
