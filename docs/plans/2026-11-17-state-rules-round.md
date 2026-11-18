# State and rules: the round plan

This file is what `pass-core` asks a round of several passes to carry on top of its design. The
design is [2026-11-17-state-and-rules-design.md](../specs/2026-11-17-state-and-rules-design.md),
which Geoff approved in his terminal session on 2026-11-17. Where the two differ, the design wins.
A planner writes each pass's plan just before the pass runs.

- Initiative: `cc-desk`
- Round id: `state-rules`
- Repo: `/var/home/glw907/Projects/cc-desk`, pushing to `glw907/cc-desk`, default branch `main`
- Conductor: `conductor-cc-desk`, a session Geoff starts in his terminal on Opus (his message of
  2026-11-17: "I'll restart and execute in a fresh opus session")
- Conductor prompt: `docs/plans/2026-11-17-state-rules-conductor-prompt.txt`
- Chain: 315 planned minutes. The round starts nothing new past 630.

## Pre-flight

**Permissions checked:** I read `~/.local/state/claude-denials.jsonl` on 2026-11-17 for 2026-10-18
to 2026-11-17. The log holds 129 rows, from 2026-11-01T00:55Z to 2026-11-18T03:06Z, so all of it
falls inside the window. Its classes, with row counts:

| Class | Rows |
| --- | --- |
| Classifier unavailable | 47 |
| Destructive or Unwanted Actions | 14 |
| Self-Modification, Interfere With Workloads | 11 each |
| Server-side classifier, no reason given | 10 |
| Undermining Oversight, Unauthorized Access | 4 each |
| Merge Without Review, Auto-Mode Bypass | 3 each |
| Irreversible Local Destruction, Git Destructive, Create Public Surface, Unverifiable Deletion Scope, Sensitive-Source Provenance, Security Weaken | 2 each |
| Production Deploy, Secret-Store Writes, DNS / Domain / Cert Changes, Bot Check Circumvention, Code from External, Shared Scratch Sweep, Safety Bypass Flag, Out-of-Place Publication, PII Data Handling, Modify Shared Resources | 1 each |

The round's sessions run these commands that the log shows refused. Each has its alternative.

| Command | Refusals in the log | Alternative |
| --- | --- | --- |
| `claude stop <id>` chained with another command | 3 rows, 2026-11-10 to 2026-11-13 | Run it alone, as `round-run.md` says. A refusal is reported and never retried |
| `claude agents --json` piped into a script | 1 row, 2026-11-17 | Redirect the output to a file, then read the file with `python3 -I` |
| `gh pr merge` | 3 rows, Merge Without Review | No session runs it. A pull request lands through `approval-queue land` |
| Any command, in a classifier outage | 47 rows | Wait 60 seconds and run it once more |
| An edit to a Claude settings file | 11 rows, Self-Modification | No pass makes one. A pass that needs one stops |
| Fetching code from outside | 1 row, Code from External | Pass 1 adds cobra with `go get github.com/spf13/cobra`. If that is refused twice, the pass stops as a step only Geoff can do |

No row refuses `go build`, `go test`, `go vet`, `gofmt`, `bash scripts/gate.sh`, `git push` of a
pass branch, or `gh pr create`. The Prove pass ran all of them in this repo on 2026-11-17.

**Reviewed:** the design had its builder check before Geoff read it, and every finding is folded.
This file adds no mechanism to it. Each pass's plan gets one `standards:plan-reviewer` read, as
`round.md` "A plan inside a round" asks.

**Evidence:** the shared assumptions and the launch checks below.

**Steps for Geoff:** one. He starts the conductor from the prompt in `docs/STATUS.md`. Nothing
else needs him: no secret, no setting, no spend past the usage stop. The repo's trust prompt was
accepted during the Prove pass.

## The passes

Each pass's acceptance list is the named items of the design's "What done looks like", which its
planner copies word for word, plus the items written here. Every pass also carries done item 8,
the gate, and writes its pull request body and its STATUS lines by
`~/.claude/docs/voice/written-for-people.md`.

