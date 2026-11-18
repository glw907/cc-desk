# Prove first: findings

The pass built the repo's scaffold and gate, and the helper that every proof session is meant to
launch through. It then read the week's usage and found it at 87 percent, above the plan's limit of
85. The plan's stop rule says to launch no more proof sessions, so six of the seven proofs weren't
attempted. Proof 7 needs no session for its second look, and that look held. Each unattempted
proof closes with the design's fallback, and the follow-up
`docs/follow-ups/2026-11-17-rerun-prove-proofs.md` says how to rerun them after the allowance resets
on Nov 20 at 11pm Alaska time.

A "Fallback applies" below means the design's fallback is the safe assumption for the later
rounds, not that the behavior was tested and failed.

| Proof | Result | Why |
| --- | --- | --- |
| 5. A new worktree raises no trust prompt | Fallback applies, not attempted | Usage at 87 percent |
| 1. Wake a session blocked on a lost connection | Fallback applies, not attempted | Usage at 87 percent |
| 4. A pass runs with no permission prompt | Fallback applies, not attempted | Usage at 87 percent |
| 2. Tell the states apart | Fallback applies, not attempted | Usage at 87 percent |
| 6. Relaunch from the state files | Fallback applies, not attempted | Usage at 87 percent |
| 3. Dropped tools leave the context | Fallback applies, not attempted | Usage at 87 percent |
| 7. Read the usage without a model | Held | `claude -p "/usage"` reports no tokens and no cost |

## The stop

The plan's "Stops and their answers" says to read usage before a proof's first launch with
`claude -p "/usage" --model haiku < /dev/null`, on its "Current week (all models)" line, and at 85
or more to launch no proof sessions and close. The read, on 2026-11-17:

```text
Current week (all models): 87% used · resets Nov 20, 11pm (America/Anchorage)
Current week (Fable): 81% used · resets Nov 20, 11pm (America/Anchorage)
```

The pass launched no proof session. `bash spikes/own.sh start`, `finish`, and `audit` ran with an
empty owned list, and the audit printed `strangers: 0`, `foreign kills: 0`, `foreign restarts: 0`,
`bypass: 0`, and `allow hooks: none`.

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

No session was launched, so no connection was cut. The plan's mechanism table shows no documented
way to message or wake a blocked background session, and `claude respawn` and
`claude --resume <id> --bg` are the two candidates the rerun tries. `daemon.log` shows no
relaunch after a reboot.

Fallback applies: the check ends the session and relaunches it from the files.

## Proof 4

This is the goal Geoff set on 2026-11-17, and no run of the pass shape was made. The scratch repo
`glw907/cc-desk-scratch` and its gate workflow exist, so the rerun can start at once. The prompt
count, the denial table, and whether a prompt can be answered from outside the session are all
open. Until a rerun measures them, the later rounds assume the narrowest allow list and show any
remaining prompt as Needs you now.

Fallback applies: the launch carries the narrowest allow list that works, and a prompt that still appears shows as Needs you now.

## Proof 2

Not attempted. Its live samples come from proofs 1 and 4, and the plan's source of the six states is
`claude agents --json --all`, whose `status` field held only `busy`, `idle`, or nothing. That field
alone may not separate a lost connection from a permission prompt. The rerun tests whether
`claude logs` does, before it falls back on the job files the docs tell scripts not to read.

Fallback applies: a state it can't tell shows as Needs you now with the session's last line.

## Proof 6

Not attempted, so no owned process was ended and no relaunch was tried.

Fallback applies: a fresh session starts from the pass's plan and branch.

## Proof 3

Not attempted. The plan's mechanism table gives the five unused tools at 54,000 characters, but no
launch with `--disallowedTools` was measured.

Fallback applies: sessions launch with every tool, at the token cost the prototype paid.

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
