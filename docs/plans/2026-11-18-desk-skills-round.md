# Desk and skills: the round plan

This file is what `pass-core` asks a round of several passes to carry on top of its design. The
design is [2026-11-18-desk-skills-design.md](../specs/2026-11-18-desk-skills-design.md). Where
the two differ, the design wins. A planner writes each pass's plan just before the pass runs.

- Initiative: `cc-desk`
- Round id: `desk`
- Repo: `/var/home/glw907/Projects/cc-desk`, pushing to `glw907/cc-desk`, default branch `main`
- Conductor: `conductor-cc-desk`, a background session on Opus at medium effort. The session
  that takes Geoff's approval launches it with the line under "Start line". The sessions
  cc-desk launches for itself use other names: its conductor is `cc-desk-conductor`
- Conductor prompt: `docs/plans/2026-11-18-desk-conductor-prompt.txt`
- Chain: 485 planned minutes on the longest path, 575 over all six passes. The round starts
  nothing new past 970.

## Pre-flight

**Permissions checked:** I read `~/.local/state/claude-denials.jsonl` on 2026-11-18 for
2026-10-19 to 2026-11-18. The log holds 133 rows, from 2026-11-01T00:55Z to 2026-11-18T19:38Z:
114 on Bash, 14 on Edit, 2 on Write, and 1 each on Agent, ScheduleWakeup, and SendMessage. The
round's sessions run these commands that the log shows refused, or that the log has never
seen. Each has its alternative.

| Command | Refusals in the log | Alternative |
| --- | --- | --- |
| `kill` or `pkill` of a running process | 9 rows, one of them a `kill` of one process id run alone | The recovery checks of passes 5 and 6 end only a session the `cc-desk` binary launched and recorded, by the process id on its record, with `kill` run alone and never `pkill -f`. The entry below covers it. A refusal parks the pass |
| `claude stop <id>` chained with another command | 5 rows | Run it alone. In the live runs the `cc-desk` binary runs it |
| `git worktree add` or `remove` chained | 2 rows | Run alone |
| `git push origin HEAD:main` chained after a commit | 3 rows | No pass pushes to `main`. The conductor's hand-over push runs alone |
| `claude --bg` from a session | 1 row, in a classifier outage | The planner's launch line, run alone |
| `claude agents --json` piped into a script | 1 row | Redirect to a file and read the file |
| `gh pr merge` | 1 row, and `claude-merge-guard` denies it for every session | No session runs it. The round's own pull requests land by `approval-queue land`. In the live runs the timed check merges |
| An edit to a Claude settings file | 10 rows | No pass makes one. The settings file a launch carries is written by the `cc-desk` binary into the state folder. The designing session adds the entry below after Geoff's approval |
| `stow -R claude` in the dotfiles repo | 1 row, refused as "[Self-Modification]" | `claude-tooling-sync update-skills` runs it when the conductor installs the two skills, and the entry below names that install. A refusal goes under "Open decisions" and holds pass 6 |
| `claude -p` with a skill and a brief, for pass 2's fixture checks | None in the log | A refusal is standing answer 2's |
| Any command, in a classifier outage | 50 rows | Wait 60 seconds and run it once more |
| `git tag`, `git push origin v0.1.0`, `systemctl --user restart`, `script -qfc`, `ln -s` into `~/.local/bin` | None in the log | Covered by the entry below, or run by this session today with no refusal (`script -qfc`). A refusal is standing answer 2's |

No row refuses `go build`, `go test`, `go vet`, `gofmt`, `bash scripts/gate.sh`, `npm ci`, a
`git push` of a pass branch, or `gh pr create`. The last two rounds ran all of them here.

**Reviewed:** the design had its builder check on 2026-11-18, and every finding is folded or
listed in the design's "Calls Claude made". Geoff approved the design and this plan in his
terminal, as `docs/STATUS.md` records. Each pass's plan gets one `standards:plan-reviewer`
read, as `round.md` "A plan inside a round" asks of every class.

**Evidence:** the shared assumptions and the launch checks below.

**Steps for Geoff:** two, both in the design under "Steps only Geoff can do".

1. `[you]` Approve the auto mode entry below, with the design. The designing session adds it
   before it launches the conductor, and `claude auto-mode` then shows it.
2. `[you]` The sitting at the end of the round: five steps, which the conductor writes under
   "Open decisions" in `docs/STATUS.md` once pass 6 has landed and the real release is set up.

