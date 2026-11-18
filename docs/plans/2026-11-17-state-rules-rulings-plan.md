# Rulings and the guard: pass plan

Initiative: cc-desk
Round: state-rules
Pass: state-rules-rulings

Class `ordinary`. Planned 60 minutes. The eight tasks run in one lane, so the path is their sum,
115 minutes. The park line is 120, twice the estimate, by round standing answer 14. The minutes
come from pass 3's measured pace: its plan set a 175-minute path, and its code commits ran 13
minutes, from 07:15Z to 07:28Z on 2026-11-18. These estimates keep a wide margin over that pace.

Design: `docs/specs/2026-11-17-state-and-rules-design.md` (approved 2026-11-17), section "Rulings
and the guard". Round plan: `docs/plans/2026-11-17-state-rules-round.md`. Where the round design
is silent, `docs/specs/2026-11-17-cc-desk-design.md` holds. Writing a person reads (the PR body,
STATUS lines, follow-ups, README lines) follows `~/.claude/docs/voice/written-for-people.md`,
loaded through the `writing-voice` skill.

Worktree `/var/home/glw907/Projects/cc-desk/.claude/worktrees/state-rules-rulings`, branch
`state-rules-rulings`, cut from `origin/state-rules-order` at `9be82ae` and fast-forwarded to
`origin/main` at `3b65cda` once pass 3 landed as PR #5, then merged with `origin/main` at
`92dac54`. The code is the same at every one of these heads; the merges brought docs files only. PR base: `main`. Push: `glw907/cc-desk`. This is the round's last
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
  command this pass runs; the near ones are in the table below. `gh pr create` sits near the
  Create Public Surface class (2 rows, neither a PR), and pass 3's `gh pr create` went through.
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
| A controlling terminal, on Ubuntu | The plan review ran the same desk in an `ubuntu:24.04` container under `podman`, with no terminal, `/bin/sh` as dash, and util-linux 2.39.3. The desk read `tty_nr=34816`, stayed alive with stdin from `/dev/null`, and `script` exited 0. A background job `( ... ) &` read the desk shell as its parent there and on Fedora; `sh -c "..."` read a middle process as its parent on both. `setsid` read 0. A container isn't the runner VM, so task 1's CI run is the proof |
| The caller's terminal | A command the desk shell runs inherits the desk's terminal (`tty_nr=34820`). Under `setsid` inside the desk it reads `tty_nr=0` and still has the desk shell as an ancestor, on bash and on dash. That is a desk session's command as the design describes it |
| The Python fallback | `python3 -c 'import pty, sys; pty.spawn(sys.argv[1:])' sh -c '...'` with stdin from `/dev/null` gave the shell `tty_nr=34820` and exited 0 locally. `ubuntu-latest` carries `python3`; the bare `ubuntu:24.04` image doesn't |
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
| `internal/state/types.go` | `Proc` and `Caller` (task 2); `RulingRequest`, the three guard refusals, the ruling-id format, and the dirty marks for `desks.json` and a new ruling (task 3) | 2, 3 |
| `internal/state/ruling.go`, `ruling_test.go` | New: `SameStart` (task 2); `RecordRuling`, `RegisterDesk`, and the guard (task 3) | 2, 3 |
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
ruling, which `checkAnswers` fails. A retry after such a crash adds a second line for the same
question; the first stays as an orphan. Task 8 files that as a follow-up.

`RecordRuling` checks in this order, and the first failure wins: the three guard tests; then the
words; then the four question fields, the project, round, question, and option; then whether the
question is already answered. A caller that isn't a desk always gets a guard refusal, whatever
else is wrong with the request.

### The guard

`cmd` collects the caller before the lock, with no checks: the session id from
`CLAUDE_CODE_SESSION_ID`, and for each process from its own PID up through its parents, the PID,
the start time, and `tty_nr`. The walk stops at PID 1 or at a parent PID of 0, which is how the
top of a PID namespace shows. Collecting the caller never fails the verb: a process that exits or
can't be read mid-walk ends the chain there, and the guard then decides. `RecordRuling` checks
the three tests after the strict `Load`, so a broken state file still names the file first. It
reports the first test that fails, and writes nothing, `usage.json` included.

The types, in `internal/state/types.go`:

