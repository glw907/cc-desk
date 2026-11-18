# Prove first: findings

Five of the seven proofs held and two fell back. A pass can run from a worktree to a merge with no
permission prompt, a new worktree raises no trust prompt, a script can tell the states apart,
dropping five tools saves about 19,000 tokens per session, and usage can be read without a model.
A session blocked on a lost connection can't be woken by `respawn` or `resume`, and a killed
session can't be relied on to come back as one row, so the check restarts such sessions from the
files. Each section ends with one closing line, and the follow-up
`docs/follow-ups/2026-11-17-rerun-prove-proofs.md` lists what stays open.

A "Fallback applies" below means the design's fallback is the safe assumption for the later rounds.
It doesn't mean the behavior failed in every form: each section says what was tried.

| Proof | Result | What decided it |
| --- | --- | --- |
| 5. A new worktree raises no trust prompt | Held | The session finished in 25 seconds and never showed `blocked` |
| 1. Wake a session blocked on a lost connection | Fallback applies | No wake brought the row back to `working`, for 614 seconds after the link returned |
| 4. A pass runs with no permission prompt | Held | Two runs of the eight-step pass: zero prompts, zero denials |
| 2. Tell the states apart | Held | All 16 cases labeled right. Expired sign-in, usage limit, and EAI_AGAIN rest on an undocumented file |
| 6. Relaunch from the state files | Fallback applies | `respawn` left the session idle, and `resume` made a second row |
| 3. Dropped tools leave the context | Held | The first request was 18,990 tokens smaller |
| 7. Read the usage without a model | Held | `claude -p "/usage"` reports no tokens and no cost |

## Usage and the sessions

The weekly figure read 86 percent at proof 1 and 87 percent at proofs 4, 2, 6, and 3, under the
pass's ceiling of 100 percent (raised from 85 on 2026-11-17). No proof was cut short or marked on the
figure. Every proof session launched and stopped through `bash spikes/own.sh`. Its closing audit
printed `strangers: 0`, `foreign kills: 0`, `foreign restarts: 0`, `bypass: 0`, and one allow hook,
`spikes/4-permissions/decide-hook.py`, the hook proof 4 used.

Three things happened outside the plan, and each is recorded here:

- A session `e329f69b` was live in the main checkout, so `claude-worktree-busy` exited 1. I
  worked only in my own worktree. I read its `claude logs` once before I saw that it wasn't mine,
  discarded the text, and used none of it.
- Auto mode refused two of my actions, both named in proofs 1 and 4: a `sleep 90 && echo` step, and
  a write of an `approval-queue land` allow rule into a spike settings file.
- Two scripts had bugs the proofs found: `own.sh`'s `runs()` (it mishandled a first argument with a
  space) and proof 1's `tunnel.py` (a tuple used where a string belonged). Both are fixed, and
  `own_test.sh` passes.

## Proof 5

Usage before the launch, read with `claude -p "/usage" --model haiku < /dev/null`: 87 percent
for the week (all models), resetting Nov 20 at 11pm Alaska time.

`spikes/5-trust/run` adds a worktree `.claude/worktrees/ccd-spike-5` from `origin/main` and launches
`ccd-spike-5-trust` there through `own.sh`, told to reply "ready" and change nothing. The
launch printed no "Workspace not trusted" error, and the row was `done` 25 seconds later, having
never shown `blocked`:

```text
backgrounded · 7f267061 · ccd-spike-5-trust
{"pid": 2922984, "id": "7f267061", "cwd": ".../.claude/worktrees/ccd-spike-5", "kind": "background", "name": "ccd-spike-5-trust", "status": "idle", "state": "done"}
```

The session's screen ended on "Worked for 3s · done". The worktree and branch were removed
afterwards with `git worktree remove` and `git branch -d`, neither forced. The session finished
before `own.sh daemonpid` could find a live owned process, so proof 1 finds the supervisor from its
own live session.

Held.

## Proof 1

Usage before the launch: 86 percent of the weekly limit (all models).

