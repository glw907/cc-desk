# The check: pass plan

Initiative: cc-desk
Round: sessions
Pass: sessions-check

Class `ordinary`. Planned 90 minutes; the path below is 158. Design:
`docs/specs/2026-11-17-sessions-design.md` (approved 2026-11-17), sections "What the command
records", "The check", "When something goes wrong", "Left to the builder", and the first row of
"The passes". Round plan: `docs/plans/2026-11-17-sessions-round.md`. The State and rules design,
`docs/specs/2026-11-17-state-and-rules-design.md`, holds where the round design is silent. Writing
a person reads (the PR body, STATUS lines, follow-ups) follows
`~/.claude/docs/voice/written-for-people.md`, loaded through the `writing-voice` skill.

Worktree `/var/home/glw907/Projects/cc-desk/.claude/worktrees/sessions-check`, branch
`sessions-check`, cut from `origin/main` at `4e7e78e`. PR base: `main`. Push: `glw907/cc-desk`.
This is the round's first pass. The last round's PRs carry no `[advisor]` line this pass builds on.

Scope is the design's "What the command records" and "The check", less every action that
launches, ends, or merges. The check decides, prints, and records each decision. Pass 2 carries
out a launch or an end. This pass edits none of `scripts/gate.sh`, `.github/workflows/`, or
`web/`, which are the Page round's.

## Pre-flight

No step in this pass needs Geoff. Every write is inside the worktree or a test's `t.TempDir()`, and
no test runs the real `claude` or touches the network.

- Permissions checked: `~/.local/state/claude-denials.jsonl`, read 2026-11-18, 130 rows,
  2026-11-01T00:55:05Z to 2026-11-18T06:27:23Z, all inside the 30-day window. Classes and rows:
  Classifier unavailable 47; Destructive or Unwanted Actions 14; Self-Modification 11; Interfere
  With Workloads 11; server-side classifier, no reason given 10; Undermining Oversight 5;
  Unauthorized Access 4; Merge Without Review 3; Auto-Mode Bypass 3; Irreversible Local
  Destruction 2; Git Destructive 2; Create Public Surface 2; Unverifiable Deletion Scope 2;
  Sensitive-Source Provenance 2; Security Weaken 2; Production Deploy 1; Secret-Store Writes 1;
  DNS / Domain / Cert Changes 1; Bot Check Circumvention 1; Code from External 1; Shared Scratch
  Sweep 1; Safety Bypass Flag 1; Out-of-Place Publication 1; PII Data Handling 1; Modify Shared
  Resources 1. Rows naming this pass's commands: `go build`/`go test`/`go vet`/`gofmt` 1 (a
  chained `kill`, Interfere With Workloads); `gate.sh` 0; `git push` 4 (all chained or
  destructive); `gh pr create` 0; `gh run` 1 (the same chained `kill`); `claude agents` 1 (piped
  into python, 2026-11-17); `claude logs` 0; `claude stop` 5 (all chained); `claude --bg` 1 (in a
  classifier outage); `git worktree` 4 (chained); `git commit` 4 (`--no-verify`, an outage, Modify
  Shared Resources); `curl` 3. The alternative for each is in the table below.
- Reviewed: one `standards:plan-reviewer` read of this plan (Opus, medium, 2026-11-18), as the
  round plan's Pre-flight asks of every pass: 4 blocking, 12 should-fix, 9 notes. Every finding
  is folded except one owner fork, which is filed (the per-model limit line in
  `## Decisions`).
- Evidence: the mechanism table and the recorded searches below. Each acceptance check was run on
  today's code before any code was written; the outputs are under `## Acceptance`.

### Refused commands and their alternatives

| Refused in the log | This pass does instead |
| --- | --- |
| `claude agents --json` piped into a script | No session needs the real list. Tests use the stand-in. If a live look is ever needed, redirect it to a file in `$CLAUDE_JOB_DIR/tmp` and read the file |
| A chained `claude stop` | This pass ends no session. If one is ever needed, `claude stop <id>` runs alone |
| `claude --bg` | This pass launches nothing |
| `kill` or `pkill` before or around `go test` | Stop a background command with `TaskStop`. Tests end their child processes through `exec.Cmd` |
| `git push` chained, or to `main` | Push only `sessions-check`, alone |
| `git commit --no-verify` | Commit by path with hooks on |
| `git stash` | No stash. A temporary commit sets work aside |
| `curl` | No test or task needs it. The probe is Go code, and tests point it at a local `httptest` server |
| An edit to a Claude settings file (Self-Modification) | No settings edit. A task that needs one parks (round-run standing answer 3) |
| A classifier outage | Wait 60 seconds and run it once more, then flag the task and go on |
| `gh pr merge` (Merge Without Review) | The pass merges nothing. It ends with the PR ready |

## Mechanism table

