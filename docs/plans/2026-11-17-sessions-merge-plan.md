# Merge: pass plan

Initiative: cc-desk
Round: sessions
Pass: sessions-merge

Class `runner`. Planned 60 minutes; the path below is 100, under the 120-minute park line
(round plan answer 14). Design: `docs/specs/2026-11-17-sessions-design.md` (approved
2026-11-17), section "Merging", the merge row of "The decision table", and the fourth row of
"The passes". The state design's "A decision carried to the Sessions round"
(`docs/specs/2026-11-17-state-and-rules-design.md:358-364`) sets the wait for a first check.
Round plan: `docs/plans/2026-11-17-sessions-round.md`. Writing a person reads (the PR body,
STATUS lines, follow-ups) follows `~/.claude/docs/voice/written-for-people.md`, loaded through
the `writing-voice` skill.

Worktree `/var/home/glw907/Projects/cc-desk/.claude/worktrees/sessions-merge`, branch
`sessions-merge`, cut from `origin/sessions-prompts` at `73da88e` (pull request 12, open, which
carries pull requests 10 and 8). PR base: `main`. Push: `glw907/cc-desk`. Before the PR is
ready, merge `origin/sessions-prompts` in again if it moved, and `origin/main` if it moved. On
2026-11-18, `origin/sessions-prompts` was `73da88e` and `origin/main` was `3c60331`, ahead by
STATUS commits only.

Scope: the ready verb a pass session reports with; the merge decisions and the merge itself,
pinned to the reported commit; the merge record and the pass's finish at the merge time; the
pull of the main checkout when it is safe; removal of the worktree and branch with no force
flag; the fix relaunch and the question after the third failure; the wait for a first check;
and a pull request merged by another hand. This pass edits none of `scripts/gate.sh`,
`.github/workflows/`, or `web/`, which are the Page round's.

## Pre-flight

No step in this pass needs Geoff. No step runs the real `claude`, `systemd-run`, or `gh pr
merge`: `claude-merge-guard` denies `gh pr merge` to every Claude session on this workstation,
and the tests run a `gh` stand-in. Every write is inside the worktree or a test's
`t.TempDir()`.

- Permissions checked: `~/.local/state/claude-denials.jsonl`, read 2026-11-18, 130 rows,
  2026-11-01T00:55:05Z to 2026-11-18T06:27:23Z. The file starts on 2026-11-01, so the 30-day
  window from 2026-10-19 holds only 18 days of rows. By class, the largest are Bash with the
  classifier unavailable (38), Bash Destructive or Unwanted Actions (13), Bash Interfere With
  Workloads (11), and Bash refused by the server classifier (8). Rows naming this pass's
  commands: three Merge Without Review rows, for `gh pr checks 9` in a merge script
  (2026-11-01), `gh pr merge 175 --merge --match-head-commit` (2026-11-10), and `gh pr view 234
  --json headRefOid,statusCheckRollup` in a merge context (2026-11-11). `claude stop` was
  refused several times as Interfere With Workloads, Destructive, or Undermining Oversight.
  `claude agents --json | python3` was refused once as Destructive (2026-11-17). A commit, a
  push, and a `gh pr create` were refused once each as Create Public Surface or Out-of-Place
  Publication. `go test`, `gofmt`, and `gate.sh`: 0.
- Reviewed: one `standards:plan-reviewer` read of this plan (Opus, medium), 2026-11-18. Its
  findings and their folds are the last section, "Review findings and their folds".
- Evidence: the mechanism table below. Each acceptance check was run on `73da88e` before any
  code was written; the outputs are under `## Acceptance`.

### Refused commands and their alternatives

| Refused in the log | This pass does instead |
| --- | --- |
| `gh pr merge`, or a `gh pr view`/`gh pr checks` read in a merge context | The executor runs none of these against GitHub. The binary calls them only through `--gh-command`, which every test sets to the stand-in. The CI watch is `gh run list --branch`, and the executor reads its own PR with a plain `gh pr view <n>` |
| `claude stop` | The binary calls it only through `--claude-command`, set to the stand-in |
| `claude agents --json` piped to a script | Read the list without a pipe, or not at all; the tests read the stand-in's log files |
| `git commit --no-verify`, `git stash` | Commit by path with hooks on. No stash |
| A chained `git push` or `git fetch` | Fetch, then push `sessions-merge`, each alone |
| A classifier outage | Wait 60 seconds and run it once more, then flag the task and go on |

## Mechanism table

