# Desk run: pass plan

Initiative: cc-desk
Round: desk
Pass: desk-run

Class `runner` and `live-account`. Planned 90 minutes. The path below is 675 minutes, with the
gate, the review, two runs of each live script, and CI time included. That is over 180, so the
pass parks at planning under the round's standing answer 35. Under that answer the conductor
sets the estimate to 675 and launches the executor. Its stop is then 1350 minutes, and answer
38 gives one relaunch, stopping at 2025. The scope is the design's section "The rehearsal and
the real run", less the real run, done items 8 to 11, and the two follow-ups the round plan
folds into this pass. Nothing was trimmed, and nothing was added past the design. The two
small verbs, `release done` and `session end`, and the fix to how a report is matched are the
builder's means for sentences the designs already hold: a release's file moves to
`releases/done/` when it ends, sessions end "by the binary", and a release's report falls due
when its last round does. Verb and field names are "Left to the builder" (see `## Decisions`).
The park PR carries two questions for Geoff, on item 8's word "killed" and on item 11's "last
commit" (the two `[ask]` decisions). Neither stops the run: each has a recommended reading the
executor takes until a ruling comes.

Design: `docs/specs/2026-11-18-desk-skills-design.md`, sections "The rehearsal and the real
run", "When something goes wrong", "Left to the builder", and done items 8 to 11. The design
wins over this plan. Round plan: `docs/plans/2026-11-18-desk-skills-round.md`, standing answers
13 to 38. Where the round is silent, the main design (`docs/specs/2026-11-17-cc-desk-design.md`,
done item 6), the State and rules design, and the Sessions design hold. The real release comes
from `docs/specs/2026-11-18-stuck-sessions-design.md`. Anything a person reads (the PR body,
STATUS lines, README lines, follow-ups, a skill) follows
`/var/home/glw907/.claude/docs/voice/written-for-people.md`, loaded through the `writing-voice`
skill before drafting.

Worktree `/var/home/glw907/Projects/cc-desk/.claude/worktrees/desk-run`, branch `desk-run`, cut
from `origin/main` at `d7279c7`. PR base: `main`. Push: `glw907/cc-desk`. Passes 1 to 5 have
landed, and no pass builds beside this one, so this pass owns `cmd/`, `internal/`, `e2e/`,
`scripts/`, `skills/`, and the follow-ups named in Task 9. It changes nothing under `web/`.

## Pre-flight

- Owner-only steps: none. The round's auto mode entry `cc-desk-desk-round` is in
  `~/.claude/settings.json` and names `cc-desk-desk-run`. Through the binary, and only on
  sessions it launched and recorded, it lets this pass launch `cc-desk-scratch-*` sessions,
  "end those sessions by the binary or by `kill` of the process id on the session's record",
  and lets the timed check merge on `glw907/cc-desk-scratch`. It does not let this pass add a
  `cc-desk` release, restart a unit, or install the skills again. Those are the conductor's. Each
  live prep's `cc-desk timer install` writes and enables the two check units; passes 3 to 5
  ran it under the same entry, and the denials log holds no row for it.
- Permissions checked: a fresh read on 2026-11-20 at 08:30 UTC of
  `~/.local/state/claude-denials.jsonl`, 145 rows from 2026-11-01T00:55:05Z to
  2026-11-20T06:38:53Z. By class: classifier unavailable 52, Destructive or Unwanted Actions 23,
  Self-Modification 11, Interfere With Workloads 11, unlabeled server-side refusals 11,
  Undermining Oversight 5, Unauthorized Access 4, Merge Without Review 3, Auto-Mode Bypass 3, and
  under 3 each for the rest. The rows since the last plan's read are drawn `git stash drop`
  refusals in the scratch refuse passes. Rows matching this pass's commands: `claude stop` 6,
  the last a lone `claude stop d5647ebf` refused at 2026-11-18T21:28:10Z with the round's entry
  in force; `kill` 5, all in cairn-cms with no entry, four chained with other commands and one
  a lone `kill 2088698` at 2026-11-17T19:49:16Z; a chained `mv ...; rm -f` of a release file
  in the real state folder 1 (2026-11-19T06:45:46Z); `git worktree` 4 (chained); `git push` 4
  (none on a cc-desk feature branch); `gh pr merge` 1; `claude agents` 1 (piped); `systemctl`
  0; `git switch` or `git checkout` 0. The table below gives the form each kind takes. No step
  needs Geoff.
- The real state folder holds pass 5's leftovers: one live release
  `releases/cc-desk-scratch.json` (name `cc-desk-scratch-20261119T074459Z`), 15 entries under
  "Needs you now", 6 unread round reports (`rep-0001` to `rep-0006`, each for a `merge-` round),
  and no live scratch session (`claude agents --json` read on 2026-11-20). `release add`
  refuses a second live release for one project (`internal/state/events.go:65`), so no live run
  can start until that release ends. No verb ends a release today. Task 1 adds one, and Task 4
  clears the leftovers with it. A probe on a copy (below) showed the clearing leaves 0 needs,
  0 waits, and 0 unread.
- The scratch checkout `~/Projects/cc-desk-scratch` is on `main` at `549108a`, clean, with 16
  worktrees and 30 local branches that passes 3 to 5 left. This pass leaves them alone: every
  name it makes carries a UTC stamp, so none collides. `glw907/cc-desk-scratch` has
  `delete_branch_on_merge` false, and its CI (`.github/workflows/gate.yml`) runs
  `bash scripts/gate.sh` on pull requests.
- The install: `cc-desk version 0.1.0` at `~/.local/share/cc-desk/bin/cc-desk`, the skills
  under `~/.claude/skills/` at `v0.1.0`, the timer running the check every 5 minutes
  (`OnUnitActiveSec=5min`), and the page on `127.0.0.1:7908` (HTTP 200 at 08:36 UTC), started
  with `CC_DESK_BINARY` set to the installed binary.
- Reviewed: one `standards:plan-reviewer` read (Opus, medium) on 2026-11-20, which returned 3
  blockers, 13 fixes, and 8 minor findings. Each is folded, and the `[review]` lines under
  `## Decisions` say how.

### Refused commands and their alternatives

| Kind | Form this pass uses |
| --- | --- |
| `claude stop` | Never typed. A session ends through `cc-desk session end` (Task 2), whose stop runs inside the binary, or by a lone `kill` of the recorded process id |
| `kill` | Only of the process id a script prints for a record, after its guards, run alone as `kill <pid>`. Never `pkill`, never `-9`, never chained. A refusal waits 60 seconds and retries once, then the session ends through `session end` (standing answer 70) |
| `mv` or `rm` in the real state folder | None. A release ends through `cc-desk release done <project>` (Task 1), and a report is read through `cc-desk report read` or the page's request |
| `claude agents --json` | Redirected to a file, then read. Never piped |
| `gh pr merge` | Never typed. Only the timed check's binary calls it, on a `glw907/cc-desk-scratch` pull request |
| `git push` | `git push origin desk-run` from the worktree, and the scripts' pushes of their stamped scratch branches to `glw907/cc-desk-scratch`. Never `main`, never `--force`, never `--delete` |
| `git switch` in the scratch checkout | Run alone, by the rehearsal's prep and close, only on a clean checkout (mechanism table, "The scratch base") |
| `git worktree`, `git branch -d` | Never typed by the executor. The check's cleanup runs them |
| `systemctl --user` | Read only (`is-active`, `show`, `cat`). The timer changes only through `cc-desk timer install`. No unit is stopped or restarted |
| Settings edits | None. No `.claude/settings*.json` change, no `autoMode` entry, no hook. A launched session's settings file comes only from `WriteSettings` in the built binary |
| `sleep` | Inside a script's wait loop, never typed as a foreground command |

