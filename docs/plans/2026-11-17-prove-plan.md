# Prove first: plan

This pass sets up the cc-desk repo's scaffold and gate, then tests the seven behaviors the design
lists under "Prove first", each in a throwaway build under `spikes/`. It ends in one findings
report that says, for each behavior, whether it held or which fallback the later rounds build.
No proof acts on a session it didn't launch, and none turns permission checks off.

- Design: [2026-11-17-cc-desk-design.md](../specs/2026-11-17-cc-desk-design.md), "Prove first"
- Route: [2026-11-17-route.md](2026-11-17-route.md), the Prove round
- Round: a one-pass round, with no conductor; the executor closes the pass
- Class: `runner`
- Executor: `cc-desk-prove`, in a worktree of `/var/home/glw907/Projects/cc-desk`, pushing to
  `glw907/cc-desk`
- Path estimate: 530 minutes

## Pre-flight

**Permissions checked:** I read `~/.local/state/claude-denials.jsonl` on 2026-11-17 for
2026-10-18 to 2026-11-17. The log holds 127 rows. The first is dated 2026-11-01T00:55Z and the last
2026-11-17T21:24Z, so the whole log falls inside the window. Its classes, with row counts:

| Class | Rows |
| --- | --- |
| Classifier unavailable | 47 |
| Destructive or Unwanted Actions | 13 |
| Interfere With Workloads | 11 |
| Self-Modification | 10 |
| Server-side classifier, no reason given | 10 |
| Undermining Oversight | 4 |
| Unauthorized Access | 4 |
| Merge Without Review | 3 |
| Auto-Mode Bypass | 3 |
| Irreversible Local Destruction, Git Destructive, Create Public Surface, Unverifiable Deletion Scope, Sensitive-Source Provenance, Security Weaken | 2 each |
| Production Deploy, Secret-Store Writes, DNS / Domain / Cert Changes, Bot Check Circumvention, Code from External, Shared Scratch Sweep, Safety Bypass Flag, Out-of-Place Publication, PII Data Handling, Modify Shared Resources | 1 each |

The tasks run these commands that the log shows refused. Each has its alternative.

| Command a task runs | Refusals in the log | Alternative |
| --- | --- | --- |
| `claude stop <id>` (tasks 3 to 9, 11) | 5 rows dated 2026-11-10 to 2026-11-16, classed Interfere With Workloads, Destructive or Unwanted Actions twice, Undermining Oversight, and one with no reason | Run it only through `spikes/own.sh`, under the auto mode entry in pre-flight item 2. If it is still refused, leave the owned session to finish its bounded prompt and let idle retirement end it, and list it in the findings |
| `kill <pid>` (task 7) | 5 rows dated 2026-11-09 to 2026-11-17, all classed Interfere With Workloads but one Destructive | The same helper and entry. If still refused, proof 6 uses `claude stop` on the owned session as its stand-in for a death and says so |
| `claude agents --json` piped into a script (every proof) | 1, Destructive or Unwanted Actions, 2026-11-17 | Run `claude agents --json --all` alone with its output redirected to a file under `spikes/`, then read the file with `python3 -I`. The allow rule `Bash(claude agents --json *)` covers that form |
| `claude --bg` (every proof) | 1, Classifier unavailable, 2026-11-17 | Wait 60 seconds and run it once more |
| `git push origin HEAD:main` | 1, no reason given | The pass pushes only the pass branch and lands its PR with `approval-queue land` |
| `gh pr merge` | 3, Merge Without Review; `claude-merge-guard` also denies every `gh pr merge` | No task runs it. Proof 4 and the close land a PR with `approval-queue land` |
| An edit to a Claude settings file | 10, Self-Modification | None is made. A launch setting goes in a file under `spikes/` passed to the owned session's launch, which the auto mode entry covers |

**Reviewed:** one `standards:plan-reviewer` read ran on 2026-11-17. Its verdict was "not ready to
launch", and all 16 of its findings are folded into this plan, the executor prompt, and the route.

**Evidence:** the mechanism table and the recorded searches below.

### Steps for Geoff

1. **[you] Trust the repo.** Run `claude` once in `/var/home/glw907/Projects/cc-desk` and accept the
   workspace trust dialog. No session has opened this folder: `~/.claude.json` held no entry for
   it on 2026-11-17. Check: that file's `projects` entry for the path has
   `hasTrustDialogAccepted` set to true, and the executor's launch prints no "Workspace not
   trusted".
2. **[you] Approve one auto mode entry.** The proofs must stop, restart, and remove the sessions
   they launch, end their processes, and land one PR on the scratch repo. The log shows those
   commands refused. Geoff approves the text below in a session, and that session lands it in
   `autoMode.allow` through the `claude-infra` skill.
   Standing answer 3 in `pass-core` `references/round-run.md` makes this a pre-launch step, never
   an executor task. Check: `claude auto-mode config` prints an entry that opens
   `cc-desk-prove-owned:`.

   > cc-desk-prove-owned: under the approved cc-desk design
   > (/var/home/glw907/Projects/cc-desk/docs/specs/2026-11-17-cc-desk-design.md, "Prove first")
   > and its plan (docs/plans/2026-11-17-prove-plan.md in that repo), the background session
   > `cc-desk-prove` may launch background sessions named `ccd-spike-*` with their working
   > folder in its worktree under /var/home/glw907/Projects/cc-desk/.claude/worktrees/. Each
   > launch passes a settings file under spikes/ in that worktree. That file may hold only
   > exact allow rules, each naming one command its proof's steps run; an `env` block that
   > points a proxy at 127.0.0.1; a status line command; and a PermissionRequest hook that
   > denies unless a decision file under spikes/ names the request. Through its script
   > spikes/own.sh, and only on the sessions that script recorded at launch, it may run
   > `claude stop`, `claude respawn <id>`, `claude rm`, `claude logs`,
   > `claude --resume <id> --bg`, and `kill` of such a session's process. It may run a CONNECT
   > tunnel on 127.0.0.1 that only those sessions use. On the private repo
   > glw907/cc-desk-scratch, and no other, its sessions may push branches, open pull requests,
   > run `gh pr ready`, and run `approval-queue land`. It never acts on any other session,
   > never runs `claude respawn --all` or `claude daemon`, and never launches a session with
   > `bypassPermissions` or `--dangerously-skip-permissions`.

