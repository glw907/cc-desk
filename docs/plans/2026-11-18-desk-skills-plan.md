# Desk skills: pass plan

Initiative: cc-desk
Round: desk
Pass: desk-skills

Class `runner`. Planned 75 minutes. The path below is 170 minutes, over twice the plan (150),
so the pass parks at planning under the round's standing answer 14, as `round.md` "Parking at
planning" says. The plan is otherwise complete, and its executor prompt assumes an estimate
raised to 90 minutes, so its stop falls at 180. The estimate grew because the round plan priced
neither the prompt lines that pass 1 left to this pass nor the fixtures and the live wording
loop. Design: `docs/specs/2026-11-18-desk-skills-design.md`, sections "The desk opens",
"Rulings", "The two skills and the guide", "The passes", "Left to the builder", and done items
2 and 11. The design wins over this plan. Round plan:
`docs/plans/2026-11-18-desk-skills-round.md`. Where the round is silent, the main,
State-and-rules, Sessions, and Page designs under `docs/specs/` hold.

Anything a person reads (`docs/guide.md`, the PR body, follow-ups) follows
`/var/home/glw907/.claude/docs/voice/written-for-people.md`, loaded through the `writing-voice`
skill before drafting. The two skills follow
`/var/home/glw907/.claude/docs/voice/agent-facing.md`, also loaded through `writing-voice`.

Worktree `/var/home/glw907/Projects/cc-desk/.claude/worktrees/desk-skills`, branch
`desk-skills`, cut from `origin/main` at `0d992b2`. PR base: `main`. Push: `glw907/cc-desk`.
Pass 4 (`desk-page`) runs beside this one. Its sessions aren't this pass's to message, end,
or wait for, and `web/` and the page's unit are its files (standing answer 17).

This pass builds on pass 1's `[advisor]` lines in PR #17 and its `## Decisions`. Three of them
bind the skills: `question add` accepts only a pass that hasn't started, so a pass session
can't ask about its own building pass and stops instead; `--replaces` lands now only for a
pass that hasn't started; and pass 2 owns the allow rules for `question add` and `report add`
(pass 1 plan, standing answer 32).

## Pre-flight

No step in this pass needs Geoff once the estimate is settled. Every write is inside the
worktree, a test's `t.TempDir()`, or `$CLAUDE_JOB_DIR/tmp`. The live script starts `claude -p`
sessions from `/var/home/glw907/Projects/cc-desk-scratch` that can only read (`--tools Read`),
so they write nothing there.

- Permissions checked: `~/.local/state/claude-denials.jsonl`, read again on 2026-11-18 against
  this pass's commands. It holds 135 rows, from 2026-11-01T00:55:05Z to 2026-11-18T21:28Z, all
  inside 30 days. The two rows added since the round plan's read of 133 are a classifier
  outage at 21:25Z and a denied `claude stop d5647ebf` at 21:28Z (Destructive or Unwanted).
  No row refuses `claude -p`, `go build`, `go test`, `go vet`, `gofmt`, `vale`, `tellgrader`,
  `bash scripts/gate.sh`, `git add <path>`, `git merge origin/main`, a push of a pass branch,
  `gh pr create`, `gh run list`, `gh run watch`, an `Agent` call (one row is a classifier
  outage), or an edit under `.github/workflows/`. The rows near this pass are under "Refused
  commands and their alternatives".
- Reviewed: one `standards:plan-reviewer` read of this plan (Opus, medium effort), 24 findings,
  all folded or answered under `## Decisions`. Its estimate finding is the reason this pass
  parks.
- Evidence: the mechanism table, the recorded searches, and the checks under `## Acceptance`,
  each run on today's code (`0d992b2`) before any code was written, with what they printed.
- Launch: the planner commits both plan files on `desk-skills` and pushes them, so
  `claude-worktree-busy` on the worktree exits 0 before an executor enters it.
- No step runs under a systemd unit, so no unit's `PATH` applies. The live script runs from
  the executor's shell, and the binary it calls is the one it builds into
  `$CLAUDE_JOB_DIR/tmp`, never the installed one. The round's standing answer 21 (the real
  state folder) doesn't apply: this pass's live runs read fixtures only.

### Refused commands and their alternatives

| Refused in the log | This pass does instead |
| --- | --- |
| `claude stop <id>` (21:28Z), and a chained `claude stop` | Never needed: every live session is `claude -p`, which ends by itself. A hung one is ended by its `timeout 120` |
| A write or edit to a file named `settings.json` (Self-Modification) | No file named `settings.json` is written, and no fixture is copied wholesale from a folder that holds one. The live sessions take no `--settings` |
| `git commit -qam` and `git commit --no-verify` | `git add <path> ...` by name, then `git commit -m`, with hooks |
| `git push origin HEAD:main`, and any force push | Push only `desk-skills`. A branch behind `main` takes `git merge origin/main`, never a rebase |
| `gh pr merge`, and a `gh pr view` read before a merge (Merge Without Review) | Nothing merges. The executor reads checks with `gh run list --branch desk-skills` |
| `claude agents --json` piped into Python (17 Nov) | Not needed in this pass |

## Mechanism table