## Mechanism table

| Clause | Mechanism | Source |
| --- | --- | --- |
| A release ends | New verb `cc-desk release done <project>`. A new event moves the project's live release to `releases/done/<name>-<UTC stamp>.json` and marks it done. `Commit` writes the new file atomically and then removes `releases/<project>.json`, both under its lock. The verb refuses when no release is live, and when a session one of the release's records holds is live in a fresh list read, naming the record. Live means the Sessions design's live row: listed, and not reading `stopped`, `done`, or `failed`. A looser test would refuse every close, since a stopped row stays listed `done`. It takes `check.lock` first, as `session start` does (`cmd/cc-desk/session.go:66`), so no check run can relaunch into a release that just moved. It prints the new path. The snapshot then reads the release only for the pace | State and rules design, "When a release ends, its file moves to `releases/done/`" (`docs/specs/2026-11-17-state-and-rules-design.md:99`); Sessions design, "A live row" (`docs/specs/2026-11-17-sessions-design.md:159`); `internal/state/commit.go:58`, `events.go:36` to `:70`, `read.go:180` |
| A report belongs to its release | A report is matched today by project and round name only (`internal/sessions/decide.go:219`, `internal/state/report.go:137`). So once a project holds one release report, every later release of that project reads as reported, and the `report` verb refuses with "project already has a release report". The fix adds the release's name to each report, as an `omitempty` field, and matches on it in `ReportDue`, `ReportRecorded`, `AddReport`, the `report` verb, the check's carry-out, and the snapshot's `buildReports`, which today folds a report's passes from the project's first release. A round report is tagged with the live release that holds the round, or else the newest done one; a release report with the live release. A report recorded before the fix, with no release name, keeps the old match. `main`'s build reads state with `DisallowUnknownFields` (`internal/state/read.go:38`), so once a run writes the field, only this branch's build can load the folder until the pass lands | Design line 378, "a release's [report is due] when its last round does"; `internal/state/types.go:462` `Report`; `internal/state/report.go:126` to `:143`; `cmd/cc-desk/report.go:38`; `cmd/cc-desk/carryout.go:194`, `:239`; `internal/snapshot/snapshot.go:86` to `:118`; State and rules design "Left to the builder", the new fields |
| A session ends | New verbs `cc-desk session end pass <project> <name>` and `cc-desk session end conductor <project>`, in the kind-first shape of `session start` (`cmd/cc-desk/session.go:51`). The rest of this plan writes them `session end`. It takes `check.lock` first, as `session start` does, so a check run in flight can't relaunch past it. It reads the record's open session and runs `stopSession`, the stop-and-reread helper the merge cleanup uses, then prints that helper's step ("stopped session <id>", "session <id> already gone", or the done-row line). It refuses a record with no open session, so it acts only on a session the binary launched and recorded. It records nothing: the check reads the ended row as it reads any other | Round entry, "end those sessions by the binary"; `cmd/cc-desk/merge.go:143` `ended`, `:152` `step`, `:169` `stopSession`; `session.go`; follow-ups `live-abort-waits-on-stopped-sessions.md` and `leftover-live-session-parks-a-pass.md` |
| The check reads a death | A row missing or reading `stopped` or `failed` gives a counted `relaunch`, reason `missing`, at the next check run, but only once a check has seen the session. A relaunch builds a new record with `last_seen` empty (`internal/state/checkrun.go:85` sets it), and a death before any check sees it gives an uncounted `launch_again`, then `stop` `launch_failed` on the second. So every death waits until the record's `last_seen` is later than its `launched`. A row reading `done` waits until 15 minutes after the later of the launch and the last time it was seen working, then gives a counted `end_and_relaunch`, reason `idle`. With checks 5 minutes apart, that fires 20 to 25 minutes after the launch. In both branches, a session already at 2 counted relaunches gets `stop` `relaunch_limit` instead. Pass 5's run drew one counted `missing` relaunch live (`check-log.jsonl`, 2026-11-20T03:06:37Z, `pass:refuse-20261120t030028z`) | `internal/sessions/decide.go:37`, `:40`, `decideRead` `:488` to `:536` (the limit at `:511` to `:512`), `doneRow` `:538` to `:553` (the limit at `:547` to `:548`) |
| A killed process | The Sessions design says Claude Code's supervisor brings a killed process back within about 20 seconds, and proof 6 saw one kill come back as `working busy` under a new PID. Whether a kill ever gives a row the check reads as ended is unproved, so Task 4 probes it before any death is scripted | `docs/specs/2026-11-17-sessions-design.md:161`; `docs/research/2026-11-17-prove-findings.md:220` to `:263` |
| The kill target | `state.Session` holds no process id. A script reads the list into a file, takes the row whose `sessionId` equals the record's `session_id`, and checks that the row's name starts `cc-desk-scratch-`, its `cwd` is at or under `~/Projects/cc-desk-scratch`, `/proc/<pid>/comm` reads `claude`, and `/proc/<pid>/cmdline` holds no `--bg-spare` (every background host and spare process also reads `claude`). Only then does it print the PID. The executor runs `kill <pid>` alone | `internal/sessions/list.go`; `internal/state/types.go` `Session` |
| Recovery: the conductor | The recovery release's first round has one pass whose pull request the script writes itself, as `live-repairs.sh` does with `scripts/live/repairs/merge-plan.md`. Its merge makes the round's report due, and the check launches `cc-desk-scratch-conductor` for it. The script prints its PID within the conductor's first minute, the executor kills it, and the script lists it each minute for 10 minutes. It prints which brought it back: the same session id under a new PID (the supervisor), or a new id from the check. A conductor killed before any check has seen it comes back on the check's uncounted `launch_again`, the usual case; one already seen comes back on a counted `relaunch`. Item 8 asks for one death here, so either path passes, and the script names which ran. The page at `http://127.0.0.1:7908/`, fetched each minute, shows no stop for it | `internal/sessions/decide.go:274`, `:299` (the conductor's limit), `:505` to `:510`; design "The conductor starts when a report is due" |
| Recovery: the pass | The second round holds a pass on a new fixture, `scripts/live/recover/work-plan.md`, that keeps working: 300 steps, each running the scratch gate (`bash scripts/gate.sh`), resumed at the next step after a relaunch. It keeps `hold-plan.md`'s lines that say not to plan the pass again, run a review, or update STATUS. The gate is already a scratch allow rule, and the release carries it. A foreground `sleep` can't serve: Claude Code's Bash tool blocks it (probed 2026-11-20: "Blocked: sleep 50"), and a session refused its one command goes idle, which brings back the death that proves nothing. Task 4 measures a step's minutes, and the executor sizes the step count so each launch works at least 15 minutes. Each death waits until the record's `last_seen` is later than its `launched`, so it reaches a counted relaunch. The pass on `hold-plan.md` replies `ready` and goes idle within a minute, so the idle rule would relaunch it with no death at all, and three "deaths" on it would prove nothing. Each death's pass condition is the order in the check log and the saved lists: a check saw the row `working` after the previous launch, then the kill or end line, then the death's decision. After the second relaunch, the next death gives `stop` `relaunch_limit`, and the page reads "Needs you now" with the round's name. A list saved every 30 seconds through the run never shows two live rows for the pass. A row counts as the pass's when its name, short id, or session id is on the pass's records, or its `cwd` is under the pass's worktree, since proof 6's revived copy carried another name | `decide.go:511` to `:512`, `:547` to `:548`; `docs/research/2026-11-17-prove-findings.md:220` to `:263` |
| Recovery: the ask | The release's `ask_rules` holds one rule, `Bash(cksum ask-probe.txt)`. No session runs `cksum` unprompted, and only the third round's plan, `scripts/live/recover/ask-plan.md`, runs that exact command. Its prompt goes through the `PermissionRequest` hook onto the record, the session blocks, and within 10 minutes the page reads "Needs you now" and the line names `cksum ask-probe.txt`. Then the script ends the session through `session end` (standing answer 18) | `internal/launch/settings.go` `WriteSettings`; `decide.go:557` `lastPrompt` |
| One recovery release | `ask_rules` are release-wide, but only the ask pass runs `cksum`, so one release holds all three rounds. The two-pass limit orders them: the merge pass and the working pass first, the ask pass once the merge frees a slot | Sessions design, "Limits on a relaunch" |
| The rehearsal's desk | `scripts/live-run.sh` reuses `live-desk.sh`'s `phase_open` and `phase_rule` shapes: a `tmux` session `cc-desk-scratch-desk-<stamp>` runs the installed binary with no arguments and no `--skill-file`, so the installed `cc-desk` skill loads. It opens with `-c` set to the run folder under `$CLAUDE_JOB_DIR/tmp`, never the scratch checkout: the cleanup's `Pull` skips while a live row's `cwd` sits in the checkout outside `.claude/worktrees` (`internal/launch/worktree.go:129` to `:134`), and the branch would then stay. The run ends the desk right after the ruling line with `tmux kill-session -t cc-desk-scratch-desk-<stamp>`, run alone. `lib.sh`'s kill guards cover pass and conductor sessions only, and the desk's `cwd` is outside the scratch checkout. Prep also checks that `refs/heads/conductor` and `.claude/worktrees/conductor` are absent from the scratch checkout, since a left `conductor` branch makes the conductor's `Make` fail three times (`worktree.go:56` to `:57`, `launcher.go:107`). The script types one answer with `tmux send-keys -l` and checks `rulings.jsonl`. It saves `tmux pipe-pane` output and counts its typed entries | `scripts/live-desk.sh` `phase_open`, `phase_rule`, `type_in` |
| The rehearsal's pass | A release `run-<stamp>` with one round and one pass whose question the typed answer settles. The check starts the pass with the real `pass.txt` prompt, so the session loads `pass-core` and `cc-desk-work` from `~/.claude/skills/`, plans from the scratch base's `docs/`, adds one function and its test to `toy.go` and `toy_test.go`, runs the scratch gate, opens its pull request, and runs `pass ready`. The timed check merges after the scratch CI passes. This is the first full `pass-core` session the runs have launched, so its minutes are unmeasured. The toy plan says to start no background subagent or watch and to watch CI in the foreground, since a turn that ends to wait reads `done` and meets the idle rule. Task 8's first run measures it | `internal/launch/templates/pass.txt`; `~/Projects/cc-desk-scratch/scripts/gate.sh` |
| The scratch base | Prep cuts `run-base-<stamp>` from the scratch `main`, adds `docs/specs/` and `docs/plans/` for the toy change from `scripts/live/run/`, pushes it, records it as the release's default branch, and switches the scratch checkout to it, alone, after checking the checkout is clean. The cleanup pulls only a checkout on the default branch, and only a pulled checkout gets `git branch -d` (PR 21's miss). Close switches the checkout back to `main` | `internal/launch/worktree.go` `Remove`; `cmd/cc-desk/merge.go` cleanup |
| Gone within 10 minutes | A listing each minute from the merge for 10 minutes. The session is gone when its row is missing or reads `stopped`, `done`, or `failed`. The worktree is gone when `git worktree list` lacks its path, and the branch when `git branch --list` lacks it. The remote branch stays, since nothing deletes it | `docs/specs/2026-11-17-sessions-design.md:159` |
| The shared listing reader | `start_listing` moves from `scripts/live-repairs.sh` into `scripts/live/lib.sh`, which `live-repairs.sh`, `live-run.sh`, and `live-recover.sh` source. The kill-target guards are new code in the same file; the only `kill` in `live-repairs.sh` today ends the listing reader (`:744`). The listing fix lands in `lib.sh` | `scripts/live-repairs.sh:407` |
| The reports | A one-round release falls due for two reports in turn, the round's and then the release's (`ReportDue`), and the check starts the conductor once for each. Item 9's clause is checked on the first start: it records exactly one report, which the page lists. The script then waits for the release report, checks that it too is recorded once and listed, and marks both read with the page's request so the run ends with nothing unread | `decide.go:239`; design line 378; `[conflict]` under `## Decisions` |
| Marked read | `curl` posts to `http://127.0.0.1:7908/reports/<id>?/read`, form-encoded, with `Origin: http://127.0.0.1:7908`, as the page's button does. A missing or foreign origin gets 403 | `web/src/routes/reports/[id=report]/+page.server.ts` |
| Prompts and refusals | The release file's pass record holds `prompts` and `refusals`. The script prints the count of prompts (want 0) and each refusal's command. A list of none would pass the clause with nothing proved, so it also counts the rows of `~/.local/state/claude-denials.jsonl` whose session id is in the run's id set, and prints both counts. It passes when they are equal | `internal/state/types.go` `Session`; pass 5's refusal hook |
| No merge command | The script collects every session id the run's records hold, across every relaunch, from the release file and the check log, finds each transcript, and greps it for `gh pr merge`. It prints the ids it read | `live-check-cleanup-gaps.md` item 1 |
| One allow-rules source | `scripts/live/run/allow-rules.json` lists the rules with `{{repo}}` and `{{branch}}` in place of the repo and the pass's branch. The rehearsal fills them with `glw907/cc-desk-scratch` and its pass name, and the setup script with `glw907/cc-desk` and `stuck-sessions`. A check fills both and diffs them with those two values put back. The staging rule is `Bash(git add *)`, since the rehearsal's pass stages `toy.go` and `toy_test.go` and the real pass stages Go files under `internal/`; a rule naming one file can't serve both | Design "The rehearsal": "The scratch release carries the allow rules the real release will carry"; `scripts/live/desk/release.json` |
| The setup script | `scripts/setup-release.sh --state-dir <dir>` runs `release add` (project `cc-desk`, repo `glw907/cc-desk`, one round `stuck-sessions` with the design path and the filled allow rules), `round design` (one pass `stuck-sessions` at 60 minutes waiting on question `ss1`, `never_guess` true, its background, both answers' causes, the recommendation `yes`, and its reason, each from the design's section "The question"), and `release checkout cc-desk ~/Projects/cc-desk --default-branch main`. It refuses a folder that holds a live `cc-desk` release | `internal/state/events.go:79` to `:106` `Design`; `docs/specs/2026-11-18-stuck-sessions-design.md:24` to `:43` |
| Item 10's check | `scripts/setup-release-check.sh` waits, reading only, until `cc-desk-check.service` is inactive, so no check run is writing the folder, then copies `~/.local/state/cc-desk` to a folder under `$CLAUDE_JOB_DIR/tmp`, or under the system temp folder when that is unset, ends any `cc-desk-scratch` release on the copy with `release done`, marks the copy's reports read, runs the setup script on the copy, and prints one line for each check: one entry waiting for `cc-desk` with one question, `never_guess` true in the release file, the allow rules equal to the shared source, and `check --dry-run` deciding nothing. Only `check --dry-run` runs on the copy, since the copy's checkout is the real `~/Projects/cc-desk` | Round plan, item 10's acceptance |
| The listing fix | `start_listing` skips the restart when `listing.out` already holds its ten reads, and never truncates the file once the merge phase has written to it | `docs/follow-ups/live-listing-restarts-after-merge.md` |
| The cleanup gaps | In `scripts/live-check.sh`: item 1 collects every session id from the check log and the release file; item 2 ends a stopped pass's live session through `session end`; item 3 waits, reading only, until `cc-desk-check.service` is inactive before it carries out the stops; item 4 adds a trap to `prep` that prints the state and puts the timer back as `abort` does; item 5 adds a trap to the `scope` phase that stops the probe's scope. Item 6 stays filed | `docs/follow-ups/live-check-cleanup-gaps.md` |
| The installed build | Each live run's prep runs `timer install` from the branch build, so the timer and the page use it. Once a run writes a report with a release name, `main`'s v0.1.0 build can't load `reports.json` (`DisallowUnknownFields`), so for this pass each script prints a warning in place of the line that puts `main`'s build back: that line is unsafe until the pass lands. The conductor reinstalls `main` after the pass lands (answer 24), and that `main` holds the field. Task 12 says so in the PR | `cmd/cc-desk/timer.go`; `scripts/live-desk.sh` `phase_prep` |
| The unit's `PATH` | `cc-desk-check.service` sets `PATH=%h/.local/share/cc-desk/bin:%h/.local/bin:/home/linuxbrew/.linuxbrew/bin:/usr/local/bin:/usr/bin:/bin`. Under it, read on 2026-11-20: `claude`, `gh`, `git`, and `go` resolve to `/home/linuxbrew/.linuxbrew/bin/`, `systemd-run` to `/usr/bin/systemd-run`, and `cc-desk` to `~/.local/share/cc-desk/bin/cc-desk`. Each script's prep checks the same with `env -i PATH=<unit PATH> bash -c 'command -v <cmd>'` | `systemctl --user show cc-desk-check.service -p Environment` |

