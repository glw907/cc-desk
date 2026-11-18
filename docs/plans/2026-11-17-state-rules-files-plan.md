# State files and fixtures: pass plan

Initiative: cc-desk
Round: state-rules
Pass: state-rules-files

Class `runner`. Planned 90 minutes; the path below is 123. Design:
`docs/specs/2026-11-17-state-and-rules-design.md` (approved 2026-11-17), sections "The state
files", "What the command does in this round", "The snapshot", "When something goes wrong", and
the first row of "The passes". Round plan: `docs/plans/2026-11-17-state-rules-round.md`. Where
the round design is silent, `docs/specs/2026-11-17-cc-desk-design.md` holds. Writing a person
reads (the PR body, STATUS lines, follow-ups) follows `~/.claude/docs/voice/written-for-people.md`.

Worktree `/var/home/glw907/Projects/cc-desk/.claude/worktrees/state-rules-files`, branch
`state-rules-files`, cut from `origin/main` at `f91f9c2`. PR base: `main`. Push: `glw907/cc-desk`.
This is the round's first pass, so no earlier PR carries `[advisor]` lines to build on.

## Pre-flight

No step in this pass needs Geoff. Every write is inside the worktree or a test's `t.TempDir()`.

- Permissions checked: `~/.local/state/claude-denials.jsonl`, 129 rows, 2026-11-01T00:55:05Z to
  2026-11-18T03:06:17Z. Classes and rows: Classifier unavailable 47; Destructive or Unwanted
  Actions 14; Self-Modification 11; Interfere With Workloads 11; server-side classifier, no
  reason 10; Undermining Oversight 4; Unauthorized Access 4; Merge Without Review 3; Auto-Mode
  Bypass 3; Irreversible Local Destruction 2; Git Destructive 2; Create Public Surface 2;
  Unverifiable Deletion Scope 2; Sensitive-Source Provenance 2; Security Weaken 2; Production
  Deploy 1; Secret-Store Writes 1; DNS/Domain/Cert Changes 1; Bot Check Circumvention 1; Code from
  External 1; Shared Scratch Sweep 1; Safety Bypass Flag 1; Out-of-Place Publication 1; PII Data
  Handling 1; Modify Shared Resources 1. No row refuses `go build`, `go test`, `go vet`, `gofmt`,
  `bash scripts/gate.sh`, a push of a pass branch, or `gh pr create`. The rows that touch this
  pass's commands, with the alternative for each, are in "Refused commands and their
  alternatives" below.
- Reviewed: one `standards:plan-reviewer` read of this plan, findings folded (see `## Decisions`).
- Evidence: the mechanism table and the recorded searches below. Each acceptance check was run on
  today's code before any code was written; the outputs are under `## Acceptance`.

### Refused commands and their alternatives

| Refused in the log | This pass does instead |
| --- | --- |
| `go run …@latest` of an outside tool (Code from External, 2026-11-09) | `go get github.com/spf13/cobra@v1.10.2`, a pinned version. If refused, the offline route in the mechanism table. A second refusal is a step only Geoff can do: write the stop |
| `kill <pid>` before `go test`, and `kill` or `pkill` in loops (2026-11-09, 2026-11-15) | Stop a background command with `TaskStop`. Tests end their own child processes through `exec.Cmd` |
| `rm -rf /tmp/tmp.*` sweeps (Shared Scratch Sweep) | Scratch lives in `t.TempDir()` or `$CLAUDE_JOB_DIR/tmp`, which clean themselves. No manual sweep |
| `git push origin HEAD:main` | Push only `state-rules-files` |
| `git stash drop`, refused twice (2026-11-17) | No stash. A temporary commit sets work aside |
| A chained `claude stop` | Run `claude stop <id>` alone, if ever needed |
| A write to a Claude `settings.json` (Self-Modification, 2026-11-18) | No settings edit. A task that needs one parks (round-run standing answer 3) |
| A classifier outage | Wait 60 seconds and retry once, then flag the task and go on |
| `gh pr merge` (Merge Without Review) | The pass merges nothing. It ends with the PR ready |

## Mechanism table