| Mechanism | Probe or source |
| --- | --- |
| Tool versions | `go.mod`: `go 1.27`, `github.com/spf13/cobra v1.10.2`. `go version`: `go1.27.4`; `.github/workflows/gate.yml` pins `go-version: 1.27.4` and `PIN_VALE_VERSION: 3.27.0`. `vale --version`: `3.27.0`. `claude --version`: `2.1.327 (Claude Code)`. tellgrader has no version flag (`unknown flag: --version`); its source is `glw907/workstation` at `claude/.claude/skills/writing-voice/evals/tellgrader`, `go.mod` `go 1.27.0` and `cobra v1.10.2` with no `replace`, last changed at commit `c60c2f5858c69a18deffecee9dd530bbfbc1dbee` (2026-11-10), which is on `origin/main` |
| tellgrader gates | `tellgrader --help`: exit 2 only when a finding gates, and only `trailing-hinge-run` and `appositive-stack` gate, and only under the `docs-register` profile. cc-desk has no `.tellgrader.json`, so a bare run can't fail. With `--register agent --profile docs-register`, a three-sentence hinge run printed `trailing-hinge-run ... "gate":true`, exit 2; a stub skill and `pass-core`'s `SKILL.md` exited 0 |
| tellgrader in CI | The runner has no tellgrader. Probed twice in scratch folders: `git clone --filter=blob:none --no-checkout https://github.com/glw907/workstation.git <dir>`, `git -C <dir> sparse-checkout set --no-cone claude/.claude/skills/writing-voice/evals/tellgrader`, `git -C <dir> checkout c60c2f5858c69a18deffecee9dd530bbfbc1dbee`, then `go build -o <bin>/tellgrader ./cmd/tellgrader` inside the tellgrader folder. The build ran, and it printed the same JSON on `pass-core`'s `SKILL.md` as `~/.local/bin/tellgrader`. `go install <module>@<sha>` isn't used: the module path `github.com/glw907/workstation/tellgrader` doesn't match its folder in the repo |
| vale on a skill | `.vale.ini` has `[*.md]` with `BasedOnStyles = Vale, Survey`, so a skill needs no new section. A stub `skills/x/SKILL.md` with YAML frontmatter ran clean under `vale --config .vale.ini --minAlertLevel=error`, exit 0, and a misspelled one exited 1. The gate today runs vale only on `git ls-files 'docs/*.md'`, so an untracked file isn't linted |
| The launch prompt | `internal/launch/prompt.go`: `Prompt` holds `Kind, Name, Project, Round, Pass, Checkout, Worktree, Branch, Base, Verbs, FollowUp, Ready`, and at most one of `Relaunch *RelaunchNote` (`Count int; Cause string`), `Answer`, `Fix`. `Fill` renders `templates/<kind>.txt`, then the note's file, with `missingkey=error`. `pass.txt`, `design.txt`, and `conductor.txt` each load `pass-core` and carry "Until cc-desk runs, a question or a stop goes under `## Stop` in `docs/STATUS.md` on the branch. Never run `approval-queue send`." `TestTemplates` (`internal/launch/launch_test.go:151`) refuses the word "relaunch" in a first launch's prompt |
| A launch's command lines and allow rules | `cmd/cc-desk/launcher.go:231-246`: `launch.Verbs` gives the stop or finish lines, `FollowUpCommand` and `ReadyCommand` give prefixes, each prefix becomes `Bash(<prefix> *)`, and `WriteSettings(dir, binary, name, AllowRules, AskRules, verbs, rules)` writes them. Nothing gives a session a `question add` or `report add` line or rule today, and the session has no other way to learn the binary's path or `--state-dir`. `launch.SessionReasons` (`internal/launch/settings.go:18`) is `usage`, `irreversible`, `secret`, `auth_data`, `scope`: the five stop reasons a session holds rules for, of the ten `pass stop` accepts (`cmd/cc-desk/pass.go:81`) |
| `question add` and `report add` | `question add <project> <round> <file>` (`cmd/cc-desk/question.go:17`) and `report add <project> <file> [--round <round>]` (`cmd/cc-desk/report.go:21`). `question add` refuses a pass that has started (`internal/state/question.go:83`). Neither restricts where its file lies; only `followup` does (`cmd/cc-desk/followup.go:67-84`, worktree or `CLAUDE_JOB_DIR`) |
| The relaunch number | `cmd/cc-desk/launcher.go:251-266`: `count := max(old.Counted(), 1)`, and `+1` only when `req.Counted`. A usage stop relaunches uncounted (`internal/sessions/decide.go:246-249`). So an uncounted usage relaunch with no counted one before it prints "This is relaunch 1 of 2", and one after a counted relaunch repeats that relaunch's number. `answer.txt` and `fix.txt` already say they don't count |
| The launch rig | `e2e/launch_rig_test.go`: `newLaunchRig`, `r.run`, `r.launches(t)[i].Args`, and `relaunchPrompt(c call)` (`e2e/recovery_test.go:49`). A pass launches from `pass start proj a` then `check` (`e2e/launch_test.go:173`); a design writer from `session start design proj r1` (`:213`); a conductor from `session start conductor proj` (`e2e/recovery_test.go:86`). `TestRecoveryPassRowVanishesThreeTimes` (`:135`) asserts "This is relaunch n of 2" for counted relaunches |
| The brief | `brief` prints the forecast in UTC (`cmd/cc-desk/brief.go:52`), each waiting question with its background, answers and causes, recommendation and reason, then a Waits line that holds only the pass name and minutes (`:270-275`), then what could come back in six hours, then the check log since the last desk. A merge question prints `pull request <n>, "<summary>"`, the summary being the pass's (`:233`). Questions sort by held-up minutes. `brief` takes the hidden flag `--now` (`cmd/cc-desk/root.go:118-119`); the `CC_DESK_NOW` variable does nothing. It never prints a ruling id. It opens `<dir>/lock` with `O_CREATE` (`internal/state/lock.go:29-30`), so a run on a tracked fixture leaves a stray `lock` file (seen in a probe) |
| Fixture shapes | `e2e/testdata/desk/releases/beta.json` holds a building pass with `questions:["merge-b4-1"]` and a `pull_request.question`. The reviewer copied `e2e/testdata/state/`, removed `usage.json`, `settings.json`, `snapshot.json`, and `alpha.json`, and ran `brief --now 2026-11-20T23:00:00Z`: it printed `Next question forecast: 2026-11-21T07:00:00Z`, `Waiting: nothing`, and nothing under "Could come back". That is after the clock and outside the six-hour window |
| A fresh session given a skill by path | Probes from the scratch checkout, 2026-11-18. Sonnet at low effort, told in `--append-system-prompt` to read a skill file, didn't read it, and ran `git` and `gh` on its own. Opus 5.5 at high, with `--add-dir <skill folder>` and the system prompt "You are the desk session. Your skill is not installed; it is the file `<path>`. Read that file with the Read tool before your first reply, and follow it.", read the file first and followed it. Adding `--tools Read` kept that, and the session made no other tool call. A stub skill with the nothing-waiting brief above took 11 seconds and said "Nothing is waiting for you; the next question should arrive around 10:00 pm on Friday, November 20." A 153-line stand-in skill with the `state` fixture's full brief took 34 seconds |
| The first message | The design ("The desk opens") has the command pass one short line, such as "What is waiting for me?", because a session speaks only after a first message. The live script sends that line. "Opens without being asked" means the opening needs no second message from Geoff |
| The no-context reader | One Sonnet trial, given a hand-written one-question opening and "State the question this asks, and what each answer would cause.", kept "the Juneau archive", "30 GB", "15 minutes", and "40 minutes" in its reply |
| `claude` flags | `claude --help` on 2.1.327 lists `--add-dir`, `--append-system-prompt`, `--tools` ("" for none, or names such as `Read`), `--model`, `--effort`, `--output-format`, and `-p`. It lists no `--max-turns` and no flag that loads a skill |
| A skill's description | `claude-tooling-sync` sets `DESCRIPTION_CAP = 500` (`~/.dotfiles/bin/.local/bin/claude-tooling-sync:62`) |
| The gate on today's code | `bash scripts/gate.sh` at `0d992b2`, and again with this plan in place: every Go package ok, vale clean, svelte-check 0 errors, 130 Playwright tests passed, `gate: ok`. It takes about 3 minutes here |

