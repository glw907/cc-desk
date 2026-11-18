# Timer: pass plan

Initiative: cc-desk
Round: sessions
Pass: sessions-timer

Class `runner` and `live-account`. Planned 75 minutes; the path below is 150, at the
150-minute park line and not over it. Design: `docs/specs/2026-11-17-sessions-design.md` (approved 2026-11-17),
section "The timer" and done items 6, 7, and 8. Round plan:
`docs/plans/2026-11-17-sessions-round.md`. Pass 4's plan,
`docs/plans/2026-11-17-sessions-merge-plan.md`, holds the ready verb and the merge this pass
runs live. The probe this plan rests on is
`docs/research/2026-11-18-auto-mode-hook-probe/README.md`. Writing a person reads (the PR body,
STATUS lines, follow-ups, the README) follows `~/.claude/docs/voice/written-for-people.md`,
loaded through the `writing-voice` skill.

Worktree `/var/home/glw907/Projects/cc-desk/.claude/worktrees/sessions-timer`, branch
`sessions-timer`, cut from `origin/sessions-merge` at `082627b` (pull request 14, open, which
carries pull requests 12, 10, and 8). PR base: `main`. Push: `glw907/cc-desk`. Before the PR is
ready, merge `origin/sessions-merge` in again if it moved, then `origin/main` if it moved.

Scope: the two unit files and an embedded copy of them; `cc-desk timer install` and `cc-desk
timer status`; the kick, a no-wait start of the service from the verbs that make work for the
check; `cc-desk desk register`, which the conductor asked for; the real hook input as a test
fixture; one script, `scripts/live-check.sh`, that runs done items 6 and 7 on the workstation;
and the follow-ups below. This pass edits none of `scripts/gate.sh`, `.github/workflows/`, or
`web/`, which are the Page round's.

## Pre-flight

No step needs Geoff before the run. Two steps the round plan put here are done:
`claude auto-mode config` shows the `cc-desk-sessions-page` entry, and `~/.claude.json` reads
`hasTrustDialogAccepted: true` for `~/Projects/cc-desk-scratch`. Proof 5 found that a new
worktree under it raises no trust prompt.

The live run uses the workstation's real state folder, `~/.local/state/cc-desk`, which did not
exist on 2026-11-18, so the scratch release is the only one the timer's check can act on. Usage
read 97 percent on 2026-11-18, against the stop line of 100, so the run is two small Sonnet
passes and no more.

- Permissions checked: `~/.local/state/claude-denials.jsonl`, read 2026-11-18, 130 rows,
  2026-11-01T00:55:05Z to 2026-11-18T06:27:23Z. No row names `systemctl`, `systemd-run`, or
  `systemd-analyze`. `claude stop` was refused when chained with `| tail`, and several times as
  Interfere With Workloads. One `claude --bg` row was a classifier outage on a chained `cd &&`.
  `gh pr create` was refused once as Create Public Surface. `claude agents --json | python3` was
  refused once as Destructive. `go test`, `gofmt`, and `gate.sh`: 0.
- Reviewed: one `standards:plan-reviewer` read of this plan (Opus, medium), 2026-11-18. Its
  findings and their folds are the last section.
- Evidence: the mechanism table below, and the three probe sessions in the research folder. Each
  acceptance check was run on `082627b` before any code; the outputs are under `## Acceptance`.

### Refused commands and their alternatives

| Refused in the log | This pass does instead |
| --- | --- |
| `claude stop` chained with a pipe or `&&` | Run `claude stop <id>` alone, as its own Bash call, with nothing after it |
| `claude agents --json` piped to a script | Write it to a file with `>`, then read the file in a second call |
| `claude --bg` after `cd &&` | Launch from the checkout with no `cd` in the same command |
| `gh pr create` as a public surface | Only the scratch sessions run it, against `glw907/cc-desk-scratch`, with the release's allow rule |

## Mechanism table

