# Move: the round plan

This file is what `pass-core` asks a round of several passes to carry on top of its design. The
design is [2026-11-19-move-design.md](../specs/2026-11-19-move-design.md). Where the two differ,
the design wins. A planner writes each pass's plan just before the pass runs.

- Initiative: `cc-desk`
- Round id: `move`
- Repos: `/var/home/glw907/Projects/cc-desk` (push `glw907/cc-desk`, branch `main`),
  `/var/home/glw907/Projects/cairn-cms` (push `glw907/cairn-cms`, branch `main`), and
  `/var/home/glw907/Projects/dubplate` (push `glw907/dubplate`, branch `master`)
- Conductor: `conductor-cc-desk-move`, a background session on Opus at medium effort. The
  session that takes Geoff's approval launches it with the line under "Start line". The
  sessions cc-desk launches for the two releases use other names: the check's pass sessions
  are `cairn-cms-<pass>` and `dubplate-<pass>`, and its conductors `cairn-cms-conductor` and
  `dubplate-conductor`
- Conductor prompt: `docs/plans/2026-11-19-move-conductor-prompt.txt`
- Chain: 310 planned minutes on the longest path, 535 over all five passes. The round starts
  no Move pass past 1,070. cairn's round 1 runs outside this clock, under standing answer 45.

## Pre-flight

**Permissions checked:** I read `~/.local/state/claude-denials.jsonl` on 2026-11-20. It holds
145 rows, from 2026-11-01T00:55Z to 2026-11-20T06:38Z: 126 on Bash, 14 on Edit, 2 on Write,
and 1 each on Agent, ScheduleWakeup, and SendMessage. The round's sessions run these commands
that the log shows refused, or that the log has never seen. Each has its alternative.

| Command | Refusals in the log | Alternative |
| --- | --- | --- |
| `claude stop <id>` | 6 rows, from 2026-11-10 to 2026-11-18, as "[Interfere With Workloads]", "[Undermining Oversight]", and "[Destructive or Unwanted Actions]" | The conductor stops one session, `2d4cd0a4` (`cairn-cms-tool-delete`), run alone. The entry below names it. A refusal is standing answer 41's |
| `git push origin --delete <branch>` | 1 row, "[Git Destructive]" | `gh pr close 259 -R glw907/cairn-cms --delete-branch`, run alone, which the entry names |
| `git stash drop` | 10 rows, from 2026-11-17 to 2026-11-20, as "[Destructive or Unwanted Actions]" | Nobody drops the stash `foundation3-task3-wip`. The report lists it |
| `kill` or `pkill` of a process | 10 rows, most as "[Interfere With Workloads]" | No Move session kills a process. A stuck cairn or dubplate session ends through `cc-desk session end pass <project> <name>`, pass 6's verb |
| `gh pr merge` | 1 row, "[Merge Without Review]", and `claude-merge-guard` denies it for every session | No session runs it. Move's pull requests land by `approval-queue land`. The timed check merges cairn and dubplate pull requests, as the entry names |
| An edit to a Claude settings file | 5 of the 16 Edit and Write rows were on a `settings.json`, and 11 rows in all read "[Self-Modification]" | No pass makes one. The designing session adds the entry below after Geoff's approval. The conductor writes only `~/.local/state/cc-desk/settings.json`, which the entry names |
| `git worktree remove` and `git branch -D` in cairn-cms | None in the log for this use | Run alone, on `foundation-7` only, once pass 3 has kept its plan. The entry names them |
| `approval-queue withdraw` and `approval-queue land` | None in the log | Run alone. The entry names both |
| `cc-desk release add`, `release checkout`, `rules add` on the real state folder | None in the log | Run alone by the conductor, through the installed binary. The entry names them |
| Any command, in a classifier outage | 52 rows read "Classifier unavailable" | Wait 60 seconds and run it once more |

No row refuses `go build`, `go test`, `bash scripts/gate.sh`, `npm ci`, a `git push` of a pass
branch, or `gh pr create`. The last three rounds ran all of them in cc-desk. cairn's and
dubplate's gates ran in their repos through November with no refusal of the gate itself.