- `Proc{PID int; Started time.Time; TTY int}`, where `TTY` is `tty_nr`.
- `Caller{SessionID string; Chain []Proc}`, where `Chain[0]` is the calling process.
- `RulingRequest{Words, Project, Round, Question, Option string}`.

Test 3 reads the `TTY` of the desk's process, the chain entry that test 2 matched. It never reads
the caller's own `TTY`: a desk session's commands have no terminal, as the design's probe found.

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

One lane, no pair. The executor runs tasks 1 to 8 in order. Task 1's CI run proceeds while tasks
2 to 5 run, and its verdict is needed before task 6 starts. The path is the sum: 10 + 15 + 25 +
5 + 10 + 25 + 15 + 10 = 115 minutes. A pair would have saved 20 minutes, but no pass in this
round has run one, and nothing has shown that a background dispatch survives an unattended run.

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

Outcome: `e2e/desk_test.go` holds `startDesk(t, terminal bool, env []string, script string)`.

1. It starts a long-lived `sh` that writes its PID to a file and blocks reading a FIFO. The shell
   runs under `script -qec "<sh command>" /dev/null` when `terminal` is true, and under `setsid`
   when it is false. Standard input is `/dev/null`, as `go test` gives it, and the desk command's
   `Stdout` and `Stderr` go to a file in the test's temp folder, never to the test binary's pipes.
   `cmd.WaitDelay` is set, as pass 3's stand-in sets it.
2. It waits up to 5 seconds for the PID file. Past that it fails the test and prints the desk's
   output file.
3. It returns the shell's PID and a function that releases the shell by writing to the FIFO. Once
   released, the shell runs `script`, a command line the test passes, then exits.
4. `t.Cleanup` releases the shell and kills the desk through its own `*exec.Cmd`.

A command line passed as `script` reports through files only, because the pseudo-terminal turns
line ends into CRLF. Each command in it writes its output and its exit code to its own files in
the temp folder, in this form: `( setsid "$BIN" ... > "$OUT/r1.out" 2>&1; echo $? > "$OUT/r1.code" ) &`,
then `wait` at the end. The `( ... ) &` form is the one whose parent is the desk shell on both
bash and dash; `sh -c` adds a middle process.