| Mechanism | Source or probe | What it means here |
| --- | --- | --- |
| Under `--permission-mode auto`, an unlisted command raises no permission request and the `PermissionRequest` hook doesn't fire | Probes `04eb1db9` and `7a406e12`, and the default-mode control `6f7ef48d` (research README) | Item 6's "a third session launched with one rule left out records its prompt" can't be met. See "Acceptance" |
| The hook input's `session_id` is the row's `sessionId` | Probe `6f7ef48d`: input, row, and `state.json` all read `6f7ef48d-7f68-41a0-a8d7-ca8e9f8258a6` | Closes the session-id half of `hook-input-from-a-real-session.md` |
| A suggestion can be `addDirectories`, with `directories` and no `rules` | `hook-input-default-mode.json` | `parseHookPrompt` (`cmd/cc-desk/prompt.go:89-99`) skips it and records no rule. Task 5 tests that |
| The launcher starts every session with `--permission-mode auto` | `internal/launch/run.go:62` | No task changes it (standing answer 3) |
| The check launches each building pass that has no session record | `internal/sessions/decide.go:158-163` | `pass start` makes the work; the kick runs the check at once |
| `pass start` refuses a pass whose `after` passes aren't done, or whose round has no design written | `state.Breaks`, `internal/state/rules.go:124-139` | The fixture round carries `design_written: true` and the passes have no `after`, so both start together |
| `release add` refuses a round with `needs_geoff: false` and no `no_need_reason`, and a project that already has a live release | `internal/state/check.go:59-60`, `internal/state/events.go:66-67` | The input carries the reason; a second `release add` of scratch can't run, so no answer restarts from `prep` |
| Two passes may build at once; a third waits on `LimitTwoPass`, and every launch waits on `LimitUsage` at the stop line | `internal/state/rules.go:205-216`, `DefaultUsageStop` 100 | At 97 percent both passes launch |
| The launcher names the binary from `os.Executable()` and the folder from `filepath.Abs` | `cmd/cc-desk/launcher.go:60` | The service's binary, `~/.local/share/cc-desk/bin/cc-desk`, is the path every scratch session's verbs name |
| A pass's name is `<project>-<pass>`, its worktree `<checkout>/.claude/worktrees/<pass>`, its branch `<pass>` | `place()` in `cmd/cc-desk/launcher.go` | The sessions are `cc-desk-scratch-timer-live-a` and `-b` |
| A first launch cuts the worktree from `origin/<default branch>` | `wt.Make()`; a relaunch reuses it with `wt.Reuse()` | The release's default branch is the run's base branch |
| The cleanup pulls the main checkout only when it is on the default branch, then runs `git branch -d` with no force | `Worktree.Pull`, `internal/launch/worktree.go:96-98`, pass 4's answers 48 and 49 | The script switches the scratch checkout to the base branch first, or `-d` refuses |
| A missing row is relaunched counted, with cause `missing`, only when the session record has a `last_seen`; with none it is an uncounted launch-again | `internal/sessions/decide.go:284-300`; `last_seen` is set when a check reads the row live, `internal/state/checkrun.go:65-66` | The stop must wait until a check has seen the second pass's row, or the release file shows no relaunch |
| A stopped pass isn't building, so the check doesn't relaunch it; `pass stop` needs no desk | Probe 2026-11-18 on `082627b`, in a temporary folder: `release add`, `release checkout`, `pass start`, then `pass stop probe pa --reason scope` printed `stopped probe/pa: scope`, and `check --dry-run` printed no decision |
| A stopped session's row reads `stopped`, which is `ClassNone`, so the check takes the missing-row branch | `internal/sessions/list_test.go:56`, `decide.go:284` | Every stop of a scratch session the run means to end runs `pass stop` first, then `claude stop` |
| Allow rules sit on the release, not the pass | `AllowRules`, `internal/state/types.go:186` | "One rule left out" is one rule left out of the release's list |
| A follow-up lands at `<worktree>/docs/follow-ups/<slug>.md`, and the verb prints its path | `cmd/cc-desk/followup.go:131-138` | Pass b commits the printed path |
| `systemctl --user show -p ...` prints the properties in the order systemd gives, and its timestamps in local time | Probe on `check-drift.timer`, 2026-11-18 | `timer status` reads `key=value` lines by key and passes `--timestamp=utc` |
| Timer properties `ActiveState`, `LastTriggerUSec`, `NextElapseUSecRealtime`; service properties `Result`, `ExecMainStatus` | The same probe | What `timer status` prints beside `last_run` |
| `last_run` in `check.json` is the time a check recorded its decisions, before it carried them out | `CheckFile.LastRun`, `internal/state/types.go:319`, set in `checkrun.go:39` | `timer status` reads it and prints the minutes since |
| A timer's `AccuracySec=` defaults to 1 minute | `systemd.timer(5)` | The timer sets `AccuracySec=1s`, so runs stay 5 minutes apart and item 7's 6-minute bound holds |
| `gh` and the push credential come from the keyring over D-Bus; the user manager's environment carries `DBUS_SESSION_BUS_ADDRESS` | `gh auth status` reads `(keyring)`; `systemctl --user show-environment` | `prep` probes `gh` from a transient user unit before `start` |
| A user unit's `PATH` here is `/usr/lib64/ccache:/usr/local/sbin:/usr/local/bin:/usr/bin` | `systemctl --user show-environment` | `claude`, `gh`, `git`, `go`, `python3`, and `jq` are all in `/home/linuxbrew/.linuxbrew/bin`, so the unit names it |
| Lingering is on | `loginctl show-user ... -p Linger`: `Linger=yes` | The user's systemd starts at boot |
| `~/.config/systemd/user` is a real folder; the Page round's `cc-desk-page.service` sits in it as a plain file | `ls -la`, 2026-11-18 | `timer install` writes plain files beside it. No stow package owns them |
| Scratch CI runs on every pull request, with no branch filter | `cc-desk-scratch/.github/workflows/gate.yml`: `on: pull_request` | A PR against the base branch gets its check |
| The scratch gate is gofmt, `go vet`, and `go test` | `cc-desk-scratch/scripts/gate.sh` | A fixture edit to a Markdown file passes it |
| The scratch checkout is clean, on `main`, at `549108a` | `git status -sb`, 2026-11-18 | The script refuses to start otherwise |
| `desk.Chain()` walks the caller's parents; `state.RegisterDesk(desk.Entry(id, pid))` records a desk | `internal/desk/desk.go:110-143`, `internal/state/ruling.go:134` | `desk register` reuses both |
| A row has `pid`, `id`, `sessionId`, `name`, `cwd`, `kind`, `startedAt`, `status`, `state`, `waitingFor`, `needs` | `internal/sessions/list.go` | `desk register` finds its row by `sessionId` and checks `kind` |
| The conductor's row: `cc-desk-ae`, `kind` `interactive`, PID 3412696 | `claude agents --json`, 2026-11-18 | The session that would run `desk register` |

## File list

From recorded searches on `082627b`, with `git grep -c` per file:

- `systemctl`: 0 hits in Go. `systemd-run`: 17, in `cmd/cc-desk/launcher.go` (3),
  `e2e/guard_test.go` (2), `e2e/launch_rig_test.go` (8), `e2e/prompt_test.go` (1),
  `internal/launch/launch_test.go` (2), and `internal/launch/run.go` (1). The kick is new code;
  none of these change.
- `timer`: 0. `desk register`: 0. Both verbs are new.
- `RegisterDesk`: 7, in `internal/state/ruling.go` (2) and `ruling_test.go` (5). The verb calls
  it and changes nothing there.
- `go:embed`: 1, `internal/launch/prompt.go`. The units' package follows its shape.
- `addDirectories`: 0. Task 5 adds the first.

Files this pass adds or changes:

| File | Task |
| --- | --- |
| `systemd/cc-desk-check.service`, `systemd/cc-desk-check.timer`, `systemd/units.go`, `systemd/units_test.go` (all new) | 1 |
| `cmd/cc-desk/timer.go` (new), `e2e/timer_test.go` (new), `cmd/cc-desk/root.go` (the `AddCommand` line, `:120`, and one hidden flag) | 2 |
| `cmd/cc-desk/kick.go` (new), `cmd/cc-desk/pass.go`, `cmd/cc-desk/ruling.go`, `e2e/kick_test.go` (new) | 3 |
| `cmd/cc-desk/desk.go` (new), `e2e/register_test.go` (new), `cmd/cc-desk/root.go` (`:120`) | 4 |
| `e2e/testdata/prompts/hook-add-directories.json` (new), `cmd/cc-desk/prompt_test.go`, `e2e/prompt_test.go` | 5 |
| `README.md`, lines 6 to 10 only | 6 |
| `scripts/live-check.sh`, `scripts/live/release.json`, `scripts/live/timer-live-a-plan.md`, `scripts/live/timer-live-b-plan.md` (all new) | 7 |
| `docs/follow-ups/*.md` | 9 |

The fixture plans live under `scripts/live/`, not `docs/`, so the gate's Vale run never reads
them.

## What the pass builds

### The units

`cc-desk-check.service` is a one-shot:

- `Type=oneshot`
- `ExecStart=%h/.local/share/cc-desk/bin/cc-desk --state-dir %h/.local/state/cc-desk check`
- `Environment=PATH=%h/.local/share/cc-desk/bin:%h/.local/bin:/home/linuxbrew/.linuxbrew/bin:/usr/local/bin:/usr/bin:/bin`
- `KillMode=process`

`cc-desk-check.timer` has `OnStartupSec=1min`, `OnUnitActiveSec=5min`, and `AccuracySec=1s`, and installs into
`timers.target`. The `systemd` package embeds both files, so `timer install` writes the same
bytes the repo holds.

### `timer install` and `timer status`

`timer install` copies `os.Executable()` to `$HOME/.local/share/cc-desk/bin/cc-desk`, through a
temporary file and a rename, so a running check is never left with half a binary. It writes
both units to `$HOME/.config/systemd/user/`, then runs `systemctl --user daemon-reload` and
`systemctl --user enable --now cc-desk-check.timer`. The paths come from `$HOME`, never the XDG
variables, because the units name `%h`.

`timer status` prints `last_run` from `check.json` with the minutes since it, or `no run
recorded` when the file is absent. It then prints the timer's `ActiveState`, `LastTriggerUSec`,
and `NextElapseUSecRealtime`, and the service's `Result` and `ExecMainStatus`, read from
`systemctl --user show --timestamp=utc`. A `systemctl` failure prints its message and exits 1.
The hidden persistent flag `--systemctl-command` names the command; every test sets it to a
stand-in.

### The kick

After a successful commit, `pass start`, `pass ready`, and `ruling record` with `--question` run
`systemctl --user start --no-block cc-desk-check.service`. The kick fires only when the state
folder, made absolute and with symlinks resolved, is `$HOME/.local/state/cc-desk` resolved the
same way, or when `--systemctl-command` is set. Every e2e test uses a temporary state folder, so
no old test kicks and none needs an edit. A failed kick prints one line to stderr, `kick: ...`,
and the verb still exits 0.

### `desk register`

`cc-desk desk register` reads `claude agents --json` (the hidden `--claude-command` flag names a
stand-in) and finds the row whose `sessionId` is `$CLAUDE_CODE_SESSION_ID`. It refuses, naming
the failed test, when the variable is empty, no row matches, the row's `kind` isn't
`interactive`, the row's `pid` isn't in `desk.Chain()`, or that process has no terminal. Then it
commits `state.RegisterDesk(desk.Entry(id, pid))`. This is off the live run's path: the
executor never runs it against the live state, and the conductor may run it after the merge.

### The live script

`scripts/live-check.sh <phase>` runs one phase per call and appends what it prints to
`$CLAUDE_JOB_DIR/tmp/live/run.log`. Each phase prints the commands it ran and their output. It
reads `CC_DESK_STATE_DIR` (default `~/.local/state/cc-desk`) and `SCRATCH_CHECKOUT` (default
`~/Projects/cc-desk-scratch`), so task 7 can try its refusals on temporary copies. The executor
runs `watch` and `wait` with the Bash tool's `timeout` set to 600000, since the default is 120000
milliseconds; every other phase finishes in under 2 minutes.

1. `prep`:
   1. Refuse unless the scratch checkout is clean and on `main`, and the state folder holds no
      release but scratch's (answer 64).
   2. Build the binary and run `timer install`, then `systemd-analyze --user verify` on both
      installed units.
   3. Run `systemd-run --user --wait --pipe -p Environment=PATH=<the unit's PATH> gh pr list
      --repo glw907/cc-desk-scratch --limit 1`, which shows `gh` reaches its keyring from a user
      unit.
   4. In scratch: `git switch -c live-2026-11-18 origin/main`, copy the two fixture plans to
      `docs/plans/<pass>-plan.md`, `git add` them by path, `git commit`, and `git push origin
      live-2026-11-18`.
   5. Add `.claude/worktrees/` to scratch's `.git/info/exclude`.
   6. Run `release add scripts/live/release.json`, then `release checkout cc-desk-scratch
      <scratch checkout> --default-branch live-2026-11-18`.
   7. Print the usage reading.
2. `start` writes the start time and runs `pass start cc-desk-scratch timer-live-a` and `pass
   start cc-desk-scratch timer-live-b` with the installed binary. Each kicks the service. A kick
   that lands while the first check runs may merge into it, so `start` then waits until the service
   reads `inactive` and runs `systemctl --user start --no-block cc-desk-check.service` once more.
3. `watch` waits up to 9 minutes a call for two things: the second pass's first commit on
   `origin/timer-live-b`, and a `last_seen` on its session record in the release file. Then it
   saves that record's `session_id` to the run folder and prints the record's short id and the
   `claude stop` line. The executor calls it a second time if the first call ends without both.
   Then the executor runs `claude stop <id>` alone.
