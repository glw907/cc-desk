# Rulings and the guard: pass plan

Initiative: cc-desk
Round: state-rules
Pass: state-rules-rulings

Class `ordinary`. Planned 60 minutes. The tasks sum to 135 minutes, and the path is 105 because
tasks 2 and 4 run in one pair beside the main lane (see "Lanes"). The park line is 120, twice the
estimate, by round standing answer 14. Pass 3 planned a 175-minute path and its code commits ran
from 22:15 to 22:28 on 2026-11-17, so these minutes sit on the same scale as pass 3's.

Design: `docs/specs/2026-11-17-state-and-rules-design.md` (approved 2026-11-17), section "Rulings
and the guard". Round plan: `docs/plans/2026-11-17-state-rules-round.md`. Where the round design
is silent, `docs/specs/2026-11-17-cc-desk-design.md` holds. Writing a person reads (the PR body,
STATUS lines, follow-ups, README lines) follows `~/.claude/docs/voice/written-for-people.md`,
loaded through the `writing-voice` skill.

Worktree `/var/home/glw907/Projects/cc-desk/.claude/worktrees/state-rules-rulings`, branch
`state-rules-rulings`, cut from `origin/state-rules-order` at `9be82ae` and fast-forwarded to
`origin/main` at `3b65cda` once pass 3 landed as PR #5. The code is the same at both heads; the
merge brought two docs files. PR base: `main`. Push: `glw907/cc-desk`. This is the round's last
pass, so its PR shows done items 1 to 8 passing on the branch.

## Pre-flight

No step in this pass needs Geoff. Every write is inside the worktree, a test's `t.TempDir()`, or
`$CLAUDE_JOB_DIR/tmp`. The one live usage read (task 8) is the zero-cost local command pass 3
ran.

- Permissions checked: `~/.local/state/claude-denials.jsonl`, read again on 2026-11-17 for this
  pass. 130 rows, 2026-11-01T00:55:05Z to 2026-11-18T06:27:23Z, the same rows pass 3 read.
  Classes and rows: Classifier unavailable 47; Destructive or Unwanted Actions 14;
  Self-Modification 11; Interfere With Workloads 11; server-side classifier, no reason 10;
  Undermining Oversight 5; Unauthorized Access 4; Merge Without Review 3; Auto-Mode Bypass 3;
  Irreversible Local Destruction 2; Git Destructive 2; Create Public Surface 2; Unverifiable
  Deletion Scope 2; Sensitive-Source Provenance 2; Security Weaken 2; one each for Production
  Deploy, Secret-Store Writes, DNS/Domain/Cert Changes, Bot Check Circumvention, Code from
  External, Shared Scratch Sweep, Safety Bypass Flag, Out-of-Place Publication, PII Data
  Handling, and Modify Shared Resources. This pass adds no dependency. It runs `go build`, `go
  test`, `gofmt`, `go vet`, `vale`, `bash scripts/gate.sh`, `git merge origin/main`, `git push
  origin state-rules-rulings`, `gh pr create`, `gh pr ready`, `gh pr checks`, `gh run list
  --branch`, `gh run view`, and once, `cc-desk usage --raw`. Its tests run `script`, `setsid`,
  and `sh`, and read `/proc`. A search of every row's target for `script `, `setsid`, `/proc`,
  `go test`, `gate.sh`, `git push`, `gh pr`, `gh run`, and `claude` found 13 rows. None refuses a
  command this pass runs; the near ones are in the table below.
- Reviewed: one `standards:plan-reviewer` read; the `[review]` line under `## Decisions` gives
  its verdict and folds. The class needs none before code; the launch prompt asked for one.
- Evidence: the mechanism table and the recorded searches below. Each acceptance check ran on
  today's code before any code was written; the outputs are under `## Acceptance`.

### Refused commands and their alternatives

| Refused in the log | This pass does instead |
| --- | --- |
| `kill <pid>` before `go test`, and a `pgrep` sweep of test processes (Interfere With Workloads, 2026-11-09 and 2026-11-15) | A test stops the desk processes it started through `t.Cleanup` and `exec.Cmd.Process.Kill` on its own `*exec.Cmd`. The session stops a background command with `TaskStop`, never `kill` or `pkill` |
| `git push origin HEAD:main` (2026-11-08, 2026-11-10, 2026-11-15) | Push only `state-rules-rulings` |
| `gh pr merge`, and `gh pr view` used to judge a merge (Merge Without Review) | The pass merges nothing. It ends with the PR ready. Read checks with `gh pr checks <n>` |
| `gh pr checks --watch` chained into a merge (2026-11-01) | Watch alone: `gh pr checks <n> --watch`, with nothing after it |
| `rm -rf /tmp/tmp.*` sweeps (Shared Scratch Sweep) | Scratch lives in `t.TempDir()` or a fresh, named folder under `$CLAUDE_JOB_DIR/tmp`. No sweep |
| `git stash drop` | No stash. A temporary commit sets work aside |
| A write to a Claude `settings.json` (Self-Modification) | No settings edit. A task that needs one parks (round-run standing answer 3) |
| A classifier outage | Wait 60 seconds and retry while the cause is still the outage |