## Recorded searches

Each count is `grep -rn` over `cmd/`, `internal/`, `e2e/`, `scripts/`, and `.github/`, on
`0d992b2`. `web/` is pass 4's and holds no launch prompt; `docs/` holds plans and designs,
which no test reads.

| Pattern | Hits | Files |
| --- | --- | --- |
| `RelaunchNote` | 7 | `internal/launch/prompt.go`, `internal/launch/launch_test.go`, `cmd/cc-desk/launcher.go` |
| `launch.Fill(` or `func Fill(` | 2 | `internal/launch/prompt.go`, `cmd/cc-desk/launcher.go` |
| `Verbs(` | 4 | `internal/launch/settings.go`, `internal/launch/launch_test.go`, `cmd/cc-desk/launcher.go` |
| `FollowUpCommand(` | 4 | the same three |
| `ReadyCommand(` | 4 | the same three |
| `WriteSettings(` | 7 | the same three |
| `This is relaunch` | 5 | `internal/launch/templates/relaunch.txt`, `internal/launch/launch_test.go`, `e2e/recovery_test.go` |
| `pass-core` | 7 | `pass.txt`, `design.txt`, `conductor.txt`, `internal/launch/launch_test.go`, `scripts/live/timer-live-a-plan.md`, `scripts/live/timer-live-b-plan.md` |
| `approval-queue` | 11 | the three templates, `internal/launch/settings.go` and `launch_test.go` (the deny rule `Bash(approval-queue land *)`), `internal/state/prompt.go` and `prompt_test.go`, `e2e/launch_test.go`, two settings files in `e2e/testdata/real-2026-11-18/launch/` |
| `Until cc-desk runs` | 3 | the three templates |
| `## Stop` | 4 | the three templates, `internal/launch/launch_test.go` |
| `docs/STATUS.md` | 3 | the three templates |
| `round-run.md` | 4 | the three templates, `internal/launch/launch_test.go` |
| `cc-desk-work` | 0 | none |
| `tellgrader` | 0 | none |
| `causeText(` | 2 | `cmd/cc-desk/launcher.go` |
| `Counted()` | 8 | `cmd/cc-desk/check.go`, `cmd/cc-desk/launcher.go`, `internal/sessions/decide.go`, `internal/snapshot/sessions.go`, `internal/state/launch_test.go`, `internal/state/merge_test.go`, `internal/state/types.go` |
| `skills/` folder | 0 | none |
| `docs/guide.md` | 0 | none |