4. `wait` polls every 30 seconds, up to 9 minutes a call, until both passes read done with their
   cleanup recorded, and prints the state of each. The executor calls it again until done or 45
   minutes have passed since `start` (answer 67).
5. `scope` runs item 7's scope probe on its own, as a transient unit:
   `systemd-run --user --wait --unit=cc-desk-scope-probe -p Type=oneshot -p KillMode=process -p
   Environment=PATH=<the unit's PATH> systemd-run --user --scope -q
   --unit=cc-desk-scope-probe-sleep <stand-in>`. The stand-in, written to the run folder, starts
   `sleep 901` in the background, prints `backgrounded · probe`, and exits. After `--wait`
   returns, the phase prints `systemctl --user is-active cc-desk-scope-probe.service`, which
   reads `inactive`, and `systemctl --user is-active cc-desk-scope-probe-sleep.scope`, which
   should read `active`. Then it runs `systemctl --user stop cc-desk-scope-probe-sleep.scope`.
6. `abort` is for a run that ends before `evidence` (answer 67). It runs `systemctl --user
   disable --now cc-desk-check.timer` and prints, for each `cc-desk-scratch` pass still
   building, the `pass stop` line and the `claude stop` line to run.
7. `evidence` prints `timer status` first, then each done item's evidence as "Acceptance" lists
   it. Then it runs `git -C <scratch checkout> switch main` and `systemctl --user disable --now
   cc-desk-check.timer`.

Each session's transcript, `~/.claude/projects/*/<session id>.jsonl`, outlives its worktree.
`evidence` reads every transcript of each pass, the second pass's first session included, and
counts only Bash tool calls whose `command` holds `gh pr merge`, read with `jq`. A plain grep
would count the string anywhere in a tool's output. "Gone" is one command per item: no row named
`cc-desk-scratch-timer-live-a` in a saved `claude agents --json`, `test ! -e` on the worktree
path, and `git -C ~/Projects/cc-desk-scratch show-ref --verify refs/heads/timer-live-a` failing.

### The fixture passes

Both run on Sonnet at medium effort. Each fixture plan says its steps are the whole pass: no
planning, no review, no STATUS, no other file. The launcher's prompt says "Plan the pass first",
so each plan opens by saying it is the pass's plan and no other is written. Neither plan holds
the string `gh pr merge`. Pass a, `timer-live-a`, runs these commands from
its worktree:

1. Write `notes-timer-live-a.md` with the one line `timer-live-a ran.`
2. `bash scripts/gate.sh`
3. `git add notes-timer-live-a.md`
4. `git commit -m "Add timer-live-a notes"`
5. `git push origin timer-live-a`
6. `gh pr create --repo glw907/cc-desk-scratch --base live-2026-11-18 --head timer-live-a --title "Add timer-live-a notes" --body "Fixture pass timer-live-a."`
7. `git rev-parse HEAD`
8. The ready command its prompt names, with `--pr <number> --commit <sha> --verdict accept --design inside`

Pass b, `timer-live-b`, first runs `git log --oneline origin/live-2026-11-18..HEAD` and skips
any step whose commit is already there. When step 1's commit is there, the session is a
relaunch, and it skips both sleeps too. Then:

1. Write `notes-timer-live-b.md` with the line `step 1`, then run `bash scripts/gate.sh`, `git
   add notes-timer-live-b.md`, `git commit -m "Add timer-live-b step 1"`, and `git push origin
   timer-live-b`.
2. `sleep 240`
3. Add the line `step 2`, then run the gate, the add, `git commit -m "Add timer-live-b step 2"`,
   and the push.
4. `sleep 240`
5. Write the body `A fixture follow-up from timer-live-b.` to `$CLAUDE_JOB_DIR/tmp/followup.md`,
   run the follow-up command its prompt names with the title `Fixture follow-up from
   timer-live-b` and `--body-file` that path. It prints an absolute path, so the session stages
   it relative to the worktree, `git add docs/follow-ups/<slug>.md`, then runs `git commit -m
   "File the timer-live-b follow-up"`, and the push.
6. The `gh pr create`, `git rev-parse HEAD`, and ready steps of pass a, with the second pass's names.

The release input's allow rules, shared by both passes:

- `Bash(bash scripts/gate.sh)`
- `Bash(git add notes-timer-live-a.md)`, `Bash(git add notes-timer-live-b.md)`, and
  `Bash(git add docs/follow-ups/*)`
- `Bash(git commit -m *)`
- `Bash(git push origin timer-live-a)` and `Bash(git push origin timer-live-b)`
- `Bash(gh pr create --repo glw907/cc-desk-scratch *)`
- `Bash(git rev-parse HEAD)` and `Bash(git log --oneline *)`

`Bash(sleep 240)` is the one rule left out. Under auto mode it should raise no prompt, and the
release file should record none. The two sleeps keep the first session running for 8 minutes
after its first push, so the stop lands mid-run after a check has read its row. The relaunched
session skips them, so the run stays short, and the first session's transcript holds the
left-out rule's evidence.

## Acceptance

Done items 6, 7, and 8 from the design, word for word:

> 6. **A real pass, no prompt.** On the workstation, with the timer installed, a desk session
> adds a fixture pass to a release for the scratch repo, and the timer's check launches it.
> The session edits a file, runs the gate, opens a pull request, and reports ready. The
> check merges it after its check passes: the release file's merge record names the check,
> and the session's log holds no merge command. Within 10 minutes of the merge the session,
> the worktree, and the branch are gone. The hook is shown to fire first: a third session
> launched with one rule left out records its prompt. Then the release file records zero
> prompts for the first pass, and the denials log holds no row for it. Nobody typed anything
> after the start. A second such pass has its session ended with `claude stop` mid-run. It
> is running again within 10 minutes, and the release file shows one relaunch with its
> cause. The round report gives each run's times and the prompt count. Any prompt that did
> appear is in the state file with its cause, counted in the report, and filed as a
> follow-up.
>
> 7. **The timer.** On the workstation, `systemd-analyze verify` passes on both unit files, and
> `cc-desk timer status` names a run in the last 6 minutes. A stand-in launch that starts a
> long sleep, run by the service through the scope, is still running after the service
> exits. A Go test reads the unit files in the repo and finds the `PATH` line and
> `KillMode=process`.
>
> 8. **The gate.** `bash scripts/gate.sh` passes, with no skipped test. Items 6 and 7 are not Go
> tests. Their script's output is in the round report, and a pass that leaves it out isn't
> done.