`TestDeskHarness` reads `/proc/<pid>/stat` directly and asserts four things. The terminal shell's
`tty_nr` isn't 0. The `setsid` shell's `tty_nr` is 0. A `( cat /proc/self/stat > f ) &` job the
released terminal shell runs reads the shell's PID as its parent. A `( setsid cat
/proc/self/stat > f ) &` job reads `tty_nr` 0 and has the shell as an ancestor, found by walking
parents. Files: `e2e/desk_test.go`. Class `ordinary`.

Test: `TestDeskHarness`, never skipped.

Acceptance check: `go test ./e2e/ -run TestDeskHarness -v` prints `--- PASS` locally. Then open
the draft PR, push, and read the `gate` check by branch (`gh run list --branch
state-rules-rulings`, then `gh run view <id>`). Record the CI verdict. If the check fails, record
`script --version` from the runner's log, and standing answer 34 holds.

### Task 2: reading the process chain (15 minutes)

Outcome: `internal/desk` holds four functions.

- `Read(pid int) (state.Proc, error)` reads one process from `/proc`.
- `Chain() []state.Proc` walks from the calling process up through its parents, as "The guard"
  says. It returns no error: a read that fails ends the chain there.
- `Caller() state.Caller` holds the chain and the session id from `CLAUDE_CODE_SESSION_ID`.
- `Entry(sessionID string, pid int) (state.Desk, error)` builds a desk entry from a PID.

Start times follow "The guard", in one conversion that `Read` and `Entry` share. The first commit
of this task also adds the `Proc` and `Caller` types and `state.SameStart` with its unit test, in
`internal/state`. Files: `internal/desk/desk.go`, `desk_test.go`, `internal/state/types.go`,
`internal/state/ruling.go`, `internal/state/ruling_test.go`. Class `ordinary`.

Test:

- `TestParseStat` on recorded `stat` lines. One has a `)` and spaces inside the command name. Each
  asserts the parent PID, `tty_nr`, and `starttime` against values written out in the test.
- `TestChain` on the test's own process: the chain starts with its own PID and ends at PID 1 or
  at a parent PID of 0.
- `TestEntry` starts a child process, notes `time.Now()` just before and just after `cmd.Start`,
  and asserts that `Entry`'s start time falls inside that window, widened by 2 seconds on each
  side. That checks the conversion against the wall clock, not against itself.
- `TestSameStart`: 1 second apart matches, 3 seconds apart doesn't.

Acceptance check: `go test ./internal/desk/ -v` prints `--- PASS` for the first three, and `go
test ./internal/state/ -run TestSameStart -v` prints `--- PASS`.

### Task 3: the ruling event, the guard, and the writes (25 minutes)

Outcome: `RecordRuling(c Caller, r RulingRequest) Event` and `RegisterDesk(d Desk) Event` follow
"The ruling line" and "The guard". The three guard refusals are exported constants. `Commit`
writes the new ruling line, then the release files, `usage.json`, the forecast line, `desks.json`
when it changed, and `snapshot.json`. One append helper writes both logs. A ruling sets no pass
event; a pass that settling starts still sets one. Files: `internal/state/types.go`, `ruling.go`,
`ruling_test.go`, `commit.go`, `commit_test.go`. Class `ordinary`.

Test: `TestRecordRuling`, a table over hand-built states.

- Each guard test fails alone and names itself.
- An empty session id fails test 1. A dead desk PID fails test 2.
- A start time 1 second off passes, and 3 seconds off fails test 2.
- A caller whose own `TTY` is 0, under a desk process whose `TTY` isn't, passes. A caller whose
  own `TTY` is set, under a desk process whose `TTY` is 0, fails test 3.
- A caller that isn't a desk and names an unknown question gets test 1's refusal.
- Empty words, and words that aren't valid UTF-8.
- An unknown project, round, question, or option, and a request carrying only some of the four.
- An answered question is refused. A ruling with no question is recorded.
- The next id after `r-0001` is `r-0002`. After a hand-written `r-0003` with two lines, it is
  `r-0004`.

`TestRegisterDesk`: a new entry, a replaced one, and each bad field. `TestCommitRuling`: the line
lands with a newline put in first when the last line lacks one; a refused ruling leaves the
folder as it was; a registered desk writes `desks.json`.

Acceptance check: `go test ./internal/state/ -run 'TestRecordRuling|TestRegisterDesk|TestCommitRuling' -v`
prints `--- PASS` for the three, and `go test ./internal/...` passes.

### Task 4: the guess outcome (5 minutes)

Outcome: `guess_outcome` follows "The guess outcome", as constants `held` and `missed` in
`snapshot`. Files: `internal/snapshot/snapshot.go`, `snapshot_test.go`. Class `ordinary`.

Test: `TestSnapshotGuessOutcome` on hand-built states: held, missed, open (null), a carried-over
answer (null), and a pass with no guess (null).

Acceptance check: `go test ./internal/snapshot/ -run TestSnapshotGuessOutcome -v` prints `--- PASS`,
and `go test ./e2e/ -run TestFixtureSnapshotIsCurrent` passes with no golden change.

### Task 5: the verb (10 minutes)

Outcome: `cc-desk ruling record` follows "The verb". It builds the caller with `desk.Caller()`
before `f.commit`, so the reading and the chain both come before the lock. `e2e/broken_test.go`
gains `{"ruling record", ... {"ruling", "record", "go"}}`. That case needs no session id: `Load`
fails on the broken file before the guard runs. Files: `cmd/cc-desk/ruling.go`,
`cmd/cc-desk/root.go`, `e2e/broken_test.go`. Class `ordinary`.

Test: `TestBrokenFile` with the new verb.

Acceptance check: `go test ./e2e/ -run TestBrokenFile -v` prints `--- PASS` with a `ruling record`
case under each broken file, and `cc-desk ruling record --help` prints the usage line.

### Task 6: done item 4 end to end (25 minutes)

Outcome: `e2e/ruling_test.go` `TestRuling` runs six cases. Each case works on a fresh copy of
`e2e/testdata/order`, plus two files the test writes:

- `desks.json`, from `desk.Entry` for the desk the case started;
- `rulings.jsonl`, one line with id `r-0001` and no trailing newline, so the refusals compare real
  bytes and the success case checks the newline repair. A new ruling's id is `r-0002`.

A ruling runs inside a desk through `startDesk`, as `setsid "$BIN" ruling record ...` in the
`( ... ) &` form, so the ruling verb itself has no terminal, as a desk session's commands don't.
The words and flags travel in environment variables that the desk's command line expands in
double quotes, so the words reach the argument list unchanged. The words are `Use the "short
form" for the café list.`