The `approval-queue` hits outside the templates are the deny rule and the prompt hook's denied
rules. They stay: the deny rule is a guard, not a line that sends Geoff a document. The two
`real-2026-11-18/launch/*.json` files are copies of real state that nothing regenerates. The
two `scripts/live/timer-live-*-plan.md` files are the sessions-timer pass's live fixtures.
Nothing reads their text as a prompt. No golden or settings fixture pins the allow rules:
`e2e/prompt_test.go` matches with `Contains`, and `launch_test.go:60` counts rules from a
direct `WriteSettings` call only.

## File list

From the searches above. "New" marks a file the pass adds.

- `internal/launch/prompt.go`: `RelaunchNote` gains `Uncounted bool`, so every existing literal
  keeps its number; `Prompt` gains `Question` and `Report` prefixes.
- `internal/launch/settings.go`: `QuestionCommand` and `ReportCommand` beside
  `FollowUpCommand`.
- `internal/launch/templates/pass.txt`, `design.txt`, `conductor.txt`, `relaunch.txt`.
- `internal/launch/launch_test.go`: `TestTemplates` and the relaunch case change; new cases for
  the two commands.
- `cmd/cc-desk/launcher.go`: the two new prefixes and their rules, and `Uncounted` on the
  relaunch note.
- `e2e/launch_test.go`: `TestLaunchPrompts` (new test).
- `e2e/testdata/skills-none/`, `e2e/testdata/skills-one/`, `e2e/testdata/skills-two/` (new
  state folders), `e2e/testdata/README.md`.
- `scripts/live/skills/brief-none.txt`, `brief-one.txt`, `brief-two.txt` (new, printed by the
  brief verb), and `e2e/skills_test.go` (new, `TestSkillBriefs`).
- `skills/cc-desk/SKILL.md` (new), `skills/cc-desk-work/SKILL.md` (new).
- `docs/guide.md` (new).
- `scripts/gate.sh`: the skills step.
- `.github/workflows/gate.yml`: the tellgrader build. The round plan's file list for this pass
  doesn't name it, and CI can't lint the skills without it; recorded under
  `## Planning misses` in the PR as `[file-list]`.
- `scripts/live-skills.sh` (new), the live script, and `scripts/live/skills/system.txt` (new).
- `styles/config/vocabularies/cc-desk/accept.txt`, only if vale flags a correct term.
- `docs/follow-ups/launch-review-notes.md`: item 5 dropped, the rest renumbered.
- `.vale.ini`: no change expected; `[*.md]` already covers `skills/*/SKILL.md`.

## The launch prompts

Each of `pass.txt`, `design.txt`, and `conductor.txt` changes in the same four ways:

1. It names the working skill: "Load the `cc-desk-work` skill. It says how to report a pull
   request ready, file a follow-up, write a question, stop, and record a report." The
   `pass-core` line stays, because `pass-core` keeps the inside of a pass.
2. The paragraph that sends a question or a stop to `## Stop` in `docs/STATUS.md` goes. In
   its place, one sentence keeps the ban without the strings the test refuses: where
   `pass-core` sends Geoff a document or writes to STATUS for him, `cc-desk-work` says what to
   do instead, and no session sends him a queue document. The new text holds neither
   `approval-queue` nor "Until cc-desk runs", and no new first-launch sentence holds the word
   "relaunch", which `TestTemplates` refuses.
3. A pass and a design writer get the `question add` line. `QuestionCommand(binary, stateDir,
   project, round)` returns `<binary> --state-dir <dir> question add <project> <round>`, and
   the template prints it followed by `<file>`. A conductor gets the `report add` line.
   `ReportCommand(binary, stateDir, project)` returns `<binary> --state-dir <dir> report add
   <project>`, and the template prints it followed by `<file> [--round <round>]`. Each prefix
   gets its `Bash(<prefix> *)` allow rule, as the follow-up and ready prefixes do. The round is
   in the question prefix, so a session asks only within its own round.
4. The conductor's role sentence ("Conduct the rounds of the project ...") stays as it is.
   Starting the conductor for a report is pass 5's.

`relaunch.txt` reads "This is relaunch N of 2, after <cause>." only for a counted relaunch.
An uncounted one reads "This relaunch, after <cause>, doesn't count toward the two." Both keep
"Carry on from the branch as it stands; the commits on it are yours."

## The fixtures

Three new state folders, each with one release and no `usage.json`, `settings.json`, or
`snapshot.json` (the brief needs none of them). Each is written from the shapes in
`e2e/testdata/state/releases/beta.json`, `desks.json`, and `e2e/testdata/desk/releases/beta.json`,
never by copying a whole folder. The worked clock is `2026-11-20T23:00:00Z`, 2:00 pm in Alaska.

Each fixture question and pull request carries one marked phrase per part: a name or a number
that a paraphrase keeps, such as "the Juneau archive" or "30 GB". Each marked phrase appears
in one part only. Ids are distinctive and checked on word boundaries: questions such as
`q-juneau-7` and `merge-dock-2`, and one ruling id, `r-0042`, seeded into a background
paragraph ("this replaces ruling r-0042") so the no-id check has something to refuse. The pull
request's number is `4817`, used nowhere else. Its title's marked phrase lives in the pass
`summary`, which is what the brief prints. "What waits" is the minute total the Waits line
prints, a number used nowhere else in the fixture.