Item 6's clause "The hook is shown to fire first: a third session launched with one rule left
out records its prompt" is unmet, and this pass doesn't try to meet it. The research folder
shows why: under the launcher's `--permission-mode auto`, an unlisted command runs with no
prompt and the hook never fires. With the hook unreachable, "zero prompts" no longer says
anything about the allow list. It says only that no session stopped on a prompt. The run still
leaves `Bash(sleep 240)` out and records that the release file holds no prompt for it. The
choice of what replaces the clause is Geoff's, under `## Calls for Geoff`, and the follow-up
`auto-mode-prompt-clause.md` carries it.

Each check, run on `082627b` before any code:

| Check | Output before code |
| --- | --- |
| `cc-desk timer status` | `unknown command "timer" for "cc-desk"`, exit 1 |
| `systemd-analyze --user verify systemd/cc-desk-check.service systemd/cc-desk-check.timer` | `Unit cc-desk-check.service not found.` and `Unit cc-desk-check.timer not found.`, exit 1 |
| `systemctl --user list-timers --all` lines naming cc-desk | 0 |
| `cc-desk desk register` | `unknown command "desk" for "cc-desk"`, exit 1 |
| `bash scripts/gate.sh` | `gate: ok`, exit 0 |

What `scripts/live-check.sh evidence` prints for each part, and what counts as met:

- Item 6, pass a: the release file's merge record for `timer-live-a`, naming the check; the
  `jq` count of Bash calls holding `gh pr merge` in its transcripts, 0; the merge time and the three "gone"
  commands with their times, each within 10 minutes of the merge; the release file's prompt
  count for the pass, 0; and the count of denials rows whose `cwd` is under
  `cc-desk-scratch/.claude/worktrees/` with a time after `start`, 0.
- Item 6, pass b: the `claude stop` time; the time the row read `working` again, within 10
  minutes; the release file's relaunch count, 1, with cause `missing`; the follow-up file on
  `origin/live-2026-11-18`; and the same prompt and denials counts.
- Item 6, the left-out rule: the release file's prompts, 0, beside the time pass b ran `sleep
  240`, read from its transcript.
- Item 6, the times: each pass's start, launch, ready, merge, and cleanup, for the round report.
- Item 7: `systemd-analyze --user verify` on both installed units, exit 0; `timer status` with
  `last_run` under 6 minutes old; the scope probe's line showing the sleep alive after the
  service's `inactive`; and `go test ./systemd/ -run TestUnits -v` printing `--- PASS:
  TestUnits`.
- Item 8: `bash scripts/gate.sh` printing `gate: ok`, with a `SKIP` count of 0.

The round report, `~/Projects/initiatives/cc-desk/reports/sessions-report.md`, is the
conductor's (round plan answer 20). This pass puts the `evidence` output in the PR body under
`## Acceptance`, ready for the conductor to copy.

## Tasks

Path: task 1, task 2, task 3, task 5, task 6, then tasks 4 and 7 side by side in a pair, then
task 8, task 9, and task 10. The pair counts once, at its longer task's 25 minutes. That is
10 + 20 + 15 + 5 + 5 + 25 + 45 + 10 + 15 = 150 minutes, at the park line and not over it.

The live run, from `start`: the second pass launches at about 1 minute and pushes at about 3. A
check reads its row by about 6, the stop lands by about 7, and the next check relaunches it by
about 12. The relaunch skips the sleeps and reports ready by about 16. Pass a is ready by about
7. Each merge waits for scratch CI and the next check, so both are merged and cleaned up by
about 27. `prep`, `scope`, and `evidence` add about 8 minutes, and answer 67's 45-minute cap
leaves room for one slow check.

### Task 1: the units (10 minutes)

- Outcome: the two unit files exist in the repo, embedded in package `systemd`, and a Go test
  reads them.
- Files: `systemd/cc-desk-check.service`, `systemd/cc-desk-check.timer`, `systemd/units.go`,
  `systemd/units_test.go`, all new.
- Class: `runner`.
- Check: `go test ./systemd/ -run TestUnits -count=1 -v` prints `--- PASS: TestUnits`.
- Test: `TestUnits` reads both files from disk, not through the embed, and finds the exact
  `Environment=PATH=` line, `KillMode=process`, `Type=oneshot`, the `ExecStart` line,
  `OnStartupSec=1min`, `OnUnitActiveSec=5min`, and `AccuracySec=1s`. It also checks that the embedded bytes equal
  the files.

### Task 2: `timer install` and `timer status` (20 minutes)

- Outcome: both verbs work against a stand-in `systemctl` and a temporary home.
- Files: `cmd/cc-desk/timer.go` and `e2e/timer_test.go`, new; `cmd/cc-desk/root.go`, the
  `AddCommand` line and the hidden persistent `--systemctl-command` flag.
- Class: `runner`.
- Check: `go test ./e2e/ -run 'TestTimerInstall|TestTimerStatus' -count=1 -v` prints `--- PASS:`
  for both.
- Test: each case sets `HOME` to a temporary folder and `--systemctl-command` to a stand-in
  script that logs its arguments and prints canned `show` output. Install: the binary copy
  exists and runs `--version`; both units hold the repo's bytes; the stand-in log holds
  `--user daemon-reload` and then `--user enable --now cc-desk-check.timer`. A second install
  replaces both files and still passes. Status: with a `check.json` whose `last_run` is 3
  minutes before `--now`, it prints that time and `3 minutes ago` and the five properties;
  with no `check.json`, it prints `no run recorded`; a stand-in that exits 1 makes it exit 1
  with the stand-in's message.

