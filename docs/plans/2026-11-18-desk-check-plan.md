# Desk check: pass plan

Initiative: cc-desk
Round: desk
Pass: desk-check

Class `runner` and `live-account`. Planned 100 minutes. The path below is 315 minutes, with gate,
review, live wall clock, and CI time included. That is over twice the plan (200), so the pass
parks at planning under the round's standing answer 14. The plan's scope is the design's section
"The check's repairs" and done item 7, plus the third verdict that standing answer 29 added by
Geoff's word. Nothing was trimmed to fit. Standing answer 28 would let the conductor raise the
estimate to half the path plus about 15 minutes (about 175, a stop at 350), but only when the
path fits the round's clock, and it doesn't. The conductor started at 11:44 on 2026-11-18
(Alaska time), and its prompt starts nothing new past 970 minutes, at 03:55 on 2026-11-19
(12:55 UTC). Planning ended near 09:30 UTC, so a 315-minute path ends near 14:45 UTC, past the
cutoff. The stop in the draft pull request's body asks the conductor to put that to Geoff.

Design: `docs/specs/2026-11-18-desk-skills-design.md`, sections "The check's repairs", "Failure
modes" (the rows for the conductor, the refusal hook, a usage stop with no time, a merge call
past its limit, and a stop on a session already gone), "Left to the builder", and done items 7
and 11. The design wins over this plan. Round plan: `docs/plans/2026-11-18-desk-skills-round.md`.
Where the round is silent, the main, State-and-rules, and Sessions designs under `docs/specs/`
hold. Anything a person reads (the PR body, STATUS lines, README lines, follow-ups, the skill)
follows `/var/home/glw907/.claude/docs/voice/written-for-people.md`, loaded through the
`writing-voice` skill before drafting.

Worktree `/var/home/glw907/Projects/cc-desk/.claude/worktrees/desk-check`, branch `desk-check`,
cut from `origin/main` at `3dc674b`. PR base: `main`. Push: `glw907/cc-desk`. Pass 3
(`desk-open`, PR 20) landed the free-pass start, the replacement, the merge answers, and the
desk. Pass 4 (`desk-page`) may build beside this pass. Everything under `web/` and the page's
unit is pass 4's (standing answer 17). This pass owns `cmd/`, `internal/`, `e2e/`, `skills/`,
`scripts/`, and the follow-ups named in Task 9. The page needs nothing new from this pass: the
snapshot already carries `refusals` (`internal/snapshot/sessions.go:45`), and the stop reasons
and decisions this pass adds reuse words the page has (`usage`, `relaunch_limit`, `launch`,
`ask`, `wait_usage`).

## Pre-flight

- Permissions checked: the denials log `~/.local/state/claude-denials.jsonl`, 137 rows from
  2026-11-01T00:55:05Z to 2026-11-19T06:45:46Z (UTC). By class: classifier unavailable 52,
  Destructive or Unwanted Actions 16, Self-Modification 11, Interfere With Workloads 11, the
  server-side classifier's unlabeled refusals 10, Undermining Oversight 5, Unauthorized Access 4,
  Merge Without Review 3, Auto-Mode Bypass 3, and two each of Irreversible Local Destruction,
  Git Destructive, Create Public Surface, Unverifiable Deletion Scope, Sensitive-Source
  Provenance, and Security Weaken. Rows matching this pass's commands: `claude stop` 6 (the last,
  a lone `claude stop d5647ebf` at 2026-11-18T21:28:10Z), `kill` 6, `git push` 4 (none on a
  cc-desk feature branch), `gh pr merge` 1, `gh pr view` or `gh pr checks` 3, `git stash drop` 2
  (both `git stash drop stash@{0}`, 2026-11-17), `git -C <path> reset --hard` 1, `claude agents`
  1 (piped), `sleep` 7, `claude --bg` 1, a settings-file write 10, and a chained `mv` of a release
  file out of the real state folder 1 (2026-11-19T06:45:46Z). `systemctl` 0, `cc-desk` 2
  (classifier outages). The table below gives the form each refused kind takes. No step needs
  Geoff.
- The state folder holds the desk-open live run's leftovers. `releases/cc-desk-scratch.json` is
  live, round `desk-live`: `desk-live-a` is stopped with reason `scope` and its session record
  open; `desk-merge` and `desk-close` are done. The session list shows `f0402e97` and `781526f0`
  (`cc-desk-scratch-desk-live-a`, blocked) and `adeb426a` and `c7a8dd66` (done). `release add`
  refuses a second live release for one project (`internal/state/events.go:65`), and no verb
  retires a release. So this pass's live script never adds a release while one is live: it
  appends its own round to the live scratch release (mechanism table, "The live release"). The
  two blocked orphan rows are the conductor's to stop before it launches the executor; if they
  remain, standing answer 52 holds.
- Reviewed: one `standards:plan-reviewer` read (Opus, medium) on 2026-11-19, with 4 blocking, 11
  major, and 11 minor findings. Each is folded, and the `[review]` lines under `## Decisions` say
  how.
- Evidence: the pre-code runs under `## Acceptance`, the hook probe under
  `docs/research/2026-11-18-auto-mode-hook-probe/`, and the counts under "Recorded searches".

### Refused commands and their alternatives

