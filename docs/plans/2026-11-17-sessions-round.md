# Sessions: the round plan

This file is what `pass-core` asks a round of several passes to carry on top of its design. The
design is [2026-11-17-sessions-design.md](../specs/2026-11-17-sessions-design.md). Where the two
differ, the design wins. A planner writes each pass's plan just before the pass runs.

- Initiative: `cc-desk`
- Round id: `sessions`
- Repo: `/var/home/glw907/Projects/cc-desk`, pushing to `glw907/cc-desk`, default branch `main`
- Conductor: `conductor-cc-desk`, a session Geoff starts in his terminal on Opus. It conducts
  this round and the Page round together
- Conductor prompt: `docs/plans/2026-11-17-sessions-page-conductor-prompt.txt`
- Chain: 375 planned minutes. The round starts nothing new past 750.

## Pre-flight

**Permissions checked:** I read `~/.local/state/claude-denials.jsonl` on 2026-11-17 for
2026-10-18 to 2026-11-17. The log holds 130 rows, from 2026-11-01T00:55Z to 2026-11-18T06:27Z.
The round's sessions run these commands that the log shows refused, or that the log has never
seen. Each has its alternative.

| Command | Refusals in the log | Alternative |
| --- | --- | --- |
| `claude stop <id>` chained with another command | 5 rows | Run it alone. In this round the built `cc-desk` binary runs it, and no session types it |
| `claude --bg` from a session | 1 row | Only the `cc-desk` binary launches, and only in pass 5's live run, under the entry below |
| `claude agents --json` piped into a script | 1 row | Redirect to a file and read the file. The tests use a stand-in |
| `gh pr merge` | 1 row, and `claude-merge-guard` denies it for every session | No session runs it. Tests use a `gh` stand-in. In the live run the check merges, started by the timer |
| `git worktree add` or `remove` chained | 2 rows | Run alone |
| `systemctl --user enable`, `systemd-run --user --scope` | None in the log | Covered by the entry below. A refusal parks pass 5 |
| Any command, in a classifier outage | 47 rows | Wait 60 seconds and run it once more |
| An edit to a Claude settings file | 11 rows | No pass makes one. The settings file a launch carries is written by the `cc-desk` binary into the state folder, which the proof found is allowed |

No row refuses `go build`, `go test`, `go vet`, `gofmt`, `bash scripts/gate.sh`, `git push` of a
pass branch, or `gh pr create`. The last round ran all of them here.

**Reviewed:** the design had its builder check on 2026-11-17, and every finding is folded.
Geoff's start of the conductor is his approval of the design, as `docs/STATUS.md` says. Each pass's plan gets one `standards:plan-reviewer` read.

**Evidence:** the shared assumptions and the launch checks below.

**Steps for Geoff:** three.

1. `[you]` Start the conductor from the prompt in `docs/STATUS.md`.
2. `[you]` Approve the auto-mode entry under "The auto-mode entry". Passes 1 to 4 run without
   it. Pass 5 doesn't start until `claude auto-mode` shows it.
3. `[you]` Accept Claude Code's trust prompt once in `~/Projects/cc-desk-scratch`: run `claude`
   there, accept, and leave. The designing session cloned the repo on 2026-11-17. Pass 5 doesn't
   start until `~/.claude.json` shows the folder trusted.

### The auto-mode entry

One entry in `autoMode.allow` in `~/.claude/settings.json`, in the form of the Prove pass's:

```text
cc-desk-sessions-page: under the approved cc-desk Sessions and Page designs
(docs/specs/2026-11-17-sessions-design.md and docs/specs/2026-11-17-page-design.md in
/var/home/glw907/Projects/cc-desk) and their round plans, the session `cc-desk-sessions-timer`
and the conductor of those rounds may run the `cc-desk` binary built from that repo. Through
that binary, and only on sessions it launched and recorded, they may launch background sessions
named `cc-desk-scratch-*` with their working folder under
/var/home/glw907/Projects/cc-desk-scratch/.claude/worktrees/, each with `--permission-mode
auto` and a settings file the binary writes, holding only allow rules that each name one
command the pass's steps run and a PermissionRequest hook that records the request and returns
no decision. They may end those sessions, remove them, read their logs, and end one such
session mid-run for the recovery test. They may install and enable the user units
cc-desk-check.service and cc-desk-check.timer from that repo, and the check those units run may
merge a pull request on the private repo glw907/cc-desk-scratch, and no other, after a check
has passed on its last commit. The launched sessions may push branches and open pull requests
on glw907/cc-desk-scratch only. The sessions `cc-desk-page-surface` and `cc-desk-page-link`
may install and enable the user unit cc-desk-page.service, listening on 127.0.0.1:7908, and
register a desktop handler for the cc-desk:// scheme. None of them acts on a session cc-desk
did not launch, runs `claude respawn --all` or `claude daemon`, or launches a session with
`bypassPermissions` or `--dangerously-skip-permissions`.
```

## The passes

Each pass's acceptance list is the named items of the design's "What done looks like", which
its planner copies word for word. Every pass also carries item 8, the gate, and writes its pull
request body and its STATUS lines by `~/.claude/docs/voice/written-for-people.md`.

| Pass | Slug | Outcome | Acceptance | Files | Class | Minutes | Depends on | Shares a file with |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | `sessions-check` | The command records sessions and pull requests in the release file, reads the session list, and decides and records what each session needs | Item 3 | `internal/state/` (types, strict checks, stop reasons), a new package for the session list and the decisions, `internal/snapshot/`, `cmd/cc-desk/` (the check verb), fixtures | `ordinary` | 90 | Nothing | 2, 3, 4, 5 |
| 2 | `sessions-launch` | A start launches a session in a worktree the command made, and the check ends and relaunches sessions | Items 1 and 2 | A new package for worktrees and launches, the prompt templates, the start verb, the check's actions, fixtures | `runner` | 90 | 1 | 1, 3, 4, 5 |
| 3 | `sessions-prompts` | A prompt is recorded, answered from a desk session, and filed, and a session files a follow-up through the command | Item 4 | The hook verb, the answer verb, the allow-rules verb, the follow-up verb, `internal/state/`, fixtures | `runner` | 60 | 2 | 1, 2, 4, 5 |
| 4 | `sessions-merge` | The check merges a clean pull request and sends back or raises any other | Item 5 | The ready verb, the merge decisions and actions, `internal/state/`, fixtures | `runner` | 60 | 3 | 1, 2, 3, 5 |
| 5 | `sessions-timer` | The timer runs the check, and a pass in the scratch repo runs to its merge with no prompt | Items 6 and 7 | `systemd/`, the timer verbs, the README | `runner`, `live-account` | 75 | 4 | 1 to 4 |

Passes 2 to 5 are `runner`: they write the settings a session launches with, answer its
prompts, merge, and install a unit. Pass 5 is `live-account` because it pushes to and merges in
`glw907/cc-desk-scratch`.

No pass in this round edits `scripts/gate.sh`, `.github/workflows/`, or anything under `web/`.
Those are the Page round's. A pass that finds it must edit one stops.

None of the checks can run today: the command has no `check` verb and no session record. Each
planner runs its checks on the code as the earlier passes left it, before any code is written.

## Shared assumptions