### Task 3: the kick (15 minutes)

- Outcome: the three verbs start the service without waiting, only on the live state folder or
  the stand-in.
- Files: `cmd/cc-desk/kick.go` and `e2e/kick_test.go`, new; `cmd/cc-desk/pass.go`
  (`start`, `ready`), and `cmd/cc-desk/ruling.go` (`record`).
- Class: `runner`.
- Check: `go test ./e2e/ -run TestKick -count=1 -v` prints `--- PASS: TestKick` and one `---
  PASS:` per case.
- Test: with `--systemctl-command` set to a logging stand-in, `pass start` and `pass ready`
  each log `--user start --no-block cc-desk-check.service` once, after the state file holds
  the event. A ruling with `--question`, run under `startDesk` in `e2e/desk_test.go`, logs
  it; a ruling without one doesn't. A refused `pass start` logs nothing. A stand-in that exits
  1 leaves the verb at exit 0 with one `kick:` line on stderr. With no flag, a temporary state
  folder, and a logging `systemctl` first on `PATH`, nothing is logged. With no flag, `HOME` set to
  a temporary folder, `--state-dir $HOME/.local/state/cc-desk`, and the same logging `systemctl`
  first on `PATH`, `pass start` logs the kick once. Every old e2e test
  passes with no edit.

### Task 4: `desk register` (15 minutes, a pair with task 7)

- Outcome: a terminal session can register itself as a desk, and nothing else can.
- Files: `cmd/cc-desk/desk.go` and `e2e/register_test.go`, new; `cmd/cc-desk/root.go`, the
  `AddCommand` line.
- Class: `runner`.
- Check: `go test ./e2e/ -run TestDeskRegister -count=1 -v` prints `--- PASS: TestDeskRegister`
  and one `--- PASS:` per case.
- Test: through `startDesk(t, terminal, env, script)`. The desk shell blocks on its FIFO until
  `release()` (`e2e/desk_test.go:17-19`), so the test writes the `claude` stand-in after
  `startDesk` returns the desk's PID and before it calls `release()`. The stand-in prints a row
  whose `pid` is that PID: an interactive row under a terminal registers, and
  `desks.json` holds the session id and the start time; a `bg` row, a row whose `pid` is not in
  the chain (never PID 1, which `desk.Chain()` holds), a desk with no terminal (`setsid`), an empty `CLAUDE_CODE_SESSION_ID`, and no
  matching row are each refused with the failed test named. After a registration, `ruling
  record` from the same desk passes the guard.

### Task 5: the real hook input (5 minutes)

- Outcome: the hook input from probe `6f7ef48d` is a fixture and a test case.
- Files: `e2e/testdata/prompts/hook-add-directories.json` (new; every path under
  `/var/home/...` replaced with one under `/tmp/cc-desk-fixture/`), `cmd/cc-desk/prompt_test.go`,
  `e2e/prompt_test.go`.
- Class: `runner`.
- Check: `go test ./cmd/cc-desk/ -run TestParseHookPrompt -count=1 -v` and `go test ./e2e/ -run
  TestPromptHook -count=1 -v` print `--- PASS:` for the new cases.
- Test: `parseHookPrompt` on the fixture gives the session id, `Bash`, the `date` command, and
  an empty rule. A new `TestPromptHook` case feeds the fixture, with its `session_id` set to the
  rig's session, and the prompt record holds an empty `suggested_rule`.

### Task 6: the README (5 minutes)

- Outcome: the paragraph at lines 6 to 10 names the check, the timer, and the two new verbs.
- Files: `README.md`, lines 6 to 10 only. The Page round appends after line 13.
- Class: `ordinary`.
- Check: `git diff origin/sessions-merge -- README.md` touches no line past 10; Vale passes in
  the gate.
- Test: none.

### Task 7: the live script (25 minutes, a pair with task 4)

- Outcome: `scripts/live-check.sh` and its three inputs exist, and `bash -n` passes.
- Files: `scripts/live-check.sh`, `scripts/live/release.json`,
  `scripts/live/timer-live-a-plan.md`, `scripts/live/timer-live-b-plan.md`, all new.
- Class: `runner`, `live-account`.
- Check: `bash -n scripts/live-check.sh` exits 0, and `bash scripts/live-check.sh help` prints the
  seven phases. In a temporary state folder, with `--usage-command e2e/testdata/usage40.sh`:
  `release add` of `release.json`, `release checkout cc-desk-scratch <a temporary folder>
  --default-branch live-2026-11-18`, and both `pass start` calls each exit 0, and each start
  prints `started`.
- Test: the release input carries `format`, `project` `cc-desk-scratch`, `name`, `repo`
  `glw907/cc-desk-scratch`, one round (`name`, `summary`, `needs_geoff: false`,
  `no_need_reason`, `design_written: true`), the two passes (`model` sonnet, `effort` medium, `state`
  `not_started`, `planned_minutes`, `summary`, `round`), and the allow rules above. `prep`'s
  refusals are tried with `CC_DESK_STATE_DIR` and `SCRATCH_CHECKOUT` pointed at temporary
  copies: a dirty checkout, and a state folder holding another release. No try touches the real
  state folder or the scratch checkout.

### Task 8: the live run (45 minutes)

- Outcome: items 6 and 7 have their evidence, or a recorded unmet line with the reason.
- Files: none in the repo; the run log is under `$CLAUDE_JOB_DIR/tmp/live/`.
- Class: `live-account`.
- Check: the `evidence` phase's output.
- Test: run `prep`, `start`, `watch` (once or twice), `claude stop <id>` alone, `wait` until
  done, `scope`, and `evidence`, in that order, each as its own Bash call, with `timeout` 600000
  on `watch` and `wait`. Read the usage reading `prep` prints before `start`. A run that ends
  early goes through answer 67.

### Task 9: the follow-ups (10 minutes)