### Checks the launching session runs

3. The desk leaves the proof sessions alone. The launching session adds one line to
   `~/.local/state/desk/notes.md` saying that sessions named `ccd-spike-*` belong to the prove
   pass and are never woken, stopped, or relaunched. Check: `grep -c ccd-spike` on that file
   prints 1 or more.
4. The sign-in is valid. Check: `claude auth status --json` shows `loggedIn` true. On
   2026-11-17 it did, with `loginExpiresAt` 2026-12-15T15:40Z. If it reads false, signing in is a
   step only Geoff can do.
5. The main checkout is level. Check: after a fetch, `git -C /var/home/glw907/Projects/cc-desk
   rev-list --count HEAD..origin/main` prints 0 (it printed 0 on 2026-11-17), and
   `.claude/worktrees/` is in `.gitignore` (it is).
6. `claude-worktree-busy /var/home/glw907/Projects/cc-desk` exits 0.
7. A private scratch repo exists for proof 4's push, PR, and landing. The launching session
   creates `glw907/cc-desk-scratch` as a private repo with one commit on `main`: a toy Go
   package, a gate script, and `.github/workflows/gate.yml`, which runs the gate on every pull
   request. Check: `gh repo view glw907/cc-desk-scratch --json visibility` shows `PRIVATE`, and
   `gh workflow list -R glw907/cc-desk-scratch` lists the gate. If auto mode refuses the
   create, proof 4 pushes to a local bare repo instead, and the finding lists the PR and the
   landing as untested. The Retire round removes the scratch repo.

Nothing else needs Geoff. The `gh` token carries the `workflow` scope (probe below), so pushing a
workflow file needs no new grant. The pass installs no systemd unit and starts no service or
timer. It reads no secret and holds none.

## Acceptance

| Outcome | Task | Shown by |
| --- | --- | --- |
| The gate passes on the branch, and the PR's Actions run of the same gate is green | 1 | `bash scripts/gate.sh` output ending in exit 0; `gh pr checks` on the PR |
| The repo has a Go module, `scripts/gate.sh`, a workflow that runs it, `.vale.ini`, and a `docs/follow-ups/` folder, and has no `BACKLOG.md` | 1 | `git ls-files go.mod scripts/gate.sh .github/workflows .vale.ini docs/follow-ups` lists each; `test ! -e BACKLOG.md` exits 0 |
| `docs/STATUS.md` is 60 lines or fewer and names the findings report and the next action | 10 | `wc -l docs/STATUS.md`; the file's "Next action" line |
| Each of the seven proofs has a section in the findings report with its commands, their output, and exactly one closing line that reads "Held." or "Fallback applies: " and the design's fallback | 3 to 10 | `python3 -I spikes/closing.py docs/research/2026-11-17-prove-findings.md` prints `7 sections, one closing line each` and exits 0 |
| Each proof has a folder under `spikes/` with a script that reruns it | 3 to 9 | `ls spikes` shows seven numbered folders, each holding a `run` script |
| No session the pass didn't launch was stopped, woken, relaunched, or acted on | 2 to 11 | `bash spikes/own.sh audit` prints `strangers: 0`, `foreign kills: 0`, and `foreign restarts: 0`, and `spikes/audit.txt` holds the lists those counts rest on |
| No owned session ran with permission checks turned off | 2 to 9 | `bash spikes/own.sh audit` prints `bypass: 0`, and its `allow hooks:` line names no file but `spikes/4-permissions/decide-hook.py` |
| No proof session pushed to `glw907/cc-desk`; only proof 4's sessions pushed, and only to the scratch repo | 3 to 9 | `git ls-remote --heads origin` lists only `main` and the pass's branch; `gh pr list -R glw907/cc-desk --state all` lists only the pass's PR |
| A small pass ran from worktree to merge with zero prompts, or each remaining prompt is named with its cause | 5 | The proof 4 section's last run count reads 0, or its residue list names each prompt |
| Every denial class in the log has a line saying which session kind hits it and what removes it | 5 | The filled "Denials by class" table in the findings |
| The findings say whether a prompt can be answered from outside the session | 5 | The proof 4 section's "Answering a prompt from outside" paragraph |
| At close no proof session or proof worktree is left | 11 | `claude agents --json --all` saved to a file shows no row named `ccd-spike-*`; `git worktree list` shows only the main checkout and the pass's worktree |

## Mechanism table

Each row was probed on 2026-11-17 or is quoted from its source.