**Reviewed:** filled in after the `standards:plan-reviewer` read of the design and this plan.

**Evidence:** the shared assumptions and the launch checks below.

**Steps for Geoff:** two, both in the design under "Steps only Geoff can take".

1. `[you]` Approve the design, this plan, and the auto mode entry below. The designing session
   adds the entry before it launches the conductor, and `claude auto-mode` then shows it.
2. `[you]` One desk session once both releases are added, for the questions they carry. cairn's
   round 1 doesn't wait for it.

Claude Code already trusts the three folders the round's sessions open in: `~/.claude.json`
shows `hasTrustDialogAccepted` true for cc-desk, cairn-cms, and dubplate.

### The auto mode entry

One entry in `autoMode.allow` in `~/.claude/settings.json`, in the form of the last round's:

```text
cc-desk-move-round: under the approved cc-desk Move design
(docs/specs/2026-11-19-move-design.md in /var/home/glw907/Projects/cc-desk) and its round
plan, the sessions named `cc-desk-move-*` and the round's conductor `conductor-cc-desk-move`
may push one branch each and open one pull request each on glw907/cc-desk, glw907/cairn-cms,
and glw907/dubplate. The conductor may land those pull requests with `approval-queue land`. It
may withdraw the queue document `20261117-230223-743a55`, stop the one session
`cairn-cms-tool-delete` (id 2d4cd0a4) with `claude stop`, and land cairn-cms pull request 260
at commit 1fa506cc2 with `approval-queue land`. It may close cairn-cms pull request 259 and
delete its branch `foundation-ci-edits`, and remove the cairn-cms worktree `foundation-7` and
its branch once that worktree's plan is on main. Through the installed `cc-desk` binary it may
add the releases in docs/releases/cairn-cms.json and docs/releases/dubplate.json at main to
~/.local/state/cc-desk, record each release's checkout as its repo's main checkout, add an
allow rule to either release that names one command a pass's steps run, write
`usage_stop_percent` in ~/.local/state/cc-desk/settings.json, end a session of either release
with `cc-desk session end`, and reinstall the binary and the user units from that repo. For
the projects `cairn-cms` and `dubplate`, the timed check may launch pass sessions, conductors,
and design writers in those two repos, which may push their branches and open pull requests
there, and the check may merge such a pull request after every check has passed on its last
commit. None of them acts on a session cc-desk did not launch other than the one named above,
drops a stash, removes an `agent-*` worktree, records a ruling, runs `claude respawn --all` or
`claude daemon`, or launches a session with `bypassPermissions` or
`--dangerously-skip-permissions`.
```

## The passes

Each pass's acceptance list is the named items of the design's "What done looks like", which
its planner copies word for word. Every pass also carries item 7, the gate, and writes its pull
request body and its STATUS lines by `~/.claude/docs/voice/written-for-people.md`.

| Pass | Slug | Outcome | Acceptance | Files | Class | Minutes | Depends on | Shares a file with |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | `move-prompt` | The pass prompt names the round's design and context, and the report conductor checks earlier answers | Item 1 | `internal/launch/templates/pass.txt`, `internal/launch/templates/conductor.txt`, the template data in `internal/launch/` and its tests, `cmd/cc-desk/launcher.go`, `docs/follow-ups/earlier-answers-after-a-round.md` (removed) | `runner` | 60 | Nothing | 5 |
| 2 | `move-releases` | The two release input files and the test that loads them | Item 2 | `docs/releases/cairn-cms.json`, `docs/releases/dubplate.json`, `e2e/release_inputs_test.go` (all new) | `ordinary` | 120 | Nothing | none |
| 3 | `move-cairn-docs` | The route, its 45 prompts, and the regime map are in cairn-cms | Item 3 | In cairn-cms: `docs/internal/plans/route/**` (new), `docs/internal/plans/2026-11-17-foundation-7-routing-plan.md` (new), `docs/STATUS.md` (the foundation block) | `ordinary` | 60 | #260 landed | #260, on cairn's `docs/STATUS.md` |
| 4 | `move-dubplate-docs` | dubplate's three designs, their prompts, and the kept drafts are in dubplate | Item 3 | In dubplate: `docs/specs/2026-11-17-*.md` (three new), `docs/plans/2026-11-17-intake-*`, `docs/plans/2026-11-17-structure/**`, `docs/archive/kept-drafts/**` (all new) | `ordinary` | 45 | Nothing | none |
| 5 | `move-writers` | Writers start on the timer, their pull requests merge through the check, and each project gets a slot | Item 4 | `internal/sessions/decide.go`, `internal/sessions/merge.go`, `cmd/cc-desk/merge.go`, `cmd/cc-desk/round.go`, `cmd/cc-desk/launcher.go`, `internal/state/{events,types,check,rules,commit}.go`, `internal/launch/templates/design.txt`, `skills/cc-desk-work/SKILL.md`, `skills/cc-desk/SKILL.md` (the `round design` line), `README.md`, `scripts/live/` (a new writer script), their tests and fixtures, and the three folded follow-ups (removed) | `runner`, `live-account` | 250 | 1 | 1 |