| Pass | Slug | Outcome | Acceptance | Files | Class | Minutes | Depends on | Shares a file with |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | `state-rules-files` | The command reads, checks, and writes every state file, records pass events, and prints a snapshot | Done item 5, and items A to D below | `cmd/cc-desk/`, new packages for the state and the snapshot, the fixtures, `scripts/gate.sh`, `go.mod`, `go.sum` | `runner` | 90 | Nothing | 2, 3, 4 |
| 2 | `state-rules-estimate` | The snapshot carries the next session's time, range, and state, and each event appends a forecast | Done item 1 | The estimate's files, the snapshot builder, the fixtures | `ordinary` | 75 | 1 | 1, 3, 4 |
| 3 | `state-rules-order` | A start is started, held, or refused by the rules, and the command reads usage | Done items 2, 3, and 7 | The rules' files, the usage reader, the start verb, the fixtures | `ordinary` | 90 | 2 | 1, 2, 4 |
| 4 | `state-rules-rulings` | A ruling is recorded only from a desk session, and it frees what waited | Done items 4 and 6 | The rulings and guard files, the ruling verb, the fixtures | `ordinary` | 60 | 3 | 1, 2, 3 |

Pass 1 is `runner` because it changes `scripts/gate.sh`, the gate itself. The guard in pass 4 is
`ordinary`: the design says it stops a mistake and isn't built to stop a forgery.

Pass 1's further items:

| Item | What Geoff would see |
| --- | --- |
| A. Events | On a fixture, adding a release and printing the snapshot shows its rounds and passes with their states. Start, finish, and stop each change the pass's state and record the time the test's clock gave |
| B. Bad requests | Finishing a pass twice, or before it started, is refused, and the state folder is byte for byte what it was. A verb that names a project, round, pass, or question that doesn't exist exits with that name |
| C. The lock | While a test holds the state lock, a second command exits with an error after 10 seconds and writes nothing |
| D. Skipped tests | `bash scripts/gate.sh` fails when any Go test reports a skip. The test for this adds a skipping test in a scratch copy and shows the gate's exit code |

None of these checks can run today. The command is a 24-line version printer, and every check
runs the built binary against fixtures that pass 1 creates. Each planner runs its checks on the
code as the earlier passes left it, before any code is written.

### Why the passes run in one chain

The design left open which of passes 2 to 4 run side by side. None do. Pass 2 builds the "Ready
now" state that done item 2 reads from the snapshot. Pass 3 builds the 200-minute guess that done
item 4 rules on. Item 6 runs rulings beside the other events with the slow usage stand-in, and
only pass 4 has all of those. All four also add verbs to the same command and fields to the same
snapshot.

The least-proven assumption is the guard's terminal test, and it sits in the last pass because
of those dependencies. A wrong assumption there costs pass 4 alone, since nothing builds on it
in this round.

## Shared assumptions

| Assumption | Passes | Probe or source |
| --- | --- | --- |
| A command run by Claude Code's Bash tool has `CLAUDE_CODE_SESSION_ID` set, has no terminal, and has the `claude` process as its parent, which has one | 1, 4 | Probe in a terminal session, 2026-11-17, recorded in the design under "Rulings and the guard" |
| A test can give a process a controlling terminal with no new dependency | 4 | `script -qec 'awk "{print \"tty_nr=\" \$7}" /proc/self/stat' /dev/null` printed `tty_nr=34816` on pavilion, 2026-11-17. The same `awk` under `setsid` with no terminal printed `tty_nr=0`. Not yet run on the CI runner: pass 4's first task runs it there |
| The usage read costs nothing and names the week's figure | 3 | `claude -p "/usage" --model haiku --output-format json < /dev/null`, 2026-11-17: the result held "Current week (all models): 88% used · resets Nov 20, 11pm (America/Anchorage)", with `num_turns` 0 and `total_cost_usd` 0 |
| Go is the same locally and in CI | all | `go version` printed `go1.27.4`, and `.github/workflows/gate.yml` pins `go-version: 1.27.4` |
| The repo has no dependency yet | 1 | `go.mod` holds the module line and `go 1.27` only |
| The gate runs `go test ./...` with no check for skips | 1 | `scripts/gate.sh`, the "gate: go test" step |

## Order table

Passes run one at a time, in this order, and merge in this order.