- Outcome: one new follow-up is filed, the planner's is updated, and four are closed.
- Files: `docs/follow-ups/auto-mode-prompt-clause.md`, written by the planner, gets the run's
  result; `docs/follow-ups/cold-boot.md` is new;
  `hook-input-from-a-real-session.md`, `desk-registration-gap.md`, `ready-verb-service-start.md`,
  and `followup-rule-under-auto-mode.md`, each closed with one line naming this pass and its
  evidence, or removed if that is how the repo closes one.
- Class: `ordinary`.
- Check: Vale passes in the gate.
- Test: none. `auto-mode-prompt-clause.md` already gives the probe, the three choices, and the
  recommendation; task 9 adds whether the left-out rule raised a prompt and any denial row. `cold-boot.md` covers the boot where the
  timer's check is the first to start Claude Code's supervisor, which the round can't test
  without ending every session, for the Move round's first restart. It also records that the
  check's `gh` and pushes need the keyring over D-Bus, which a cold boot may not have unlocked.
  `ready-verb-service-start.md` misread the design as the page service; its closing line says
  the design means `cc-desk-check.service`.

### Task 10: close (15 minutes)

- Outcome: the PR is ready, with its evidence, review, and green CI.
- Files: `docs/STATUS.md` only if a stop arises.
- Class: `runner`.
- Check: `gh pr view <n> --repo glw907/cc-desk --json isDraft` reads `false`; CI is green on the
  last commit.
- Test: one whole-branch `standards:diff-reviewer` read, folded; merge `origin/sessions-merge`
  and then `origin/main` if they moved, with the gate after each.

## Standing answers

The twelve in `round-run.md` hold, with the round plan's changes and its answers 13 to 20.
Pass 1's 23 to 29, pass 2's 30 to 38, pass 3's 39 to 46, and pass 4's 47 to 53 hold, with 150
minutes as this pass's park line. The last round's 19 to 22 hold too. This pass adds:

- **54. A test would kick the real service.** Every test uses a temporary state folder, and
  every kick test sets `--systemctl-command`. A test that sees a real kick means the condition
  is wrong; fix the condition, never the test's folder.
- **55. A test would write to the real home.** Every `timer` test sets `HOME` to a temporary
  folder and the stand-in `systemctl`. No test runs the real `systemctl` or `systemd-analyze`.
- **56. `systemd-analyze --user verify` fails on an installed unit.** Fix the unit file in the
  repo and reinstall. Never edit another unit, and never write a drop-in.
- **57. A fixture session strays from its plan.** It may read `pass-core` and write files the
  plan doesn't name. If a pass hasn't reported ready 30 minutes after `start`, run
  `~/.local/share/cc-desk/bin/cc-desk --state-dir ~/.local/state/cc-desk pass stop
  cc-desk-scratch <pass> --reason scope` alone, then `claude stop <id>` alone, and record the
  part as unmet with its transcript lines. There is no rerun: no verb adds a pass to a live
  release.
- **58. The classifier denies a step in a fixture session.** That is a finding: record the
  denial row, file it as a follow-up, and count it in `## Acceptance`. If the session can't go
  on, stop it as answer 57 says. Never edit a Claude settings file, an `autoMode` entry, or
  a hook.
- **59. Usage reads 100 percent or more.** Before `start`, run nothing live: write the stop
  under `## Stop` in `docs/STATUS.md`, with tasks 8 and 9 named as left, and end. During the
  run, let the check wait out the hold until answer 67's 45 minutes run out.
- **60. The check launches nothing within 5 minutes of `start`.** Read `timer status` and
  `journalctl --user -u cc-desk-check.service -n 50 --no-pager`. A cause in this pass's code is
  fixed once: commit it, run `timer install` again from the fixed build, and run `systemctl
  --user start --no-block cc-desk-check.service`. Never rerun `prep`. Any other cause goes to
  answer 67.
- **61. A fixture PR doesn't merge.** Read `check.json`'s decisions and the PR's checks with
  `gh pr view <n> --repo glw907/cc-desk-scratch --json statusCheckRollup`. A red scratch gate is
  recorded as the reason item 6's merge is unmet.
- **62. The pull or `git branch -d` refuses in the live run.** Pass 4's answers 48 and 49 hold:
  find the cause with `git -C ~/Projects/cc-desk-scratch status --porcelain --ignored` and the
  branch's state, and never force.
- **63. `claude stop` is refused.** Run it once more, alone, with nothing else on the line. If it
  is refused again, record the second pass's relaunch clause as unmet with the denial row and go on.
- **64. The state folder holds a release that isn't scratch's, or the scratch checkout is dirty
  or off `main`, at `prep`.** Stop before the install. Never move another release.
- **65. A scratch session is still live at the end.** Run `pass stop ... --reason scope` alone,
  as answer 57 writes it, then `claude stop <id>` alone, and record it. Never stop a session that
  isn't `cc-desk-scratch-*`.
- **66. `desk register` against the live state.** The executor never runs it against
  `~/.local/state/cc-desk`. The conductor may, after the merge.
- **67. The run ends before `evidence`.** That covers a stop, answer 60's other causes, a usage
  hold, and `wait` passing 45 minutes. Run `scripts/live-check.sh abort`, then each `pass stop`
  and `claude stop` line it prints, each alone, then `scripts/live-check.sh evidence` for the
  parts that exist. Record every part with no evidence as unmet with its reason. Never leave the
  timer enabled.

## Decisions

- [advisor] Item 6's hook clause is marked unmet with the probe evidence, not redefined. No
  task changes `--permission-mode` or adds `ask` rules, and the run is two passes, not three.
  The follow-up recommends the `PermissionDenied` hook.
  Call: Geoff's (the design's done item), listed under `## Calls for Geoff`.
- [advisor] With the hook unreachable, the zero-prompt count says only that no session stopped
  on a prompt. The PR body and the conductor's message both say so.
  Call: mine (method).
- [advisor] The kick fires only on the live state folder or a set stand-in, so no temporary
  folder in any test can start the real service. `pass start` kicks as well as `pass ready` and
  a ruling that answers a question, since a started pass is work for the check.
  Call: mine (method).