Pass 1 changes only the templates and the data the launcher hands them, and it lands before
pass 5 starts, so pass 5 reads its launcher change from `main` (standing answer 15).

Pass 1 is `runner`: it changes what every pass session is told. Pass 5 is `runner` because it
changes the check that starts sessions and merges, and `live-account` because its live run
pushes to and merges in `glw907/cc-desk-scratch`. Passes 2, 3, and 4 change no code that runs:
pass 2's files are data and a test, and passes 3 and 4 copy documents. No pass is
`auth-data`.

Pass 2 can't meet item 5 by itself, since that item is the conductor's `release add` on the
real state folder. Its acceptance for item 5 is the test's snapshot. Item 6 is the conductor's,
after every pass has landed.

Pass 5 writes the regime map's one line about writers only if cairn needs one, which it
doesn't: every cairn round arrives designed.

## Shared assumptions

| Assumption | Passes | Probe or source |
| --- | --- | --- |
| The page refuses a design path outside the project's checkout | 2, 3, 4 | `web/src/lib/server/design.ts:58-60`: `realpath` of the checkout and of the file, then `within` |
| The check starts nothing for a release with no checkout | conductor | `internal/sessions/decide.go:321`: `if rf.Done \|\| rf.Checkout == ""` skips the release |
| `release add` takes passes not started, questions only with a carried answer or none, and refuses a second live release for one project | 2 | `internal/state/events.go:39`, `AddRelease`. cc-desk's live release blocks no Move pass: Move runs under its conductor, not as a release |
| A question must name a pass that exists | 2 | `internal/state/question.go:44` (`DecodeQuestion` needs `passes`) and `:78-80` (`AddQuestion` checks each) |
| A pass waits on its round's `after` rounds, and an undesigned round blocks the rounds after it | 2, 5 | `internal/state/rules.go:76`, `eachRoundAfter` |
| `round design` refuses a round already written, and its file needs format, passes, context, and calls | 5 | `cmd/cc-desk/round.go`, `DecodeDesign`, and `DesignRound` at `internal/state/events.go:110` |
| A design writer's prompt tells it to add a question on a pass of its round, which has no pass yet, so the line always fails | 5 | `internal/launch/templates/design.txt:15` and `internal/state/question.go:78` |
| The merge waits with `no_check` until a check registers, and stops on a failed check | 5, conductor | `internal/sessions/merge.go:148-151`, and the reasons `checks_failed` and `merge_failed` |
| A pull request closed without a merge clears its pass's wait and leaves the pass building | 5 | `internal/sessions/merge.go:125` |
| Release order in `startFree` is name order, so `cairn-cms` comes before `dubplate` | 5 | `internal/sessions/decide.go:318`, which walks releases in the order `os.ReadDir` returns their files (`internal/state/read.go:125`) |
| `cc-desk` reads one usage figure from `settings.json`, and the file is absent, so the stop is 100 | conductor | `internal/state/read.go:224-234` and `DefaultUsageStop`. `types.go:512` says the file is edited by hand |
| cairn-cms is on `main` at `79d326bed`, and its pull request 260 is open at `1fa506cc2` with every check passed | 3, conductor | `gh pr view 260` on 2026-11-20: `OPEN`, `MERGEABLE`, three checks `SUCCESS` |
| cairn-cms pull request 259 is open on branch `foundation-ci-edits` | conductor | `gh pr view 259` on 2026-11-20 |
| Both repos are public, cairn-cms on `main` and dubplate on `master` | 2, 3, 4 | `gh repo view` on 2026-11-20: `false main` and `false master` |
| The route and its 45 prompts are committed in the initiatives repo at `f461b3f`, and that commit is in the local checkout | 3 | `git -C ~/Projects/initiatives ls-tree -r f461b3f theme-garden/plans` lists 45 `route-*-planner-prompt.txt` files |
| dubplate's three designs and their prompts are untracked files in the initiatives checkout | 4 | `git status` there lists `dubplate-server/specs/` and `dubplate-server/plans/` as untracked. The pass copies from the working tree |
| The kept structure and intake drafts are byte-identical to the initiatives copies | 4 | `cmp` on 2026-11-20, during the design |
| cairn-cms and dubplate ignore `.claude/worktrees/` | 3, 4 | `git check-ignore -q .claude/worktrees/x` exits 0 in both |
| cairn's `docs/internal/**` is outside its published-docs checks | 3 | `scripts/checks/check-stable-source.mjs:83` |
| Measured Desk passes took 1.44 times their planned minutes | all | Passes 1 to 5 planned 630 and took about 906 (`~/Projects/initiatives/cc-desk/reports/desk-report.md`) |