- `skills-none`: nothing waits, and the forecast is `2026-11-21T07:00:00Z`, 10:00 pm in Alaska
  on November 20 (a round that needs Geoff, with no design, after a building pass, as round
  `b3` of the `state` fixture's `beta`).
- `skills-one`: one question with a background, both answers' causes, a reason, and the pass
  that waits.
- `skills-two`: a merge question on ready pull request 4817 that holds more planned minutes,
  and a plain question that holds fewer. The merge question comes first, so the one opening
  shows both the order and the merge question's three answers.

The README gives each folder one paragraph naming each marked phrase, its part, and the
forecast's expected Alaska time.

The brief verb, run on a copy of each folder at the worked clock with `--now`, prints
`scripts/live/skills/brief-<case>.txt`. `TestSkillBriefs` copies each folder to `t.TempDir()`,
runs the brief there, and fails when the committed file differs, so a later change to the brief
can't leave the fixtures stale. Nothing runs the brief on the tracked folders, which would leave
a `lock` file in them.

## The live script

`scripts/live-skills.sh` follows `scripts/live-check.sh`'s shape: phases, `show` printing each
command, and a log under `$CLAUDE_JOB_DIR/tmp/live-skills/`.

- `run` builds the binary from the worktree into `$CLAUDE_JOB_DIR/tmp`, copies each fixture
  to `$CLAUDE_JOB_DIR/tmp/live-skills/<case>`, prints the three briefs there, and diffs each
  against `scripts/live/skills/`. Then, from `/var/home/glw907/Projects/cc-desk-scratch`, it
  starts one session per fixture, one after another: `timeout 120 claude -p --model
  claude-opus-5-5 --effort high --tools Read --add-dir <worktree>/skills/cc-desk
  --output-format stream-json --verbose --append-system-prompt "<system text>" "What is
  waiting for me?"`. The system text is `scripts/live/skills/system.txt` with `@SKILL_PATH@`
  set to `<worktree>/skills/cc-desk/SKILL.md` and `@BRIEF@` set to the case's brief. The file
  holds the probe's proven sentence and the placeholders, never an absolute path, so pass 3 can
  lift it. The script saves each opening, checks that the first tool call read the skill, and
  grades the opening. The probes took 11 and 34 seconds a session, so the three fit one Bash
  call; the executor runs it with the Bash tool's `timeout: 600000`.
- `reader <file>` grades the no-context reader's answer, saved to a file.

Each check prints one line, `pass <case>: <check>` or `fail <case>: <check>: <what it found>`,
and the script ends with `result: pass` or `result: fail`. The checks:

| Case | Check | Item 2 sentence |
| --- | --- | --- |
| every | The first tool call is a Read of the skill's path | "a fresh session given the desk skill" |
| every | The result text isn't empty, and holds none of "Would you like", "Shall I", "Do you want", or "Should I" | "opens without being asked" |
| every | The opening holds no fixture question id, no `r-0042`, and nothing matching `\br-[0-9]{4}\b` | "holds no ruling id" |
| none | The opening is one sentence (one terminal `.`, `?`, or `!` after trimming) | "the opening is one sentence" |
| none | It matches `\b10:00 ?pm\b`, case ignored, and holds no "tomorrow" | "holds the fixture's forecast time" |
| one | It holds the marked phrase of the background, of each answer's cause, of the reason, and the Waits minute total | "holds a marked phrase from each part" |
| one | A line that ends in `?`, with a blank line or the text's edge on each side, holds the question's marked phrase | "puts the plain question on a separate line" |
| two | The merge question's marked phrase comes before any of the plain question's | "the question that holds more planned minutes comes first" |
| two | It holds the title's marked phrase, holds no `\b4817\b`, and names merge, fix again, and close (case ignored) | "names its pull request by what it does and gives the three answers" |
| reader | The reader's answer holds the question's marked phrase and each answer cause's marked phrase | "a reader with no context ... states the question and what each answer causes" |

The no-context reader is one blocking `Agent` call on Sonnet at medium effort, made by the
executor, never by the script. Its prompt is the one-question opening alone, and the
instruction "State the question this asks, and what each answer would cause." The executor
saves the reply to `$CLAUDE_JOB_DIR/tmp/live-skills/reader.txt` and runs `bash
scripts/live-skills.sh reader <that file>`. A new opening needs a new reader reply. The PR holds
both runs' output and the reader's reply.

## Acceptance

Item 2 and item 11, word for word from the design's done list:

| Item | Task | Check |
| --- | --- | --- |
| 2. **Skills.** Each skill is under 160 lines and passes `vale` and `tellgrader`. The brief verb makes three fixture briefs: nothing waiting, one question, and a question with a merge question. For each, a fresh session given the desk skill and the brief opens without being asked. With nothing waiting, the opening is one sentence and holds the fixture's forecast time. With one question, the opening holds a marked phrase from each part of the fixture question: its background, what each answer causes, the reason for the recommendation, and what waits. It puts the plain question on a separate line and holds no ruling id. A reader with no context, given only that opening, states the question and what each answer causes. With two, the question that holds more planned minutes comes first, and the merge question names its pull request by what it does and gives the three answers. A Go test builds the three launch prompts the way a launch does. It finds `cc-desk-work` in each, and finds neither `approval-queue` nor "Until cc-desk runs". The prompt of a usage relaunch that isn't counted holds no "relaunch 1 of 2", and a counted relaunch's prompt still gives its number. | 1 to 4 | The gate's skills step; `TestSkillBriefs`; `bash scripts/live-skills.sh run` and `reader`; `TestLaunchPrompts` |
| 11. **The gate.** `bash scripts/gate.sh` passes with no skipped test. Items 2 to 5 and 7 to 10 have parts that run live and aren't Go tests. Each live script prints one line for each check and ends with its result. A pass's pull request holds that output from a run on the pass's last commit, and a pass that leaves it out isn't done. | 6 | `bash scripts/gate.sh` run bare on the last commit; the `gate` check on that commit; the live script's two outputs from a run on that commit, in the PR |

Item 2, sentence by sentence, with what each check printed on today's code (`0d992b2`):

| Sentence | Task | Command | Today's code |
| --- | --- | --- | --- |
| Each skill is under 160 lines and passes `vale` and `tellgrader` | 3 | `bash scripts/gate.sh` (its `gate: skills` step) | `ls skills`: `No such file or directory`. `tellgrader --register agent --profile docs-register skills/cc-desk/SKILL.md`: exit 1 |
| The brief verb makes three fixture briefs | 2 | `go test -count=1 ./e2e/ -run '^TestSkillBriefs$' -v` | No fixture fits: see "The brief" and "Fixture shapes" in the mechanism table |
| A fresh session opens without being asked; the nothing-waiting, one-question, and two-question openings | 4 | `bash scripts/live-skills.sh run` | No skill exists, so no session can be given one |
| A reader with no context states the question and what each answer causes | 4 | the `Agent` call, then `bash scripts/live-skills.sh reader <file>` | As above |
| A Go test builds the three launch prompts; `cc-desk-work` in each, neither banned string | 1 | Task 1's check command | `grep -c cc-desk-work` on the three templates: 0, 0, 0. `grep -l -e approval-queue -e "Until cc-desk runs"`: all three |
| An uncounted usage relaunch holds no "relaunch 1 of 2"; a counted one gives its number | 1 | Task 1's check command | `launcher.go:251-266` sets `count := max(old.Counted(), 1)`, so the uncounted prompt reads "This is relaunch 1 of 2" |

Every Go command must show each named test with `--- PASS`. Output with `no tests to run`
fails the item. Item 11 on today's code: `bash scripts/gate.sh` at `0d992b2` printed
`gate: ok`, with 130 Playwright tests passed and no skip.

## Tasks

Each task commits on its own after `bash scripts/gate.sh` passes, run bare, and pushes. A new
file is staged with `git add <path>` before that gate run, so the gate's `git ls-files` lint
reaches it. Each task's minutes include its gate run (about 3 minutes) and exclude waiting on
CI, which runs in the background, as pass 1's plan counted. No task holds code; the sections
above fix the behavior.

### Task 1: the launch prompts (35 minutes)

- Outcome: a pass and a design writer get a `question add` line and its allow rule, a
  conductor gets a `report add` line and its rule, the relaunch note knows whether it counts,
  and the four templates read as "The launch prompts" sets out.
- Files: `internal/launch/prompt.go`, `internal/launch/settings.go`,
  `internal/launch/templates/pass.txt`, `design.txt`, `conductor.txt`, `relaunch.txt`,
  `internal/launch/launch_test.go`, `cmd/cc-desk/launcher.go`, `e2e/launch_test.go`.
- Class: `runner`.
- Check: `go test -count=1 ./internal/launch/ ./cmd/cc-desk/`, then `go test -count=1 ./e2e/
  -run '^(TestLaunchPrompts|TestRecoveryPassRowVanishesThreeTimes)$' -v`, each test showing
  `--- PASS`.
- Test: table cases in `launch_test.go` for `QuestionCommand` and `ReportCommand`, and that
  `WriteSettings` given the new prefixes writes their rules. `TestTemplates` drops its
  `## Stop` and `approval-queue send` wants, adds `cc-desk-work`, and sets `Uncounted` where a
  case needs it. `TestLaunchPrompts` on the launch rig has subtests `pass`, `design`,
  `conductor`, `uncounted_relaunch`, and `counted_relaunch`. Each of the first three reads the
  launch's prompt argument and finds `cc-desk-work` and its `question add` or `report add`
  line, and finds neither `approval-queue` nor "Until cc-desk runs". `uncounted_relaunch`
  stops a launched pass with `pass stop --reason usage`, runs `check`, and finds no "relaunch
  1 of 2" in the relaunch's prompt; `counted_relaunch` finds "This is relaunch 1 of 2". If a
  helper is missing, standing answer 32 holds.

### Task 2: the fixtures and their briefs (25 minutes)

- Outcome: the three state folders and the three briefs exist as "The fixtures" sets out, and a
  test keeps them in step.
- Files: `e2e/testdata/skills-none/`, `skills-one/`, `skills-two/`,
  `e2e/testdata/README.md`, `scripts/live/skills/brief-none.txt`, `brief-one.txt`,
  `brief-two.txt`, `e2e/skills_test.go`.
- Class: `ordinary`.
- Check: `go test -count=1 ./e2e/ -run '^TestSkillBriefs$' -v`, three subtests showing
  `--- PASS`, and `git status --short e2e/testdata` shows no `lock` file.
- Test: `TestSkillBriefs` with one subtest per case: the brief on a copy at the worked clock
  equals the committed file byte for byte. `brief-none.txt` holds `2026-11-21T07:00:00Z` and
  `Waiting: nothing`; `brief-two.txt` lists the merge question first.

### Task 3: the two skills and the gate's skills step (45 minutes)

- Outcome: `skills/cc-desk-work/SKILL.md` and `skills/cc-desk/SKILL.md` exist, and the gate and
  CI fail when a skill has 160 lines or more, fails vale, or gates under tellgrader.
- Files: `skills/cc-desk-work/SKILL.md`, `skills/cc-desk/SKILL.md`, `scripts/gate.sh`,
  `.github/workflows/gate.yml`, and `styles/config/vocabularies/cc-desk/accept.txt` if needed.
- Class: `runner`.
- Check: `bash scripts/gate.sh` prints `gate: skills` and `gate: ok`. Then, on a scratch copy
  of one skill in `$CLAUDE_JOB_DIR/tmp` with a three-sentence hinge run added, the gate's
  tellgrader command exits 2. The `gate` check on the pushed commit is green.
- Test: the gate step lists `git ls-files 'skills/*/SKILL.md'` and fails when it finds none,
  when a file has 160 lines or more, on vale at error level, or when `tellgrader --register
  agent --profile docs-register` exits 2. It fails with "install it, do not skip the gate" when
  `tellgrader` is absent, as the vale step does. CI adds `PIN_TELLGRADER_COMMIT:
  c60c2f5858c69a18deffecee9dd530bbfbc1dbee` and, before the gate, a step that runs the three
  git commands in the mechanism table's "tellgrader in CI" row into `$RUNNER_TEMP/workstation`,
  runs `go build -o "$RUNNER_TEMP/bin/tellgrader" ./cmd/tellgrader` in the tellgrader folder,
  and appends `$RUNNER_TEMP/bin` to `$GITHUB_PATH`.

`cc-desk-work` says how a launched session reports a pull request ready, files a follow-up,
writes a question, stops, and records a report. It holds:

- Run each command line exactly as the launch prompt prints it, since only the prompt knows
  the binary and the state folder.
- Before `question add`, give the question file's text alone to a no-context reader (one
  `Agent` call) and rewrite it until that reader can state the question and what each answer
  causes.