## Recorded searches

Run in the worktree on `d7279c7`, 2026-11-20:

- `ls cmd/cc-desk/`: `release.go` and `session.go` hold the two verb groups.
- `grep -rln stopSession cmd internal e2e`: `cmd/cc-desk/merge.go` and `carryout.go`.
- `grep -rn 'releases/done' internal cmd scripts docs/specs`: read only, in `internal/state/read.go:180`, `internal/state/types.go:528`, and the tests `internal/state/read_test.go:130` and `internal/snapshot/estimate_test.go:114`, plus the State and rules design at line 100. Nothing writes there.
- `grep -n 'func hasReport' internal/sessions/decide.go` and `report.go:137`: both match a report on project and round name alone.
- `grep -rln 'live-check.sh\|live-repairs.sh\|live-desk.sh'`: the three scripts, five plans, and the follow-ups `live-listing-restarts-after-merge.md`, `live-scripts-start-a-conductor.md`, `live-check-cleanup-gaps.md`, and `live-abort-waits-on-stopped-sessions.md`. No Go test or gate step runs a live script.
- `ls e2e/`: 40 test files and `testdata/`. `conductor_test.go`, `launch_test.go`, `events_test.go`, `recovery_test.go`, and the launch rig (`launch_rig_test.go`, `standin_test.go`) cover the `session` and `release` verbs and the stand-in `claude`.
- `grep -n 'IdleLimit\|RelaunchLimit' internal/sessions/decide.go`: `IdleLimit = 15 * time.Minute` (line 37), `RelaunchLimit = 2` (line 40).
- Go: `go.mod` reads `go 1.27` and requires `github.com/spf13/cobra v1.10.2`; `go version` prints `go1.27.4 linux/amd64`. No new module.