| Mechanism | Probe or source | What it showed |
| --- | --- | --- |
| Claude Code version | `claude --version` | `2.1.326 (Claude Code)` |
| The session list a script can read | `claude agents --json --all` | 392 rows with keys `cwd, id, kind, name, pid, sessionId, startedAt, state, status`. States seen: done 293, stopped 88, failed 8, working 1, and one interactive row. `status` held only `busy`, `idle`, or nothing: no reason text |
| Where scripts should read state | <https://code.claude.com/docs/en/agent-view> | "Scripts should read state with `claude agents --json --all`, not the files under `~/.claude/jobs/`." |
| Reason text for a stopped session | Read of `~/.claude/jobs/*/state.json` and `timeline.jsonl` | `state.json` holds `state`, `detail`, `tempo`, `needs`, `needsAsk`; timeline rows hold `at, detail, state, text`. Recorded texts: `API Error: Can't reach the API server — check your internet or DNS (EAI_AGAIN)` (2026-11-14), `Login expired` (2026-11-17T17:58Z), `You've hit your weekly limit · resets Nov 6, 11pm (America/Anchorage)`, and `You've hit your Fable limit · resets Nov 20, 11pm (America/Anchorage)` |
| A permission prompt in a background session | `~/.local/state/desk/efficiency-notes.md` line 182 | `tempo: blocked`, `needs: "approve Bash: python3 - <<'E' ..."` |
| What blocked means | agent-view docs, states table | Blocked is "waiting on you: a question it asked, a permission or sandbox decision, an error only you can clear such as an expired login, or its first prompt" |
| A session's recent screen | `claude logs --help` | "Print the background session's recent terminal output." |
| No scripted message to a session | agent-view docs | No shell command for sending a message is documented, and `claude attach` is interactive |
| Restarting one session | `claude respawn --help`; agent-view docs | "Restart a background session (or all of them)"; "A session that has a saved conversation resumes it." |
| Resuming under the same id | agent-view docs, v2.1.257 row | `claude --resume <session-id> --bg` continues the session under the same id, or starts a copy under a new id and prints a `note:` line |
| Resuming a live session starts a copy | `~/.local/state/desk/notes.md` line 53 | A `--resume` of a session still running started a second copy |
| The process behind a session | `ps -o pid,ppid,args` on a live session's `pid` | The row's `pid` runs `claude bg-spare`. Its parent runs `claude bg-pty-host`, and that one's parent runs `claude daemon run`, a child of the user's systemd |
| Proxy variables the supervisor holds | Variable names only in `/proc/<daemon pid>/environ`, the process id found from the process tree | No `HTTPS_PROXY`, `HTTP_PROXY`, `ALL_PROXY`, or `NO_PROXY` in upper case or lower case. Values weren't read |
| Nothing relaunches after a reboot | `~/.claude/daemon.log` | `2026-11-17T02:55:23 bg adopt: adopted=0 respawned=0 dead=9` |
| Ended sessions in the supervisor log | `grep -c` on `~/.claude/daemon.log` | 314 lines of the form `bg settled <id> (killed)` and 304 `bg retire` lines before the pass starts, so an audit counts only lines after the pass's first launch |
| Idle sessions are ended | agent-view docs; `daemon.log` | Unattached finished or waiting sessions stop after about an hour, while working or permission-blocked ones keep their process; log lines `bg retire <id>: idle-prompt, idle 60m` |
| Background sessions may push and open PRs | agent-view docs, v2.1.198 row | A background session that isolated its work in a worktree commits, pushes that branch, and opens a draft pull request when it finishes |
| Landing a PR | `approval-queue` source, `land.ts`; `~/.claude/settings.json` | `land` refuses a draft, a failing check, a base other than the default branch, a `## Before merge` section, or a waiting document. It polls checks for up to 900 seconds, merges with `--match-head-commit`, and records its sentence in the queue store. `Bash(approval-queue *)` is in the allow list |
| `gh pr merge` is denied | `~/.local/bin/claude-merge-guard`, a hook in `~/.claude/settings.json` | "PreToolUse hook denying every `gh pr merge`, so a merge goes through `approval-queue`" |
| A hook that decides a prompt | <https://code.claude.com/docs/en/hooks> | `PermissionRequest` fires "When a tool call needs a permission decision". Its answer is `decision.behavior`, `allow` or `deny`, in `hookSpecificOutput`; exit code 2 "isn't honored for this event". The page doesn't say whether it fires in a background session |
| `--disallowedTools` drops tools from context | <https://code.claude.com/docs/en/cli-reference> | "A bare tool name removes the matching tools from Claude's context" |
| Size of the unused tools | `efficiency-notes.md` lines 41 and 49 | Tool definitions are 82,000 characters; `Artifact`, `SendFeedback`, `Workflow`, `AskUserQuestion`, and `ReportFindings` come to 54,000 |
| Permission modes | cli-reference; <https://code.claude.com/docs/en/permission-modes> | `--permission-mode` takes `default`, `acceptEdits`, `plan`, `auto`, `dontAsk`, `bypassPermissions`, `manual`. In `dontAsk`, "Claude Code auto-denies every tool call that would otherwise prompt you" and "the session never waits for input". In auto mode a blocked action is denied. After 3 blocks in a row or 20 in all, auto mode pauses and goes back to prompting. "Narrow rules like `Bash(npm test)` stay in effect" in auto mode |
| `EnterWorktree` on an existing path prompts | `efficiency-notes.md` line 218; `pass-core` `references/executor.md` lines 11 and 12 | Both record a permission prompt that stalled the run |
| Launch settings for one session | cli-reference | `--settings` takes "Path to a settings JSON file or an inline JSON string" whose values override the same keys for that session |
| Proxy variables reach a session | <https://code.claude.com/docs/en/network-config> | "Claude Code respects standard proxy environment variables", and every variable on the page "can also be configured in `settings.json`". For background agents, set them "in the `env` block" of settings. "When the session's launch environment already sets a variable, Claude Code ignores the settings file's value" |
| Trust and background launches | agent-view docs, v2.1.281 row | `claude --bg` checks workspace trust, and "where no dialog can appear, such as in a script, the command exits with a `Workspace not trusted` error" |
| Trust is inherited below a trusted repo | Read of `~/.claude.json` `projects`; count of job folders | `/var/home/glw907/Projects/cairn-cms` is trusted; no entry exists for any path under a `.claude/worktrees/` folder; 14 jobs were launched with their folder inside one; no entry exists for `cc-desk` |
| Usage figures a script can see | <https://code.claude.com/docs/en/statusline> | The status line's input carries `rate_limits.five_hour.used_percentage`, `rate_limits.seven_day.used_percentage`, and their `resets_at`. It appears only for Pro, Max, and Ultra subscribers, and "only after the first API response in the session" |
| The account's plan | `claude auth status --json` | `subscriptionType` is `max`. Keys: `analyticsDisabled, apiProvider, authMethod, configDirectory, consoleProfileSignInOffered, email, loggedIn, loginExpiresAt, orgId, orgName, planTier, projectsDirectory, storedClaudeAiLogin, subscriptionType`. No usage figure |
| The house usage read | `~/.claude/docs/unattended-work-guards.md` line 145 | `claude -p "/usage" --model haiku < /dev/null`, read on the "Current week (all models)" line. Whether it spends model tokens is not recorded anywhere |
| Auto mode entries are prose rules | Read of `~/.claude/settings.json` | `autoMode` holds `allow` and `environment` lists of prose entries; `permissions.defaultMode` is `auto`; allow rules include `Bash(claude agents --json *)` |
| GitHub token scopes | `gh auth status` | `delete_repo, gist, read:org, repo, workflow` |
| Tools on the host | `go version`; `vale --version`; `command -v tellgrader` | go1.27.4; vale 3.27.0; `/var/home/glw907/.local/bin/tellgrader` |
| Vale in CI, prior art | `/var/home/glw907/Projects/dubplate/.vale.ini`; `dubplate/.github/workflows/check.yml` lines 23, 24, and 78 to 80 | The Survey package from a pinned release, and vale 3.27.0 fetched with a pinned SHA-256 |
| Vale finds no config in cc-desk | `vale README.md` in the repo | "0 errors, 0 warnings and 0 suggestions in 0 files": the repo needs a `.vale.ini` |