## Order table

Two passes build at once, and never more. Pass 1 and pass 2 start together.

| Pass | Planner prompt | Depends on | Planned minutes | Waits for Geoff |
| --- | --- | --- | --- | --- |
| 1 `move-prompt` | `2026-11-19-move-1-planner-prompt.txt` | Nothing | 60 | no |
| 2 `move-releases` | `2026-11-19-move-2-planner-prompt.txt` | Nothing | 120 | no |
| 3 `move-cairn-docs` | `2026-11-19-move-3-planner-prompt.txt` | #260 landed (order: after 2) | 60 | no |
| 4 `move-dubplate-docs` | `2026-11-19-move-4-planner-prompt.txt` | Nothing (order: after 3) | 45 | no |
| 5 `move-writers` | `2026-11-19-move-5-planner-prompt.txt` | 1 | 250 | no |

Slot A runs pass 1 and then pass 5, 310 minutes. Slot B runs passes 2, 3, and 4 in that order,
225 minutes, so slot B frees first. The `(order)` marks are slot order, not a file dependency,
so a parked pass 2 doesn't hold pass 3 (`round-run.md` standing answer 5).

Before each launch the conductor reads `claude agents --json` into a file and reads the file.
A 1.1 session there whose pass shares a file with the Move pass due next holds that pass until
the 1.1 pass lands (standing answer 40). Only passes 1 and 5 touch files a 1.1 pass might.

The conductor's steps follow the passes, as the design's "The run" lists them:

0. Before any launch: the first-step test (standing answer 44), then the land of #260 on
   Question 1. Pass 3 waits for that land, since #260 edits cairn's `docs/STATUS.md` beside
   the block pass 3 replaces.
1. After passes 2, 3, and 4 have landed: `release add` for both files, with no checkout.
2. The cairn leftovers: #259 and its branch, the `foundation-7` worktree and branch.
3. After pass 5 has landed and is installed: the usage figure, then `release checkout` for
   both projects.
4. The watch of cairn's round 1 to its report.
5. The Move report, STATUS, and the end.

## Usage ceiling

The figure Geoff picks in the design's Question 2, 85 percent on the recommendation. It stops
the Move conductor's launches as it stops the check's. The figure read 15 percent on
2026-11-20, with the reset at Nov 20, 10:59pm Alaska time. At the ceiling the conductor follows
standing answer 4 as the Desk plan replaced it: the passes in flight finish, nothing new
starts, and the figure goes under "Open decisions".

## Standing answers