| Mechanism | Probe or source |
| --- | --- |
| The command today | `go build -o $CLAUDE_JOB_DIR/tmp/pre/cc-desk ./cmd/cc-desk`, then `cc-desk check --dry-run`: `Error: unknown command "check" for "cc-desk"`, exit 1 |
| The session list | Round plan, shared assumptions: `claude agents --json --all` is "the supported way to read session state from outside Claude Code". The first row in `spikes/2-states/fixtures.json` carries `cwd`, `id`, `kind`, `name`, `pid`, `sessionId`, `startedAt`, `state`, `status`. `waitingFor` appears only on the prompt row, so rows are decoded tolerantly: unknown keys ignored, never through `decodeStrict` |
| The row states | Prove findings, proof 2: working reads `state` `working`; a finished session reads `done` with `status` `idle`; a prompt reads `blocked`, `waiting`, `waitingFor: "permission prompt"`; a lost connection reads `blocked idle` with no reason in the row. So "blocked" is read before the idle rule's `status: idle` |
| A blocked session's text | Prove findings, proof 2: the lost connection's reason is on the `claude logs` screen ("API Error: Couldn't connect through your proxy (ERR_PROXY_TUNNEL)"); "a 700-character tail of it did not [hold the line], so the check must search the whole screen". The sign-in and limit texts are recorded `state.json` `detail` text, not live: "Login expired", "You've hit your weekly limit · resets Nov 6, 11pm (America/Anchorage)", "You've hit your Fable limit · resets Nov 20, 11pm (America/Anchorage)" |
| The classifier | `spikes/2-states/classify.py`: `Login expired`; `You've hit your .{0,30}limit`; `Can't reach the API server\|EAI_AGAIN\|Couldn't connect through your proxy\|ERR_PROXY_TUNNEL\|proxy refused the tunnel`. It searches `state.json` `detail` then `needs`, then the screen with `\x1b\[[0-9;?]*[A-Za-z]` stripped. All 16 cases pass today (`python3 -I spikes/2-states/classify.py spikes/2-states/fixtures.json`) |
| `claude logs` | `spikes/own.sh:173-175` `cmd_simple` runs `"$CLAUDE_BIN" "$cmd" "$id"` (here `logs`) with the short id, and the spike scripts read its output to the end, so it exits. The check gives it a 10-second context timeout anyway |
| The state file | `~/.claude/jobs/<short id>/state.json`. This session's `~/.claude/jobs/5bd1155a/state.json` carries `detail`, `state`, `sessionId`, and 21 other keys; the check reads `detail` and `needs` tolerantly. The docs tell scripts not to read it (proof 2), so its absence or a parse failure is no evidence, never an error |
| The probe | Design, "The check": "one HTTPS request to `api.anthropic.com`, and any HTTP answer counts as the link being up". Go `net/http` with a 10-second timeout: any response is up, any error is down |
| The state lock | `internal/state/lock.go:14` `LockWait = 10 * time.Second`; `Lock(dir, wait)` takes `dir/lock` with `LOCK_EX\|LOCK_NB`, polling every 25 ms. A 30-second second read held inside it would make every other verb wait out its 10 seconds and fail |
| Two checks never at once | Design: "The second takes the state lock's 10-second wait and exits." The check takes a second flock on `<state>/check.lock` with the same wait for its whole run, then commits under the state lock as every verb does |
| The commit path | `internal/state/commit.go:55` `Commit(dir, now, wait, ev, render)`: lock, `Load`, apply `ev`, `settle`, write the dirty releases atomically, the snapshot last. `cmd/cc-desk/root.go` `commitReading` takes the usage reading before the lock and folds it in inside. The check uses the same shape: evidence outside the lock, a pure decide event inside |
| A loaded state with many building passes | `internal/state/rules.go:188` `Building()` counts; nothing in `Load` or `Check` refuses more than two building. The two-pass limit applies only at a start. So one fixture can hold nine building passes |
| Stop through the last round's path | `internal/state/events.go:255` `StopPass(project, pass, reason)` refuses a pass that isn't building and a reason not in `StopReasons`; it opens a stop interval and sets `passEvent`, so a forecast line follows |
| Stand-ins | `internal/usage/usage.go:65` `Read(ctx, command)` runs `exec.CommandContext(ctx, command)` when the hidden `--usage-command` is set. `e2e/main_test.go:69` `runWith` gives the binary `PATH=/usr/bin:/bin`, so no real `claude` is on the path |
| No skipped test | Last round's standing answer 19: "No Go test may skip, conditionally or not." `scripts/gate.sh` fails on a skip event |
| The snapshot goldens | `e2e/events_test.go:68` `TestFixtureSnapshotIsCurrent` compares the binary's snapshot at `workedClock` with `e2e/testdata/state/snapshot.json` byte for byte. `internal/snapshot/snapshot_test.go:27` reads `testdata/empty.golden.json` and takes `-update`. Top-level keys today: `format`, `generated_at`, `guess_minutes_limit`, `guess_minutes_total`, `next_session`, `open_guesses`, `releases`, `rulings`, `waiting`, `waits_for_geoff` |
| The Page round's reading | Page round plan, standing answer 16: a Sessions snapshot change is expected, and its type test ignores unknown fields. The Page design shows an unknown stop code with underscores turned to spaces |
| The gate | `bash scripts/gate.sh` in the worktree on 2026-11-18: gofmt, go vet, and go test ok (e2e 27.7 s), vale, `gate: ok`, exit 0 |
| CI | `.github/workflows/gate.yml` runs the gate. `gh run list --repo glw907/cc-desk --branch main --limit 3`: the `gate` workflow succeeded, latest 44275394549 in 54 seconds |

## Recorded searches

`grep -rnE <term> --include=*.go .` in the worktree, 2026-11-18:

| Search | Hits | Each hit |
| --- | --- | --- |
| `StopReasons\|one of the five\|irreversible, secret` | 11 | `types.go:32,33` and `check.go:18,19`: the list and its comment, changed (task 1). `check.go:183,186` and `events.go:253,258`: "one of the five" wording, changed to name no count. `pass.go:70,80`: the stop verb's help, changed to list the ten. `check.go:139` "one of the five states" is pass states, dismissed |
| `\.Repo\b` | 2 | `snapshot.go:256` copies it (left); `check.go:32` requires it (left). The checkout is a new field beside it |
| `"repo"` under e2e (`grep -rl '"repo"' e2e`) | 9 | Six release fixtures (`state/releases/alpha,beta`, `order/releases/north,south`, `writers/releases/east,west`), `state/snapshot.json`, and `inputs/gamma.json`: none gains a field, since the new release fields are optional. `broken_test.go`: the find strings, left |
| `testdata/state/snapshot.json\|empty.golden.json` | 3 | `events_test.go:74,79` and `snapshot_test.go:27`: both goldens regenerate (task 6, standing answer 24) |
| `needs_you_now` | 8 | `estimate.go:433` sets it; changed so a conductor stop or the list-failure entry sets it too. `snapshot.go:102`, the `NextSession` comment: gains the new causes. The six test hits (`e2e/estimate_test.go:307,308`, `e2e/order_test.go:362`, `snapshot_test.go:213`, `nextsession_test.go:122,123`) assert today's cases: left |
| `MarkHidden\|AddCommand` | 9 | `root.go:120`: the check verb registers here. `root.go:117,119`: the persistent hidden flags; the check's are local, so left. `pass.go:52,54,82`, `release.go:16`, `round.go:16`, `ruling.go:54`: subcommand wiring. `release.go:16` gains the `checkout` subcommand beside it; the rest left |
| `LockWait\|state.Lock\(\|func Lock` | 5 | `lock.go:13,14`: the constant and its comment, reused for the check lock. `lock.go:19`: `Lock` gains a sibling that takes a file name (task 5). `root.go:85`, `lock_test.go:38`: left |
| `exec.CommandContext` | 2 | `usage.go:70,72`: the model for the `claude` stand-in. Left |
| `rel == "lock"` in `e2e/main_test.go` | 1 | `folder()` at `:129` skips the state lock. It skips `check.lock` too (task 7), so a check test's before-and-after compare works |

## File list

| File | Task | Change |
| --- | --- | --- |
| `internal/state/types.go` | 1 | The session record, the pull request record, the relaunch and prompt entries, the release checkout fields, five stop reasons, `check.json`'s type |
| `internal/state/check.go`, `read.go` | 1 | Strict checks on the new fields; one worktree and one session name per live record; `check.json` loaded strictly and optionally |
| `internal/state/events.go` | 1, 2, 5 | "One of the five" wording; `SetCheckout`; the event that records a check's results |
| `internal/state/commit.go` | 5 | Writes `check.json` when an event changed it |
| `internal/state/lock.go` | 5 | A lock on a named file beside `Lock` |
| `cmd/cc-desk/pass.go` | 1 | The stop help names the ten reasons |
| `cmd/cc-desk/release.go` | 2 | `release checkout` |
| `cmd/cc-desk/check.go` (new), `root.go` | 5 | The `check` verb and its registration |
| `internal/sessions/` (new): `list.go`, `classify.go`, `read.go`, `decide.go` | 3, 4 | The tolerant row reader, the classifier, the outside reads, the decision |
| `internal/snapshot/snapshot.go`, `sessions.go` (new), `estimate.go` | 6 | `sessions`, `needs_you_now`, `last_check`; the next-session state |
| `internal/snapshot/testdata/empty.golden.json`, `e2e/testdata/state/snapshot.json` | 6 | Regenerated, additions only |
| `*_test.go` beside each | 1 to 6 | Unit tests |
| `e2e/check_test.go` (new), `e2e/main_test.go` (`folder()` only), `e2e/broken_test.go` (one added case) | 7 | Item 3 |
| `e2e/testdata/check/`, `check-launch/`, `check-fail/` (new) | 7 | The item 3 fixtures: state folders, list files, screens, job state files |
| `docs/STATUS.md` (below the round-report block), `docs/follow-ups/` | 8 | Close |

## What the command records

Field names are the builder's (standing answer 21). The names below are the plan's proposal.

A **session record** sits under what the session works on: `Pass.Session` for a pass session,
`Round.Session` for a design writer, `Release.Conductor` for a conductor. Every one is optional
with `omitempty`, so every existing fixture still loads. It holds:

| Field | Meaning |
| --- | --- |
| `name`, `kind` | The `claude --bg` name; `pass`, `design`, or `conductor`, which must match where the record sits |
| `model`, `effort` | As launched |
| `short_id`, `session_id` | What `claude --bg` gave it. `short_id` is the row's `id` |
| `worktree`, `branch` | Absolute path; branch name |
| `launched` | Time of the last launch. Every launch clears `last_seen` and `last_working` |
| `launch_tries` | Launches since the session last showed a row, 1 to 2. The first launch sets it to 1; pass 2 owns every change to it |
| `relaunches` | Each with `time`, `cause`, and `counted` (a relaunch after a prompt's answer or a fix doesn't count, per the design's calls) |
| `last_seen`, `last_state` | When the check last saw a live row, and the state it read (`working`, `blocked`, `idle`, `missing`) |
| `last_working` | The last time the check saw the row read `working` |
| `prompts` | Each with `time`, `tool`, `command`, `suggested_rule`, `answer`. Pass 3 fills them; this pass adds the type and its checks |
| `finished` | For a conductor or design writer: the time it reported its work finished |
| `stop_reason`, `stopped` | For a conductor or design writer only, which have no pass state. A pass's stop reason stays on the pass |
| `stop_detail` | On every session record: the cause the check read at its last stop (the reset text, the last line). The snapshot's `cause` reads it, so it stands after the run that wrote it |
| `ended` | The time the session was ended for good. Pass 2 sets it. A record without it is live |

A **pull request record**, `Pass.PullRequest`, holds `number`, `commit` (the last one the session
reported), `verdict`, `fix_attempts`, `ready` (the time of the standing ready report, cleared by a
fix relaunch), and `merged`.

A **release** gains `checkout` (absolute path) and `default_branch`. The verb
`cc-desk release checkout <project> <path> --default-branch <branch>` sets both; the path must be
absolute and an existing folder. A start refused for a release with neither is pass 2's.

**Strict checks** (`internal/state/check.go`), each a refusal through the existing load path:

1. A session record's `kind` matches its place, and `name`, `short_id`, `session_id`, `worktree`
   (absolute), `branch`, and `launched` are set.
2. `launch_tries` is 1 to 2; each relaunch carries a time and a cause; each prompt a time and a tool.
3. A session record on a pass only when the pass is building, stopped, or done.
4. `stop_reason` on a session record only for a conductor or design writer, and only one of the
   ten reasons, with `stopped` set with it.
5. `pull_request` only on a pass; `number` above 0.
6. `checkout` absolute when set, and `default_branch` set exactly when `checkout` is.
7. Across every live release, no two live records name one worktree or one session name. The
   error names both records.
8. `release add` and `round design` refuse an input that carries a session record, a pull request
   record, `checkout`, or `default_branch`, as their guards at `events.go:42-48` and `:119-121`
   refuse the pass fields today. Only the check and the verbs of passes 2 to 4 set them.

**Stop reasons.** Five join the list, named to read well with underscores turned to spaces:
`permission_prompt`, `sign_in_expired`, `unreadable_stop`, `relaunch_limit`, `launch_failed`. A
session's usage limit is the existing `usage`.

**`check.json`** is a new state file: `format`, `list_failures` (failed list reads in a row),
`last_run` (time or absent), and `decisions` (the last run's, each with `project`, `record`,
`session`, `decision`, `reason`, and `detail`). It loads strictly and optionally: absent reads as
zero. No format bump: every field is additive, and every reader is this binary (`## Decisions`, no format bump).

## The check

`cc-desk check` runs in this order:

1. Create the state folder if it is missing, as `Commit` does at `commit.go:56`, then take
   `<state>/check.lock` with the 10-second wait, held to the end. A second check exits 1 with an
   error naming `check.lock`, and writes nothing.
2. Take the usage reading, as `commitReading` does.
3. Load the state folder without the state lock, to learn the records. A load error goes straight
   to step 9 with no event, so the error file is written as every verb writes it.
4. Read the session list through the `claude` command. A failure, or output that isn't a JSON
   list, exits 1 with the error after one write. That write is its own commit path: a `noSettle`
   flag on `State`, set by the failed-read event, makes `Commit` skip `settle`, and the event
   doesn't go through `WithReading` or fold the usage reading. It writes `list_failures + 1` to
   `check.json` and the snapshot, and no release file and no forecast line (`## Decisions`, the failed-read count).
5. Sort each row into one class. A row with the record's `short_id` reading `working` or `blocked`
   is live. A `done` row is the idle case. A `failed` or `stopped` row, or none, is no row. When
   `--all` gives two rows for one id, a live one wins. If any record not stopped or ended has no
   row, wait the second-read time (30 seconds) and read the list once more. Its rows replace the
   first read's for every record. A failure here is a failed read, as in step 4.
6. For each blocked row with no `waitingFor`, read `jobs/<short id>/state.json` and
   `claude logs <short id>`, with a 10-second timeout. A missing or unparsable job file, and a
   failed or timed-out `logs`, read as empty text.
7. If any record classifies as a lost connection and usage doesn't hold its relaunch, send the
   probe once.
8. Hold the evidence: both reads, the screens, the job files, the probe's answer.
9. `state.Commit` with an event that calls `sessions.Decide(state, evidence, now)` on the freshly
   locked state, then records each decision. A record that appeared after step 3 has no evidence
   and is left alone this run.
10. Print the decisions to stdout as JSON lines, one per record, and exit 0.

`--dry-run` runs steps 1 to 8 and `Decide` on the step 3 state, prints the decisions, and writes
nothing: the folder is byte for byte what it was, `check.lock` aside. On a load error or a failed
list read it exits 1 and writes nothing either.

**The decision** is a plain function over the state, the evidence, and a clock. It goes record by
record, first match wins:

| # | When | Decision | Reason |
| --- | --- | --- | --- |
| 1 | The record carries `ended` | `none` | |
| 2 | The record is on a pass in state `done` | `none` (pass 4 ends it) | |
| 3 | A conductor or design writer with `finished` set | `end` | |
| 4 | A pass, or a conductor's or design writer's record, stopped with `usage` | `relaunch` (the resume; `counted` false) | `usage` |
| 5 | A pass or record stopped for any other reason | `none` (its "Needs you now" stands) | its reason |
| 6 | The row reads `working` | `none`; record `last_working` and `last_seen` | |
| 7 | Blocked, `waitingFor` is `permission prompt` | `stop` | `permission_prompt` |
| 8 | Blocked, the text matches "Login expired" | `stop` | `sign_in_expired` |
| 9 | Blocked, the text matches the limit pattern | `stop` | `usage` |
| 10 | Blocked, the text matches the connection pattern, two counted relaunches | `stop` | `relaunch_limit` |
| 11 | The same under two, probe answered | `end_and_relaunch` | `lost_connection` |
| 12 | The same, probe failed | `wait_link` | |
| 13 | Blocked, none of the above | `stop` | `unreadable_stop`, detail the last line |
| 14 | `done` row, idle 15 minutes or more, no `ready` report, two counted relaunches | `stop` | `relaunch_limit` |
| 15 | The same under two | `end_and_relaunch` | `idle` |
| 16 | `done` row under 15 minutes, or a `ready` report stands | `none` | |
| 17 | No row at the second read, `last_seen` unset, `launch_tries` 1 | `launch_again` | |
| 18 | The same, `launch_tries` 2 | `stop` | `launch_failed` |
| 19 | No row at the second read, seen before, two counted relaunches | `stop` | `relaunch_limit` |
| 20 | The same under two | `relaunch` | `missing` |

Idle time runs from the later of `launched` and `last_working`. Usage is checked before rows 4, 11,
15, 17, and 20: while `state.UsageStop()` holds (the reading at the stop line or unreadable), each
of them decides `wait_usage` instead, and that wait counts nothing. Under that hold step 7 sends no
probe, so a lost connection decides `wait_usage`, never `wait_link`. A `stop` on a pass goes
through `StopPass`; on a conductor or design writer it sets the record's `stop_reason` and
`stopped`. Every stop also sets the session record's `stop_detail`. Its text for row 9 is the
reset text, and for row 13 the last line.

The last line is the job file's `detail` when it is set. Otherwise it comes from the screen: strip
`\x1b\[[0-9;?]*[A-Za-z]`, OSC sequences (`\x1b]` to BEL or `\x1b\\`), and `\x1b7`/`\x1b8`; take
the last non-empty line; collapse runs of white space to one space; keep its last 200 runes. With no
text at all it is "no text".

**What a check writes in this pass** (standing answer 23): `last_seen`, `last_state`, and
`last_working`; each `stop`; and `check.json`. It never changes `launch_tries` or `relaunches`
and never adds a relaunch entry, because pass 2 does that when it carries the decision out. So
"is ended and relaunched" reads, in this pass, as the decision `end_and_relaunch` printed by
`--dry-run` and kept in `check.json`.

**Stand-ins**, each a hidden local flag on `check`, named in the tests and nowhere else:

| Flag | Default | Test value |
| --- | --- | --- |
| `--claude-command` | `claude` | A script the test writes. It logs each call's arguments to a file and answers `agents --json --all` from `list-1.json` on its first call and `list-2.json` after (falling back to `list-1.json`), and `logs <id>` from `screens/<id>.txt` |
| `--jobs-dir` | `~/.claude/jobs` | A folder under `t.TempDir()` holding `<id>/state.json` |
| `--probe-url` | `https://api.anthropic.com/` | An `httptest.NewServer` URL that logs each call. For a failing probe, its handler logs the call, then hijacks the connection and closes it, so the client gets an error and the test still sees one attempt |
| `--second-read-wait` | `30s` | `0s`, plus one case at `200ms` that asserts the gap in the stand-in's call times |

The stand-in script prints nothing and exits 0 for `logs <id>` with no screen file. One e2e helper
runs `check` and always passes all four flags and `--usage-command`, so no test can reach the
real `claude` or the network by leaving one out.

## The snapshot

Three top-level fields, each always present:

- `sessions`: `[]`, one entry per session record: `project`, `round`, `pass` (null for a
  conductor), `kind`, `name`, `short_id`, `last_state`, `last_seen`, `last_working`, `relaunches`
  (the counted number), `stop_reason` (null when none), and `prompts` (`[]`).
- `needs_you_now`: `[]`, one entry per cause: `project`, `round` (null for a conductor or the list
  entry), `session` (null for the list entry), `reason`, `cause` (the session record's
  `stop_detail`, or the prompt's command once pass 3 records one), and `todo`, the one thing to do.
  Proposed texts: permission prompt "answer it at the desk"; sign-in "sign in to Claude Code";
  usage "wait for the reset, or reset early"; unreadable stop and relaunch limit "read the
  session's log, then start the pass again"; failed launch "read the launch error, then start the
  pass again". Three failed list reads add an entry with reason `session_list_unreadable`, which is
  not a stop reason and sits on no record.
- `last_check`: null, or `time`, `list_failures`, and `decisions` from `check.json`.

`next_session.state` reads `needs_you_now` when any `needs_you_now` entry stands, as a stopped pass
already makes it.

## Acceptance

Every check runs the built `cc-desk` binary against a copy of a fixture state folder in
`t.TempDir()` through `go test ./e2e/...`, with the stand-ins above. Byte-for-byte comparisons
leave out `lock` and `check.lock`.

| Item, word for word | Task | Check |
| --- | --- | --- |
| **Stops the check reads.** Fixture rows, one each: a permission prompt, "Login expired", a usage limit, an unreadable block, a lost connection with the probe failing, the same with the probe answering, a session idle 16 minutes, one idle 14, one idle 16 with a ready report standing, a launch with no row after one check, and the same after two. The snapshot names the first four as "Needs you now", each with its own stop reason and the one thing to do. The fifth is left alone, and its relaunch count doesn't change. The sixth and seventh are ended and relaunched. The eighth and ninth are left alone. The tenth launches again, and the eleventh reads "Needs you now". A usage-limit session relaunches when the usage stand-in drops under the stop line, and its count doesn't change. Three failed list reads in a row read "Needs you now". | 4, 7 | `go test ./e2e/ -run 'TestCheck' -count=1 -v`, read for a `--- PASS` line for each subtest `TestCheckStops/<clause>` below and for `TestCheckLock` and `TestCheckDryRun`. An `ok` with no such lines, or `[no tests to run]`, is a fail. `go test ./internal/sessions/ -v` runs the same rows through `Decide` as one table |
| **The gate.** `bash scripts/gate.sh` passes, with no skipped test. Items 6 and 7 are not Go tests. Their script's output is in the round report, and a pass that leaves it out isn't done. | 8 | `bash scripts/gate.sh` run bare; the PR's `gate` check. Items 6 and 7 are pass 5's; this pass's round report line says so |

Item 3, clause by clause. The fixture `e2e/testdata/check/` holds one release with nine building
passes, `s01` to `s09`, each with a session record, in row order, plus the list files, screens,
and job files. Clauses 10 and 11 have their own small folder, `e2e/testdata/check-launch/`, with
`s10` and `s11` only, so that only their records can cause the second read. The clock is each
fixture's `now`, and neither folder holds a `usage.json`. Every run passes `--usage-command
e2e/testdata/usage40.sh` unless a clause names another reading: without a reading, usage is
unreadable and every relaunch becomes `wait_usage`. A reading stays fresh for 5 minutes
(`usage.Cached`, `internal/usage/usage.go:88-94`), so a run that needs a new reading on the same
folder moves the clock on 5 minutes. The probe is sent at most once a run, so clauses 5 and 6 are
two runs on fresh copies of the folder: run A with the probe failing, run B with it answering.
Clauses 1 to 4 and 6 to 9 are asserted on run B.

| Clause | Fixture | Assertion |
| --- | --- | --- |
| A permission prompt (1) | Row `blocked`, `waiting`, `waitingFor` `permission prompt`, copied from `spikes/2-states/fixtures.json` "permission prompt, save 1" | Snapshot `needs_you_now` names `s01`, reason `permission_prompt`, todo set; `s01` reads stopped |
| "Login expired" (2) | Row `blocked idle`; job file `detail` "Login expired" | Same for `s02`, `sign_in_expired` |
| A usage limit (3) | Row `blocked idle`; job file `detail` "You've hit your weekly limit · resets Nov 6, 11pm (America/Anchorage)" | Same for `s03`, `usage`, cause holds the reset text. The usage-limit relaunch's first run, below, asserts that the causes of `s03` and `s04` still stand |
| An unreadable block (4) | Row `blocked idle`; no job file; a screen with no matching text (`spikes/2-states/out/logs-idle-1.txt`) | Same for `s04`, `unreadable_stop`, cause exactly `ng… 3 Brewing… ✶ * ✢ Brewing… · Brewing… Brewing… Brewing… Brewing… ✢ Brewing… 4 ✻ Crunched for 4s · done 6:34 PM ❯ You've used 87% of your weekly limit · resets Nov 20, 11pm (America/Anchorage) 2 s 2` (the last 200 runes; it reads "used", so the limit pattern doesn't match) |
| Lost connection, probe failing (5) | Row `blocked idle`; the screen `spikes/1-wake/out/logs-blocked.txt`, which holds "ERR_PROXY_TUNNEL"; run A, probe at the hijack-and-close server | The server logged one call; decision `wait_link`, with no `counted` field; `s05` stays building; its `relaunches` unchanged |
| The same, probe answering (6) | `s06` as `s05`; run B, probe at the `httptest` server | Decision `end_and_relaunch`; the server logged exactly one call although `s05` and `s06` both read lost |
| Idle 16 minutes (7) | Row `done idle`; `last_working` 16 minutes before the clock | Decision `end_and_relaunch`, reason `idle` |
| Idle 14 (8) | The same at 14 minutes | Decision `none`; not stopped |
| Idle 16 with a ready report (9) | As 7, with `pull_request.ready` set | Decision `none` |
| No row after one check (10) | `check-launch`. No row in either list; `last_seen` unset; `launch_tries` 1 | Decision `launch_again`; the stand-in's log shows two list reads |
| The same after two (11) | `check-launch`, `launch_tries` 2 | `needs_you_now` names `s11`, `launch_failed` |
| Usage-limit relaunch | On run B's folder, now holding `s03` stopped with `usage`: run at the clock plus 5 minutes at 100 percent through a stand-in the test writes with `newStandIn`, then on the same folder at the clock plus 10 minutes at 40 percent | Each usage stand-in logged exactly one call. At 100: `s03` decides `wait_usage`, and the snapshot still gives `s03` and `s04` their causes from clauses 3 and 4. At 40: `s03` decides `relaunch` with `counted` false, and its `relaunches` is unchanged. Only `s03` is asserted in these two runs, since the later clock moves `s08` past 15 minutes |
| Three failed list reads | Its own small folder, `e2e/testdata/check-fail/`: one release with one waiting pass held by `usage`, and a `usage.json` at 40 percent read at the folder's clock, so a `settle` would start the pass. A list stand-in that exits 1, run three times | Each run exits 1, the waiting pass stays waiting, every release file is byte for byte, and `forecasts.jsonl` gains no line; after the third, `needs_you_now` holds the `session_list_unreadable` entry. A good read resets `list_failures` to 0 |

Two design lines outside item 3 get a test in the same file: `TestCheckLock` (a held `check.lock`
makes a second check exit 1 after 10 seconds and write nothing) and `TestCheckDryRun` (the folder is
unchanged after `--dry-run`, and the printed decisions match a real run's `check.json`).

What today's code printed, before any change:

- Item 3: `cc-desk check --dry-run` on `e2e/testdata/state` printed
  `Error: unknown command "check" for "cc-desk"` and exited 1. Fails today.
- Item 8: `bash scripts/gate.sh` in the worktree printed `gate: ok`, exit 0, and the last CI run on
  `main` (44275394549) passed. It holds today and must still hold at close.

## Tasks

Each task commits on its own, runs `bash scripts/gate.sh` bare before the commit, and is pushed.
No task holds code; the sections above fix the behavior.

### Task 1: the record fields and their checks (22 minutes)

- Outcome: the release file carries session records, pull request records, and the checkout
  fields; the strict reader checks them; the stop list holds ten reasons and no message counts
  them; `check.json` loads.
- Files: `internal/state/types.go`, `check.go`, `read.go`, `events.go`, `cmd/cc-desk/pass.go`.
- Class: `ordinary`.
- Check: `go test ./internal/state/ -v`.
- Test: one table in `check_test.go` with a passing record and one failing case per strict check
  1 to 7; one refusal case per new field in a `release add` input and in a `round design` input
  (check 8); `check.json` absent, valid, and broken. Every existing fixture still loads (the e2e suite
  stays green).

### Task 2: the checkout verb (10 minutes)

- Outcome: `cc-desk release checkout <project> <path> --default-branch <branch>` records both
  fields and refuses a relative or missing path.
- Files: `internal/state/events.go`, `cmd/cc-desk/release.go`, `e2e/events_test.go` (one added
  test), `e2e/broken_test.go` (`verbs` gains `release checkout`).
- Class: `ordinary`.
- Check: `go test ./internal/state/ ./cmd/... -v`, then `go test ./e2e/ -run 'TestReleaseCheckout|TestBrokenFile' -count=1 -v`.
- Test: the event sets both and marks the release dirty. A relative path and a missing folder are
  each refused through `state.Refuse`, so the binary exits 2 and the folder is byte for byte.

### Task 3: the session list and the classifier (20 minutes)

- Outcome: `internal/sessions` reads the list tolerantly, reads a screen and a job file, sends the
  probe, and labels a blocked row from its text.
- Files: `internal/sessions/list.go`, `classify.go`, `read.go` and tests.
- Class: `ordinary`.
- Check: `go test ./internal/sessions/ -v`.
- Test: the 16 cases of `spikes/2-states/fixtures.json`, ported as one table (copied into
  `internal/sessions/testdata/`, the spike file unchanged), each giving `classify.py`'s label. A
  list with an unknown key, one that isn't a list, an empty one, and two rows with one id. The
  last-line rule on both spike screens, with escape bytes absent from the result. The probe against
  an `httptest` server answering 404 (up) and one that hijacks and closes (down).

### Task 4: the decision (22 minutes)

- Outcome: `sessions.Decide` gives the decision table's answer for every row of it.
- Files: `internal/sessions/decide.go`, `decide_test.go`.
- Class: `ordinary`.
- Check: `go test ./internal/sessions/ -run TestDecide -v`.
- Test: one table entry per decision-table row (20), plus the eleven item 3 rows by name, plus
  rows 4, 11, 15, 17, and 20 turned to `wait_usage` at the stop line and with an unreadable
  reading, plus the second read finding a row the first missed (decision `none`), plus idle time
  measured from a `launched` later than `last_working`.

### Task 5: the check verb (22 minutes)

- Outcome: `cc-desk check` runs the ten steps, with `--dry-run` and the four hidden flags, and the
  check lock.
- Files: `cmd/cc-desk/check.go`, `root.go`, `internal/state/lock.go`, `events.go`, `commit.go`.
- Class: `ordinary`.
- Check: `go test ./cmd/... ./internal/state/ -v`. The e2e check is task 7's.
- Test: the lock on a named file has `Lock`'s tests run against it; the record event writes
  `check.json` and the session fields, and leaves `launch_tries` and `relaunches` alone. The
  failed-read write, run on a state with a waiting pass no limit holds, leaves that pass waiting
  and writes no release file and no forecast line. The lock sibling's error names `check.lock`.

### Task 6: the snapshot (12 minutes)

- Outcome: the snapshot carries `sessions`, `needs_you_now`, and `last_check`, always present, and
  the next-session state reads `needs_you_now` for any entry.
- Files: `internal/snapshot/snapshot.go`, `sessions.go`, `estimate.go`, both goldens.
- Class: `ordinary`.
- Check: `go test ./internal/snapshot/ -v`; the additions-only check below.
- Test: `TestSnapshotEveryFieldPresent` gains the three keys; a state with one stopped conductor
  and one with three list failures each read `needs_you_now`; a stopped pass's entry takes its
  cause from `stop_detail`. Regenerate `empty.golden.json` with `-update`. Regenerate the fixture
  golden from a copy, never in place, because the run rewrites the stale `usage.json`:

  ```
  T=$(mktemp -d) && cp -r e2e/testdata/state "$T/s" && go run ./cmd/cc-desk --state-dir "$T/s" --now 2026-11-20T23:00:00Z --usage-command "$PWD/e2e/testdata/usage40.sh" snapshot > e2e/testdata/state/snapshot.json
  ```

  Then `git status --short e2e/testdata` names only `snapshot.json`. The additions-only check, run
  for each golden, prints nothing:
  `diff <(git show origin/main:<golden> | jq -S .) <(jq -S 'del(.sessions, .needs_you_now, .last_check)' <golden>)`.

### Task 7: the item 3 fixture and tests (35 minutes)

- Outcome: item 3 passes end to end.
- Files: `e2e/check_test.go`, `e2e/testdata/check/`, `check-launch/`, `check-fail/`, `e2e/main_test.go` (`folder()` skips
  `check.lock`), `e2e/broken_test.go` (`verbs` gains `check`, through the helper that passes every
  stand-in flag; and `TestBrokenCheckFile`, a test of its own that writes a valid `check.json`
  into the copied folder, breaks it, and runs the existing assertions. `e2e/testdata/state` gains
  no `check.json`).
- Class: `ordinary`.
- Check: `go test ./e2e/ -run 'TestCheck|TestBroken' -count=1 -v`, read for the `--- PASS` lines
  `## Acceptance` names, then the whole `go test ./e2e/ -count=1 -v`.
- Test: the tests are the acceptance checks. Screens and rows are copied from
  `spikes/1-wake/out/logs-blocked.txt`, `spikes/2-states/out`, and `spikes/2-states/fixtures.json`, with ids and names changed to the fixture's. The
  stand-in script lives in the test as a heredoc string, as `usageStandIn` does. No file under
  `e2e/testdata/state`, `order`, `writers`, or `inputs` changes except the regenerated
  `snapshot.json`.

### Task 8: close (15 minutes)

- Outcome: the PR is ready, with review accepted and CI green.
- Files: `docs/STATUS.md` (below the round-report block), each follow-up under `docs/follow-ups/`.
- Class: `ordinary`.
- Check: the PR body's `## Acceptance` evidence; the `gate` check on the last commit.
- Test: none of its own. Follow-ups to file: how an expired sign-in, an unreadable stop, a
  relaunch limit, and a failed launch clear once Geoff acts (the design says how a prompt and a
  usage limit clear, and is silent on these four); and a live sample of the sign-in and limit
  texts on the `claude logs` screen (proof 2's open test); and the uncounted usage relaunch that can
  loop on a per-model limit (the per-model limit line in `## Decisions`).

Path: tasks 1 to 8 in order, 158 minutes. Task 2 may run beside task 3 in a pair after task 1. At
the last round's pace of about 0.6, that is about 95 minutes. Both are under the 180-minute park
line.

The CI watch, run bare after each push: `gh run list --repo glw907/cc-desk --branch sessions-check
--limit 3`, then `gh run watch <id> --repo glw907/cc-desk --exit-status` on the newest.

## Standing answers

The twelve in `round-run.md` hold, with the round plan's changes (4 replaced, 1 and 9 changed) and
its answers 13 to 20. The last round's 19 to 22 hold too: no Go test skips; the lock test waits its
full 10 seconds; a field name is the builder's; a path past 180 stops. This pass adds:

- **23. What a check writes in pass 1.** The `last_*` fields, each stop, and `check.json`. Never
  `launch_tries`, `relaunches`, a worktree, or a session. A launch, an end, or a relaunch is a
  printed and recorded decision. A clause of item 3 that says "ended and relaunched" or "launches
  again" is met by that decision. Rows 10 and 11 have their try count set in the fixture.
- **24. The goldens change.** Answer 16 protects the files the tests feed the binary. The two
  goldens are outputs: they gain the three keys and change in nothing else, which task 6's
  additions-only check shows. Every input file stays byte for byte. The PR lists both goldens.
- **25. A text the classifier doesn't know.** It is `unreadable_stop`. Never widen a pattern past
  the recorded texts in the mechanism table.
- **26. The design's table and the decision table disagree.** The design wins. Change the row, keep
  the unit table in step, and list it under `## Decisions`.
- **27. A design line needs a launch or an end to test.** It is pass 2's. Test the decision and
  file nothing.
- **28. The Page round's work.** Its files and sessions are not this pass's. A merge conflict in
  `e2e/testdata/state/snapshot.json` after the Page round lands: regenerate it on the merged code,
  rerun the additions-only check against the new `origin/main`, and push.
- **29. A pass stopped for one of the older five reasons.** It gets a `needs_you_now` entry too,
  with `session` null and the cause from its stop. Its todo is "read the stop in the repo's STATUS";
  for `usage` it is the usage todo. The fixture snapshot gains these entries as additions, and
  task 6's check still holds.

## Decisions

- [advisor] Evidence is gathered outside the state lock and the decision runs inside it, as
  `commitReading` takes its reading. A 30-second read inside the lock would time out every other
  verb.
  Call: mine (method).
- [advisor] A separate `check.lock` held for the whole run meets "two checks never run at once";
  the second check waits the same 10 seconds and exits.
  Call: mine (method).
- [advisor] Blocked is read before the idle rule, because a lost connection reads `blocked idle`.
  A `done` row is the idle case, not a dead session: the design's idle row names `done`, and the
  more specific row wins over "no live row".
  Call: mine (method).
- [advisor] The failed-read count lives in a new `check.json`. A failed read writes it and the
  snapshot, and runs no `settle`, so a waiting pass can't start on it. The design's "changes
  nothing" is read as no release file changes.
  Call: mine (method).
- [advisor] A non-dry-run check in this pass writes the `last_*` fields, the stops, and
  `check.json`, and leaves the counts for pass 2.
  Call: mine (scope, inside the round plan's line).
- [advisor] The goldens regenerate with additions only, proved by a `jq` diff.
  Call: mine (method).
- No format bump for the new fields: each is optional and additive, and this binary is the only
  reader.
  Call: mine (method).
- The probe stand-in is a URL flag pointed at an `httptest` server, not a script, so the test runs
  the real probe code. The design leaves the stand-in to the builder.
  Call: mine (method).
- A record that appears between the unlocked load and the commit has no evidence and is left alone
  until the next run.
  Call: mine (method).
- Five new stop codes: `permission_prompt`, `sign_in_expired`, `unreadable_stop`,
  `relaunch_limit`, `launch_failed`. The list-failure entry's code, `session_list_unreadable`, is
  not a stop reason.
  Call: mine (naming, standing answer 21).
- [advisor] The plan-reviewer read is folded whole: the usage cache moves the clock on 5 minutes
  between readings; a usage hold turns every relaunch row to `wait_usage`, ahead of `wait_link`;
  the `check.json` broken case has a test of its own, so no fixture changes; the failed read gets
  a `noSettle` commit path; the relaunch limit applies to every cause; a done pass is left alone;
  every stop keeps its cause in `stop_detail`; the probe's failing server logs its one call; and
  clauses 10 and 11 run in their own folder.
  Call: mine (method).
- [advisor] The reviewer's owner fork, the per-model limit. A per-model limit stop, such as "You've hit
  your Fable limit", relaunches as soon as the usage reading is under the stop line, which it may
  already be, and that relaunch counts nothing. The plan builds it as the design says (rows 4 and
  9) and files a follow-up naming three ways out: as designed; hold until the reset time in the
  stop's text; or count usage relaunches after the first. Recommendation in the follow-up: hold
  until the reset time.
  Call: Geoff's (product), filed, not asked; the design stands until he rules.