## Recorded searches

| Search | Hits | Each hit |
| --- | --- | --- |
| `git ls-files` in cc-desk | 27 | `.gitignore`: task 1 adds the sandbox folders. `README.md`: dismissed, unchanged. `docs/STATUS.md`: task 10. The design: read only. 23 files under `docs/specs/2026-11-17-cc-desk-mockups/`: dismissed, the Page round's reference, left out of the vale run |
| `git ls-files` for `BACKLOG.md`, `.github`, `go.mod`, `.vale.ini`, `spikes`, `scripts`, `docs/research`, `docs/follow-ups` | 0 for each | Each is new, made by tasks 1 to 10 |
| `grep -rniE 'backlog\|follow-?up'` over the repo's Markdown, mock-ups left out | 7, all in the design (lines 203, 204, 271, 297, 298, 315, 349) | Lines 203, 204, 271, 297, and 298 state the rule task 1 follows: one file per follow-up and no `BACKLOG.md`. Line 315 names two follow-ups outside cc-desk, dismissed as workstation items. Line 349 is about cairn's re-cut, dismissed |
| `grep` for `seven_day`, `five_hour`, `rate_limits`, `used_percentage` in `~/.dotfiles`, `~/.local/bin`, and `approval-queue` | 0 | No existing usage reader to reuse |
| `grep -l` for the same words in JSON under `~/.claude`, leaving out `projects/`, `jobs/`, and `plugins/cache` | 0 | No usage file on disk |
| Live sessions, `claude agents --json` | 6 at probe time | `desk`, `dubplate-intake-design`, `dubplate-structure-design`, `dubplate-engine`, `cairn-cms-tool-delete`, and one interactive session. All dismissed: never acted on by this pass |
| `ready.md` "A retired form" search | Not run | This pass retires no command or form |

## Tasks

Every proof session is launched and controlled through `spikes/own.sh` (task 2). It runs on Sonnet
at low effort and is named `ccd-spike-<proof>-<word>`. At most two are alive at once.

Each session works in a sandbox: a fresh git repo with no remote, under its proof's folder and
ignored by the outer repo. A session that isolates its work there has nowhere to push and can't
open a PR. Proof 4's main session, which works in a clone of the scratch repo, and proof 5's
session, which makes no edits, are the exceptions. Every prompt ends in a bounded task. A session
left alone finishes and goes idle.

Proof 5 runs first, because every later launch sits in a worktree and its result says where those
launches can go. Then the proofs run in order of how little is known about them: proof 1 has no
documented mechanism, proof 4 is contradicted by the records, and proofs 3 and 7 have documented
mechanisms. Proof 2 comes after 1 and 4 because it reads the states those two produce. Proof 6
follows it, and proofs 3 and 7 come last.

There are no pairs. Every proof drives live sessions through one owned list, and a pair would
share that list and the host's two-building-pass limit.

Each Bash call a task or a proof session makes stays under the Bash tool's default timeout of
120,000 ms, or sets its timeout explicitly.

### 1. Scaffold and gate

- Outcome: the repo builds and checks itself locally and in CI.
- Files: `go.mod`, `cmd/cc-desk/main.go`, `cmd/cc-desk/main_test.go`, `scripts/gate.sh`,
  `.github/workflows/gate.yml`, `.vale.ini`, `.gitignore`, `docs/follow-ups/.gitkeep`.
- Class: `runner`.
- Work: a stdlib-only `cc-desk` that prints its version, so the gate has a real package. The
  gate runs a `gofmt` check, `go vet`, `go test`, and vale over `docs/` with the mock-ups left out.
  `.vale.ini` takes the Survey package as `dubplate` pins it. The workflow installs Go and vale
  3.27.0 as the `dubplate` repo's `check.yml` does, with its pinned SHA-256, and runs the gate.
  `.gitignore` gains the proofs' sandbox folders. `docs/follow-ups/` holds one Markdown file per
  item, named `<date>-<slug>.md`, and starts with `.gitkeep` until task 10 adds an item.
- Acceptance: `bash scripts/gate.sh` exits 0, and the workflow's run on the first push is
  green.
- Test: `main_test.go` checks the version output; a deliberate `gofmt` break makes the gate
  exit non-zero, then is reverted.
- Minutes: 30.

### 2. The owned-session helper

- Outcome: one script launches proof sessions, records them, refuses to act on any other, and
  audits the pass's effect on every session it doesn't own.
- Files: `spikes/own.sh`, `spikes/own_test.sh`, `spikes/closing.py`; it writes
  `spikes/owned.tsv`, `spikes/control.log`, `spikes/others-before.tsv`,
  `spikes/others-after.tsv`, and `spikes/audit.txt`.
- Class: `runner`.
- Work:
  - `own.sh start` runs once, before the first launch. It saves `claude agents --json --all` to a
    file and writes the id, state, and `startedAt` of every row to `others-before.tsv`, with the
    time. `own.sh finish` writes the same columns to `others-after.tsv` at close.
  - `own.sh launch <name> <folder> <flags...>` refuses any flag naming `bypassPermissions` or
    `--dangerously-skip-permissions`, changes folder only in a child shell, starts the session,
    finds its row by its unique name in a saved `claude agents --json --all`, and appends the id,
    session id, name, folder, and launch line to `owned.tsv`.
  - `stop`, `respawn`, `rm`, `logs`, `resume`, and `kill` take an id, refuse one not in
    `owned.tsv`, and log each call, with its time, to `control.log`. A copy that `resume` starts
    is recorded from its `note:` line.
  - `kill` saves a fresh session list just before it acts and reads the target's row from it. It
    refuses unless every check holds: `/proc/<pid>/cmdline` starts with `claude bg-spare`; the
    process's grandparent, read from `/proc/<pid>/status` at that moment, runs
    `claude daemon run`, and the target isn't that grandparent; no other row in the list has the
    same process id; and the row's folder is under the pass's worktree. No process id is ever
    written into the script.
  - `daemonpid` prints the supervisor's process id, found the same way from an owned session's
    process, and checks its `cmdline` before it prints.
  - `audit` prints five lines and writes the lists behind them to `audit.txt`:
    - `strangers: <n>`: targets in `control.log` that aren't ids in `owned.tsv`.
    - `foreign kills: <n>`: `bg settled <id> (killed)` lines in `~/.claude/daemon.log` dated
      after the pass's first launch, naming an id outside `owned.tsv`, within 10 seconds after a
      `control.log` entry. Every such line after the first launch is listed, whatever its time.
    - `foreign restarts: <n>`: rows in `others-before.tsv` whose `startedAt` differs in
      `others-after.tsv`. Rows whose state changed are listed too, since other sessions go on
      working, but aren't counted.
    - `bypass: <n>`: matches for `bypassPermissions`, a `defaultMode` of bypass, or
      `dangerously` in every `spikes/**/settings.json`, every hook file those files name, and
      `owned.tsv`.
    - `allow hooks: <files>`: every hook file under `spikes/` that can print an `allow`
      decision.
  - `closing.py` reads the findings report, finds each of the seven `## Proof <n>` sections, and
    checks that each holds exactly one line starting `Held.` or `Fallback applies: `. It prints
    `7 sections, one closing line each` and exits 0, or names each section that fails and exits
    1.
