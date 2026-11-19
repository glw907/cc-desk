# Desk open: pass plan

Initiative: cc-desk
Round: desk
Pass: desk-open

Class `runner`. Planned 100 minutes. The path below is 255 minutes, gate, review, and CI time
included, over twice the plan (200), so the pass parks at planning under the round's standing
answer 14. The stop is in the draft pull request's body, and the plan launches as written once
Geoff raises the estimate or splits the pass. Design:
`docs/specs/2026-11-18-desk-skills-design.md`, sections "The desk opens", "Rulings", "Left to the
builder", and done items 3, 4, 5, and 11. The design wins over this plan. Round plan:
`docs/plans/2026-11-18-desk-skills-round.md`. Where the round is silent, the main,
State-and-rules, and Sessions designs under `docs/specs/` hold. Anything a person reads (the PR
body, STATUS lines, README lines, the guide, follow-ups, the skill) follows
`/var/home/glw907/.claude/docs/voice/written-for-people.md`, loaded through the `writing-voice`
skill before drafting.

Worktree `/var/home/glw907/Projects/cc-desk/.claude/worktrees/desk-open`, branch `desk-open`, cut
from `origin/main` at `81afc1d`. PR base: `main`. Push: `glw907/cc-desk`. Pass 1
(`desk-records`, PR 17) built the records, the check log, the `brief` verb, and the ruling guard.
Pass 2 (`desk-skills`, PR 18) wrote `skills/cc-desk/SKILL.md` and `docs/guide.md`. Pass 4
(`desk-page`) may build beside this pass. Everything under `web/` and the page's unit is pass 4's
(standing answer 17). This pass owns `cmd/`, `internal/`, `skills/`, `e2e/`, `scripts/live-desk.sh`,
`README.md`, `docs/guide.md`, and the follow-ups named in Task 8. Pass 5 (`desk-check`) follows
this pass and owns `internal/launch/settings.go`'s hooks, so the desk's settings writer lives in
`internal/desk/`, not in `internal/launch/`.

## Pre-flight

- Permissions checked: the denials log `~/.local/state/claude-denials.jsonl`, 136 rows, last
  2026-11-18T23:53:28Z. Three rows are new since the round plan read 133 rows to 19:38Z: two classifier
  outages and one `claude stop d5647ebf` judged Destructive at 21:28Z. Counts for this pass's
  commands: `script -q` 0, `tmux` 0, `setsid` 0, `ln -s` 0, `go build` 0, `gate.sh` 0,
  `systemctl` 0, `systemd-analyze` 0, `timer install` 0, `claude --session-id` 0,
  `gh pr create` 0, `gh pr comment` 0, `claude agents --json` 1 (piped to python, 2026-11-17),
  `claude stop` 6, `kill` 12, `gh pr close` 1 (another repo, 2026-11-07), `gh pr view` or
  `gh pr checks` 2 ("Merge Without Review"), `git push` 4 (none on a cc-desk feature branch),
  `cc-desk` rows are classifier outages only. The table below gives the form each refused kind
  takes. No step needs Geoff.
- Reviewed: one `standards:plan-reviewer` read (Opus, medium) on 2026-11-18, six blocking findings
  and nine others, all folded; its line is under `## Decisions`.
- Evidence: the pre-code acceptance runs under `## Acceptance`, the no-terminal probe under
  "Mechanism table", and the fixture survey under "Recorded searches".

### Refused commands and their alternatives

| Kind | Form this pass uses |
| --- | --- |
| `claude stop` | Only through the `cc-desk` binary (the check's carry-out, or the live script's abort phase), never typed by the executor. If one must be typed, it runs alone, never chained |
| `kill` | Only of a process id the script recorded, run alone. Never `pkill -f` |
| `claude agents --json` | Redirected to a file, then read. Never piped |
| `gh pr close` | Only the check's binary calls it, on a `glw907/cc-desk-scratch` pull request |
| `gh pr view`, `gh pr checks` | Only the check's binary reads checks. The executor reads CI with `gh run list` and `gh run watch` |
| `git push` | `git push origin desk-open` from the worktree, and the live script's push of its own scratch branch to `glw907/cc-desk-scratch`. Never main, never `--force` |
| Settings edits | None. No `.claude/settings*.json` change, no `autoMode` entry |

## Mechanism table