**Setup.** `spikes/1-wake/tunnel.py` is a CONNECT proxy on 127.0.0.1:18907 that logs hosts only and
refuses every connection while a block file exists. The session launched through `own.sh` with
`HTTPS_PROXY` in the launch settings' `env` block. The supervisor's environment held no proxy
variable name, so the proxy reached the session through the settings file alone. Before the cut
the tunnel carried 14 connections to `api.anthropic.com`, so the session's API traffic did go
through it. The session ran a six-step prompt, each step a 90-second Bash call that appends a
line to `progress.txt`.

**The cut.** At 03:11:59 UTC two steps were done and the block file went in. Step 3 finished in
the next minute, since the Bash call needs no network. The row stayed `working busy` for four
more minutes while the session retried ("attempt n/10"), and at minute 5 it read `blocked idle`.
The screen's last line then read "API Error: Couldn't connect through your proxy
(ERR_PROXY_TUNNEL)". The state was made, and a lost connection takes about five minutes to show
as `blocked`.

**The wakes.** The link came back at 03:17:04.

| Wake | Command | Three minutes later |
| --- | --- | --- |
| Nothing | none | Row `blocked idle`, 3 steps |
| Restart | `claude respawn <id>` | Row `blocked idle`, 3 steps. It printed "restarted 5305f47d". |
| Resume | `claude --resume <sid> --bg` | Original row `blocked idle`, 4 steps. It printed a note that the session was already running, so it started a copy, `b2ddeeb4`. |

No wake brought the row back to `working`. `progress.txt` still reached step 5 by 03:27, so
something kept running the Bash calls after `respawn` and `resume`. I did not read a job file
or transcript to say which process wrote steps 4 and 5. The copy `b2ddeeb4` ran as a second row,
`working busy`, named `step-3-sequential-run`, with its folder at the worktree root, and most
likely wrote them, the same pattern as proof 6. The one result is that the row, the only signal the later rounds can read,
stayed `blocked idle` for 614 seconds after the link returned, whatever ran underneath.

A real drop on the laptop (WiFi lost, lid shut) was not tested. The tunnel models a refused
connection, and a dropped one may behave otherwise. That test is filed as a follow-up.

Fallback applies: the check ends the session and relaunches it from the files.

## Proof 4

This is the goal Geoff set on 2026-11-17: a small pass runs from worktree to merge with no
permission prompt. Two runs of the eight-step pass shape on `glw907/cc-desk-scratch` each counted
zero prompts and zero denials. Usage read 87 percent before each launch.

| Run | Launch line | Prompts | Denials | Result |
| --- | --- | --- | --- | --- |
| 1 | `claude --bg --name ccd-spike-4-main-1 --worktree ccd-spike-4-1 --permission-mode auto --settings spikes/4-permissions/settings.json --model sonnet --effort low` | 0 | 0 | Scratch PR 1 merged at 03:28:32 UTC, 25 seconds after the 03:28:07 launch |
| 2 | The same with `-2` in both names | 0 | 0 | Scratch PR 2 merged; the row read `working` at six polls 10 seconds apart, then `done` |

The settings file holds five allow rules, two exact and three with a wildcard: `git clone` of the scratch repo, `git -C scratch *`,
`bash scratch/scripts/gate.sh`, `gh pr create -R glw907/cc-desk-scratch *`, and
`gh pr ready * -R glw907/cc-desk-scratch`. It has no rule for `approval-queue land`, and step 7
ran anyway: the session reported "returned with no output and no error", and both PRs show
`MERGED`. PR 2 merged at 03:30:25, after its `gate` check passed at 03:30:19. PR 1 merged at
03:28:32, three seconds before its `gate` check started at 03:28:35, so `approval-queue land` merged
it before any check existed. The wildcard rules (`git -C scratch *`, `gh pr create -R ... *`, and
`gh pr ready * -R ...`) go beyond the auto-mode entry's "exact rules, each naming one command", so the
zero-prompt result was measured under wider rules than the entry allows. Auto mode allowed the landing command without a rule, and every
screen line for an edit or a write reads "Allowed by auto mode". Run 1 polled every 60 seconds and
the session finished inside one poll, so run 1 alone would miss a short block. Run 2 polled every
10 seconds and saw none. The tally also reads the denials log by session id, and it found none for
either session. `dontAsk` was not tried: the goal was met with `auto`, so the plan's condition for
trying it did not arise.