- Acceptance: `own_test.sh` passes, and `own.sh audit` prints `strangers: 0` on an empty log.
- Test: `own_test.sh` uses a stub `claude` on the `PATH` and fake `/proc` and log files. It checks
  that an id absent from `owned.tsv`, a launch carrying a bypass flag, and a `kill` whose target
  is the supervisor, shares another row's process id, or doesn't run `claude bg-spare` are all
  refused with
  no `claude` command run. It also feeds `audit` a fake daemon log and both snapshot files and
  checks each count, and feeds `closing.py` a report with one section that has two closing lines.
- Minutes: 25.

### 3. Proof 5: a new worktree raises no trust prompt

- Outcome: the findings say whether a session launched in a new worktree of the trusted repo
  starts with no trust prompt.
- Files: `spikes/5-trust/` (`run`, notes). It also makes and then removes
  `.claude/worktrees/ccd-spike-5` and its branch in the main checkout.
- Class: `runner`.
- How it is attempted: run after pre-flight item 1 and `own.sh start`. `git -C` on the main
  checkout adds the worktree from `origin/main`. `ccd-spike-5-trust` launches with its folder
  there and is told to reply "ready" and change nothing. The script records the launch output and
  the row, and runs `own.sh daemonpid` while the session is alive, saving the result to
  `spikes/daemon.pid` for proof 1. Then the worktree goes with `git worktree remove` and the
  branch with `git branch -d`, neither forced.
- Held when: the launch prints no "Workspace not trusted" and the row reaches done without a
  `blocked` state.
- Fallback: "Fallback applies: worktrees go in one folder Geoff trusts once, at setup." The
  finding proposes that folder, and proofs 1, 4, and 6 launch from it.
- Acceptance: the findings section for proof 5 holds the launch output and the row.
- Test: `run` reruns the proof.
- Minutes: 20.

### 4. Proof 1: wake a session blocked on a lost connection

- Outcome: the findings say whether a script can bring a session blocked on a lost
  connection back to work.
- Files: `spikes/1-wake/` (`run`, `tunnel.py`, `settings.json`, `prompt.txt`, `wake.sh`, sandbox,
  notes).
- Class: `runner`.
- How it is attempted: first, the proxy check. A session's process descends from the supervisor,
  so a proxy variable in the supervisor's environment would most likely override the settings
  file. The script reads only the variable names in `/proc/<pid>/environ` for the process id in
  `spikes/daemon.pid`, after checking that process still runs `claude daemon run`, and records
  whether any proxy name is present. It never reads a value. A stdlib `python3 -I` CONNECT tunnel listens on 127.0.0.1 and logs only the host
  of each tunnel; it never reads or keeps the traffic. One session, `ccd-spike-1-wake`, launches
  with a settings file whose `env` block points `HTTPS_PROXY` at the tunnel. Its prompt runs six
  steps, each a separate Bash call that runs `sleep 90` and then appends a line to `progress.txt`.
  Once the tunnel log shows `api.anthropic.com` and `progress.txt` has two lines, the tunnel stops.
  The script then saves the session list every 60 seconds for up to 20 minutes, until the
  session's row leaves `working`, and records the row and `claude logs`. Then it restarts the
  tunnel and tries each wake in turn, allowing about 3 minutes for each, so that all three fit in
  the 10 minutes after the link returns. It stops at the first that works:
  1. Nothing: the session retries by itself.
  2. `own.sh respawn <id>`.
  3. `own.sh resume <id>` with the prompt "Carry on with the next step."
- Held when: one of the three brings the row back to `working` with no typed input and no attach,
  within 10 minutes of the link's return, and `progress.txt` reaches six lines with no step
  written twice. The finding names which wake worked and its time after the link's return. A
  session that retries by itself counts as held with no wake needed.
- If none works in its window: the script keeps saving rows until 10 minutes after the link's
  return, and the finding records whether any wake brought the session back later and when. The
  closing line reads "Fallback applies: the check ends the session and relaunches it from the
  files."
- If the state can't be made: when the tunnel never sees API traffic, or the session never
  leaves `working` in 20 minutes, the state can only be observed when a real drop occurs. The
  finding then reads "Fallback applies: the check ends the session and relaunches it from the
  files." The pass files a follow-up that describes how to try the same three wakes on the next
  real drop, and leaves `spikes/1-wake/wake.sh` ready for it.
- Acceptance: the findings section for proof 1 holds the proxy check, the commands, the saved
  rows, the tunnel log, `progress.txt`, and its closing line.
- Test: `run` reruns the proof from a clean sandbox.
- Minutes: 60.

### 5. Proof 4: a pass runs with no permission prompt

Geoff ruled on 2026-11-17, quoted as he wrote it:

<!-- vale Vale.Spelling = NO -->
> make sure that the new system isn't going to run into an avalance of
> permission issues. (That's a common problem. I often need to attach to background sessions to
> give them permission for one thing or another.)
<!-- vale Vale.Spelling = YES -->

This proof is therefore a goal cc-desk is measured on, not only a fallback check. Its target is
zero prompts.

- Outcome: the findings name a launch under which a small pass runs from worktree to merge with
  no prompt and no denial, or list by name each prompt that remains and why no launch setting
  removes it.