| Pass | Planner prompt | Depends on | Planned minutes | Waits for Geoff |
| --- | --- | --- | --- | --- |
| 1 `state-rules-files` | `2026-11-17-state-rules-1-planner-prompt.txt` | Nothing | 90 | no |
| 2 `state-rules-estimate` | `2026-11-17-state-rules-2-planner-prompt.txt` | 1 | 75 | no |
| 3 `state-rules-order` | `2026-11-17-state-rules-3-planner-prompt.txt` | 2 | 90 | no |
| 4 `state-rules-rulings` | `2026-11-17-state-rules-4-planner-prompt.txt` | 3 | 60 | no |

No pass is marked to wait. A pass lands as soon as `round-run.md` "Reading a pass" lets it, and
the next worktree is cut from `origin/main`. A pass that can't land yet holds the later landings,
and each later worktree is cut from the branch of the pass before it.

## Usage ceiling

100 percent of the weekly allowance (Geoff, 2026-11-17, quoted in the design). It replaces the 85
percent in `round-run.md` standing answer 4. The figure read 88 percent on 2026-11-17, with the
reset at Nov 20, 11pm.

## Standing answers

The twelve in `round-run.md` hold, with these changes and additions.

- **4, replaced.** At 100 percent or more, or when the figure can't be read, the conductor
  finishes the pass in flight, starts nothing new, tells Geoff in the terminal, and waits. Geoff
  has a reset he means to use, and his word in the terminal resumes the round.
- **1 and 9, changed by `pass-core` "Until cc-desk runs".** Nothing runs `approval-queue send`.
  The conductor asks a stop's question in Geoff's terminal session and records it under "Open
  decisions" in `docs/STATUS.md`. His answer there is the ruling.
- **13. The CI runner gives a test no controlling terminal.** Pass 4 learns this in its first
  task. Done item 4's outcome stands, so the session asks its advisor for a check that shows the
  same outcome and flags it, as `advisor.md` "Acceptance items" sets out. It never skips the
  test, because the gate fails on a skip.
- **14. A pass's plan comes out over twice its minutes.** It parks, as `round.md` "Parking at
  planning" says.
- **15. A later pass needs a change in an earlier pass's files.** That is expected: the passes
  share the command, the snapshot, and the fixtures. The pass makes the change and lists it in
  its pull request.
- **16. The live usage check of done item 7.** It runs on the workstation, outside CI, and it
  must not be a Go test that skips. The round report gives its output.
- **17. The report.** The conductor writes it at
  `~/Projects/initiatives/cc-desk/reports/state-rules-report.md` and gives Geoff its first
  paragraph in the terminal. It sends nothing to the queue.

## Start line

The first planner launches when all of these read true.

1. `docs/STATUS.md` on `origin/main` names the conductor prompt as the next action.
2. `claude agents --json` lists no other session named `conductor-cc-desk` and none named
   `cc-desk-state-rules-*`.
3. The main checkout is on `main`, and `git rev-list --count HEAD..origin/main` prints 0 after a
   fetch.
4. `claude-worktree-busy /var/home/glw907/Projects/cc-desk` exits 0.
5. The weekly usage figure is under 100.

No other round is approved or running in this repo. The Prove round closed with pull request 2.

### Launch checks, run on 2026-11-17

| Check | Output |
| --- | --- |
| `claude -p "/usage" --model haiku < /dev/null` | "Current week (all models): 88% used" |
| `git status -sb` after a fetch | `## main...origin/main [ahead 1]` before this file's commit was pushed. The push that carries this file levels it |
| `claude-worktree-busy .` | exit 0 |
| `git check-ignore -q .claude/worktrees/x` | exit 0 |
| `git worktree list` | The main checkout only |
| `claude --help` | Lists `--bg`, `--model`, `--effort`, `--name`. It doesn't list `--advisor`. `claude --advisor fable --version` printed the version with no error, and another round's conductor launched planners with the flag on 2026-11-17 |
| `claude agents --json` | Three rows: one in another repo, a finished `preflight-checks-spike-launch`, and `cc-desk-e2` with no state. Neither cc-desk row is this round's. The conductor reads the list again before the first launch |

## Decisions

- The conductor runs in Geoff's terminal session on Opus, not as a background Sonnet session.
  He said he would run it that way, and under "Until cc-desk runs" his terminal is where
  questions and reports reach him.
- The four passes run as one chain, for the reasons above.
- Pass 1 holds done item 5 for every file, so it builds the strict reader for `desks.json`,
  `settings.json`, and `rulings.jsonl` as well as the release files.