| Case | Desk | Session id | Request | Expect |
| --- | --- | --- | --- | --- |
| `answers_and_frees` | terminal | the desk's | north-q1, `yes`, the words | exit 0; `rulings.jsonl` gains exactly one line, `r-0002`; its decoded words equal the words byte for byte; in the snapshot, north-q1 reads `answered: true` and `answer_kind: "ruling"`, `rulings` holds `r-0002` with `question: "north-q1"` and `option: "yes"`, and `n-big` reads `free: true` and `why_not: []`; the release file's north-q1 `answer.ruling` is `r-0002` |
| `guess_held` | terminal | the desk's | after `pass start north n-guess --guess north-q2` at `at(10)`, which reads total 200: north-q2, `yes` | `n-guess` reads `guess_outcome: "held"`, `open_guesses` is `[]`, `guess_minutes_total` drops from 200 to 0 |
| `guess_missed` | terminal | the desk's | the same start, then north-q2, `no` | `guess_outcome: "missed"`, total 0 |
| `unknown_session` | terminal | `not-a-desk` | north-q1, `yes` | exit 2, stderr holds test 1's refusal, folder byte for byte as before |
| `not_under_desk` | terminal, registered | the desk's | run by the test directly, outside the desk | exit 2, test 2's refusal, folder unchanged |
| `no_terminal` | `setsid`, registered | the desk's | north-q1, `yes` | exit 2, test 3's refusal, folder unchanged |

Files: `e2e/ruling_test.go`. Class `ordinary`.

Acceptance check: `go test ./e2e/ -run TestRuling -v` prints `--- PASS` for `TestRuling` and each of
the six cases.

### Task 7: done item 6 end to end (15 minutes)

Outcome: a fixture `e2e/testdata/writers/` holds releases `east` and `west`. Each has one designed
round that needs nothing from Geoff (with its `no_need_reason`), seven building passes, and three
unanswered questions that no pass rests on. The fixture has no `usage.json`, so every verb takes
a reading. Nothing waits, so settling starts nothing. Files:
`e2e/testdata/writers/releases/east.json`, `west.json`, `e2e/testdata/README.md`,
`e2e/writers_test.go`. Class `ordinary`.

`TestTwoWriters` launches twenty verbs at once with a 15-second stand-in that appends `date
+%s.%N` to a log on each call:

1. It starts one terminal desk, registered in `desks.json`.
2. It starts seven `pass finish` and seven `pass stop` verbs across both releases, each as its own
   process, then releases the desk at once. The desk runs six `ruling record` verbs, one per
   question, each in the `( setsid ... ) &` form with its own result files, then `wait`.
3. Every verb runs at the same `--now`.

Test:

- Every verb exits 0.
- The stand-in log holds 20 call times, all within 5 seconds of each other, and the twenty verbs
  finish under 40 seconds after the first launch. That shows the reads ran at once, outside the
  lock.
- Both release files load. Every finished pass reads `done`, every stopped pass `stopped`, and
  every question an answer naming its ruling.
- `rulings.jsonl` holds six lines with six distinct ids.
- `forecasts.jsonl` holds 14 lines, one per finish and stop.
- The test reads `snapshot.json` into memory and the stand-in log's line count, then runs the
  snapshot verb at the same `--now`. The bytes it read equal the verb's output, and the log still
  holds 20 lines, because the verb uses the cached reading.

Acceptance check: `go test ./e2e/ -run TestTwoWriters -v` prints `--- PASS`.

### Task 8: done items 1 to 8, the pull request, and close (10 minutes)

