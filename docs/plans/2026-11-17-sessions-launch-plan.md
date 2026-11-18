# Launching: pass plan

Initiative: cc-desk
Round: sessions
Pass: sessions-launch

Class `runner`. Planned 90 minutes; the path below is 160. Design:
`docs/specs/2026-11-17-sessions-design.md` (approved 2026-11-17), sections "What the command
records", "Launching", "The check", "When something goes wrong", "Left to the builder", and the
second row of "The passes". Round plan: `docs/plans/2026-11-17-sessions-round.md`. Pass 1's plan,
`docs/plans/2026-11-17-sessions-check-plan.md`, holds the decision table this pass carries out.
Writing a person reads (the PR body, STATUS lines, follow-ups) follows
`~/.claude/docs/voice/written-for-people.md`, loaded through the `writing-voice` skill.

Worktree `/var/home/glw907/Projects/cc-desk/.claude/worktrees/sessions-launch`, branch
`sessions-launch`, cut from `origin/sessions-check` at `31dc089` (pull request 8, open, waiting
for Geoff's approve on one call). PR base: `main`. Push: `glw907/cc-desk`. Before the PR is
ready, merge `origin/sessions-check` in again if it moved, and `origin/main` if it moved.

Scope is the design's "Launching" and the actions of "The check" that launch, end, and relaunch.
The command makes the worktree, writes the settings file, fills the prompt template, and runs
the launch through `systemd-run --user --scope`. The check carries out the decisions pass 1
records, in the order end, confirm, launch. A desk session starts and finishes a conductor or a
design writer through a new verb. Merging, prompt recording, and follow-ups are passes 3 and 4.
This pass edits none of `scripts/gate.sh`, `.github/workflows/`, or `web/`, which are the Page
round's.

## Pre-flight

No step in this pass needs Geoff. Every write is inside the worktree or a test's `t.TempDir()`.
No test runs the real `claude` or `systemd-run`, and no test reaches the network. Git is real,
in a repo and a bare remote each test makes.

- Permissions checked: `~/.local/state/claude-denials.jsonl`, read again 2026-11-18, 130 rows,
  2026-11-01T00:55:05Z to 2026-11-18T06:27:23Z, all inside the 30-day window and unchanged since
  pass 1's read. Classes and rows: Classifier unavailable 47; Destructive or Unwanted Actions 14;
  Self-Modification 11; Interfere With Workloads 11; server-side classifier, no reason given 10;
  Undermining Oversight 5; Unauthorized Access 4; Merge Without Review 3; Auto-Mode Bypass 3;
  Irreversible Local Destruction 2; Git Destructive 2; Create Public Surface 2; Unverifiable
  Deletion Scope 2; Sensitive-Source Provenance 2; Security Weaken 2; Production Deploy 1;
  Secret-Store Writes 1; DNS / Domain / Cert Changes 1; Bot Check Circumvention 1; Code from
  External 1; Shared Scratch Sweep 1; Safety Bypass Flag 1; Out-of-Place Publication 1; PII Data
  Handling 1; Modify Shared Resources 1. Rows naming this pass's commands: `systemd-run` 0;
  `git worktree` 4 (chained, 2026-11-02 to 11-13); `git fetch` 3 (chained); `claude --bg` 1 (a
  classifier outage, 2026-11-17); `claude stop` 5 (all chained); `go test` and `go build` 1 (a
  chained `kill`); `gate.sh` 0; `git push` 4 (chained or destructive); `gh pr create` and
  `gh pr ready` 0; a Claude settings file named in the target 10 (Self-Modification and the
  server classifier). The alternative for each is in the table below.
- Reviewed: one `standards:plan-reviewer` read of this plan (Opus, medium), as the round plan's
  Pre-flight asks of every pass. Its findings and how each was folded are under `## Decisions`.
- Evidence: the mechanism table and the recorded searches below. Each acceptance check was run on
  today's code before any code was written; the outputs are under `## Acceptance`.

### Refused commands and their alternatives

| Refused in the log | This pass does instead |
| --- | --- |
| `claude --bg` from a session | Only the built binary launches, and only in a test, against the `claude` stand-in. No task launches a real session |
| A chained `claude stop` | The binary runs `claude stop <id>` alone, through `exec.Command`. No task stops a real session |
| `claude agents --json` piped into a script | Tests use the stand-in. A live look, if ever needed, redirects to a file in `$CLAUDE_JOB_DIR/tmp` |
| A chained `git worktree add` or `remove` | The binary runs each git command alone, in a test's repo. The executor runs no `git worktree` but the ones `EnterWorktree` and the close make |
| A chained `git fetch` or `git push` | Fetch and push `sessions-launch` alone |
| A write to a Claude settings file | The executor edits no settings file. The binary writes each session's file under the state folder, `<state>/launch/<name>.json`, and in tests that folder is a `t.TempDir()` |
| `kill` or `pkill` around `go test` | Stop a background command with `TaskStop` |
| `git commit --no-verify`, `git stash` | Commit by path with hooks on. No stash; a temporary commit sets work aside |
| `gh pr merge` | The pass merges nothing. It ends with the PR ready, and the conductor lands it |
| A classifier outage | Wait 60 seconds and run it once more, then flag the task and go on |

## Mechanism table

| Mechanism | Probe or source |
| --- | --- |
| The verbs today | `go build -o $CLAUDE_JOB_DIR/tmp/cc-desk ./cmd/cc-desk`, then `cc-desk conductor start x`: `unknown command "conductor" for "cc-desk"`, exit 1. `cc-desk round start x y`: `unknown command "start" for "cc-desk round"`, exit 1. `cc-desk --help` lists `check`, `free`, `pass`, `release`, `round`, `ruling`, `snapshot`, `usage` |
| `pass start` records and launches nothing | The `writers` fixture copied to a temp folder: `cc-desk pass start east e-1` prints `pass "e-1" is building and cannot start`, exit 2. The release holds `e-1` to `e-7` building with no `session` key and no `checkout` |
| No fixture has a checkout | A scan of every `e2e/testdata/*/releases/*.json`: none carries `checkout`, the three `check*` fixtures included. Only `writers` holds building passes with no session (seven in each of `east` and `west`) |
| The old tests call `pass start` | `grep -rnE '"pass", "start"' e2e`: about 25 calls in `estimate_test.go`, `events_test.go`, `order_test.go`, `ruling_test.go`, `lock_test.go`, and `broken_test.go`, all on releases with no checkout. So `pass start` stays a recording verb (Decisions) |
| The start refusal text | `internal/state/events.go:205` `pass %q is %s and cannot start`. Tests assert on substrings only: `start_test.go:162` "is waiting and cannot start", `events_test.go:178` a done pass refused |
| The claude flags | `claude --help > $CLAUDE_JOB_DIR/tmp/claude-help.txt` on Claude Code 2.1.327 lists `--bg, --background` ("Prints the id that `claude attach`, `logs`, `stop` and `rm` take"), `-n, --name`, `--model`, `--effort`, `--permission-mode`, `--settings <file-or-json>`, and `--disallowedTools <tools...>` |
| What a launch prints | `spikes/*/out/launch*.txt`: `backgrounded · 0c026eb1 · ccd-spike-2-idle`, and eight more of that shape. One line, from a `resume`, reads `backgrounded · e3b28ac7` with no name. The short id is the token after `backgrounded · ` |
| The session id | `spikes/own.sh:146-165` saves the list before and after the launch and takes the one new row with the name, giving `id` and `sessionId`. A job's `~/.claude/jobs/<short id>/state.json` also carries `sessionId` (its keys read on this machine: `backend`, `children`, `cliVersion`, `createdAt`, `cwd`, `detail`, `name`, `sessionId`, `state`, and 15 more) |
| A failed launch | `pass-core` `references/planner.md` "Launch": "`needs: \"send a prompt to start\"` is a failed launch" |
| The prompt before the tools | `docs/research/2026-11-17-prove-findings.md`, proof 3: `--disallowedTools` "takes every word after it"; leaving out Artifact, SendFeedback, Workflow, AskUserQuestion, and ReportFindings saved about 19,000 tokens a request |
| Auto mode and allow rules | `docs/research/2026-11-17-prove-findings.md`, proof 4: auto mode with `--settings` allow rules ran a pass with zero prompts, and `approval-queue land` ran with no rule, so the deny rules are needed. Spike settings shape: `spikes/4-permissions/settings.json` `{"permissions": {"allow": ["Bash(git -C scratch *)", ...]}}` |
| The deny rule form | [Configure permissions](https://code.claude.com/docs/en/permissions): "Bash(ls *) matches ls", so a space before a trailing `*` also matches the bare command; `Bash(ls*)` with no space matches `lsof` too. A deny rule "isn't a security boundary around the program": `claude-merge-guard` stays the merge's guard. So the rules are `Bash(gh pr merge *)` and `Bash(approval-queue land *)`. A bare `gh pr merge` merges the current branch's PR, which the space form catches |
| The hook entry | `spikes/4-permissions/answer-settings.json`: `{"hooks": {"PermissionRequest": [{"matcher": "Bash", "hooks": [{"type": "command", "command": "...", "timeout": 150}]}]}}`. [Hooks reference](https://code.claude.com/docs/en/hooks): a `PermissionRequest` hook that exits 0 with no output makes no decision, and the call goes through the normal permission flow. Proof 4: with no hook, a background session's prompt shows and the row reads `waitingFor: "permission prompt"`, which pass 1's check stops on. So the stub verb that prints nothing leaves every prompt as it is today |
| The scope | `systemd-run --version`: systemd 259. `systemd-run --help` lists `--user`, `--scope`, `-q --quiet`. `systemd-run --user --scope -q /bin/sh -c 'pwd; echo "args: $*"' sh "a prompt" --disallowedTools X` printed its cwd and `args: a prompt --disallowedTools X`, exit 0. Run with no `--working-directory` from a folder, `/bin/pwd` printed that folder: the scope keeps the caller's cwd, so `exec.Cmd.Dir` sets the worktree. Round plan, shared assumptions: the scope outlives the starting unit (probed 2026-11-17) |
| `git worktree add` | A probe repo with a bare remote, in `$CLAUDE_JOB_DIR/tmp/wt`: an add into an existing empty folder succeeds (exit 0), so git won't refuse an empty folder and the command must stat the path first. An add whose branch exists fails, `fatal: a branch named 'p2' already exists`, exit 255. An add into a non-empty folder fails, `fatal: '.claude/worktrees/p5' already exists`, exit 128, and leaves the branch `p5` it made. So a failed add can leave a branch behind, and the command cleans up only a branch this call made |
| Tracking | The same probe: a plain add prints "branch 'p1' set up to track 'origin/main'". With `--no-track` it prints no tracking line; a bare `git push` then fails with "has no upstream branch", and `git push -u origin p6` sets `origin/p6`. So the add uses `--no-track`, and the prompt names `git push -u origin <branch>` |
| The validators | `internal/state/check.go:219` `validateSession` requires `name`, `short_id`, `session_id`, `branch`, an absolute `worktree`, a nonzero `launched`, and `launch_tries` of 1 or 2. `check.go:421` `checkSessionsUnique` refuses two live records naming one worktree or one session name. So a conductor and a design writer each get a worktree of their own, never the main checkout |
| The commit path | `internal/state/commit.go` `Commit(dir, now, wait, ev, render)` and `WithReading(u, forced, ev)`: fold the reading, `settle`, then `ev`. A waiting pass that `settle` starts becomes building with no session inside the check's record commit, so the launch decision is made in that same event, after `settle` |
| The two-pass slot | `internal/state/rules.go:188` `Building()` counts building passes only. A relaunch of a building pass changes no pass state, so it keeps its slot. A usage-stopped pass is `stopped` and holds no slot |
| The check today | `cmd/cc-desk/check.go:36-37` Long help: "A launch, an end, or a relaunch is decided and printed, never carried out." `check.go:101-120`: the record commit, then one JSON line per decision. `check.go:20-27` `checkFlags{dryRun, claude, jobsDir, probeURL, secondWait}` |
| What pass 1's check tests assert after a run | `e2e/check_test.go:259` and `:331`: `releases/chk.json` holds no `"relaunches"`. `:438` `"launch_tries": 1` and `:469` `"launch_tries": 2` unchanged. `:271`, `:281`, `:328`: relaunch and end-and-relaunch decisions. Their release has no checkout and their worktrees, `/w/chk/s01` to `/w/chk/s11`, don't exist |
| The decision code | `internal/sessions/decide.go:130-165` `early`: an ended record or a done pass is `none`; a finished conductor or design writer is `end`; a usage stop is `wait_usage` under the hold, otherwise an uncounted `relaunch`. `:194-212`: a missing row gives `launch_again`, `stop launch_failed`, `stop relaunch_limit`, `wait_usage`, or a counted `relaunch` |
| The needs-you-now entry | `internal/snapshot/sessions.go:85-120`: a session with a stop gives an entry with its round and cause. A pass stopped with no session gives an entry with `session` null and its round. `todoFor(launch_failed)` reads "read the launch error, then start the pass again" |
| Stand-ins | `e2e/check_test.go:43` `claudeStandIn`, a heredoc script with `agents` and `logs` cases and no `--bg` or `stop` case. `e2e/main_test.go` `runWith` sets `PATH=/usr/bin:/bin`, where `git` and `systemd-run` live, so every launch test names both stand-ins and leaves git real |
| The gate | `bash scripts/gate.sh` in the worktree, 2026-11-18: gofmt, go vet, go test (e2e 38.5 s), vale, `gate: ok`, exit 0 |
| CI | `.github/workflows/gate.yml` runs the gate on each push. The watch: `gh run list --repo glw907/cc-desk --branch sessions-launch --limit 3`, then `gh run watch <id> --repo glw907/cc-desk --exit-status` |

## Recorded searches

`grep -rnE <term> --include=*.go .` in the worktree, 2026-11-18, `spikes/` left out:

| Search | Hits | Each hit |
| --- | --- | --- |
| `Checkout` | 14 | `types.go:146,148`: the field, beside which `AllowRules` goes (task 2). `events.go:52`: `AddRelease` refuses a checkout; it now accepts `allow_rules`, left as is for the checkout. `events.go:291-308` `SetCheckout`: left. `check.go:259-263`: left. `release.go:42` and `events_test.go:233-250`: left |
| `LaunchTries` | 8 | `decide.go:198`: read, left. `check.go:236,237`: left. `types.go:84`: left. The four test hits build records: left |
| `Relaunches` | 8 | `types.go:85,104` and `check.go:239`: left. `snapshot/sessions.go:30,91` shows `Counted()`: left. Three test hits: left |
| `AddCommand` | 8 | `root.go:120` gains `newSessionCmd` and `newPromptCmd` (task 5). `pass.go:52,54,82`, `release.go:16,51`, `round.go:16`, `ruling.go:54`: left |
| `never carried out` | 1 | `check.go:37`: the Long help, rewritten (task 4) |
| `claudeStandIn` | 3 | `check_test.go:40,43,70`: left. The launch tests get a stand-in of their own (task 1); answer 16 keeps the old one byte for byte |
| `Session != nil\|Session == nil` | 9 | `events.go:48,56,129`: the new-release and design refusals, left. `snapshot/sessions.go:112`: left. Five test hits: left |
| `cannot start` | 6 | `events.go:205` gains the live record's name when the pass has a session (task 2). `events.go:185,187,213`: left. `start_test.go:74,162`: left, since they match substrings |
| `StopLaunchFailed` | 5 | `decide.go:199`, `snapshot/sessions.go:64`, `types.go:36,42`: left. The check stops a pass with it when a launch fails (task 4) |
| `DryRun\|dryRun` | 6 | `check.go:23,41,87,101,111`: the dry run still carries out nothing (task 4). `check_test.go:548`: left |
| `decodeStrict` | 6 | `read.go:35,54` reads a release strictly, so `allow_rules`, `model`, and `effort` must be real fields before any input carries them (task 2) |

## File list

| File | Task | Change |
| --- | --- | --- |
| `e2e/launch_rig_test.go` (new) | 1 | The git repo with a bare remote, the `claude` and `systemd-run` stand-ins, and the rig that runs the binary with both |
| `internal/state/types.go`, `check.go` | 2 | `Release.AllowRules`, `Pass.Model` and `Pass.Effort`, `Decision.Outcome`; strict checks on them |
| `internal/state/events.go`, `checkrun.go` | 2 | The refusal names the live record; events that record a launch, a relaunch, an end, and a finish |
| `internal/launch/` (new): `worktree.go`, `settings.go`, `prompt.go`, `run.go`, `templates/*.txt` | 3 | The worktree, the settings file, the prompt from an embedded template, the launch and its ids |
| `internal/sessions/decide.go` | 4 | A `launch` decision for a building pass with no session |
| `cmd/cc-desk/check.go` | 4 | Carries the decisions out after the record commit, and its help says so |
| `cmd/cc-desk/session.go`, `prompt.go` (new), `root.go` | 5 | `session start` and `session finish` for a conductor and a design writer; `prompt hook`, the stub |
| `e2e/launch_test.go` (new) | 6 | Done item 1 |
| `e2e/recovery_test.go` (new) | 7 | Done item 2 |
| `docs/STATUS.md`, `docs/follow-ups/*.md` | 8 | The close |

## What the pass builds

### Records

A release gains `allow_rules`, a list of exact Claude Code permission rules. It can arrive in the
input file at `release add`; nothing in this pass changes it afterward, and pass 3 adds the guarded
verb. A pass gains optional `model` and `effort`, which override Opus at high for its session. A
check decision gains `outcome`, empty until the carry-out writes what happened.

The launch writes the session record in one commit after `claude --bg` returns: the name, kind,
model, effort, both ids, worktree, branch, and launch time. A first launch sets `launch_tries` to
1. A launch again sets it to 2 and keeps `last_seen` empty. A relaunch appends `{time, cause,
counted}` to `relaunches`, sets `launch_tries` to 1, and clears `last_seen` and `last_state`, so
pass 1's never-started rule applies to each new launch. The pass's `started` stays as `settle` or
`pass start` set it; the launch records no second start.

### Names and places

| Kind | Session name | Worktree | Branch | Model, effort |
| --- | --- | --- | --- | --- |
| Pass | `<project>-<pass>` | `<checkout>/.claude/worktrees/<pass>` | `<pass>` | `opus`, `high`, or the pass's override |
| Design writer | `<project>-<round>-design` | `<checkout>/.claude/worktrees/<round>-design` | `<round>-design` | `opus`, `high` |
| Conductor | `<project>-conductor` | `<checkout>/.claude/worktrees/conductor` | `conductor` | `sonnet`, `medium` |

The settings file is `<state>/launch/<session name>.json`, written fresh at every launch:

```json
{
  "permissions": {
    "allow": ["<each of the release's allow_rules>", "Bash(<binary> --state-dir <state> <each report verb>)"],
    "deny": ["Bash(gh pr merge *)", "Bash(approval-queue land *)"]
  },
  "hooks": {
    "PermissionRequest": [
      {"hooks": [{"type": "command", "command": "<binary> --state-dir <state> prompt hook", "timeout": 30}]}
    ]
  }
}
```

`<binary>` is `os.Executable()`. The report verbs are the exact commands the prompt names: for a
conductor or design writer, its `session finish`; for a pass, `pass stop <project> <pass>
--reason <r>` for each reason a session may report. The hook has no matcher, so every tool's
prompt reaches it. `prompt hook` reads its input to the end, prints nothing, and exits 0, which
leaves the prompt to the normal flow. It records nothing; pass 3 makes it record and answer.

### The launch

The launch, in order, with each command run alone:

1. Refuse before any side effect when:
   - the release has no checkout recorded;
   - another live record names the worktree or the session name (the error names that record);
   - the worktree path exists and no record names it;
   - for a first launch, the branch exists.
2. On a first launch:
   1. `git -C <checkout> fetch origin <default>`.
   2. `git -C <checkout> worktree add --no-track <path> -b <branch> origin/<default>`.

   A relaunch or a launch again reuses the recorded worktree. It first checks that
   `git -C <path> rev-parse --abbrev-ref HEAD` prints the recorded branch, and a mismatch or a
   missing folder fails the launch.
3. Write the settings file.
4. Fill the template for the kind, plus the relaunch paragraph on a relaunch.
5. From the worktree, run:

   ```sh
   systemd-run --user --scope -q claude --bg --name <name> --model <m> --effort <e> \
     --permission-mode auto --settings <file> "<prompt>" --disallowedTools Artifact SendFeedback \
     Workflow AskUserQuestion ReportFindings
   ```
6. Take the short id from `backgrounded · <id>` in its output. Take the session id from the row
   with that id in one `claude agents --json --all` read, or else from the job's `state.json`
   `sessionId`. Fail the launch when its `needs` reads `send a prompt to start`.
7. Commit the record.

A failure after step 2 on a first launch undoes what this call made:

1. `git -C <checkout> worktree remove <path>`, with no force flag.
2. `git -C <checkout> branch -D <branch>`, only while the branch tip still equals
   `origin/<default>`.

A failure after step 5 has also started a session, so the launch first runs `claude stop <id>`
on that session. A failed undo leaves the folder and names it in the error.

The templates live in `internal/launch/templates/` as `pass.txt`, `design.txt`, `conductor.txt`,
and `relaunch.txt`, embedded with `embed`, and filled with `text/template`. Each names:

- the session, its kind, the project, and the round;
- the worktree, the branch, the PR base, and the push command;
- `pass-core` with its role reference (`references/executor.md` for a pass, `references/conductor.md` for a conductor, `references/planner.md` for a design writer) and `references/round-run.md`;
- the report verbs;
- `~/.claude/docs/voice/written-for-people.md` for what it writes for a person;
- "Until cc-desk runs": a stop goes under `## Stop` in `docs/STATUS.md` on the branch, and no session runs `approval-queue send`.

A pass session is told to plan its pass first, then build it. Its PR ends ready and is not merged,
since the merge is the check's. These lines stay consistent with `planner.md` "Launch": the
prompt carries the goal, the scope, the plan path, the checkout, and the push target. The
relaunch paragraph reads: "This is relaunch <n> of 2, after <cause>. Carry on from the branch as
it stands; the commits on it are yours."

### The check carries out

The record commit stays as pass 1 built it. Inside the same event, after `settle`, `Decide` adds
a `launch` decision for each building pass with no session record: `wait_usage` under the usage
hold, otherwise `launch`. After that commit, still under `check.lock` and outside the state lock,
the check carries out each decision in the order printed:

| Decision | What runs |
| --- | --- |
| `launch` | The launch, first kind. A refusal or a failure stops the pass as `launch_failed`, with the error in the outcome |
| `launch_again` | The launch on the recorded worktree, `launch_tries` 2 |
| `relaunch` with both reads missing | The launch on the recorded worktree, with the relaunch entry |
| `relaunch` for a usage stop | One list read. A live row for the old id goes through the end below first. Then the pass resumes through `StartPass`, which closes its stop interval, and launches. While two passes are building, it waits (standing answer 34) |
| `end_and_relaunch` | `claude stop <old id>`, one list read, and the launch only when the old row is missing or reads `stopped`, `done`, or `failed`. A failed stop or a row still live launches nothing this run |
| `end` | `claude stop <id>`, then the record's `ended` |
| `none`, `stop`, `wait_usage`, `wait_link` | Nothing more; pass 1 already recorded them |

Before any side effect, each action checks its release's checkout and, for a relaunch, the
recorded worktree. An action on a release with no checkout does nothing, and its outcome reads
`not carried out: project <p> has no checkout recorded`. That keeps pass 1's fixtures, whose
release has none, at their recorded results. Each outcome is written to `check.json` in one last
commit, and each printed decision line carries it. A dry run carries out nothing and prints no
outcome. The check exits 0 when every action ran or was held, and 2 when a launch was refused,
naming the record in the error.

### Session verbs

- `cc-desk session start conductor <project>` and `cc-desk session start design <project>
  <round>` launch at once, under the same hidden stand-in flags as the check.
  - A live record, finished or not, is refused, and the error names it.
  - An ended record is replaced by a fresh launch that reuses its worktree and branch.
  - A release with no checkout is refused.
- `cc-desk session finish conductor <project>` and `cc-desk session finish design <project>
  <round>` set `finished`. The next check ends the session and never relaunches it.

The hidden flags are `--claude-command` (default `claude`), `--systemd-run-command` (default
`systemd-run`), and `--jobs-dir`. The check already has the first and third. A test sets the
row wait through `--launch-row-wait` (default 10 s), the time the launch polls the list for its
new row.

## Acceptance

Copied from the design's "What done looks like", items 1, 2, and 8, word for word.

1. **Launch.** Starting a free fixture pass fetches, makes one worktree on a new branch cut from the default branch's remote head, and writes a settings file holding the project's allow rules, the hook, and the two merge deny rules. The `claude` stand-in's log shows one call. That call runs inside `systemd-run --user --scope` from the worktree, carries `--permission-mode auto` and the settings file, puts the prompt before `--disallowedTools`, and holds neither `bypassPermissions` nor `--dangerously-skip-permissions`. The release file holds the session's name, ids, worktree, and branch. A second start of that pass is refused with an error naming the live record, and a launch whose worktree another record holds is refused with an error naming that record. A start for a release with no checkout recorded is refused.
2. **Recovery.** A stub conductor's row vanishes, and the next check launches it again with no stop in the snapshot. A conductor that reported finished is not launched again. A pass session's row vanishes three times. The first two checks relaunch it in the same worktree, the commit at the branch's tip is the same before and after each, the release file's count reads 1 and then 2 with a cause each time, and each relaunch prompt holds the relaunch paragraph. The third check writes "Needs you now" with the round's name and launches nothing. With a row still present, the log shows `claude stop` for the old id, a list read that shows it ended, and only then the launch. When the stop fails, no launch follows. A row missing at the first read and back at the second is left alone. After all rows vanish at once with three passes recorded as building, one check relaunches each, and none while the usage stand-in reads at the stop line.
8. **The gate.** `bash scripts/gate.sh` passes, with no skipped test. Items 6 and 7 are not Go tests. Their script's output is in the round report, and a pass that leaves it out isn't done.

| Item | Task | Check | Today's output |
| --- | --- | --- | --- |
| 1 | 6 | `go test ./e2e/ -run TestLaunch -count=1 -v` | `testing: warning: no tests to run`, `ok ... [no tests to run]` |
| 1, unit | 3 | `go test ./internal/launch/ -count=1 -v` | `stat .../internal/launch: directory not found`, `FAIL ./internal/launch [setup failed]` |
| 2 | 7 | `go test ./e2e/ -run TestRecovery -count=1 -v` | `testing: warning: no tests to run`, `ok ... [no tests to run]` |
| 8 | 8 | `bash scripts/gate.sh` | `gate: ok`, exit 0 |

Item 1's clauses map to tests this way. "Starting a free fixture pass" is `pass start` on a
release with a checkout and then one `check`: the check is the launch, since "Work that became
free" is the check's row in the design. "A second start" is `pass start` again, refused naming
the session. "A launch whose worktree another record holds" is a check whose new pass's path a
live design writer's record names. "A start for a release with no checkout" is `session start
conductor`, refused, plus a check whose `launch` outcome reads not carried out. Item 2's
"stub conductor" is a conductor record made through `session start conductor` against the
stand-in. Each `--- PASS` line goes under `## Acceptance` in the PR body with its test name.
Items 6 and 7 are pass 5's; this PR's body says so in one line.

## Tasks

Each task commits on its own, runs `bash scripts/gate.sh` bare before the commit, and is pushed.
No task holds code. Each test is table-driven where it has cases, and every new Go file follows
`standards:go-conventions`.

### Task 1: the launch rig and one probe launch (20 minutes)

- Outcome: a test makes a real repo with a bare remote, runs the stand-ins, and a launch from a
  git-made worktree through `systemd-run` reaches the `claude` stand-in. This is the round plan's
  check of its shared assumption.
- Files: `e2e/launch_rig_test.go`.
- Class: `ordinary`.
- Check: `go test ./e2e/ -run TestLaunchRig -count=1 -v` prints `--- PASS: TestLaunchRig`.
- Test: the rig does the following.
  - `git init --bare` a remote, clone it, commit once, and push `main`.
  - Write the `claude` stand-in as a heredoc script. On `--bg` it appends its argv as one JSON
    array line to `calls.log`, with its cwd. It adds a working row to `rows.json`, prints
    `backgrounded · <id> · <name>` with ids from a counter, and writes `<jobs>/<id>/state.json`
    holding `sessionId`.
  - On `agents --json --all` the stand-in logs a read and prints `rows.json`. A test edits that
    file to make a row vanish or come back, or scripts a sequence of lists.
  - On `stop <id>` it logs the call and marks the row `stopped`. It exits 1 when `stop-fails`
    exists.
  - The `systemd-run` stand-in logs its own flags and cwd to `scope.log`, then drops `--user`,
    `--scope`, and `-q` and execs the rest.
  - The probe: `git worktree add` in the test, then the binary's launch path runs from it. The
    stand-in's cwd is the worktree, and `scope.log` holds `--user --scope`.

  Since the binary has no launch yet, the probe at this task runs the `systemd-run` stand-in
  with the `claude` stand-in through `exec.Command` from the worktree. Task 6 replaces it with
  the binary.

### Task 2: the record fields and events (25 minutes)

- Outcome: the release, pass, and decision fields; the events that record a launch, a relaunch,
  an end, and a finish; and the refusal naming the live record.
- Files: `internal/state/types.go`, `check.go`, `events.go`, `checkrun.go`, and their tests.
- Class: `runner`.
- Check: `go test ./internal/state/ -count=1 -v`, read for the new cases.
- Test: unit tables cover the following.
  - `AddRelease` accepts `allow_rules` and the pass overrides, and the strict reader refuses an
    unknown key beside them.
  - A model or effort outside the known sets is refused. `model` takes `opus`, `sonnet`,
    `fable`, `haiku`, or a full `claude-*` id; `effort` takes `low`, `medium`, `high`, `xhigh`,
    or `max`.
  - `RecordLaunch` makes a record with `launch_tries` 1, a launch again gives 2, and a relaunch
    appends its entry and resets the tries and `last_seen`.
  - `EndSession` sets `ended`, and `FinishSession` sets `finished` on a conductor or a round.
  - `pass start` on a building pass with a session reads `pass "x" is building and cannot start:
    session alpha-x is live`.
  - Every existing state test passes unchanged.

### Task 3: the launch package (35 minutes)

- Outcome: `internal/launch` makes or reuses the worktree, writes the settings file, fills the
  prompt, runs the launch, reads its ids, and undoes a first launch that fails.
- Files: `internal/launch/worktree.go`, `settings.go`, `prompt.go`, `run.go`,
  `templates/{pass,design,conductor,relaunch}.txt`, and their tests.
- Class: `runner`.
- Check: `go test ./internal/launch/ -count=1 -v`, read for each case below.
- Test:
  - The argv puts the prompt before `--disallowedTools` and the five tools after it. It holds
    `--permission-mode auto` and `--settings <file>`, and neither `bypassPermissions` nor
    `--dangerously-skip-permissions`. Its first words are `--user --scope -q claude --bg`.
  - The settings file decodes to the allow rules, the two deny rules, and the hook entry.
  - Each template names its session, worktree, branch, push command, `pass-core` reference, and
    report verbs, and the relaunch paragraph holds "carry on from the branch as it stands; the
    commits on it are yours".
  - Worktree cases run against a real repo in `t.TempDir()`:
    - a fresh add;
    - an existing empty folder with no record, refused before any git command;
    - an existing branch, refused, with the branch left;
    - a failed add, which removes only the branch it made.
  - The id parse reads `backgrounded · 0c026eb1 · ccd-spike-2-idle` and
    `backgrounded · e3b28ac7`. A missing row falls back to `state.json` `sessionId`, and neither
    found fails the launch with a `claude stop`.

### Task 4: the check carries out (35 minutes)

- Outcome: the check makes the `launch` decision and carries out every action in the table under
  "The check carries out", in the order end, confirm, launch, and records each outcome.
- Files: `internal/sessions/decide.go` and its test, `cmd/cc-desk/check.go`.
- Class: `runner`.
- Check: `go test ./internal/sessions/ ./e2e/ -run 'TestDecide|TestCheck' -count=1 -v`. Every
  pass 1 check test passes with no edit, and `TestCheckDryRun` still prints 9 decisions.
- Test:
  - `decide_test.go` gains the `launch` rows: a building pass with no session, under the hold
    and not, with and without a checkout.
  - The carry-out's e2e cases come in tasks 6 and 7.
  - `check --help` no longer says "never carried out".

### Task 5: the session verbs and the hook stub (20 minutes, beside task 4 in a pair)

- Outcome: `session start` and `session finish` for a conductor and a design writer, and
  `prompt hook`.
- Files: `cmd/cc-desk/session.go`, `cmd/cc-desk/prompt.go`, `root.go`.
- Class: `runner`.
- Check: `cc-desk session --help` and `cc-desk prompt hook --help` print their usage, and `echo
  '{}' | cc-desk prompt hook` prints nothing and exits 0.
- Test: the e2e cases come in tasks 6 and 7. The pair works in its own worktree off task 3's
  commit, and its commit merges into the branch before task 6.

### Task 6: done item 1 (20 minutes, beside task 7 in a pair)

- Outcome: item 1 passes end to end.
- Files: `e2e/launch_test.go`, `e2e/launch_rig_test.go`.
- Class: `ordinary`.
- Check: `go test ./e2e/ -run TestLaunch -count=1 -v`, read for one `--- PASS` line per clause.
- Test: each clause in the mapping under `## Acceptance`, through the built binary and the rig.
  - The release comes in through `release add` with `allow_rules`, then `release checkout` names
    the test's clone.
  - The fetch is shown by a commit pushed to the remote after the clone: the new worktree's
    HEAD equals that commit.
  - The settings file is read back from `<state>/launch/`.

### Task 7: done item 2 (30 minutes, beside task 6 in a pair)

- Outcome: item 2 passes end to end.
- Files: `e2e/recovery_test.go`.
- Class: `ordinary`.
- Check: `go test ./e2e/ -run TestRecovery -count=1 -v`, read for one `--- PASS` line per
  clause.
- Test: each clause of item 2 through the built binary, with `--second-read-wait 1ms` and a
  short `--launch-row-wait`.
  - The branch tip is `git -C <worktree> rev-parse HEAD` before and after each relaunch.
  - The relaunch prompt is the argv in `calls.log`.
  - The usage stand-in is the one `e2e/standin_test.go` already writes, read at 100 percent.
  - The three-pass case builds its release through the verbs, since no old fixture changes
    (answer 16).

### Task 8: close (15 minutes)

- Outcome: the PR is ready, with its review accepted and CI green.
- Files: `docs/STATUS.md` (below the round-report block), each follow-up under
  `docs/follow-ups/`.
- Class: `ordinary`.
- Check: the PR body's `## Acceptance` evidence; the `gate` check on the last commit.
- Test: none of its own. One whole-branch `standards:diff-reviewer` read (class `runner`).

Path: tasks 1 to 4 in order, task 5 beside task 4, then tasks 6 and 7 side by side, then task 8:
20 + 25 + 35 + 35 + 30 + 15 = 160 minutes. The tasks sum to 200. At the last round's pace of about
0.6, the path is about 95 minutes. Both are under the 180-minute park line.

The CI watch, run bare after each push: `gh run list --repo glw907/cc-desk --branch
sessions-launch --limit 3`, then `gh run watch <id> --repo glw907/cc-desk --exit-status` on the
newest.

## Standing answers

The twelve in `round-run.md` hold, with the round plan's changes (4 replaced, 1 and 9 changed by
"Until cc-desk runs", 6 changed) and its answers 13 to 20. Pass 1's 23 to 29 hold where this pass
doesn't replace them: 23 and 27 are replaced by 30 below. The last round's 19 to 22 hold too: no
Go test skips; a field name is the builder's; a path past 180 stops. This pass adds:

- **30. What a check writes now.** Everything pass 1's check wrote, plus each carried-out
  action's record and outcome. A release with no checkout gets nothing carried out, so pass 1's
  fixtures and their assertions stand byte for byte. A pass 1 test that fails is a bug in this
  pass, never a reason to edit that test.
- **31. A relaunch whose recorded worktree is missing or on another branch.** The launch fails
  before any side effect. A pass stops as `launch_failed` with the path in the outcome. A
  conductor or design writer takes the stop on its session record. Nothing is recreated.
- **32. `claude --bg` fails, or its id or session id can't be read.** The launch runs `claude
  stop` on any id it printed. A first launch undoes the worktree and branch it made. The pass
  stops as `launch_failed`. Never retry inside one run.
- **33. The record commit fails after a session is live.** Run `claude stop <id>` on it, then
  undo a first launch's worktree and branch, and exit with both errors. If the undo fails, the
  next check finds an unrecorded folder and refuses it, which is the design's row.
- **34. A usage relaunch while two passes are building.** The pass stays stopped and the outcome
  reads `held: two passes building`. The next run tries again. The relaunch is still uncounted,
  as PR 8 built it.
- **35. The `writers` fixture or any fixture with building passes and no checkout.** A check
  decides `launch` for each and carries out none. Never add a checkout to an old fixture.
- **36. A real `claude` or `systemd-run` would run.** Every e2e call to `check` or `session start`
  passes both stand-in flags. A test helper that could omit them refuses to build the command
  line. `PATH=/usr/bin:/bin` holds the real `systemd-run`, so the flag is the only guard.
- **37. The deny rule form.** `Bash(gh pr merge *)` and `Bash(approval-queue land *)`, as the
  mechanism table quotes. Don't widen them and don't add more deny rules.
- **38. The tools left out.** The five proof 3 measured, for all three kinds. A kind that needs
  one of them is a follow-up, not a change in this pass.

## Decisions

- [advisor] `pass start` stays a recording verb. The check launches a building pass with no
  session, as the design's "Work that became free" row says. The no-checkout refusal lives in
  the launch, which the check and `session start` share. About 25 old e2e calls run `pass start`
  on releases with no checkout, and answer 16 keeps those fixtures, so a refusal in `start()`
  would break them. Item 1's "a start for a release with no checkout" maps to `session start`
  and to the check's not-carried-out outcome.
  Call: mine (method).
- [advisor] A conductor and a design writer each get a worktree and branch of their own, because
  `validateSession` needs both and `checkSessionsUnique` refuses two live records on one
  worktree. The design names only the pass's worktree.
  Call: mine (method, "Left to the builder").
- [advisor] The command stats the worktree path and checks the branch before `git worktree add`,
  because the probe showed git accepts an empty existing folder. The undo deletes only a branch
  this call made, and only while its tip equals `origin/<default>`. The add uses `--no-track`.
  Call: mine (method).
- [advisor] Each half-launch has a standing answer (32, 33), and every check runs before any side
  effect.
  Call: mine (method).
- Builds on PR 8's flagged call: a usage-stop relaunch counts nothing. Pass 1 decides it as an
  uncounted `relaunch`, and this pass carries that out as decided, resuming the pass through
  `StartPass`. If Geoff rules for holding until the reset time, the change sits in pass 1's
  `early` and touches none of this pass's carry-out. Answer 34 adds a hold when two passes are
  building, so the resume never puts a third pass in flight.
  Call: Geoff's (product), already filed as `docs/follow-ups/usage-relaunch-loop-on-a-model-limit.md`.
- Builds on PR 8's `[advisor]` lines: evidence is gathered outside the state lock and decided
  inside it, and the carry-out runs after the record commit under `check.lock` with each launch
  recorded by its own commit. Blocked-before-idle and the `noSettle` failed read stand. A
  non-dry-run check now writes the counts that PR 8 left for this pass.
  Call: mine (method).
- A relaunch resets `launch_tries` to 1 and clears `last_seen`, so pass 1's never-started rule
  covers each new launch. The pass's `started` is not set again by the launch, since `settle`
  already sets it.
  Call: mine (method).
- Verb names: `session start|finish conductor <project>`, `session start|finish design <project>
  <round>`, and `prompt hook`. Field names: `allow_rules`, `model`, `effort`, `outcome`. The
  settings file is `<state>/launch/<name>.json`.
  Call: mine (naming, standing answer 21).
- The hook stub prints nothing, so a prompt goes through the normal flow and pass 1's check stops
  on it. This is the hooks reference's stated behavior for a hook that makes no decision.
  Call: mine (method).
- The report verbs a session may run get exact allow rules in its settings file, beside the
  release's own rules.
  Call: mine (method, within "the rules are exact commands").