| Mechanism | Probe or source |
| --- | --- |
| The verbs today | `go run ./cmd/cc-desk pass ready --help` on `73da88e` prints the `pass` group's help, listing `finish`, `start`, and `stop`, exit 0. No ready verb exists |
| What a pull request holds today | `internal/state/types.go:129-137`: `PullRequest{Number, Commit, Verdict, FixAttempts, Ready, Merged}`, each optional but `number`. `internal/state/check.go:228-229` refuses a number of 0 or less. Nothing writes the record yet: no event in `internal/state/events.go` sets it |
| Ready holds the idle rule off | `internal/sessions/decide.go:360-364` `readyStands`: a pass whose `PullRequest.Ready` is set. `doneRow` at `decide.go:286` returns `none` while it stands |
| One read for a pass's pull request | `gh pr list -R glw907/cc-desk --head state-rules-order --state all --json number,state,headRefOid,mergedAt,mergeCommit,statusCheckRollup` printed `[{"headRefOid":"1c128e6f...","mergeCommit":{"oid":"3b65cda9..."},"mergedAt":"2026-11-18T07:40:07Z","number":5,"state":"MERGED","statusCheckRollup":[{"__typename":"CheckRun","completedAt":"2026-11-18T07:39:42Z","conclusion":"SUCCESS",...,"name":"gate","startedAt":"2026-11-18T07:39:06Z","status":"COMPLETED","workflowName":"gate"}]}]`, exit 0 (`gh` 2.105.0). An open one: `gh pr view 12 -R glw907/cc-desk --json number,headRefOid,state,mergedAt,mergeCommit` gave `"number":12,"state":"OPEN"`, and `gh pr checks 12 --json name,state,bucket` gave one `gate` entry, bucket `pass`, state `SUCCESS`, both exit 0 |
| No pull request for a branch | `gh pr list -R glw907/cc-desk --head no-such-branch-xyz --state all --json number` printed `[]`, exit 0. `gh pr view no-such-branch-xyz -R glw907/cc-desk` printed `no pull requests found for branch "no-such-branch-xyz"`, exit 1. So the check uses `pr list`, where exit 0 always means GitHub answered |
| No check registered | `gh pr list -R glw907/workstation --head lean-cutover --state all --json number,statusCheckRollup` printed `[{"number":1,"statusCheckRollup":[]}]`, exit 0. `gh pr checks 1 -R glw907/workstation --json name,bucket` printed `no checks reported on the 'lean-cutover' branch`, exit 1. cli/cli trunk `pkg/cmd/pr/checks/checks.go` `populateStatusChecks` returns that error before the JSON exporter runs. So `gh pr checks` can't tell "no check" from "no network" by its exit, and the check reads the rollup instead |
| GitHub can't be reached | `gh pr list -R glw907/no-such-repo-xyz --head x --state all --json number` printed `GraphQL: Could not resolve to a Repository with the name 'glw907/no-such-repo-xyz'. (repository)`, exit 1. Any non-zero exit is "GitHub can't be reached": merging waits, and nothing else is held |
| A check's bucket | cli/cli trunk `pkg/cmd/pr/checks/aggregate.go` `aggregateChecks`, quoted: `state := string(c.State); if state == "" { if c.Status == "COMPLETED" { state = string(c.Conclusion) } else { state = c.Status } }`. Then `SUCCESS` is pass; `SKIPPED` and `NEUTRAL` skipping; `ERROR`, `FAILURE`, `TIMED_OUT`, and `ACTION_REQUIRED` fail; `CANCELLED` cancel; anything else (`EXPECTED`, `REQUESTED`, `WAITING`, `QUEUED`, `PENDING`, `IN_PROGRESS`, `STALE`) pending. A `CheckRun` carries `status` and `conclusion`; a `StatusContext` carries `state`. The check copies this mapping |
| The merge call | The guard refused `gh pr merge --help` on this workstation, so the manual page is the source: <https://cli.github.com/manual/gh_pr_merge>, `--match-head-commit <SHA>`: "Commit SHA that the pull request head must match to allow merge", and `-m, --merge`: "Merge the commits with the base branch". What a mismatch prints is undocumented, so the stand-in exits 1 on a wrong SHA, and the check treats any non-zero exit as "not merged". The call never carries `-d`/`--delete-branch`, `--auto`, or `--admin` |
| Where the binary looks for `gh` | `e2e/main_test.go:73` runs the binary with `PATH=/usr/bin:/bin`. The real `gh` is `/home/linuxbrew/.linuxbrew/bin/gh` and `/usr/bin` holds none, so a test that misses `--gh-command` fails to find `gh` rather than reaching GitHub. Real `git` and `python3` are in `/usr/bin` |
| The worktree a launch makes | `internal/launch/worktree.go:45` `Make` runs `git worktree add` from `origin/<Default>` with `--no-track`, at `<checkout>/.claude/worktrees/<branch>`, and its undo uses `dropBranch` at `:73`, which runs `git branch -D` (`:78`). The cleanup after a merge is a new function that never calls `dropBranch` |
| What a launch leaves in the worktree | `internal/launch/settings.go:87-92` writes the settings file under `<state>/launch/`, and the prompt hook writes unmatched input under the state folder (`cmd/cc-desk/prompt.go:114-135`). No launch file lands in the worktree, so a plain `git worktree remove` can succeed |
| `git branch -d` and a `--no-track` branch | With no upstream, `git branch -d` checks that the branch is merged into `HEAD` of the checkout it runs in. So the pull must land before the delete, and the stand-in's merge must be a real merge commit on the bare remote's `main` |
| The default branch | `internal/state/types.go` `ReleaseFile.DefaultBranch`; `cmd/cc-desk/launcher.go:233` passes it as the worktree's base |
| A live session in the checkout | `internal/sessions/list.go:14` `Row`, its `Cwd` at `:19`. The pull's live-session test reads the rows the check already lists |
| The commit path | `cmd/cc-desk/check.go:49` `runCheck`: `gatherEvidence` (`:165`, outside the lock, the probe at `:203`), `sessions.Plan` (`decide.go:136`), `state.RecordCheck`, then `carryOut` (`carryout.go:19`) and `recordOutcomes` (`:164`). `carry` (`carryout.go:47`) acts on `launch`, `launch_again`, `relaunch`, `end_and_relaunch`, and `end` only |
| A failed action | `carryout.go`: a failed action's outcome reads `failed: ...`, and the check exits 2 (pass 2) |
| A relaunch's prompt | `internal/launch/prompt.go:29-60`: `Prompt.Relaunch *RelaunchNote` and `Prompt.Answer *AnswerNote`; `Fill` refuses both at once and appends `relaunch.txt` or `answer.txt`. `templates/relaunch.txt` says "This is relaunch {{.Relaunch.Count}} of 2", so a fix relaunch takes a third note and its own template |
| The pass template | `internal/launch/templates/pass.txt:5` and `design.txt:5`: "End with a ready pull request. Do not merge it: the merge belongs to the check." |
| A pass session's verbs | `internal/launch/settings.go:14` `Deny = [Bash(gh pr merge *), Bash(approval-queue land *)]`; `settings.go:57-59` `FollowUpCommand`, a prefix rule ending in ` *`; `launcher.go:224-226` writes the verbs and that rule. `internal/launch/launch_test.go:42-43` and `:89-90` assert them |
| Finishing a pass | `internal/state/events.go:236` `FinishPass` sets `Finished` to the event clock and refuses a pass that isn't building. A pass merged by another hand may be stopped, so the merge needs its own event |
| A question | `internal/state/types.go` `Question{ID, Text, Options, Recommended, Answer}` in `Release.Questions`, validated at `check.go:72-86`: an id, text, options, and a recommended option among them. Only a round's design adds one today (`events.go:134-139`). A pass names the ones it rests on in `Pass.Questions`, checked only for existence (`check.go:146-150`). `internal/snapshot/estimate.go:330-354` `questionItem` times a question by the passes resting on it that have yet to start. With none, it times the question "now" and gives it the round of the pass that rests on it, and `fillSession` (`:358`) then counts it under `waits_for_geoff`. So a question the building pass rests on shows under "Waiting for you" at once |
| Answering a question | `internal/state/ruling.go:79` `RecordRuling` answers a question under the desk guard. `ruling.go:133` says of `RegisterDesk`: "The Desk round calls it when a desk opens; no verb does" (pull request 12's flagged gap) |
| "Needs you now" | `internal/snapshot/sessions.go:79` `fillSessions` adds an entry for each session or pass stop reason, with `todoFor` (`:54`). `Build` renders the snapshot from state alone (`snapshot.go:172`, `fillSessions` at `:210`), so every merge state it shows must be a field in the release file |
| The fixture snapshot | `e2e/testdata/state/snapshot.json` is compared byte for byte (`e2e/testdata/README.md`). New fields are `omitempty`, and no fixture's bytes change |

## File list

Hit counts are from `grep -rn` over `cmd` and `internal`, tests left out, on `73da88e`.

| File | Change | Hits |
| --- | --- | --- |
| `internal/state/types.go` | New optional `PullRequest` fields | `PullRequest` 3, `FixAttempts` 1, `Merged` 1 |
| `internal/state/check.go` | Validate the new fields | `PullRequest` 1 (`:228`) |
| `internal/state/merge.go` (new), `merge_test.go` (new) | The ready, wait, fix, ask, merge, and cleanup events | 0 |
| `internal/state/checkrun.go` | `RecordCheck` writes a pull request's wait | the record step |
| `cmd/cc-desk/pass.go` | The `pass ready` verb | `passCmd` 1 |
| `internal/launch/settings.go`, `launch_test.go` | The ready verb's allow rule: `WriteSettings` (`:67`) takes a list of prefix rules in place of one `followUp` string | `func WriteSettings` 1, `Deny = ` 1 |
| `cmd/cc-desk/launcher.go` | Pass launches pass the ready rule and fill `Prompt.Ready` | `FollowUpCommand` 1 (`:225`) |
| `internal/launch/prompt.go` | A `Ready` field beside `FollowUp` (`:25-27`), and the fix note | `FollowUp string` 1, `RelaunchNote` 3 |
| `internal/launch/templates/pass.txt` | The ready line | `Do not merge it` 1 |
| `internal/launch/templates/fix.txt` (new) | The fix note's text | 0 |
| `internal/launch/worktree.go` | `Remove`, with no force | `dropBranch` 4, `branch", "-D` 1 |
| `internal/sessions/github.go` (new), `github_test.go` (new) | Parse the read and bucket the checks | 0 |
| `internal/sessions/decide.go`, `merge.go` (new), `merge_test.go` (new) | The merge decisions | `readyStands` 3, `func early` 1 |
| `cmd/cc-desk/check.go` | `--gh-command`, the read in `gatherEvidence` | `probeURL` 3, `func gatherEvidence` 1 |
| `cmd/cc-desk/carryout.go`, `launcher.go` | The merge, the fix relaunch, the question, the cleanup | `func (l *launcher) carry` 2, `RelaunchNote` 1 |
| `internal/snapshot/sessions.go`, `snapshot.go` | The no-check entry and the cleanup note; `fillSessions` gains `now` | `func todoFor` 1, `fillSessions(` 2 |
| `e2e/merge_test.go` (new), `e2e/merge_rig_test.go` (new), `e2e/testdata/check-merge/` (new), `e2e/testdata/README.md` | The `gh` stand-in and done item 5 | 0 |

## What the pass builds

### The ready verb

`cc-desk pass ready <project> <pass> --pr <n> --commit <sha> --verdict accept|fix --design
inside|outside`. All four flags are required. It commits one event, `ReportReady`, which
refuses unless the pass is building and has a session record, so a stopped pass can't report.
The event sets the pull request's number, commit, verdict, inside-design flag, and `Ready` to
the clock, and clears its wait, failures, question, and merge marker. `FixAttempts` carries over. A second report replaces the first.

A pass launch adds one prefix allow rule beside the follow-up rule:
`Bash(<binary> --state-dir <dir> pass ready <project> <pass> *)`. Design and conductor launches
don't get it. `WriteSettings` takes its prefix rules as a list, and the launcher passes the
follow-up rule alone for design and conductor launches. `Prompt` gains a `Ready` field, the
command prefix, and `pass.txt` gains one line that uses it:
once the pull request is ready and reviewed, report it with that command, naming the PR, its
last commit, the review's verdict, and whether the change stays inside the approved design.
The line keeps "Do not merge it". The service start the design gives this verb is pass 5's
(Decisions).

### The read

`gatherEvidence` reads each live release that has a repo and a checkout. For each pass that is
building or stopped, has a session record with a branch, and has no recorded merge, it runs one
call through a hidden `--gh-command` flag (default `gh`):

`gh pr list -R <repo> --head <branch> --state all --json number,state,headRefOid,mergedAt,mergeCommit,statusCheckRollup`

A non-zero exit is "GitHub can't be reached" for that pass. Exit 0 with `[]` is "no pull
request". With more than one, the one whose number the ready report names wins, else the first.
`--dry-run` makes the same read and prints the decisions, and calls nothing else.

### The decision

For a building pass whose read came back, `decideMerge` runs before the row rules and decides
in this order. The first row that matches wins.

| # | When | Decision | Then the row rules |
| --- | --- | --- | --- |
| 1 | State `MERGED` | `merged`: recorded at `mergedAt`, outside cc-desk unless the merge marker is set (below) | Don't run |
| 2 | No ready report stands, or the state is `CLOSED` | none | Run as today |
| 3 | A merge question was asked under the current report (the pull request's `question` is set), answered or not | none, detail "question <id> asked" | Run, idle held off |
| 4 | The report says outside the design | `ask`, reason `outside_design` | Don't run |
| 5 | The head isn't the reported commit | wait `commit_moved` | Run, idle held off |
| 6 | The rollup is empty | wait `no_check` | Run, idle held off |
| 7 | A check is pending | wait `check_running` | Run, idle held off |
| 8 | A check failed or was cancelled, or the verdict is `fix`, and `FixAttempts` is under 2 | `fix`, uncounted, cause `fix` | Don't run |
| 9 | The same, with `FixAttempts` at 2 | `ask`, reason `fix_limit` | Don't run |
| 10 | Every check passed or was skipped, the verdict is `accept`, and the report says inside | `merge` | Don't run |

A pass that is stopped gets row 1 only. Every other stopped pass stays with pass 1's and pass
2's rows, PR 8's usage rows among them. A pass whose read failed gets wait `unreachable`, and
the row rules run. Each wait is written to the pull request's `wait` field in the check's
record step, so the snapshot can show it.

`early()` gains one row, placed first, before the `sess.Ended` test at `decide.go:186`: a
done pass whose pull request is merged and not yet cleaned up decides `cleanup`. Cleanup step 1
ends the session, so the row must come before that test for a cleanup that failed part way to
run again on the next check. Every other done pass keeps `none`. A release marked done drops
out of `LiveSessions` (`internal/state/check.go:418`), but no code sets a release done yet,
so that case is a follow-up.

### Carrying out

- **merge.** Commit the merge marker (`merge_called`, the clock) first. Then run `gh pr merge
  <n> -R <repo> --merge --match-head-commit <sha>`. On exit 0, commit `MergePass` at the check's
  clock and go on to the cleanup. On any other exit, the outcome reads `failed: merge: <last
  stderr line>`, the check exits 2, and the marker stays. On the next run, a `MERGED` read with
  the marker set is cc-desk's own merge, and an `OPEN` one is decided again.
- **merged.** Commit `MergePass` at `mergedAt`, with `merged_outside` set when no marker is,
  then the cleanup.
- **fix.** Stop the session when its row is live, then relaunch it uncounted with cause `fix`
  and a fix note naming the pull request, the commit, and the failed checks (or "the reviewer
  asked for a fix"). The note is a third kind beside `RelaunchNote` and `AnswerNote`, with its
  own template, `fix.txt`, which tells the session to fix the branch, push, and report ready
  again. `Fill` refuses a prompt with more than one note. One commit holds the relaunch and
  the fix: `FixAttempts` up by one, `Ready` cleared, and the failures saved. A failed stop or
  launch spends no attempt.
- **ask.** One commit adds a question with the id `merge-<pass>-<k>`, where `k` is one more
  than the pass's merge questions so far, and adds the id to the pass's `questions` and to the pull request's `question`. Its text
  names the pull request and the failures or "outside the approved design". Its options are
  `merge`, `fix again`, and `close`, with `close` recommended after the third failure and `fix
  again` for a change outside the design. The ready report stays, so the idle rule stays off and
  row 3 holds every later decision, after an answer too, until a new report clears `question`. What an answer does is a follow-up (Decisions).
- **cleanup**, after `MergePass`, in this order:
  1. Stop the session when its row is live, and commit `EndSession`.
  2. Pull the checkout with `git -C <checkout> pull --ff-only` when all three hold: it is on the
     release's default branch; no file the merge touched has an uncommitted change, found by
     `git -C <checkout> fetch origin <default>`, then `git diff --name-only
     HEAD...origin/<default>`, then `git status --porcelain -- <those files>`; and no live row's `cwd` is the checkout or a folder
     under it outside `.claude/worktrees/`. Otherwise note "checkout not pulled: <reason>".
  3. `git worktree remove <worktree>`, with no force flag. A refusal notes "worktree left:
     <the last line git printed>".
  4. `git branch -d <branch>`, only when the pull ran, with no force flag. A skip or a refusal
     notes "branch left: <reason>".
  5. Commit `RecordCleanup` with the notes joined, or none, and `cleaned` set to the clock.

  A note is a report line, not a failure: the check still exits 0. Only a git error before
  step 3's refusal can be read, such as a missing checkout, is `failed:`.

### What the snapshot shows

`fillSessions(snap, s)` becomes `fillSessions(snap, s, now)` (`snapshot.go:210`). A pull
request whose wait is `no_check` with `Ready` 30 minutes or more before the clock gives one
`needs_you_now` entry: reason `no_check_registered`, cause the PR number, todo "See why no
check has started on pull request <n>". That entry also sets `next_session.state` to
`needs_you_now` (`estimate.go:361`, `:432`), which is right: Geoff must act. A merge question
shows under `waits_for_geoff` through the existing estimate code. A cleanup note shows as a new
`cleanup` field, `omitempty`, on the pass session's `SessionEntry` (`sessions.go:20`). None of these change the bytes of any old fixture's snapshot.

## Acceptance

Copied from the design's "What done looks like", items 5 and 8, word for word.

5. **Merge.** Six fixture pull requests: clean; no check registered; a check running; a check failed; outside the design; last commit moved. Only the first merges, with a merge call pinned to its commit, and its pass reads done with the merge time. Its worktree and branch are gone, and the log holds no force flag. The failed one relaunches its session once, and a second check with no new report does nothing. After a new report on a new failed commit it relaunches again, and after a third it becomes a question. A pull request with no check after 31 minutes reads "Needs you now". The `gh` stand-in's log shows no merge call for the other five.

8. **The gate.** `bash scripts/gate.sh` passes, with no skipped test. Items 6 and 7 are not Go tests. Their script's output is in the round report, and a pass that leaves it out isn't done.

Items 6 and 7 are pass 5's.

| Item | Task | Check | On `73da88e` |
| --- | --- | --- | --- |
| 5, the six | 7 | `go test ./e2e/ -run TestMergeSix -count=1 -v` prints `--- PASS: TestMergeSix` | `testing: warning: no tests to run`, `ok github.com/glw907/cc-desk/e2e 0.330s [no tests to run]` for `-run TestMerge` |
| 5, the fix limit | 7 | `go test ./e2e/ -run TestMergeFixLimit -count=1 -v` prints `--- PASS: TestMergeFixLimit` | The same |
| 5, the 31 minutes | 7 | `go test ./e2e/ -run TestMergeNoCheck -count=1 -v` prints `--- PASS: TestMergeNoCheck` | The same |
| 5, the ready verb | 2 | `go test ./e2e/ -run TestPassReady -count=1 -v` prints `--- PASS: TestPassReady` | `pass ready --help` prints the group's help with `finish`, `start`, and `stop` only |
| 8 | 8 | `bash scripts/gate.sh`, bare, ends `gate: ok`; `go test ./... -count=1 -v 2>&1 \| grep -c -- '--- SKIP'` prints 0 | `gate: ok`, exit 0 |

What each test asserts:

- `TestMergeSix`: one release with six building passes, launched through the rig. For each,
  the test commits a file in the worktree and pushes the branch to the bare remote, as the
  session would. It writes one stand-in answer per branch, then reports ready through `pass
  ready`. It sets every session's stand-in row live and `done`, so the row rules leave the
  waiting passes alone. The check runs once, through a new `runEnv` helper in
  `merge_rig_test.go` that adds `GIT_TRACE=<file>` to the binary's environment only, and exits
  0. Then:
  - The stand-in's log holds exactly one `pr merge` call, for the clean PR, with `--merge` and
    `--match-head-commit <its head>`, and none of `--admin`, `-d`, `--delete-branch`, or
    `--auto`.
  - The clean pass reads `done`, its `finished` equal to the check's clock, its pull request
    merged at that time. Its worktree folder is gone, `git branch --list <branch>` in the
    checkout prints nothing, and the checkout's `main` holds the merge commit.
  - On the trace's `trace: built-in:` and `run_command:` lines, split on spaces, no token is
    `-D`, `-f`, or `--force`. A substring match would fail on `--ff-only`.
  - The failed one's session was relaunched once, uncounted, cause `fix`, and its prompt holds
    the fix note. The outside one has an open `merge-<pass>-1` question, and the snapshot
    counts it under `waits_for_geoff`. The other three have the waits `no_check`,
    `check_running`, and `commit_moved`, and no launch.
- `TestMergeFixLimit`: the failed PR from the start. A second check with no new report adds no
  launch and no merge call. A new report on a new failed commit relaunches again. A third
  report on a third failed commit adds the question, and the log holds no third relaunch.
- `TestMergeNoCheck`: a PR with an empty rollup. At 29 minutes after its report the snapshot's
  `needs_you_now` holds no `no_check_registered` entry; at 31 minutes it holds one, and
  `next_session.state` reads `needs_you_now`.
- `TestMergeOutside` and `TestMergeDirty` cover a PR merged by another hand (recorded outside,
  finished at `mergedAt`, cleaned up) and a worktree with an uncommitted file (it stays, the
  branch stays, and the note shows in the snapshot). `TestMergeUnreachable` covers a stand-in
  that exits 1: no merge, wait `unreachable`, and the row rules still decide.

## Tasks

Each task commits on its own, runs `bash scripts/gate.sh` bare before the commit, and is
pushed. Each accepted task's report names the gate check's verdict on its pushed commit. No
task holds code. Each test is table-driven where it has cases, and every Go file follows
`standards:go-conventions`.

### Task 1: the state events (15 minutes)

- Outcome: the pull request fields and the merge events exist and are tested.
- Files: `internal/state/types.go`, `check.go`, `merge.go` (new), `merge_test.go` (new),
  `checkrun.go`.
- Class: `runner`.
- Check: `go test ./internal/state/ -run 'TestReportReady|TestMergePass|TestAskMerge|TestRecordFix|TestRecordCleanup' -count=1 -v`
  prints `--- PASS:` for each of the five names.
- Test:
  - New `PullRequest` fields, each `omitempty`: `inside_design`, `wait`, `failures`,
    `question`, `merge_called`, `merged_outside`, `cleaned`, and `cleanup_note`. Validation refuses a
    verdict other than `accept` or `fix`, a `Ready` with no commit or verdict, a wait outside
    the four, and a `merged` time on a pass that isn't done.
  - `ReportReady` as "The ready verb" says, refusing a pass that isn't building (a stopped
    pass among the cases) or has no session.
  - `MergePass(project, pass, at, outside)` closes an open stop at `at`, then finishes the pass
    at `at` with its actual minutes and sets `merged`. It refuses a pass not started or already
    done.
  - `AskMerge`, `RecordFix`, `MarkMerge`, and `RecordCleanup` as "Carrying out" says. `RecordFix`
    composes with `RecordLaunch` in one commit.

### Task 2: the ready verb (10 minutes, a pair with task 3)

- Outcome: a pass session can report its pull request ready, and only through its own verb.
- Files: `cmd/cc-desk/pass.go`, `launcher.go`, `internal/launch/settings.go`, `prompt.go`,
  `launch_test.go`, `templates/pass.txt`, `e2e/merge_test.go` (new, `TestPassReady` only).
- Class: `runner`.
- Check: `go test ./e2e/ -run TestPassReady -count=1 -v` prints `--- PASS: TestPassReady`, and
  `go test ./internal/launch/ -count=1` passes.
- Test: the verb writes the four fields and `Ready`; each missing flag is refused by cobra; a
  pass that isn't building is refused with exit 1. The settings file holds the prefix rule, the
  deny list is unchanged, a design launch's settings hold no ready rule, and the first-launch prompt holds the ready line and still holds "Do
  not merge it". `launch_test.go` gains the rule in its expected list; no other old test
  changes.

### Task 3: the read (10 minutes, a pair with task 2)

- Outcome: the check reads each pass's pull request in one call and buckets its checks as `gh`
  does.
- Files: `internal/sessions/github.go` (new), `github_test.go` (new), `cmd/cc-desk/check.go`.
- Class: `runner`.
- Check: `go test ./internal/sessions/ -run 'TestParsePulls|TestBucket' -count=1 -v` prints
  `--- PASS:` for both.
- Test: the parse of the four recorded outputs in the mechanism table (merged, open with a
  passing run, no checks, `[]`), a `StatusContext` entry, and each state in the bucket table.
  `--gh-command` is hidden, defaults to `gh`, and the read runs outside the lock.

### Task 4: the decisions (15 minutes)

- Outcome: `Plan` gives the merge table's decisions and the cleanup row.
- Files: `internal/sessions/decide.go`, `merge.go` (new), `merge_test.go` (new).
- Class: `runner`.
- Check: `go test ./internal/sessions/ -run 'TestDecideMerge|TestCleanupRow' -count=1 -v`
  prints `--- PASS:` for both, and `go test ./internal/sessions/ ./e2e/ -count=1` passes with no
  old test changed.
- Test: one case per row of the merge table, row order included (an outside report with a
  failed check asks; a moved head with a failed check waits), the stopped pass that gets row 1
  only, the unreachable read, an answered merge question that still holds row 3, and a done
  pass with and without a merged, uncleaned pull request, with its session ended and not.

### Task 5: carrying out (20 minutes)

- Outcome: the check merges, relaunches for a fix, asks, and cleans up as "Carrying out" says.
- Files: `cmd/cc-desk/carryout.go`, `launcher.go`, `internal/launch/prompt.go`,
  `templates/fix.txt` (new), `worktree.go`, `launch_test.go`.
- Class: `runner`.
- Check: `go test ./internal/launch/ -run 'TestFixNote|TestRemove' -count=1 -v` prints `---
  PASS:` for both, and `go test ./... -count=1` passes.
- Test: `Fill` with a fix note, and its refusal with two notes. `Remove` removes a clean
  worktree and a merged branch, and leaves a dirty worktree and an unmerged branch with their
  notes, all in a `t.TempDir()` repo. The merge path is covered end to end in task 7.

### Task 6: the snapshot (5 minutes)

- Outcome: the 31-minute entry and the cleanup note show.
- Files: `internal/snapshot/sessions.go`, `snapshot.go`, their tests.
- Class: `runner`.
- Check: `go test ./internal/snapshot/ ./e2e/ -count=1` passes, and `git diff --stat
  73da88e -- e2e/testdata/state/snapshot.json` prints nothing.
- Test: the entry at 29 and 31 minutes, and `next_session.state` at each; the note on a
  cleaned pass.

### Task 7: done item 5 (25 minutes)

- Outcome: item 5 passes end to end against the stand-ins.
- Files: `e2e/merge_rig_test.go` (new), `e2e/merge_test.go`, `e2e/testdata/check-merge/`
  (new), `e2e/testdata/README.md`.
- Class: `runner`.
- Check: the five `go test ./e2e/ -run <name>` commands under `## Acceptance` each print their
  `--- PASS:` line.
- Test: as "What each test asserts" says. The `gh` stand-in is a Python file beside the
  `claude` one. It logs each call's argument list as one JSON line, answers `pr list --head <b>`
  from that branch's answer file or `[]`, and exits 1 when its `down` control file exists. On
  `pr merge <n> ... --match-head-commit <sha>`, it exits 1 when the SHA isn't the answer's head.
  Otherwise it merges the branch into `main` on the bare remote with a real `git merge
  --no-ff` in its own clone, pushes, and rewrites the answer as `MERGED` with `mergedAt`. It
  clears `GIT_TRACE` for its own git calls, so the trace holds only the binary's.

### Task 8: close (10 minutes)

- Outcome: the PR is ready, with its review accepted and CI green.
- Files: `docs/STATUS.md` (below the round-report block), and one file each under
  `docs/follow-ups/`: what an answer to a merge question does; the ready verb's service start,
  for pass 5; a pull request closed without a merge; and the cleanup of a release marked done.
- Class: `ordinary`.
- Check: the PR body's `## Acceptance` evidence; the `gate` check on the last commit.
- Test: none of its own. One whole-branch `standards:diff-reviewer` read (class `runner`),
  folded. A Low finding from a later delta read is folded once, checked by the executor's own
  diff read and CI, or filed; it never opens another review round.

Path: task 1, tasks 2 and 3 as a pair, task 4, task 5, task 6, task 7, task 8: 15 + 10 + 15 +
20 + 5 + 25 + 10 = 100 minutes. The tasks sum to 110. At the last rounds' pace of about 0.6, the
path is about 60 minutes.

The CI watch, run bare after each push: `gh run list --repo glw907/cc-desk --branch
sessions-merge --limit 3`, then `gh run watch <id> --repo glw907/cc-desk --exit-status` on the
newest.

## Standing answers

The twelve in `round-run.md` hold, with the round plan's changes and its answers 13 to 20.
Pass 1's 23 to 29, pass 2's 30 to 38, and pass 3's 39 to 46 hold, with 120 minutes as the
ceiling in answer 39. The last round's 19 to 22 hold too. This pass adds:

- **47. A test would run the real `gh`.** Every e2e call to `check` passes `--gh-command` with
  the stand-in. The executor never runs `gh pr merge`, `gh pr checks`, or a `gh pr view` of
  another pass's PR, even to probe. The mechanism table holds the recorded outputs.
- **48. `git branch -d` refuses in a test.** The pull didn't land first, or the stand-in's merge
  isn't real. Fix the order or the stand-in. Never switch to `-D`, and never call `dropBranch`
  from the cleanup.
- **49. `git worktree remove` refuses on a clean worktree in a test.** Find the file that makes
  it dirty with `git -C <worktree> status --porcelain --ignored`, and fix whatever wrote it.
  Never add `--force`.
- **50. A field or a reason name doesn't fit.** The names here are the builder's to change, as
  long as the JSON names stay `omitempty` and the snapshot bytes of old fixtures don't change.
- **51. A row of the merge table seems wrong for a case the design names.** The design's
  "Merging" table wins. Record the change under `## Decisions` as an `[advisor]` line.
- **52. A rig change is needed.** Add `e2e/merge_rig_test.go` beside the old rig, with its own
  helper that adds `--gh-command` to `launchRig.args`' list, and change no old helper.
- **53. The pair's merge conflicts.** Tasks 2 and 3 touch different files. A conflict in
  `check.go` or `pass.go` means one task strayed; keep task 3's `check.go` and task 2's
  `pass.go`, and rerun both checks.

## Decisions

- [advisor] One read per pass: `gh pr list --head <branch> --state all` with the rollup. Exit 0
  always means GitHub answered, `[]` means no PR, and an empty rollup means no check. `gh pr
  checks` exits 1 on zero checks (probed), and `gh pr view <branch>` exits 1 on no PR, so
  either would blur "no check" or "no PR" into "unreachable". Reading by branch also finds a PR
  merged by another hand before any ready report.
  Call: mine (method).
- [advisor] Skipped and neutral checks count as passed, and a cancelled check as failed. This
  follows how GitHub's required checks treat them; the design says "every registered check has
  passed" and names no bucket.
  Call: mine (method).
- [advisor] The merge marker, committed before the merge call, tells cc-desk's own merge from
  one by another hand when the check crashes between the call and its record. Without it, that
  crash would count a cc-desk merge in the round report as an outside one.
  Call: mine (method).
- [advisor] The merge time is the check's clock for cc-desk's own merge, and `mergedAt` for any
  merge found on a read. `MergePass` takes the time, closes an open stop, and finishes the
  pass, since `FinishPass` uses the event clock and refuses a stopped pass.
  Call: mine (method).
- [advisor] `PullRequest.Ready` is the 30-minute clock. The design's wait starts at the report,
  and the state design's rule is about the reported commit, so no new time field is needed.
  Call: mine (method).
- [advisor] Waits let the row rules run with idle held off, so a session that blocks or dies
  while its PR waits is still caught. Merge, fix, ask, and merged replace the row decision for
  that run.
  Call: mine (method).
- [advisor] The question's options and recommendations (`merge`, `fix again`, `close`) are
  mine, since the state model needs them and the design gives none. What an answer does is out
  of this pass's scope and goes to a follow-up for pass 5 or later; until then a ruling only
  marks the question answered, and row 3 holds every decision while it is open.
  Call: mine (method), filed as a follow-up.
- [advisor] This builds on pull request 12's flagged gap, "no verb registers a desk": a merge
  question is answered through `ruling record`, which needs a desk. Round plan answer 17 has
  pass 5's conductor register as one. This pass adds no verb under the ruling guard.
  Call: mine (method), names a flagged call.
- [advisor] This builds on pull request 8's flagged usage relaunch call only at its edge: a
  pass stopped on usage gets merge row 1 alone, and every other stopped-pass decision stays
  with PR 8's rows unchanged.
  Call: mine (method), names a flagged call.
- [advisor] The design has the ready verb also start the service. The unit and timer are pass
  5's, so this pass defers that start and files a follow-up for it.
  Call: mine (method, scope as the round plan sets it).
- [advisor] The path is 100 minutes against 60 planned, under the 120-minute park line. The
  design's cases are kept whole rather than trimmed to fit.
  Call: mine (method).

## Review findings and their folds

One `standards:plan-reviewer` read (Opus, medium), 2026-11-18: 2 High, 6 Medium, 4 Low. I
checked each against the code before folding. All twelve are folded.

1. High. The cleanup row could never run again after step 1, since `early()` returns at
   `sess.Ended` first. Folded: the row goes first in `early()`, with a test case for an ended
   session. The done-release half is filed for task 8: no code sets a release done yet.
2. High. The trace check matched `-f` inside `--ff-only`. Folded: tokens on `trace: built-in:`
   and `run_command:` lines only.
3. Medium. The touched-files diff had no commit to diff against. Folded: `mergeCommit` joins the
   read, and step 2 names the fetch, diff, and status commands.
4. Medium. An answered merge question let the next check ask again, without end. Folded: a
   `question` field that only a new report clears, and row 3 matches it answered or not.
5. Medium. Task 2's files left out `launcher.go` and `prompt.go`. Folded: both listed, with the
   `WriteSettings` list change, `Prompt.Ready`, and pass launches only.
6. Medium. `TestMergeSix` could relaunch a waiting pass with no stand-in row. Folded: every row
   set live and `done`, and the exit code asserted.
7. Medium. No helper set `GIT_TRACE` for the binary alone. Folded: `runEnv` in
   `merge_rig_test.go`.
8. Medium. The snapshot change named no signature, field, or todo. Folded: all three, and
   `next_session.state` asserted.
9. Low. The `gh pr merge` call wasn't checked for force flags. Folded.
10. Low. The marker from a failed merge call survived a new report. Folded: `ReportReady` clears it.
11. Low. A report on a stopped pass was unstated. Folded: refused, with a test case.
12. Low. Two hit counts were off. Folded: recounted, with `WriteSettings` listed.