| Clause | Mechanism | Source |
| --- | --- | --- |
| Bare `cc-desk` on a terminal opens the desk | The root command gets `RunE` and `Args`. With no `args`, stdin and stdout must both be terminals (`TCGETS` through `syscall.Syscall(SYS_IOCTL, ...)`, since `/dev/null` is a char device too and a mode check passes it), and the process must have a controlling terminal (`tty_nr` in `/proc/self/stat` is not 0). The second test keeps a desk under `setsid` from opening, so no desk opens that can't record | `cmd/cc-desk/root.go` has no `RunE` today; pre-code run below; design "Left to the builder": "no desk opens that can't record" |
| `cc-desk <project> <round>` | `Args` accepts zero or two words. Two words must name a live release's project and one of its rounds, or the command refuses with the names it knows, so a mistyped verb never opens a desk | cobra's `legacyArgs` answers "unknown command" for a root with subcommands and no `Args`; pre-code run below |
| No terminal: help, nothing registered | When either test fails, `cmd.Help()` and exit 0, before any state read or write. The rig runs it with `</dev/null` and under `setsid` on a `pty` | Design "The desk opens" |
| A state file that fails to parse | `state.Load` runs first. On an error the command prints it, exits 2, and writes nothing | Done item 3 |
| The brief | `brief.Build` (bare) or `brief.BuildRound` (round form), placed in the system prompt by `--append-system-prompt`, never on screen | `internal/brief/brief.go:44`, `:64`; PR 18's `scripts/live/skills/system.txt` is the proven form |
| The skill | The system prompt says to load the `cc-desk` skill. A hidden `--skill-file <path>` replaces that with "read the file at `<path>` with the Read tool before your first reply", as `system.txt` does | Design "The desk opens", last paragraph |
| The session id and registration | A new UUID v4 from `crypto/rand`. `desk.Entry(id, os.Getpid())` and `state.RegisterDesk` in one `state.Commit`. A second `cc-desk` registers a second desk; `RegisterDesk` appends entries | `internal/state/ruling.go` `RegisterDesk` |
| The settings file | `internal/desk` writes `<state>/launch/desk-<id>.json` with `permissions.allow` only: `Bash(<prefix> ruling record *)`, `Bash(<prefix> brief*)`, `Bash(<prefix> prompt *)`, and `Bash(<prefix> pass start *)`, where `<prefix>` is `os.Executable()` plus `--state-dir <dir>` when one was given; `Read(//<state>/**)` and `Read(//<skill file>)`; and `Edit(//<state>/desk/words/**)`. A `Write(path)` rule matches nothing, so the words file takes an `Edit` rule (Claude Code permissions doc). No hook, no deny change | `internal/launch/settings.go` `WriteSettings` is the pattern; pass 5 owns that file |
| Permission mode | `--permission-mode default` on the `argv`, since `~/.claude/settings.json:177` sets `defaultMode` to `auto`. With the allow rules above, the recorded path asks nothing; anything else asks Geoff at the keyboard | Done item 3: "No permission prompt and no trust question appears" |
| The process the list shows | `syscall.Exec` of `claude`, found on `PATH` by `exec.LookPath`, so the registered PID is the listed PID | Probe of 2026-11-18 in the design |
| Arguments | `claude --session-id <id> --model claude-opus-5-5 --effort high --permission-mode default --settings <file> --append-system-prompt <text> --name <project>-desk "What is waiting for me?"`. The system prompt names `<prefix>` as the exact command the session types for each verb, so the typed text matches the allow rules | Done item 3 names Opus and high effort |
| The environment | Drop `CLAUDECODE`, `CLAUDE_CODE_CHILD_SESSION`, `CLAUDE_CODE_SESSION_ID`, and `CLAUDE_CODE_ENTRYPOINT` before the exec, or the session isn't listed | Probe of 2026-11-18 |
| The folder | The main checkout of the project whose question comes first in the brief (`rf.Checkout`), or of the named project. With nothing waiting, or no checkout recorded, the current folder | Design "The desk opens" |
| His words, unchanged | `ruling record` gains `--words-file <path>`, read whole with trailing newlines stripped (as `$(cat)` does), exclusive with the trailing argument. The skill tells the session to write the words to `<state>/desk/words/<n>.txt` with the Write tool (allowed by the `Edit` rule), then run `<prefix> ruling record ... --words-file <that path>`. This avoids a `$(...)` in a Bash rule, which no run has shown matches an allow rule | `skills/cc-desk/SKILL.md` names the job's `tmp` folder, which an interactive session doesn't have, and bare `cc-desk` |
| The guard | Called unchanged. `State.Guard` tests the registered id, the desk PID in the caller's chain, and the desk process's terminal | `internal/state` `Guard`; PR 17 |
| No-terminal refusal, live | Since `cc-desk` won't open a desk under `setsid`, the live script starts real `claude` under `setsid` in a `tmux` pane by hand, with a copy of the state folder whose `desks.json` the script writes with that session's id and PID (the seam pass 1's e2e uses). That session runs the record against the copy, and the unchanged guard answers `RefuseNoTerminal`. The probe below shows Claude Code opens and answers that way | Probe below; `e2e/ruling_test.go` `inSetsidDesk` |
| `desk register`'s list read | `exec.CommandContext` with a 30-second context, `cmd.WaitDelay` set so a grandchild holding the pipe can't hold `Output()` (golang/go#23019), and a hidden `--list-wait` flag so a test can shorten it | `cmd/cc-desk/desk.go:70`; `docs/follow-ups/live-check-cleanup-gaps.md` item 6 |
| The free-pass start | In `sessions.Plan`, after `s.Settle(now)`, so a waiting pass takes a free slot first (`settle` starts waiting passes oldest first, `internal/state/commit.go:233`): for each live release with a checkout, each `NotStarted` pass in file order with `Free` true, while `Limits()` is empty, run `state.StartPass` and append a `Result` with the new decision `start` and reason `free`. `Limits()` is read again after each start, and it holds whenever the usage reading is missing (`rules.go:213`). A held pass stays not started, and the next run starts it. The existing launch step then adds the pass's `launch` decision in the same run | `internal/sessions/decide.go:164`; `internal/state/rules.go` `Free`, `Limits` |
| The start in the check log | `RecordCheck` logs only stops (`checkrun.go:38`), and `RecordOutcomes` logs only a non-empty outcome (`launch.go:197`). `carry` returns `""` for a decision outside its list (`carryout.go:51`), so `carry` gains a `start` case that returns the outcome `started`. `applyRecorded` lets a decision with no session record through only for `launch` and `wait_usage`, so `start` joins that list | `internal/state/checkrun.go:59`; `cmd/cc-desk/carryout.go:46` |
| No guess, no checkout | `StartPass` passes an empty guess, so `Breaks` never weighs one. `hasCheckout` gates the release | `internal/state/events.go:184`; `decide.go:197` |
| Replacement rows | `checkReplace` keeps its first four refusals. Not started, waiting, and stopped passes take the replacement. A building pass takes it and gets a mark. A done pass is refused: "has finished; a new change is needed", true of merged and closed passes alike | `internal/state/ruling.go:140` to `:181` |
| The mark | A new optional field on the pass, `replaced` (the new ruling id and time), set by `ReplaceRuling` when the pass is building, or stopped with a session record. An older file reads, since the field is `omitempty`. A question added by `AskMerge` takes no mark: `ReplaceRuling` already moves `q.Answer` to the new ruling, so the merge rows read the latest answer, and a replacement after the merge carry acted (`MergeCalled` or `Closed` set) is refused | Design "Left to the builder" allows new fields; `internal/state/ruling.go:134` |
| The relaunch | The check reads the mark first, before the merge rows, and decides `end_and_relaunch` with reason `replaced`, uncounted. `Relaunch` with a reason other than usage skips the stop (`carryout.go:92`), so it must not be that decision. The carry-out stops the live session through the binary, confirms it gone from the list, then launches with `Counted: false` and a `Replace` prompt paragraph holding the question's text and the new words. While the mark stands, `decideMerge` holds, and the launch's `Also` event clears both the mark and `pr.Ready`, so a pre-replacement commit never merges. `launchRequest` adds the paragraph whenever the pass carries a mark, so a stopped pass's next launch (a usage relaunch) carries it too | `cmd/cc-desk/carryout.go:85`; `cmd/cc-desk/launcher.go:152`, `:167`, `:270`; `internal/state/launch.go` `RecordLaunch` |
| The paragraph | `internal/launch/templates/replace.txt`, filled from a `ReplaceNote{Question, Words}`. It composes with a relaunch, answer, or fix note, so `Fill`'s one-note rule doesn't count it | `internal/launch/prompt.go:84` |
| Merge answers | `decideMerge` gains the answered question's option and words, read by its caller in `decide.go` (`Answer.Ruling` names a ruling; `Ruling.Option` gives the option), and `MergeInfo` gains `Words`. The answered-question rows come before the `CLOSED` and session-ended returns (`internal/sessions/merge.go:87`, `:92`), so a close that failed halfway runs again, and before the outside-design and fix-limit rows, so a "merge" doesn't ask again. An unanswered question holds, as today | `internal/sessions/merge.go:56`, `:97`; `internal/state/types.go:321` |
| The commit asked about | `AskMerge` gains the commit it asks about and stores it on the pull request as `asked_commit`. A first question asks about `pr.Commit`; a question asked again because the head moved asks about the new head. `ReportReady` replaces the whole pull request record, so a new report clears both `Question` and `asked_commit` | `internal/state/merge.go:30` to `:60`, `:156` |
| Merge | Merge when every check passed and the pull request's head equals `asked_commit`, through the existing `MarkMerge`, `callMerge` (`--match-head-commit`), `MergePass`, and cleanup. Otherwise ask again: `AskMerge` adds `merge-<pass>-<k+1>` with the new head | `cmd/cc-desk/merge.go` `carryMerge` |
| Fix again | `sendBack` with `FixNote.Words` set from the ruling. `RecordFix` raises `FixAttempts` past `FixLimit`, so a second return at the limit reaches the existing `ReasonFixLimit` ask, whose text counts `FixAttempts` rather than reading "failed FixLimit+1 times" (`cmd/cc-desk/merge.go:140`). `fix.txt` prints the failures only when there are some, since a "fix again" on an outside-design question has none. The outside-design rule asks again on a return | `internal/sessions/merge.go:136`; `internal/launch/templates/fix.txt` |
| Close | A new carry, safe to run twice: `gh pr close <n> -R <repo>` unless the pull request already reads `CLOSED`, then stop the session through the binary and confirm it gone, then commit `Sequence(EndSession, ClosePass(project, pass, rulingID))` | `internal/state/events.go:294`, no caller today; `internal/state/launch.go:124` |
| After a close | `BreakClosed` already blocks a pass after a closed one, so the free-pass step never starts it, and the snapshot's `why_not` names it. A closed pass's worktree and branch stay, since cleanup needs `Merged` (`merge.go:198`); Task 8 files that as a follow-up | `internal/state/rules.go` |
| Commands under the check's unit | `cc-desk-check.service` sets `PATH=%h/.local/share/cc-desk/bin:%h/.local/bin:/home/linuxbrew/.linuxbrew/bin:/usr/local/bin:/usr/bin:/bin`. Under it: `claude`, `gh`, and `git` resolve to `/home/linuxbrew/.linuxbrew/bin/`, `systemd-run` to `/usr/bin/systemd-run` (resolved 2026-11-18). `gh pr close` adds no new command | `systemd/cc-desk-check.service:7` |
| The live run's binary | The live script runs `timer install` from its branch build, as pass 5's `live-check.sh` does, so the timer runs this pass's check. It prints the line that puts main's build back, and the conductor reinstalls from `main` once the pass lands | `scripts/live-check.sh`; round plan |
| The install line | The README's install lines add `ln -sf ~/.local/share/cc-desk/bin/cc-desk ~/.local/bin/cc-desk`, so Geoff's shell finds `cc-desk`. The terminal link's handler already falls back to the installed path | Design "Left to the builder", last item; `docs/follow-ups/cc-desk-on-browser-path.md` |

The no-terminal probe, 2026-11-18 at 18:23 local, in `~/Projects/cc-desk-scratch` on `main`,
Claude Code 2.1.327. `script -qfec "setsid -w inner.sh"`, where `inner.sh` printed its own
`/proc` fields and then ran `exec claude --session-id <uuid> --name cc-desk-scratch-probe-notty
--model sonnet --effort low "Reply with the one word ready."` with the four variables cleared:

```text
probe PID 1780876 tty_nr 0 isatty0 yes
```

The screen showed Claude Code's banner ("Sonnet 5.5 with low effort"), and the transcript
`6b3c3f58-dbcc-42bb-a04b-a39be39beccc.jsonl` holds the user line and the reply `ready`. The
45-second `timeout` ended it, and no process with the name remained. A session list read 6
seconds after the start didn't show the row yet. The live script waits for the row with the
`--launch-row-wait` pattern before it reads the list.

## Recorded searches

Run from the worktree with `grep -rn` over `*.go`, `*.md`, `*.sh`, and `*.txt`, excluding
`docs/plans/` and `docs/specs/`. Each count is the hit count in that file.

| Search | Hits |
| --- | --- |
| `checkReplace` | `internal/state/ruling.go` 3, `docs/follow-ups/desk-skills-review-notes.md` 1 |
| `ReplaceRuling\|Replaces` | `internal/state/ruling.go` 12, `replace_test.go` 8, `launch_test.go` 4, `launch.go` 4, `types.go` 2, `read.go` 2, `cmd/cc-desk/launcher.go` 2, `carryout.go` 2, `internal/state/merge_test.go` 1, `e2e/replace_test.go` 1 |
| `comes in pass 3` | `internal/state/replace_test.go` 3, `internal/state/ruling.go` 1, `e2e/replace_test.go` 1 |
| `decideMerge` | `internal/sessions/merge.go` 2, `decide.go` 1 |
| `carryMerge` | `cmd/cc-desk/merge.go` 2, `carryout.go` 1 |
| `sendBack` | `cmd/cc-desk/merge.go` 3 |
| `ClosePass` | `internal/state/closed_test.go` 6, `events.go` 2, `docs/follow-ups/desk-skills-review-notes.md` 1 |
| `FixNote` | `internal/launch/prompt.go` 4, `launch_test.go` 2, `cmd/cc-desk/merge.go` 1, `launcher.go` 1 |
| `agentRows` | `cmd/cc-desk/desk.go` 3 |
| `RegisterDesk` | `internal/state/ruling_test.go` 5, `ruling.go` 2, `cmd/cc-desk/desk.go` 1 |
| `Settle(` | `internal/state/start_test.go` 1, `commit.go` 1, `internal/sessions/decide.go` 1 |
| `logCheck` | `internal/state/checkrun.go` 3, `launch.go` 1 |
| `job.s .tmp.\|job .tmp` | `skills/cc-desk/SKILL.md` 1, `skills/cc-desk-work/SKILL.md` 1, `internal/launch/templates/pass.txt` 1, `design.txt` 1, `conductor.txt` 1, `internal/launch/launch_test.go` 1 |
| `within a minute` | `skills/cc-desk/SKILL.md` 1, `docs/follow-ups/desk-skills-review-notes.md` 1 |
| `cobra.Command{` | 13 files under `cmd/cc-desk/`; the root is `root.go` (2) |
| `checkout` (case blind) in tests | `e2e/launch_test.go` 8, `internal/sessions/plan_test.go` 7, `e2e/events_test.go` 6, `e2e/launch_rig_test.go` 3, `e2e/broken_test.go` 2, `e2e/records_test.go` 2, `e2e/guard_test.go` 2, `e2e/ask_test.go` 1 |
| Follow-up names | `cc-desk-on-browser-path`, `merge-question-answer`, `live-check-cleanup-gaps`, `desk-skills-review-notes`: cited in specs and plans only, plus `docs/follow-ups/page-link-chromium-click.md` 1 |

The fixture survey. Every `e2e/testdata` folder holding `releases/` was copied and given
`snapshot` at `2026-11-21T00:00:00Z` by the build of `6b42644`. Free passes: `order` 4
(`north/n-free`, `north/n-two`, `south/s-free`, `south/s-two`), `skills-none` 1 (`harbor/h-three`),
`state` 2 (`alpha/p13`, `beta/b-three`). None of those releases records a checkout. The one
fixture with a `checkout` key, `real-2026-11-18`, has no free pass. So no fixture's behavior
changes by itself. The tests that add a checkout at run time are the eight files in the last
search row. One that also holds a free pass and a fresh usage reading now sees a `start` and a
`launch` it didn't before; with no usage reading, `Limits()` holds and nothing starts. Task 2 runs them, and each changed expectation is listed in the PR. No fixture file is
edited (standing answer 16).

## File list

| File | Change | Task |
| --- | --- | --- |
| `cmd/cc-desk/desk.go` | The 30-second list read and `--list-wait` | 1 |
| `cmd/cc-desk/desk_test.go` or `e2e/register_test.go` | A stand-in `claude` that sleeps; the verb gives up at the limit | 1 |
| `internal/sessions/decide.go` | The free-pass step in `Plan`, the `start` decision | 2 |
| `internal/sessions/plan_test.go` | Rows for free, held by the limit, no checkout, on a guess, after a closed pass | 2 |
| `internal/state/checkrun.go` | `start` joins `applyRecorded`'s list | 2 |
| `cmd/cc-desk/carryout.go` | `carry`'s `start` case and its `started` outcome | 2 |
| `e2e/start_test.go` (new) | A ruling frees a pass, the check starts it and logs the start | 2 |
| `e2e/*_test.go` in the checkout row above | Expectations that change with the new start | 2 |
| `internal/state/ruling.go`, `types.go`, `read.go` | The replacement rows, the `replaced` mark, the refusal after a merge carry acted | 3 |
| `internal/state/replace_test.go`, `e2e/replace_test.go` | The new rows; the "comes in pass 3" rows change | 3 |
| `internal/sessions/decide.go`, `merge.go` | The `replaced` decision, read before the merge rows; merges hold while the mark stands | 3 |
| `cmd/cc-desk/carryout.go`, `cmd/cc-desk/launcher.go` | The `end_and_relaunch` route for `replaced`; `launchRequest`, `causeText`, and the note switch | 3 |
| `internal/state/launch.go` | The launch's `Also` event clears the mark and `pr.Ready` | 3 |
| `internal/launch/prompt.go`, `templates/replace.txt` (new), `launch_test.go` | `ReplaceNote` and its paragraph | 3 |
| `internal/sessions/merge.go`, `merge_test.go` | The answered-question rows, before the `CLOSED`, ended, outside-design, and fix-limit rows; `MergeInfo.Words` | 4 |
| `internal/sessions/decide.go` | `decideMerge`'s caller passes the answer's option and words | 4 |
| `internal/state/merge.go`, `types.go`, `merge_test.go` | `AskMerge` takes the commit; `asked_commit` on the pull request | 4 |
| `cmd/cc-desk/merge.go` | The close carry, safe to run twice; words into `FixNote`; the fix-limit text | 4 |
| `internal/launch/prompt.go`, `templates/fix.txt` | `FixNote.Words`; failures printed only when present | 4 |
| `e2e/merge_rig_test.go`, `e2e/testdata/check-merge-answers/gh.py` (new, copied from `check-merge`) | `gh pr close` in the new stand-in, the three answers | 4 |
| `cmd/cc-desk/root.go` | `RunE`, `Args`, the terminal test, the hidden `--skill-file` | 5 |
| `cmd/cc-desk/open.go` (new) | Build, register, write settings, scrub, exec | 5 |
| `internal/desk/settings.go` (new), `settings_test.go` | The desk's allow-only settings file | 5 |
| `cmd/cc-desk/ruling.go` | `--words-file` | 5 |
| `e2e/open_test.go` (new) | A stand-in `claude` that records `argv`, environment, and folder | 5 |
| `skills/cc-desk/SKILL.md` | The corrected lines (Task 5's list) | 5 |
| `scripts/live-desk.sh` (new), `scripts/live/desk/` (new fixtures) | The live checks for items 3 to 5 | 6 |
| `README.md` | The install lines | 8 |
| `docs/guide.md` | Line 8's caveat removed | 8 |
| `docs/follow-ups/` | Task 8's edits, and one new file for a closed pass's worktree and branch | 8 |
| `docs/STATUS.md` | The pass's line | 8 |

No `go.mod` change unless Task 5 takes `golang.org/x/sys`. `go.mod` records Go 1.27 and cobra
v1.10.2 with `pflag` v1.0.9, and no `x/sys` or `x/term`. If the terminal test needs `x/sys/unix`,
it is a new direct dependency and goes through `standards:dependency-upgrade` with its version
named as `go.mod` records it. The stdlib route is `syscall.Syscall(SYS_IOCTL, fd, TCGETS, ...)`
through `syscall.Termios`; Task 5 prefers it, since it adds no module.

## Fixtures

New fixtures only. `e2e/testdata/desk-open/` holds a release for `north` with a checkout field
pointing at a test folder, a round with a written design, one question that frees `n-wait`, and
a pass after a closed one. `e2e/testdata/check-merge-answers/` copies `check-merge`'s shape with
three answered merge questions, one per option, and its own copy of `gh.py` that also answers
`gh pr close`; `check-merge` stays as it is. `scripts/live/desk/` holds the release the live
script adds for the project `cc-desk-scratch`, under a name that starts `cc-desk-scratch-` and
carries the run's time.

## Acceptance

The design's items, word for word:

> 3. **The desk opens.** On a terminal a test makes, in the scratch repo's main checkout as it
>    stands, `cc-desk` starts a session that the session list shows as `interactive` under the
>    id and process the command registered. The process's arguments name Opus and high effort.
>    The screen holds the one line the command passed and then the opening. It holds neither
>    the brief as the command printed it nor the session's instructions. No permission prompt
>    and no trust question appears. `cc-desk <project> <round>` opens on that round's questions
>    only. With no terminal, `cc-desk` prints help and registers nothing. With a state file that
>    fails to parse, it prints the error, starts no session, and registers nothing.
> 4. **Ruling.** In that session a typed answer to a fixture question adds one line to the
>    rulings log with the time and the typed words, byte for byte. The test types an apostrophe,
>    a double quote, a dollar sign, and a second line. The snapshot drops the question within 5
>    seconds. The pass that waited starts within 5 minutes with nothing else typed: the check's
>    log holds the start, and the desk session ran no start command. The same record from a
>    background session is refused. So is one from a shell that the desk session started with
>    no terminal, and that refusal names the missing terminal.
> 5. **Replacing and merging.** A replacement before the pass starts leaves both lines in the log
>    with the first marked replaced, and the pass starts on the second. A replacement while the
>    pass builds ends its session and starts one with the new words in its prompt. The relaunch
>    count is unchanged, and the session list never shows two sessions for the pass. A
>    replacement after the merge is refused. For a fixture merge question, "merge" merges at
>    the commit asked about, and asks again when the commit has moved. "Fix again" relaunches
>    the pass once with the words, and asks again when that attempt comes back the same way.
>    "Close" closes the pull request, ends the session, records the pass as closed with the
>    ruling, and leaves "Needs you now" empty. A pass that comes after the closed one doesn't
>    start.
> 11. **The gate.** `bash scripts/gate.sh` passes with no skipped test. Items 2 to 5 and 7 to 10
>     have parts that run live and aren't Go tests. Each live script prints one line for each
>     check and ends with its result. A pass's pull request holds that output from a run on the
>     pass's last commit, and a pass that leaves it out isn't done.

Item 4's no-terminal clause reads as the design's "Rulings" paragraph states it: "A ruling is
refused ... from a shell that holds the desk session's id with no terminal." With the guard
unchanged, the shell has no terminal when the desk process has none, which is what pass 1's
e2e `no_terminal` case builds. Every Bash call a desk session makes runs with no controlling
terminal, so the clause can't mean the desk's own shells, or no desk could record. `cc-desk`
itself refuses to open with no controlling terminal ("no desk opens that can't record"). So the
live check starts real Claude Code under `setsid` by hand, registers it in a copy of the state
folder, and has it record. This is a call, listed under `## Decisions`.

### Clause table

Live means `scripts/live-desk.sh` against real Claude Code on a `tmux` pane in
`~/Projects/cc-desk-scratch`, with sessions named `cc-desk-scratch-*` and the real state folder
holding only that project's releases. Rig means a Go test with stand-ins, run by the gate.

| Item | Clause | Where | Task |
| --- | --- | --- | --- |
| 3 | Listed `interactive` under the registered id and PID | Live: `desks.json` against the saved session list | 6 |
| 3 | Arguments name Opus and high effort | Live: `/proc/<pid>/cmdline`; rig: the recorded `argv` | 5, 6 |
| 3 | The screen holds the one line, then the opening, and neither the brief nor the instructions | Live: `tmux capture-pane`, searched for the first line, a brief line, and a system-prompt sentence | 6 |
| 3 | No permission prompt, no trust question | Live: the capture holds neither prompt's text | 6 |
| 3 | `<project> <round>` opens on that round's questions only | Rig: the recorded `--append-system-prompt` holds that round's questions and no other; live: a second open's screen | 5, 6 |
| 3 | No terminal: help, nothing registered | Rig: `</dev/null` runs and `setsid` runs on a `pty`, `desks.json` unchanged | 5 |
| 3 | Parse failure: the error, no session, nothing registered | Rig: a broken release file, the stand-in never runs | 5 |
| 4 | A typed answer with `'`, `"`, `$`, and a second line adds one line, byte for byte | Live | 5, 6 |
| 4 | The snapshot drops the question within 5 seconds | Live | 6 |
| 4 | The pass starts within 5 minutes; the log holds the start; the desk ran no start command | Live: `check-log.jsonl`, the desk transcript's Bash calls; rig: `plan_test.go`, `e2e/start_test.go` | 2, 6 |
| 4 | Refused from a background session | Live: the executor's own Bash runs the verb on a copy of the state folder | 6 |
| 4 | Refused from a shell with no terminal, naming it | Live: real Claude Code started by hand under `setsid`, registered in a copy of the folder, records; rig: the existing `no_terminal` case | 6 |
| 5 | Replacement before the start: both lines, the first marked, the pass starts on the second | Rig: `e2e/replace_test.go` and `e2e/start_test.go` | 3 |
| 5 | Replacement while building: session ended, new one with the words, count unchanged, never two | Live: the pass the item 4 run started; the list saved every 5 seconds | 3, 6 |
| 5 | Replacement after the merge is refused | Rig | 3 |
| 5 | "Merge" merges at the commit asked about; asks again when it moved | Live on a scratch pull request; rig in `merge_rig_test.go` | 4, 6 |
| 5 | "Fix again" relaunches once with the words; asks again when it comes back the same | Rig: `launch_rig` stand-in records the prompt | 4 |
| 5 | "Close" closes, ends, records closed, empties "Needs you now"; the next pass doesn't start | Live on a scratch pull request; rig for the session end | 4, 6 |
| 11 | The gate passes with no skipped test; the script prints a line per check and a result | Gate; the PR holds the script's output from the last commit | 7, 9 |

### Pre-code runs on `6b42644` (`81afc1d` changes only the round plan)

Built to `$CLAUDE_JOB_DIR/tmp/precheck/cc-desk`, with a copy of `e2e/testdata/state`:

```text
$ cc-desk --state-dir X </dev/null
Keep the state of Claude Code's work and print its snapshot
Usage: ...                                     (exit 0)
$ script -qfc 'cc-desk --state-dir X'          (on a terminal)
Keep the state of Claude Code's work and print its snapshot
Usage: ...                                     (item 3 fails today)
$ cc-desk --state-dir X cc-desk-scratch desk
Error: unknown command "cc-desk-scratch" for "cc-desk"
$ go test ./internal/state -run TestReplace -v
--- PASS: TestReplaceRuling, with a_building_pass, a_stopped_pass, a_waiting_pass
    expecting "has started; replacing it while it builds comes in pass 3"
```

`internal/sessions/merge.go:97` reads `case pr.Question != "": ... detail = "question " +
pr.Question + " asked"`, so nothing reads an answer today. `ClosePass(` has one non-test site,
its definition at `internal/state/events.go:294`.

## Tasks

The executor commits after each task by path. Tasks 2 to 4 are lane A and Tasks 1, 5, and 6 are
lane B. Tasks 2, 3, and 4 share `decide.go`, `carryout.go`, `merge.go`, and `prompt.go`, so they
run in that order in one lane. Lane B shares no file with lane A. The two lanes run as pairs, as `executor.md` "Pairs" says, with
descriptions `desk-open task <n>`.

### Task 1: the 30-second list read (10 minutes)

Outcome: `desk register` gives up on `claude agents --json` after 30 seconds with an error that
says so, and registers nothing. Files: `cmd/cc-desk/desk.go`, its test. Class `runner`. Check: a
stand-in `claude` that starts a child `sleep 5` holding its stdout, with `--list-wait 1s`, exits
non-zero within 3 seconds (`WaitDelay` set) and leaves `desks.json` unchanged. Test: `go test ./e2e -run TestDeskRegister`.

### Task 2: the check starts free passes (40 minutes)

Outcome: the mechanism rows "The free-pass start" to "No guess, no checkout". Files: the Task 2
rows of the file list. Class `runner`. Check: in `plan_test.go`, with a fresh usage reading, a
free pass in a release with a checkout gets `start` then `launch`. A waiting pass that `Settle`
starts takes its slot before a free pass. With one pass already building, two free passes yield one
start; with two building, none. A pass with no checkout, one free only on a guess, and one after
a closed pass are never started. `e2e/start_test.go` records a ruling, runs `check` with the
stand-ins, and finds a line for the pass in `check-log.jsonl` with decision `start` and outcome `started`. Then the eight files in
the checkout row run, and each changed expectation goes in the PR's `## Decisions` with its
reason. Test: `go test ./internal/sessions ./internal/state ./e2e`.

### Task 3: a wrong record is replaced (60 minutes)

Outcome: the mechanism rows "Replacement rows" to "The paragraph". Files: the Task 3 rows.
Class `runner`. Check: `replace_test.go`'s rows become not started (replaced, starts on the new
ruling), waiting (replaced, no mark), stopped (replaced, marked), building (replaced, marked),
done by merge and done by close (refused, "has finished; a new change is needed"). A merge
answer is replaced with no mark, and refused once `MergeCalled` or `Closed` is set. A check run
on a building pass with a mark and a live session decides one `end_and_relaunch` with reason
`replaced`, ahead of any merge row, and with a ready report standing it merges nothing. The
carry-out stops the session, waits for it to leave the list, and launches with the paragraph.
The counted relaunch total doesn't change, `Relaunches` gains one entry with `counted: false`,
and the mark and `pr.Ready` are gone afterwards. A stopped pass with a mark gets the paragraph
on its usage relaunch. Test: `go test
./internal/state ./internal/sessions ./internal/launch ./e2e -run 'Replace|Launch'`.

### Task 4: the three merge answers (70 minutes)

Outcome: the mechanism rows "Merge answers" to "After a close". Files: the Task 4 rows. Class
`runner`. Check: in `merge_test.go` and `merge_rig_test.go`, with the stand-in `gh`:

- "merge" with checks passed and head at `asked_commit` merges once with `--match-head-commit`.
- "merge" with the head moved adds `merge-<pass>-2` asking about the new head, and calls no
  merge. "merge" on that question, with the head unchanged, merges at it.
- An answered question on an outside-design pull request merges without asking again.
- "fix again" launches once with the words in the prompt, raises `FixAttempts` to 3, and a
  second failure asks again with reason `fix_limit`.
- "close" calls `gh pr close`, stops the session, writes `closed` with the ruling id, and the
  snapshot shows no waiting question for the pass. A run where the stop fails after the close
  leaves the pass building, and the next run skips `gh pr close` and finishes. The pass after it shows `why_not` naming the
  closed pass and gets no `start`.

Test: `go test ./internal/sessions ./internal/state ./e2e -run 'Merge|Close|Fix'`.

### Task 5: the desk opens (65 minutes)

Outcome: the mechanism rows "Bare `cc-desk`" to "His words, unchanged". Files: the Task 5 rows.
Class `runner`. Check: `e2e/open_test.go` runs `cc-desk` on a `pty` the test makes (`script -qec`,
as `startDesk` does) with `--claude-command` set to a stand-in that writes its `argv`,
environment, folder, and PID to a file. The test checks:

- The PID equals the one in `desks.json`.
- The `argv` has the model, the effort, `--permission-mode default`, the settings path, the first
  line, and a system prompt holding the brief and the exact command prefix.
- The four variables are gone, and the folder is the fixture's checkout.
- The settings file has the allow rules, each on that prefix, the `Edit` rule for the words
  folder, and no hook.
- `cc-desk north n1` gives a system prompt with n1's questions only, and a mistyped project is
  refused with the names it knows.
- `</dev/null` and a `setsid` run on a `pty` (stdin and stdout terminals, `tty_nr` 0) print help
  and leave `desks.json` absent.
- A broken release file prints the error, exits 2, and the stand-in never runs.
- `ruling record --words-file` keeps `'`, `"`, `$`, and an inner newline byte for byte.

The skill is corrected and each line listed in the PR (standing answer 15):

- The `$(cat <file>)` line becomes the words-file form with the state folder's `desk/words/`.
- Each command the skill types uses the prefix the system prompt gives, never bare `cc-desk`.
- The replace table matches the design's three rows.
- "The build starts within a minute" becomes "The check starts the build within five minutes".
- The merge answers section names what each answer does.
- The prompts section's verb forms are checked against `cc-desk --help`.

The skill stays under 160 lines and passes `vale` and `tellgrader`, as item 2 asks. Test: `go
test ./cmd/... ./internal/desk ./e2e -run 'Open|Ruling|Settings'`.

### Task 6: the live script (50 minutes)

Outcome: `scripts/live-desk.sh`, in `live-check.sh`'s phased shape, with its refusals: a dirty
scratch, a scratch branch other than `main`, or a release in the real state folder other than
its own. It prints `pass <check>` or `fail <check>` per clause in the table's live rows, then
`result: pass` or `result: fail`. It logs to `$CLAUDE_JOB_DIR/tmp/live-desk/run.log`. Phases, in
order:

1. `prep`: build to `$RUN/cc-desk`, `timer install` from it, `systemd-analyze --user verify`,
   add the run's release, and check `command -v` for `claude`, `gh`, `git`, and `systemd-run`
   under `unit_path()`.
2. `open`: a `tmux` session `cc-desk-scratch-desk-<time>`, 200 by 50, runs `cc-desk --skill-file
   <repo>/skills/cc-desk/SKILL.md` in the scratch checkout. The phase waits for the row, saves
   the list, and captures the pane.
3. `round`: the round form in a second pane, captured and then ended.
4. `rule`: `tmux send-keys` types the answer with the four characters (a second line through
   the key Claude Code takes for a newline, found and recorded in the run). The phase checks the
   log line, the snapshot within 5 seconds, and the desk transcript for a start command.
5. `start`: waits up to 5 minutes for the `start` line and the pass session's row, saving the
   list every 5 seconds from here to `replace`'s end.
6. `refuse`: the executor's own session id on a copy of the folder. Then, in a third pane, real
   `claude` started under `setsid` with a known session id, written with its PID into the
   copy's `desks.json`, records against the copy and is refused for the missing terminal.
7. `replace`: a correction typed in the desk. The phase waits for the check's relaunch, then
   reads the new session's first prompt for the words and the relaunch count.
8. `merge`: a scratch branch and pull request, made and pushed by the script. The run's release
   gets a merge question on it. The phase answers "merge" after a second push (asks again about
   the new head), then "merge" on the new question (merges at that head).
9. `close`: a second scratch pull request answered "close". The phase checks it closed, the pass
   closed, and the snapshot's "Needs you now" rows empty for it.
10. `abort` and `evidence`: as `live-check.sh` has them, ending every session the run started
    through the binary.

Files: `scripts/live-desk.sh`, `scripts/live/desk/`. Class `runner`. Check: `bash -n` and
`shellcheck` pass, and `bash scripts/live-desk.sh prep` refuses on a dirty scratch. Test: the run
in Task 9.

### Task 7: fold the lanes, review, gate (20 minutes)

Outcome: both pair branches merged into `desk-open`, one whole-branch `standards:diff-reviewer`
read (Opus, medium) folded, and `bash scripts/gate.sh` passing with no skipped test. Class
`runner`. Check: the gate's last line and `go test ./... -count=1 -json` with no `"skip"`
action.

### Task 8: words (15 minutes)

Runs before the live run, so the run is on the last commit. Outcome:

- The README's install lines with the link.
- `docs/guide.md` line 8's caveat removed.
- `docs/follow-ups/merge-question-answer.md` and `cc-desk-on-browser-path.md` removed, since
  this pass folds them.
- `live-check-cleanup-gaps.md` reworded without item 6.
- `desk-skills-review-notes.md` reworded to item 5 alone (where a permission prompt's details
  come from, with no source yet).
- One new follow-up: a closed pass's worktree and branch stay, since cleanup needs a merge.
- The STATUS line.

Class `runner`. Check: `vale` and `tellgrader --register docs` on the README, guide, and
follow-ups; the gate.

### Task 9: the live run on the last commit (50 minutes)

Outcome: `bash scripts/live-desk.sh` from start to end on the commit the PR will hold, every line
`pass`, and `result: pass`, then the PR body and the CI watch. Class `runner`. Check: the
output, pasted whole into the PR. Each permission prompt it meets is handled as standing answer
18 says, and each refusal as standing answer 19 says. A fix found here makes a new commit, and
the gate and the run repeat on it.

### The path

| Step | Lane A | Lane B |
| --- | --- | --- |
| Build | Tasks 2, 3, 4: 170 | Tasks 1, 5, 6: 125 |
| Fold | Task 7: 20 | |
| Words | Task 8: 15 | |
| Run | Task 9: 50 | |

The path is 170 + 20 + 15 + 50 = 255 minutes. The sum over every task is 380. The figures are in
the unit the round's earlier plans used, and they aren't shaved to fit.

## The gate and CI

Run `bash scripts/gate.sh` bare, never piped, with the Bash tool's `timeout: 600000`. A plan or
doc is staged before the gate, since the gate lints tracked files only. CI: `gh run list --branch
desk-open --limit 3`, then `gh run watch <id> --exit-status` with `run_in_background`. No review
runs on a commit while CI runs.

## Standing answers

The round plan's answers 13 to 27 and the twelve in `round-run.md` hold. These are this pass's.

- **35. The terminal test needs a module.** It takes the stdlib `TCGETS` route. If that fails on
  this kernel, it takes `golang.org/x/sys` through `standards:dependency-upgrade` and lists it.
- **36. Claude Code started by hand under `setsid` doesn't open in the live run.** The no-terminal
  clause rests on the e2e `no_terminal` case. The PR lists it under `## Planning misses` with the run's output.
- **37. A fixture-driven test changes with the free-pass start.** The test's expectation changes,
  never the fixture, and the PR lists it.
- **38. The words file meets a permission prompt.** Standing answer 18 holds: end the desk, file
  it, fix the allow rule, run once more.
- **39. The session list shows two sessions for the pass during `replace`.** That is a failed
  check, never retried away. The executor traces it in the carry-out's stop-and-confirm step.
- **40. A live pass session started by `start` begins real work in the scratch repo.** Expected.
  It runs until `replace` ends it, and `abort` ends its successor.
- **41. `gh pr close` or `merge` on a scratch pull request fails.** The script reports `fail` with
  the binary's stderr. Nothing acts on `glw907/cc-desk` or any other repo.
- **42. The timer runs this pass's build after the run.** Expected. The script prints the line
  that puts main's build back, and the conductor reinstalls from `main` once the pass lands.
- **43. The guard would need to change.** That stops the pass: the guard is `auth-data`, and its
  reads are outside this plan.

## Decisions

- [advisor] The pass parks at planning. Line 27 of `planner.md` sets the design's estimate from
  past scores, and the plan's path is compared with twice that estimate, never discounted by
  those scores. PR 18 parked at 170 against 75 with no discount.
- [advisor] Each clause is priced live or rig from the design's words before the path is summed.
  Integration, review, and CI are on the path, as the desk-page plan counted them.
- [call] Item 4's no-terminal clause is the desk with no terminal, with the guard unchanged.
  `cc-desk` refuses to open without a controlling terminal, so the live check starts Claude
  Code under `setsid` by hand and registers it in a copy of the state folder. The probe of
  2026-11-18 shows Claude Code opens that way.
- [call] The brief goes in the system prompt, the skill by `--skill-file` in tests, and words by
  `--words-file`. A second `cc-desk` opens a second working desk.
- [call] The desk runs with `--permission-mode default`, since the user settings default to
  auto, and allow rules only, on the exact command prefix the system prompt gives.
- [call] A re-asked merge question records the commit it asks about (`asked_commit`).
- [call] A replaced merge answer takes no mark: the merge rows read the latest answer, so the
  replacement acts on the next check. Once `MergeCalled` or `Closed` is set, it's refused.
- [call] A closed pass's worktree and branch stay, filed as a follow-up; the design asks only
  that the close ends the session and records the pass.
- [review] One `standards:plan-reviewer` read (Opus, medium) found six blocking mechanics errors
  and nine others. All are folded into the mechanism rows, files, and checks above, and the path
  rose from 210 to 255.
- [call] A replacement while building is a mark the check carries out, never an end in the verb.
- [call] A pass held by the limit of two stays not started rather than becoming `Waiting`.
- [call] Review note 5 stays as a follow-up. Notes 1 to 4 and 6 are folded by Tasks 5, 4, 2, 3,
  and 5.