Claude Code already trusts both folders the round's sessions open in: `~/.claude.json` shows
`hasTrustDialogAccepted` true for `/var/home/glw907/Projects/cc-desk` and for
`/var/home/glw907/Projects/cc-desk-scratch`.

### The auto mode entry

One entry in `autoMode.allow` in `~/.claude/settings.json`, in the form of the last round's:

```text
cc-desk-desk-round: under the approved cc-desk Desk and skills design
(docs/specs/2026-11-18-desk-skills-design.md in /var/home/glw907/Projects/cc-desk) and its
round plan, the sessions `cc-desk-desk-open`, `cc-desk-desk-page`, `cc-desk-desk-check`, and
`cc-desk-desk-run`, and the round's conductor `conductor-cc-desk`, may run the `cc-desk` binary
built from that repo. Through that binary, and only on sessions it launched and recorded, they
may launch sessions named `cc-desk-scratch-*` with their working folder at or under
/var/home/glw907/Projects/cc-desk-scratch: background sessions with `--permission-mode auto`
and a settings file the binary writes, holding only allow rules that each name one command the
pass's steps run, `ask` rules, and hooks that record a permission request or a refusal and
return no decision, and desk sessions the binary opens on a pseudo-terminal the test makes. They may end
those sessions by the binary or by `kill` of the process id on the session's record, remove
them, and read their logs. The timed check may merge a pull request on the private repo
glw907/cc-desk-scratch after a check has passed on its last commit. `cc-desk-desk-page` and
the conductor may reinstall and restart the user units cc-desk-check.service,
cc-desk-check.timer, and cc-desk-page.service from that repo, with the page's unit given write
access to ~/.local/state/cc-desk and nothing else, and may link ~/.local/bin/cc-desk to the
installed binary. The conductor may tag the repo `v0.1.0` or a later patch tag and push that
tag. It may add the two skills `cc-desk` and `cc-desk-work` from that repo, at that tag, to the
workstation's vendored-skills manifest through a pull request on the dotfiles repo, and run
`claude-tooling-sync update-skills` for those two, which restows the claude package. It may
add one release for the project `cc-desk` to ~/.local/state/cc-desk with one
round and one pass under docs/specs/2026-11-18-stuck-sessions-design.md. For that one pass, and
after Geoff's ruling at the desk, the timed check may launch one pass session and one
conductor in /var/home/glw907/Projects/cc-desk, which may push one branch and open one pull
request on glw907/cc-desk, and the check may merge that one pull request after every check has
passed on its last commit. None of them acts on a session cc-desk did not launch, runs
`claude respawn --all` or `claude daemon`, records a ruling, or launches a session with
`bypassPermissions` or `--dangerously-skip-permissions`.
```

## The passes

Each pass's acceptance list is the named items of the design's "What done looks like", which
its planner copies word for word. Every pass also carries item 11, the gate, and writes its
pull request body and its STATUS lines by `~/.claude/docs/voice/written-for-people.md`.

