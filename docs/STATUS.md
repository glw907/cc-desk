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

The core design (`docs/specs/2026-11-20-core-design.md`) is being built in the dotfiles
worktree `~/.dotfiles/.claude/worktrees/cc-desk-core`, on branch `cc-desk-core`. Pull request
254 in `glw907/workstation` landed the build's auto-mode entry, and the entry is live. The
settings edits are commit 8a24900 on the branch, made by the session `cc-desk-core-spec`.

A fresh background session, `cc-desk-core-build`, builds the rest there under `/goal`. Its
launch prompt is `docs/plans/2026-11-20-core-build-prompt.txt`. Two subagents review the
result, and the build session opens one pull request, which deletes the entry. `cc-desk-core-spec`
reports that pull request in Geoff's terminal, he approves it by number, and that session merges
it with `approval-queue land` and runs steps 4 to 7 of the design's "How it lands".

After the core lands, the session brainstorms the process spec from
`docs/specs/2026-11-20-process-decisions.md`, then the other-skills spec. Pull request 253 in
the dotfiles repo is still open. It adds the sentence to the global `CLAUDE.md` that exempts
cc-desk's spec work from the pass process.

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