| Assumption | Passes | Probe or source |
| --- | --- | --- |
| `claude agents --json --all` is the supported way to read session state, and a row carries `state`, `status`, and `waitingFor` | 1, 2 | Claude Code docs, "agent view": "the supported way to read session state from outside Claude Code". A live row on pavilion, 2026-11-17, read `"status": "busy", "state": "blocked"` |
| A permission prompt reads `waitingFor: "permission prompt"` | 1 | Prove findings, proof 2: two live samples 60 seconds apart |
| A lost connection shows only on the `claude logs` screen, and an expired sign-in or a usage limit shows in the session's `state.json` `detail` | 1 | Prove findings, proof 2. The two `detail` texts are recorded, not live: the open-tests file says so. Pass 1 builds from the recorded texts and falls to "unreadable" for anything else |
| A blocked session can't be woken, and `resume` leaves a second row | 1, 2 | Prove findings, proofs 1 and 6 |
| `claude --help` lists `--bg`, `--name`, `--model`, `--effort`, `--permission-mode`, `--settings`, `--disallowedTools`, and `--worktree` | 2 | Run on 2026-11-17, Claude Code 2.1.327 |
| Auto mode with a `--settings` file of allow rules runs a pass with no prompt, and the prompt must come before `--disallowedTools` | 2, 5 | Prove findings, proofs 3 and 4. The proof launched with `--worktree`. A launch from a worktree `git` made is proof 5's case. Pass 2's first task runs one such launch through the stand-in and pass 5 runs it live |
| A `PermissionRequest` hook that prints no decision leaves the normal prompt up | 3 | Claude Code docs, hooks: exit 0 with no output is "no decision; normal permission flow applies". Not run here: pass 3's first task records one real hook input from a scratch session, which also gives the test its fixture |
| A process started with `systemd-run --user --scope` outlives the unit that started it, and a plain child doesn't | 2, 5 | Probe on pavilion, 2026-11-17: the scope's `sleep 120` was still running after the one-shot exited, the scope read `active`, and the plain `sleep 121` was gone |
| Claude Code's supervisor stays in the control group of whatever first started it | 2, 5 | `/proc/2989551/cgroup` on 2026-11-17 named a `ptyxis-spawn-…scope`. Not run: a supervisor first started inside a scope by `cc-desk`. Pass 5's first task runs it. If the supervisor dies with the service anyway, the service takes `KillMode=process`, and the plan says so |
| A user unit's `PATH` has neither `claude` nor `gh` | 5 | `systemd-run --user … bash -lc 'command -v claude'` printed "command not found", with `PATH=/usr/lib64/ccache:/usr/local/sbin:/usr/local/bin:/usr/bin`. The same unit ran `/var/home/linuxbrew/.linuxbrew/bin/claude agents --json` by its full path |
| The user's systemd runs from boot | 5 | `loginctl show-user` printed `Linger=yes` |
| The scratch repo is private, has a `gate` workflow, and its default branch is `main` | 5 | `gh repo view glw907/cc-desk-scratch`: `private=true main`; `.github/workflows/gate.yml` exists |
| Go is the same locally and in CI | all | `go1.27.4` both, as the last round recorded |

## Order table

Passes run one at a time within this round, in this order, and merge in this order. One pass
of the Page round runs beside each.

| Pass | Planner prompt | Depends on | Planned minutes | Waits for Geoff |
| --- | --- | --- | --- | --- |
| 1 `sessions-check` | `2026-11-17-sessions-1-planner-prompt.txt` | Nothing | 90 | no |
| 2 `sessions-launch` | `2026-11-17-sessions-2-planner-prompt.txt` | 1 | 90 | no |
| 3 `sessions-prompts` | `2026-11-17-sessions-3-planner-prompt.txt` | 2 | 60 | no |
| 4 `sessions-merge` | `2026-11-17-sessions-4-planner-prompt.txt` | 3 | 60 | no |
| 5 `sessions-timer` | `2026-11-17-sessions-5-planner-prompt.txt` | 4 | 75 | no, once pre-flight steps 2 and 3 read done |

A pass lands as soon as `round-run.md` "Reading a pass" lets it, and the next worktree is cut
from `origin/main`.

## Usage ceiling

100 percent of the weekly allowance (Geoff, 2026-11-17). The figure read 91 percent on
2026-11-17, with the reset at Nov 20, 10:59pm.

## Standing answers

The twelve in `round-run.md` hold, with these changes and additions.

- **4, replaced.** At 100 percent or more, or when the figure can't be read, the conductor
  finishes the passes in flight, starts nothing new, tells Geoff in the terminal, and waits.
  His word in the terminal resumes the round.