## Mechanism table

| Mechanism | Probe or source |
| --- | --- |
| Base | `git merge --no-edit origin/main` fast-forwarded the branch to `3b65cda`; `git diff --stat 9be82ae HEAD` lists only `docs/STATUS.md` and `docs/plans/2026-11-17-state-rules-order-plan.md` |
| The gate on today's code | `bash scripts/gate.sh` printed `ok` for all six packages (e2e in 12.02s), then `gate: ok`. CI on `state-rules-order` at `9be82ae`: run 44259265664, `success` |
| CI | `.github/workflows/gate.yml` runs the gate on `ubuntu-latest` with Go 1.27.4 and Vale 3.27.0, on pull requests and pushes to `main` only. A run is listed by branch (`gh run list --branch state-rules-rulings`), never by commit: pass 3's `## Planning misses` found `--commit` returns nothing |
| One commit path | `internal/state/commit.go` `Commit(dir, now, wait, ev, render)`: lock, strict `Load`, `ev`, `settle`, write dirty releases and `usage.json`, `render`, append a forecast line when `passEvent` is set, write `snapshot.json`. It writes neither `rulings.jsonl` nor `desks.json` today. `appendForecast` rewrites `forecasts.jsonl` whole through `atomicfile`, adding a newline when the last line lacks one |
| Every verb takes a reading | `cmd/cc-desk/root.go` `rootFlags.commit(ev)` calls `commitReading`, which runs `usage.Take` before the lock and wraps `ev` in `state.WithReading` |
| Exit codes | `state.Refuse(msg)` gives a `*RefusedError`, which `main` maps to exit 2. Any other error exits 1. Refusal text today starts `cannot start <project>/<pass>: ` |
| The strict reader | `internal/state/read.go` reads `desks.json` (each desk needs `session_id`, `pid` above 0, and `started`) and `rulings.jsonl` (each line needs `id`, `time`, `desk`, and `words`). `check.go` `checkAnswers` fails a question whose `answer.ruling` names no line in `rulings.jsonl`, and `Check` fails a ruling id that appears twice |
| The types | `internal/state/types.go`: `Ruling{Format, ID, Time, Desk, Words, Project, Round, Question, Option}`, the last four `omitempty`; `Desks{Format, Desks}`; `Desk{SessionID, PID, Started time.Time}`; `Question{ID, Text, Options, Recommended, Answer}` on the release, not on a round; `Answer{Ruling, Carried}`; `Guess{Question, Recommendation, At}` |
| A guess closes on its own | `internal/state/rules.go` `OpenGuess(p)` is true while the guess's question has no answer, and `OpenGuessMinutes` sums those. Setting `Answer` closes the guess and lifts the one-pass rule with no change to `rules.go`, as pass 3's "What pass 4 adds" says |
| The guess outcome field | `internal/snapshot/snapshot.go` `Pass.GuessOutcome *string` (`guess_outcome`) is always nil today. `snapshot_test.go:128` checks it is nil on a pass with no guess, which stays true |
| The fixture's ruling id | `e2e/testdata/state/rulings.jsonl` holds one line with id `r-0001`. The design leaves "the format of the short ruling id" to the builder |
| No guess in the worked fixture | `grep -n '"guess"' e2e/testdata/state/releases/*.json` prints nothing, so a derived outcome leaves `e2e/testdata/state/snapshot.json` as it is |
| A controlling terminal, locally | On pavilion, util-linux 2.42.4, `getconf CLK_TCK` 100. A long-lived `sh` started by `script -qec "sh -c '...'" /dev/null` with standard input from `/dev/null` wrote its PID, then blocked on a FIFO. Its `/proc/<pid>/stat` read `tty_nr=34820`. Once released, its child read the shell's PID as its parent and `tty_nr=34820`, and a grandchild read the shell as its parent too. `setsid awk ... /proc/self/stat` read `tty_nr=0`, and so did this background session's own process |
| Reading `/proc` | `/proc/<pid>/stat` is parsed after the last `)`: the parent PID is field 4, `tty_nr` field 7, and `starttime` field 22, in clock ticks since boot. Boot time is the `btime` line of `/proc/stat`, whole seconds. Linux fixes the `/proc` tick (`USER_HZ`) at 100 |
| The pseudo-terminal's output | Output through the pseudo-terminal comes back with CRLF line endings, so the desk helper reads results from files the desk shell writes, never from the pseudo-terminal |
| Twenty writers on today's code | A probe folder under `$CLAUDE_JOB_DIR/tmp/w` held two releases, `east` and `west`, each with ten building passes. Twenty `pass finish` verbs launched at once with a 15-second stand-in all exited 0 after 15 seconds; the stand-in ran 20 times, all twenty passes read `done`, `forecasts.jsonl` held 20 lines, and `cmp` found `snapshot.json` equal to what the snapshot verb then printed. `Check` sets no cap on building passes, so a fixture may hold twenty |
| The environment | This session's process has `CLAUDE_CODE_SESSION_ID` set. `e2e/main_test.go` `runWith` passes `os.Environ()` plus a `PATH` of `/usr/bin:/bin`, so every test that runs the ruling verb sets `CLAUDE_CODE_SESSION_ID` itself, last in the list (`os/exec` keeps the last value of a repeated key) |
| Label wording | Geoff, 2026-11-18: settle how a state or label reads from the Survey prose standard. `/var/home/glw907/Projects/survey/skills/survey/writing/editor.md` (Microsoft voice) asks for "short sentences and short words", the everyday word, and copy that names what the product really does; its error model says what failed in plain words |