**Where the session works.** A nested `git init` folder under the worktree raised "Workspace not
trusted" in proof 1's first attempt, and a clone of the scratch repo is the same case. The pass
shape therefore launches with `--worktree` from a cc-desk worktree, which inherits trust, and the
prompt has the session clone the scratch repo into a plain `scratch/` folder inside it and work
there with `git -C scratch`. The shell's directory moved into `scratch/` after step 1 in run 2, and
the session corrected for it. A later round that runs a pass in another repo should launch from
a worktree of that repo, the case proof 5 held.

**Answering a prompt from outside.** A prompt can be answered without attaching, through a
`PermissionRequest` hook in the launch settings.

- `claude --help` names no command that answers a prompt. `--permission-prompts none` is tied to
  `--print`: a `--bg` session launched with it in `default` mode still blocked on the prompt (the
  run's output was not saved, so this is recorded from the session, not from a file).
- `ccd-spike-4-prompt` ran `date -u +%s > probe.txt` in `default` mode with no hook. Its row read
  `state: blocked`, `status: waiting`, `waitingFor: "permission prompt"` 11 seconds after launch,
  and `claude logs` showed the command and "Do you want to proceed?". That is the sample proof 2
  uses.
- `ccd-spike-4-answer` ran the same command with `answer-settings.json`, whose hook runs
  `decide-hook.py`. The hook fired in the background session and logged the request to `hook.log`.
  The row stayed `blocked waiting` while it waited. Thirty seconds after the log line a script wrote
  `decisions/<sha1>.txt`, and the next poll read `done idle` with `probe.txt` written. The screen
  said "Allowed by PermissionRequest hook".

So the desk can clear one prompt on Geoff's word by writing a decision file the hook looks for,
provided the session launched with that hook. A session launched without it can be answered only by
attaching. The hook waits up to 120 seconds, then denies, so the desk has two minutes to write the
file. I did not drive `claude attach`, the control socket, or anything under `~/.claude/daemon/`.

**Denials by class.** The log held 129 rows on 2026-11-18, 2 more than the plan read. The two are
this pass's: proof 1's second session had `sleep 90 && echo step1 >> progress.txt` refused as
Destructive or Unwanted Actions (the later runs use a `step.sh` script instead), and my write of
proof 4's settings file with an `approval-queue land` allow rule was refused as Self-Modification.
`denials.py` gives the same classes as the plan's table, and the lines below confirm or correct each.

| Class | Rows | Session kind that hits it | What removes it |
| --- | --- | --- | --- |
| Classifier unavailable | 47 | Every kind, in an outage | No setting. 13 sessions: 8 later settled `killed`, 3 `done`, 1 `claimed-spare`, 1 not named in `daemon.log`. An exact allow rule did not come into play: the runs above hit no outage, so whether a rule skips the classifier is untested. A bounded retry after 60 seconds is the plan's rule |
| No explanation from the server-side classifier | 10 | Every kind | Mostly matched another class on the target; the answer is the same rule as the matching class |
| Destructive or Unwanted Actions | 14 | Pass executor, Retire round | A gate script that cleans up what it starts, with an exact allow rule |
| Self-Modification | 11 | Desk session; a pass that changes its launch | Launch settings in a file the `cc-desk` command writes. Confirmed this pass: a write of an allow rule for `approval-queue land` into a spike settings file was refused with this class, and the same file without that rule went through |
| Interfere With Workloads | 11 | Pass executor, conductor | The gate script, and session control in the `cc-desk` command |
| Undermining Oversight | 4 | Conductor, desk session | Session control in the `cc-desk` command |
| Unauthorized Access | 4 | Desk session, executor | Each session gets the paths it needs and reads no other session's files |
| Merge Without Review | 3 | Pass executor at a merge | `approval-queue land`. Step 7 ran it with no rule and the PRs merged |
| Auto-Mode Bypass | 3 | Pass executor | CI runs on push and PR, so no pass dispatches a workflow |
| 16 classes of one or two rows | 22 | Mostly the Retire round | Listed in `denials.py`'s output, each with its session kind |

A `cc-desk check` could not wake or relaunch a `Classifier unavailable` session without Geoff by
proof 1's method: the wakes there did not bring a blocked row back. Relaunching from the files is
proof 6's question.

Held.

## Proof 2

`spikes/2-states/classify.py` (stdlib, `python3 -I`) labels a case from its list row, the
`claude logs` screen, and `state.json`, and `spikes/2-states/run` made the live samples. Usage read
87 percent before the run. All 16 cases in `fixtures.json` get the expected label.

| Sample | Source of the label | Label | Saves |
| --- | --- | --- | --- |
| Working (proof 1, during a sleep step) | `row.state` is `working` | working | 2, 60 seconds apart |
| Finished session (this proof) | `row.state` `done` with `row.status` `idle` | idle | 2, 60 seconds apart |
| Permission prompt (this proof) | `row.waitingFor` is `permission prompt`, with `state` `blocked` and `status` `waiting` | permission prompt | 2, 60 seconds apart |
| Lost connection (proof 1) | The row reads `blocked idle` and carries no reason. The `claude logs` screen holds "API Error: Couldn't connect through your proxy (ERR_PROXY_TUNNEL)" | lost connection | 2, 10 minutes apart |
| EAI_AGAIN, 2026-11-14 (recorded text) | `state.json` `detail`: `API Error: Can't reach the API server — check your internet or DNS (EAI_AGAIN)` | lost connection | fixture |
| Expired sign-in (recorded text) | `state.json` `detail`: "Login expired" | expired sign-in | fixture |
| Weekly limit and Fable limit (recorded texts) | `state.json` `detail`: "You've hit your ... limit · resets ..." | usage limit | 2 fixtures |

Four things this showed:

- The list row alone separates three states. A working session reads `working`, a finished one
  `done idle`, and a permission prompt carries `waitingFor: "permission prompt"`. The prompt row
  also reads `blocked waiting`, where the lost connection reads `blocked idle`.
- A lost connection, an expired sign-in, and a usage limit are all `blocked` rows with no reason in
  the row. The lost connection's reason is on the `claude logs` screen, a documented command. The
  screen held the error line ten minutes after the link came back, but a 700-character tail of it
  did not, so the check must search the whole screen it saves, not a cut of it. The `—` in the EAI_AGAIN text
  and the `·` in the limit texts are in `fixtures.json` as recorded.
- The expired sign-in and the two usage limits come only from recorded `state.json` text. No live
  sample of either was made (no sign-in was expired and no limit was hit), so whether
  `claude logs` shows them is untested. Until it is, those two labels rest on a file the docs tell
  scripts not to read. If the check can't read it, they show as Needs you now with the session's
  last line, the design's fallback for a state it can't tell.
- `loginExpiresAt` from `claude auth status --json` was not read in this proof, so the sign-in
  expiry warning is open.

Held.

## Proof 6

Usage read 87 percent before the run. `spikes/6-relaunch/run` makes a linked worktree of this repo
as the sandbox (branch `ccd-spike-6`, since a nested `git init` is a separate trust root), launches
`ccd-spike-6-relaunch` on a five-step prompt (each step sleeps 60 seconds, then commits, and the
prompt has the session check the log first), and runs `own.sh kill` after the second commit.

| Time | What happened | Commits | Row of the original |
| --- | --- | --- | --- |
| 03:40:09 | Before the kill | 2 | `working busy`, PID 2956315 |
| 03:40:10 | `own.sh kill` ended the process | 2 | |
| 03:40:20 | Ten seconds later | 2 | `working`, no PID, no status |
| 03:40:30 | Twenty seconds later, with no command from me | 2 | `working busy`, new PID 2961551 |
| 03:41:24 | Step 3 committed | 3 | same |
| 03:42:12 | End of the two-minute watch for a wake | 3 | same |
| 03:42:13 | `own.sh respawn` printed "restarted 96fdc1a7" | 3 | `blocked idle`, PID 2959603, for 5.5 minutes |
| 03:47:52 | `own.sh resume` with "Carry on from the last commit." | 3 | note: the session "is already running in the background, so this started a copy as e3b28ac7" |
| 03:49:01 and 03:50:04 | Steps 4 and 5 committed by the copy | 5 | original still `blocked idle` |

The branch at the end, with no step written twice:

```
5c66eb9 Step 5
872d0da Step 4
61b2ff5 Step 3
7a890cc Step 2
4ac0399 Step 1
```

Four things this showed:

- Ending the process did not leave the session dead. Within 20 seconds the row had a new PID and read
  `working busy`. Whether the supervisor relaunched it or the step in flight kept the row busy, the
  saved rows can't say. Step 3 did commit about 74 seconds later, 49 seconds before `respawn`, which is less than one
  step, so the saves can't show a stall after it.
- `respawn` printed success and left the session `blocked idle` with an empty prompt. It did not
  carry on, and nothing in the row says why.
- `resume` with a prompt did the work. The copy ran steps 4 and 5 and read the log first, so it
  skipped steps 1 to 3.
- The copy sat in a different folder (the worktree root) and was named `missing-step-one-resume`.
  The run's saves list rows by folder, so they show one row throughout. The full list saved at the
  end of the run (`out/list.json`) shows two: the original, idle, and the copy, working. The condition "no save shows two live
  rows for the work" is therefore not met, though only one row was ever working. The third
  attempt, a fresh launch, was not needed and was not tried.

Fallback applies: a fresh session starts from the pass's plan and branch.

## Proof 3

Usage read 87 percent before the run. `spikes/3-tools/run` launched two sessions in one sandbox on
Sonnet at low effort, and `tokens.py` read only those two transcripts, found by session id in
`owned.tsv`.

Launch lines:

```
full: claude --bg --name ccd-spike-3-full --model sonnet --effort low --permission-mode default "<asks about the five tools>"
lean: claude --bg --name ccd-spike-3-lean --model sonnet --effort low --permission-mode default "Reply with the single word ""ready"". Run no tool." --disallowedTools Artifact SendFeedback Workflow AskUserQuestion ReportFindings
```

Per-tool record. No transcript held a deferred-tools record that named any of the five. The full
session's answer was "full definition is in my tool list" for each of Artifact, SendFeedback,
Workflow, AskUserQuestion, and ReportFindings, so all five were present in full.

| First request | Input | Cache creation | Cache read | Total |
| --- | --- | --- | --- | --- |
| Full | 2 | 24,073 | 30,107 | 54,182 |
| Lean | 2 | 35,190 | 0 | 35,192 |

The lean request is 18,990 tokens smaller, about 3,800 for each of the five tools. The threshold was
2,000 tokens for each tool present, which is 10,000 for five. The split between cache creation and
cache read says only how warm the cache was, so the finding rests on the total.

Two things to know:

- The first lean launch put the prompt after `--disallowedTools`, which takes every word that
  follows, so the prompt became a sixth tool name and the session sat at an empty prompt. The
  prompt now goes first. Any launcher built from this must do the same.
- The saving is measured on the first request only. Later requests were not compared.

Held.

## Proof 7

The second place the plan lists, `claude -p "/usage" --model haiku`, answers the question. Run with
`--output-format json`, it returned a result with `num_turns` 0, `total_cost_usd` 0, `duration_api_ms`
0, every token count 0, and `local_command` set to `usage`, which marks it as answered by the
client with no model turn. The figure sits in its `result` text. `spikes/7-usage/run` reruns it and
`spikes/7-usage/read.py` parses it:

```text
weekly used: 87%  resets: Nov 20, 11pm (America/Anchorage)  model tokens: 0  cost: 0
```

The limits of this source: the reset time is local text in the `Anchorage` zone, so a reader parses
it as text, and the output is meant for people, so a wording change can break the parser. The first
look, the status line in a background session, wasn't run. The third and fourth looks stay as the
plan cites them: `claude auth status --json` shows no usage field, and no usage file exists
under `~/.claude`.

Held.