- Files: `spikes/4-permissions/` (`run`, `launch.sh`, `settings.json`, `prompt.txt`, `tally.py`,
  `denials.py`, `decide-hook.py`, `answer-settings.json`, the scratch clone, notes).
- Class: `runner`.
- The pass shape: the session launches as cc-desk is to launch a pass, with
  `claude --bg`, a name, the `--worktree` flag, a permission mode, and a settings file. Its
  folder is a clone of `glw907/cc-desk-scratch` (pre-flight check 7) under the proof folder. The
  scratch repo holds a toy Go package, a gate script, and a workflow that runs the gate. The
  prompt runs these steps in order and reports each one's result:
  1. Edit a file and write a new one.
  2. Run the scratch repo's gate.
  3. Commit twice.
  4. Push the branch.
  5. Open a draft PR.
  6. Mark it ready with `gh pr ready <n>`.
  7. Land it with `approval-queue land glw907/cc-desk-scratch#<n> "Scratch proof: <one plain
     sentence>"`, with the Bash call's timeout set to 600,000 ms, since `land` waits for the
     workflow's check. `land` records that sentence in the queue store and puts nothing before
     Geoff.
  8. `EnterWorktree` on an existing path, the step the records show prompting.
- If the landing can't be tested: when `land` is refused twice or the scratch repo couldn't be
  made, the session closes the PR with `gh pr close`, and the finding lists the merge step as
  untested, with the refusal quoted.
- How it is counted: the script saves the session's row every 60 seconds. `tally.py` counts
  each `blocked` state, with its `claude logs` text, and each denial row in
  `~/.local/state/claude-denials.jsonl` whose session id is the proof session's.
- How each one is removed: each prompt or denial is traced to a cause, then to the launch setting
  that removes it: the permission mode, `--allowedTools`, the settings file, or an exact allow
  rule. The proof edits only the spike settings file and launch line, never a Claude settings
  file, and reruns with a fresh scratch branch. It stops when a run counts zero, or after four
  runs, with the residue listed by name. The first run uses `--permission-mode auto` and the
  narrow allow rules the pass needs. A `dontAsk` run is tried only after that, and the finding
  notes that `dontAsk` turns a prompt into a refusal the session sees.
- A prompt for the state proof: a separate session, `ccd-spike-4-prompt`, launches in `default`
  mode with no hook and is asked to run one Bash command outside any allow list. It must block on
  a prompt, and task 6 reads that state.
- Answering a prompt from outside: the finding says whether a blocked prompt can be answered
  without attaching, so the desk could clear one on Geoff's word. The places to look, in order:
  1. `claude --help`, `claude agents --help`, and the agent-view docs, for a command that answers
     a prompt. Version 2.1.326's help names none. Its `--permission-prompts` and
     `--permission-prompt-tool` choose who answers prompts, but the help ties them to `--print`.
     The proof tries whether either takes effect with `--bg`.
  2. The `PermissionRequest` hook in the mechanism table. A third session,
     `ccd-spike-4-answer`, launches in `default` mode with `answer-settings.json`, whose
     `PermissionRequest` hook runs `decide-hook.py` with a timeout of 150 seconds. The hook
     waits up to 120 seconds for a file under `spikes/4-permissions/decisions/` that names the
     tool and command of the request, prints an `allow` decision if one does, and prints a
     `deny` decision in every other case. It logs each request it sees to `hook.log`. The
     session is asked to run one Bash command outside any allow list, and the script writes the
     decision file 30 seconds after `hook.log` shows the request. The finding
     records whether the hook fired in a background session and whether the command ran.
  3. If none works, the finding says plainly that a prompt can only be answered by attaching.

  Never tried: driving `claude attach` through a script, the supervisor's control socket, or
  anything under `~/.claude/daemon/`.
- The denials log by class: `denials.py` reads the log's classes and tools, and the finding
  carries the table under "Denials by class" below with one line per class. For "Classifier
  unavailable", the finding reads only the denials log and `~/.claude/daemon.log`, never another
  session's job files or transcripts. For each session id in those rows, it records what the
  supervisor log shows next for that session: settled done, settled killed, retired idle, or
  nothing. It says whether `cc-desk check` could wake or relaunch such a session without Geoff,
  by proof 1's and proof 6's methods.
- Held when: one run counts zero prompts and zero denials across all eight steps.
- Fallback when none does: "Fallback applies: the launch carries the narrowest allow list that
  works, and a prompt that still appears shows as Needs you now." The finding lists each
  remaining prompt by name, with its cause.
- Never tried: `bypassPermissions` or `--dangerously-skip-permissions`, on any session.
- Acceptance: the findings section for proof 4 shows each run's launch line, its count, each
  prompt's cause and the setting that removed it, the residue if any, the landing's result or
  why it is untested, the answer to "Answering a prompt from outside", and the filled class table.
- Test: `tally.py` and `denials.py` run on a fixture of three denial rows and two saved rows and
  print the expected counts. `decide-hook.py` runs on a fixture request with and without a
  decision file and prints `allow` and `deny`. `run` reruns the last launch against a fresh
  scratch branch.
- Minutes: 200.

#### Denials by class

From the log read on 2026-11-17 (127 rows, 2026-11-01 to 2026-11-17). The proof confirms or
corrects each line.

