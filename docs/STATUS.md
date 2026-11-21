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
- Geoff's sitting of 2026-11-20 finished the desk round, and he has read its report. The
  `stuck-sessions` pass landed as pull request 24, the desk's question layout as 25 (skill tag
  `v0.1.1`), and the terminal link's `PATH` fix as 26.

## Next action

The core design (`docs/specs/2026-11-20-core-design.md`) is live. Geoff approved pull request
255 in `glw907/workstation` on 2026-11-20 and it is merged. Steps 4 to 7 of the design's "How
it lands" are done: the settings are applied, every item under "What done looks like" holds,
and both auto-mode outputs are in a comment on the pull request. What the reviews left is
items 101 to 108 in the dotfiles `BACKLOG.md`; item 108 is the look at one week of refusals,
due 2026-11-27. Pull request 253 is closed unmerged, because 255 carries its sentence.

The brainstorm of the process spec is under way. Its documents live in the initiatives repo,
which Geoff raised on 2026-11-20 because the work crosses every repo. The decisions record is
`~/Projects/initiatives/cc-desk/specs/2026-11-20-process-decisions.md`, with nine research
reports in `../research/2026-11-20-process/`, and both specs get written beside it. Each
ruling goes into the record as he gives it. He took agent view's four group names as they are.

He ruled that the workstation comes first: its Claude configuration reaches the new standard
in full, and only then does work on the page begin. The process pieces land first, then the
workstation's other skills, agents, and docs. The page, the removal of the Go, and the repos
wait behind them, and his sitting opens in agent view meanwhile.

He took the `ask` rule on `gh pr merge`, with one confirmation per pull request, provided it
follows Anthropic's standards. Background sessions run in `dontAsk` mode, and
`approval-queue land` retires when the rule lands. The budget caps are the next question. The
rest are in `docs/plans/2026-11-20-process-spec-prompt.txt`, where a fresh session picks up.

## Open decisions

- Where the 1.2 work (scheduled check-ins) comes against the rewrite of Move is asked once, at
  the end of the design work.
- The Move design (`docs/specs/2026-11-19-move-design.md`) is rewritten after the process spec,
  which is the bar cairn and dubplate join cc-desk at. Geoff said on 2026-11-20 that both
  projects' current work is complete: dubplate's `structure` round is paused at his request,
  and cairn's `checks` round is closing on pull request 262. Both hold new work, and he
  reshapes them once the new process is in place, so the process spec replaces the old process
  outright and keeps no second set of rules alive.
- The conductors of dubplate and cairn-cms share one rule with cc-desk's sessions (agreed
  2026-11-20): at most four planners or executors on pavilion, no new session past 85 percent
  weekly usage, and a message to the others before each launch.