## File list

| File | Change | Task |
| --- | --- | --- |
| `cmd/cc-desk/release.go` | The `done` verb | 1 |
| `internal/state/events.go`, `internal/state/commit.go` | The event that moves a release, and the commit's removal of the old file | 1 |
| `internal/state/types.go`, `internal/state/report.go`, `internal/sessions/decide.go` | The release name on a report, and the match on it | 1 |
| `cmd/cc-desk/report.go`, `cmd/cc-desk/carryout.go`, `internal/snapshot/snapshot.go` | The match in the `report` verb, the carry-out, and `buildReports` | 1 |
| `internal/state/*_test.go`, `internal/sessions/*_test.go`, `internal/snapshot/*_test.go`, `e2e/release_done_test.go` (new) | Their tests | 1 |
| `cmd/cc-desk/session.go` | The `end` verb, calling `stopSession` | 2 |
| `e2e/session_end_test.go` (new) | Its test, on the launch rig's stand-in `claude` | 2 |
| `scripts/live/lib.sh` (new) | The shared listing reader and kill-target guards | 3 |
| `scripts/live-repairs.sh` | Sources `lib.sh`; `abort` ends sessions through `session end` | 3 |
| `scripts/live-check.sh` | Cleanup gaps 1 to 5 | 3 |
| `scripts/live/recover/work-plan.md` (new), `scripts/live/probe/release.json` (new) | The fixture that keeps working, and Task 4's probe release | 3 |
| `docs/research/2026-11-20-kill-probe.md` (new) | Task 4's record | 4 |
| `scripts/live/run/allow-rules.json` (new) | The shared allow rules | 5 |
| `scripts/setup-release.sh`, `scripts/setup-release-check.sh` (new) | The setup script and item 10's check | 5 |
| `scripts/live/run/` (new): `release.json`, `design.json`, the toy design and plan under `docs/` | The rehearsal's fixtures | 6 |
| `scripts/live-run.sh` (new) | The rehearsal | 6 |
| `scripts/live/recover/`: `release.json`, `merge-plan.md`, `ask-plan.md` (new) | The recovery fixtures; the pass that dies runs `work-plan.md` | 7 |
| `scripts/live-recover.sh` (new) | The recovery checks | 7 |
| `README.md` | The two verbs, the two live scripts, and the setup script | 9 |
| `docs/follow-ups/` | The files Task 9 removes, rewords, and adds | 9 |
| `docs/STATUS.md` | A stop only, under `## Stop` | any |