| Mechanism | Probe or source |
| --- | --- |
| Go version | `go version` printed `go1.27.4`; `.github/workflows/gate.yml` pins `go-version: 1.27.4` |
| No dependency yet | `go.mod` holds `module github.com/glw907/cc-desk` and `go 1.27`; no `go.sum` |
| cobra offline | `GOPROXY=off GOFLAGS=-mod=mod go get github.com/spf13/cobra@v1.10.2`, then `GOPROXY=off go mod tidy`, in a scratch copy: `require github.com/spf13/cobra v1.10.2` with `mousetrap v1.1.0` and `pflag v1.0.9` indirect, a 10-line `go.sum` |
| Strict JSON | A `json.Decoder` with `DisallowUnknownFields()` on `{"extra":1}` returned `json: unknown field "extra"`. The plan reviewer's probe on Go 1.27.4: `dec.More()` is false after `{"a":1}}` and `{"a":1}]`, so trailing data is caught only by a second `Decode` that must return `io.EOF`. A 0-byte file returns `EOF` on the first `Decode`. Missing fields decode silently, so required fields need a check after decode |
| The lock | `syscall.Flock(fd, LOCK_EX|LOCK_NB)` on a second descriptor of the same file, in the same process, returned `EWOULDBLOCK` while the first held it, on btrfs (`/var/home`) and tmpfs (`/tmp`). Across processes: with `flock <dir>/lock sleep 3` holding it, a separate Go binary's `LOCK_NB` flock printed `resource temporarily unavailable`, and `<nil>` once the holder exited |
| Skip events | `go test -json ./...` on a module with `t.Run("inner", … t.Skip)` and a package with no test files emitted `{"Action":"skip","Package":"p/b"}` (no `Test` key) and a skip with `"Test":"TestS/inner"`. `go test` exited 0. `jq -r 'select(.Action=="skip" and .Test!=null)'` printed `p/a TestS/inner` only. Cached runs replay the skip events |
| Build failures under `-json` | The plan reviewer's probe: a compile error made `go test -json` exit 1 and emit `build-output` and `build-fail` events that carry `ImportPath`, with no `Package` or `Test`. A filter on failed tests alone would show a bare FAIL |
| jq on CI | `gate.yml` runs on `ubuntu-latest`, which is ubuntu-24.04 today; its runner images list `jq 1.7.1` (actions/runner-images release notes). Local: `/home/linuxbrew/.linuxbrew/bin/jq`. The gate checks `command -v jq` as it checks vale, so a runner without it fails loudly |
| Gate on a skip today | In a scratch clone, `cmd/cc-desk/skip_test.go` with `t.Skip("probe")`: `bash scripts/gate.sh` printed `gate: ok`, exit 0 |
| Vale in a scratch copy | `.vale.ini` sets `StylesPath = styles`, and `styles/` is committed (8 files), so a copy runs with no `vale sync` |
| Gate with no docs | `vale --minAlertLevel=error --output=line` with no file argument printed its help and exited 0 (the plan reviewer's probe, and mine). The scratch copy still carries one `docs/*.md` so its vale step lints a real file; "This folder holds a test." passed Survey at error level |
| CI | `.github/workflows/gate.yml` runs `bash scripts/gate.sh` after setup-go and a pinned vale. The skip check rides inside `gate.sh`, so the workflow needs no change. The last three runs on `main` passed, latest 44242272799 |

## Recorded searches

`git grep -n <term> -- ':!docs/specs/2026-11-17-cc-desk-mockups'`, run in the worktree:

| Search | Hits | Each hit |
| --- | --- | --- |
| `go test` | 6 | `scripts/gate.sh:2,18,19`: changed (task 7). The rest are docs describing the gate: left |
| `0.0.0-prove` | 1 | `cmd/cc-desk/main.go:11`: the version moves into the root command (task 1) |
| `XDG_STATE_HOME` | 1 | Design line 86: the source of the default folder |
| `local/state/cc-desk` | 1 | Design line 86: same |
| `state-dir` | 0 | The flag is new |
| `go.sum` | 1 | Round plan line 66, and its assumption row at line 107, say the repo has no dependency. Both go stale after task 1. Left unedited; the PR lists it under `## Planning misses` as `[stale-fact]` |
| `scripts/gate.sh` | 26 | `gate.yml:38` runs it (no change). STATUS, plans, prompts, findings, design line 433: describe it, left. `spikes/4-permissions/prompt.txt`, `settings.json`, `spikes/owned.tsv`: refer to `scratch/scripts/gate.sh` in another repo, dismissed |
| `Nothing is built` | 1 | `README.md:6`: changed to say what the command does now (task 8) |

`git ls-files '*.go'` lists two files, `cmd/cc-desk/main.go` and `cmd/cc-desk/main_test.go`.

## File list

| File | Task | Change |
| --- | --- | --- |
| `go.mod`, `go.sum` | 1 | cobra v1.10.2 |
| `cmd/cc-desk/main.go` | 1 | Build the root command, run it, map the error to its exit code |
| `cmd/cc-desk/root.go` | 1 | Root command, persistent `--state-dir` and hidden `--now`, version |
| `cmd/cc-desk/main_test.go` | 1 | Replaces the `run` test with a version test on the root command |
| `cmd/cc-desk/release.go`, `round.go`, `pass.go`, `snapshot.go` | 5 | One file per verb group |
| `internal/state/types.go` | 2 | Every state file's type, the five pass states, the five stop reasons |
| `internal/state/read.go` | 2 | The strict reader for every file |
| `internal/state/check.go` | 2 | Required fields, references, loops, format versions |
| `internal/state/events.go` | 5 | Add release, record design, start, finish, stop |
| `internal/state/commit.go` | 4 | Lock, load, check, apply, settle, write, snapshot |
| `internal/state/lock.go` | 3 | The flock with its 10-second wait, under `//go:build unix` |
| `internal/atomicfile/atomicfile.go` | 3 | Temp file, sync, rename |
| `internal/snapshot/snapshot.go` | 4 | The snapshot type and one `Build` that fills pass 1's fields. Passes 2 to 4 add their builders in their own files (`estimate.go`, `order.go`, `rulings.go`), each called from `Build` |
| `*_test.go` beside each | 2 to 5 | Unit tests |
| `e2e/main_test.go` | 6 | `TestMain` builds the binary once |
| `e2e/broken_test.go`, `events_test.go`, `lock_test.go`, `gate_test.go` | 6, 7 | Items 5, A, B, C, D |
| `e2e/testdata/state/` | 6 | The base fixture with the worked case |
| `e2e/testdata/inputs/` | 6 | Release and design inputs for the verbs |
| `scripts/gate.sh` | 7 | Fails on a skipped test |
| `README.md` | 8 | Line 6 |
| `docs/STATUS.md` | 8 | Below the round-report block only |

## The command this pass builds

The design leaves names, layout, and exit codes to the builder. These are fixed here so passes 2
to 4 add files and fields without reshaping them.

| Verb | Arguments |
| --- | --- |
| `cc-desk release add <file>` | Installs the release in `<file>` as `releases/<project>.json` |
| `cc-desk round design <project> <round> <file>` | Adds the passes and questions in `<file>` to the round and marks its design written |
| `cc-desk pass start <project> <pass>` | Records the start. In this pass there is no scheduling rule |
| `cc-desk pass finish <project> <pass>` | Records the finish and the actual minutes |
| `cc-desk pass stop <project> <pass> --reason <reason>` | Records the stop with one of the five reasons |
| `cc-desk snapshot` | Prints the snapshot as JSON to stdout |

- Persistent `--state-dir`: default `$XDG_STATE_HOME/cc-desk` when that is set, else
  `~/.local/state/cc-desk`. A verb creates the folder (0755) when it is missing.
- Hidden persistent `--now`, RFC 3339: the one clock. Absent, it is the system time. Every verb
  reads time only through it.
- Pass names are unique within a release, and so are question ids. A pass is addressed by
  project and pass name.
- The root command reports the version through cobra's `Version`; `cc-desk --version` prints
  `cc-desk version 0.1.0`.
- Later passes add `--guess` on start, and verbs `free`, `usage`, and `ruling`.

| Exit code | Meaning |
| --- | --- |
| 0 | Done |
| 1 | Error: a broken file, a name that doesn't exist, the lock wait ran out, a bad input, bad usage |
| 2 | Refused: finishing twice or before a start, a state change the pass's state doesn't allow, a release already live for the project, a design already written |

A refusal and an error each print one line to stderr. A missing name prints the name.

### Every verb's run

1. Take the lock on `<state>/lock` (an empty file), waiting up to 10 seconds of real time. On
   timeout, exit 1 and write nothing, not even `snapshot-error.json`.
2. Load every file strictly, in this fixed order: `releases/*.json` by sorted name,
   `releases/done/*.json` by sorted name, `desks.json`, `settings.json`, `usage.json`,
   `rulings.jsonl`, `forecasts.jsonl`. An absent optional file is empty; an absent
   `settings.json` means a usage stop of 100.
3. Run the cross-file checks, release files in the same order.
4. On a load or check failure, the first failure wins: write `snapshot-error.json` with that file
   (relative to the state folder) and the message, leave every other file as it was, and exit 1
   with the same message.
5. The load and check are clean, so remove `snapshot-error.json` if it exists. This is the one
   write a refusal may make, and it makes "restoring the file and running any verb removes it"
   true for every verb.
6. Apply the verb's event. A refusal writes nothing more and exits 2.
7. Settle. In this pass the step is an empty function in the one commit path; pass 3 fills it.
8. Write each changed file atomically and rewrite `snapshot.json`. The snapshot verb runs every
   step too.

A JSON file that is empty, or holds anything after its one value, is a syntax error naming the
file: decode one value, then a second `Decode` must return `io.EOF`. In a JSONL file each line
is one value under the same rule; a blank line is a syntax error, and the final newline is
optional. An error names the line number for a JSONL file.

### State files

Every JSON file and every JSONL line carries `"format": 1`; another value is an error naming
the file. Times are RFC 3339 in UTC. The worked case's 2:00 pm is 2026-11-20T14:00:00-09:00,
stored as 2026-11-20T23:00:00Z.

- Release file: project, name, repo; rounds (name, summary, needs-Geoff mark, the reason when it
  needs nothing, rounds before it, planned minutes while it has no passes, design written);
  passes (name, summary, round, planned minutes, passes before it, questions it rests on, state,
  started and finished times, stop intervals, stop reason, actual minutes, and the optional guess
  record that pass 3 writes); questions (id, text, options, recommended option, answer). An
  answer is a ruling reference or a carried-over answer (words, date, source).
- The five pass states: `not_started`, `waiting`, `building`, `stopped`, `done`. `waiting`
  carries an optional limit field that pass 3 writes.
- The five stop reasons: `usage`, `irreversible`, `secret`, `auth_data`, `scope`.
- `rulings.jsonl`: one ruling per line. The id is an opaque non-empty string in this pass; pass 4
  sets its format and the guard.
- `forecasts.jsonl`, `desks.json`, `usage.json`: their types land now. Nothing in this pass
  writes them.
- `settings.json`: `format` and `usage_stop_percent`, an integer from 1 to 100.

The strict checks name the file for each of these: a syntax error, an unknown field, a missing
required field, a reference to a pass or question that doesn't exist, an answer that points at
no ruling, and a loop in what rests on what, across passes and across rounds.

### Pass events

| Verb | From | To | Records |
| --- | --- | --- | --- |
| start | `not_started` | `building` | Start time |
| start | `stopped` | `building` | Closes the open stop interval (the resume) |
| finish | `building` | `done` | Finish time; actual minutes, the time between start and finish less stopped time, rounded to the nearest minute |
| stop | `building` | `stopped` | Opens a stop interval; the reason |

Any other start, finish, or stop is refused with exit 2. No verb ends a release in this round.
A design input carries no answers, so a carried-over answer comes in only through `release add`.

### The snapshot

Every field is always present. An empty list encodes as `[]`, an unset time or number as
`null`. An object a later pass fills is present as an object whose members are `null` or `[]`,
never as `null` itself.

| Field | Filled by |
| --- | --- |
| `format`, `generated_at` (the clock) | 1 |
| `releases`: every round and pass with name, state, planned and actual minutes, event times, and what it needs from Geoff (its questions, or nothing with the reason, or not designed yet) | 1 |
| Per question: `answered` (true when an answer is present) and the answer's kind | 1 |
| `rulings`: every ruling | 1 |
| Per release: `pace`, `pace_low`, `pace_high`, `rough` | 2 (`null`) |
| Per pass and round: `minutes_before`, `needed_time`; per not-designed round: `forecast_time` | 2 (`null`) |
| `next_session`: an object with state, time, range, rough mark, question count, rounds and projects, rounds not designed yet | 2 (members `null` or `[]`) |
| `waits_for_geoff`: one entry per round with needed written questions | 2 (`[]`) |
| Per pass: `free` and `why_not` | 3 (`null` and `[]`) |
| `waiting`: each pass waiting on the two-pass limit or usage, with the reset text | 3 (`[]`) |
| `open_guesses` with question and minutes, `guess_minutes_total`, `guess_minutes_limit` 300 | 3 (`[]`, 0, 300) |
| Per guessed pass: `guess_outcome`, held or missed | 4 (`null`) |

## Acceptance

Every check runs the built `cc-desk` binary against a copy of the fixture state folder in
`t.TempDir()`, through `go test ./e2e/...`. Byte-for-byte comparisons leave out `<state>/lock`.
The base fixture carries a `snapshot.json`, and every "unchanged" check first asserts that file
exists and is not empty, so an absent file can't pass as unchanged.

| Item, word for word | Task | Check |
| --- | --- | --- |
| **A broken file.** Each of these makes every verb exit with an error that names the file: a syntax error, an unknown field, a missing required field, a reference to a pass that doesn't exist, an answer that points at no ruling, and a loop in what rests on what, each in a release file; then a syntax error in `desks.json`, in `settings.json`, and in one line of `rulings.jsonl`. Each time `snapshot.json` is unchanged and `snapshot-error.json` exists. Restoring the file and running any verb removes it. | 6 | `go test ./e2e/ -run TestBrokenFile -v`: one table of the nine cases, each made by changing a copy of the base fixture, run against every verb in a second table that later passes append to |
| **The gate.** `bash scripts/gate.sh` passes, and so does CI on the pull request. The gate fails if any test was skipped, so a runner without a terminal can't pass item 4 by skipping it. | 7, 8 | `bash scripts/gate.sh` run bare; the PR's `gate` check; item D |
| A. Events. On a fixture, adding a release and printing the snapshot shows its rounds and passes with their states. Start, finish, and stop each change the pass's state and record the time the test's clock gave | 6 | `go test ./e2e/ -run TestEvents -v` |
| B. Bad requests. Finishing a pass twice, or before it started, is refused, and the state folder is byte for byte what it was. A verb that names a project, round, pass, or question that doesn't exist exits with that name | 6 | `go test ./e2e/ -run TestBadRequests -v`. The question case is a `round design` input whose pass rests on an undefined question id; passes 3 and 4 add their own |
| C. The lock. While a test holds the state lock, a second command exits with an error after 10 seconds and writes nothing | 6 | `go test ./e2e/ -run TestLock -v`: the test holds the flock, runs `pass start`, and asserts exit 1, elapsed at least 10 and under 15 seconds, and the folder unchanged |
| D. Skipped tests. `bash scripts/gate.sh` fails when any Go test reports a skip. The test for this adds a skipping test in a scratch copy and shows the gate's exit code | 7 | `go test ./e2e/ -run TestGateFailsOnSkip -v` |

Items A to D sit in a table in the round plan, with a column mark between the label and the
text. Here a period stands in for that mark; the words are unchanged.

What today's code printed, before any change, with the binary built at
`$CLAUDE_JOB_DIR/tmp/pre/cc-desk` and a broken `releases/alpha.json` (`{"format":1,`):

- Items 5, A, B, C: `snapshot`, `--state-dir <state> snapshot`, `pass start alpha p1`,
  `pass finish alpha p1`, and `release add <file>` each printed `cc-desk 0.0.0-prove` and exited
  0. No `snapshot-error.json` was written; the folder held only `releases`. The binary ignores
  every argument, so each item fails today.
- Item D: a scratch clone with a skipping test printed `gate: ok`, exit 0. Fails today.
- Item 8: `bash scripts/gate.sh` in the worktree exited 0, and the last CI run on `main` passed.
  The skip half fails today, per item D.

## Tasks

Each task commits on its own, runs `bash scripts/gate.sh` bare before the commit, and is pushed.
No task holds code; the sections above fix the behavior.

### Task 1: cobra and the root command (10 minutes)

- Outcome: `cc-desk --version` prints `cc-desk version 0.1.0`; `--state-dir` and the hidden `--now`
  parse on every subcommand; `main()` is the only printer and maps errors to codes 1 and 2.
- Files: `go.mod`, `go.sum`, `cmd/cc-desk/main.go`, `cmd/cc-desk/root.go`, `cmd/cc-desk/main_test.go`.
- Class: `runner`.
- Check: `go build ./cmd/cc-desk && ./cc-desk --version` prints the line.
- Test: `main_test.go` runs a fresh root command with `--version` and reads the output; a test
  maps a refusal error to 2 and any other to 1.

### Task 2: types and the strict reader (30 minutes)

- Outcome: one load call reads every state file under a folder and returns the state or an error
  naming the file relative to the folder.
- Files: `internal/state/types.go`, `read.go`, `check.go`, and their tests.
- Class: `runner`.
- Check: `go test ./internal/state/ -run 'TestRead|TestCheck' -v`.
- Test: one table per file kind covering each strict failure in "State files", an empty file,
  trailing data, and a blank JSONL line; absent optional files; an empty folder loads as empty;
  with two broken files, the error names the first in load order.

### Task 3: atomic writer and the lock (8 minutes)

- Outcome: `atomicfile.WriteFile` writes temp, syncs, renames; the lock waits up to 10 seconds,
  then returns an error that names the state folder.
- Files: `internal/atomicfile/atomicfile.go`, `internal/state/lock.go`, and their tests.
- Class: `runner`.
- Check: `go test ./internal/atomicfile/ ./internal/state/ -run 'TestWrite|TestLock' -v`.
- Test: a write leaves no temp file and the target holds the bytes; a held lock times out. The
  unit test takes the wait as a parameter so it runs in under a second; the e2e test in task 6
  holds the full 10 seconds.

### Task 4: the commit path and the snapshot (15 minutes)

- Outcome: one function runs "Every verb's run" for any event; the snapshot type has every field
  in the table above; `snapshot-error.json` is written and removed as set out.
- Files: `internal/state/commit.go`, `internal/snapshot/snapshot.go`, and their tests.
- Class: `runner`.
- Check: `go test ./internal/... -v -run 'TestCommit|TestSnapshot'`.
- Test: a golden of the snapshot of an empty state shows every field in the table, with `[]`,
  `null`, `0`, and `300`, and `next_session` as an object; a load failure leaves `snapshot.json`
  as it was and writes the error file; a refusal after a clean load removes the error file and
  writes nothing else.

### Task 5: the verbs (15 minutes)

- Outcome: the six verbs in "The command this pass builds", with the pass events table.
- Files: `internal/state/events.go`, `cmd/cc-desk/release.go`, `round.go`, `pass.go`,
  `snapshot.go`, and their tests.
- Class: `runner`.
- Check: `go test ./internal/state/ ./cmd/... -v`.
- Test: a table of every from-state against every verb; resume after stop leaves stopped time out
  of the actual minutes; `release add` refuses a live project; `round design` refuses a written
  design and names an undefined question.

### Task 6: fixtures and the e2e tests (30 minutes)

- Outcome: `e2e/testdata/state/` holds the base fixture, and items 5, A, B, and C pass.
- Files: `e2e/main_test.go`, `e2e/broken_test.go`, `events_test.go`, `lock_test.go`,
  `e2e/testdata/state/`, `e2e/testdata/inputs/`.
- Class: `runner`.
- Check: `go test ./e2e/ -v`.
- Test: the tests are the acceptance checks. The base fixture holds:
  - `releases/alpha.json`, the worked case: twelve finished passes, each planned at 100, finish
    times oldest first, actual minutes 300, 300, 150, 80, 120, 120, 120, 120, 120, 120, 125,
    125; then two unfinished passes planned at 100 and 200, and after them the pass the worked
    case's question rests on. The worked clock is 2026-11-20T23:00:00Z.
  - `releases/beta.json`: two finished passes, a pass needed later, a round marked as needing
    Geoff with no design, and one carried-over answer.
  - One question answered by a ruling present in `rulings.jsonl`.
  - `desks.json`, `settings.json`, `usage.json`, and `forecasts.jsonl`, each valid.
  - `snapshot.json`, as the snapshot verb writes it for this fixture at the worked clock.
  A `README` line in `e2e/testdata/` says the worked case is for passes 2 to 4 and gives its
  numbers: a pace of 1.2 with a range of 0.8 to 1.5 from the last ten.

### Task 7: the gate fails on a skip (12 minutes, beside tasks 2 to 6)

- Outcome: `scripts/gate.sh` checks `command -v jq` as it checks vale. It writes
  `go test -json ./...` to a temp file and captures the status with `|| status=$?`, so
  `set -euo pipefail` can't abort it or mask the status. It prints the `Output` of every
  `build-output` event and of every failing package or test, then names each test whose event is
  a skip with a `Test` key. It exits non-zero when the status or the skip scan fails. A package
  with no test files doesn't fail it.
- Files: `scripts/gate.sh`, `e2e/gate_test.go`.
- Class: `runner`.
- Check: `go test ./e2e/ -run TestGateFailsOnSkip -v`, and `bash scripts/gate.sh` bare.
- Test: the e2e test builds a scratch module in `t.TempDir()`, outside the worktree: a copy of
  the working tree's `scripts/gate.sh`, `.vale.ini`, and `styles/`, a `go.mod`, one small
  `docs/*.md`, and a package with a skipping test, then `git init` and `git add`. It runs the
  gate and asserts a non-zero exit and the skipped test's name in the output. A control copy
  with no skip passes, and a copy with a compile error fails and shows the compiler's message.
  The scratch module doesn't hold `gate_test.go`, so the gate can't recurse.

### Task 8: close (15 minutes)

- Outcome: the PR is ready, with review accepted and CI green.
- Files: `README.md`, `docs/STATUS.md` (below the round-report block), any follow-up under
  `docs/follow-ups/`.
- Class: `runner`.
- Check: the PR body's evidence lines and `## Acceptance`; the `gate` check on the last commit.
- Test: none of its own.

Path: tasks 1, 2, 3, 4, 5, 6, and 8 in order, 123 minutes. Task 7 runs beside them after task 1.
The sum over every task is 135. Both are under the 180-minute park line.

## Standing answers

The twelve in `round-run.md` hold, with the round plan's changes (4 replaced, 1 and 9 changed)
and its answers 13 to 17. This pass adds:

- **18. cobra can't be fetched.** Use the offline route in the mechanism table, which reads the
  local module cache with `GOPROXY=off GOFLAGS=-mod=mod` (cobra v1.10.2 and `pflag` v1.0.9 are
  there today). Only if that is refused too, or the cache is empty, is it a step only Geoff can
  do: write the stop under `## Stop` in `docs/STATUS.md` on the branch and end.
- **19. A test is tempted to skip.** No Go test may skip, conditionally or not. This overrides
  the go-conventions allowance. A test that can't run on a host fails there.
- **20. The lock test is slow.** Item C waits the full 10 seconds. The gate takes that cost.
- **21. A field name in the design table reads wrong once coded.** The names are the builder's;
  change it, keep the snapshot's field table in step, and list it in `## Decisions`.
- **22. The path runs past 180 minutes.** Finish the task in flight, push, and write the stop.

## Decisions

- [advisor] Exit codes 0, 1, 2 depart from go-conventions' single exit 1, because the design asks
  for a distinct refusal code.
- [advisor] Start on a stopped pass is the resume, and actual minutes leave out stopped time.
- [advisor] Item 5's cases come from one table of mutations to a copy of the base fixture, run
  against a verb table that later passes append to.
- [advisor] The lock timeout writes nothing, not even `snapshot-error.json`: the files may be fine.
- [advisor] Byte-for-byte comparisons leave out `<state>/lock`.
- [advisor] The skip check parses `go test -json` with jq and fails only on events with a `Test`
  key, so a package with no test files passes.
- [advisor] cobra v1.10.2 by pinned `go get`, with the offline route as the first alternative.
- The snapshot verb writes `snapshot.json` like any other verb.
- A clean load removes `snapshot-error.json` before the event runs, so a refusal clears it too
  (plan review finding 1).
- `gate.yml` needs no change; the skip check lives in `gate.sh`.
- The plan review's twelve findings are folded: the error-file order, a fixture `snapshot.json`,
  the strict JSON rule, the gate's build output and jq check, the load order and `next_session`
  form, the per-task CI line and hand-over wording in the prompt, the cross-process lock probe
  and 15-second bound, the vale row, the cobra cache route, the snapshot placeholders and
  separate builder files, 30-minute tasks 2 and 6, and the `go.sum` stale fact.
