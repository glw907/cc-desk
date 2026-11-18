# The waiting order and caps: pass plan

Initiative: cc-desk
Round: state-rules
Pass: state-rules-order

Class `ordinary`. Planned 90 minutes; the path below is 175, under the round's 180-minute park
line. Pass 2's path was 115 and it ran in about 17 minutes of wall clock (plan commit 21:16,
last code commit 21:32 on 2026-11-17), so these minutes sit on the same scale as pass 2's.
Design: `docs/specs/2026-11-17-state-and-rules-design.md` (approved 2026-11-17), sections
"What is free to start" and "The usage stop", with "Settling" from "What the command does in
this round". Round plan: `docs/plans/2026-11-17-state-rules-round.md`. Where the round design is
silent, `docs/specs/2026-11-17-cc-desk-design.md` holds. Writing a person reads (the PR body,
STATUS lines, follow-ups) follows `~/.claude/docs/voice/written-for-people.md`, loaded through
the `writing-voice` skill.

Worktree `/var/home/glw907/Projects/cc-desk/.claude/worktrees/state-rules-order`, branch
`state-rules-order`, cut from `origin/main` at `480eecf` (passes 1 and 2 landed as PRs #3 and
#4). PR base: `main`. Push: `glw907/cc-desk`.

## Pre-flight

No step in this pass needs Geoff. Every write is inside the worktree, a test's `t.TempDir()`, or
`$CLAUDE_JOB_DIR/tmp`. The one live usage read (task 8) is a zero-cost local command.

- Permissions checked: `~/.local/state/claude-denials.jsonl`, read again on 2026-11-17 for this
  pass. 130 rows, 2026-11-01T00:55:05Z to 2026-11-18T06:27:23Z. The one row newer than pass 2's
  read is an `echo` into a memory index in another repo (Undermining Oversight). Classes and
  rows: Classifier unavailable 47; Destructive or Unwanted Actions 14; Self-Modification 11;
  Interfere With Workloads 11; server-side classifier, no reason 10; Undermining Oversight 5;
  Unauthorized Access 4; Merge Without Review 3; Auto-Mode Bypass 3; Irreversible Local
  Destruction 2; Git Destructive 2; Create Public Surface 2; Unverifiable Deletion Scope 2;
  Sensitive-Source Provenance 2; Security Weaken 2; Production Deploy 1; Secret-Store Writes 1;
  DNS/Domain/Cert Changes 1; Bot Check Circumvention 1; Code from External 1; Shared Scratch
  Sweep 1; Safety Bypass Flag 1; Out-of-Place Publication 1; PII Data Handling 1; Modify Shared
  Resources 1. This pass adds no dependency and runs `go build`, `go test`, `gofmt`, `go vet`,
  `vale`, `bash scripts/gate.sh`, `git push origin state-rules-order`, `gh pr create`, `gh pr
  ready`, `gh pr checks`, `gh run view`, and once, `cc-desk usage --raw`, which runs `claude -p
  "/usage" --model haiku --output-format json`. No row refuses any of them. The rows near them
  are in the table below.
- Reviewed: one `standards:plan-reviewer` read (Opus, medium effort) of this plan at `439de32`,
  verdict "ready after folds", 2 blockers and 13 other findings, all folded; the `[review]` line
  under `## Decisions` lists them. The class needs none before code; the launch prompt asked
  for one.
- Evidence: the mechanism table and the recorded searches below. Each acceptance check was run
  on today's code before any code was written; the outputs are under `## Acceptance`.

### Refused commands and their alternatives

| Refused in the log | This pass does instead |
| --- | --- |
| `kill <pid>` before `go test` (Interfere With Workloads, 2026-11-09) | Stop a background command with `TaskStop` |
| `rm -rf /tmp/tmp.*` sweeps (Shared Scratch Sweep) | Scratch lives in `t.TempDir()` or a fresh, named folder under `$CLAUDE_JOB_DIR/tmp`. No sweep, no `rm -rf` |
| `git push origin HEAD:main` (2026-11-08, 2026-11-10) | Push only `state-rules-order` |
| `git stash drop` | No stash. A temporary commit sets work aside |
| A chained `claude stop`, `claude agents --json` piped to a script, `claude --bg` in an outage (2026-11-17) | Not needed by the executor. The live usage read runs alone, unchained, through the built binary |
| A chained `claude-worktree-busy` with `ps` (Interfere With Workloads) | Not needed by the executor |
| A write to a Claude `settings.json` (Self-Modification) | No settings edit. The cc-desk `settings.json` in a test's temp folder is a state file, not a Claude setting. A task that needs a Claude settings edit parks (round-run standing answer 3) |
| A classifier outage | Wait 60 seconds and retry while the cause is still the outage. A task that can't finish parks by standing answer 31 |
| `gh pr merge`, and `gh pr view` used to judge a merge (Merge Without Review) | The pass merges nothing. It ends with the PR ready. Read checks with `gh pr checks <n>` |

## Mechanism table

| Mechanism | Probe or source |
| --- | --- |
| Base | `git rev-parse --short HEAD` and `git rev-parse --short origin/main` each printed `480eecf` after a fetch; branch `state-rules-order` |
| The gate on today's code | `bash scripts/gate.sh` printed `ok` for all five packages (e2e in 11.461s), then `gate: ok`. CI on `main` at `480eecf`: run 44252534393, `success` |
| The gate covers this plan | `scripts/gate.sh` runs `gofmt`, `go vet`, `go test -count=1 -json ./...` through `jq` (any skipped test fails it), and vale over every tracked `docs/*.md`. Every commit of a `docs/` file runs vale first |
| CI has no `claude` | `.github/workflows/gate.yml` installs Go 1.27.4 and Vale 3.27.0 on Ubuntu and runs the gate. No step installs `claude`, so no test may reach the real usage command |
| One path for every verb | `internal/state/commit.go:48` `Commit(dir, now, wait, ev, render)`: lock, strict `Load`, remove `snapshot-error.json`, `ev`, `settle(s, now)` at `:76` (empty at `:113`), write dirty releases, `render`, append a forecast line when `passEvent` is set, write `snapshot.json`. An `ev` error returns before any write. Its one caller is `cmd/cc-desk/root.go:63`, through `rootFlags.commit` |
| The snapshot verb writes | `cmd/cc-desk/snapshot.go` runs `f.commit(nil)`, so it settles and rewrites `snapshot.json` like any verb |
| Exit codes | `state.Refuse(msg)` (`internal/state/refuse.go:12`) gives a `*RefusedError`, which `main` maps to exit 2 (`cmd/cc-desk/main_test.go:32-33`). Any other error exits 1 |
| Start today | `internal/state/events.go:151` `StartPass` starts a `not_started` pass and resumes a `stopped` one; any other state is refused at `:169` ("pass %q is %s and cannot start"). It checks no rule |
| Waiting and guess fields | `internal/state/types.go:75-79` `Pass.Limit *Limit` (`Kind`, `Since`) and `Pass.Guess *Guess` (`Question`, `Recommendation`, `At`), unwritten. `events.go:44` and `:116` reject a new release or design whose pass carries either |
| Pass checks | `internal/state/check.go:143` treats `not_started` and `waiting` alike: no start, finish, or actual minutes. Nothing checks `design_written` |
| Usage and settings today | `internal/state/read.go:173` defaults the stop to `DefaultUsageStop` (100, `types.go:178`); `:207-211` requires `usage_stop_percent` 1 to 100 when `settings.json` exists; `:216-226` reads `usage.json` strictly (format, `read_at` required) into `s.Usage`, nil when absent. Nothing reads usage live |
| The snapshot's fields today | `internal/snapshot/snapshot.go:86-87` `Free *bool`, `WhyNot []string` (always null and `[]`); `:29` `Waiting []WaitingPass` (`Project`, `Pass`, `Limit`, `ResetText`, `:131-136`); `:30-32` `OpenGuesses`, `GuessMinutesTotal`, `GuessMinutesLimit` (`GuessMinutesLimit = 300` at `:19`) |
| `needs_you_now` today | `internal/snapshot/estimate.go:430` sets it when a live pass is stopped; nothing else does |
| A pass in an undesigned round is reachable | `release add` of a gamma variant with `design_written: false` and a pass in that round printed "added release v1 of gamma" and exited 0 |
| The usage command | `spikes/7-usage/read.py` reads `Current week \(all models\): (\d+)% used · resets (.+)`. A live run on 2026-11-17 saved `$CLAUDE_JOB_DIR/tmp/usage-probe/raw.json` in 5.4 seconds, exit 0: `type` "result", `subtype` "success", `is_error` false, `num_turns` 0, `total_cost_usd` 0, every usage token 0. Its `result` text's second line reads "Current week (all models): 89% used · resets Nov 20, 10:59pm (America/Anchorage)", the next line is the Fable line at 83%, and a later "What's contributing" section has lines that start with a percent such as "96%" |
| The live line comes first | In the live result the all-models line is the first percent line, so item 7's recorded result is built with other percent lines ahead of it |
| The e2e test setup | `e2e/main_test.go`: `TestMain` builds the binary into a temp folder; `run(t, dir, now, args...)` at `:49` prepends `--state-dir` and `--now`, and its `exec.Command(binary, ...)` at `:52` is the only exec of the binary; `newState(t)` at `:68` copies the hard-coded `testdata/state` (`:71`, `:75`); `folder` and `sameFolder` compare a whole folder |
| The fixture's usage | `e2e/testdata/state/usage.json` holds 40 percent read at 22:30Z, 30 minutes before `workedClock` (`2026-11-20T23:00:00Z`), so every existing e2e verb reads usage once this pass lands. `settings.json` holds stop 90 |
| Existing starts stay legal | The 17 existing `"pass", "start"` calls start p13, b-three, p14 after p13, gamma's g-one (start, stop, resume), or are refusals. At most two passes build at once in any test, and every pass started is free under the new rules |
| Unit-test states | `internal/state/events_test.go:10` `loaded` writes one release and runs `Load`, so `s.Usage` is nil. That release, `goodRelease` (`internal/state/read_test.go:10-14`), has round `r1` with no `design_written`, so its pass breaks "unfinished work before it". `forecast_test.go`'s `TestPassEventMark` and `TestCommitForecast` use it too and write no `usage.json`. `internal/snapshot` tests build `&state.State{Releases: ...}` by hand (`nextsession_test.go:31` `sessionOf`, `estimate_test.go`), and `snapshot_test.go:20`, `:49` render `&state.State{}` against `testdata/empty.golden.json` |
| Label wording | The Survey prose standard (`/var/home/glw907/Projects/survey/skills/survey/writing/editor.md`, 99 lines) holds no rule on CLI labels, so the refusal names and the snapshot's reason text are the design's words, verbatim |

## Recorded searches

Each ran with `git grep -n -E '<pattern>' -- '*.go' e2e/testdata/README.md` at `480eecf`.

| Pattern | Hits | Each hit |
| --- | --- | --- |
| `exec\.Command\(binary` | 1 | `e2e/main_test.go:52` (task 4 adds `--usage-command` and the `PATH`) |
| `"pass", "start"` | 17 | `e2e/broken_test.go:19`; `e2e/estimate_test.go:162`, `:205`, `:254`, `:266`, `:292`, `:305`, `:312`, `:314`; `e2e/events_test.go:108`, `:110`, `:142`, `:159`, `:177`, `:206`, `:207`; `e2e/lock_test.go:28`. All stay legal; none changes. Each target is free under the new rules: p13 and g-one sit in designed rounds with nothing before them, b-three's questions are answered, and p14 starts after p13 finishes. No folder ever holds more than one building pass, and the worked fixture's `usage.json` reads 40, so task 3's rules pass these before task 4 takes readings |
| `settle\(` | 2 | `internal/state/commit.go:76` (call), `:113` (empty body). Task 3 fills it |
| `\bLimit\b\|\bGuess\b\|OpenGuess\|GuessMinutes` | 20 | `internal/snapshot/snapshot.go:18-19` (the constant moves to `state`, task 1), `:30-32`, `:134`, `:138-139`, `:180-181` (task 6); `internal/state/events.go:44`, `:116` (task 1, `Limit` becomes `Wait`); `internal/state/types.go:75-79`, `:89-90`, `:95-97` (task 1) |
| `Free\b\|WhyNot\|why_not` | 5 | `internal/snapshot/snapshot.go:86-87`, `:302` (task 6); `internal/snapshot/snapshot_test.go:128` (the done pass, whose `free` stays null), `:136` (no list encodes as null); both hold |
| `Usage\|usage\.json\|UsageStop\|usage_stop` | 30 | `cmd/cc-desk/root.go:82` (`SilenceUsage`, unrelated); `e2e/events_test.go:139` (`TestBadUsage`, command-line usage, unrelated); `internal/state/events_test.go:113`, `:116` (`StopUsage`, unchanged); `internal/state/read.go:173`, `:207-211` (unchanged), `:216-226` (task 2: the `error` form); `internal/state/read_test.go:32-36`, `:43`, `:52-53`, `:80` (unchanged), `:81` (task 2 adds cases); `internal/state/types.go:25`, `:33` (`StopUsage`, unchanged), `:163-178` (task 2), `:197` (unchanged) |
| `needs_you_now` | 6 | `e2e/estimate_test.go:307-308`, `internal/snapshot/nextsession_test.go:122-123` (unchanged); `internal/snapshot/estimate.go:430` (task 6 adds the usage reason); `internal/snapshot/snapshot.go:102` (comment, task 6) |
| `Pass 3\|pass 3` | 5 | `internal/snapshot/snapshot.go:130`, `:138`; `internal/state/commit.go:111`; `internal/state/types.go:75`, `:77`. Each comment is rewritten where its code changes |
| `Waiting\b\|"waiting"` | 9 | `internal/snapshot/estimate.go:266` (unchanged: a waiting pass counts as not started); `nextsession_test.go:108`, `snapshot.go:29`, `:179`, `snapshot_test.go:59` (task 6 fills the list; the empty case stays `[]`); `internal/state/check.go:12`, `:143` (task 1); `internal/state/events_test.go:43` (a hand-set `waiting` pass, task 1 gives it a `Wait`); `types.go:15` (unchanged) |
| `Refuse\(` | 11 | `cmd/cc-desk/main_test.go:32-33`, `internal/state/commit_test.go:98`, `refuse.go:12` (unchanged); `internal/state/events.go:54`, `:101`, `:190-194`, `:226` (unchanged), `:169` (task 3 keeps it for the states that can't start) |
| `testdata/state` | 5 | `e2e/events_test.go:74`, `:79` (the fixture golden, regenerated in task 6); `e2e/main_test.go:2`, `:71`, `:75` (task 5: `newState` takes the fixture name) |
| `StartPass\(` | 15 | `cmd/cc-desk/pass.go:21` (task 4 adds `--guess`); `internal/state/events.go:151` (task 3); `internal/state/events_test.go:75`, `:107`, `:119`, `:140`, `:158`, `:179`, `:188`, `:193`; `internal/state/forecast_test.go:20`, `:26`, `:27`, `:47`, `:62`. The signature stays; these tests need a readable usage reading in their state (task 3) |
| `newState\(` | 16 | `e2e/main_test.go:68` (definition, task 5) and 15 callers in `broken_test.go`, `estimate_test.go`, `events_test.go`, `lock_test.go` (unchanged if `newState` keeps its one-argument form; see task 5) |
| `func run\(` | 1 | `e2e/main_test.go:49` (task 4; task 5 adds the per-test variant) |

Golden files that change: `internal/snapshot/testdata/empty.golden.json` (task 6, the new
`reasons` list) and `e2e/testdata/state/snapshot.json` (task 6, the `free` and `why_not` values
and `reasons`). Each is regenerated by the command the task gives and read by eye before commit.

## File list

| File | Change | Task |
| --- | --- | --- |
| `internal/state/types.go` | `Pass.Limit` becomes `Pass.Wait` (`Since`, `Limits`); `Usage` gains `Error`; the 300-minute constant and the four refusal names move here | 1, 2 |
| `internal/state/check.go` | A waiting pass needs a `Wait`; a `Guess` needs a started pass and a question it rests on | 1 |
| `internal/state/rules.go`, `rules_test.go` | New: what is free, the names a start breaks, open guesses, building count, the usage verdict | 1 |
| `internal/state/events.go`, `events_test.go` | `StartPass` and the new `StartOnGuess` apply the rules; `:44` and `:116` check `Wait` | 1, 3 |
| `internal/state/read.go`, `read_test.go` | `usage.json`'s error form | 2 |
| `internal/state/commit.go`, `commit_test.go` | `settle`; `WithReading`; `usage.json` written after the event | 2, 3 |
| `internal/state/forecast_test.go` | The test state gets a readable reading | 3 |
| `internal/usage/usage.go`, `usage_test.go` | New: `Parse`, `Read`, and the cache decision | 4 |
| `cmd/cc-desk/root.go` | The hidden `--usage-command` flag; `commit` takes the reading before the lock | 4 |
| `cmd/cc-desk/pass.go` | `pass start --guess QUESTION`; the wait line | 4 |
| `cmd/cc-desk/usage.go`, `cmd/cc-desk/free.go` | New: the `usage` and `free` verbs | 4 |
| `internal/snapshot/snapshot.go`, `estimate.go`, `snapshot_test.go`, `nextsession_test.go`, `testdata/empty.golden.json` | `free`, `why_not`, `waiting`, `open_guesses`, `guess_minutes_total`, and `next_session.reasons` | 6 |
| `e2e/main_test.go` | `run` passes the default stand-in and runs the binary with a `PATH` that holds no `claude` (task 4); the per-test stand-in helper and a second fixture (task 5) | 4, 5 |
| `e2e/testdata/usage40.sh` | New, executable: the default stand-in, a 40 percent result in the live shape | 4 |
| `e2e/broken_test.go` | The `verbs` list gains `usage` and `free` (its comment at `:10-11` asks each pass to add its verbs) | 4 |
| `e2e/testdata/order/releases/north.json`, `south.json`, `e2e/testdata/README.md` | New fixture; the README names it | 5 |
| `e2e/testdata/state/snapshot.json` | Regenerated golden | 6 |
| `e2e/order_test.go` | New: done items 2 and 3, a guess start that waits, and the `free` verb | 5, 6, 7 |
| `e2e/usage_test.go` | New: done item 7's recorded result and the cache | 7 |
| `docs/STATUS.md`, `docs/follow-ups/*.md` | The pass's STATUS lines and follow-ups | 8 |

## The rules this pass builds

These are plain functions in `internal/state/rules.go` over the loaded `State` and the clock.
The start events, `settle`, and the snapshot's `free` and `why_not` all call them, so a rule has
one home.

**Before it.** The passes before a pass are found transitively, as pass 2's estimate finds them:
each pass in its `after`, and every pass of every round in its round's `after`, followed through
their own `after` lists. An `after` round that has no passes, or whose passes are not all done,
is unfinished work.

**What a start breaks.** For a pass P in a live release, the names it breaks are, in this order:

| Name, verbatim | When P breaks it |
| --- | --- |
| "unfinished work before it" | P's round's design isn't written, or a pass or round before P isn't done |
| "an unanswered question" | P rests on an unanswered question other than the one a guess start names. A plain start names every unanswered question P rests on |
| "the one-pass rule" | A pass before P carries an open guess, whatever that pass's state |
| "the 300-minute rule" | Guess starts only: the open-guess minutes plus P's planned minutes are over 300 |

A guess is open when the pass carries a `Guess` and its question has no answer. The open-guess
minutes are the planned minutes of every pass in every live release that carries an open guess,
whether it is building, stopped, waiting, or done. A waiting guess start counts from the moment
it is accepted.

**Free.** A pass is free when it breaks none of the first three names. The snapshot sets `free`
for `not_started` and `waiting` passes and leaves it null for the rest; `why_not` holds the
names a plain start breaks, `[]` when free.

**The limits.** The two-pass limit holds when two or more passes in live releases are
`building`. A `stopped` pass doesn't count. The usage limit holds when the reading is missing,
carries an error, or its percent is at or over the stop (`settings.json`, 100 when absent). A
nil `s.Usage` counts as unreadable here: the command always takes a reading before an event, so
nil reaches the rules only from a unit test's hand-built state.

**A start request.** In order:

1. P must exist (exit 1 with the name, as today). A guess start's question must exist in P's
   release (exit 1 with its name). It must be one P rests on and still unanswered; if not, the
   request is refused (exit 2) with a line that names the question and no rule name.
2. A `stopped` P resumes as today, outside both limits ("Passes already building carry on").
   Any state but `not_started` and `stopped` is refused as today at `events.go:169`. A plain
   start of a P that already carries a `Guess` resumes or is refused the same way.
3. Collect the names P breaks. Any name means the request is refused (exit 2), nothing is
   written, and the refusal is one line: `cannot start <project>/<pass>: <name> (<detail>); <name> (<detail>)`.
   The detail is plain words a person reads: the unanswered question ids, the unfinished pass or
   round names, the guessed pass, or the minutes against 300. Detail never repeats one of the four names.
4. With no name broken, a guess start records `Guess{Question, Recommendation, At: now}` where
   the recommendation is the question's `recommended` option. Both kinds of request then check
   the limits: none holds, and P is `building` with `started` now and a pass event set; any
   holds, and P is `waiting` with `Wait{Since: now, Limits: [...]}`, no pass event. The limit
   list holds `two_pass_limit`, then `usage`, for those that hold.

**Settling.** `settle(s, now)` takes the `waiting` passes in live releases oldest `Wait.Since`
first; ties go to release order (by project name), then pass order in its file. For each, it
recomputes the limits with the passes it has already started counted. None holds, and the pass
starts: `building`, `started` now, `Wait` cleared, a pass event set, the release dirty. Otherwise
its limit list is updated in place, and the release is dirty only if the list changed. Settling
never re-checks the four names: a pass that passed them when it was accepted keeps its place.

**The reading.** Before the lock, the command reads `usage.json` leniently. It takes a fresh
reading when that file is absent or doesn't parse, when its `read_at` is 5 minutes or more
before the clock, when `read_at` is after the clock, or when the verb forces one (`usage
--raw`). A fresh reading is `{percent, reset_text, read_at: now}`, or `{error, read_at: now}`
when the command exits non-zero, times out after 60 seconds, prints JSON it can't decode,
reports `is_error`, or has no "Current week (all models)" line. Inside the lock,
`state.WithReading(u, forced, ev)` keeps one reading and marks `usage.json` dirty when the fresh
one wins. A forced reading always wins. Otherwise the fresh one wins when the loaded `read_at`
is older, or after the clock (a reading from the future counts as older than any). On a tie the
loaded one stays. It then settles and runs `ev`. So a waiting pass that now
fits starts before a new request can take its place. `Commit` writes `usage.json` atomically
after the releases and before the forecast and `snapshot.json`, and only once the event has
succeeded: a refusal writes nothing. A broken `usage.json` still fails the strict `Load`, as any
state file does.

**The snapshot.** `waiting` lists every waiting pass, oldest first by the settle order: project,
pass, `limits`, `since`, and `reset_text`, which holds the reading's reset text when `usage` is
among the limits and the reading is readable, and null otherwise. `open_guesses` lists each pass
with an open guess (project, pass, question, planned minutes), and `guess_minutes_total` is
their sum. `next_session` gains `reasons`, `[]` when empty; a reading that carries an error adds
"the usage figure can't be read" and makes the state `needs_you_now`, ahead of every other
state. A nil `s.Usage` adds no reason (see `## Decisions`).

**Forecast lines.** One line per commit that has any pass event, as pass 2 built it. A finish
whose settle starts a waiting pass writes one line. A recorded wait writes none.

## Acceptance

Every check runs the built `cc-desk` binary against a copy of a fixture state folder in
`t.TempDir()`, through `go test ./e2e/...`, with the usage read replaced by a stand-in script.
The new fixture `e2e/testdata/order/` (task 5) holds two projects, north and south, with no
`usage.json` and no `settings.json`, so the stop is 100.

| Item, word for word | Task | Check |
| --- | --- | --- |
| **Waiting order.** Five cases run with two projects, as the main design's done item 5 lists them. The set of names in each refusal is exactly the set given here. - A free pass starts. A plain start of the pass resting on a question is refused with "an unanswered question". - With only a 400-minute pass resting on a question, a guess start is refused with "the 300-minute rule", nothing starts, and the snapshot reads "Ready now". - A 200-minute pass starts on a guess, and the file records the question, the recommendation, and the time. - A 150-minute guess start in the other project is refused with "the 300-minute rule". - The guessed pass finishes with its question still open. A 50-minute pass after it is then refused with "the one-pass rule" alone. After each refusal the state folder is byte for byte what it was. | 1, 3, 4, 5 | `go test ./e2e/ -run TestWaitingOrder -v` prints `--- PASS` for `TestWaitingOrder` and each of its five `t.Run` cases |
| **Limits.** With one pass building in each project, two more free passes are requested and both are recorded as waiting on the two-pass limit. Finishing one building pass starts the earlier request and leaves the later one waiting. With a recorded usage reading of 99 a free pass starts. With 100 it waits and the snapshot shows the reset text, and a later reading of 99 followed by any verb starts it. With `usage_stop_percent` set to 85 and a reading of 87, it waits. With a reading the command can't parse, it waits and the snapshot reads "Needs you now". | 2, 3, 4, 6, 7 | `go test ./e2e/ -run TestLimits -v` prints `--- PASS` for `TestLimits` and each of its six `t.Run` cases |
| **The usage reader.** Given a recorded result whose first percent is a different line, the reader returns the figure on the "Current week (all models)" line. Run live on the workstation, the usage verb prints the same percent as that line of the raw result it saved, and the result reports a cost of zero and no model turn. The live run is outside CI, and the round's report gives its output. | 4, 7, 8 | `go test ./internal/usage/ -run TestParse -v` and `go test ./e2e/ -run TestUsageReader -v` print `--- PASS`. The live run is task 8's three commands, run once on the workstation; the PR body quotes their output |
| **The gate.** `bash scripts/gate.sh` passes, and so does CI on the pull request. The gate fails if any test was skipped, so a runner without a terminal can't pass item 4 by skipping it. | 8 | `bash scripts/gate.sh` run bare on the final head prints `gate: ok`; the PR's `gate` check on that head is `pass`. The skip check landed in pass 1 (`e2e/gate_test.go`) and stays |

What each check printed on today's code, before any change. A binary built into
`$CLAUDE_JOB_DIR/tmp/pre` ran on a copy of `e2e/testdata/state` at `workedClock`:

```text
$ cc-desk pass start alpha p15                       # p15 rests on alpha-q1, unanswered
started alpha/p15                                    # exit 0: no rule is checked
$ cc-desk pass start alpha p15 --guess alpha-q1
unknown flag: --guess                                # exit 1
$ cc-desk release add gamma.json; pass start alpha p13; pass start beta b-three; pass start gamma g-one
                                                     # all exit 0: three passes building
snapshot: "waiting": [], and every pass "free": null, "why_not": []
$ (usage.json set to 100 percent) cc-desk pass start alpha p13
started alpha/p13                                    # exit 0: the stop isn't read
$ cc-desk usage
unknown command "usage" for "cc-desk"                # exit 1
```

`TestWaitingOrder`, `TestLimits`, `TestParse`, and `TestUsageReader` don't exist yet, so each
`go test -run` check reports no tests to run, which fails its `--- PASS` requirement. The gate
printed `gate: ok`.

## Tasks

Every check below that runs `go test -run` names the new tests, so today's tree fails it. Each
task's commit runs `bash scripts/gate.sh` bare first, so the gate stays green at every commit.

To regenerate the fixture golden, run these from the worktree, with a new folder name each time
(`fx1`, `fx2`). The `--usage-command` flag and `e2e/testdata/usage40.sh` exist from task 4:

```bash
go build -o "$CLAUDE_JOB_DIR/tmp/cc-desk" ./cmd/cc-desk
cp -r e2e/testdata/state "$CLAUDE_JOB_DIR/tmp/fx1"
"$CLAUDE_JOB_DIR/tmp/cc-desk" --state-dir "$CLAUDE_JOB_DIR/tmp/fx1" --now 2026-11-20T23:00:00Z --usage-command "$PWD/e2e/testdata/usage40.sh" snapshot > e2e/testdata/state/snapshot.json
```

Then read `git diff e2e/testdata/state/snapshot.json` by eye.

### Task 1: the types, the checks, and the rules (25 minutes)

Outcome: `internal/state/rules.go` answers, for any pass, which of the four names a plain or a
guess start breaks, whether the pass is free, the open-guess minutes, the building count, and
whether the usage limit holds, all by "The rules this pass builds". `Pass.Limit` is `Pass.Wait`
(JSON `wait`: `since`, `limits`). The four names and the 300-minute cap are exported constants
in `state`; `snapshot.GuessMinutesLimit` refers to the cap. `validatePass` requires a `Wait`
with at least one known limit on a `waiting` pass and none on any other; a `Guess` on any
state but `not_started` (a guess start that waits carries one), naming a question the pass rests
on. `events.go:44` and `:116` reject a
new pass carrying `Wait` or `Guess`. Files: `internal/state/types.go`, `check.go`, `rules.go`,
`rules_test.go`, `events.go`, `events_test.go` (`:43`'s hand-set waiting pass gains a `Wait`),
`internal/snapshot/snapshot.go` (the constant). Class `ordinary`.

Test, as `TestRules<Case>` in `rules_test.go`, on states the test builds:

- `TestRulesBefore`: unfinished work through an `after` pass, an `after` round with no passes,
  an `after` round not done, and an undesigned round.
- `TestRulesQuestions`: a plain start names every unanswered question; a guess start leaves its
  own question out and names any other.
- `TestRulesOnePass`: a done pass before P with an open guess breaks the one-pass rule; an
  answered guess doesn't.
- `TestRulesGuessMinutes`: 200 open plus 100 is allowed, 200 plus 150 is not, a waiting guess
  start counts, and a plain start never breaks the 300-minute rule.
- `TestRulesLimits`: two building passes hold the two-pass limit and a stopped one doesn't; a
  nil, an error, a 99, and a 100 reading under the default stop, and 87 under a stop of 85.
- `TestCheckWait`: the strict checks above, one table row each, including a `waiting` pass that
  carries a `Guess` (it loads) and a `not_started` pass that carries one (it doesn't).

Acceptance check: `go test ./internal/state/ -run 'TestRules|TestCheckWait' -v` prints `--- PASS`
for each of the six, and `go test ./internal/...` passes.

### Task 2: the reading in the state files (15 minutes)

Outcome: `Usage` gains `Error string` (`error`, omitted when empty). The strict reader takes a
`usage.json` with `read_at` and either an error with no percent and no reset text, or a reset
text and no error; anything else names the file. `state.WithReading(u Usage, forced bool, ev Event) Event`
works as "The reading" says, and `Commit` writes `usage.json` when it is dirty, in the order
releases, `usage.json`, forecast, snapshot. Its signature stays. Files: `internal/state/types.go`,
`read.go`, `read_test.go`, `commit.go`, `commit_test.go`. Class `ordinary`.

Test: `read_test.go`'s table gains "usage error with percent" and "usage without reset text or
error", each refused with the file. `TestCommitReading` in `commit_test.go`: a fresh reading is
written; an older fresh reading than the loaded one leaves `usage.json` alone; a refused event
leaves the whole folder as it was, `usage.json` included.

Acceptance check: `go test ./internal/state/ -run 'TestRead|TestCommitReading' -v` prints `--- PASS`
for `TestCommitReading` and the `TestRead` cases.

### Task 3: the start events and settling (30 minutes)

Outcome: `StartPass(project, pass)` and the new `StartOnGuess(project, pass, question)` follow
"A start request", sharing one start path; `settle` follows "Settling". Files:
`internal/state/events.go`, `events_test.go`, `commit.go`, `commit_test.go`, `forecast_test.go`.
Class `ordinary`.

The unit tests of `StartPass` (15 hits) now fail for two reasons. Their states have a nil
`Usage`, which counts as a stop, and their release, `goodRelease`, has a round with no written
design, which breaks "unfinished work before it". Give the start tests a release with
`"design_written": true` (a start-test copy of `goodRelease`, or the constant itself if no
read test depends on its absence) and a readable reading: the `loaded` helper writes a fresh
`usage.json`, and so do `forecast_test.go`'s `TestPassEventMark` and `TestCommitForecast`. Leave
the rules as they are (standing answer 28).

Test, in `events_test.go`:

- `TestStartOutcomes`: started, waiting on the two-pass limit, waiting on usage, waiting on both
  (one `Wait` listing both), and refused with the exact name set and nothing changed. A waiting
  start sets no pass event.
- `TestStartOnGuess`: records the question, the recommended option, and the clock; a question
  that doesn't exist exits with its name (not a refusal); a question the pass doesn't rest on,
  or one already answered, is a refusal with no rule name; a guess start that waits records the
  guess and counts toward 300.
- `TestStartWaitingPass`: a plain start of a `waiting` pass is refused as at `events.go:169`.
- `TestSettle`: oldest `since` first, ties by project then pass order; one free slot starts the
  oldest and updates the other's limits; a pass waiting on usage alone starts when the reading
  drops; a settle start sets a pass event; a resumed stopped pass ignores both limits.

Acceptance check: `go test ./internal/state/ -run 'TestStartOutcomes|TestStartOnGuess|TestStartWaitingPass|TestSettle' -v`
prints `--- PASS` for each of the four, and `go test ./internal/...` passes.

### Task 4: the usage reader and the verbs (30 minutes)

Outcome: `internal/usage` holds `Parse(raw []byte) (percent int, resetText string, err error)`,
which decodes the result loosely and reads the first "Current week (all models)" line with the
spike's pattern, and `Read(ctx, command string)`, which runs the command with no arguments, or,
when `command` is empty, `claude -p "/usage" --model haiku --output-format json` with stdin from
the null device and a 60-second timeout. It sets `cmd.WaitDelay` (1 second, well under any stand-in's sleep) beside the
context's timeout: a probe showed a killed shell's child holds stdout open, so without it a
500 ms timeout on a script that sleeps 4 seconds returned after 4 seconds. The cache decision ("The reading") also lives here, so
`cmd` stays wiring. In `cmd/cc-desk`:

- A hidden persistent `--usage-command PATH` flag beside `--now`.
- `rootFlags.commit` takes the reading before the lock, wraps the event in
  `state.WithReading`, for every verb, and returns the reading it kept beside the snapshot's
  bytes.
- `pass start <project> <pass> [--guess QUESTION]` prints `started <project>/<pass>` as today,
  or `waiting <project>/<pass> on <limits>` on the same stream, exit 0. It tells them apart by
  P's state and `waiting` entry in the snapshot `commit` returns.
- `usage [--raw FILE]` forces a fresh reading when `--raw` is given and saves the raw bytes to
  FILE, commits through `commit(nil)`, and prints `<n>% used, resets <reset text>` on stdout
  from the reading `commit` returns.
  An unreadable reading is still committed, prints its reason on stderr, and exits 1.
- `free` runs through `commit(nil)`, decodes the snapshot it returns, and prints one line per
  live pass that is `not_started` or `waiting`: `free <project>/<pass>`, `waiting
  <project>/<pass> on <limits>`, or `not free <project>/<pass>: <names>`.

The same task gives the e2e suite its default stand-in, because from this commit every verb
takes a reading and the fixture's `usage.json` is 30 minutes old at the worked clock. Without
it, `go test ./e2e/...` would run the real `claude` (standing answer 32). Commit the stand-in as
`e2e/testdata/usage40.sh`, executable, printing a 40 percent result in the live shape. `run`
always passes `--usage-command` with its absolute path, and sets `cmd.Env` to the test's
environment with `PATH` replaced by `/usr/bin:/bin`. The binary needs no other tool, so a run
that misses the flag finds no `claude`, reads an error, and fails its test where you see it.
`e2e/broken_test.go`'s `verbs` list gains `{"usage"}` and `{"free"}`. Each takes its reading
before the lock, then the strict `Load` fails, so nothing is written and the existing checks
hold.

Files: `internal/usage/usage.go`, `usage_test.go`, `cmd/cc-desk/root.go`, `pass.go`, `usage.go`,
`free.go`, `e2e/main_test.go`, `e2e/testdata/usage40.sh`, `e2e/broken_test.go`. Class `ordinary`.

Test, in `usage_test.go`: `TestParse` with a result trimmed from the live shape (the all-models
line first), a built result whose earlier lines carry the Fable percent and a "96%" contribution
line before the all-models line, and one case for each unreadable kind: bad JSON, `is_error`
true, and no matching line. `TestRead` runs a stand-in script from `t.TempDir()`: it returns its
output, reports a non-zero exit as an error, and times out: with a 500 ms timeout on a stand-in
that sleeps 4 seconds, it returns an error in under 4 seconds. `TestFresh` covers the cache: absent, broken, aged 4 and 5 minutes, and in the future.

Acceptance check: `go test ./internal/usage/ -run 'TestParse|TestRead|TestFresh' -v` prints
`--- PASS` for each of the three, and the e2e suite passes with no `claude` on the `PATH`. `claude`
shares Homebrew's bin folder with the tools the suite needs (its gate test runs `gofmt`, `jq`,
and `vale`), so make a folder that links only those four, then run the suite with that folder
and `/usr/bin` as the whole `PATH`:

```bash
mkdir "$CLAUDE_JOB_DIR/tmp/goonly"
for b in go gofmt jq vale; do ln -s "$(command -v "$b")" "$CLAUDE_JOB_DIR/tmp/goonly/$b"; done
env PATH="$CLAUDE_JOB_DIR/tmp/goonly:/usr/bin" go test -count=1 ./e2e/...
```

On today's code that run printed `ok  github.com/glw907/cc-desk/e2e 11.436s`.

### Task 5: the stand-in, the fixture, and done item 2 (25 minutes)

Outcome: each test can set its own reading, and `TestWaitingOrder` passes. Files:
`e2e/main_test.go`, `e2e/testdata/order/releases/north.json`, `south.json`,
`e2e/testdata/README.md`, `e2e/order_test.go`. Class `ordinary`.

1. **The per-test stand-in.** Task 4's default stays. A helper writes a per-test stand-in into `t.TempDir()` with a settable delay (a `sleep`
   before it prints, which pass 4 sets to 15 seconds), a result file the test can rewrite
   between verbs, and a log file that gains one line per call. A variant of `run` takes that
   stand-in's path.
2. **The fixture.** `newState` keeps its one-argument form for `testdata/state`, and a sibling
   takes a fixture name, so the 15 existing callers don't change. `e2e/testdata/order/` holds
   two releases, each one round, designed, `needs_geoff: true`. Every question is unanswered,
   with options `yes` and `no` and `recommended` `yes`.

   | Release | Round | Pass | Planned | After | Rests on |
   | --- | --- | --- | --- | --- | --- |
   | north | n1 | n-free | 60 | | |
   | north | n1 | n-two | 30 | | |
   | north | n1 | n-big | 400 | | north-q1 |
   | north | n1 | n-guess | 200 | | north-q2 |
   | north | n1 | n-after | 50 | n-guess | |
   | south | s1 | s-free | 45 | | |
   | south | s1 | s-two | 30 | | |
   | south | s1 | s-guess | 150 | | south-q1 |

   The README gains a section for it.
3. **`TestWaitingOrder`**, in `e2e/order_test.go`, five `t.Run` cases. Each refusal check reads
   the exit code (2), collects which of the four names appear in stderr, and compares that set
   to the expected set exactly, then compares the whole folder with `sameFolder` against a copy
   taken just before.

   | Case | Copy | Steps, each at a later clock | Expected |
   | --- | --- | --- | --- |
   | `free_and_question` | A | `pass start north n-free`; then `pass start north n-guess` | Exit 0, n-free `building`; then exit 2, set {"an unanswered question"}, folder unchanged |
   | `too_big` | B, fresh | `pass start north n-big --guess north-q1`; then `snapshot` | Exit 2, set {"the 300-minute rule"}, folder unchanged; no pass `building`; `next_session.state` `ready_now` |
   | `guess_starts` | A | `pass start north n-guess --guess north-q2` | Exit 0, n-guess `building`; `north.json` holds `guess` with `question` `north-q2`, `recommendation` `yes`, `at` the clock (task 6 adds the `open_guesses` check) |
   | `other_project` | A | `pass start south s-guess --guess south-q1` | Exit 2, set {"the 300-minute rule"}, folder unchanged |
   | `one_pass` | A | `pass finish north n-guess`; then `pass start north n-after` | Exit 0; then exit 2, set {"the one-pass rule"}, folder unchanged |

   Copy A's cases run in order on one folder.

Acceptance check: `go test ./e2e/ -run TestWaitingOrder -v` prints `--- PASS` for
`TestWaitingOrder` and its five cases, and `go test ./e2e/...` passes.

### Task 6: the snapshot fields (15 minutes)

Outcome: the snapshot carries `free`, `why_not`, `waiting`, `open_guesses`,
`guess_minutes_total`, and `next_session.reasons` by "The snapshot". `WaitingPass` becomes
`Project`, `Pass`, `Limits []string`, `Since`, `ResetText *string`. Files:
`internal/snapshot/snapshot.go`, `estimate.go`, `snapshot_test.go`, `nextsession_test.go`,
`testdata/empty.golden.json`, `e2e/testdata/state/snapshot.json`, `e2e/order_test.go`. Class `ordinary`.

Test: `TestSnapshotFree` builds a state with a free pass, a pass behind unfinished work and a
question (two names, in table order), and a waiting pass with both limits and a readable
reading (its reset text shows); `TestSnapshotUsageUnreadable` gives `needs_you_now` with the
one reason and a null reset text; `TestSnapshotOpenGuesses` sums a building and a waiting
guess. `snapshot_test.go:128` checks the done pass, whose `free` stays null, and `:136` checks
only that no list encodes as null, so both hold. Its not-started pass `b` sits in the undesigned
round `ask` and now reads `free: false` with `["unfinished work before it"]`. Task 6 also adds
to `TestWaitingOrder`'s `guess_starts` case: `open_guesses` holds one entry and
`guess_minutes_total` reads 200.
Regenerate `empty.golden.json` with `go test ./internal/snapshot/ -update` and the fixture golden
as above, and read both diffs: p13 and b-three read `free: true`; p14 reads `why_not`
["unfinished work before it"]; p15 reads both names.

Acceptance check: `go test ./internal/snapshot/ -run 'TestSnapshot' -v` prints `--- PASS` for
the three new tests and the existing ones, and `go test ./e2e/ -run 'TestFixtureSnapshotIsCurrent|TestWaitingOrder' -v`
prints `--- PASS` for both.

### Task 7: done items 3 and 7 end to end (25 minutes)

Outcome: `TestLimits`, `TestGuessWaits`, `TestFreeVerb`, and `TestUsageReader` pass. Every reading goes through a stand-in; no test
writes `usage.json` by hand. Files: `e2e/order_test.go`, `e2e/usage_test.go`. Class `ordinary`.

`TestLimits`, six `t.Run` cases, each on a fresh copy of the order fixture:

| Case | Steps | Expected |
| --- | --- | --- |
| `two_pass_limit` | Start n-free and s-free; request n-two at a later clock, then s-two later still; then `pass finish north n-free` | Both requests exit 0 and print the wait line; both are `waiting` with `limits` `["two_pass_limit"]`. After the finish n-two is `building` and s-two still waits, listed in `waiting` |
| `usage_99` | Stand-in reads 99; start n-free | `building` |
| `usage_100_then_99` | Stand-in reads 100; start n-free; `snapshot`; at least 5 minutes later the stand-in reads 99; `snapshot` | `waiting` with `["usage"]`, its `reset_text` equal to the stand-in's text; after the later snapshot, `building` |
| `stop_85` | A `settings.json` with `usage_stop_percent` 85; stand-in reads 87; start n-free | `waiting` with `["usage"]` |
| `unparseable` | Stand-in prints text that isn't a result; start n-free; `snapshot` | `waiting` with `["usage"]`, `reset_text` null; `next_session.state` `needs_you_now`, `reasons` `["the usage figure can't be read"]` |
| `cache` | A logging stand-in; `snapshot` twice, 1 minute apart; `snapshot` again 5 minutes after the first | The log holds one call after the first two verbs and two after the third |

`TestGuessWaits`: start n-free and s-free, then `pass start north n-guess --guess north-q2`. It
exits 0 and waits on `["two_pass_limit"]` with its `guess` recorded. A following `snapshot`
exits 0 (the release still loads), `open_guesses` holds n-guess, and `guess_minutes_total`
reads 200.

`TestFreeVerb`: start n-free and s-free, request n-two, then run `free`. Its lines include
`free south/s-two`, `waiting north/n-two on two_pass_limit`,
and `not free north/n-big: an unanswered question`.

`TestUsageReader`: a stand-in prints the built result from task 4 with the all-models line at 61
percent after a Fable line and a "96%" line. `cc-desk usage` prints `61% used, resets <its reset
text>`.

Acceptance check: `go test ./e2e/ -run 'TestLimits|TestGuessWaits|TestFreeVerb|TestUsageReader' -v`
prints `--- PASS` for the four tests and each of `TestLimits`'s six cases. Then rerun task 4's
suite command with no `claude` on the `PATH`.

### Task 8: the live read, gate, pull request, and close (10 minutes)

Outcome: the live check of done item 7 has run once and its output is in the PR body; the gate
and CI pass on the final head; the PR is ready; STATUS carries the pass's lines.

1. The live read, outside CI, from the worktree, each command alone:

   ```bash
   go build -o "$CLAUDE_JOB_DIR/tmp/live/cc-desk" ./cmd/cc-desk
   "$CLAUDE_JOB_DIR/tmp/live/cc-desk" --state-dir "$CLAUDE_JOB_DIR/tmp/live/state" usage --raw "$CLAUDE_JOB_DIR/tmp/live/raw.json"
   python3 -I -c 'import json,re,sys; r=json.load(open(sys.argv[1])); m=re.search(r"Current week \(all models\): (\d+)% used", r["result"]); print("line percent:", m.group(1)); print("total_cost_usd:", r["total_cost_usd"]); print("num_turns:", r["num_turns"])' "$CLAUDE_JOB_DIR/tmp/live/raw.json"
   ```

   Use a fresh `live` folder; never point `--state-dir` at a real state folder. The PR body
   quotes the verb's line and the three printed lines, and nothing else from the raw result.
   The percent the verb prints must equal the line percent, the cost must be 0, and the turns 0.
   A difference goes by standing answer 30.
2. Open the PR as `round-run.md` "Evidence" says, and run the whole-branch
   `standards:diff-reviewer` read the class calls for. Fold its findings once. A delta review
   after the fold has its Low findings folded once and checked by your own diff read and CI, or
   filed as follow-ups; it never starts another round.
3. After the last edit (STATUS and the review fold included), run `bash scripts/gate.sh` bare and
   never piped, push, and watch the `gate` check with `gh pr checks <n> --watch`. Name the CI
   verdict on that head in the PR body. Merge nothing.

Acceptance check: `bash scripts/gate.sh` prints `gate: ok` on the final head; the PR's `gate`
check on that head is `pass`; the PR body holds the live read's four lines.

The tasks sum to 175 minutes: 25, 15, 30, 30, 25, 15, 25, and 10.

## What pass 4 adds

Pass 4 records rulings. A ruling that answers a guessed question closes the guess, so its
minutes leave the 300 total and the one-pass rule lifts; `rules.go` reads "open" from the
question's answer, so pass 4 needs no change there. Its done item 6 sets the stand-in's delay to
15 seconds, which task 5's helper already takes. A ruling verb is a writing verb, so it settles
through the same `WithReading` path.

## Left alone

`docs/follow-ups/orphan-question-forecast.md` stays open; this pass doesn't build on it. A
`stopped` pass that resumes doesn't check the limits, so a resume can put three passes building;
task 8 files that as a follow-up for Geoff (see `## Decisions`).

## Standing answers

The twelve in `round-run.md` hold, with the round plan's changes and its answers 13 to 17,
pass 1's 18 to 22 (in `docs/plans/2026-11-17-state-rules-files-plan.md`), and pass 2's 23 to 26
(in `docs/plans/2026-11-17-state-rules-estimate-plan.md`). This pass adds:

- **27. A start case the plan doesn't cover.** Take the reading closest to the design's text,
  write a unit test for it in `rules_test.go` or `events_test.go`, and list it under
  `## Decisions`.
- **28. A rule breaks a hand-built test state.** Fix the test's state (give it a readable
  reading, a written design, a `Wait`, or a started time), never the rule. List the tests changed in the PR body.
- **29. A signature change breaks a caller this plan didn't list.** Fix the caller in this pass
  and add it to the PR body's file list.
- **30. The live read disagrees with the reader, costs anything, or takes a model turn.** A cost
  above zero is spend: stop, commit and push what you have, write the stop with the four printed
  lines, and end. A percent that differs with a traced cause (the line's wording changed) gets a
  fix to `Parse` and its test, one more live run, and a `## Decisions` line. A difference with no
  traced cause, or a model turn at zero cost, goes under `## Calls for Geoff` with the four
  lines, and the close goes on. Never retry in a loop: one more run is allowed if the first
  timed out.
- **31. The path runs past 180 minutes.** This is round standing answer 14 applied to this
  pass's 90 planned minutes. Finish the task in flight, push, and write the stop.
- **32. A test would run the real `claude`.** It is a bug: every e2e run passes
  `--usage-command`, and every unit test of `Read` passes a stand-in. Fix the test. No test
  skips when `claude` is absent.
- **33. The four names or the reason text.** They are the design's words, exact. Never reword
  them, and never let a detail repeat one.

## Decisions

- [advisor] A resumed `stopped` pass skips both limits, by "Passes already building carry on",
  and only `building` passes count toward two. A resume can then put three passes building.
  Call: the PR's one call for Geoff, with a follow-up file. Recommended: keep it, since a
  stopped pass holds its worktree and its session and stopping it twice costs more than one
  extra session.
- [advisor] A pass in an undesigned round breaks "unfinished work before it". The design's
  free list puts "its round's design is written" first, and the four names have no other home
  for it. The probe shows such a release installs.
- [advisor] `next_session` gains `reasons`, `[]` when empty, for "the usage figure can't be
  read". Both golden files change. A stopped pass adds no reason text; the design names none, and
  each pass already shows its stop reason.
- [advisor] `Pass.Limit` becomes `Pass.Wait` with a list of limits and one request time,
  because "a wait records every limit that holds the pass" and settling takes the oldest
  request first. `Since` is the request time and settle never resets it.
- [advisor] A second fixture, `e2e/testdata/order/`, carries done items 2 and 3, so the worked
  fixture and its golden stay as pass 2 left them apart from the new fields.
- [advisor] The usage read sits behind a hidden `--usage-command` flag, beside `--now`, so the
  tests set the read the way they set the clock. The design leaves "how fixtures replace the
  clock and the usage read" to the builder.
- [advisor] The reading is taken before the lock and folded in inside it, keeping the newer
  reading, then settled before the event. A refusal writes nothing, `usage.json` included,
  by "A start breaks a rule: Refused ... Nothing is written".
- A nil `s.Usage` is a stop in the rules and adds no reason in the snapshot. The command
  always takes a reading, so nil comes only from a unit test's state, and treating it as a
  stop keeps a missing reading from ever starting a pass. Not adding the reason keeps every
  hand-built snapshot test and the empty golden as they are.
- A broken `usage.json` fails the strict `Load` like any state file, and the command exits with
  its name. The pre-lock read is lenient only to decide whether to read again.
- The 300-minute rule applies to guess starts. The design's text sets the cap on "this pass"
  joining the open guesses, and a plain start adds no guess.
- Settling doesn't re-check the four names. A waiting pass passed them when it was accepted,
  and in this round nothing between acceptance and settling can make one true: a ruling only
  answers questions, and an open guess before a waiting pass would have refused it.
- Ties in settling go to project name, then pass order. Item 3's requests use distinct clocks,
  so the test doesn't rest on the tie.
- A wait prints one line and exits 0: a wait is an outcome the design lists beside "Started",
  not a refusal.
- One forecast line per commit with a pass event, as pass 2 built it. A finish whose settle
  starts a pass writes one line. Pass 4's item 6 counts lines per commit with a pass event.
- The labels are the design's words. The Survey prose standard
  (`/var/home/glw907/Projects/survey/skills/survey/writing/editor.md`) holds no rule for CLI
  labels.
- [review] One `standards:plan-reviewer` read at `439de32` gave "ready after folds". Folded: the
  `open_guesses` check moved from task 5 to task 6; a waiting guess start may carry a `Guess`,
  with `TestCheckWait` rows and `TestGuessWaits`; task 3 names `goodRelease`'s undesigned round
  beside the nil reading; `e2e/broken_test.go` joins the file list; `Read` sets `WaitDelay`,
  confirmed by a probe; `TestFreeVerb` covers the `free` verb; "The reading" states which
  reading wins; `commit` returns its reading; the `cache` case names its verbs; the default
  stand-in is committed as `e2e/testdata/usage40.sh`; standing answer 30 now stops only on
  spend, as `planner.md` "Stops" says; the `snapshot_test.go` sentence is corrected; `run`
  hides `claude` from the binary's `PATH`; and the executor prompt opens the draft PR before
  task 1's push, because CI runs only on pull requests. Task 7 grew to 25 minutes.
- The executor runs on Sonnet at medium effort, as the round's launch line names it. Each
  task's rules are spelled out here, and the pass carries no novel design call.

## Executor decisions

## Rulings

Carried from pass 2: Geoff, 2026-11-18, terminal session, via conductor-cc-desk, on the round
that needs Geoff with no design and no item left (`ready_now` with a null time, or
`nothing_waiting`): "Those both work." The built `ready_now` stands, and this pass's new
`needs_you_now` reason sits ahead of it.