## Recorded searches

Each ran with `git grep -n -E '<pattern>' -- '*.go'` at `3b65cda`.

| Pattern | Hits | Each hit |
| --- | --- | --- |
| `rulings\.jsonl` | 11 | `e2e/broken_test.go:42` (the broken-line case, unchanged); `internal/state/check.go:282`, `:294`; `read.go:235`, `:241`; `read_test.go:44`, `:84-86`, `:114`; `types.go:133`. All unchanged; task 3 adds the writer in `commit.go` |
| `Desks\b\|desks\.json` | 16 | `e2e/broken_test.go:40`; `internal/state/read.go:172-197`; `read_test.go:78`, `:113`, `:121-122`; `types.go:158-161`, `:208`. All unchanged; task 3 adds the writer |
| `state\.Ruling\|\bRuling\b` | 19 | `internal/snapshot/snapshot.go:26`, `:180`, `:385` (task 4 reads the ruling's option); `internal/state/check.go:106`, `:275-282`; `events.go:50`; `events_test.go:283`; `read.go:235`; `rules_test.go:121`, `:139`, `:201`; `start_test.go:122`; `types.go:122`, `:133-134`, `:211`. Unchanged |
| `GuessOutcome\|guess_outcome` | 2 | `internal/snapshot/snapshot.go:88` (task 4 fills it); `snapshot_test.go:128` (stays nil on a pass with no guess) |
| `OpenGuess` | 32 | `e2e/order_test.go`, `internal/snapshot/snapshot.go`, `snapshot_test.go`, `internal/state/rules.go`, `rules_test.go`, `start_test.go`. Unchanged: an answer closes the guess |
| `appendForecast` | 3 | `internal/state/commit.go:113`, `:186`, `:189` (task 3 turns it into one append helper for both logs) |
| `AddCommand\(` | 6 | `cmd/cc-desk/root.go:120` (task 5 adds the ruling command); `pass.go:52`, `:54`, `:82`; `release.go:16`; `round.go:16` |
| `CLAUDE_CODE_SESSION_ID` | 0 | Task 2 reads it in `internal/desk` |
| `/proc/` | 0 | Task 1 reads it in the desk helper test, task 2 in `internal/desk` |
| `var verbs` | 1 | `e2e/broken_test.go:17` (task 5 adds the ruling verb; its comment asks each pass to add its verbs) |

Golden files: none changes. `e2e/testdata/state/snapshot.json` holds no guess, and
`internal/snapshot/testdata/empty.golden.json` holds no pass. `TestFixtureSnapshotIsCurrent`
fails if that reading is wrong.

## File list

| File | Change | Task |
| --- | --- | --- |
| `e2e/desk_test.go` | New: the desk helper (`startDesk`) and `TestDeskHarness` | 1 |
| `internal/state/types.go` | `Proc`, `Caller`, the three guard names, the ruling-id format, and the dirty marks for `desks.json` and a new ruling | 3 |
| `internal/state/ruling.go`, `ruling_test.go` | New: `RecordRuling`, `RegisterDesk`, the guard, and `SameStart` | 3 |
| `internal/state/commit.go`, `commit_test.go` | Write a new ruling line before the release files, and `desks.json` when it changed; one append helper for both logs | 3 |
| `internal/desk/desk.go`, `desk_test.go` | New: read a process from `/proc`, walk the caller's chain, build a desk entry | 2 |
| `internal/snapshot/snapshot.go`, `snapshot_test.go` | `guess_outcome` reads `held` or `missed` once a ruling answers the guessed question | 4 |
| `cmd/cc-desk/ruling.go`, `cmd/cc-desk/root.go` | New `ruling record` verb; `root.go` adds it | 5 |
| `e2e/broken_test.go` | The `verbs` list gains `ruling record` | 5 |
| `e2e/ruling_test.go` | New: `TestRuling`, done item 4 | 6 |
| `e2e/testdata/writers/releases/east.json`, `west.json`, `e2e/testdata/README.md` | New fixture for done item 6; the README names it | 7 |
| `e2e/writers_test.go` | New: `TestTwoWriters`, done item 6 | 7 |
| `README.md`, `docs/STATUS.md`, `docs/follow-ups/*.md` | The README's verb sentence names rulings; the pass's STATUS lines; follow-ups | 8 |

## What this pass builds

### The ruling line

`RecordRuling` appends one `Ruling` line to `rulings.jsonl`. Its id is `r-` and a four-digit
number, one more than the count of lines already there (`r-0002` after the fixture's `r-0001`);
the number widens past 9999, and an id already taken moves to the next number. Its time is the
clock, its desk is the caller's session id, and its words are the typed words, unchanged. When
the request names a question, the line also carries the project, round, question, and option.
The four come together or not at all. Words that are empty or not valid UTF-8 are a bad request
(exit 1).

A ruling that answers a question needs a live release for the project, a round of that name in
it, a question of that id, and an option in the question's options; each miss is a bad request.
A question that already has an answer is refused (exit 2): `cannot record the ruling: question
"<id>" is already answered`. The event then sets the question's `answer` to `{"ruling": "<id>"}`
and marks the release dirty. Settling runs as for any verb, so a waiting pass the ruling frees
starts at once; a `not_started` pass shows `free: true` in the snapshot.

`Commit` writes the new ruling line before any release file. A crash between the two writes
then leaves a ruling no question points at, which loads, never an answer that points at no
ruling, which `checkAnswers` fails.

### The guard

`cmd` collects the caller before the lock, with no checks: the session id from
`CLAUDE_CODE_SESSION_ID`, and for each process from its own PID up through its parents to PID 1,
the PID, the start time, and `tty_nr`. `RecordRuling` checks the three tests after the strict
`Load`, so a broken state file still names the file first. It reports the first test that
fails, and writes nothing, `usage.json` included.

| Test | Fails when | Refusal, exact |
| --- | --- | --- |
| 1 | The session id is empty, or no `desks.json` entry carries it | `cannot record the ruling: this session isn't a registered desk` |
| 2 | No process in the caller's chain has the entry's PID and a start time that matches | `cannot record the ruling: this command isn't running under the desk's process` |
| 3 | That process's `tty_nr` is 0 | `cannot record the ruling: the desk's process has no terminal` |

A start time is boot time plus `starttime` ticks at 100 a second, in UTC. `desk.Entry` and the
chain read use that one conversion in `internal/desk`. `state.SameStart(a, b)` matches two start
times less than 2 seconds apart, because `btime` is whole seconds and two reads of it can differ
by one.

`RegisterDesk(d Desk)` is the function the Desk round calls; no verb reaches it. It refuses an
entry with an empty session id, a PID of 0 or less, or a zero start time, replaces any entry with
the same session id, and marks `desks.json` for `Commit` to write atomically. The tests seed
`desks.json` by writing the file, as the design says.

### The verb

`cc-desk ruling record <words> [--project P --round R --question Q --option O]`. The words are
one argument, taken byte for byte; a desk passes words that start with `-` after `--`. On success
it prints `recorded ruling <id>` on stderr and exits 0.

### The guess outcome

The snapshot derives `guess_outcome` for a pass with a guess whose question a ruling answers:
`held` when the ruling's option equals the guess's recommendation, `missed` otherwise. A guess
whose question is still open, or answered by a carried-over answer, reads `null`. Either way the
answer closes the guess, so its minutes leave the 300 total.

## Lanes

The executor runs tasks 1, 3, 5, 6, 7, and 8 in order. Task 3's first commit adds the `Proc` and
`Caller` types alone; the executor then dispatches one pair, Sonnet at medium effort, with
`isolation: "worktree"` and the description `state-rules-rulings task 2`, which runs task 2 and
then task 4, as `references/executor.md` "Pairs" sets out. The executor merges the pair's commits
before task 5. Task 1's CI run proceeds while task 3 runs, and its verdict is needed before task 6.

The path is tasks 1, 3, 5, 6, 7, and 8: 10 + 30 + 10 + 30 + 15 + 10 = 105 minutes. Tasks 2 and 4
(20 + 10) run beside task 3.

## Acceptance

Every check runs the built `cc-desk` binary against a copy of a fixture state folder in
`t.TempDir()`, through `go test ./e2e/...`, with the usage read replaced by a stand-in script.

| Item, word for word | Task | Check |
| --- | --- | --- |
| **Ruling.** A ruling recorded from a process under a registered desk process that has a terminal adds exactly one line to `rulings.jsonl`. Its words equal the typed words byte for byte, in a case that includes quotation marks and a non-ASCII character. The question shows as answered in the snapshot and the pass that waited shows as free. One case rules for the recommended option on the 200-minute guess: the guess reads held and the open-guess total drops from 200 to 0. Another rules for a different option, and the guess reads missed. Three requests are refused, each naming its test: a session id that isn't in `desks.json`, the desk's id from a process that isn't under the desk's process, and a registered desk process that has no terminal. After each refusal the log is byte for byte what it was. | 1, 2, 3, 4, 5, 6 | `go test ./e2e/ -run TestRuling -v` prints `--- PASS` for `TestRuling` and each of its six `t.Run` cases |
| **Two writers.** Twenty event verbs launched at once, spread over both releases and including rulings, all succeed. Both release files parse and together hold every event, `forecasts.jsonl` holds one line for each pass event, and `snapshot.json` equals what the snapshot verb then prints. The usage read in this test is a stand-in that takes 15 seconds, longer than the wait for the lock. | 3, 5, 7 | `go test ./e2e/ -run TestTwoWriters -v` prints `--- PASS` for `TestTwoWriters` |
| **The gate.** `bash scripts/gate.sh` passes, and so does CI on the pull request. The gate fails if any test was skipped, so a runner without a terminal can't pass item 4 by skipping it. | 1, 8 | `bash scripts/gate.sh` run bare on the final head prints `gate: ok`; the PR's `gate` check on that head is `pass`. The skip check from pass 1 (`e2e/gate_test.go`) stays |

The design changes one case of the main design's done item 4, in its words:

> This changes one case in the main design's done item 4. That item refuses a ruling "from a
> shell with the desk session's id and no terminal". The probe below shows that a desk session's
> commands never have a terminal themselves, so the test moves to the desk's process. The case
> becomes a ruling from a shell with the desk session's id that isn't under the desk's process.
> As the main design says, the guard stops a mistake and isn't built to stop a forgery.

`TestRuling`'s `not_under_desk` case is that changed case, and its `no_terminal` case is the
moved test.

What each check printed on today's code, before any change. A binary built into
`$CLAUDE_JOB_DIR/tmp/cc-desk` ran on a copy of `e2e/testdata/order` at `2026-11-20T23:00:00Z`:

```text
$ cc-desk ruling record x
unknown command "ruling" for "cc-desk"                  # exit 1
$ cc-desk snapshot                                      # order fixture
n-big    not_started  free false  why_not ["an unanswered question"]      guess_outcome null
n-guess  not_started  free false  why_not ["an unanswered question"]      guess_outcome null
n-after  not_started  free false  why_not ["unfinished work before it"]   guess_outcome null
open_guesses [], guess_minutes_total 0
$ go test ./e2e/ -run 'TestRuling|TestTwoWriters' -v
testing: warning: no tests to run
ok  github.com/glw907/cc-desk/e2e  0.174s [no tests to run]
$ bash scripts/gate.sh
gate: ok
```

No ruling verb exists, so item 4 fails and item 6 has no rulings to launch. The twenty-writer
probe in the mechanism table shows today's lock already holds for twenty `pass finish` verbs.

### Done items 1 to 8 on the branch

The last pass shows every done item. Task 8 runs each check on the final head and quotes its
output in the PR body.

| Item | Check |
| --- | --- |
| 1 Estimate | `go test ./e2e/ -run TestEstimate -v` |
| 2 Waiting order | `go test ./e2e/ -run TestWaitingOrder -v`, all five cases |
| 3 Limits | `go test ./e2e/ -run TestLimits -v`, all six cases |
| 4 Ruling | `go test ./e2e/ -run TestRuling -v`, all six cases |
| 5 A broken file | `go test ./e2e/ -run TestBrokenFile -v`, with the ruling verb in `verbs` |
| 6 Two writers | `go test ./e2e/ -run TestTwoWriters -v` |
| 7 The usage reader | `go test ./internal/usage/ -run TestParse -v` and `go test ./e2e/ -run TestUsageReader -v`, and the live run, once, as pass 3's task 8 ran it |
| 8 The gate | `bash scripts/gate.sh` and the PR's `gate` check |

On today's code items 1, 2, 3, 5, and 7's tests printed `--- PASS`; 4 and 6 have no tests.

## Tasks

Every `go test -run` check below names a new test, so today's tree fails it. Each task's commit
runs `bash scripts/gate.sh` bare first, so the gate stays green at every commit. Load
`standards:go-conventions` and `standards:golang-spf13-cobra` before the first Go file.

### Task 1: the desk helper and the CI terminal probe (10 minutes)

Outcome: `e2e/desk_test.go` holds `startDesk(t, terminal bool, env []string, script string)`. It
starts a long-lived `sh` that writes its PID to a file and blocks on a FIFO, under `script -qec
"<sh command>" /dev/null` when `terminal` is true and under `setsid` when it is false, with standard input
from `/dev/null`, as `go test` gives it. It returns the shell's PID and a function that releases
the shell. Once released, the shell runs `script`, a command line that writes its output and exit
code to files in the test's temp folder. `t.Cleanup` releases and kills the desk through its own
`*exec.Cmd`. `TestDeskHarness` reads `/proc/<pid>/stat` directly and asserts: the terminal shell's
`tty_nr` is not 0; the `setsid` shell's is 0; and a child the released shell runs reads the shell's
PID as its parent and the same `tty_nr`. Files: `e2e/desk_test.go`. Class `ordinary`.

Test: `TestDeskHarness`, never skipped.

Acceptance check: `go test ./e2e/ -run TestDeskHarness -v` prints `--- PASS` locally. Then commit,
push, and read the PR's `gate` check by branch (`gh run list --branch state-rules-rulings`,
then `gh run view <id>`). Record the CI verdict and `script --version` from the runner's log line
if one fails. A failure goes by standing answer 34.

### Task 2: reading the process chain (20 minutes, pair)

Outcome: `internal/desk` reads a process from `/proc` (`Read(pid) (state.Proc, error)`), walks
the caller's chain from its own PID to PID 1 (`Chain() ([]state.Proc, error)`), builds the caller
(`Caller() (state.Caller, error)`, with the session id from `CLAUDE_CODE_SESSION_ID`), and builds a
desk entry from a PID (`Entry(sessionID string, pid int) (state.Desk, error)`). Start times follow
"The guard". A process that exits during the walk is an error that names the PID. Files:
`internal/desk/desk.go`, `desk_test.go`. Class `ordinary`.

Test: `TestParseStat` on recorded `stat` lines, one with a `)` and spaces inside the command
name; `TestChain` on the test's own process, whose chain ends at PID 1 and starts with its own
PID; `TestEntry` builds an entry for the test's PID whose start time `SameStart` matches a second
read.

Acceptance check: `go test ./internal/desk/ -v` prints `--- PASS` for the three.

### Task 3: the ruling event, the guard, and the writes (30 minutes)

Outcome: `RecordRuling(c Caller, r RulingRequest) Event` and `RegisterDesk(d Desk) Event` follow
"The ruling line" and "The guard". The three guard names are exported constants.
`Commit` writes the new ruling line, then the release files, `usage.json`, the forecast line,
`desks.json` when it changed, and `snapshot.json`, through one append helper for the two logs.
A ruling sets no pass event; a pass that settling starts still does. Files:
`internal/state/types.go`, `ruling.go`, `ruling_test.go`, `commit.go`, `commit_test.go`. Class
`ordinary`. The first commit adds `Proc` and `Caller` alone, so the pair can branch from it.

Test: `TestRecordRuling`, a table over hand-built states: each guard test fails alone and names
itself; an empty session id fails test 1; a dead desk PID fails test 2; a start time 1 second off
passes and 3 seconds off fails; empty words and invalid UTF-8; an unknown project, round,
question, or option; part of the four; an answered question refused; a ruling with no question;
the next id after `r-0001`, and after a hand-written `r-0002` gap. `TestRegisterDesk`: a new
entry, a replaced one, and each bad field. `TestCommitRuling`: the line lands with a newline when
the last line lacks one, a refused ruling leaves the folder as it was, and a registered desk
writes `desks.json`.

Acceptance check: `go test ./internal/state/ -run 'TestRecordRuling|TestRegisterDesk|TestCommitRuling' -v`
prints `--- PASS` for the three, and `go test ./internal/...` passes.

### Task 4: the guess outcome (10 minutes, pair)

Outcome: `guess_outcome` follows "The guess outcome", as constants `held` and `missed` in
`snapshot`. Files: `internal/snapshot/snapshot.go`, `snapshot_test.go`. Class `ordinary`.

Test: `TestSnapshotGuessOutcome` on hand-built states: held, missed, open (null), and a pass with
no guess (null).

Acceptance check: `go test ./internal/snapshot/ -run TestSnapshotGuessOutcome -v` prints `--- PASS`,
and `go test ./e2e/ -run TestFixtureSnapshotIsCurrent` passes with no golden change.

### Task 5: the verb (10 minutes)

Outcome: `cc-desk ruling record` follows "The verb". It builds the caller with `desk.Caller()`
before `f.commit`, so the reading and the chain both come before the lock. `e2e/broken_test.go`
gains `{"ruling record", ... {"ruling", "record", "go"}}`, run with the fixture's session id, so
each broken file is named before the guard runs. Files: `cmd/cc-desk/ruling.go`,
`cmd/cc-desk/root.go`, `e2e/broken_test.go`. Class `ordinary`.

Test: `TestBrokenFile` with the new verb.

Acceptance check: `go test ./e2e/ -run TestBrokenFile -v` prints `--- PASS` with a `ruling record`
case under each broken file, and `cc-desk ruling record --help` prints the usage line.

### Task 6: done item 4 end to end (30 minutes)

Outcome: `e2e/ruling_test.go` `TestRuling` runs six cases, each on a fresh copy of
`e2e/testdata/order`, with `desks.json` written by the test from `desk.Entry` for the desk it
started. Every ruling runs inside a desk through `startDesk`, with the words and flags in
environment variables the desk's command line expands in double quotes, so the words reach the argument list
unchanged. The words are `Use the "short form" for the café list.`

| Case | Desk | Session id | Request | Expect |
| --- | --- | --- | --- | --- |
| `answers_and_frees` | terminal | the desk's | north-q1, `yes`, the words | exit 0; `rulings.jsonl` gains exactly one line; its decoded words equal the words byte for byte; the snapshot shows north-q1 answered by that id and `n-big` with `free: true` and `why_not: []` |
| `guess_held` | terminal | the desk's | after `pass start north n-guess --guess north-q2` at `at(10)`, which reads total 200: north-q2, `yes` | `n-guess` reads `guess_outcome: "held"`, `open_guesses` is `[]`, `guess_minutes_total` drops from 200 to 0 |
| `guess_missed` | terminal | the desk's | the same start, then north-q2, `no` | `guess_outcome: "missed"`, total 0 |
| `unknown_session` | terminal | `not-a-desk` | north-q1, `yes` | exit 2, stderr is test 1's refusal, folder byte for byte as before |
| `not_under_desk` | terminal, registered | the desk's | run by the test, outside the desk | exit 2, test 2's refusal, folder unchanged |
| `no_terminal` | `setsid`, registered | the desk's | north-q1, `yes` | exit 2, test 3's refusal, folder unchanged |

Files: `e2e/ruling_test.go`. Class `ordinary`.

Acceptance check: `go test ./e2e/ -run TestRuling -v` prints `--- PASS` for `TestRuling` and each of
the six cases.

### Task 7: done item 6 end to end (15 minutes)

Outcome: a fixture `e2e/testdata/writers/` holds releases `east` and `west`, each with one
designed round that needs nothing from Geoff (with its `no_need_reason`), seven building passes,
and three unanswered questions, and no `usage.json`, so every verb takes a reading. Nothing waits,
so settling starts nothing. `TestTwoWriters` launches twenty verbs at once with a 15-second
stand-in: seven `pass finish` and seven `pass stop` across both releases, run by the test, and
six `ruling record` verbs, one per question, run as background jobs inside one terminal desk.
Files: `e2e/testdata/writers/releases/east.json`, `west.json`, `e2e/testdata/README.md`,
`e2e/writers_test.go`. Class `ordinary`.

Test: every verb exits 0; the stand-in ran 20 times and the test took at least 15 seconds, so the
read really ran; both release files load; every finished pass reads `done`, every stopped pass
`stopped`, and every question an answer naming its ruling; `rulings.jsonl` holds six lines with
six distinct ids; `forecasts.jsonl` holds 14 lines, one per finish and stop; `snapshot.json`
equals the snapshot verb's output at the same clock.

Acceptance check: `go test ./e2e/ -run TestTwoWriters -v` prints `--- PASS`.

### Task 8: done items 1 to 8, the pull request, and close (10 minutes)

1. Run each check in "Done items 1 to 8 on the branch" on the final head. Run the live usage
   read once, as pass 3's task 8 ran it, in a fresh `$CLAUDE_JOB_DIR/tmp/live` folder, and quote
   the verb's line and the three printed lines.
2. Update `README.md`'s verb sentence to name rulings, and `docs/STATUS.md` with the pass's lines.
3. Get one whole-branch `standards:diff-reviewer` read and fold it once.
4. After the last edit, run `bash scripts/gate.sh` bare, push, and watch with `gh pr checks <n>
   --watch` alone. Name the CI verdict on that head in the PR body.

Acceptance check: `bash scripts/gate.sh` prints `gate: ok` on the final head; the PR's `gate`
check on that head is `pass`; the PR body quotes each item's output.

The tasks sum to 135 minutes: 10, 20, 30, 10, 10, 30, 15, and 10.

## Left alone

The Desk round calls `RegisterDesk` when `cc-desk` opens a desk, and passes rulings to the
sessions they affect. What to do with work built on a missed guess is the conductor's call in a
later round. A ruling that changes what a release ends up as is the Sessions round's. The open
follow-ups in `docs/follow-ups/` stay open.

## Standing answers

The twelve in `round-run.md` hold, with the round plan's changes and its answers 13 to 17,
pass 1's 18 to 22, pass 2's 23 to 26, and pass 3's 27 to 33. Answer 30 (the live read) holds for
task 8. This pass adds:

- **34. CI gives the desk helper no terminal.** If `TestDeskHarness` fails on the runner because
  `script` is missing, exits early on standard input's end, or gives `tty_nr` 0, switch the desk helper's
  terminal wrapper to `python3 -c 'import pty, sys; pty.spawn(sys.argv[1:])'`, which gives the
  child a controlling terminal the same way, and push once more. That is a mechanism change with
  the same outcome: a `## Decisions` line, not a call. If that fails too, round answer 13 holds:
  ask the advisor for a check that shows the same outcome, flag it under `## Calls for Geoff`, and
  never skip.
- **35. The guard names, the refusal prefix, or the id format.** They are this plan's exact
  strings. Never reword them.
- **36. A ruling case the plan doesn't cover.** Take the reading closest to "Rulings and the
  guard", write a unit test for it in `ruling_test.go`, and list it under `## Decisions`.
- **37. A start time fails to match.** The 2-second window is fixed. Trace the cause; a fix goes
  in `internal/desk`'s one conversion with a `## Decisions` line. Never widen the window to pass
  a test.
- **38. A writer fails in `TestTwoWriters`.** Trace it. Never change `LockWait` or the
  stand-in's 15 seconds, which the design sets. A cause in the commit path is fixed in this pass.
- **39. The path runs past 120 minutes.** Round answer 14 applied to this pass's 60 minutes.
  Finish the task in flight, push, and write the stop.

## Decisions

- [advisor] The guard runs inside the event, after the strict `Load`, on a caller `cmd` collected
  before the lock. A broken state file then names the file for the ruling verb, as done item 5
  needs, and `state` never reads `/proc`.
- [advisor] Ruling ids are `r-` and four digits, the fixture's `r-0001` form. The design leaves
  the format to the builder; a count under the lock is the same on every run.
- [advisor] The guess outcome is derived in the snapshot from the answering ruling's option, not
  stored on the pass. The ruling line already holds the option, so nothing can disagree.
- [advisor] Start times match within 2 seconds. `btime` is whole seconds, and reusing a PID
  inside 2 seconds is out of the guard's reach, which stops mistakes, not forgery.
- [advisor] The new ruling line is written before the release files, so a crash leaves a state
  that loads.
- The guard names follow the Survey prose standard
  (`/var/home/glw907/Projects/survey/skills/survey/writing/editor.md`): short everyday words that
  say what failed, with contractions as the Microsoft voice uses them. The prefix `cannot record
  the ruling:` matches the command's `cannot start` refusals.
- A ruling for an answered question is refused (exit 2): the state doesn't allow it. An unknown
  project, round, question, or option is a bad request (exit 1), as `round "x" not found` is
  today.
- A ruling sets no pass event, so `forecasts.jsonl` gains a line only when its settle starts a
  pass, as pass 3 built it. Item 6's fixture has nothing waiting, so its count is 14.
- The desk helper uses `script` because the local probe proved it, with a Python pseudo-terminal as the named
  fallback (standing answer 34). Results come back through files because the pseudo-terminal turns line ends
  into CRLF.
- Pass 3's resume rule is Geoff's ruling of 2026-11-18, "Yes as built." This pass doesn't build
  on it.
- The executor runs on Sonnet at medium effort, as the round's launch line names it. The plan
  spells out each string and test, and the pair takes the two tasks that touch no other file.

## Executor decisions

## Rulings

Carried from pass 3: Geoff, 2026-11-18, terminal session, via conductor-cc-desk, on a resumed
stopped pass and the limits: "Yes as built." Carried from pass 2: Geoff, 2026-11-18, on the round
that needs Geoff with no design: "Those both work." Geoff, 2026-11-18, on wording: "In general
you can look at survey prose standard for that sort of thing."