## Acceptance

Done items 8, 9, and 11, word for word from the design, and item 10 as the round plan restates
it:

> 8. **Recovery on the page.** On the workstation, with the real conductor: its process is
> killed, it is running again within 10 minutes, and the page shows no stop. A pass session is
> killed three times. After the second relaunch the page reads "Needs you now" with the round's
> name, and the session list never shows two sessions for that pass. A session launched under a
> release that names one command as always asking reads "Needs you now" within 10 minutes, and
> the line names the command.

> 9. **The rehearsal.** In the scratch repo, from one typed answer on a made terminal: the pass
> starts, opens a pull request, and the timed check merges it. Within 10 minutes of the merge
> the session, worktree, and branch are gone. The conductor starts by itself and records one
> report, which the page lists. The release file records no permission prompt for the pass and
> lists each command auto mode refused. The transcripts hold no merge command. Nothing was typed
> after the answer.

> 10 (round plan). "the release setup script, run against a copy of the state folder, yields a
> snapshot with one question waiting for the project `cc-desk`, and that the rehearsal passed
> with the allow rules that release carries"

> 11. **The gate.** `bash scripts/gate.sh` passes with no skipped test. Items 2 to 5 and 7 to
> 10 have parts that run live and aren't Go tests. Each live script prints one line for each
> check and ends with its result. A pass's pull request holds that output from a run on the
> pass's last commit, and a pass that leaves it out isn't done.

### Clause table

| Clause | Task | Command | Pass when |
| --- | --- | --- | --- |
| 8: the conductor is killed and running again within 10 minutes | 4, 7, 8, 11 | `bash scripts/live-recover.sh` | The kill line, then a listing each minute shows a live conductor row within 10 minutes, with the mechanism named |
| 8: the page shows no stop | 7, 8, 11 | the same | The page at `http://127.0.0.1:7908/`, fetched each minute, shows no stop for the conductor |
| 8: a pass killed three times; "Needs you now" with the round's name after the second relaunch | 4, 7, 8, 11 | the same | For each death in turn: a check saw the row `working` after the previous launch, then the death line, then its decision. Two counted relaunches, then `relaunch_limit`, and the page's text holds "Needs you now" and the round's name |
| 8: never two sessions for that pass | 7, 8, 11 | the same | Every saved list holds at most one live row for the pass, by the mechanism table's row key |
| 8: the ask reads "Needs you now" within 10 minutes and names the command | 7, 8, 11 | the same | The page's line holds "Needs you now" and `cksum ask-probe.txt` within 10 minutes of the launch |
| 9: from one typed answer, the pass starts | 6, 8, 11 | `bash scripts/live-run.sh` | One ruling line, then a `start` decision for the pass |
| 9: it opens a pull request and the timed check merges it | 6, 8, 11 | the same | A `merged pull request <n>` outcome from a run of `cc-desk-check.service`, and no merge by the script |
| 9: within 10 minutes the session, worktree, and branch are gone | 3, 6, 8, 11 | the same | The minute-by-minute listing, from the merge, shows all three gone |
| 9: the conductor starts by itself and records one report, which the page lists | 1, 6, 8, 11 | the same | A `launch` `report_due` decision, then exactly one report recorded by that conductor, listed on the page |
| 9: no permission prompt for the pass; each refused command listed | 6, 8, 11 | the same | `prompts` 0 on every record of the pass; each refusal printed, and the record's refusal count equals the denials log's count for the run's ids |
| 9: no merge command in the transcripts | 3, 6, 8, 11 | the same | The ids read and 0 matches |
| 9: nothing typed after the answer | 6, 8, 11 | the same | One typed entry in the pane log |
| 10: the setup script on a copy gives one question waiting for `cc-desk` | 5, 8, 11 | `bash scripts/setup-release-check.sh` | Each of its lines reads ok |
| 10: the rehearsal passed with the allow rules that release carries | 5, 6, 8, 11 | the same, and `live-run.sh`'s rules line | The two filled rule lists match |
| 11: the gate passes with no skip | 10 | `bash scripts/gate.sh` | `gate: ok` |
| 11: each live script prints one line per check and its result; the PR holds the output from the last commit | 8, 11, 12 | the three scripts | Each ends with one result line; the PR holds the output whole |
| Each run ends with no question waiting, no report unread, and no session running (answer 21) | 6, 7, 8, 11 | each script's `close` | Its snapshot line reads 0, 0, and 0, and `release done` moved the run's release |

### Pre-code runs on `d7279c7`

- Item 11: `bash scripts/gate.sh` in the worktree, run bare on 2026-11-20. It printed
  `gate: gofmt`, `gate: go vet`, `ok` for each Go package (`e2e` in 69.8 s), `gate: vale`,
  `gate: skills`, the web steps, `175 passed (2.0m)`, and `gate: ok`. The gate fails on a
  skipped test, so none skipped.
- Item 10, by hand on a copy of the state folder (`cp -a` to `$CLAUDE_JOB_DIR/tmp/state-copy`).
  The 6 reports were marked read with `report read`, and the leftover release file was moved
  aside on the copy, since no verb ends a release yet. Then a hand-written `cc-desk` release
  ran through the three verbs the setup script will run:

  ```text
  added release cc-desk-stuck-sessions of cc-desk
  recorded the design of cc-desk/stuck-sessions
  recorded the checkout of cc-desk
  ```

  The snapshot read `waits_for_geoff` `[{"project":"cc-desk","round":"stuck-sessions","questions":1}]`,
  with 0 needs, 0 sessions, and 0 unread. The release file kept `ss1` with `never_guess` true,
  the round's `design_written` true with the design path, and the pass at 60 minutes,
  `not_started`. `check --dry-run` printed nothing and exited 0. Its allow rules were a
  placeholder, so the rules clause can't pass until Task 5.