- `question add` takes only a pass that hasn't started. A pass session never asks about its
  own building pass; it stops with `pass stop --reason scope`, or the reason that fits.
- A follow-up's body file lies in the worktree or the job folder.
- The five reasons a session may stop with (`usage`, `irreversible`, `secret`, `auth_data`,
  `scope`), each in a line of when to use it.
- `pass ready`'s four flags and when `--design outside` applies.
- A report's four parts, with an empty list allowed.

`cc-desk` holds what "The two skills and the guide" lists for it, with the design's sample
opening as the model for its wording:

- The opening: what waits, what could come back in the next few hours, what the check settled
  since the last desk; then the first question, in the brief's order, with its background,
  the plain question on a line of its own, each answer and what it causes, the recommendation
  and its reason, and what waits. With nothing waiting, one sentence with the forecast.
- Times in Alaska time as Geoff's clock shows them (`4:10 pm`, "today", "tomorrow" measured
  against Alaska dates), never UTC stamps. The brief prints UTC; Alaska is UTC-9 from the first
  Sunday of November to the second Sunday of March and UTC-8 otherwise, with one worked
  example (`2026-11-21T07:00:00Z` is 10:00 pm on November 20).
- No ids: no question id, no ruling id, no pass name used as a name. A pull request is named
  by what it does, never by its number.