- **1 and 9, changed by `pass-core` "Until cc-desk runs".** Nothing runs `approval-queue send`.
  The conductor asks a stop's question in Geoff's terminal session and records it under "Open
  decisions" in `docs/STATUS.md`.
- **6, changed.** Two rounds run in this repo, so the busy check on the main checkout sees the
  other round's sessions. A hit that names a session or a process of the other round's pass in
  flight, with its folder under that pass's worktree, isn't a stop. Any other hit is.
- **13. Pre-flight step 2 or 3 isn't done when pass 4 lands.** The conductor tells Geoff which,
  in the terminal, and waits. The Page round goes on.
- **14. A pass's plan comes out over twice its minutes.** It parks, as `round.md` "Parking at
  planning" says.
- **15. A later pass needs a change in an earlier pass's files.** Expected: the passes share
  the command, the release file, and the fixtures. The pass makes the change and lists it.
- **16. A pass would change a fixture the last round's tests read.** It adds a fixture and
  leaves the old one as it is, because the Page round's tests print snapshots from them.
- **17. A live run in pass 5 meets a permission prompt.** That is a finding, not a failure of
  the pass. The prompt shows as "Needs you now" in `cc-desk snapshot`. The conductor runs in
  Geoff's terminal, so it can register as a desk. It reads the snapshot at each wait on pass
  5 and answers yes when the command is one the scratch pass's steps run. Any other prompt
  goes to Geoff. The executor files the prompt and runs the live pass once more, and the
  report gives every count. A third run with a prompt parks the pass.
- **18. The live checks.** Item 6 runs on the workstation, outside CI, and is never a Go test
  that skips. The round report gives its output.
- **19. The scratch release.** Pass 5's live run uses the real state folder,
  `~/.local/state/cc-desk`, with one release for the project `cc-desk-scratch`. It is the only
  release there, and it stays as a finished release after the run. Its allow rules arrive in
  the input file the executor adds it from.
- **20. The report.** The conductor writes it at
  `~/Projects/initiatives/cc-desk/reports/sessions-report.md` and gives Geoff its first
  paragraph in the terminal. It counts the permission prompts, as the route asks.

## Start line

The first planner launches when all of these read true.

1. `docs/STATUS.md` on `origin/main` names the conductor prompt as the next action.
2. `claude agents --json` lists no other session named `conductor-cc-desk` and none named
   `cc-desk-sessions-*` or `cc-desk-page-*`.
3. The main checkout is on `main`, and `git rev-list --count HEAD..origin/main` prints 0 after a
   fetch.
4. `claude-worktree-busy /var/home/glw907/Projects/cc-desk` exits 0.
5. The weekly usage figure is under 100.

The Page round is the only other round approved in this repo, and it starts on the same line.
The State and rules round closed with pull request 6.

### Launch checks, run on 2026-11-17

| Check | Output |
| --- | --- |
| `claude -p "/usage" --model haiku < /dev/null` | "Current week (all models): 91% used · resets Nov 20, 10:59pm" |
| `git status -sb` after a fetch | `## main...origin/main`, with the two new designs uncommitted. The commit that carries this file levels it |
| `claude-worktree-busy .` | Exit 1, "uncommitted file docs/specs/2026-11-17-page-design.md edited in the last 60 minutes". That is this session's writing. The conductor runs it again |
| `git check-ignore -q .claude/worktrees/x` | exit 0 |
| `git worktree list` | The main checkout only |
| `claude --help` | Lists `--bg`, `--model`, `--effort`, `--name`. `claude --advisor fable --version` printed `2.1.327` with no error |
| `claude agents --json` | A blocked `cairn-cms-tool-delete` in another repo, a finished `planner-executor-handoff-pass-1` in this checkout, and this session. None is this round's. The conductor reads the list again before the first launch |

## Decisions

- One conductor runs both rounds from one prompt, one pass of each at a time. That is the
  host's limit of two building passes.
- Five passes, not the route's four, for the reasons the design gives.
- The live run uses the real state folder. Nothing else is in it yet.