- Items 8 and 9 can't run on today's code: neither script exists, and the leftover live
  release makes `release add` refuse any new scratch release. The real snapshot at 08:36 UTC
  read 15 needs, 0 waits, and 6 unread.

## Tasks

Lane A is Tasks 1 and 2, and lane B is Tasks 3 and 5. Each lane runs as a pair per
`executor.md` "Pairs" (Sonnet, medium, descriptions `desk-run task <n>`). Task 4 is the
executor's, once both lanes (Tasks 1, 2, 3, and 5) are on the branch. Tasks 6 and 7 start once Task 4 has settled
standing answer 70's branch, Task 6 on lane B and Task 7 on lane A. Then the executor runs
Tasks 9, 10, 8, 11, and 12, in that order. The docs and the review come before the live runs,
so a clean Task 8 run is already on the last commit.

### Task 1: a release ends, and its reports are its own (55 minutes)

Class `runner`. Outcome: `cc-desk release done <project>` moves the live release under
`releases/done/`, and the snapshot stops listing it as live. A report carries its release's
name, so a second release of one project falls due for its own release report and can record
it. Files: the first five rows of the file list. Check: a state test moves a release and reads
both folders; an e2e test runs the verb and then `release add` of a new release for the same
project, which now succeeds. The verb refuses with no live release, and refuses while the
stand-in list shows a live row for one of the release's records, but not a `done` one. It
waits while `check.lock` is held, which a test shows. A decide test gives two releases of one
project in turn, each falling due for its release report once. A snapshot test with two
releases of one project lists each report's passes from its own release. An old report with
no release name keeps its match. Test: `go test ./internal/state/ ./internal/sessions/
./internal/snapshot/ ./e2e/ -run 'Release|Report'`.

### Task 2: a session ends through the binary (35 minutes)

Class `runner`. Outcome: `cc-desk session end pass <project> <name>` and `cc-desk session end
conductor <project>` end that record's session through `stopSession` and print its step.
Files: `cmd/cc-desk/session.go`, `e2e/session_end_test.go`. Check: on the launch rig, the verb
stops a working row and prints "stopped session <id>"; on a row already stopped it prints
"already gone"; with no open session on the record it refuses and runs no stop; it waits while
`check.lock` is held; a following check reads the ended row as the mechanism table's "The
check reads a death" row says. Test: `go test ./e2e/ -run 'SessionEnd'`.

### Task 3: the live scripts' fixes (40 minutes)

Class `runner`. Outcome: `scripts/live/lib.sh` with the listing reader and the new guards, the
listing fix in it, `live-repairs.sh` sourcing it and ending sessions through `session end` in
`abort`, cleanup gaps 1 to 5 in `live-check.sh`, and the two fixtures in the file list's
Task 3 row, as the mechanism table's rows say. Check: `bash -n` on the three files. The
dirty-checkout refusal is proved live in Task 8, which runs each new script's `prep` once with
an untracked file in the scratch checkout, then removes it. The listing fix shows in Task 8's
rehearsal, which runs the shared reader. Test: `bash -n scripts/live/lib.sh
scripts/live-repairs.sh scripts/live-check.sh`.

### Task 4: install, clear the leftovers, and probe a death (60 minutes)

Class `runner`, live. Order, each step run alone:

1. Build from the branch and run `timer install`. Check the unit's `PATH` as the table says.
2. Run `setup-release-check.sh` once on its copy, now that `release done` is on the branch.
   Then mark `rep-0001` to `rep-0006` read with `cc-desk report read`, and run `cc-desk
   release done cc-desk-scratch`. The snapshot reads 0 needs, 0 waits, and 0 unread.
3. Cut `probe-base-<stamp>` from the scratch `main`, copy `work-plan.md` into its
   `docs/plans/` as `live-desk.sh` does at line 212, push it, and add the probe release from
   `scripts/live/probe/release.json` on that base, which carries the gate's allow rule. Wait
   for the check to start the pass, then for a later check to see its row reading `working`,
   with `last_seen` later than `launched`. Record a gate step's minutes.
4. Print the pass's PID through the guards, then run `kill <pid>` alone. Save the list every 5
   seconds for 90 seconds, then once a minute through the next check run. Record the row's
   state, PID, and session id at each read, and the check's decision.
5. If the session is live, wait for a check to see it, then run `session end` on it. Save the
   list each minute through the next check runs, until the check decides, and record the row's
   state and the decision: a counted `missing` relaunch, or a `done` row waiting out the idle
   rule to `end_and_relaunch`. Stop reading after 30 minutes. A `session end` refused twice
   is met here first (answer 72).
6. End any session the probe release still holds through `session end`, then `release done
   cc-desk-scratch`, then switch the scratch checkout back to `main`.

The record goes into `docs/research/2026-11-20-kill-probe.md`. It settles standing answer 70's
branch for Tasks 7, 8, and 11, and the minutes a death costs. Test: the saved lists and the check
log lines in that file.

### Task 5: the allow rules and the real release (35 minutes)

Class `runner`. Outcome: the shared allow-rules source, the setup script, and item 10's check,
as the mechanism table says. The first rules are the scratch fixture's six, generalized, with
`Bash(git add *)` as the staging rule, plus the test command the real pass will run in
cc-desk (`go test ./...`). Each rule names one command the pass's steps run. Check: `bash -n`
on both scripts in the pair. The live check runs `release done`, which is lane A's, so the
executor runs it in Task 4 step 2 and again in Task 11. Test: `bash -n scripts/setup-release.sh
scripts/setup-release-check.sh`.

### Task 6: the rehearsal (60 minutes)

Class `runner`. Outcome: `scripts/live-run.sh` with phases `prep`, `open`, `rule`, `pass`,
`merge`, `listing`, `report`, `evidence`, and `close`, and a `run` that chains them. Each phase
prints one line per check, and the script ends with one result line. `prep` refuses a dirty
scratch checkout, any live `cc-desk-scratch-` session, or a left `conductor` branch or
worktree. `rule` ends the desk with `tmux kill-session` on its recorded name, alone, right after the
ruling line. `close` ends any session the run's records hold through `session end`, marks every
report read, runs `release done`, switches the scratch checkout back to `main`, and prints the
snapshot line. Check: `bash -n`, then the live run in Task 8. Test: `bash -n
scripts/live-run.sh`.

### Task 7: the recovery checks (45 minutes)

Class `runner`. Outcome: `scripts/live-recover.sh` with phases `prep`, `conductor`, `pass`,
`ask`, `evidence`, and `close`, as the three recovery rows of the mechanism table say, drawing
each death in the form standing answer 70 settled. Before each pass death the phase waits for
a check to see the relaunched session. A phase that needs a `kill` prints the PID and stops,
and the executor runs the `kill` alone and then the phase's `--after` step. Check: `bash -n`,
then the live run in Task 8. Test: `bash -n scripts/live-recover.sh`.

### Task 9: follow-ups and words (10 minutes)