| Class | Rows | Who hit it in the log | cc-desk session kind that would hit it | What removes it |
| --- | --- | --- | --- | --- |
| Classifier unavailable | 47: 38 Bash, 8 Edit, 1 `ScheduleWakeup` | Any session, during a classifier outage | Every kind | No setting: it's an outage. The proof tests whether an exact allow rule skips the classifier for the pass's routine commands, and records what the supervisor log shows a session did after one |
| Destructive or Unwanted Actions | 13 | Executors stopping test runners they started, `git stash drop`, `claude stop` | Pass executor; Retire round | A reviewed gate script that cleans up the processes it starts, with an exact allow rule. Session stops move to the `cc-desk` command |
| Interfere With Workloads | 11 | `kill` and `pkill` of test runners; conductors running `claude stop` | Pass executor; conductor | The same gate script; session control done by the `cc-desk` command, which isn't a session |
| Self-Modification | 10 | Sessions editing `settings.json` or running `stow` for `claude` | Desk session; a pass that changes how it was launched | Launch settings live in a file the `cc-desk` command writes, and Claude settings change only through `claude-infra` before a launch |
| Server-side classifier, no reason given | 10 | Mixed: a push to main, `claude stop`, a settings edit, a secret read | Every kind | Trace each in the proof. Most match another class in this table and go the same way |
| Undermining Oversight | 4 | `claude stop` from a session; edits near the queue | Conductor; desk session | Session control in the `cc-desk` command |
| Unauthorized Access | 4 | Reads of other sessions' job folders and transcripts; a download | Desk session; executor | The command passes each session the paths it needs, and no session reads another's files |
| Merge Without Review | 3 | `gh pr merge` from a session | Pass executor at a clean merge | `approval-queue land` with an exact allow rule, the path `pass-core` uses until cc-desk runs; proof 4's step 7 tests it |
| Auto-Mode Bypass | 3 | `gh workflow run` | Pass executor | CI runs on push and PR, so no pass dispatches a workflow |
| Classes of 1 or 2 rows (16 classes) | 26 | Mixed | Mostly the Retire round | Listed by name in the finding, each with its session kind |

### 6. Proof 2: tell the states apart

- Outcome: the findings show which of the six states a script can tell apart from what
  Claude Code records.
- Files: `spikes/2-states/` (`run`, `classify.py`, `fixtures.json`, notes).
- Class: `runner`.
- How it is attempted: `classify.py` is stdlib Python, run as `python3 -I`. It takes a saved
  `claude agents --json --all` file and, for owned sessions only, the output of
  `claude logs <id>` and that session's `~/.claude/jobs/<id>/state.json`. It labels each
  session with one of the six states or "can't tell", and names the field each label rests on.
  The live samples are a working session (proof 1 during a sleep step), an idle one (any finished
  proof session), the lost-connection block (proof 1, if it was made), and the permission prompt
  (`ccd-spike-4-prompt`). A lost connection that proof 1 couldn't make, an expired sign-in, and a
  usage limit come from the recorded texts in the mechanism table. `fixtures.json` copies each
  text character for character into the field it was recorded in (`detail` or `needs`),
  including the EAI_AGAIN text with its dash and the `·` in both limit texts, so that the
  classifier is tested on the real strings and not on paraphrases. `loginExpiresAt` from
  `claude auth status --json` is recorded as a way to see a sign-in expiry coming.
- Held when: each live sample gets the right label on two saves 60 seconds apart, and each
  fixture gets the right label. The finding lists the source of each label. Any label that rests
  on `~/.claude/jobs/`, which the docs tell scripts not to read, is marked as resting on an
  undocumented file.
- Partial result: if some states can't be told apart, the closing line reads "Fallback
  applies:", names those states, and says they show as Needs you now with the session's last
  line.
- Acceptance: the findings section for proof 2 holds a table of every sample, its label, and
  its source.
- Test: `classify.py` runs against `fixtures.json` and the saved live rows, and exits non-zero
  on any wrong label.
- Minutes: 40.

### 7. Proof 6: relaunch from the state files

- Outcome: the findings say whether a session that died can carry on from its branch, with
  only one session ever holding the work.
- Files: `spikes/6-relaunch/` (`run`, `prompt.txt`, sandbox, notes).
- Class: `runner`.
- How it is attempted: `ccd-spike-6-relaunch` works on a branch in its sandbox. Its prompt
  makes five commits, one for each step, each after a 60 second sleep, and checks the log before
  each commit so that no step is written twice. After the second commit, `own.sh kill` ends its
  process, which stands in for a restart. The supervisor is never stopped, since that would
  reach every live session, and the reboot record in the mechanism table shows what a restart
  leaves. The script records the row's state after the process ends, and whether anything
  relaunched it. Then it tries, in turn:
  1. `own.sh respawn <id>`.
  2. `own.sh resume <id>` with "Carry on from the last commit."
  3. A fresh `own.sh launch` in the same sandbox and branch. Its prompt is the prompt file plus
     one line: carry on from the branch.
- Held when: the first or second brings the branch to five commits. The first two are kept and
  none is repeated. No save of the session list in that time shows two live rows for the work.
  The finding names which one worked.
- Fallback: if only the third works, "Fallback applies: a fresh session starts from the
  pass's plan and branch."
- Acceptance: the findings section for proof 6 holds the `git log` of the sandbox branch and
  the saved rows.
- Test: `run` reruns the proof from a clean sandbox.
- Minutes: 35.

### 8. Proof 3: dropped tools leave the context

- Outcome: the findings say how many tokens a session saves when it launches without the
  tools it never uses.
- Files: `spikes/3-tools/` (`run`, `tokens.py`, notes).
- Class: `runner`.
- How it is attempted: two sessions start in one sandbox on the same model and effort.
  `ccd-spike-3-full` launches as is. Its prompt asks it to say, for each of the five tools in the
  mechanism table, whether its full definition is in its tool list or it is only named as a
  deferred tool, and then to reply "ready". `ccd-spike-3-lean` launches with `--disallowedTools`
  naming the five tools, and is told to reply with the single word "ready". `tokens.py` reads the
  transcript of each of the two sessions, found by its session id in `owned.tsv`. It reads the
  `usage` fields of the first assistant reply and adds up input, cache creation, and cache read
  tokens. In the full session's transcript it also reads any record that lists deferred tool
  names. It reads no other message text and no other session's transcript.
- Which tools count: the finding first records, for each of the five, whether it was present in
  full in the first request, from the deferred-tool record if there is one and from the session's
  own answer if not. A tool that already loads deferred costs little, so it doesn't count toward
  the saving.
- Held when: the lean session's first request is smaller by at least 2,000 tokens for each tool
  that was present in full. If all five were present, that is the 10,000 token threshold. The
  five tools come to 54,000 characters, about 13,500 tokens at four characters a token, or about
  2,700 a tool, and the threshold leaves room for that guess. If none was present, the proof is
  held without a saving to test, and the finding says the tools already stay out of the context.
  The finding gives both totals and the per-tool record.
- Fallback: "Fallback applies: sessions launch with every tool, at the token cost the
  prototype paid."
- Acceptance: the findings section for proof 3 holds both launch lines, the per-tool record, the
  threshold it applied, and both totals.
- Test: `tokens.py` runs on a fixture line built from the documented `usage` fields and
  prints the expected sum.
- Minutes: 25.

