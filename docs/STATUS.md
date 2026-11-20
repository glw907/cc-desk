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

<!-- round-report desk -->
**The desk round has landed all six passes, and Geoff's sitting is next.** The round builds the
desk session that `cc-desk` opens, the two skills, the page's report pages, and the check's
repairs, then rehearses one run with nobody tending it. In the scratch repo, one typed answer
led to a merged pull request and a report with nothing else typed, and the recovery checks
passed on the live page. The real release is set up, and the page shows one question waiting
for cc-desk. Geoff's sitting finishes the round.
<!-- /round-report -->

## Next action

Geoff's sitting, with the steps below; the timer launches the `stuck-sessions` pass once he
answers. After the sitting, the 1.1 design session runs in his terminal:

```bash
claude --model claude-opus-5-5 --effort high --name cc-desk-1.1-design "$(cat /var/home/glw907/Projects/cc-desk/docs/plans/2026-11-19-1.1-design-prompt.txt)"
```

## Open decisions

- Geoff's sitting. The page shows one question waiting for cc-desk.
  1. Type `cc-desk`. The desk reads out the small design about stuck sessions and its
     question. Answer in your words.
  2. Look at the page and confirm the ruling shown is what you said.
  3. Walk away. The build starts, merges, and cleans up after itself. This is the first time
     the timed check merges into a real repo with no approve from you, as your ruling of
     2026-11-17 allows for a change inside its approved design.
  4. When the report appears under "Reports to read", read it and click "Mark as read".
  5. Click the terminal link once in Chromium, the click still owed from the Page round. What
     to look for is in `docs/follow-ups/page-link-chromium-click.md`.
  If anything else asks for you during step 3, the run has failed its test.
- The Move design (`docs/specs/2026-11-19-move-design.md`) needs a rewrite before Geoff
  approves it. dubplate's `structure` round and cairn's `checks` round run under their own
  conductors for now (Geoff, 2026-11-20: "We'll bring both other [initiatives] under the cc-desk
  regime once cc-desk is ready"), so Move's first act and Questions 1 and 4 must match what
  those rounds leave behind. cairn #260 and #259 have landed.
- The conductors of dubplate and cairn-cms share one rule with cc-desk's sessions (agreed
  2026-11-20): at most four planners or executors on pavilion, no new session past 85 percent
  weekly usage, and a message to the others before each launch.