Remove `live-listing-restarts-after-merge.md` and `live-abort-waits-on-stopped-sessions.md`.
Reword `live-check-cleanup-gaps.md` to hold item 6 alone. Reword item 1 of
`leftover-live-session-parks-a-pass.md` to say `session end` exists and the Move design settles
whether the desk or the page offers it. File one follow-up on the two reports a one-round
release makes, for the 1.1 round. Add the two verbs and three scripts to `README.md`. A
finding Task 8 files gets its own follow-up there, under standing answer 81.

### Task 10: fold the lanes, review, gate (35 minutes)

One whole-branch `standards:diff-reviewer` read (its frontmatter model, medium effort), folded.
It is the pass's one whole-branch read (standing answer 81). `bash scripts/gate.sh`, bare.

### Task 8: the live runs and the fixes they show (215 minutes)

Class `runner`, live. Run `live-run.sh run`, then `setup-release-check.sh`, so its rules line
compares the final lists, then `live-recover.sh` phase by phase. Each
finding is fixed inside the design with a test, or filed, by standing answers 18, 19, 71, 75,
and 81. A rule a run needed goes into the shared source and the scratch release (`cc-desk
rules add`). Priced at 60 minutes for the rehearsal, 5 for the setup check, 125 for the recovery, and 25
for fixes.
The rehearsal waits up to 5 minutes for the timer before each of the pass's start, its merge,
and the two conductor starts, about 15 for the pass's build and CI, and 10 for the listing.
The recovery takes about 15 to reach the merge, 10 for the conductor's watch, about 25 for
each of three deaths when an ended row waits out the 15-minute idle rule (about 10 each when a
kill gives a counted `missing`), 10 for the ask, and 15 for prep and close. A run with no fix
after it is on the last commit, and its output is item 11's evidence.

### Task 11: the live runs again after a fix (190 minutes)

Only when Task 8 ended with a fix. `setup-release-check.sh` (5 minutes), `live-run.sh run`
(60), and `live-recover.sh` (125) on the PR's last commit, with the output kept whole, under
standing answer 82. Priced in full, since a first live run usually shows a fix.

### Task 12: the PR (15 minutes)

The body as the executor prompt says, with every clause's output. It names the commit each
script's output came from, and says the line that puts `main`'s build back is unsafe until the
pass lands.

### The path

| Step | Minutes |
| --- | --- |
| Lane A (Tasks 1 and 2, 90) beside lane B (Tasks 3 and 5, 75) | 90 |
| Task 4 | 60 |
| Tasks 6 and 7 beside each other | 60 |
| Tasks 9 and 10 | 45 |
| Task 8 | 215 |
| Task 11 | 190 |
| Task 12 | 15 |
| Total | 675 |

## The gate and CI

The gate is `bash scripts/gate.sh`, run bare with the Bash timeout at 600000, before every
commit. Stage a new file first: the gate lints tracked files only. After each push, run `gh run
list --branch desk-run --limit 3`, then `gh run watch <id> --exit-status` on the newest run for
the pushed commit, in the background. A red run is fixed before any other task.

## Standing answers

The round plan's 13 to 38 hold. These add to them.

- **70. What a death is.** Task 4 settles it. Every pass death, in either form, waits until a
  check has seen the session since its last launch (`last_seen` later than `launched`), or it
  gives an uncounted `launch_again` and then `launch_failed`. If the killed row goes missing or
  reads `stopped` or `failed` and the check relaunches it, every pass death in Tasks 8 and 11
  is a lone `kill`. If the row comes back under a new PID, or stays listed `working` with no
  PID through a check run, or the `kill` is refused twice 60 seconds apart, the conductor's kill
  still runs as the clause says, with its mechanism named. The pass's three deaths and every
  teardown then go through `session end`, in the form step 5 of Task 4 showed the check
  relaunching. That reading of "killed" is the park PR's question for Geoff (decision `[ask]`).
  Until a ruling says otherwise, the executor takes the recommended reading and the PR lists it
  under `Calls:`, with the kill output beside it. Nothing stops for it. If neither a kill nor
  `session end` gives a death the check relaunches, item 8's pass clause can't be drawn: write
  it under `## Stop` with the probe's record and end.
- **71. A command the run needed a rule for.** Answer 18: end the stuck session through
  `session end`, file the prompt, add the rule to the shared source and to the scratch release
  with `cc-desk rules add`, and run once more. A third run that meets such a prompt waits for
  Geoff under answer 38, while the work that doesn't rest on it goes on.
- **72. A refused step the executor typed.** Wait 60 seconds and retry once, citing the round's
  entry. Then use the table's alternative. Never change a setting, hook, or rule (answer 38).
  A refusal of `release done` on the leftovers, twice, is a stop: write it under `## Stop`
  with the recommendation that the conductor run that one verb, and end. A refusal of `session
  end`, twice, is the same stop, recommending that the conductor end the session (answer 33),
  since then no teardown can end a session and the next run can't add its release. Task 4 step
  5 meets it first.
- **73. A classifier outage.** Wait 60 seconds and retry once.
- **74. A skill needs a fix.** Make it on the branch and name it in the PR body's first
  paragraph (answer 23). The rehearsal's desk can prove a `cc-desk` skill fix with
  `--skill-file`. A `cc-desk-work` fix can't be proved live before the conductor installs
  `v0.1.1`, so the PR lists it as a call, inside the design.
- **75. The rehearsal shows cc-desk broken.** A fix the design already describes is made with a
  test and listed. A fix that would change what the design says stops for Geoff (answer 26's
  last line).
- **76. A usage limit or a long wait.** The check holds relaunches past the stop line by
  design. The scripts' waits are long enough for two check runs past the expected step. A run
  that times out on a wait the check was holding is run again, not counted as a finding.
- **77. A session the run can't trace.** Answer 33: the conductor's. The executor ends only
  sessions on the records its own runs made.
- **78. The page doesn't answer.** The scripts fall back to the snapshot for evidence and say
  so. The page's unit is the conductor's to restart, so a page down for a whole run is a stop.
- **79. The conductor finishes before the kill.** The kill runs within the conductor's first
  minute, before any check has seen it, so `launch_again` is its usual way back. If its report is already recorded, the run kills the release report's conductor
  instead, or reruns the phase with a fresh release.
- **80. The executor's time.** Past twice the planned minutes, finish the task in flight, push,
  and write the stop. Answer 38's relaunch follows.
- **81. Findings after the review.** Task 10's read is the pass's one whole-branch review. A
  fix after it gets the executor's own read of its diff and a green CI run. Its Low findings are
  folded once or filed under `docs/follow-ups/`, and never start a second review round.
- **82. What a fix reruns.** A change under `cmd/` or `internal/` changes the binary every
  script runs, so all three scripts run again on the new last commit. A change to one script or
  its fixtures reruns that script. The PR then names the commit each output came from, with
  `git diff --stat <that commit> HEAD` showing that no file the other scripts run has changed.
  That diff-backed reading of item 11's "a run on the pass's last commit" is the park PR's
  second question for Geoff. Until a ruling says otherwise, the executor takes it and lists it
  under `Calls:`.