### 9. Proof 7: read the usage left without a model

- Outcome: the findings say where cc-desk can read the remaining usage without calling a model
  itself, or that it can't.
- Files: `spikes/7-usage/` (`run`, `statusline.sh`, `settings.json`, `read.py`, notes).
- Class: `runner`.
- Where to look, in order:
  1. The status line. `ccd-spike-7-usage` launches with a settings file whose `statusLine`
     command is `statusline.sh`, which saves its input to `last.json` in the proof folder. After
     the session's first reply, `read.py` reads `rate_limits.seven_day.used_percentage` and
     `resets_at` from that file. The finding also records whether a background session runs a
     status line at all.
  2. `claude -p "/usage" --model haiku` with JSON output, run once. The finding records whether
     the result reports any input or output tokens or any cost, to see whether the command calls
     a model.
  3. `claude auth status --json`. The probe already shows no usage field, so the finding cites it.
  4. Files under `~/.claude`. The probe found none, so the finding cites it.
- Never tried: the sign-in token, an endpoint the docs don't name, the supervisor's control
  socket, or anything under `~/.claude/daemon/`.
- Held when: the first or second gives the weekly figure and its reset time to a script with
  no model call. The finding states the limit of the source it found: a status line file is only
  as fresh as the last session's last reply.
- Fallback: "Fallback applies: each conductor reads the usage and reports it to the command."
- Acceptance: the findings section for proof 7 holds each look and its output.
- Test: `read.py` runs on a fixture shaped like the documented status line input and prints
  the figure.
- Minutes: 25.

### 10. Findings, follow-ups, and STATUS

- Outcome: one report says what held and which fallbacks apply, and STATUS points at it.
- Files: `docs/research/2026-11-17-prove-findings.md`, `docs/follow-ups/*.md`,
  `docs/STATUS.md`.
- Class: `ordinary`.
- Work: the report opens with a short summary table of the seven results, then gives one
  section for each proof, headed `## Proof <n>` and built up during tasks 3 to 9. Each section
  holds the exact commands, their output trimmed to what shows the result, and its one closing
  line. A follow-up file is written for each untested path the proofs leave, such as proof 1's
  real-drop test. STATUS names the report and the next action: design the State and rules round
  from these findings. The report and STATUS follow `~/.claude/docs/voice/written-for-people.md`.
- Acceptance: `python3 -I spikes/closing.py` on the report exits 0, and `wc -l docs/STATUS.md`
  prints 60 or less.
- Test: the gate's vale run passes on both files.
- Minutes: 30.

### 11. Close

- Outcome: nothing the proofs made is left running, and the PR is reviewed and landed.
- Files: none new.
- Class: `runner`.
- Work: end and remove every owned session through `own.sh`, stop the tunnel, remove any proof
  worktree, run `own.sh finish`, and run `own.sh audit`. Then one whole-branch
  `standards:diff-reviewer` read and the PR body as `pass-core` "Close" sets out. The hand-over
  follows `pass-core` "Until cc-desk runs": with the checks green, the review accepted, and no
  call for Geoff, the executor lands the PR with `approval-queue land glw907/cc-desk#<n>
  "<one plain sentence>"`. A call for Geoff goes under "Open decisions" in `docs/STATUS.md` with
  a recommendation, the PR waits, and the session ends.
- Acceptance: every row of "Acceptance" shows its evidence in the PR.
- Test: the acceptance commands, run at close.
- Minutes: 40.

## Stops and their answers

| Stop | Standing answer |
| --- | --- |
| Auto mode refuses a command | Use the alternative in the pre-flight table, once. After a second refusal, mark the proof "Fallback applies" with the refusal quoted, and go on (standing answer 2) |
| Auto mode reports "Classifier unavailable" | Wait 60 seconds and run it once more. That is not a refusal |
| Proof 5's fallback applies | Launch proofs 1, 4, and 6 from the folder that fallback names, and say so in each finding |
| The main checkout's launch prints "Workspace not trusted" | Pre-flight item 1 isn't done. That step is Geoff's: write the question under "Open decisions" in `docs/STATUS.md` with a recommendation, and end the session |
| A proof session blocks on something its proof didn't plan for | Record the row and `claude logs`, stop it through `own.sh`, and relaunch it once with the prompt amended. A second block becomes part of the finding |
| `own.sh` refuses an id or a `kill` | Never work around it. The id or process id isn't the pass's to act on |
| A session the pass didn't launch changes state during a proof | Ignore it. Its rows are never read into a finding; only `own.sh audit` compares them |
| A proof session pushes to a repo other than the scratch repo, or opens a PR there | End it through `own.sh`, close the PR, delete the branch it pushed, and record it as a finding about isolation |
| The desk or any other session acts on a `ccd-spike-*` session | Record it as interference, and rerun that proof once |
| The sign-in expires during the pass | Signing in is Geoff's step: write the question under "Open decisions" in `docs/STATUS.md` with a recommendation, and end the session |
| Weekly usage reads 85 percent or more, or can't be read, before a proof's first launch | Read it with `claude -p "/usage" --model haiku < /dev/null`, on its "Current week (all models)" line. At 85 or more, or with no figure, launch no more proof sessions, mark each untried proof "Fallback applies" with the reason, and close (standing answer 4) |
| The pass passes 1,060 minutes, twice its estimate | Finish the proof in flight, mark the rest "Fallback applies" as not attempted, and close |
| The gate or CI is red | Fix it inside the pass; a cause that returns gets one structural fix (`pass-core` `references/planner.md` "Stops") |
| A question none of these settles | Ask the advisor (`pass-core` `references/advisor.md`), record the answer under `## Decisions`, and go on |

## Path estimate

The tasks run one after another, so the path is their sum: 30 + 25 + 20 + 60 + 200 + 40 + 35 +
25 + 25 + 30 + 40 = 530 minutes. Proof 4 is the largest task. It can rerun its pass shape up to
four times, and each run waits on the scratch repo's check before it lands. Proof 1 holds the
longest single wait: up to 20 minutes for the blocked state and 10 for the three wakes. A run near
the 1,060 minute ceiling most likely means proof 4 took every rerun.

The Page round's terminal link (see the route) isn't probed here. Its probe registers a URL
handler on the desktop and needs Chromium, which puts it past this pass's files and past 15
planned minutes.