1. Run each check in "Done items 1 to 8 on the branch" on the final head.
2. The live read, once, outside CI, from the worktree, each command alone, as pass 3 ran it:

   ```bash
   go build -o "$CLAUDE_JOB_DIR/tmp/live/cc-desk" ./cmd/cc-desk
   "$CLAUDE_JOB_DIR/tmp/live/cc-desk" --state-dir "$CLAUDE_JOB_DIR/tmp/live/state" usage --raw "$CLAUDE_JOB_DIR/tmp/live/raw.json"
   python3 -I -c 'import json,re,sys; r=json.load(open(sys.argv[1])); m=re.search(r"Current week \(all models\): (\d+)% used", r["result"]); print("line percent:", m.group(1)); print("total_cost_usd:", r["total_cost_usd"]); print("num_turns:", r["num_turns"])' "$CLAUDE_JOB_DIR/tmp/live/raw.json"
   ```

   Use a fresh `live` folder; never point `--state-dir` at a real state folder. The PR body
   quotes the verb's line and the three printed lines, and nothing else from the raw result. A
   difference goes by standing answer 30.
3. Update `README.md`'s verb sentence to name rulings, and `docs/STATUS.md` with the pass's
   lines. File two follow-ups under `docs/follow-ups/`: the orphan ruling a retry after a crash
   leaves (see "The ruling line"), and a desk that lives through a wall-clock step of more than 2
   seconds. `btime` moves when the clock is stepped, so such a desk fails test 2; storing the raw
   `starttime` ticks with the boot id would fix it, in the Desk round.
4. Get one whole-branch `standards:diff-reviewer` read and fold it once. A delta review after the
   fold has its Low findings folded once and checked by your own diff read and CI, or filed as
   follow-ups; it never starts another round.
5. After the last edit, run `bash scripts/gate.sh` bare, push, and watch with `gh pr checks <n>
   --watch` alone. Name the CI verdict on that head in the PR body.

Acceptance check: `bash scripts/gate.sh` prints `gate: ok` on the final head; the PR's `gate`
check on that head is `pass`; the PR body quotes each item's output.

The tasks sum to 115 minutes: 10, 15, 25, 5, 10, 25, 15, and 10.

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
  never skip. Before pushing the fallback, confirm the runner has `python3` from the failed run's
  log or a one-line `command -v python3` step; `ubuntu-latest` carries it.
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
- **40. Item 4 or 8 can't be reached.** If both terminal wrappers fail on the runner and the
  advisor finds no equivalent check, tasks 5, 7, and 8 still run. The PR closes with the bar
  unmet, named under `## Calls for Geoff` with every probe's output, and no test skips.

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
  spells out each string and test.
- One lane, no pair. A pair would cut the path by 20 minutes, but none has run in this round,
  and nothing shows that a background Agent dispatch survives an unattended run. The minutes come
  from pass 3's measured pace instead, which leaves the 115-minute path a wide margin.
- `Caller()` never fails the verb. A process that exits mid-walk ends the chain, and the guard
  refuses with test 2 inside the event, so a broken state file is still named first.
- The guard runs before every other check of the request, so a caller that isn't a desk always
  gets a guard refusal.
- A retry after a crash between the ruling write and the release write leaves an orphan ruling
  line. The state still loads; task 8 files it as a follow-up.
- [review] One `standards:plan-reviewer` read (Opus, medium), 2026-11-17. Two blockers, six
  majors, and twelve minors, all folded. The blockers: the pair's task 2 needed `SameStart`
  before task 3 had written it, and the 105-minute path rested on a pair dispatch nobody had
  probed; both are gone with the single lane. The majors: the rulings now run under `setsid`
  inside the desk, so a guard that read the caller's terminal would fail (and a unit row checks
  it); `TestEntry` checks start times against the wall clock; `TestTwoWriters` reads
  `snapshot.json` before the snapshot verb runs, and checks the 20 call times fall within 5
  seconds; the snapshot check reads `answered`, `answer_kind`, and `rulings[]`, adding no field;
  and `Caller()` never fails the verb. The minors: the `( ... ) &` job form, deadlines and
  `WaitDelay`, no session id in the broken-file case, a seeded `rulings.jsonl`, questions no pass
  rests on, the check order, the orphan follow-up, the live commands, the delta-review floor,
  standing answer 40, the base at `92dac54`, and the runner's `python3`.

## Executor decisions

## Rulings

Carried from pass 3: Geoff, 2026-11-18, terminal session, via conductor-cc-desk, on a resumed
stopped pass and the limits: "Yes as built." Carried from pass 2: Geoff, 2026-11-18, on the round
that needs Geoff with no design: "Those both work." Geoff, 2026-11-18, on wording: "In general
you can look at survey prose standard for that sort of thing."