## Decisions

- `[conflict]` The conductor in the scratch project is `cc-desk-scratch-conductor`. The binary
  names a conductor `<project>-conductor` (`cmd/cc-desk/launcher.go:107`,
  `internal/sessions/decide.go:283`), so the launch prompt's `cc-desk-conductor` is read as
  that.
- `[conflict]` Item 9 says the conductor "records one report". The design's rule makes a
  one-round release due for two in turn, the round's and the release's (design line 378,
  `ReportDue`), and the real run will meet the same. Pass 5's six leftover reports are all
  round reports, because its release held rounds that never finished, so no run has yet shown
  both. The rehearsal checks the clause on the conductor's first start, then records and reads
  the release report too. A follow-up asks the 1.1 round whether a one-round release needs
  both. The park PR names this call.
- `[advisor]` `release done` and `session end` stay in the pass. `release done` carries out the
  State and rules design's sentence that a release's file moves to `releases/done/` when it
  ends, and verb names are the builder's (design, "Left to the builder"). Its live test is the
  Sessions design's live row. `session end` carries out the round entry's "end those sessions
  by the binary", ends only a session on a record, and answers two filed follow-ups. The Move
  design still settles whether the desk or the page offers it to Geoff.
- `[advisor]` Item 8's deaths are probed before they are scripted, with `session end` as the
  fallback, and Task 4 proves the fallback too. `doneRow` gives `stop` `relaunch_limit` to a
  session at the limit (`decide.go:547` to `:548`), so a third death drawn through a `done` row
  still reaches "Needs you now".
- `[ask]` Item 8 says a pass session "is killed three times". If Task 4 shows a kill always
  comes back under the supervisor, the three deaths are drawn with `session end` instead. Then
  each death is a stop, a `done` row, and the idle rule. The clause's outcome ("Needs you now"
  after the second relaunch, never two sessions) stays as written, and the Sessions design
  counts a stopped row as ended. Because the word "killed" is the design's, the park PR asks
  Geoff whether that reading holds. Recommended: yes. The executor takes it until a ruling
  comes and lists it under `Calls:`.
- `[ask]` Item 11 asks for each script's output "from a run on the pass's last commit". Under
  answer 82, a fix to one script reruns only that script, and the PR shows by diff that nothing
  the other two run has changed. The park PR asks Geoff whether that reading holds.
  Recommended: yes, since a full rerun after each script fix costs about 190 minutes. The
  executor takes it until a ruling comes and lists it under `Calls:`.
- `[conflict]` Item 8 says "the session list never shows two sessions for that pass". The
  check counts a live row, so the plan reads it as at most one live row: a stopped row stays
  listed as `done`, and the list would otherwise always show two after a relaunch. Proof 6's
  reading is the precedent. A row is the pass's by name, short id, session id, or a `cwd` under
  its worktree.
- `[planner]` A report carries its release's name. Without it, the rehearsal's release report
  would make the recovery release read as reported, and the real project's second release could
  never record its own. This carries out design line 378, so it is inside the design.
- `[planner]` The ask rule is `Bash(cksum ask-probe.txt)`, a command no session runs
  unprompted, so the merge and working passes of the same release never block on it.
- `[planner]` The staging rule in the shared source is `Bash(git add *)`, so the rehearsal and
  the real pass carry one list.
- `[planner]` "The real conductor" is a conductor the binary launched for the scratch release,
  never `conductor-cc-desk`. The recovery run has its own merged round for it, so the
  rehearsal's conductor is never killed.
- `[planner]` A kill's target is the row whose `sessionId` is the record's `session_id`,
  guarded by name, folder, process name, and a command line with no `--bg-spare`, since the
  record holds no PID. The planner ran no kill: the round's entry names the executor.
- `[planner]` The rehearsal never starts `cc-desk-check.service` by hand. The timed check does
  each step, so the run shows the timer working.
- `[planner]` "Branch gone" is the local branch. Nothing deletes the remote one, and the scratch
  repo's `delete_branch_on_merge` is false.
- `[planner]` The plan prices two runs of each live script and trims nothing.
- `[review]` The `standards:plan-reviewer` read (Opus, medium, 2026-11-20) found the hold
  fixture can't prove a death: it replies `ready` and goes idle, so the idle rule relaunches it
  with no death. Folded: the pass that dies runs `work-plan.md`, which keeps working, and each
  death's evidence is its order in the check log.
- `[review]` A death before any check sees the relaunched session gives an uncounted
  `launch_again`, then `launch_failed`. Folded: answer 70 and Tasks 4 and 7 wait for `last_seen`
  before each pass death, and the conductor row names `launch_again` as its usual way back.
- `[review]` The plan called its scope grown, which answer 35 excludes. Folded: the two verbs
  and the report fix carry out sentences the designs hold, and the header says so.
- `[review]` Both new verbs take `check.lock`, with a test each. The report fix reaches the
  `report` verb, the carry-out, and `buildReports`, with a tag rule and a snapshot test. The
  scripts warn that `main`'s build can't read the folder until the pass lands.
- `[review]` The desk opens outside the scratch checkout, so the cleanup's pull isn't skipped,
  and prep checks for a left `conductor` branch. Task 4 waits for both lanes, and item 10's
  live check runs once `release done` is on the branch.
- `[review]` Ending a pass by `session end` costs about 25 minutes a death, so Task 8's recovery
  is 125 and the probe's cutoff 30. The refusals clause cross-checks the denials log. Answer 72
  covers a refused `session end`. Answers 81 and 82 set the review floor and what a fix reruns,
  and Tasks 9 and 10 move ahead of the live runs. The rehearsal's toy plan runs no background
  watch.
- `[review]` The minor findings are folded: the cites, the `--bg-spare` guard, the page's URL,
  the kind-first `session end` shape, the copy under the check's lock, the unit writes in the
  pre-flight, and the dirty-checkout refusal proved live in Task 8.
- `[advisor]` The final advisor read found four gaps, all folded. The reviewer's `sleep 50`
  fixture is blocked by Claude Code's Bash tool, so the working pass repeats the scratch gate. Task 4 waits
  for both lanes, since step 2 runs Task 5's script. Task 8 runs `setup-release-check.sh`, so
  item 10 has output on the last commit when no fix follows. The desk ends through `tmux
  kill-session`, outside the kill guards. Task 8's fix minutes absorb the setup check, so the
  path stays at the 675 the conductor raised to.

## Rulings

- **Estimate raised to 675 minutes (conductor-cc-desk, 2026-11-20, standing answer 35).** The
  plan parked only because its path is over 180 minutes. Its scope matches the designs: `release
  done` carries out the State and rules design's move to `releases/done/`, `session end` carries
  out the round entry's "end those sessions by the binary", and the report's release name carries
  out design line 378. The planner trimmed nothing. The executor's stop is 1350 minutes, and
  answer 38 gives one relaunch from the branch with its stop at 2025.