| Kind | Form this pass uses |
| --- | --- |
| `claude stop` | Only through the `cc-desk` binary (the check's carry-out, or a live script phase), never typed by the executor. Task 3's one probe of `claude stop` on an id already stopped runs alone, never chained |
| `kill` | Only of a process id the script recorded, run alone. Never `pkill -f` |
| `claude agents --json` | Redirected to a file, then read. Never piped |
| `gh pr merge`, `gh pr view`, `gh pr checks` | Only the check's binary calls them, on a `glw907/cc-desk-scratch` pull request. The executor reads CI with `gh run list` and `gh run watch` |
| `git push` | `git push origin desk-check` from the worktree, and the live script's push of its own scratch branch to `glw907/cc-desk-scratch`. Never main, never `--force` |
| `git stash drop`, `git reset --hard` | Never run by the executor. Task 1's live probe asks the scratch session to run one, so auto mode refuses it there |
| `mv` or `rm` in the real state folder | None. The live script edits only the scratch release, through `cc-desk` verbs or under `flock` on the state lock (mechanism table) |
| Settings edits | None by the executor. No `.claude/settings*.json` change, no `autoMode` entry. The launched session's settings file is written by `WriteSettings` in the built binary, never by hand: a hand-written one was refused as Self-Modification |
| `sleep` in a stand-in | Inside a stand-in script the test or live script writes, never typed as a foreground command |

## Mechanism table

| Clause | Mechanism | Source |
| --- | --- | --- |
| The refusal hook | `WriteSettings` adds `hooks.PermissionDenied`, beside `PermissionRequest`: one command, `<binary> --state-dir <dir> refusal hook`, `timeout` 30. The binary is the absolute path `WriteSettings` already demands | `internal/launch/settings.go:90`, `:102` |
| Its input | Read whole from stdin and kept as the fixture `e2e/testdata/refusal/hook-input.json` from Task 1's live probe. The hooks reference gives the fields `session_id`, `tool_name`, `tool_input`, `tool_use_id`, and `reason`, and says the event fires on auto mode's classifier denials, not on a deny rule. The target is a named key per tool: `command` for Bash, `file_path` for Read, Edit, and Write, `url` for WebFetch, `pattern` for Grep and Glob, else `tool_input` as compact JSON. An input with no reason records an empty one | code.claude.com hooks reference, "PermissionDenied"; `~/.local/bin/claude-denied-log` reads `.reason` |
| The record | A new event `RecordRefusal(sessionID, Refusal, *found)` appends to `Session.Refusals` of the one open record with that `session_id`, with `noSettle` set, as `RecordPrompt` does. The `Refusal` type and the snapshot field exist; nothing writes them today. Both relaunch paths in `RecordLaunch` carry `Refusals` to the next record, as they carry `Prompts` | `internal/state/prompt.go:59`; `internal/state/types.go:125`; `internal/state/launch.go:102`; `internal/snapshot/sessions.go:111` |
| It doesn't stop the session | The hook prints nothing and exits 0 on success. It never writes `hookSpecificOutput.retry` | Design "Refused commands are recorded" |
| The hook can't write | The verb exits 0 with no output, so the session carries on, and appends one line (time, session id, error) to `<state>/refusal-hook.log`. An unmatched session id keeps the raw input under `<state>/hook-unmatched/` as `prompt hook` does | Design "Failure modes", the refusal hook row; `cmd/cc-desk/prompt.go:130` |
| The usage reset | `early` gains `now`, and so do `Watched` (`internal/sessions/decide.go:102`) and its caller (`cmd/cc-desk/check.go:181`). For a record stopped on usage (a pass's `StopReason`, or a session record's for the conductor), a narrow parser reads the reset from the stop's detail: `resets <Mon> <D>, <H>[:<MM>]am\|pm (<Zone>)`, or `resets <H>[:<MM>]am\|pm (<Zone>)` with no date, read in `time.LoadLocation(<Zone>)`. The binary imports `time/tzdata`, so a zone loads with no system database. A dated reset is read in the stop's year, and in the next year when that falls more than 180 days before the stop began. A time alone takes its next occurrence after the stop began. While `now` is before the reset, the decision is `wait_usage` with the detail "waits for the reset at <time>". At or after it, the relaunch runs uncounted | `internal/sessions/decide.go` `early`, `resetText`; the recorded texts in `docs/follow-ups/live-sample-of-limit-texts.md` and `e2e/testdata/usage40.sh` |
| No readable time | The usage carry-out already writes `Relaunch{Cause, Counted}` through `RecordLaunch` (`cmd/cc-desk/carryout.go:91`). The rule is in the decision: a usage relaunch whose stop gave no readable time is counted when the record already holds a `Relaunch` with cause `usage`, and uncounted otherwise. A stop that gives a readable time waits for it and relaunches uncounted, whatever came before. Counted usage relaunches share the limit of two with every other counted relaunch | Design "A model limit holds the relaunch" and its failure row |
| The fourth stop | At the limit of two counted relaunches the decision is `stop` with reason `relaunch_limit`. The pass is already stopped on usage and `StopPass` refuses a pass that isn't building, so a new event `RestopPass(project, pass, reason, detail)` changes the open stop's reason and detail and adds no stop interval. `applyRecorded` calls it when the pass is already stopped. The decision comes from the count, after `blockedRow`'s `stop usage` (`decide.go:435`) would otherwise give `usage` again | `internal/state/checkrun.go:98`; `internal/state/events.go:337` |
| The stop step | `cleanup` and the `End` and `EndAndRelaunch` carry-outs read the list (`sessions.List`, which runs `claude agents --json --all`) and stop a row for the session whose state is working, blocked, or done, through `runner.Stop`. Then they read the list again. The session is gone when the second read shows no row for the id or a row whose `Class()` is `ClassNone` (stopped or failed). A failed stop on a session gone by that rule counts as ended; otherwise the error stands. A row already stopped or failed takes no stop. Task 3 first runs `claude stop` once on an id that is already stopped and records its exit code and text, so the rig's `stop-fails` control copies the real CLI | `cmd/cc-desk/merge.go:229`, `sessionLive` at `:92`; `internal/sessions/read.go:26`, `list.go`; `internal/launch/run.go:175`; `docs/follow-ups/launch-review-notes.md` item 1 |
| The step in the log | The cleanup outcome names it: "merged pull request <n>; stopped session <id>; cleaned up", or "...; session <id> already gone; cleaned up" | `internal/state/launch.go:207` `RecordOutcomes` |
| The 60-second limit | `callMerge`, the fetch, and the pull each run under their own `context.WithTimeout(ctx, limit)`. `internal/launch/worktree.go`'s `git` gains a context form, and `Pull(ctx, cwds)` gives each call its own context. Each limited `exec.Cmd` sets `SysProcAttr.Setpgid` and a `Cancel` that kills the process group, so a grandchild holding the pipe dies with it, and `WaitDelay` of 5 seconds as a backstop. A probe in review showed `WaitDelay` alone returns at the delay with the grandchild still running. An error after the context's deadline reads "stopped at the <limit> limit". A hidden `check` flag `--call-limit` (default `60s`) lets a test shorten it | `cmd/cc-desk/merge.go:71`; `internal/launch/worktree.go:19`, `:93` to `:123`; `cmd/cc-desk/check.go` hidden flags |
| The stand-ins that sleep | The e2e `gh` stand-in and a `git` stand-in on the test's `PATH` start a child `sleep 120` that holds stdout, when a control file says so, and write its process id to a file. The `git` stand-in passes every other call to the real `git` by absolute path | `--gh-command`; `PATH` for `git`; `e2e/launch_rig_test.go` control files |
| Three failed merge calls | A new `PullRequest.MergeFailures int`. A failed or timed-out merge call raises it in the same run's record commit, and the carry-out returns the outcome "merge call failed (<n> of 3): <last line>" as a plain outcome, not an error, so the log keeps that line and the check exits 0. The third failure in a row asks in that same commit, with a new reason `merge_failed`, through `AskMerge`. `AskMerge` resets the count, so each "merge" answer gets three more calls; `answeredRow`'s merge row takes the same rule. `MergePass` clears the count, and `ReportReady` keeps it when the pull request's number is unchanged | Design "A merge call can't hang or fail forever"; `internal/state/merge.go:30`, `:68`, `:159`; `docs/follow-ups/merge-call-fails-for-good.md` |
| A success between | The count lives on one pull request. A merge call that succeeds merges it, and `MergePass` zeroes the count, so the count is never read again. A new pull request on the same pass starts at 0 | Done item 7; `[design-gap]` under `## Decisions` |
| The third verdict | `VerdictEscalate = "escalate"`. `pass ready` takes `--verdict accept\|fix\|escalate` and `--calls <n>` (default 0, so old fixture plans still run). `ReportReady` stores `PullRequest.Calls`. `internal/state/check.go:298` accepts the new verdict | Standing answer 29; `internal/state/merge.go:41` |
| Calls for Geoff | In `decideMerge`, after the checks pass and before the merge row: `Calls > 0` asks with a new reason `calls_for_geoff`, recommending merge. `escalate` with no calls merges, like `accept`: standing answer 26 lands such a PR when the body says `Calls: none` and each escalated point has an `[advisor]` line. `fix` with calls is still a fix: the fix row comes first. An open finding is reported as `fix`, so `pass ready` takes no count of findings | Standing answers 26 and 29 |
| The skill's ready line | `skills/cc-desk-work/SKILL.md:37` (the ready line) and `:42` (the verdict bullet) name the three verdicts and `--calls`. The table row at `:25` stays. `internal/launch/templates/pass.txt:15` matches | Standing answer 29 |
| A report due | In `sessions.Plan`, for a live release with a checkout: a round's report is due when it has a pass, every pass is done, and no report exists for the project and round. The release's report is due when every round has a pass, every pass in the release is done, and no report has an empty round. With one due and no open conductor record, the check adds `launch` with reason `report_due` on the `conductor` record. `applyRecorded` admits that decision with no record, as it does a pass's first launch | Design "The conductor starts when a report is due"; `internal/state/report.go:110`, `:132`; `internal/state/checkrun.go:67` |
| The conductor's record | A new field, `Session.ForReport` (the round's name, or `release`), set by the launch and carried by both relaunch paths in `RecordLaunch`. The limit of two is per report. When the last conductor record for a due report ended without the report recorded (it died, it ran `session finish` with no report, or its launch failed), the next start for that report is a counted relaunch of that record, and the third reads `relaunch_limit` and "Needs you now". A first launch that fails writes an ended conductor record with the failure as its stop, so it counts. A conductor whose report is recorded ends as today. A due report for another round starts a fresh record | `internal/sessions/decide.go:307`, `decideRead`; `internal/state/launch.go:79` to `:121`, `:183`; `internal/snapshot` `todoFor` |
| The conductor's worktree | `place()` puts the conductor at `.claude/worktrees/conductor` on branch `conductor`, and `Make` refuses a path or branch that exists. When a conductor record ends with its report recorded, the carry-out removes that worktree and branch with the helper the pass cleanup uses, so the next report's launch can make it | `cmd/cc-desk/launcher.go:102`; `internal/launch/worktree.go:46` to `:71` |
| The conductor's prompt | When `ForReport` is set, the template's line "Conduct the rounds of the project." is replaced by a block: write the report for that round (or the release), record it, finish, and do nothing else. A relaunched conductor first runs the brief and writes no second report when one is recorded. `relaunch.txt` says the same | Design: "A relaunched conductor that finds the report recorded writes no second one. It has no other job in this round"; `internal/launch/templates/conductor.txt:5` |
| The live release | The live script never adds a release while one is live. With none live, `prep` adds `scripts/live/repairs/release.json` under a name with the run's UTC stamp. With one live, `prep` appends this run's round `repairs-<stamp>` and its passes to the live scratch release with `jq`, as `live-desk.sh` `add_pass` does, holding `flock` on `<state>/lock` for the write. Each run's round has a held pass (`scripts/live/desk/hold-plan.md`, which only waits), so no release report ever comes due in the scratch project: a release report is keyed on the project alone, and one recorded here would stand for pass 6's rehearsal | `scripts/live-desk.sh:226`, `:459`; `internal/state/lock.go`; `internal/state/report.go:137` |
| Each new step in the log | The refusal hook isn't a check step. Each other step's decision carries an outcome line through `RecordOutcomes`: `wait_usage` with the reset, the counted usage relaunch, `stop` `relaunch_limit`, the stop step, a merge call failed, `ask` `merge_failed`, `ask` `calls_for_geoff`, `launch` `report_due`, and the conductor worktree's removal | `internal/state/launch.go:207`; `checkrun.go:38` |