The twelve in `round-run.md` hold, as the Desk round plan changed them (answers 1, 4, 6, and
9), with its answers 13 to 39. Answers 15, 17, 18, 19, 26, 27, 28, 33, 34, and 38 bear on this
round most. These are new.

**40. A 1.1 pass holds a file a Move pass needs.** The Move pass waits for the 1.1 pass to
land, and the conductor runs every step that doesn't rest on it. For pass 1, the wait has no
limit. For pass 5, once every other step through step 2 is done and the 1.1 pass still holds
the file, the conductor doesn't wait. It runs step 3 without the writers. Under "Open
decisions" it writes that pass 5 goes to 1.1 with its planner prompt, and that dubplate's
`catalog` round needs a writer started from the desk.

**41. Auto mode refuses a step the entry names.** The conductor runs the step once more, with
a line that cites the entry and this answer. A second refusal goes under "Open decisions", and
the conductor goes on with every step that doesn't rest on it. A refused land of #260 holds
`release checkout` for cairn-cms, and pass 3 then runs with cairn's STATUS left alone, as a No
to Question 1 has it. A refused cleanup of #259 or `foundation-7` holds
nothing.

**42. A cairn or dubplate session stops at a permission prompt.** This is answer 18 on a real
release. The conductor reads the prompt's command from `cc-desk snapshot`. If the command is
one the pass's steps run and the release's rules should allow, the conductor adds the rule
with `cc-desk rules add <project> <rule>`, files the gap as a follow-up, and ends the session
with `cc-desk session end pass <project> <name>`, so the check relaunches it under the new rule. A command the pass
shouldn't run, or a third prompt in one pass, parks the pass under "Open decisions". The
conductor never answers a prompt in Geoff's name.

**43. The busy check names a 1.1 session.** In the cc-desk repo, a hit that names a 1.1
session whose pass shares no file with the Move pass due next isn't a stop (answer 6, changed
the same way for the Move design session). A shared file is answer 40's.

**44. The first step finds Geoff was asked for more than the sitting's five steps.** The
conductor launches nothing. It writes what asked him, from which record, under "Open
decisions" in `docs/STATUS.md`, and ends.

**45. cairn's round 1 runs past twice its planned minutes.** This is cairn ruling 9, which
cc-desk has no field for. The conductor writes the overrun under "Open decisions" with each
session's first forecast, and the report names it. It ends no session for it, and the check
carries on with the pass in flight.

**46. CI stops running mid-round.** A GitHub Actions block, such as the billing block of
2026-11-20, leaves a pull request with no check, and the merge waits with `no_check`. The
conductor lands nothing without its checks. It writes the block under "Open decisions" with the
first run that failed and the message GitHub gave, and goes on with every step that rests on no
merge. The check's merges in cairn and dubplate wait the same way, which needs nothing from the
conductor. Once a run passes again, the conductor reads each waiting pull request afresh.

### Every stop and its answer

| Stop | Answer |
| --- | --- |
| A question `advisor.md` "What always stops" lists | 1, changed |
| Auto mode refuses a command a pass runs | 2 and 19; a step the entry names is 41 |
| A setting, hook, or rule must change | 3; a missing allow rule on a release is 42 |
| Usage at the ceiling | 4, replaced |
| A parked pass's dependents | 5 |
| The busy check exits 1 | 6, changed; a 1.1 session is 43 |
| The round passes twice its plan | 7; cairn's round 1 is 45 |
| A plan can't go ahead as designed, or runs over its estimate | 8, 14, 28 |
| Geoff answers while the round runs | 9, changed |
| The conductor dies or the machine restarts | 10 |
| A session goes idle unfinished | 11 |
| An acceptance item can't be met as written | 12 |
| A later pass changes an earlier pass's file | 15 |
| A live run meets a permission prompt | 18; on a real release, 42 |
| A review returns fix or escalate | 26 |
| A step after a pass lands is refused | 27, for its rule; 41, for this round's steps |
| A leftover scratch session | 33 |
| A 1.1 pass holds a file | 40 |
| The first step finds the sitting failed | 44 |
| GitHub Actions runs no job | 46; at the start, the start line's state 7 |

