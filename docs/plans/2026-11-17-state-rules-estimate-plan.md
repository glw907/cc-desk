# The estimate: pass plan

Initiative: cc-desk
Round: state-rules
Pass: state-rules-estimate

Class `ordinary`. Planned 75 minutes; the path below is 115, under the round's 150-minute park
line. Design: `docs/specs/2026-11-17-state-and-rules-design.md` (approved 2026-11-17), section
"The estimate", with the snapshot fields "The snapshot" lists for it. Round plan:
`docs/plans/2026-11-17-state-rules-round.md`. Where the round design is silent,
`docs/specs/2026-11-17-cc-desk-design.md` holds. Writing a person reads (the PR body, STATUS
lines, follow-ups) follows `~/.claude/docs/voice/written-for-people.md`.

Worktree `/var/home/glw907/Projects/cc-desk/.claude/worktrees/state-rules-estimate`, branch
`state-rules-estimate`, cut from `origin/main` at `f81af60` (pass 1 landed as PR #3). PR base:
`main`. Push: `glw907/cc-desk`.

## Pre-flight

No step in this pass needs Geoff. Every write is inside the worktree, a test's `t.TempDir()`, or
`$CLAUDE_JOB_DIR/tmp`.

- Permissions checked: `~/.local/state/claude-denials.jsonl`, read again on 2026-11-17 for this
  pass. 129 rows, 2026-11-01T00:55:05Z to 2026-11-18T03:06:17Z, no row newer than the round
  plan's read. Classes and rows: Classifier unavailable 47; Destructive or Unwanted Actions 14;
  Self-Modification 11; Interfere With Workloads 11; server-side classifier, no reason 10;
  Undermining Oversight 4; Unauthorized Access 4; Merge Without Review 3; Auto-Mode Bypass 3;
  Irreversible Local Destruction 2; Git Destructive 2; Create Public Surface 2; Unverifiable
  Deletion Scope 2; Sensitive-Source Provenance 2; Security Weaken 2; Production Deploy 1;
  Secret-Store Writes 1; DNS/Domain/Cert Changes 1; Bot Check Circumvention 1; Code from External
  1; Shared Scratch Sweep 1; Safety Bypass Flag 1; Out-of-Place Publication 1; PII Data Handling
  1; Modify Shared Resources 1. This pass adds no dependency and runs `go build`, `go test`,
  `gofmt`, `go vet`, `vale`, `bash scripts/gate.sh`, `git push origin state-rules-estimate`, `gh
  pr create`, `gh pr ready`, `gh pr checks`, and `gh run view`. No row refuses any of them; the one row that names `vale` is a classifier outage on a `grep`
  in another repo, and no row names `gh pr ready`. The rows near them are in the table below.
- Reviewed: one `standards:plan-reviewer` read (Opus, medium effort) of this plan. Its verdict
  was "ready after fixes"; all fifteen findings are folded, as `## Decisions` lists. The class
  needs none before code; the launch prompt asked for one.
- Evidence: the mechanism table and the recorded searches below. Each acceptance check was run on
  today's code before any code was written; the outputs are under `## Acceptance`.

### Refused commands and their alternatives

| Refused in the log | This pass does instead |
| --- | --- |
| `kill <pid>` before `go test` (Interfere With Workloads, 2026-11-09) | Stop a background command with `TaskStop` |
| `rm -rf /tmp/tmp.*` sweeps (Shared Scratch Sweep) | Scratch lives in `t.TempDir()` or a fresh, named folder under `$CLAUDE_JOB_DIR/tmp`. No sweep, no `rm -rf` |
| `git push origin HEAD:main` (2026-11-10, 2026-11-15) | Push only `state-rules-estimate` |
| `git stash drop` | No stash. A temporary commit sets work aside |
| A chained `claude stop`, and `claude agents --json` piped to a script (2026-11-17) | Not needed by the executor. If ever needed, run `claude stop <id>` alone and redirect `claude agents --json` to a file |
| A write to a Claude `settings.json` (Self-Modification, 2026-11-18) | No settings edit. A task that needs one parks (round-run standing answer 3) |
| A classifier outage | Wait 60 seconds and retry while the cause is still the outage, inside the 150-minute ceiling. A task that can't finish parks by standing answer 26 |
| `gh pr merge`, and `gh pr view` used to judge a merge (Merge Without Review) | The pass merges nothing. It ends with the PR ready. Read checks with `gh pr checks <n>` |

## Mechanism table

| Mechanism | Probe or source |
| --- | --- |
| Base | `git rev-parse --short HEAD` and `origin/main` both printed `f81af60` after a fetch; branch `state-rules-estimate` |
| The gate on today's code | `bash scripts/gate.sh` printed `ok` for all five packages (e2e in 11.386s), then `gate: ok`. CI on `main` at `f81af60`: run 44246989796, `success` |
| The gate covers this plan | `scripts/gate.sh` runs vale over every tracked `docs/*.md`, `docs/plans/` included. A vale run on this plan's first draft flagged two words as spelling errors; this text drops them. Every commit of a `docs/` file runs vale first |
| One path for every verb | `internal/state/commit.go:42` `Commit(dir, now, wait, ev, render)`: lock, `Load`, remove `snapshot-error.json`, `ev`, `settle` (empty, pass 3), write dirty releases with `atomicfile.WriteFile`, `render`, write `snapshot.json`. An `ev` error returns before any write. Its one caller is `cmd/cc-desk/root.go:63` |
| No import cycle | PR #3's `[advisor]` line: `state.Commit` takes the render function as an argument so `state` never imports `snapshot`. `internal/state/commit.go:27` `type Render func(s *State, now time.Time) ([]byte, error)` |
| The forecast line today | `internal/state/types.go:139` `Forecast{Format; Made time.Time; Project; Round; Time time.Time}`. `internal/state/read.go:241-246` rejects a line missing made, project, round, or time. Nothing writes the file: after `pass start` and `pass finish` on a fixture copy, `forecasts.jsonl` still held its one fixture line |
| Pass events | `internal/state/events.go:151` `StartPass` (start and resume), `:180` `FinishPass` (actual minutes are start to finish less closed stops, `math.Round` to the minute), `:214` `StopPass` (opens an interval with a nil `End`). `FinishPass` and `StopPass` refuse a clock before the start |
| Float minutes miss literals | A probe computing `200 × (1100/1000)` in `float64` gave `220.00000000000003` minutes; `100 × 1.2 + 200 × 1.2` gave exactly 360. The result depends on the numbers, so the estimate uses integer seconds (Decisions). `json.Marshal(1100.0/1000.0)` prints `1.1`, so the display ratios compare as literals |
| Time zone | Alaska is UTC-9 on 2026-11-20. `e2e/main_test.go:18` `workedClock = "2026-11-20T23:00:00Z"`, "2:00 pm Alaska time" |
| The fixture | `alpha` r1 holds p01 to p12 at 100 planned, actual 300, 300, 150, 80, 120 six times, 125, 125, finished 22:00Z on 2026-11-08 to 2026-11-19. r2 holds p13 (100), p14 (200, after p13), p15 (60, after p14, rests on `alpha-q1`, unanswered). `beta` b1 holds b-one (40 planned, 45 actual) and b-two (50, 55), done. b2 holds b-three (60, after b-two, both questions answered). b3 is not designed, `needs_geoff`, 90 planned, after b2. At pace 1.0 b3 starts at 3:00 pm, ahead of alpha's 8:00 pm. `"planned_minutes": 60` appears once in `beta.json`, and no test depends on it |
| A design input for b3 | `e2e/testdata/inputs/design-b3.json` adds pass b-four (90, round b3, rests on `beta-q3`, unanswered) |
| The fixture golden | `e2e/events_test.go` `TestFixtureSnapshotIsCurrent` compares `e2e/testdata/state/snapshot.json` with the snapshot verb's output at `workedClock`. Any change to what `Build` prints changes it, so every task that changes the output regenerates it in the same commit |

## Recorded searches

Each ran with `git grep -n -E '<pattern>' -- '*.go' e2e/testdata/README.md docs/STATUS.md` at
`f81af60`.

| Pattern | Hits | Each hit |
| --- | --- | --- |
| `Pass 2\|pass 2` | 3 | `docs/STATUS.md:37` (STATUS line, the executor's STATUS update covers it); `internal/snapshot/snapshot.go:102` and `:120` (the comments on `NextSession` and `WaitEntry`, changed in task 2) |
| `Render\|render\b` | 16 | `cmd/cc-desk/root.go:63` (caller, task 3); `internal/snapshot/snapshot.go:143-144` (task 3); `internal/snapshot/snapshot_test.go:20`, `:46` (task 3, new return value); `internal/state/commit.go:26-27`, `:42`, `:89` (task 3); `internal/state/commit_test.go:14` (the stub, task 3) and `:39`, `:60`, `:72`, `:98`, `:118`, `:129` (calls through the stub, unchanged) |
| `Forecast\|forecasts` | 10 | `internal/snapshot/snapshot.go:61` (`ForecastTime`, task 1); `internal/state/read.go:167` (comment, unchanged), `:238-244` (task 3); `internal/state/read_test.go:85` (still fails, task 3 adds cases); `internal/state/types.go:138-139` (task 3), `:196` (the `Forecasts` slice, unchanged) |
| `StartPass\|FinishPass\|StopPass` | 25 | `cmd/cc-desk/pass.go:21`, `:34`, `:48` (callers, unchanged); `internal/state/events.go:149-151`, `:177-180`, `:212-214` (task 3 marks each as a pass event); 16 hits in `internal/state/events_test.go` (unchanged; task 3 adds a mark check) |
| `workedClock` | 11 | `e2e/main_test.go:16-18` (the constant); 10 uses in `broken_test.go`, `events_test.go`, `lock_test.go`, all on refusals or the snapshot verb, unchanged |
| `sameFolder\|folder\(t` | 14 | The two definitions at `e2e/main_test.go:93` and `:123`; the rest in `e2e/broken_test.go`, `events_test.go`, `lock_test.go`, around refused or failed verbs. A refusal appends no forecast, so none changes |
| `next_session` in tests | 4 | `internal/snapshot/snapshot_test.go:65-75`: expects `state` and `question_count` null on an empty state. Task 2 changes it |

Golden files that change: `internal/snapshot/testdata/empty.golden.json` (task 2) and
`e2e/testdata/state/snapshot.json` (tasks 1 and 2), each regenerated by the command the task
gives and read by eye before commit.

## File list

| File | Change | Task |
| --- | --- | --- |
| `e2e/testdata/state/releases/beta.json` | b-three planned 60 becomes 480 | 1 |
| `e2e/testdata/state/snapshot.json`, `e2e/testdata/README.md` | Regenerated golden; README states beta's figures | 1, 2 |
| `internal/snapshot/estimate.go` | New: pace, range, rough; needed times; the next session; `waits_for_geoff` | 1, 2 |
| `internal/snapshot/estimate_test.go` | New: unit tests over `state.State` values built in the test | 1, 2 |
| `internal/snapshot/snapshot.go` | `Build` calls the estimate; `NextSession` gains `waiting_on_design`; `Render` also returns the forecast | 1, 2, 3 |
| `internal/snapshot/snapshot_test.go` | The empty state reads `nothing_waiting` with a count of 0 (task 2); `Render`'s new return value (task 3) | 2, 3 |
| `internal/snapshot/testdata/empty.golden.json` | Regenerated | 2 |
| `internal/state/types.go` | `Forecast.Time` may be null; a private pass-event mark on `State` | 3 |
| `internal/state/read.go`, `read_test.go` | A forecast line holds project, round, and time together or none of them | 3 |
| `internal/state/events.go`, `events_test.go` | Start, finish, and stop set the pass-event mark | 3 |
| `internal/state/commit.go`, `commit_test.go` | `Render` returns the forecast; `Commit` appends it after a pass event | 3 |
| `cmd/cc-desk/root.go` | No change expected; `snapshot.Render` keeps its name | 3 |
| `e2e/estimate_test.go` | New: done item 1 | 4 |
| `docs/STATUS.md` | The pass's STATUS lines | 5 |

## The rules this pass builds

These are plain functions in `internal/snapshot/estimate.go` over the loaded `state.State` and the
clock. `state` holds no estimate logic.

**Pace.** For each project, take its finished passes from its live release and every release in
`releases/done/` with that project, sort by finish time (ties by release file order, then pass
order), and keep the last ten. With fewer than three, the pace is 1.0, `rough` is true, and
`pace_low` and `pace_high` are null. Otherwise the pace is the sum of actual minutes over the sum
of planned minutes, `rough` is false, and the range ends are the smallest and largest single-pass
ratio. `pace` is always set.

**Arithmetic.** Every duration is whole seconds in integers. A pass's paced length at a ratio
`a/p` (a sum of actual minutes over a sum of planned minutes, or one pass's actual over planned)
is `round(planned × 60 × a / p)` seconds, rounded half up. The pace, low, and high each use
their own integer pair. The snapshot's `pace`, `pace_low`, and `pace_high` are the same ratios as
`float64`, for display only.

**What comes before.** The predecessors of a pass are found transitively: each pass in its
`after`, and every pass of every round in its round's `after`. A round in that chain that has no
passes yet counts its round-level `planned_minutes` once, and its own `after` rounds are
followed. Each pass or round counts once. Only unfinished ones count.

**Counting an unfinished one.** A `not_started` or `waiting` pass, or a round with no passes
yet, counts its paced length. A `building` or `stopped` pass counts its paced length less its
active seconds so far, never below zero. Active seconds run from the start to now, less every
closed stop and less an open stop up to now, the same accounting `FinishPass` uses.

**Needed times.**

| Thing | `minutes_before` | `needed_time` |
| --- | --- | --- |
| A `not_started` or `waiting` pass | Sum of the raw planned minutes of its unfinished predecessors | now plus the counted seconds of those predecessors |
| A `building`, `stopped`, or `done` pass | null | null |
| A round that hasn't started (`not_started` or `not_designed`) | Sum of the raw planned minutes of the unfinished work its `after` rounds reach | now plus that work's counted seconds |
| A round that is building or done | null | null |
| A `not_designed` round that needs Geoff | As above | As above; `forecast_time` equals it |
| A written, unanswered question | (not in the snapshot) | The earliest `needed_time` of the `not_started` or `waiting` passes resting on it. With none (its passes are all under way or done) it is needed now |

The low and high ends of any of these use the same predecessors at the low and high ratios,
with the same building refinement. A rough project has no ends.

**The next session.** Items are every written, unanswered question in a live release, and every
`not_designed` round that needs Geoff whose start is still ahead. A `not_designed` round whose
start has arrived (needed time at or before now) is not an item; it is listed under
`waiting_on_design`. The item with the earliest needed time sets the session's time; ties go to
release order, then round order, questions before rounds. A question's round is the round of
its earliest resting pass, in release then round order; that round names it in `rounds`,
`waits_for_geoff`, and the forecast line. Names in `rounds`, `not_designed`, and
`waiting_on_design` are `<project>/<round>`, in release then round order. `projects` lists each
project that has a written, unanswered question.

| State, first that applies | `time` | `range` | `rough` | `question_count` |
| --- | --- | --- | --- | --- |
| `needs_you_now`: a live pass is `stopped` (every stop reason is one of the five). Pass 3 adds the unreadable usage figure | The setting item's needed time, or null with no item | The setting item's ends, or null | The setting item's project's rough mark, or null with no item | Count of written, unanswered questions |
| `ready_now`: a written, unanswered question is needed now | now | now to now, or null if rough | The setting item's project's rough mark | As above |
| `ready_now`, with no question: a round that needs Geoff is waiting on its design and no item is left | null | null | null | 0 |
| `scheduled`: there is an item and its time is ahead | The setting item's needed time | Its ends, or null if rough | Its project's rough mark | As above |
| `nothing_waiting`: no item, and nothing in `waiting_on_design` | null | null | null | 0 |

The third row is the one case the design's four states leave open; see `## Decisions`.

`waits_for_geoff` has one entry per round with written, unanswered questions needed now, with the
count of those questions.

**The forecast line.** Each successful pass start, finish, or stop appends exactly one line to
`forecasts.jsonl`. It holds `format`, `made` (the clock), and the next session's `time` from the
snapshot that the same commit writes, with `project` and `round` of the item that set it. When the
session has no time, the line holds `made` and `"time": null` and no project or round. A refused
event appends nothing. In pass 3, its settled starts mark a pass event; its recorded waits don't.

## Acceptance

Every check runs the built `cc-desk` binary against a copy of the fixture state folder in
`t.TempDir()`, through `go test ./e2e/...`. Times are literals in the test: Alaska time is UTC-9,
so 2:00 pm is `2026-11-20T23:00:00Z` and midnight is `2026-11-21T09:00:00Z`.

| Item, word for word | Task | Check |
| --- | --- | --- |
| **Estimate.** A fixture holds two releases in two projects, one of them the worked case above with all twelve finished passes. The snapshot's next session equals each row of that case's table, as literal times in the test: 8:00 pm with 6:00 pm to 9:30 pm at the start and again at 2:30 pm, then 6:30 pm with 4:30 pm to 7:00 pm after the fast finish, or 9:30 pm with 8:00 pm to midnight after the slow one. The question count and the project list equal stated figures. The second project has two finished passes and a later needed time: its estimate is marked rough with a pace of 1.0 and no range, and the session carries the first project's range and no rough mark. A variant makes the second project the earliest, and the session is then rough. A round marked as needing Geoff with no design is named as not designed yet, with a start time equal to a stated figure. Each line that `forecasts.jsonl` gains holds the time it was made and the same needed time the snapshot showed. | 1, 4 | `go test ./e2e/ -run TestEstimate -v` prints `--- PASS` for `TestEstimate` and each of its six `t.Run` cases |
| **The gate.** `bash scripts/gate.sh` passes, and so does CI on the pull request. The gate fails if any test was skipped, so a runner without a terminal can't pass item 4 by skipping it. | 5 | `bash scripts/gate.sh` run bare on the final head prints `gate: ok`; the PR's `gate` check on that head is `pass`. The skip check landed in pass 1 (item D, `e2e/gate_test.go`) and stays |

What each check printed on today's code, before any change. A built binary in
`$CLAUDE_JOB_DIR/tmp` ran on a copy of `e2e/testdata/state`:
`pass start alpha p13` at 2:00 pm (exit 0), `snapshot` at 2:30 pm (exit 0), `pass finish alpha
p13` at 2:50 pm (exit 0). The snapshot held:

```text
{"state":null,"time":null,"range":null,"rough":null,"question_count":null,"rounds":[],"projects":[],"not_designed":[]}
waits_for_geoff: []
{"project":"alpha","pace":null,"pace_low":null,"pace_high":null,"rough":null}
{"project":"beta","pace":null,"pace_low":null,"pace_high":null,"rough":null}
{"name":"p13","minutes_before":null,"needed_time":null}  (p14 and p15 the same)
{"name":"b3","state":"not_designed","needs":"not_designed","forecast_time":null}
```

`forecasts.jsonl` still held its one fixture line after both events. `TestEstimate` doesn't exist
yet, so `go test ./e2e/ -run TestEstimate -v` reports no tests to run, which fails the check's
`--- PASS` requirement. The gate check today is the `gate: ok` in the mechanism table.

## Tasks

Every check below that runs `go test -run` names the new tests, so today's tree, where they
don't exist, fails it. Each task's commit runs `bash scripts/gate.sh` bare first.

To regenerate the fixture golden, run these from the worktree, with a new folder name each
time (`fx1`, `fx2`):

```bash
go build -o "$CLAUDE_JOB_DIR/tmp/cc-desk" ./cmd/cc-desk
cp -r e2e/testdata/state "$CLAUDE_JOB_DIR/tmp/fx1"
"$CLAUDE_JOB_DIR/tmp/cc-desk" --state-dir "$CLAUDE_JOB_DIR/tmp/fx1" --now 2026-11-20T23:00:00Z snapshot > e2e/testdata/state/snapshot.json
```

Then read `git diff e2e/testdata/state/snapshot.json` by eye.

### Task 1: the fixture, pace, and needed times (30 minutes)

Outcome: beta's needed times sit later than alpha's in every row of the worked table; every live
release carries `pace`, `pace_low`, `pace_high`, and `rough`; every pass and round carries
`minutes_before` and `needed_time` by the table above; each `not_designed` round that needs
Geoff carries `forecast_time`. Files: `e2e/testdata/state/releases/beta.json`,
`e2e/testdata/state/snapshot.json`, `e2e/testdata/README.md`, `internal/snapshot/estimate.go`,
`estimate_test.go`, `snapshot.go` (`Build` calls in). Class `ordinary`.

First, b-three's `planned_minutes` goes from 60 to 480. At pace 1.0 b3 then starts at 10:00 pm
(`2026-11-21T07:00:00Z`) at the 2:00 pm clock, 10:30 pm at 2:30 pm, 10:50 pm at 2:50 pm, and
1:20 am at 5:20 pm. Alpha's latest is 9:30 pm. The README gains beta's figures.

Test, in `estimate_test.go`, on `state.State` values the test builds. Name each test
`TestEstimate<Case>`:

- `TestEstimateWorkedCase`: the worked case at 2:00 pm with p13 not started: pace 1.2, ends 0.8
  and 1.5; p15's `minutes_before` 300 and `needed_time` `2026-11-21T05:00:00Z`.
- `TestEstimatePaceLastTen`: the pace takes the last ten by finish time across a live release
  and a `releases/done/` release of the same project, and leaves out another project's passes.
- `TestEstimateRough`: two finished passes give pace 1.0, rough, null ends.
- `TestEstimateBuilding`: a building pass started 30 minutes ago counts its paced length less 30
  minutes; one past its paced length counts zero; a stopped pass leaves out its stopped time,
  open stop included.
- `TestEstimatePredecessors`: predecessors through `after` passes and `after` rounds, a round
  with no passes yet counted once by its planned minutes, and a pass reached twice counted once.

Regenerate the fixture golden as above; it shows alpha pace 1.2 and beta pace 1.0 rough.

Acceptance check: `go test ./internal/snapshot/ -run 'TestEstimate' -v` prints `--- PASS` for
each of the five, and `go test ./e2e/ -run TestFixtureSnapshotIsCurrent -v` passes.

### Task 2: the next session and `waits_for_geoff` (20 minutes)

Outcome: `next_session` and `waits_for_geoff` hold what "The next session" above says, and
`NextSession` gains `waiting_on_design` (`[]` when empty). Files: `estimate.go`,
`estimate_test.go`, `snapshot.go`, `snapshot_test.go`, `testdata/empty.golden.json`,
`e2e/testdata/state/snapshot.json`. Class `ordinary`.

Test, as `TestNextSession<Case>` in `estimate_test.go`, one case per row of the state table:

- `TestNextSessionOrder`: a stopped pass with a question needed now reads `needs_you_now`.
- `TestNextSessionReadyNow`: a question whose only resting pass is under way is needed now.
- `TestNextSessionScheduled`: the range and rough mark come from the setting item's project
  only.
- `TestNextSessionNotDesignedSets`: a `not_designed` round whose start is ahead is the earliest
  item; the state is `scheduled`, the time is its start, and the setting item is that round's
  project and round. (Task 3 checks the forecast line that follows from it.)
- `TestNextSessionWaitingOnDesign`: a `not_designed` round whose start has arrived moves to
  `waiting_on_design`, and the time comes from the next item. With no item left, the state is
  `ready_now` with a null time and a count of 0.
- `TestNextSessionNothingWaiting`: an empty state.

`TestSnapshotEveryFieldPresent` now expects `state` `"nothing_waiting"`, `question_count` 0,
`time`, `range`, and `rough` null, and four `[]` lists. Regenerate `empty.golden.json` with `go
test ./internal/snapshot/ -update` and the fixture golden as above, and read both diffs.

Acceptance check: `go test ./internal/snapshot/ -run 'TestNextSession|TestSnapshot' -v` prints
`--- PASS` for each of the six new tests and the existing ones.

### Task 3: the forecast log (25 minutes)

Outcome: each successful pass event appends exactly one forecast line; a refusal appends none.
Files: `internal/state/types.go`, `read.go`, `read_test.go`, `events.go`, `events_test.go`,
`commit.go`, `commit_test.go`, `internal/snapshot/snapshot.go`, `snapshot_test.go`. Class
`ordinary`.

The mechanism, all of it fixed here:

1. `State` gains a private mark that `StartPass`, `FinishPass`, and `StopPass` set once they
   succeed. (In pass 3, `settle` sets it when it starts a waiting pass; `settle` stays empty
   here.)
2. `state.Render` becomes `func(s *State, now time.Time) ([]byte, Forecast, error)`.
   `snapshot.Render` builds the snapshot once and returns its bytes with the forecast drawn from
   that same `NextSession`. So the line's time can't differ from what `snapshot.json` shows.
3. In `Commit` the order is: the release writes, then `render`, then the forecast append, then
   `snapshot.json`. A marked commit reads the raw bytes of `forecasts.jsonl` (empty if absent),
   adds a newline if the last byte isn't one, appends the marshalled line and a newline, and
   writes the whole file through `atomicfile.WriteFile`. A crash can't leave a torn line for the
   strict reader.
4. `Forecast.Time` becomes `*time.Time`, and `Project` and `Round` take `omitempty`. The reader
   requires `made`, and requires `project`, `round`, and a non-null `time` together or none of
   them.

Test: `TestCommitForecast` in `commit_test.go`, whose stub returns a fixed forecast: a marked
event appends one line and keeps the earlier lines byte for byte; an unmarked one and a refused
one append nothing; a file with no final newline gains one before the new line.
`TestRenderForecast` in `snapshot_test.go`: a `not_designed` round that sets the time gives a
forecast naming its project and round, and an empty state gives a forecast with a null time.
`read_test.go`'s table adds "forecast project without time" and "forecast time without round",
each refused with the file and line. `TestPassEventMark` in `events_test.go` checks the mark
after each of the three events and its absence after a refusal.

Acceptance check: `go test ./internal/... -run 'TestCommitForecast|TestPassEventMark|TestRenderForecast|TestRead' -v`
prints `--- PASS` for each of the three new tests and the `TestRead` ones, and `go test ./internal/...` passes.

### Task 4: done item 1 end to end (25 minutes)

Outcome: `e2e/estimate_test.go` holds `TestEstimate`, whose six `t.Run` cases run the binary on copies
of the fixture with literal times. Each case builds its own copy with `newState` and replays
its events. Files: `e2e/estimate_test.go`. Class `ordinary`.

1. **`table`.** Record b3's design (`round design beta b3` with `design-b3.json`) at 2:00 pm so
   beta has a written question. Then `pass start alpha p13` at 2:00 pm, read the snapshot, and
   `snapshot` at 2:30 pm. One copy then finishes p13 at 2:50 pm; a second copy, built the same
   way, finishes it at 5:20 pm. Each snapshot's next session:

   | Clock | `time` | `range` | `rough` |
   | --- | --- | --- | --- |
   | 2:00 pm `2026-11-20T23:00:00Z` | `2026-11-21T05:00:00Z` | `2026-11-21T03:00:00Z` to `2026-11-21T06:30:00Z` | false |
   | 2:30 pm `2026-11-20T23:30:00Z` | `2026-11-21T05:00:00Z` | `2026-11-21T03:00:00Z` to `2026-11-21T06:30:00Z` | false |
   | 2:50 pm `2026-11-20T23:50:00Z` | `2026-11-21T03:30:00Z` | `2026-11-21T01:30:00Z` to `2026-11-21T04:00:00Z` | false |
   | 5:20 pm `2026-11-21T02:20:00Z` | `2026-11-21T06:30:00Z` | `2026-11-21T05:00:00Z` to `2026-11-21T09:00:00Z` | false |

   Each: `state` `scheduled`, `question_count` 2, `projects` `["alpha","beta"]`, `rounds`
   `["alpha/r2","beta/b3"]`. Beta's release reads pace 1.0, `rough` true, null ends, and
   `beta-q3`'s pass b-four reads `needed_time` `2026-11-21T07:00:00Z` at 2:00 pm. Alpha reads
   pace 1.2, ends 0.8 and 1.5, at 2:00 pm; 1.1, 0.5, and 1.25 after the fast finish; 1.25, 0.8,
   and 2.0 after the slow one.
2. **`forecast_lines`.** In each copy from case 1, `forecasts.jsonl` holds the fixture line
   plus exactly 2 lines (the start and the finish), none for `round design` or `snapshot`. Each
   new line's `made` equals the event's clock, its `time` equals that commit's
   `next_session.time`, and it names `alpha` and `r2`.
3. **`variant`.** A copy with b-three's `planned_minutes` set back to 60 (a string replace, as
   `broken_test.go` does), b3's design recorded, and p13 started at 2:00 pm: `time`
   `2026-11-21T00:00:00Z` (3:00 pm), `range` null, `rough` true, the forecast line naming `beta`
   and `b3`.
4. **`not_designed`.** The base fixture with p13 started at 2:00 pm: b3 reads `state`
   `not_designed`, `needs.kind` `not_designed`, `minutes_before` 480, `forecast_time`
   `2026-11-21T07:00:00Z`. `next_session.not_designed` is `["beta/b3"]`, `question_count` 1,
   `projects` `["alpha"]`, `time` `2026-11-21T05:00:00Z`.
5. **`waiting_on_design`.** The base fixture: `pass start beta b-three` at 2:00 pm and `pass
   finish beta b-three` at 2:10 pm (`2026-11-20T23:10:00Z`). b3 is in `waiting_on_design`, not
   in `not_designed`, and the time is alpha's: `2026-11-21T05:10:00Z`.
6. **`other_states`.** On one copy: `pass start alpha p13` at `2026-11-20T23:00:00Z`, `pass stop
   alpha p13 --reason scope` at `2026-11-20T23:10:00Z`; the state is `needs_you_now`. On a fresh
   copy: p13 from `2026-11-20T23:00:00Z` to `2026-11-20T23:50:00Z`, then p14 from
   `2026-11-20T23:50:00Z` to `2026-11-21T01:00:00Z`; the state is `ready_now` with `time`
   `2026-11-21T01:00:00Z` and `waits_for_geoff`
   `[{"project":"alpha","round":"r2","questions":1}]`.

Acceptance check: `go test ./e2e/ -run TestEstimate -v` prints `--- PASS` for `TestEstimate`
and each of the six `t.Run` cases.

### Task 5: gate, pull request, and close (15 minutes)

Outcome: the gate and CI pass on the final head; the PR is ready with its body; STATUS carries
the pass's lines. Open the PR as `round-run.md` "Evidence" says and run the whole-branch
`standards:diff-reviewer` read the class calls for. Fold its findings once. A delta review after
the fold has its Low findings folded once and checked by your own diff read and CI, or filed as
follow-ups; it never starts another round. After the last edit (STATUS and the review fold
included), run `bash scripts/gate.sh` bare and never piped, push, and watch the `gate` check
with `gh pr checks <n> --watch`. Name the CI verdict on that head in the PR body. Merge nothing.

Acceptance check: `bash scripts/gate.sh` prints `gate: ok` on the final head; the PR's `gate`
check on that head is `pass`.

## What pass 3 adds

Pass 3 adds the unreadable usage figure to `needs_you_now`, the `free` and `why_not` fields, and
the `waiting` list. Its `settle` sets the pass-event mark for each pass it starts, so the forecast
log gains a line then. A pass waiting on a limit counts as not started in the estimate, which
this pass already does. Pass 3's done item 2 reads `ready_now` from this pass's states.

## Left alone

`docs/follow-ups/snapshot-stop-times.md` also notes that alpha r1, designed with no questions and
`needs_geoff: true`, shows needs `nothing` with a null reason. That is the "what it needs" field
of pass 1's snapshot, not the estimate, so it stays in the follow-up.

## Standing answers

The twelve in `round-run.md` hold, with the round plan's changes (4 replaced, 1 and 9 changed)
and its answers 13 to 17, and pass 1's 18 to 22 (in `docs/plans/2026-11-17-state-rules-files-plan.md`).
This pass adds:

- **23. A literal in task 4 disagrees with the code.** The literal is right: the design's table
  sets it. Fix the code. Change a literal only when the plan's arithmetic is wrong, show the
  arithmetic in `## Decisions`, and flag it.
- **24. A case in the estimate the plan doesn't cover.** Take the reading closest to the design's
  text, write a unit test for it, and list it in `## Decisions`.
- **25. Changing `Render` breaks a caller this plan didn't list.** Fix the caller in this pass and
  add it to the PR body's file list.
- **26. The path runs past 150 minutes.** This is round standing answer 14 applied to this pass's
  75 planned minutes. Finish the task in flight, push, and write the stop.

## Decisions

- [advisor] From PR #3: `state.Commit` takes the render function as an argument, which avoids an
  import cycle between `state` and `snapshot`. This pass builds on it: `Render` also returns the
  forecast, so the forecast line and `snapshot.json` come from one build.
- [advisor] Durations are integer seconds with half-up rounding per paced pass, because a
  `float64` probe gave `220.00000000000003` minutes for a literal 220.
- [advisor] The fixture's b-three goes to 480 planned minutes so beta is later in all four rows.
  The "earliest" variant is a test-local string replace. The change lands in task 1 with the
  fixture golden, so the gate stays green at every commit (review finding).
- [advisor] Beta gets its written question by recording b3's design in the test, so the count and
  project list cover two projects.
- [advisor] One forecast line per pass event, always. With no session time it holds `made` and a
  null `time`, which keeps done item 6's "one line for each pass event" exact. A later reader of
  forecasts skips lines with no time.
- [advisor] The forecast file is rewritten through `atomicfile.WriteFile`, so no torn line can
  break the strict reader.
- [advisor] For Geoff at the close, and the PR's one call under `## Calls for Geoff`: a round that needs Geoff is waiting on its design and no item
  is left. The design's four states leave this case open: "Nothing waiting" requires that no
  such round lacks a design, and "Ready now" names a question. It reads `ready_now` with a null
  time, because the overdue design is work Geoff's session should take up, and "it stops setting
  the time" rules out a time. The plan reviewer recommended `ready_now`.
- [advisor] A stopped pass counts like a building one, less its active time only, so its
  estimate slides while it is stopped.
- A question whose resting passes are all under way or done is needed now. This reads the
  design's "earlier work is all done" for a question whose own passes have started.
- A question's round is the round of its earliest resting pass, in release then round order,
  for `rounds`, `waits_for_geoff`, and the forecast line.
- A `waiting` pass counts as `not_started` for a question's needed time, so pass 3's waiting
  passes keep their questions scheduled.
- `minutes_before` and `needed_time` are null for a building, stopped, or done pass or round.
  The design lists both for "every round and pass"; a time already reached has no needed time.
- State strings are `needs_you_now`, `ready_now`, `scheduled`, and `nothing_waiting`, fixed now
  because pass 3's done item 2 reads "Ready now".
- Round names in the session's lists are `<project>/<round>`, since two projects may share a
  round name.
- A round's `needed_time` is its start; a question's needed time drives the session.
- Review folds: task 4 folded into task 1; standing answer 26 set to 150 minutes; the outage row
  no longer caps retries; task 3 raised to 25 minutes; every `go test -run` check names its new
  tests; case 6 carries clocks; the close caps review rounds and reruns the gate after the last
  edit.

## Executor decisions

- A question no pass rests on is needed now with an empty round. It counts in `question_count` and
  `projects`, and stays out of `rounds` and `waits_for_geoff`. Its forecast line carries a null
  time, which differs from the snapshot in that edge case; the follow-up
  `docs/follow-ups/orphan-question-forecast.md` holds the open fix.
- `rounds` lists only question rounds, and `not_designed` lists only rounds whose start is ahead.
- In `ready_now` the range keeps the setting pass's low and high ends, so `high` can be later than
  now when a building predecessor has used its paced length at the middle pace but not the high one.

## Rulings

- Geoff, 2026-11-18, terminal session, via conductor-cc-desk: on the round that needs Geoff with no design and no item left, `ready_now` with a null time or `nothing_waiting`, "Those both work." The built `ready_now` stands.