## Recorded searches

Run in the worktree at `3dc674b` over `cmd internal e2e skills scripts`, Go, text, Markdown, and
shell files.

| Pattern | Hits | Files |
| --- | --- | --- |
| `claude stop\|Runner.Stop\|\.Stop\(ctx` | 22 | `cmd/cc-desk/carryout.go`, `answer.go`, `merge.go`, `launcher.go`, `internal/launch/run.go`, `e2e/recovery_test.go`, `e2e/replaced_launch_test.go`, `scripts/live-check.sh` |
| `resetText` | 8 | `internal/sessions/decide.go`, `internal/usage/usage.go`, `e2e/order_test.go` |
| `StopUsage` | 19 | 9 files: `cmd/cc-desk/carryout.go`, `internal/sessions/decide.go`, `decide_test.go`, `plan_test.go`, `internal/snapshot/sessions.go`, `sessions_test.go`, `internal/state/types.go`, `events_test.go`, `checklog_test.go` |
| `Counted` | 41 | 16 files |
| `hook-unmatched` | 4 | `cmd/cc-desk/prompt.go`, `e2e/prompt_test.go` |
| `PermissionRequest\|PermissionDenied` | 16 | `cmd/cc-desk/prompt.go`, `internal/launch/settings.go`, `launch_test.go`, `e2e/launch_test.go`, `e2e/prompt_test.go`. `PermissionDenied` alone: 0 in `settings.go` |
| `WriteSettings` | 16 | `internal/launch/settings.go`, `launch_test.go`, `cmd/cc-desk/launcher.go`, `open.go`, `internal/desk/settings.go`, `settings_test.go` |
| `func git\(\|git\(` | 44 | `internal/launch/worktree.go`, `launch_test.go`, five e2e files |
| `callMerge` | 3 | `cmd/cc-desk/merge.go` |
| `WaitDelay\|WithTimeout` | 0 | in `internal/launch/run.go`, `worktree.go`, `cmd/cc-desk/merge.go` |
| `VerdictAccept\|VerdictFix` | 8 | `internal/state/types.go`, `merge.go`, `check.go`, `internal/sessions/merge.go`, `merge_test.go` |
| `--verdict` and the ready line | 17 | `cmd/cc-desk/pass.go:90`, `:97`; `e2e/merge_test.go:351`; `e2e/merge_rig_test.go:171`; `e2e/kick_test.go:87`, `:142`; `internal/launch/templates/pass.txt:15`; `internal/launch/launch_test.go:164`; `skills/cc-desk-work/SKILL.md:25`, `:37`, `:42`; `scripts/live-desk.sh:476`; `scripts/live/timer-live-a-plan.md:17`; `timer-live-b-plan.md:30`; `scripts/live/desk/desk-live-a-plan.md:15` |
| `FixAttempts` | 14 | 6 files under `internal/state` and `internal/sessions` |
| `StopRelaunchLimit` | 7 | `internal/sessions/decide.go`, `internal/snapshot/sessions.go`, `internal/state/types.go`, `sessions_test.go` |
| `KindConductor\|"conductor"` | 58 | 16 files |
| `Refused\|Refusal` | 84 | 24 files; the `Refusal` writers outside tests: none (`internal/state/check.go:343` only validates) |
| `early(` callers | 3 | `internal/sessions/decide.go` (`Decide`, `Watched`), `cmd/cc-desk/check.go:181` through `Watched`, `decide_test.go:232` |