| Pass | Slug | Outcome | Acceptance | Files | Class | Minutes | Depends on | Shares a file with |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | `desk-records` | Questions, rulings, reports, rounds, and stops carry what the desk and the page say, the check keeps a log, and one verb prints the brief | Item 1 | `internal/state/`, `internal/snapshot/`, `cmd/cc-desk/` (the new verbs), `internal/launch/` (the settings file's `ask` rules), the snapshot type under `web/src/lib/`, fixtures | `ordinary` | 120 | Nothing | 3, 4, 5, 6 |
| 2 | `desk-skills` | The two skills and the guide exist, and the launch prompts name the working skill | Item 2 | `skills/cc-desk/`, `skills/cc-desk-work/`, `docs/guide.md`, `internal/launch/templates/` and its tests, `scripts/gate.sh` and `.vale.ini` for the skills' lint | `runner` | 75 | 1 | 5, 6 |
| 3 | `desk-open` | `cc-desk` opens the session, the check starts free passes, a wrong record is replaced, and a merge question has answers | Items 3, 4, 5 | `cmd/cc-desk/` (the root command, the desk, ruling, and merge verbs), `internal/desk/`, `internal/state/`, `internal/sessions/` (the start of a free pass, the merge answers), `e2e/`, the README's install lines | `runner` | 100 | 1, 2 | 1, 5, 6 |
| 4 | `desk-page` | Reports, designs, "Mark as read", and the check and sessions in words | Item 6 | `web/`, `systemd/cc-desk-page.service`, `web/desktop/cc-desk-open` | `runner` | 90 | 1 | 1 |
| 5 | `desk-check` | Refusals are recorded, a model limit holds, the cleanup stops the session, a merge call is bounded, and the conductor starts for a report | Item 7 | `internal/sessions/`, `internal/launch/` (the settings file's hooks), `cmd/cc-desk/` (the check and the hook verb), `scripts/live-check.sh`, fixtures | `runner`, `live-account` | 100 | 2, 3 | 1, 2, 3, 6 |
| 6 | `desk-run` | The whole path runs in the scratch repo, the recovery checks pass on the live page, and the real release is ready to add | Items 8, 9, 10 | `scripts/` (the rehearsal and the release setup), `e2e/`, the scratch repo's fixture, whatever the rehearsal shows broken | `runner`, `live-account` | 90 | 1 to 5 | 1 to 5 |

Pass 1 fixes the names of the new verbs and the format of the brief. Passes 2, 3, and 4 read
both from `main` once pass 1 has landed, and none of them renames a verb.

Passes 2 to 6 are `runner`: they write what a session is told, the settings it launches with,
a unit, or the check that merges. Passes 5 and 6 are `live-account` because they push to and
merge in `glw907/cc-desk-scratch`. No pass is `auth-data`: the ruling guard was built and
reviewed in the State and rules round, and pass 3 calls it without changing what it checks. A
plan that finds it must change the guard's checks takes the `auth-data` reads.

Pass 6 can't meet item 10 by itself: that item is Geoff's sitting. Its acceptance for item 10
is that the release setup script, run against a copy of the state folder, yields a snapshot
with one question waiting for the project `cc-desk`, and that the rehearsal passed with the
allow rules that release carries. The evidence from Geoff's run is added to the round report
by the next planning session, as the design says.

Items 3, 4, and 5 need a live desk session on a pseudo-terminal, and items 7 to 9 need the
scratch repo. None is a Go test that skips. Each planner runs the checks that can run on the
code as the earlier passes left it, before any code is written.

## Shared assumptions

| Assumption | Passes | Probe or source |
| --- | --- | --- |
| A session started as `claude --session-id <uuid> "<first prompt>"` on a terminal is listed as `interactive` under that id and the launched process id | 3, 6 | Probe on pavilion, 2026-11-18, Claude Code 2.1.327, under `script -qfc`: `claude agents --json` gave a row with `kind: interactive`, the chosen id, and `pid` 698378, the launched process. Inside it `echo $CLAUDE_CODE_SESSION_ID $PPID` printed the same id and 698378 |
| A session started from inside a Claude session isn't listed unless the inherited marker is cleared | 3, 6 | Same probe: without `env -u CLAUDE_CODE_CHILD_SESSION -u CLAUDECODE -u CLAUDE_CODE_SESSION_ID -u CLAUDE_CODE_ENTRYPOINT` the session ran, printed that transcript saving is off under the inherited `CLAUDE_CODE_CHILD_SESSION` marker, and had no row. Every test that opens a desk clears the four |
| A first prompt on the command line shows on screen as a user message | 2, 3 | Same probe: the typescript shows the prompt above the reply. So the opener passes one short line and the brief travels another way |
| Claude Code asks no trust question in a project's main checkout it has run in, and asks one in a new folder | 3, 6 | Same probe: `~/.local/state/cc-desk` and a new scratchpad folder stopped at "Is this a project you created or one you trust?". `~/.claude.json` reads `hasTrustDialogAccepted: true` for the cc-desk and cc-desk-scratch checkouts |
| `claude --help` lists `--session-id`, `--append-system-prompt`, `--settings`, `--permission-mode`, `--model`, `--effort`, `--name`, and `--bg`, and no flag that loads a skill | 2, 3 | Run on 2026-11-18, Claude Code 2.1.327 |
| A `PermissionDenied` hook fires for an auto mode refusal and its input carries the tool, the target, and a reason | 5, 6 | The workstation's hook of that kind, `claude-denied-log`, wrote rows on 2026-11-18 with `tool`, `target`, and `reason` set, from sessions in auto mode. Not run: the hook in a per-session `--settings` file. Pass 5's first task runs one such launch in the scratch repo, on a command the log shows refused, and keeps the input as its fixture |
| In auto mode a command that an `ask` rule names stops the session at a permission prompt, and the `PermissionRequest` hook gets the command | 1, 5, 6 | `docs/research/2026-11-18-ask-rule-probe/`: with `Bash(date *)` as an `ask` rule, a background session in the scratch repo read `state: blocked` and `waitingFor: permission prompt` within 15 seconds, and the hook's input held `permission_mode: auto` and the command |
| In auto mode a command no rule names runs with no prompt | 5, 6 | `docs/research/2026-11-18-auto-mode-hook-probe/README.md`, and the Sessions round's live run, where `sleep 240` had no allow rule and ran. So a prompt count of zero proves nothing, and the live runs list the refused commands |
| The check starts no pass that hasn't been started | 3 | `internal/sessions/decide.go:158-195`: `Plan` makes a launch decision only for a pass already building, and `settle` in `internal/state/commit.go:194-231` moves only a waiting pass to building. `start` is called from `pass start`, `prompt yes`, and a usage relaunch. `Breaks` in `internal/state/rules.go` holds the rules a start must pass |
| The real state folder holds no project | 3, 6, conductor | `cc-desk snapshot` on 2026-11-18 printed `projects: []`. So the check's new step starts nothing on install |
| A session two merges old can still be listed as idle | 5, 6 | `docs/follow-ups/merged-pass-session-still-listed.md`: both sessions listed with `state` `done`, one 5 minutes 48 seconds after its merge. Pass 5 builds the stop step and runs the 10-minute listing as its evidence |
| A user unit's `PATH` has no `~/.local/bin`, `claude`, or `gh` | 3, 4, 5 | `systemd-run --user` on 2026-11-17 printed `PATH=/usr/lib64/ccache:/usr/local/sbin:/usr/local/bin:/usr/bin`. `cc-desk-check.service` sets its own `PATH`, and `cc-desk-page.service` sets none |
| The page refuses every write today | 4 | `web/src/hooks.server.ts` answers 405 to every method but GET and HEAD, and the Page round turned SvelteKit's origin check off |
| A ruling already starts the check | 3 | `cmd/cc-desk/ruling.go:47` calls `f.kick(cmd)` |
| The workstation can take a skill from this private repo at a tag | conductor, 6 | `gh api "repos/glw907/cc-desk/contents/README.md?ref=main"` returned the file on 2026-11-18. `claude-tooling-sync update-skills` fetches by the same call, and Survey's entry in `~/.claude/tooling/vendored-skills.json` is the model. Each installed skill is a link from `~/.claude/skills/` into the dotfiles repo |
| The repo has no tag | conductor | `git tag` printed nothing on 2026-11-18 |
| Both repos are private with `main` as the default branch | 5, 6 | `gh repo view` on 2026-11-18: `true main` for `glw907/cc-desk` and for `glw907/cc-desk-scratch` |
| Go is the same locally and in CI | all | `go1.27.4` locally on 2026-11-18, and `go.mod` reads `go 1.27` |

## Order table

Two passes build at once where the table allows it, and never more. Passes merge in number
order, except that pass 4 may land after pass 3 or pass 5.

| Pass | Planner prompt | Depends on | Planned minutes | Waits for Geoff |
| --- | --- | --- | --- | --- |
| 1 `desk-records` | `2026-11-18-desk-1-planner-prompt.txt` | Nothing | 120 | no |
| 2 `desk-skills` | `2026-11-18-desk-2-planner-prompt.txt` | 1 | 75 | no |
| 3 `desk-open` | `2026-11-18-desk-3-planner-prompt.txt` | 1, 2 | 100 | no |
| 4 `desk-page` | `2026-11-18-desk-4-planner-prompt.txt` | 1 | 90 | no |
| 5 `desk-check` | `2026-11-18-desk-5-planner-prompt.txt` | 2, 3 | 100 | no |
| 6 `desk-run` | `2026-11-18-desk-6-planner-prompt.txt` | 1 to 5, and the conductor's install | 90 | no |

Pass 1 runs alone. Passes 2 and 4 start together when it lands. Pass 3 starts when pass 2 has
landed, and pass 5 when pass 3 has. A pass lands as soon as `round-run.md` "Reading a pass"
lets it, and the next worktree is cut from `origin/main`. The least-proven assumption is the
opener's, and pass 3 is the earliest pass that can test it. The longest chain is passes 1, 2,
3, 5, and 6.

## Usage ceiling

100 percent of the weekly allowance, as in the last two rounds (Geoff, 2026-11-17: "You can go
to 100%."). The figure read 3 percent on 2026-11-18, with the reset at Nov 20, 11pm.

## Standing answers

The twelve in `round-run.md` hold, with these changes and additions.

- **4, replaced.** At 100 percent or more, or when the figure can't be read, the conductor
  finishes the passes in flight, starts nothing new, writes the figure under "Open decisions"
  in `docs/STATUS.md`, and ends.
- **1 and 9, changed by `pass-core` "Until cc-desk runs".** Nothing runs `approval-queue send`.
  The conductor is a background session, so a stop's question goes under "Open decisions" in
  `docs/STATUS.md` on `main` with what was tried, the choices, and a recommendation. It goes on
  with every pass that builds on neither the stopped pass nor one waiting for it, and ends
  when nothing can run. Geoff's answer reaches the round through a terminal session, which
  hands the round back as `conductor.md` "Hand-off" says.
- **6, changed.** Two passes run in this repo at once, so the busy check on the main checkout
  sees the other pass. A hit that names a session or a process of this round's other pass in
  flight, with its folder under that pass's worktree, isn't a stop. Any other hit is.
- **13. Two passes at once.** This replaces "passes run one at a time" in `round-run.md` "The
  order table" for this round. Pass 4 runs beside pass 2, and beside pass 3 or pass 5 if it is
  still building then. Its file list meets none of theirs. A pass whose branch falls behind
  `main` merges `origin/main` into its branch before its pull request lands, and its checks
  run again.
- **14. A pass's plan comes out over twice its minutes.** It parks, as `round.md` "Parking at
  planning" says.
- **15. A later pass needs a change in an earlier pass's files.** Expected: the passes share
  the command, the release file, and the fixtures. The pass makes the change and lists it.
- **16. A pass would change a fixture an earlier round's tests read.** It adds a fixture and
  leaves the old one as it is.
- **17. Pass 4 and the pass beside it both want a file.** `web/desktop/cc-desk-open`, the
  page's unit, and everything under `web/` are pass 4's. Everything under `cmd/`, `internal/`,
  and `skills/` is the other pass's. If pass 4 needs a verb or a snapshot field pass 1 didn't
  build, it parks on that one task and the conductor gives the task to pass 5.
- **18. A live run meets a permission prompt it didn't set up.** That is a finding. No session
  of this round can record an answer, since none is a desk. The executor ends the stuck
  session, files the prompt, adds the rule its command needs to the scratch release, and runs
  once more. A third run with such a prompt parks the pass. The prompt that item 8 sets up
  with an `ask` rule is ended the same way once the page has shown it.
- **19. Auto mode refuses a command in a live run.** After pass 5 lands the refusal is on the
  session's record. The executor files it, and the run counts as passed if the pass still
  merged. Item 7's refusal is drawn on purpose and isn't filed.
- **20. The live checks.** Items 3 to 5 and 7 to 9 run on the workstation, outside CI, and are
  never a Go test that skips. Each pull request and the round report give their output.
- **21. The state folder.** Live runs use the real state folder, `~/.local/state/cc-desk`, with
  releases for the project `cc-desk-scratch` only, one named for each run. Once pass 3 is
  installed, the timed check starts a free pass by itself, so a script adds a release only
  when its run is ready to start. The rehearsal ends with no question waiting, no report
  unread, and no session running there. No pass adds a release for the project `cc-desk`. The
  conductor adds it after pass 6 lands, and its one pass waits on Geoff's question.
- **22. The tag and the install.** When pass 5 has landed, the conductor tags `main` as
  `v0.1.0` and pushes the tag. It adds the two skills to the workstation's vendored-skills
  manifest at that tag and lands that change as `claude-infra` and
  `~/.claude/docs/claude-tooling.md` say, in a worktree of the dotfiles repo cut from its
  `origin/main` after `claude-worktree-busy ~/.dotfiles`. It installs the binary and the page
  from `main` as the README says. Pass 6 launches once `claude-tooling-sync verify` shows no
  finding about either skill and both read under `~/.claude/skills/`. If the dotfiles checkout
  can't be brought up to date, the conductor writes that under "Open decisions" and waits on
  it with `round-watch`, checking again at each wake.
- **23. Pass 6 changes a skill.** After it lands, the conductor tags `v0.1.1`, moves the pin
  the same way, and installs again before it sets up the real release.
- **24. After pass 6 lands.** The conductor installs `main` again, runs pass 6's setup script
  against the real state folder, and checks that `cc-desk snapshot` shows one question waiting
  for `cc-desk` and no session started. Then it does the hand-over in its prompt.
- **25. The report.** The conductor writes it at
  `~/Projects/initiatives/cc-desk/reports/desk-report.md`, and puts its title and first
  paragraph under "Current state" in `docs/STATUS.md`. It counts the permission prompts and
  the refused commands of every live run.
- **26. A review that escalates (Geoff, 2026-11-18).** Asked about PR 18, Geoff said "Let's
  prep so that I stop doesn't happen." A PR whose closing review reads `escalate` or `fix`
  lands like one that reads `accept` when every acceptance item carries its evidence, CI is
  green, and the body says `Calls: none`. Each escalated point must also be settled by an
  `[advisor]` line with a follow-up for a later pass. The conductor names each such landing in
  the report. A point that changes what the design says still stops and waits for Geoff.
- **27. The steps after passes 5 and 6 (Geoff, 2026-11-18).** In the same message Geoff
  approved every step in answers 22 to 24 ahead of time. That covers pushing the `v0.1.0` and
  `v0.1.1` tags and landing the dotfiles PR that pins both skills. It also covers installing
  the binary, the timer, the page, and the link handler from `main`, and setting up the real
  release in `~/.local/state/cc-desk`. When auto mode refuses one of these steps, the conductor
  retries it once and cites this answer. It edits no setting, and a second refusal goes under
  "Open decisions".
- **28. A plan over its estimate (Geoff, 2026-11-18).** After passes 2 and 3 parked on their
  estimates, Geoff chose "Yes, with limits". When a plan parks at planning only because its path
  is over twice its estimate, the conductor raises the estimate to half the path plus about 15
  minutes and launches the executor. That holds only when the plan's scope matches the design
  and the planner trimmed nothing. A plan whose scope grew, or one whose path runs past the
  round's clock, still waits for Geoff. The conductor records each raise under `## Rulings` in
  the pass's plan and in the report.
- **29. Pass 5 takes the third verdict (Geoff, 2026-11-18).** A failure review found that
  `pass ready` refuses any verdict but `accept` or `fix`, so a session holding an `escalate`
  has no honest line to run, and pass 6's real run goes through that verb. Geoff chose "Fold
  into pass 5". Pass 5 adds it: `pass ready` takes `escalate` and a count of calls for Geoff,
  and the check merges a pass whose change is inside the design, whose checks passed, and
  that has no open finding and no call for Geoff. That is answer 26, built into cc-desk. The
  working skill's ready line names the third verdict. This widens pass 5's scope by Geoff's
  word, so answer 28 still applies to its estimate.
- **30. The pass-core edits ride the dotfiles PR (Geoff, 2026-11-18).** Geoff chose "With the
  dotfiles PR". The dotfiles PR of answer 22 also carries the `pass-core` edits the failure
  review proposes, landed through `claude-infra` and reviewed with the rest. The review is
  `~/Projects/initiatives/cc-desk/reports/desk-failure-review.md`, section 3. Answer 27 covers
  landing it.
- **31. Pass 6 may start past the clock (Geoff, 2026-11-18).** At 23:27 the forecast put pass
  6's launch at about the conductor's 03:55 cutoff. Geoff said "Yes, let pass 6 start past the
  cutoff." The cutoff still holds for everything else: pass 5, the steps after it, and pass 6
  each start only when the one before has landed. Pass 6 launches whenever those steps finish,
  and the hand-over follows it as planned. A stop that waits for Geoff still goes under "Open
  decisions".
- **32. A skill line the design doesn't back (Geoff, 2026-11-18).** Geoff chose "Fix it to
  match". When the read of both skills after pass 5 finds a line that tells a session to do
  something the design doesn't, and the design says plainly what the line should say, the
  conductor fixes the skill to match in a small PR. One whole-branch `standards:diff-reviewer`
  read reviews it, and it lands by the usual test before the `v0.1.0` tag. A fix that would
  change what the design says still waits for Geoff under "Open decisions". The report names
  each such fix.
- **33. Leftover scratch sessions (Geoff, 2026-11-19).** Pass 5 parked because a blocked session
  from its own earlier probe kept its live script from starting. Geoff chose "Yes, and from now
  on". For the rest of the round, the conductor stops any `cc-desk-scratch` session that this
  round's own live runs left blocked or idle, notes it in the report, and relaunches the parked
  executor. A session the conductor can't trace to this round still waits for Geoff.
- **34. A 1.1 round answers the lessons, before Move (Geoff, 2026-11-19).** Geoff asked whether
  to collect the lessons and do an organized 1.1 after the first release, and chose "1.1 round
  before Move". This replaces answer 30. The dotfiles PR of answer 22 carries only the two
  skills' manifest entries, and no `pass-core` edit. Every lesson stays in
  `docs/follow-ups/desk-round-lessons.md`, and a stop later in this round adds its row there. The
  hand-over writes the design prompt for a 1.1 round that answers each row, for cc-desk and
  `pass-core` together, and names it as the next round. The Move design prompt is still written,
  and it runs after 1.1 lands. Pass 5's third verdict (answer 29) stays in this round.
- **35. Pass 6's estimate, past the clock (Geoff, 2026-11-19).** Answer 28 can't raise an
  estimate once the round is past its clock, and it already is, so pass 6's plan would park on
  its estimate and wait for Geoff. Asked before the launch, he chose "Raise to the plan's path
  (Recommended)". When pass 6's plan parks only because its path is over 180 minutes, its scope
  matches the design, and the planner trimmed nothing, the conductor sets the pass's estimate to
  the plan's path and launches the executor. The executor's stop is then twice that path. The
  conductor records the raise under `## Rulings` in the pass's plan and in the report. A plan
  whose scope grew still waits for Geoff.
- **36. Stop after pass 5; pass 6 in a fresh session (Geoff, 2026-11-19).** Geoff wrote, with his spelling corrected, "Let's
  stop after pass five." and then "We can launch six in a fresh session." The second conductor
  lands pass 5, puts main's build back on the timer, writes the report through pass 5 and the
  hand-over, and ends. A fresh conductor, launched from
  `docs/plans/2026-11-19-desk-pass-6-conductor-prompt.txt`, does the steps after pass 5 (the
  skills read, the `v0.1.0` tag, the dotfiles PR, the install), then pass 6, the steps after it,
  and the round's last hand-over.
- **37. Move starts at 1.0, beside 1.1 (Geoff, 2026-11-19).** Geoff wrote to the third conductor:
  "I'd like to start the work on both cairn and dubplate once we hit the cc-desk version 1.0. We
  can work on cc-desk 1.1 concurrently with those efforts." This changes answer 34's order. The
  Move round, which puts cairn and dubplate under cc-desk, no longer waits for 1.1 to land. The
  conductor reads 1.0 as the release this round finishes: pass 6 landed and Geoff's sitting
  passing done item 10. After the sitting, the 1.1 design and the Move design can both start,
  and the hand-over names both. Everything else in answer 34 holds: the dotfiles PR carries only
  the two skills, and the 1.1 round answers each row of the lessons ledger.
- **38. Pass 6 runs overnight (Geoff, 2026-11-19).** The conductor asked under "Open decisions"
  about four stops no answer covered. Geoff wrote: "I'd like this effort to run overnight." The
  conductor reads that as the recommended answers to the two stops that only need a ruling. An
  executor that reaches twice its plan's path stops as usual, and the conductor relaunches it
  once from its branch with the stop at three times the path; a second stop waits for Geoff. A
  pull request whose `Calls:` line counts one or more lands by the rest of the usual test when
  every call stays inside the design and no finding is open; the conductor names each call in
  the report, for the sitting. A fix that needs a setting, hook, or permission rule changed, and
  a third live run with a prompt it didn't set up, still wait for Geoff under "Open decisions",
  while everything that builds on neither goes on.
- **39. The Move design is written overnight, before the sitting (Geoff, 2026-11-19).** Geoff
  wrote: "Once you have the rehearsal done, I'd like you to shape both the dubplate and cairn
  work to run in cc-desk. This should be done overnight and before my sitting, so both can launch
  as soon as my sitting is done." This changes answer 37's timing for the Move design only. Asked
  whether it must wait for the rehearsal, Geoff wrote: "Do we even need to wait for the
  rehearsal?" It doesn't, so the conductor launches the Move design session at once,
  `cc-desk-move-design`, in the background from `docs/plans/2026-11-19-move-design-prompt.txt`.
  It writes the design, the round plan, a planner prompt for each pass, and the conductor prompt,
  with every call that is Geoff's listed with a recommendation, so his approval after the
  sitting is one read. It launches nothing and adds nothing to the state folder, so the sitting
  still sees one question waiting for cc-desk. Move's conductor launches only once the sitting
  has passed done item 10 and Geoff has approved the design. A busy-check hit in this repo that
  names the Move design session isn't a stop for pass 6 (answer 6, changed). The 1.1 design still
  starts after the sitting.

## Start line

The session that takes Geoff's approval launches the conductor from the main checkout, with
this line run alone, once the six states read true:

```text
claude --bg --model claude-opus-5-5 --effort medium --name conductor-cc-desk "$(cat /var/home/glw907/Projects/cc-desk/docs/plans/2026-11-18-desk-conductor-prompt.txt)"
```

The conductor checks the six again before it launches the first planner.

1. `docs/STATUS.md` on `origin/main` says Geoff approved the design, and names the conductor
   prompt as the next action.
2. `claude auto-mode` shows the entry `cc-desk-desk-round`.
3. `claude agents --json` lists no other session named `conductor-cc-desk` and none named
   `cc-desk-desk-*`.
4. The main checkout is on `main`, and `git rev-list --count HEAD..origin/main` prints 0 after
   a fetch.
5. `claude-worktree-busy /var/home/glw907/Projects/cc-desk` exits 0.
6. The weekly usage figure is under 100.

No other round is approved or running in this repo. The Sessions and Page rounds closed with
pull request 16.

### Launch checks, run on 2026-11-18

| Check | Output |
| --- | --- |
| `~/.local/share/cc-desk/bin/cc-desk usage` | `3% used, resets Nov 20, 11pm (America/Anchorage)` |
| `git status -sb` after a fetch | `## main...origin/main`, with the two new designs uncommitted. The commit that carries this file levels it |
| `claude-worktree-busy .` | Exit 1, "uncommitted file docs/specs/2026-11-18-desk-skills-design.md edited in the last 60 minutes". That is this session's writing. The conductor runs it again |
| `git check-ignore -q .claude/worktrees/x` | exit 0 |
| `git worktree list` | The main checkout only, after this session removed six finished `agent-*` worktrees, each clean and with no commit `main` lacked, with `git worktree remove` and `git branch -d` |
| `claude --help` | Lists `--bg`, `--model`, `--effort`, `--name`, `--session-id`, `--settings`, `--permission-mode`, and `--append-system-prompt` |
| `claude agents --json` | A blocked `cairn-cms-tool-delete` in another repo, a session in ecxc-ski, and this session. None is this round's |
| `claude --bg --model sonnet --permission-mode auto --settings <file> --name cc-desk-scratch-ask-probe "<prompt>"`, from the scratch checkout | `backgrounded · 218a503c · cc-desk-scratch-ask-probe`. `claude stop 218a503c` then printed `stopped 218a503c` |
| `claude-worktree-busy ~/.dotfiles` | Exit 0 |
| `gh repo view glw907/cc-desk-scratch` | Private, default branch `main` |

## Decisions

- Six passes against the route's three, and 575 planned minutes against 225, for the reasons
  the design gives. Geoff chose six on 2026-11-18.
- Pass 1 runs alone and builds every new verb and the brief, so the skills and the page are
  written against names that exist. The builder check found the first order had the skills
  written before the verbs they name.
- Two passes build at once inside one round. The last two rounds did the same as two rounds
  under one conductor.
- The conductor is a background session on Opus at medium effort, where `conductor.md` gives
  Sonnet. It does four things beyond a conductor's usual work, and each takes judgment: the
  tag, the install of the two skills through a dotfiles pull request, the setup of the real
  release, and the prompt for the next round's design.
- The executors of pass 2 and pass 6 run on Opus at medium effort, and the other four on
  Sonnet. Pass 2 writes what every later session is told and what the desk says to Geoff. Pass
  6 fixes whatever the rehearsal shows broken, which no plan can list. `pass-core` "Models"
  allows Opus where the design calls the work judgment.
- The auto mode entry names the install of the two skills. The log shows `stow -R claude`
  refused once as a change to Claude's own instructions, and the install runs it.
- Live runs use the real state folder, as the Sessions round's did.
- The four planning checks the last two rounds' reports asked for go into every planner
  prompt of this round, since the planner reference has no room for them under its word cap.