## Start line

The session that takes Geoff's approval launches the conductor from the main checkout, with
this line run alone, once the seven states read true:

```text
claude --bg --model claude-opus-5-5 --effort medium --name conductor-cc-desk-move "$(cat /var/home/glw907/Projects/cc-desk/docs/plans/2026-11-19-move-conductor-prompt.txt)"
```

The conductor checks the seven again before it launches the first planner.

1. `docs/STATUS.md` on `origin/main` says pass 6 has landed, that Geoff's sitting passed done
   item 10, and that he approved this design. It names the conductor prompt as the next action.
2. `claude auto-mode` shows the entry `cc-desk-move-round`.
3. `claude agents --json`, redirected to a file, lists no other session named
   `conductor-cc-desk-move` and none named `cc-desk-move-*`.
4. The main checkout is on `main`, and `git rev-list --count HEAD..origin/main` prints 0 after
   a fetch.
5. `claude-worktree-busy` exits 0 for the cc-desk, cairn-cms, and dubplate checkouts, or its
   hit is one standing answer 43 or 6 covers.
6. The weekly usage figure is under the ceiling.
7. GitHub Actions runs again: `gh run list -R glw907/cc-desk -L 1` shows a run created after
   the billing block was cleared, with a conclusion that isn't a billing failure.

### Launch checks, run on 2026-11-20

| Check | Output |
| --- | --- |
| `~/.local/share/cc-desk/bin/cc-desk usage` | `15% used, resets Nov 20, 10:59pm (America/Anchorage)` |
| `claude agents --json`, to a file | `cairn-cms-tool-delete` (blocked), `conductor-cc-desk`, `cc-desk-desk-run-plan` (busy), and this session. No 1.1 session |
| `gh pr view 260 -R glw907/cairn-cms` | `OPEN` at `1fa506cc2`, `MERGEABLE`, three checks `SUCCESS` |
| `gh pr view 259 -R glw907/cairn-cms` | `OPEN`, branch `foundation-ci-edits` |
| `git -C ~/Projects/cairn-cms worktree list` | `foundation-7` at `79d326bed`, with its plan untracked; 110 `agent-*` worktrees |
| `git -C ~/Projects/cairn-cms stash list` | `stash@{0}: On worktree-agent-a373c1844f2b29bf4: foundation3-task3-wip` |
| `cc-desk snapshot` | One release, `cc-desk-scratch`, from pass 6's rehearsal. No cairn-cms or dubplate release |
| `git check-ignore -q .claude/worktrees/x` in cairn-cms and dubplate | Exit 0 in both |
| `gh run list -L 3` on each repo, at 09:40 UTC | cc-desk: three `gate` runs `failure` from 09:25 UTC, the billing block. cairn-cms: last run `success` on 2026-11-19. dubplate: last run `success` on 2026-11-17. cc-desk-scratch: last run `success` at 06:39 UTC |
| `git -C .claude/worktrees/desk-run log origin/main..HEAD` | Pass 6 at `f2b0c18`, seven commits, draft pull request 22. `release done` is built. `session end` is its plan's Task 2 |

## Decisions

- Five passes against the route's two files and a round. The design gives the reasons: the
  papers must sit inside each repo for the page to show them, and the writers and the prompt
  change are this round's by the Desk design.
- Move runs under its own conductor in the old way, since cc-desk holds one live release per
  project and 1.1 shares the project `cc-desk`.
- The conductor is a background session on Opus at medium effort. It lands a pull request from
  the old regime, adds two real releases, and watches a real round, and each takes judgment.
- Executors: pass 1 and pass 5 on Opus at medium, since pass 1 writes what every session is
  told and pass 5 changes the check's merge. Passes 2, 3, and 4 on Sonnet at medium.
- No release gets a checkout until pass 5 has landed and is installed. The check can't count a
  Move pass against its two slots, so this order is what keeps three passes from building.
- The first planner launch and every relaunch follow the Desk round's lines, with
  `conductor-cc-desk-move` as the one session whose messages a Move session takes as
  instructions.
