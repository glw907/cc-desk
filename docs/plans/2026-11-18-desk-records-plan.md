# Desk records: pass plan

Initiative: cc-desk
Round: desk
Pass: desk-records

Class `ordinary`. Planned 120 minutes. The path below is 209 minutes, under twice the plan
(240), so the pass runs and doesn't park. Design: `docs/specs/2026-11-18-desk-skills-design.md`,
sections "What the records gain", "The page", "Rulings" ("A wrong record" and "A question about
a pull request"), "The check's repairs", and done item 1. The design wins over this plan. Round
plan: `docs/plans/2026-11-18-desk-skills-round.md`. Where the round is silent, the main,
State-and-rules, Sessions, and Page designs under `docs/specs/` hold. Anything a person reads
(the PR body, STATUS lines, follow-ups) follows
`/var/home/glw907/.claude/docs/voice/written-for-people.md`, loaded through the `writing-voice`
skill before drafting.

Worktree `/var/home/glw907/Projects/cc-desk/.claude/worktrees/desk-records`, branch
`desk-records`, cut from `origin/main` at `6f761e8`. PR base: `main`. Push: `glw907/cc-desk`.
This is the round's first pass, so no earlier PR carries `[advisor]` lines to build on.

Pass 1 fixes the names of the new verbs, fields, and files, and the format of the brief. Passes
2, 3, and 4 read them from `main` and rename nothing, so every name below is final once this
pass lands. The PR body lists each verb with its arguments, copied from "The verbs".

## Pre-flight

No step in this pass needs Geoff. Every write is inside the worktree, a test's `t.TempDir()`,
or `$CLAUDE_JOB_DIR/tmp`. Task 1 reads `~/.local/state/cc-desk` and never writes it.

- Permissions checked: `~/.local/state/claude-denials.jsonl`, read on 2026-11-18. It holds 133
  rows, from 2026-11-01T00:55:05Z to 2026-11-18T19:38:25Z, so every row is inside 30 days. By
  reason: Classifier unavailable 50; Destructive or Unwanted 14; Self-Modification 11; Interfere
  With Workloads 11; server classifier with no reason 10; Undermining Oversight 5; Unauthorized
  Access 4; Merge Without Review 3; Auto-Mode Bypass 3; Irreversible 2; Git Destructive 2; Create
  Public Surface 2; and single rows under other reasons. Five rows are cc-desk's: a chained
  `sleep 90 &&`, a write of `spikes/4-permissions/settings.json`, and three classifier outages.
  No row refuses `go build`, `go test`, `go vet`, `gofmt`, `bash scripts/gate.sh`, `npm`,
  `git add <path>`, a push of a pass branch, `flock`, or `gh pr create`. The rows that touch
  this pass's commands are in "Refused commands and their alternatives".
- Reviewed: one `standards:plan-reviewer` read of this plan, findings folded (see
  `## Decisions`). The class is `ordinary`, which needs no read before code; the round's
  planner prompt asked for one.
- Evidence: the mechanism table, the recorded searches, and the checks under `## Acceptance`,
  each run on today's code (`6f761e8`) before any code was written. The round plan's premise
  that the real state folder "holds no project" is out of date: it holds one release,
  `cc-desk-scratch`, with a finished round `live` of two merged passes. Item 1's copy test
  therefore covers a real release file, its sessions, and its launch settings files.
- An outside writer: `cc-desk-check.timer` is active and rewrites `check.json`, `usage.json`,
  and `snapshot.json` in the real folder every five minutes, under an exclusive lock on its
  `lock` file. Task 1 copies under a shared lock on that file, so it never reads a file half
  written.

### Refused commands and their alternatives

| Refused in the log | This pass does instead |
| --- | --- |
| A write or edit to a file named `settings.json` (Self-Modification) | No new file is named `settings.json`. The new state fixture carries none. The Go type and test that write `launch/<name>.json` are fine: those files aren't named `settings.json` |
| `git commit -qam` and `git commit --no-verify` | `git add <path> ...` by name, then `git commit -m`, with hooks |
| `git push origin HEAD:main` | Push only `desk-records` |
| A chained `claude stop` | Never needed in this pass. If it is, run `claude stop <id>` alone |
| `sleep N &&` chained before a command | Wait on a background command through its notification, or `Monitor` |
| `gh pr merge` (Merge Without Review) | The pass merges nothing. It ends with the PR ready |
| A classifier outage | Wait 60 seconds and retry once, then go on with the next task and note it |

## Mechanism table

| Mechanism | Probe or source |
| --- | --- |
| Tool versions | `go.mod`: `go 1.27`, `github.com/spf13/cobra v1.10.2`, `pflag v1.0.9`. `go version`: `go1.27.4`, and `.github/workflows/gate.yml` pins `go-version: 1.27.4` and `node-version: 26` (local `v26.13.1`). `web/package.json`: `typescript 6.0.3`, `@playwright/test 1.65.1`, `svelte 5.57.3`, `@sveltejs/kit 3.3.1`. Vale `3.27.0` locally and in CI. jq `1.8.2` |
| The strict reader | Every state file decodes with `DisallowUnknownFields`. Today `round design beta b3` with a question that carries `background` printed `json: unknown field "background"` and exited 1. So each new field needs a Go type, and a file written before this round still reads as long as every new field is optional. It also means no fixture can hold a new field before the task that adds its type |
| A missing list against an empty one | `encoding/json` decodes an absent key and `"calls": null` the same into a `*[]Call`: the pointer stays nil. `"calls": []` makes it a non-nil pointer to an empty slice. So `null` reads as missing. Output needs a non-nil slice and no `omitempty`, or `null` is written |
| The commit path | `state.Commit(dir, now, LockWait, ev, snapshot.Render)` loads, runs the event, settles unless `noSettle`, appends `newRuling`, writes dirty release files, `usage.json`, `check.json`, renders, appends a forecast line on a `passEvent`, writes `desks.json`, then `snapshot.json` (`internal/state/commit.go:58-159`). `f.commit` takes a usage reading first (`commitReading`); `RecordListFailure` shows the path that skips settle. `settle` moves only waiting passes, so a not-started pass stays not started until pass 3's check starts it |
| A refusal's exit code | A missing field or a broken rule is a plain error, exit 1 (`internal/state/events.go:95`, `missing required field "passes"`; `cmd/cc-desk/main.go:14-18`). A refusal by rule returns `state.Refuse`, which the command prints the same way. This pass adds no new exit code |
| A verb that rewrites the real release | On a copy of the real state folder, `release checkout cc-desk-scratch <same path> --default-branch live-2026-11-18` (the values it holds) left the file's sha256 at `9991682…` unchanged. With `--default-branch main-copytest` it changed to `e98a317…`, and a key-path walk gave 95 paths before and 95 after, none missing. The checkout must be a folder that exists, so the test passes a `t.TempDir()` path, which also changes the bytes |
| The clock | Every verb takes the hidden `--now <RFC 3339>` flag (`cmd/cc-desk/root.go:45-53,118`) |
| The settings file | `internal/launch/settings.go` `WriteSettings(stateDir, binary, name, allow, verbs, prefixRules)` builds `Permissions.Allow` from a non-nil `append([]string{}, ...)` and writes `launch/<name>.json`. Its one caller is `cmd/cc-desk/launcher.go:239`. The ask-rule probe (`docs/research/2026-11-18-ask-rule-probe/`) showed an `ask` rule stops a session at a prompt in auto mode |
| A prompt answered "always" | `internal/state/prompt.go:189-190,209-212` appends the prompt's rule to `AllowRules`. With an ask rule in place, that would put one rule in both lists |
| The check's outcomes | `carryOut` in `cmd/cc-desk/carryout.go` gives one string per decision: `launched <id>`, `ended <id>`, `not relaunched: ...`, `failed: <err>`, a no-checkout line, a merge outcome, or `""` for a decision with nothing to carry out. A relaunch's cause is in `Decision.Reason` (`carryout.go:90`). Stops are applied in `RecordCheck`'s commit (`internal/state/checkrun.go`, called at `cmd/cc-desk/check.go:127-129`); outcomes in a later commit (`carryout.go:167-168`). So in one run every stop is written before every outcome |
| The launch rig | `newLaunchRig` reads its session rows from `rows.json` or `lists/<n>.json`; a row's `state` and `waitingFor` (`internal/sessions/list.go:24`) are how a test makes the check stop a record. `check.json` holds only the last run |
| Merge questions | `AskMerge` (`internal/state/merge.go:156-185`) adds a question `merge-<pass>-<k>` to the building pass whose pull request is ready, and sets `pull_request.question` to its id. The pull request record holds `number` and `commit`, with no URL |
| Desk records | `desks.json` keeps every registered desk: `RegisterDesk` replaces an entry by session id and never removes one. So the latest `started` is the last desk session's start |
| Snapshot goldens | `e2e/events_test.go` `TestFixtureSnapshotIsCurrent` compares the `snapshot` verb's output on `e2e/testdata/state` at the worked clock with `e2e/testdata/state/snapshot.json`. `internal/snapshot/snapshot_test.go` has an `-update` flag for `testdata/empty.golden.json` |
| The web type check | `web/tests/snapshot-type.spec.ts` writes every snapshot from `allStates()` in `web/tests/state.ts` as a `Snapshot` literal and runs `tsc` on it, so a printed field missing from `web/src/lib/snapshot.ts` fails the gate with TS2353. Every new field is always printed, so each task that adds a snapshot field adds it to `snapshot.ts` in the same commit. `web/tests/words.spec.ts` checks every code in every printed snapshot has a word entry |
| `go test -run` | The pattern is a Go regexp split on `/`. `\|` matches a literal bar: `go test -run 'TestA\|TestB'` printed `ok ... [no tests to run]`, exit 0 (the reviewer's probe). Every check below uses a bare `|` inside a group |
| The gate on today's code | `bash scripts/gate.sh` at `6f761e8` printed `gate: ok`, exit 0: every Go package ok, svelte-check 0 errors, 130 Playwright tests passed, no skip. It also passed with this plan in place. One local gate takes about 4 minutes; one CI `gate` run took 4 to 11 minutes over the last eight |

## Recorded searches

Each count is over `cmd/`, `internal/`, `e2e/`, `web/src/`, and `web/tests/` with
`grep -rn`, on `6f761e8`.

| Pattern | Hits | Files |
| --- | --- | --- |
| `type Question struct` | 2 | `internal/state/types.go`, `internal/snapshot/snapshot.go` |
| `buildQuestion(` | 3 | `internal/snapshot/snapshot.go` |
| `.Questions = append` | 7 | `internal/state/events.go`, `internal/state/merge.go`, `internal/snapshot/snapshot.go`, `internal/snapshot/nextsession_test.go` |
| `state.Ruling` | 9 | `cmd/cc-desk/ruling.go`, `internal/snapshot/snapshot.go`, `internal/snapshot/snapshot_test.go`, `e2e/writers_test.go`, `e2e/ruling_test.go` |
| `RecordRuling(` | 7 | `cmd/cc-desk/ruling.go`, `internal/state/ruling.go`, `internal/state/ruling_test.go` |
| `AllowRules` | 22 | `internal/state/types.go`, `internal/state/prompt.go`, `internal/state/prompt_test.go`, `cmd/cc-desk/launcher.go` |
| `WriteSettings(` | 6 | `internal/launch/settings.go`, `internal/launch/launch_test.go`, `cmd/cc-desk/launcher.go` |
| `SessionEntry{` | 2 | `internal/snapshot/sessions.go` |
| `NeedEntry{` | 5 | `internal/snapshot/sessions.go` |
| `TestNeedsYouNowStoppedConductor` | 2 | `internal/snapshot/sessions_test.go`, a follow-up |
| `RecordOutcomes(` | 2 | `internal/state/launch.go`, `cmd/cc-desk/carryout.go` |
| `carryOut(` | 2 | `cmd/cc-desk/carryout.go`, `cmd/cc-desk/check.go` |
| `NoNeedReason` | 5 | `internal/state/types.go`, `internal/state/check.go`, `internal/snapshot/snapshot.go` |
| `DesignRound(` | 7 | `cmd/cc-desk/round.go`, `internal/state/events.go`, `internal/state/events_test.go`, `internal/snapshot/sessions_test.go` |
| `DecodeDesign(` | 6 | `cmd/cc-desk/round.go`, `internal/state/events.go`, `internal/state/events_test.go` |
| `Desks.Desks` | 10 | `internal/state/ruling.go`, `internal/state/ruling_test.go` |
| `(==\|!=) *(state\.)?Done\b`, non-test | 9 | see "A closed pass" |
| `case Done`, non-test | 3 | `internal/state/check.go:164`, `internal/state/events.go:244`, `internal/state/merge.go:76` |
| `readJSONL[` | 1 | `internal/state/read.go` |
| `appendLine(` | 3 | `internal/state/commit.go` |
| `noSettle` | 9 | `internal/state/types.go`, `internal/state/commit.go`, `internal/state/checkrun.go`, `internal/state/prompt.go`, `internal/state/prompt_test.go` |
| `kick(` | 4 | `cmd/cc-desk/kick.go`, `cmd/cc-desk/pass.go`, `cmd/cc-desk/ruling.go` |
| `readBody` or `--body-file` | 19 | `cmd/cc-desk/followup.go`, `internal/launch/settings.go`, the launch templates, `internal/launch/launch_test.go`, `e2e/followup_test.go` |
| `snapshot.json` | 52 | across `e2e/`, `web/tests/`, `internal/` |
| Interfaces in `web/src/lib/snapshot.ts` | 18 | one file, 200 lines |
| `allStates` or `builders[` | 7 | `web/tests/state.ts`, `web/tests/snapshot-type.spec.ts`, `web/tests/words.spec.ts` |
| A `Snapshot` literal or cast under `web/` | 5 | `web/tests/surface.spec.ts`, `web/tests/snapshot-type.spec.ts`, `web/tests/state.ts`, `web/tests/words.spec.ts`, `web/src/lib/server/state.ts`. None is handwritten: each parses the binary's output |
| `internal/snapshot/testdata` | 1 | `empty.golden.json` |
| `web/tests/fixtures/twin` | 3 files | `desks.json`, `rulings.jsonl`, `releases/alder.json`, read by `rulings`, `scan`, `items`, and `link` specs through `state.ts` |
| The three follow-ups this pass closes | 0 outside `docs/` | `pass-needs-reason.md`, `snapshot-stop-times.md`, `ended-record-stop-in-needs-you-now.md` |

## File list

From the searches above. "New" marks a file the pass adds.

- `internal/state/types.go`: the new fields and types.
- `internal/state/read.go`, `internal/state/check.go`: reading and checking `reports.json`,
  `check-log.jsonl`, and the new fields.
- `internal/state/commit.go`: writing `reports.json` and appending `check-log.jsonl`.
- `internal/state/events.go`: `DecodeDesign`, `DesignRound`, `start` (never guessed),
  `ClosePass`.
- `internal/state/rules.go`: `Breaks` and a closed pass.
- `internal/state/ruling.go`: `--replaces`, conversations.
- `internal/state/prompt.go`: an "always" answer on an ask rule.
- `internal/state/checkrun.go`, `internal/state/launch.go`: the check log.
- `internal/state/question.go` (new), `internal/state/report.go` (new), and their tests.
- `internal/state/*_test.go` beside each change.
- `internal/snapshot/snapshot.go`, `internal/snapshot/sessions.go`,
  `internal/snapshot/estimate.go`, their tests, `internal/snapshot/testdata/empty.golden.json`.
- `internal/brief/brief.go` (new) and `internal/brief/brief_test.go` (new).
- `internal/launch/settings.go`, `internal/launch/launch_test.go`.
- `cmd/cc-desk/root.go` (the command list), `cmd/cc-desk/ruling.go`, `cmd/cc-desk/launcher.go`,
  `cmd/cc-desk/question.go` (new), `cmd/cc-desk/report.go` (new), `cmd/cc-desk/brief.go` (new).
- `e2e/testdata/real-2026-11-18/` (new, the copy of the real state folder),
  `e2e/testdata/desk/` (new state fixture, grown task by task), `e2e/testdata/inputs/` (new
  input files), `e2e/testdata/README.md`, `e2e/testdata/state/snapshot.json` (generated,
  regenerated).
- `e2e/records_test.go` (new), `e2e/brief_test.go` (new), `e2e/checklog_test.go` (new), and a
  case beside the launch rig's tests for the ask rule.
- `web/src/lib/snapshot.ts`, `web/tests/state.ts` (one builder), and, only if
  `words.spec.ts` asks, `web/src/lib/words.ts`.
- `docs/follow-ups/`: remove the three, add one (task 12).
- `README.md`: the verb list, if it lists verbs.

## The verbs

These names are final. Each runs under the root's `--state-dir`.

| Verb | Arguments | What it does | Reading, guard, settle |
| --- | --- | --- | --- |
| `question add` | `<project> <round> <file>` | Adds one question from a file to a round | `f.commit` (a reading), no desk guard, settles. Records no ruling, never kicks |
| `report add` | `<project> <file> [--round <round>]` | Records a report from a file. No `--round` is the release's report | `f.commit`, no desk guard, settles. Records no ruling |
| `report read` | `<id>` | Marks a report read | `state.Commit` directly with `noSettle`. No reading, no session list, no guard, no `kick` |
| `ruling record` | `<words> [--project P --round R --question Q --option O] [--replaces <id>]` | As today, and `--replaces` names the ruling the new one replaces | The same desk guard as today |
| `brief` | `[<project> <round>]` | Prints the brief. With both arguments, only that round, with its design path, context, and calls | Reads only: loads under the lock, writes nothing, takes no reading |

`report read` refuses, in a cobra `Args` function before `RunE`, an id that is empty, starts
with a dash, or holds a slash. A report id is `rep-` and four digits, as rulings are `r-%04d`.
`brief` takes exactly zero or two arguments (`cobra.MatchAll` over a custom count check).
Every refusal exits 1, as the command's other refusals do.

### The question file

```json
{
  "id": "alpha-q9",
  "text": "Should cc-desk notice by itself that you fixed the problem?",
  "options": ["yes", "no"],
  "recommended": "yes",
  "background": ["One paragraph.", "Another."],
  "causes": {"yes": "What a yes causes.", "no": "What a no causes."},
  "reason": "Why the recommendation.",
  "never_guess": true,
  "passes": ["p14"]
}
```

This is `e2e/testdata/inputs/question-q9.json`, added to `alpha` round `r2` of the `state`
fixture (copied to a temp folder), where `p14` is not started. `passes` names one or more
not-started passes in the named round. Each gains the question in its `questions` list. The
question id is new in the release. `causes` has one entry for each option and no other key. The
four new fields are optional in a release file, so old files read; in the question file
`background`, `causes`, and `reason` are required, since a working session writes the whole
question. `validateRelease` runs on the result before it is swapped in, as `AskMerge` does.

### The report file

```json
{
  "title": "...",
  "summary": "...",
  "calls": [{"call": "...", "chosen": "...", "other": "...", "why": "..."}],
  "misses": [{"miss": "...", "cost": "..."}]
}
```

Each of the four keys is required, and a refusal names the one missing. A key set to `null`
reads as missing. `"calls": []` and `"misses": []` are accepted. A call needs all four of its
fields and a miss both of its, and the refusal names the list, the index, and the field. An
unknown key is refused by the strict reader. A second report for the same project and round (or
for the release, with no round) is refused. The round, when given, must exist in the release. A
release under `releases/done` takes a report as a live one does: the round's report is often
written after its release closes.

## Records

| Record | Change | File |
| --- | --- | --- |
| Question | `background []string`, `causes map[string]string`, `reason string`, `never_guess bool`, all `omitempty` | release file |
| Ruling | `replaces string`, `omitempty` | `rulings.jsonl` |
| Pass | `closed {ruling, time}`, a pointer, `omitempty`; `no_need_reason string`, `omitempty` | release file |
| Release | `ask_rules []string`, `omitempty` | release file |
| Session | `refusals [{time, tool, target, reason}]`, `omitempty`. Pass 5 fills it | release file |
| Round | `design string` (the design's path), `context []string`, `calls []Call`, all `omitempty`. Set from the design file by `DecodeDesign` | release file |
| Report | `{id, project, round (omitempty), title, summary, calls, misses, sent, read (null until read)}`. `calls` and `misses` have no `omitempty` and are written as `[]` when empty | new `reports.json`, `{format, reports}` |
| Design conversation | `{project, round, desk, started}` | `desks.json` gains `conversations`, `omitempty` |
| Check log line | `{format, time, project, record, decision, outcome, reason, detail}`, the last two `omitempty` | new `check-log.jsonl`, append only |

`Call` is one type, `{call, chosen, other, why}`, shared by a round and a report.

### What each change does

1. **Never guessed.** `start` refuses a guess on a question with `never_guess`, with a reason
   that names the question. The snapshot's `guess_outcome` and `free` don't change.
2. **A replacing ruling.** `RecordRuling` with `Replaces` set takes the same guard. It needs the
   four question fields. It refuses when the named ruling doesn't exist, is already replaced by
   another, names a different project, round, or question, or isn't the question's current
   answer. It refuses with "has merged; a new change is needed" when a pass resting on the
   question is done. It refuses with "has started; replacing it while it builds comes in pass
   3" when one is building, stopped, or waiting. Otherwise the answer moves to the new ruling,
   both lines stay in `rulings.jsonl`, and the pass reads `free: true` with no `why_not` naming
   the question. Starting it is the check's, in pass 3. Pass 3 also replaces the second refusal
   with the relaunch.
3. **A closed pass.** It is a mark on a done pass, never a new state, so an older file reads and
   the round's report counts it as finished. `validateRelease` refuses `closed` on a pass that
   isn't done or whose pull request merged. `ClosePass(project, pass, rulingID)` takes a
   building or stopped pass, sets it done with its finish time and actual minutes, closes an
   open stop, and sets `closed`. It refuses a pass that is not started, waiting, or done: only a
   started pass has a pull request to close. Only a unit test calls it; the verb is pass 3's.
   A pass with a closed pass anywhere among its `Before` passes gets no forecast start (null),
   adds no minutes ahead of anything, and isn't a candidate for `next_session`. Its `why_not`
   reads "pass X was closed". Each `Done` reader, with what closed means there:

   | Reader | A closed pass |
   | --- | --- |
   | `internal/state/rules.go:131` `Breaks` | Holds the passes after it: `q.State != Done \|\| q.Closed != nil`, with the reason "was closed" |
   | `internal/snapshot/estimate.go:53`, the pace | Left out: its minutes don't measure a merged pass |
   | `internal/snapshot/estimate.go:152`, the walk ahead | Done, so no minutes ahead; a pass after it gets no forecast |
   | `internal/snapshot/snapshot.go:294`, the round's state | Done: the round can finish |
   | `internal/state/check.go:164`, `:221` | Done, as today |
   | `internal/state/check.go:249` | Done, as today; the new check above refuses a merged closed pass |
   | `internal/state/events.go:244`, `internal/state/merge.go:76` | Done, as today: no event moves a done pass |
   | `internal/sessions/merge.go:57` | Done: no merge |
   | `internal/sessions/decide.go:211`, the cleanup | Not cleaned here: it needs a merge. Pass 3's "Close" ends the session |
   | `internal/sessions/decide.go:230` | Done: its session ends |

4. **Ask rules.** `WriteSettings` gains an `ask []string` parameter after `allow`, and the
   `settings` type gains `Ask []string \`json:"ask,omitempty"\`` in `Permissions`, so a release
   with none writes no `ask` key and never `"ask": null`. The caller passes `rf.AskRules`.
   `validateRelease` refuses an empty rule and a rule that is also in `allow_rules`. In
   `internal/state/prompt.go`, an "always" answer whose rule is in `ask_rules` is recorded as a
   one-time yes and never added to `allow_rules`, so no answer can leave a file every `Load`
   refuses.
5. **An ended record.** In `internal/snapshot/sessions.go`, a conductor or design-writer record
   with a stop is left out of "Needs you now" when `ended` is set, unless its stop is
   `launch_failed`. The filter sits only at the conductor and round calls (`sessions.go:114`
   and `:118`), never inside `add`, which the pass branch at `:138` also calls. A pass is keyed
   on `p.State == Stopped`, never on its session's `ended`: a stopped pass always reads "Needs
   you now", ended session or not.
6. **Stop times and a pass's reason.** The snapshot's pass gains `stopped_at` (the open stop's
   start, or null), `needs_kind` (`questions` when it rests on any, else `nothing`), and
   `needs_reason` (its own `no_need_reason`, else its round's, else null). The session entry
   gains `stopped_at` and `refusals`.
7. **The check log.** One line for each stop `RecordCheck` applies, with the outcome `stopped:
   <reason>`, written in that run's `RecordCheck` commit. One line for each decision whose
   outcome is not empty, in the same commit as `RecordOutcomes`. `failed: ...`, `not relaunched:
   ...`, and a no-checkout line each get one. Each line carries the decision's `reason` and
   `detail` when they are set, so the page and the desk can say "restarted after a dropped
   connection". Within one run, the stop lines come first in decision order, then the outcome
   lines in decision order. The `State` gains the loaded lines and a list of new ones that
   `Commit` appends after the ruling and before the release files. A crash between that append
   and the release write can leave a line for a change that didn't land; the next run's line
   follows it, and the log is a record, never an input. The snapshot carries the last 50 lines
   as `check_log`, oldest first.
8. **Conversations.** `StartConversation(project, round, desk)` adds one entry and refuses a
   second for the same round. `DesignRound` removes the round's entry. Only a unit test calls
   `StartConversation`; the verb is pass 3's. The snapshot carries `design_conversations`.
9. **Reports in the snapshot.** `reports`, every report with its record's fields and a `passes`
   list folded from the release: each pass's name, state, planned and actual minutes, its pull
   request number when it has one, and `closed`. They are folded from `s.Releases` directly,
   since the snapshot skips done releases elsewhere.

### The snapshot

Every new field is always present, as today's are: a list is `[]` and a missing value is
`null`. New top-level fields: `reports`, `design_conversations`, `check_log`. Round gains
`design`, `context`, `calls`. Pass gains `closed`, `stopped_at`, `needs_kind`, `needs_reason`.
Question gains `background`, `causes`, `reason`, `never_guess`. Ruling gains `replaces`, which
stays `omitempty` like its other optional fields, so `web/src/lib/snapshot.ts` marks it
optional. Session entry gains `stopped_at` and `refusals`. Each new field is in
`web/src/lib/snapshot.ts` with the same name, added in the task that prints it.

## The brief

`internal/brief` builds it from the loaded state and the snapshot the binary already computes,
so it reads the same forecast. It is plain text with four sections. The design leaves the
wording to the desk session; this format is what that session reads.

```text
Brief at 2026-11-20T23:00:00Z
Next question forecast: 2026-11-21T02:10:00Z

Waiting: 2 questions

1. alpha / r4, question alpha-q9, holds up 70 planned minutes
   Should cc-desk notice by itself that you fixed the problem?
   Background:
     One paragraph.
     Another.
   Answers:
     yes: What a yes causes.
     no: What a no causes.
   Recommended: yes. Why the recommendation.
   Never guessed.
   Waits: p20 (30 minutes), p21 (40 minutes)

2. beta / b2, merge question merge-b4-1 on pull request 12, "Add the report verbs", at commit 3f2a9c1, holds up 60 planned minutes
   The pull request lies outside its design.
   Answers:
     merge: Merges on the check's next run if every check has passed and the last commit is still 3f2a9c1. Otherwise it asks again.
     fix again: Sends the pass back with your words for one more attempt. If it comes back the same way, this is asked again.
     close: Closes the pull request, ends the session, and records the pass as closed by your ruling.
   Recommended: merge.
   Waits: b4 (60 minutes)

Could come back in the next 6 hours:
  The next decision session is forecast at 2026-11-21T02:10:00Z (alpha / r5).
  alpha / p9 runs on a guess for alpha-q3, 45 minutes.

Since the last desk session, 2026-11-19T16:10:00Z:
  2026-11-20T18:00:00Z alpha pass:p13 launch: launched abcd1234 (reason: dropped connection)
```

- **The opening** is the time and one line, "Next question forecast:", with `next_session`'s
  time, or "none forecast". It prints whether or not anything waits, so a desk with nothing to
  ask can still say when the next question comes.
- **Waiting** lists each unanswered question that a not-started pass rests on, and each
  unanswered merge question: a question whose id is some pass's `pull_request.question`.
- **Holds up** is the sum of planned minutes over a set of passes. The set is every pass that
  rests on the question and isn't done, plus every pass that isn't done and has one of those
  among its `Before` passes, directly or through any chain of passes and rounds. Each pass
  counts once. **Waits** lists that set.
- **The order** is most minutes held up first. Ties go by the forecast's question order (the
  order `estimate.go` gives questions, timed by the earliest resting pass), then by id.
- **A question block** carries the project and round, the id the desk passes to `ruling record`,
  the text, the background, each option with what it causes, the recommendation and its reason,
  "Never guessed." when marked, and what waits. A field the question doesn't have is left out.
- **A merge question's block** names the pull request by its pass's summary, its number, and
  the commit asked about. It prints the three answers in the order merge, fix again, close,
  each with the design's words for what it causes (with the commit filled in), unless the
  question carries its own `causes`.
- **Could come back** lists the next session from `next_session` when its time is within 6 hours
  of now, and each open guess. With neither it prints "Nothing."
- **Since the last desk session** prints the check log lines after the latest `started` in
  `desks.json`, each with its reason when set. With no desk recorded it prints the last 24
  hours, and its heading says so. With no line it prints "Nothing." Pass 3 must build the brief
  before it registers the new desk, or this reads the desk being opened. That goes in the PR's
  `## Decisions` for pass 3.
- **`brief <project> <round>`** prints that round's questions only, and first the round's
  design path, its context paragraphs, and the calls already made. A round with no design yet
  prints "No design recorded yet." in their place. A project or round that doesn't exist exits 1
  naming it.
- **A broken state file** prints the error naming the file and exits 1. It writes nothing, not
  even `snapshot-error.json`, since it isn't a commit.

## Fixtures

1. **`e2e/testdata/real-2026-11-18/`.** Task 1 copies `~/.local/state/cc-desk` into it under a
   shared lock: `flock -s ~/.local/state/cc-desk/lock cp -a ~/.local/state/cc-desk/.
   e2e/testdata/real-2026-11-18/`. It then removes `lock`, `check.lock`, and any `.*.tmp-*`
   file from the copy. `launch/` comes too; if that copy is refused, leave it out and say so in
   `## Decisions`. Nothing else is edited. The copy is taken before any binary from this pass
   runs, and the pass never writes the real folder.
2. **`e2e/testdata/desk/`.** A new state fixture with no file named `settings.json`. It grows
   with the tasks, since the strict reader refuses a field before its type exists. Each task
   adds its part and a README line:
   - task 2: two releases. In `alpha`, question `qa` rests on a pass free now (10 planned
     minutes, nothing after it). In `beta`, question `qb` rests on a pass that comes after a
     building pass (30 minutes, with a 40-minute pass after it). So the forecast puts `qa` first
     and the brief puts `qb` first. Both questions carry every new field, and `qb` is never
     guessed.
   - task 3: an ended conductor with `sign_in_expired` in `alpha`, an ended conductor with
     `launch_failed` in `beta`, a stopped pass whose session is ended, and a pass with its own
     `no_need_reason`.
   - task 4: `check-log.jsonl` with lines before and after the desk's start.
   - task 5: `desks.json` with one desk, and a merge question on a building pass with a ready
     pull request.
   - task 6: `reports.json` with one unread report.
   - task 8: a closed pass and a pass after it.
   - task 9: `ask_rules` on one release.
   - task 10: a round with `design`, `context`, and `calls`, and one conversation in
     `desks.json`.
3. **`e2e/testdata/inputs/`.** `question-q9.json`, `report.json` (the four keys),
   `report-empty.json` (empty `calls` and `misses`), and `design-b3-context.json` (a design with
   `design`, `context`, and `calls`). The missing-key and null cases are made in the test from a
   copy of `report.json`.
4. **`web/tests/state.ts`.** One builder, `desk`, that copies the new fixture, added in task 2,
   so `snapshot-type.spec.ts` checks each new field in the task that adds it.
5. **The goldens.** `e2e/testdata/state/snapshot.json` and
   `internal/snapshot/testdata/empty.golden.json` are generated outputs, and every new field is
   always present, so both are regenerated in each task that adds a snapshot field. Old input
   fixtures stay as they are (standing answer 16). Task 11 checks that nothing else moved.

## Acceptance

Item 1 and item 11, word for word from the design's done list:

| Item | Task | Check |
| --- | --- | --- |
| 1. **Records.** A fixture question with every new field reaches the snapshot unchanged. A report recorded from a file shows in the snapshot as unread, and marking it read records the time. A report file missing its title, its summary, its list of calls, or its list of misses is refused, and each refusal names the part. An empty list is accepted. The check's log holds one line for each thing two fixture runs carried out, in order. A copy of the real state folder taken before this round reads with no error, and after one verb rewrites it, every field it held is still there. An ended record with a stop is absent from "Needs you now", and a failed launch is present. The brief verb, run on a fixture with two waiting questions, puts the one that holds more planned minutes first. It prints each one's background, answers, recommendation, and what waits. A session launched under a fixture release that names a command as always asking has that rule in its settings file. | 1 to 10 | The item 1 command below, one named test or subtest for each sentence, each showing `--- PASS` |
| 11. **The gate.** `bash scripts/gate.sh` passes with no skipped test. Items 2 to 5 and 7 to 10 have parts that run live and aren't Go tests. Each live script prints one line for each check and ends with its result. A pass's pull request holds that output from a run on the pass's last commit, and a pass that leaves it out isn't done. | 12 | `bash scripts/gate.sh` run bare on the last commit, its output in the PR; the `gate` check on that commit. Pass 1 has no live part: every check of item 1 is a Go test |

The item 1 command:

```sh
go test -count=1 ./e2e/ -run '^(TestRecords|TestBrief|TestCheckLog|TestRealCopy|TestAskRule)$' -v
```

The PR shows its output, and each subtest in the table below appears with `--- PASS`. Output
with `no tests to run`, or a named subtest missing, fails the item.

Item 1, sentence by sentence:

| Sentence | Task | Test | Today's code (`6f761e8`) |
| --- | --- | --- | --- |
| A question with every new field reaches the snapshot unchanged | 2 | `TestRecords/question_fields`: on a copy of the `state` fixture, `question add alpha r2 question-q9.json`, then the snapshot's question equals the file's fields | `question --help`: `unknown command "question" for "cc-desk"`, exit 1. A design whose question carries `background`: `json: unknown field "background"`, exit 1 |
| A report shows as unread, and marking it read records the time | 6 | `TestRecords/report_read`: `report add`, the snapshot's `read` is null, `report read rep-0001` at a pinned `--now`, `read` is that time; a second read keeps it and exits 0 | `report --help`: `unknown command "report"`, exit 1 |
| A report missing a part is refused, naming it; an empty list is accepted | 6 | `TestRecords/report_parts`: four missing-key cases and one `null` case, each exits 1 with stderr naming `title`, `summary`, `calls`, or `misses`, folder bytes unchanged; `report-empty.json` exits 0 | As above |
| The check's log holds one line per thing two runs carried out, in order | 4 | `TestCheckLog` on the launch rig: run 1 launches a free pass; the test reads `check.json`; it then edits the launched session's row to `state: blocked` with a `waitingFor`; run 2 stops it; the test reads `check.json` again. `check-log.jsonl` equals, run by run, the stops then the non-empty outcomes of each `check.json`, in decision order | No `check-log.jsonl` is ever written: the fixture runs left none |
| A copy of the real state folder reads, and after one verb rewrites it every field is still there | 1 | `TestRealCopy`: copy the fixture to a temp folder, `snapshot --now <the copy's check.json last_run>` exits 0, `release checkout cc-desk-scratch <t.TempDir()> --default-branch main-copytest` exits 0 and changes the release file's bytes, then for every JSON and JSONL source file (all but `snapshot.json`, which the clock shapes) each key path before is present after | `snapshot` on a copy: exit 0. The verb with a changed branch: exit 0, 95 key paths before and after, none missing |
| An ended record with a stop is absent from "Needs you now", and a failed launch is present | 3 | `TestRecords/needs_you_now` on the `desk` fixture, and unit cases in `sessions_test.go`: an ended conductor with `sign_in_expired` is absent; an ended conductor with `launch_failed` is present; a stopped pass with an ended session is present | An ended conductor with `sign_in_expired` gave `[{"project": "alpha", "session": "x-conductor", "reason": "sign_in_expired", ...}]`: present, so the check fails today |
| The brief puts the question holding more minutes first, and prints each one's background, answers, recommendation, and what waits | 5 | `TestBrief` on the `desk` fixture: `qb`'s block comes before `qa`'s, the snapshot shows `qa`'s pass with the smaller `minutes_before` (so forecast order disagrees), each block holds its background lines, both answers with causes, the recommendation and reason, and the waits list, and the merge question's block names its pull request and the three answers | `brief --help`: `unknown command "brief"`, exit 1 |
| A session launched under a release with an ask rule has it in its settings file | 9 | `TestAskRule` on the launch rig: a release with `ask_rules: ["Bash(date *)"]`, `check` launches its pass, `launch/<name>.json` has `permissions.ask` equal to the list; a release with none has no `ask` key | `release add` of a release with `ask_rules`: `json: unknown field "ask_rules"`, exit 1 |

Item 11 on today's code: `bash scripts/gate.sh` at `6f761e8` printed `gate: ok`, exit 0, with
every Go package ok, svelte-check 0 errors, and 130 Playwright tests passed.

## Tasks

Each task commits on its own after `bash scripts/gate.sh` passes, run bare, and pushes. A task
that prints a new snapshot field adds it to `web/src/lib/snapshot.ts` (and any word
`words.spec.ts` asks for) and regenerates `e2e/testdata/state/snapshot.json` (the `snapshot`
verb on a copy of the fixture at the worked clock) and the empty golden (`go test
./internal/snapshot/ -update`), all in the same commit. No task holds code; the sections above
fix the behavior. The order puts the check log and the brief, which passes 2 and 3 need most,
ahead of the parts they don't read.

### Task 1: the real-state copy (12 minutes)

- Outcome: `e2e/testdata/real-2026-11-18/` holds the real folder as it is today, and
  `TestRealCopy` passes on it.
- Files: `e2e/testdata/real-2026-11-18/`, `e2e/records_test.go`, `e2e/testdata/README.md`.
- Class: `ordinary`.
- Check: `go test -count=1 ./e2e/ -run '^TestRealCopy$' -v`.
- Test: `TestRealCopy` as in the acceptance table. The key-path walk is a helper that task 11
  reuses for the golden check.

### Task 2: question fields and `question add` (22 minutes)

- Outcome: a question carries the four new fields into the snapshot; `question add` adds one
  from a file; a guess start on a never-guessed question is refused. The `desk` fixture and its
  web builder start here.
- Files: `internal/state/types.go`, `internal/state/question.go`, `internal/state/events.go`,
  `internal/state/check.go`, their tests, `internal/snapshot/snapshot.go`,
  `cmd/cc-desk/question.go`, `cmd/cc-desk/root.go`, `e2e/testdata/inputs/question-q9.json`,
  `e2e/testdata/desk/`, `e2e/records_test.go`, `web/src/lib/snapshot.ts`, `web/tests/state.ts`,
  the goldens.
- Class: `ordinary`.
- Check: `go test -count=1 ./internal/state/ -run 'Question' -v`, then
  `go test -count=1 ./e2e/ -run '^TestRecords$/^question' -v`. A `-run` pattern splits on `/`
  at every level, so a unit pattern and a subtest pattern never share one command.
- Test: a table over the question file: each required key missing, a `causes` key that is no
  option, a missing option's cause, a duplicate id, a pass in another round, a started pass,
  no passes. Each is refused and names the part. A guess on a never-guessed question is
  refused. An old release file with no new field reads.

### Task 3: refusals, stop times, a pass's reason, and ended records (15 minutes)

- Outcome: items 5 and 6 of "What each change does".
- Files: `internal/state/types.go`, `internal/state/events.go` (the design file's pass reason),
  `internal/snapshot/snapshot.go`, `internal/snapshot/sessions.go`, their tests,
  `e2e/testdata/desk/`, `e2e/records_test.go`, `web/src/lib/snapshot.ts`, the goldens.
- Class: `ordinary`.
- Check: `go test -count=1 ./internal/snapshot/ -run 'NeedsYouNow|StopTime|PassReason' -v`,
  then `go test -count=1 ./e2e/ -run '^TestRecords$/^needs_you_now' -v`.
- Test: the three needs-you-now cases; `stopped_at` on a stopped pass and null after it
  resumes; `needs_reason` from the pass, then the round, then null; `refusals` reads and shows.

### Task 4: the check log (20 minutes)

- Outcome: item 7 of "What each change does".
- Files: `internal/state/types.go`, `internal/state/read.go`, `internal/state/commit.go`,
  `internal/state/checkrun.go`, `internal/state/launch.go`, their tests,
  `internal/snapshot/snapshot.go`, `e2e/checklog_test.go`, `e2e/testdata/desk/`,
  `web/src/lib/snapshot.ts`, the goldens.
- Class: `ordinary`.
- Check: `go test -count=1 ./internal/state/ ./e2e/ -run 'CheckLog' -v`.
- Test: `TestCheckLog` as in the acceptance table; a unit test that a failed and a
  not-relaunched outcome each log a line with their reason; the order rule within a run; the
  snapshot holds at most 50 lines.

### Task 5: the brief (35 minutes)

- Outcome: `brief` prints "The brief" exactly.
- Files: `internal/brief/brief.go`, `internal/brief/brief_test.go`, `cmd/cc-desk/brief.go`,
  `cmd/cc-desk/root.go`, `e2e/testdata/desk/`, `e2e/brief_test.go`.
- Class: `ordinary`.
- Check: `go test -count=1 ./internal/brief/ ./e2e/ -run 'Brief' -v`.
- Test: `TestBrief` as in the acceptance table; a unit table for "holds up" (direct, through a
  pass, through a round, a pass counted once, a done pass left out, the tie-break); a merge
  question's block; the opening with nothing waiting holds the forecast time, and "none
  forecast" with none; the round form, with and without a design; "Nothing." in each empty
  section; a broken release file exits 1 naming it and writes nothing.

### Task 6: reports (25 minutes)

- Outcome: `reports.json`, `report add`, `report read`, and the snapshot's `reports`.
- Files: `internal/state/types.go`, `internal/state/report.go`, `internal/state/read.go`,
  `internal/state/commit.go`, their tests, `internal/snapshot/snapshot.go`,
  `cmd/cc-desk/report.go`, `cmd/cc-desk/root.go`, `e2e/testdata/inputs/report*.json`,
  `e2e/testdata/desk/`, `e2e/records_test.go`, `web/src/lib/snapshot.ts`, the goldens.
- Class: `ordinary`.
- Check: `go test -count=1 ./internal/state/ -run 'Report' -v`, then
  `go test -count=1 ./e2e/ -run '^TestRecords$/^report' -v`.
- Test: the acceptance cases; a call or a miss missing a field; an unknown key; a second
  report for one round refused; a round that doesn't exist refused; a done release takes a
  report; `report read` with `""`, `-x` (after `--`), and `a/b` refused before the state folder
  is touched (its bytes unchanged); `report read` writes no `usage.json` change and runs no
  stand-in command; a broken `reports.json` names the file.

### Task 7: `ruling record --replaces` (15 minutes)

- Outcome: a ruling can replace another as "What each change does" item 2 says.
- Files: `internal/state/types.go`, `internal/state/ruling.go`, `internal/state/ruling_test.go`,
  `cmd/cc-desk/ruling.go`, `e2e/ruling_test.go`, `web/src/lib/snapshot.ts`.
- Class: `ordinary`.
- Check: `go test -count=1 ./internal/state/ ./e2e/ -run 'Ruling|Replace' -v`.
- Test: each refusal in item 2; a good replacement before the pass starts leaves two lines,
  the second with `replaces`, the question's answer names the second, and the pass reads
  `free: true` with no `why_not` naming the question.

### Task 8: a closed pass (18 minutes)

- Outcome: `closed` reads and checks, `ClosePass` exists, and every reader in the table treats
  a closed pass as it says.
- Files: `internal/state/types.go`, `internal/state/events.go`, `internal/state/rules.go`,
  `internal/state/check.go`, `internal/snapshot/estimate.go`, `internal/snapshot/snapshot.go`,
  their tests, `e2e/testdata/desk/`, `web/src/lib/snapshot.ts`, the goldens.
- Class: `ordinary`.
- Check: `go test -count=1 ./internal/... -run 'Close|Breaks|Pace|Forecast' -v`.
- Test: `ClosePass` on a building pass and on a stopped one; refused on a not-started, waiting,
  or done pass; the pass after it doesn't start, reads "was closed" in `why_not`, and has a null
  forecast; it isn't `next_session`'s; the round reads done; the pace leaves it out; `closed` on
  a not-done pass or a merged one is refused.

### Task 9: ask rules (12 minutes)

- Outcome: a release's `ask_rules` reach `launch/<name>.json` under `permissions.ask`, and an
  "always" answer can't put a rule in both lists.
- Files: `internal/state/types.go`, `internal/state/check.go`, `internal/state/prompt.go`,
  `internal/state/prompt_test.go`, `internal/launch/settings.go`,
  `internal/launch/launch_test.go`, `cmd/cc-desk/launcher.go`, the launch rig's test,
  `e2e/testdata/desk/`.
- Class: `ordinary`.
- Check: `go test -count=1 ./internal/launch/ ./internal/state/ ./e2e/ -run 'TestWriteSettings|^TestAskRule$|Prompt' -v`.
- Test: `TestWriteSettings` with and without ask rules (no `ask` key without); `TestAskRule` as
  in the acceptance table; a rule also in `allow_rules` is refused; an "always" answer on an
  ask rule records a one-time yes and leaves `allow_rules` alone.

### Task 10: a round's design, context, and calls, and conversations (12 minutes)

- Outcome: the round fields from the design file, the conversation record, and their snapshot
  fields.
- Files: `internal/state/types.go`, `internal/state/events.go`, `internal/state/ruling.go`
  (conversations beside desks), their tests, `internal/snapshot/snapshot.go`,
  `e2e/testdata/inputs/design-b3-context.json`, `e2e/testdata/desk/`,
  `web/src/lib/snapshot.ts`, the goldens.
- Class: `ordinary`.
- Check: `go test -count=1 ./internal/state/ ./internal/brief/ ./e2e/ -run 'Design|Conversation|Brief' -v`.
- Test: `round design` with the new input carries the three fields to the snapshot and to the
  round form of the brief; the old `design-b3.json` still works; `StartConversation` twice for a
  round is refused; `DesignRound` removes the entry; an old `desks.json` reads.

### Task 11: the golden check (8 minutes)

- Outcome: the branch's changes to generated files only add, and no old input fixture moved.
- Files: `e2e/testdata/README.md`, a test beside `TestFixtureSnapshotIsCurrent` if the check
  is written as one.
- Class: `ordinary`.
- Check: the golden check below, then `bash scripts/gate.sh`.
- Test: every key path in `e2e/testdata/state/snapshot.json` at `6f761e8` (`git show
  6f761e8:<path>`) is present with the same value in the new one, and the same for
  `empty.golden.json`. `git diff --stat 6f761e8 -- e2e/testdata web/tests/fixtures` lists no
  file other than the two goldens, `e2e/testdata/README.md`, and the new folders.

### Task 12: close (15 minutes)

- Outcome: the PR is ready, with the review folded and CI green on the last commit.
- Files: `docs/follow-ups/` (remove `pass-needs-reason.md`, `snapshot-stop-times.md`, and
  `ended-record-stop-in-needs-you-now.md`; add
  `designed-round-with-no-questions-reads-nothing.md` for the second paragraph of
  `snapshot-stop-times.md`, which this pass doesn't do), `README.md` if it lists verbs,
  `docs/STATUS.md` if the executor reference asks.
- Class: `ordinary`.
- Check: the PR body's `## Acceptance`; the `gate` check on the last commit.
- Test: none of its own.

Path: tasks 1 to 12 in order, 209 minutes. The sum over every task is the same; no task runs
beside another, since nearly all touch `types.go` and the goldens. The minutes count work, not
waiting on CI, which runs in the background.

## The gate and CI

The gate is `bash scripts/gate.sh`, run bare before every commit. The CI watch, after each
push, runs in the background so the next task isn't held:

1. `gh run list --branch desk-records --limit 3`
2. `gh run watch <id> --exit-status` on the newest run for the pushed commit, run bare with
   `run_in_background`.

Before the next push, read that watch's result. A red run is fixed before any other task.

## Standing answers

The twelve in `round-run.md` hold, with the round plan's changes and its answers 13 to 25. This
pass adds:

- **26. A name reads wrong once coded.** The names in "The verbs", "Records", "The snapshot",
  and "The brief" are fixed for passes 2 to 4. Change one only if it collides with an existing
  name, keep this plan's tables in step in the PR body, and list it under `## Decisions`.
- **27. A golden changes a value, not only adds one.** Find why. If a new rule changes a
  fixture's meaning (for example a pass now held by a closed pass), that's a bug in the fixture
  reading, not a reason to accept it. Only `generated_at` may differ, and it doesn't at the
  worked clock.
- **28. `words.spec.ts` fails on a new code.** Add the word in `web/src/lib/words.ts` as the
  page's other codes read, and list it. Standing answer 15 covers the shared file.
- **29. The copy of the real folder is refused, or the folder changed shape since this plan.**
  Copy what can be copied. If `releases/` can't be copied, task 1's check can't run: write the
  stop. If the shared lock waits more than a minute, take the lock again once; the timer's run
  is short.
- **30. A verb seems to need a usage reading or a session list that this plan says it skips.**
  It doesn't: "Mark as read" takes neither, by the design. Find another way, or write the stop.
- **31. The path runs past 240 minutes.** Finish the task in flight, push, and write the stop,
  naming the tasks left.
- **32. A check on the live check or a session's allow rules looks needed.** Not this pass's.
  Pass 2 adds the allow rules that let a working session run `question add` and a conductor run
  `report add`. Pass 5 fills `refusals` and starts the conductor.
- **33. A review after the whole-branch read finds only low findings.** Fold them once or file
  them as follow-ups. They never start another review round.

## Decisions

- [advisor] A closed pass is a mark on a done pass, never a new state.
- [advisor] `--replaces` lands now only for a pass that hasn't started; building, stopped, and
  waiting are refused until pass 3, and merged is refused for good.
- [advisor] The ended-record rule applies to conductor and design-writer records only, except
  `launch_failed`; a pass is keyed on its own state.
- [advisor] Reports live in `reports.json` beside the releases, so the snapshot can fold a done
  release's report.
- [advisor] The check log logs every non-empty outcome and every stop `RecordCheck` applies,
  and the snapshot carries the last 50.
- [advisor] The goldens are regenerated, checked by a key-path subset, not by "the diff only
  adds lines": a new last member rewrites the line before it with a comma.
- This pass takes the never-guessed refusal in `start`, which no other pass owns. Pass 2 owns
  the allow rules for the question and report verbs.
- `question add` accepts only not-started passes. A waiting pass is waiting on its limits, and
  adding a question under it would need settle to move it back. So a pass session can't ask
  about its own building pass with it; it stops instead, as today. Pass 2's skill wording
  should say so.
- `brief` takes no usage reading and writes nothing: it is a read.
- Pass 3 builds the brief before it registers the desk.
- The check log line carries the decision's reason, so a relaunch's cause reaches the page.
  The session entry gains no list of relaunch causes; the page reads them from `check_log`.
- From the plan review, folded:
  - Each task adds its snapshot fields to `web/src/lib/snapshot.ts` in its own commit, or
    the type check fails the gate on every commit between.
  - The brief lists merge questions with the three answers, and always opens with the next
    forecast question time: done item 2 needs both, and the format is final here.
  - A replaced ruling leaves the pass free, not started: settle moves only waiting passes.
  - The item 1 command uses a group, since `\|` in `-run` matches nothing.
  - An "always" answer on an ask rule is a one-time yes, or the release would hold one rule
    in both lists and no `Load` would pass.
  - The `desk` fixture grows task by task, the real copy is taken under the state lock, a pass
    after a closed one has no forecast, the check log orders stops before outcomes within a
    run, refusals exit 1 as elsewhere, and the question input names `alpha` `p14`.
- The reviewer priced the pass at 360 to 420 minutes from the repo's PR open-to-merge times.
  Those times are mostly waiting for Geoff's batched approvals (PRs #7 to #16 all merged between
  18:00Z and 19:25Z on 2026-11-18), and the state-rules passes of similar scope (PRs #3 to #6)
  ran 15 to 25 minutes from opening to merge. So the plan keeps 120 planned minutes and prices
  the path at 209, with CI off the critical path. The order puts the check log and the brief
  early, so a stop at 240 leaves the parts passes 2 and 3 need most already landed.
- For the conductor: the real state folder holds `cc-desk-scratch`'s round `live`, finished with
  no report. By the design's rule ("a round's report is due when its last pass merges") pass 5's
  check would start a conductor for it once installed. Pass 5's planner should decide whether a
  round finished before reports existed is due. It doesn't block pass 1.
- Pass 4 reads the pass page's `needs_kind` and `needs_reason` in place of the round's, the
  page half of `pass-needs-reason.md`.