- When to record and when to ask: record at once when his words settle the choice; ask only
  when they leave the answers open. Then say what was recorded, as the design's "Recorded at
  2:43 pm as your ruling, in your words" sample does.
- The words reach `ruling record` unchanged: the session writes them with the Write tool to a
  file, then runs `cc-desk ruling record --project <p> --round <r> --question <q> --option <o>
  -- "$(cat <file>)"`. No shell reads the words themselves, and `--` guards words that start
  with `-`. Whether the desk's allow rule matches a command holding `$(...)` without a prompt
  is unverified; pass 3 proves it or changes the line (standing answer 15).
- The three merge answers and what the check does for each; a wrong record replaced with
  `--replaces`, and what happens by how far the pass got (the design's table).
- A permission prompt answered with `prompt yes` or `prompt no`, and a stopped pass started
  again with `pass start`, both when Geoff says so.
- Outside the desk session the skill answers from the brief, can't record, and says that
  `cc-desk` opens the session that can.

Each skill's frontmatter `name` matches its folder, and its `description` stays under 500
characters.

### Task 4: the live script and its runs (30 minutes)

- Outcome: `scripts/live-skills.sh` runs as "The live script" sets out, and every check
  passes.
- Files: `scripts/live-skills.sh`, `scripts/live/skills/system.txt`, and the two skills if a
  run shows a wording fault.
- Class: `runner`.
- Check: `bash scripts/live-skills.sh run` (Bash tool `timeout: 600000`), then the reader
  `Agent` call, then `bash scripts/live-skills.sh reader <file>`, each ending `result: pass`.
- Test: the script's own lines. A wording fix in a skill reruns all three sessions and the
  reader.

### Task 5: the guide (10 minutes)

- Outcome: `docs/guide.md` tells Geoff what he types, what each part of the page means, what
  "Needs you now" asks of him, and how to fix a wrong ruling.
- Files: `docs/guide.md`.
- Class: `ordinary`.
- Check: `vale --config .vale.ini --minAlertLevel=error docs/guide.md` and
  `tellgrader --register docs --profile docs-register docs/guide.md`, each exit 0.
- Test: none. The page's parts are named as the page shows them on `main` today; where
  pass 4 renames one, its PR updates the guide (standing answer 15).

### Task 6: close (25 minutes)

- Outcome: the PR is ready, with the review folded, CI green on the last commit, and the live
  output from that commit in the body.
- Files: `docs/follow-ups/launch-review-notes.md`, and whatever the review's fold touches.
- Class: `ordinary`.
- Check: the PR body's `## Acceptance`; the `gate` check on the last commit; a fresh
  `bash scripts/live-skills.sh run` and `reader` on that commit.
- Test: none of its own. The order is fixed, so no commit lands after the live evidence:
  1. Reword the follow-up: drop item 5, keep the other seven renumbered, and keep its opening
     paragraph. Commit.
  2. One whole-branch `standards:diff-reviewer` read; fold it, commit, push.
  3. Wait for the `gate` check on that commit to pass.
  4. Run `run` and the reader on that commit.
  5. Edit the PR body with that output (no commit), and mark the PR ready.

  A finding on a later read that is only low falls under standing answer 35 and never forces
  another live run.

Path: tasks 1 to 6 in order, 170 minutes. No task runs beside another: task 1's files are all
the launch's, the skills need the fixtures' briefs, and the live script needs both skills.

## The gate and CI

The gate is `bash scripts/gate.sh`, run bare before every commit and never piped.
`.github/workflows/gate.yml` runs on `pull_request` and on pushes to `main` only, so no run
exists for `desk-skills` until its PR does. The executor opens the draft PR right after task 1's
first push, before the first watch. After each push the watch runs in the background so the
next task isn't held:

1. `gh run list --branch desk-skills --limit 3`
2. `gh run watch <id> --exit-status` on the newest run for the pushed commit, run bare with
   `run_in_background`.

Before the next push, read that watch's result. A red run is fixed before any other task.

## Standing answers

The twelve in `round-run.md` hold, with the round plan's changes and its answers 13 to 25.
This pass adds:

- **26. A live opening fails one check.** Run that case once more. Two fails in a row on the
  same check are a wording fault: fix the skill and rerun all three cases and the reader. A
  rerun whose cause is already traced is skipped.
- **27. A live session reads nothing, or makes a tool call other than the Read.** Check the
  filled system text against `scripts/live/skills/system.txt` and the probe's sentence. If the
  text matches and the session still skips the Read, that's standing answer 26's wording fault.
- **28. `claude -p` is refused, times out at 120 seconds, or hits a usage limit.** A refusal
  falls to the round's standing answer 2. A timeout reruns that case once. A usage limit is
  `pass stop --reason usage` for the executor's own session, as the round plan sets out.
- **29. Item 2's live bar is out of reach.** After a third round of wording fixes that still
  fails, finish task 5 (the guide doesn't depend on it), push, and write the stop under
  `## Stop` in `docs/STATUS.md` on the branch with the last output and the failing checks.
  Then end.
- **30. The gate's tellgrader step fails on wording the agent-facing register allows.** Fix
  the wording. Neither the profile nor the gate changes in this pass.
- **31. The CI tellgrader build fails.** Read the step's log. If the workstation repo moved
  the folder at the pinned commit, the commit is still the one named; fix the path. Never
  drop the step.
- **32. `TestLaunchPrompts` needs a rig change.** Add a helper in `e2e/launch_rig_test.go`;
  change no existing helper's behavior.
- **33. A skill can't hold what it must in under 160 lines.** Cut the reasons before the
  rules, then the examples beyond one. Never move a rule into `pass-core`.
- **34. The desk-page pass changes `scripts/gate.sh` or `.github/workflows/gate.yml` first.**
  Run `git merge origin/main` once it lands and keep both changes. Never rebase, and never
  force push.
- **35. A review after the whole-branch read finds only low findings.** Fold them once or file
  them as follow-ups. They never start another review round.

## Decisions

- [advisor] The three templates keep their `pass-core` line and add `cc-desk-work`, because
  `pass-core` keeps the inside of a pass. The ban on queue documents stays in one sentence
  that holds neither refused string.
- [advisor] This pass adds the `question add` and `report add` lines and rules, as pass 1's
  plan assigned. The conductor's role stays as written; starting it for a report is pass 5's.
- [advisor] The fixtures are three new state folders (standing answer 16) with marked phrases
  that survive a paraphrase, and the two-question case puts the merge question first so one
  opening proves the order and the three answers.
- [advisor] tellgrader gates only under `--profile docs-register`, so the gate names it. CI
  builds tellgrader from the public workstation repo at a pinned commit.
- [advisor] The live sessions run on Opus 5.5 at high with `--tools Read`: the only proven
  form. Sonnet at low skipped the skill.
- [advisor] The path is priced in the unit pass 1 used (task minutes with their gate runs,
  without CI waits) and isn't shaved to fit: at 170 it parks. The plan reviewer priced it at
  190 to 220 with sessions of up to 300 seconds. The measured sessions (11 and 34 seconds) and
  folding the original nine tasks into six (three fewer gate runs) bring it to 170.
- `RelaunchNote` gains `Uncounted`, not `Counted`, so the existing literals keep their number
  without an edit.
- The live script sends "What is waiting for me?" as the first message, as the design's "The
  desk opens" gives it.
- The prompt names the skill and no fallback path. Pass 5's live sessions start before the
  conductor installs `cc-desk-work` at `v0.1.0`, so they may not find it. Pass 5's planner
  decides whether its runs need the skill installed first; this pass's PR says so under
  `## Decisions`.
- The brief reaches the live sessions in the system prompt. Pass 3 may choose a file instead
  ("Left to the builder"), and `system.txt` is what it starts from.