## File list

| Task | Files |
| --- | --- |
| 1 | `internal/launch/settings.go`, `launch_test.go`; `cmd/cc-desk/refusal.go` (new), `root.go`; `internal/state/refusal.go` (new), `launch.go` (the relaunch carry); `e2e/refusal_test.go` (new), `e2e/testdata/refusal/hook-input.json` (new, from the probe); `scripts/live-repairs.sh` (new, its `prep` and `refuse` phases only) |
| 2 | `internal/sessions/decide.go`, `reset.go` (new), `decide_test.go`, `reset_test.go` (new); `internal/state/checkrun.go`, `events.go` (`RestopPass`); `cmd/cc-desk/check.go` (the `Watched` call), `carryout.go` (the usage relaunch's count), `main.go` or `root.go` (the `time/tzdata` import); `e2e/check_test.go`, a new usage fixture with a checkout under `e2e/testdata/` |
| 3 | `cmd/cc-desk/merge.go` (`cleanup`, `sessionLive`), `carryout.go` (`End`, `EndAndRelaunch`); `e2e/merge_test.go`, `e2e/launch_rig_test.go` (the `stop-fails` control, matched to the probe) |
| 4 | `internal/launch/worktree.go`, `launch_test.go`; `cmd/cc-desk/merge.go` (`callMerge`, the `Pull` call), `check.go` (the hidden flag); `e2e/limit_test.go` (new), `e2e/launch_rig_test.go` (the sleeping `gh` and `git` stand-ins) |
| 5 | `internal/state/types.go`, `merge.go`, `merge_test.go`; `internal/sessions/merge.go`, `merge_test.go`, `answered_test.go`; `cmd/cc-desk/merge.go`; `e2e/merge_test.go` |
| 6 | `internal/state/types.go`, `merge.go`, `check.go`, `merge_test.go`; `internal/sessions/merge.go`, `merge_test.go`; `cmd/cc-desk/pass.go`, `merge.go` (`ask` wording); `internal/launch/templates/pass.txt`, `launch_test.go`; `skills/cc-desk-work/SKILL.md` (`:37`, `:42`); `e2e/merge_rig_test.go`, `kick_test.go` helpers |
| 7 | `internal/sessions/decide.go`, `plan_test.go`; `internal/state/types.go`, `launch.go`, `checkrun.go`; `cmd/cc-desk/launcher.go`, `carryout.go` (the conductor worktree's removal); `internal/launch/prompt.go`, `templates/conductor.txt`, `relaunch.txt`; `skills/cc-desk-work/SKILL.md` (the conductor's lines only); `e2e/conductor_test.go` (new) |
| 8 | `scripts/live-repairs.sh` (extends Task 1's file with the rest of its phases), `scripts/live/repairs/` (new fixture plans and release) |
| 9 | `docs/follow-ups/`: remove `auto-mode-prompt-clause.md`, `merge-call-fails-for-good.md`, `merged-pass-session-still-listed.md`, `usage-relaunch-loop-on-a-model-limit.md`; reword `launch-review-notes.md` without item 1; add `live-scripts-start-a-conductor.md` (Task 7 makes a round that finishes in `live-desk.sh` or `live-check.sh` start a scratch conductor, and those scripts don't expect one); `README.md` lines for the new flags and verb |

Files two tasks touch, and the task that owns the first change to each:

| File | Tasks | Owner |
| --- | --- | --- |
| `internal/state/types.go` | 1, 5, 6, 7 | New fields only, each in its own place, so the lanes' merge is mechanical |
| `internal/state/launch.go` | 1, 7 | Task 1 adds `Refusals` to the relaunch carry; Task 7 adds `ForReport` beside it |
| `internal/sessions/decide.go` | 2, 7 | Task 2, lane A, which runs Task 7 after it |
| `cmd/cc-desk/check.go` | 2, 4 | Task 2 changes the `Watched` call; Task 4 adds only the hidden flag |
| `cmd/cc-desk/carryout.go` | 2, 3, 7 | Task 2; Tasks 3 and 7 touch other functions |
| `cmd/cc-desk/merge.go` | 3, 4, 5, 6 | Lane B, in task order |
| `internal/sessions/merge.go` | 5, 6 | Lane B, in task order |
| `e2e/launch_rig_test.go` | 3, 4 | Lane B, in task order |
| `skills/cc-desk-work/SKILL.md` | 6, 7 | Task 6 the ready lines; Task 7 the conductor's lines |
| `scripts/live-repairs.sh` | 1, 8 | Task 1 writes `prep` and `refuse`; Task 8 extends it |

## Acceptance

The design's items, word for word:

> 7. **Repairs.** A session launched in the scratch repo on a command auto mode refuses has the
>    refusal on its record with a cause. A fixture usage stop that names a reset time isn't
>    relaunched before it and is after it. A fixture pass whose usage stops name no time is
>    relaunched after the first stop with its count unchanged, and after the second and third
>    with the count at 1 and 2. At the fourth stop it reads "Needs you now". After a fixture
>    merge in the scratch repo with the timer on, the check's log shows the stop step, and a
>    listing each minute for 10 minutes shows the session gone. The listing is in the pull
>    request. Stand-ins that sleep for the merge call, the fetch, and the pull are each ended at
>    60 seconds. The third failed merge call in a row makes a question, and a success in
>    between starts the count again. With a report due and no conductor, the next check starts
>    one. It starts none when no report is due, and none while one runs. A conductor ended
>    three times with its report due is relaunched twice and then reads "Needs you now".
> 11. **The gate.** `bash scripts/gate.sh` passes with no skipped test. Items 2 to 5 and 7 to 10
>     have parts that run live and aren't Go tests. Each live script prints one line for each
>     check and ends with its result. A pass's pull request holds that output from a run on the
>     pass's last commit, and a pass that leaves it out isn't done.

Standing answer 29, as this pass's items:

> 29a. `pass ready` takes `--verdict escalate` and `--calls <n>`. A Go test reports each of
>      `accept`, `fix`, and `escalate` ready, and the check decides `merge`, `fix`, and `merge`
>      for them with checks passed, inside the design, and no calls.
> 29b. A Go test reports `escalate` with `--calls 1`: the check asks with reason
>      `calls_for_geoff`, recommending merge, and a "merge" answer merges without asking again.
> 29c. `skills/cc-desk-work/SKILL.md`'s ready line names the third verdict and `--calls`.

### Clause table

Live means `scripts/live-repairs.sh` against real Claude Code in `~/Projects/cc-desk-scratch`,
sessions named `cc-desk-scratch-*`, and the real state folder holding only that project's
releases. Rig means a Go test with stand-ins, run by the gate.

| Item | Clause | Where | Task | Command |
| --- | --- | --- | --- | --- |
| 7 | A refused command's refusal is on the record with a cause | Live (probe); rig on the kept input | 1, 8 | `bash scripts/live-repairs.sh refuse`; `go test ./e2e -run TestRefusalHook` |
| 7 | A usage stop naming a reset isn't relaunched before it, is after it | Rig: two check runs on a new fixture with a checkout, 5 minutes before and 5 after the reset | 2 | `go test ./e2e -run TestUsageReset` |
| 7 | No time: uncounted, then counted 1 and 2, then "Needs you now" | Rig: the launch rig on a fixture with a checkout, four usage stops on one pass with a seen run between each, the snapshot's `needs_you_now` read after the fourth | 2 | `go test ./e2e -run TestUsageNoTime` |
| 7 | The check's log shows the stop step after a fixture merge with the timer on | Live: the `check-log.jsonl` line for the merge carries "stopped session" or "already gone" | 3, 8 | `bash scripts/live-repairs.sh merge` |
| 7 | A listing each minute for 10 minutes shows the session gone; the listing is in the PR | Live: 10 reads of the default `claude agents --json`, redirected to a file, one a minute from the merge; each line `pass` when the id has no row | 8 | `bash scripts/live-repairs.sh listing` |
| 7 | Stand-ins that sleep for the merge call, the fetch, and the pull are each ended at 60 seconds | Rig at `--call-limit 2s`; live at the real 60 seconds, three timed runs against a fixture state folder | 4, 8 | `go test ./e2e -run TestCallLimits`; `bash scripts/live-repairs.sh limits` |
| 7 | The third failed merge call in a row makes a question | Rig: the `gh` stand-in fails three runs; the third run's commit asks `merge_failed` | 5 | `go test ./e2e -run TestMergeFailures` |
| 7 | A success in between starts the count again | Rig: state tests that `MergePass` zeroes the count and that a new pull request on the same pass starts at 0 | 5 | `go test ./internal/state -run TestMergeFailures` |
| 7 | With a report due and no conductor, the next check starts one | Rig: a fixture round with every pass done and no report | 7 | `go test ./e2e -run TestConductorStart` |
| 7 | None when no report is due, none while one runs | Rig: the same fixture with the report recorded; with a working conductor row | 7 | the same |
| 7 | Ended three times with its report due: relaunched twice, then "Needs you now" | Rig: missing-row runs with a seen run between each; and the finish-without-report and failed-launch paths | 7 | the same |
| 7 | The conductor starts live | Live: a scratch round of one fixture pass merges, the next check launches `cc-desk-scratch-conductor`, which records one report, and no second conductor starts | 7, 8 | `bash scripts/live-repairs.sh conductor` |
| 11 | The gate with no skipped test | The gate | 10 | `bash scripts/gate.sh` |
| 11 | The live script's output, from a run on the last commit, in the PR | The PR body | 11 | `bash scripts/live-repairs.sh` |
| 29a | Three verdicts | Rig | 6 | `go test ./internal/sessions ./e2e -run 'TestVerdicts'` |
| 29b | Calls above zero | Rig | 6 | `go test ./e2e -run TestCallsForGeoff` |
| 29c | The ready line | `grep -n escalate skills/cc-desk-work/SKILL.md` and `grep -n -- --calls skills/cc-desk-work/SKILL.md` | 6 | those two |

### Pre-code runs on `3dc674b`

The gate ran bare in the worktree and ended `gate: ok`: every Go package `ok` (e2e 53.052s),
vale and the skills check passed, and the web steps ended "175 passed (2.0m)". These checks show
today's behavior where it differs from the design:

```text
$ go test ./e2e -run 'TestCheckStops' -count=1 -v
=== RUN   TestCheckStops/a_usage-limit_relaunch
--- PASS: TestCheckStops (0.08s)
ok  	github.com/glw907/cc-desk/e2e	0.475s
```

That subtest's fixture, `e2e/testdata/check/releases/chk.json`, records no checkout, so the
check never relaunches there and its "writes no relaunch" assertion holds under the new rule.
Task 2's tests use a new fixture with a checkout.

```text
$ grep -c PermissionDenied internal/launch/settings.go
0
$ grep -cE "WaitDelay|WithTimeout" internal/launch/run.go cmd/cc-desk/merge.go internal/launch/worktree.go
internal/launch/run.go:0
cmd/cc-desk/merge.go:0
internal/launch/worktree.go:0
$ go run ./cmd/cc-desk pass ready --help
  cc-desk pass ready <project> <pass> --pr <n> --commit <sha> --verdict accept|fix --design inside|outside [flags]
$ for p in merge_failed calls_for_geoff report_due ForReport MergeFailures "merge call failed (" RecordRefusal; do grep -rF -- "$p" cmd internal e2e skills | wc -l; done
0 (each of the seven)
$ grep -c -- --calls skills/cc-desk-work/SKILL.md; grep -c escalate skills/cc-desk-work/SKILL.md
0
0
```

The state folder's `releases/cc-desk-scratch.json` is live (pre-flight), and
`claude agents --json` lists eight rows, four of them the desk-open run's scratch sessions.

## Tasks

The executor commits after each task by path. Lane A is Tasks 2, 7, and 8, in that order: Tasks
2 and 7 share `decide.go`, and Task 8 extends Task 1's script, so it starts once Task 1 is on the
branch. Lane B is Tasks 1, 3, 4, 5, and 6, which share `cmd/cc-desk/merge.go` and
`internal/sessions/merge.go` and run in that order. The shared-file table above names each
owner. The two lanes run as pairs, as `executor.md` "Pairs" says, with descriptions `desk-check
task <n>`. Tasks 9 to 11 are the executor's own.

### Task 1: refusals are recorded (55 minutes)

Outcome: the mechanism rows "The refusal hook" to "The hook can't write". Files: the Task 1 row.
Class `runner`. Order:

1. `WriteSettings` writes the `PermissionDenied` entry, and the `refusal hook` verb exists, at
   first keeping its whole input under `<state>/hook-unmatched/` and the log.
2. The live probe: `bash scripts/live-repairs.sh refuse` (this task writes the script's `prep`
   and `refuse` phases) builds the binary, sets up this run's round (mechanism table, "The live
   release") with one pass whose fixture plan asks the session to run `git stash drop stash@{0}`
   in the scratch checkout, and lets the check launch it in auto mode. First it reads the
   scratch checkout's `.claude/settings*.json` deny lists: a deny rule refuses without the
   classifier, and the hook doesn't fire for it. The phase waits up to 10 minutes for a hook
   input naming the session's id, then copies it to `e2e/testdata/refusal/hook-input.json`. The
   workstation's own log, `~/.local/state/claude-denials.jsonl`, is read for a row with the same
   session id: it tells "the hook didn't fire" apart from "nothing was refused".
3. The parser and `RecordRefusal` are built against the kept input, and both relaunch paths
   carry `Refusals`.

Check: `e2e/refusal_test.go` feeds the kept input to `refusal hook` with a state folder whose
pass record has that `session_id`; the record gains one refusal with the tool, the target, and a
non-empty reason, and the verb prints nothing and exits 0. With a release file that fails to
parse, it exits 0, prints nothing, and appends one line to `refusal-hook.log`. An unknown id
keeps the input under `hook-unmatched/`. A relaunch keeps the refusal on the next record.
`launch_test.go` finds the hook entry with the absolute binary. Test:
`go test ./internal/state ./internal/launch ./e2e -run 'TestWriteSettings|TestRefusal|TestRelaunch'`.

### Task 2: a usage limit holds the relaunch (50 minutes)

Outcome: the mechanism rows "The usage reset" to "The fourth stop". Files: the Task 2 row. Class
`runner`. Check: `reset_test.go` table-tests the parser on the four recorded texts, a date that
wraps the year, a time alone, an unknown zone, and a text with no `resets`. A new fixture with a
checkout (standing answer 16) holds a usage stop at 2026-11-20T20:00Z naming "resets Nov 20,
10:59pm (America/Anchorage)" (07:59Z on the 21st): a run at 07:54Z decides `wait_usage` with the
reset in its detail, and a run at 08:04Z an uncounted relaunch. `TestUsageNoTime` runs the launch
rig through four stops with no time, with a seen run between each: the relaunches read counted
false, true, true, and the fourth stop decides `stop` `relaunch_limit` (not `usage`), with the
pass in `needs_you_now` with that reason and one open stop interval. A mixed case gives a stop
with a readable time after a no-time relaunch: it waits, then relaunches uncounted. A conductor
record stopped on usage takes the same rule. `TestWatched` follows the new signature. Each
decision's line is in `check-log.jsonl`. Test: `go test ./internal/sessions ./internal/state
./e2e`.

### Task 3: the cleanup stops the session (35 minutes)

Outcome: the mechanism rows "The stop step" and "The step in the log". Files: the Task 3 row.
Class `runner`. First, run `claude stop` once, alone, on the id of a scratch session the list
already shows as stopped, and record its exit code and text in the task's report; the rig's
`stop-fails` control copies them. Check: in the merge rig, a merge whose session row is `done`
stops it (the `claude` stand-in's log holds `stop <id>`) and the outcome names the stop. A row
already `stopped` takes no stop, and the outcome says "already gone". With `stop-fails` set and
the stand-in marking the row stopped, the cleanup counts it ended and finishes. With
`stop-fails` set and the row still `done`, the cleanup's error stands and the pass keeps no
`cleaned` time. `End` takes the same cases. Test: `go test ./e2e -run 'TestMerge|TestEnd'`.

### Task 4: the merge call, the fetch, and the pull are limited (45 minutes)

Outcome: the mechanism rows "The 60-second limit" and "The stand-ins that sleep". Files: the
Task 4 row. Class `runner`. Check: `TestCallLimits`, with `--call-limit 2s`, runs three cases,
each with a stand-in whose child `sleep 120` holds stdout: the merge call (the run returns within
6 seconds, and the outcome names the limit), the fetch (the cleanup errors within 6 seconds), and
the pull (the cleanup's note names the limit). After each case, the stand-in's recorded child
PID has no `/proc/<pid>`. Test: `go test ./internal/launch ./e2e -run 'TestCallLimits|TestPull'`.

### Task 5: three failed merge calls make a question (40 minutes)

Outcome: the mechanism rows "Three failed merge calls" and "A success between". Files: the Task
5 row. Class `runner`. Check: `internal/state` tests that a failure raises the count, `MergePass`
zeroes it, `AskMerge` resets it, a new report with the same number keeps it, and a new pull
request on the same pass starts at 0. `TestMergeFailures` runs the `gh` stand-in failing three
runs in a row: runs 1 and 2 log "merge call failed (1 of 3)" and "(2 of 3)" and exit 0, and run
3 logs "(3 of 3)" and asks `merge_failed` in the same commit. A "merge" answer gets three more
calls. A merge call stopped at the limit counts as a failure. Test:
`go test ./internal/state ./internal/sessions ./e2e -run 'MergeFail'`.

### Task 6: the third verdict (40 minutes)

Outcome: the mechanism rows "The third verdict" to "The skill's ready line". Files: the Task 6
row. Class `runner`. Check: `TestVerdicts` reports `accept`, `fix`, and `escalate` ready on
three passes with checks passed and no calls: `merge`, `fix`, `merge`. `TestCallsForGeoff`
reports `escalate` with `--calls 1`: `ask` `calls_for_geoff`, recommending merge; the answer
"merge" merges with no second question. A `fix` with calls is a fix. `--verdict review` is
refused, naming the three. Both 29c `grep` lines print a line, where both print nothing on
`3dc674b`. Test:
`go test ./internal/state ./internal/sessions ./internal/launch ./e2e -run 'Verdict|Calls|Ready'`.

### Task 7: the conductor starts for a report (75 minutes)

Outcome: the mechanism rows "A report due" to "The conductor's prompt". Files: the Task 7 row.
Class `runner`. Check: `plan_test.go` table-tests "due" for a round with a pass not done, all
done and no report, all done and a report, an empty round, the release report after every
round's, and a release holding a pass not done (never due). `TestConductorStart` runs the check
on a fixture with a due report: one `launch` `report_due` on `conductor`, the stand-in's `--bg`
call names `<project>-conductor`, and the same run's log holds every other decision too. With
the report recorded, no launch; with a working conductor row, no launch. Missing-row runs with a
seen run between each give relaunches counted true, true, then `stop` `relaunch_limit`, and
`needs_you_now` holds the conductor. A conductor that runs `session finish` with no report
recorded, and a first launch that fails, each count toward the same limit. A conductor that
records its report and finishes ends, and its worktree and branch are gone. A relaunch keeps
`ForReport`. The filled prompt names the round and the no-second-report rule and no longer says
"Conduct the rounds". Test: `go test ./internal/sessions ./internal/state ./internal/launch
./e2e -run 'Conductor|Plan|Relaunch'`.

### Task 8: the live script (60 minutes)

Outcome: `scripts/live-repairs.sh` with phases `prep`, `refuse`, `merge`, `listing`, `limits`,
`conductor`, `abort`, and `evidence`, in the shape of `scripts/live-check.sh` (the same `say`,
`try`, and `prep_refuse` helpers). `prep` refuses on a dirty scratch checkout, on a release of
another project, and on a live (working or blocked) `cc-desk-scratch-*` row that no record of
this run names; it then sets up the run's round as "The live release" says. Each check prints
one line, `pass` or `fail`, and the script ends with `result: pass` or `result: fail`. `merge`
runs a one-pass fixture round with the timer on, so the timed check merges it. `listing` starts
at the merge and runs in the background beside `conductor`: it reads the default session list
once a minute for 10 minutes into the run folder. `limits` runs the three sleeping stand-ins at
the real 60 seconds through the built binary with `--claude-command` pointed at the e2e `claude`
stand-in, `--gh-command`, and a `PATH` `git`, against a fixture state folder the script makes
and a throwaway clone of the scratch repo as the checkout, and times each call from that state
folder's check log. It never reads or copies the real state folder. `conductor` waits for the
next timed check to start the scratch conductor, waits up to 20 minutes for its report, and
checks one report recorded, no second conductor, and its worktree gone. `abort` stops this
run's sessions through the binary. Class `runner`. Check: `bash -n` and `shellcheck` on the
script, and `bash scripts/live-repairs.sh help`. Test: the whole run in Task 11.

### Task 9: follow-ups and words (15 minutes)

Outcome: the four follow-ups named in the file list are removed, each in the commit of the task
that settles it or here. `launch-review-notes.md` loses item 1 and keeps items 2 to 7 as they
stand. `live-scripts-start-a-conductor.md` is added. The README lists `refusal hook`, `--calls`,
and the third verdict. Class `runner`. Check: `vale` and `tellgrader --register docs` on the
README and the changed follow-ups. Test: the gate.

### Task 10: fold the lanes, review, gate (25 minutes)

Outcome: lane B merged into lane A's branch, the gate green, and one whole-branch
`standards:diff-reviewer` read folded. Class `runner`. Check: `bash scripts/gate.sh` ends
`gate: ok` with no skipped test.

### Task 11: the live run on the last commit (60 minutes)

Outcome: `bash scripts/live-repairs.sh` from start to end on the commit the PR will hold, every
line `pass`, and `result: pass`, then the PR body and the CI watch. Its wall clock is two to six
minutes of timer cadence, the 10-minute listing beside the conductor's report, three 60-second
limits, and CI. Class `runner`. Check: the output, pasted whole into the PR. A fix found here
makes a new commit, and the gate and the run repeat on it, as standing answer 60 limits.

### The path

| Step | Lane A | Lane B |
| --- | --- | --- |
| Build | Tasks 2, 7, 8: 185 | Tasks 1, 3, 4, 5, 6: 215 |
| Words | Task 9: 15 | |
| Fold | Task 10: 25 | |
| Run | Task 11: 60 | |

The path is 215 + 15 + 25 + 60 = 315 minutes. The sum over every task is 500. The figures are in
the unit the round's earlier plans used, and they aren't shaved to fit. The review's findings
added 15 minutes to the path (Tasks 1, 3, 4, and 7).

## The gate and CI

Run `bash scripts/gate.sh` bare, never piped, with the Bash tool's `timeout: 600000`. A plan or
doc is staged before the gate, since the gate lints tracked files only. CI: `gh run list --branch
desk-check --limit 3`, then `gh run watch <id> --exit-status` with `run_in_background`. No review
runs on a commit while CI runs.

## Standing answers

The round plan's answers 13 to 32 and the twelve in `round-run.md` hold. This pass's start at 50,
clear of the numbers the round's other plans use.

- **50. The probe's hook fires and the record holds no reason.** Record the refusal with an empty
  reason, keep the input as the fixture, and list it. The design's "with a cause" then rests on
  the workstation's log row for the same session, which the live line prints beside the record.
  A reason of "classifier unavailable" is a cause on the record, but it isn't the classifier's
  refusal: try the same command once more.
- **51. The probe's session isn't refused, or the hook doesn't fire.** If the workstation's log
  has no row for the session, nothing was refused: try the next command in order (`claude stop
  00000000`, then `git reset --hard HEAD~1` on the session's own branch), three in all. If the
  workstation's log has a row and the state folder doesn't, the hook didn't fire. Either way,
  after the third command Task 1's live clause parks, with both logs in `## Stop`, and the
  other tasks go on.
- **52. Orphan scratch sessions are still live at a live step.** The executor doesn't stop
  another run's sessions. Task 1's live step and Task 8's live phases wait; it finishes every
  other task, then writes `## Stop` naming the sessions, and ends.
- **53. A test's expectation changes with a repair.** The expectation changes, never a fixture
  an earlier round's tests read (answer 16), and the PR lists each change with its reason.
  `TestWatched` (`internal/sessions/decide_test.go:232`) changes for the new signature.
- **54. A stand-in's `sleep` outlives the test.** That is a failed check. The group kill is the
  mechanism; find which call lacks it.
- **55. The live merge doesn't stop the session within the 10 minutes.** That is a failed check,
  never retried away. The listing goes into the PR as it stands.
- **56. The scratch conductor writes more than its report.** Abort it through the script, file
  it, and fix the prompt. A third run that does it parks the pass.
- **57. A scratch conductor starts in someone else's live run.** It isn't this pass's to end.
  The follow-up from Task 9 names it.
- **58. The timer runs this pass's build after the run.** Expected. The script prints the line
  that puts main's build back, and the conductor reinstalls from `main` once the pass lands.
- **59. Auto mode refuses a command the live script itself runs.** Use the table above once.
  Task 1's refusal is drawn on purpose and isn't filed (answer 19).
- **60. The live run ends `result: fail`.** The remaining phases still run, so the output shows
  every clause. A fix for a traced cause gets one more run on its commit, as the design allows
  one run after a failed one. A failure with no traced cause, or a second failed run, ends the
  pass: write `## Stop` with the bar unmet and the output, and mark nothing ready.
- **61. Something else changes the shared state mid-run.** The timer, the real state folder,
  `check-log.jsonl`, and the scratch checkout are shared with the real conductor and pass 4. A
  clause that fails because another session changed one of them gets one rerun of that phase,
  with the other session's change named in the PR.

## Decisions

- [advisor] The pass parks at planning on its estimate alone. Its scope is the design's section
  and standing answer 29, and nothing was trimmed. Answer 28's raise would be about 175, but
  the 315-minute path runs past the round's 12:55 UTC cutoff, so its proviso fails. The stop
  asks the conductor to put three choices to Geoff: allow the path past the cutoff, as answer 31
  did for pass 6 (recommended); split the live script and Task 11 into a pass of their own; or
  wait for the next round.
- [advisor] The stop step stops every row that is working, blocked, or done, and reads the list
  again. A failed stop on a row gone by the second read (no row, or stopped or failed) counts as
  ended.
- [advisor] The merge-failure count resets when the check asks, so each "merge" answer gets
  three more calls.
- [advisor] A dated reset is read in the stop's year and counts as passed when it is before the
  stop. A time alone takes its next occurrence.
- [call] `escalate` with no calls merges like `accept`, as answer 26 lands such a PR; `fix` with
  calls is a fix. The calls row sits after the checks and before the merge.
- [call] An open finding is reported as `fix`, so answer 29's "no open finding" needs no new
  flag.
- [call] `--calls` defaults to 0, so the fixture plans earlier rounds read still run unchanged.
- [call] The hook's error log is `<state>/refusal-hook.log`, one appended line per failure.
- [call] The conductor's limit of two is per report, and each record carries the report it was
  started for. `[design-gap]`: the design doesn't say whether the limit is per report or per
  release.
- [call] `[design-gap]`: on one pull request a successful merge call merges it, so "a success in
  between" can't be followed by a later failure on that pull request. The plan proves the count
  is zeroed by the merge and starts at 0 on a new pull request.
- [call] The live probe runs before the parser is written, so the fixture is a real input.
- [call] The live script appends its round to the live scratch release under the state lock,
  with a held pass, so it never needs a release retired and the scratch project never records a
  release report.
- [review] B1: the stop step reads `--all` and treats a stopped or failed row as gone; Task 3
  probes the real `claude stop` first.
- [review] B2: "The live release" row replaces adding a release per run.
- [review] B3: `limits` runs on a fixture state folder, the `claude` stand-in, and a throwaway
  clone.
- [review] B4: `checkrun.go` is in Task 7, with a test that the run's other decisions are kept.
- [review] M1 to M4: the per-report limit, the seen runs in the test, the relaunch carry of
  `Refusals` and `ForReport`, and the conductor worktree's removal. The `live-desk.sh` and
  `live-check.sh` steps moved to a follow-up.
- [review] M5: process-group kill, with `WaitDelay` as the backstop, and one context per call.
- [review] M6: the success clause, as the `[design-gap]` line above says, and the third
  failure asks in its own run.
- [review] M7: the usage carry-out already writes the relaunch; the fixture with a checkout and
  the mixed series are named, and the conductor takes the same rule.
- [review] M8 and m2: the `Watched` sites are in Task 2, and the shared-file table names each
  owner.
- [review] M9 and m9: standing answers 60 and 61.
- [review] M10: the first `[advisor]` line. M11: the second `[call]` line.
- [review] m1, m3 to m8, m10, m11: the skill's line cites, a release file that fails to parse,
  the per-tool target keys and the deny-list read, answer 51's end state, the outcome kept as a
  plain line, the `relaunch_limit` assertion and `time/tzdata`, the replaced template line, the
  listing beside the conductor phase, and the pre-code controls.