- [advisor] The binary lives at `~/.local/share/cc-desk/bin/cc-desk`, copied by `timer install`,
  so the units and every scratch session's verbs name a path no rebuild in a worktree moves.
  Call: mine (method).
- [advisor] Install paths come from `$HOME`, never the XDG variables, because the units name
  `%h`.
  Call: mine (method).
- [advisor] The live run cuts its own base branch, `live-2026-11-18`, and points the release's
  default branch at it, so the executor pushes nothing to scratch's `main`. The script switches
  the scratch checkout to that branch for the run, so the pull after the merge and `git branch -d` work,
  and switches it back at the end.
  Call: mine (method).
- [advisor] The design's "a desk session adds a fixture pass" is read as the session that runs
  `release add`. That verb needs no desk, and `desk register` stays off the run's path.
  Call: mine (method).
- [advisor] This builds on pull request 12's flagged gap, "no verb registers a desk": `desk
  register` is that verb, refused for anything but an interactive session under a terminal.
  Call: mine (method), names a flagged call.
- [advisor] This builds on pull request 8's flagged usage relaunch call only at its edge: the
  live run launches at 97 percent under the stop line of 100, so the usage hold isn't
  exercised, and no relaunch here is a usage one.
  Call: mine (method), names a flagged call.
- [advisor] The left-out rule is `Bash(sleep 240)`, which only the second pass runs. Leaving out a push or
  `gh pr create` rule would risk a classifier denial that stops the run's other checks.
  Call: mine (method).
- [advisor] The two fixture passes run at once, since two may build together, which keeps the run
  inside the usage left.
  Call: mine (method).
- [advisor] Item 7's scope probe runs as a transient unit with the service's `Type`,
  `KillMode`, and `PATH`, starting the launcher's `systemd-run --user --scope` shape with a
  stand-in that backgrounds a long sleep in a named scope. Running it through the real unit
  would need a drop-in. The probe stops its scope with `systemctl --user stop` when it is done.
  Call: mine (method).
- [advisor] "The branch is gone" means the scratch checkout's local branch. The merge leaves the
  remote branch to GitHub's own setting.
  Call: mine (method).
- [advisor] The pass leaves the timer installed but disabled at the end, since the binary it
  runs is this branch's build. After the merge, the conductor runs `cc-desk timer install` from
  `main`, and the Move round records the install in the workstation repo, as the design says.
  Call: mine (method).
- The run goes ahead at 97 percent, as the conductor asked: keep the live run small at that
  reading. Answer 59 stops it before `start` at 100, and answer 67 bounds an early end. A stop
  before `start` is likely unless usage resets before task 8; the conductor's launch message
  says so.
  Call: the conductor's (budget), answering the reviewer's owner fork.
- [advisor] The scratch release stays in `~/.local/state/cc-desk` after the run, so the Page
  round's page shows it. The installed unit files and binary are recorded in no
  manifest until the Move round records the install.
  Call: mine (method).
- [advisor] The path is 150 minutes against 75 planned, at the park line and not over it. The
  conductor's added scope (`desk register`, the hook input, the kick) is kept whole.
  Call: mine (method).

## Calls for Geoff

1. **Item 6's hook clause can't be met under auto mode.** Every launched session runs with
   `--permission-mode auto`, and in that mode an unlisted command raises no prompt, so the
   prompt hook never fires. Recommended: accept item 6 without that clause for this round, and
   take the `PermissionDenied` hook, which fires on an auto-mode denial, as a later pass that
   records each denial the way a prompt is recorded now. The other choices are dropping the
   clause for good, or launching sessions in another mode, which brings back the prompts auto
   mode removed.

## Review findings and their folds

One `standards:plan-reviewer` read (Opus, medium), 2026-11-18. It returned 15 findings. Each is
folded; none changed the design.

| Finding | Fold |
| --- | --- |
| 1, blocker: the release input can't start a pass, because the round has no design written and no `no_need_reason` | The input's round carries both; task 7's check runs `release add`, `release checkout`, and both `pass start` calls in a temporary folder; mechanism rows added |
| 2, blocker: a stop before a check has read the row gives an uncounted launch-again | `watch` waits for `last_seen` as well as the first push; the sleeps are 240 seconds and a relaunch skips them; mechanism row fixed |
| 3, major: no verb adds a rerun pass, and a bare `claude stop` gets relaunched | Reruns removed; every intended stop runs `pass stop` first (answers 57 and 65) |
| 4, major: a second `prep` can't run | Answer 60 fixes, reinstalls, and kicks, with no new `prep` |
| 5, major: no early end disables the timer | The `abort` phase and answer 67 |
| 6, owner fork: usage at 97 percent | The run goes ahead under answers 59 and 67, as the conductor asked, named under `## Decisions` |
| 7, major: the Bash timeout | `timeout` 600000 on `watch` and `wait` |
| 8, major: the transcript check misses the relaunched session and counts output text | Every transcript of each pass, counted by `jq` over Bash `command` fields; the fixture plans don't hold the string |
| 9, major: the follow-up path is absolute, so its `git add` misses the rule | The plan stages it relative to the worktree |
| 10, minor: refusal tests would touch real state | `CC_DESK_STATE_DIR` and `SCRATCH_CHECKOUT` |
| 11, minor: `prep` commits before it switches | The steps written in order |
| 12, minor: timer accuracy and `last_run`'s meaning | `AccuracySec=1s` in the unit and `TestUnits`; `evidence` prints `timer status` first; row fixed |
| 13, minor: a kick may merge into a running start | `start` kicks once more after the service reads `inactive`; `watch` may run twice |
| 14, minor: `gh` from a user unit is unprobed, and `scope` isn't written out | `prep` probes `gh` through `systemd-run`; the `scope` phase's commands are written out; `cold-boot.md` records the keyring |
| 15, minor: a wrong cite, the desk PID in task 4, PID 1, no positive kick case, the release on the page | Each folded where it sits; the last is a decision |
