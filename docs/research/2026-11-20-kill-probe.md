# A killed pass comes back, and `session end` gives a counted relaunch

On 2026-11-20 a lone `kill` of a pass's process did not end the pass. Claude Code's supervisor
brought the session back under a new process within 11 seconds, and the check read it as live.
`cc-desk session end` did end it, and the next check counted a `missing` relaunch. Pass 6
(`desk-run`) ran the probe on the branch build installed by `timer install`, on Claude Code
2.1.329, in the cc-desk-scratch checkout. This settles standing answer 70: every pass death in
Tasks 7, 8, and 11 goes through `session end`.

## The probe pass

The probe release was `cc-desk-scratch-probe-20261120t094250z`, from
`scripts/live/probe/release.json`, on scratch branch `probe-base-20261120t094250z`. Its one pass,
`probe-work`, ran `scripts/live/recover/work-plan.md`. That fixture's 300 steps each run the
scratch gate.

The 09:44:41 check launched session `fb7528e0` at 09:44:48, with the decisions `start` `free` and
`launch`. The 09:49:41 check saw its row `working` and set `last_seen` to 09:49:43, later than
`launched`. By 09:54:58 the pass had done 170 steps, so one step takes about 3.5 seconds and 300
steps take about 18 minutes.

## The kill

`kill_target` in `scripts/live/lib.sh` printed PID 282629 from a list saved at 09:50:28, and the
executor ran `kill 282629` alone at 09:50:37. The list, saved every 5 seconds, read:

| Read (UTC) | Row state | Status | PID | Session id |
| --- | --- | --- | --- | --- |
| 09:50:37 | `working` | none | none | `fb7528e0` |
| 09:50:43 | `working` | none | none | `fb7528e0` |
| 09:50:48 to 09:54:12 | `working` | `busy` | 476489 | `fb7528e0` |

The revived row kept its name, `cc-desk-scratch-probe-work`, its session id, and its `cwd` under
the pass's worktree. The 09:54:41 check logged decision `none` for `pass:probe-work`, so it read
the row as a live session. That is the "comes back under a new PID" branch of answer 70.

The first `kill_target` refused PID 282629 because its command line held `--bg-spare`. Every
claimed background session runs as `claude bg-spare` under a `claude bg-pty-host`, real
sessions included, so the plan's guard could never pass. The guard now refuses a terminal host
(`--bg-pty-host`) and a process whose own working folder is outside the scratch checkout. It
keeps the name, row `cwd`, and `comm` checks. Run on the executor's own row, it still refused on the name.

## `session end`

At 09:54:58, after the 09:54:41 check had seen the session, `cc-desk session end pass
cc-desk-scratch probe-work` printed `stopped session fb7528e0`. The list at 09:55:11 and each
minute through 10:00:12 held no probe row, and PID 476489 was gone. The 09:59:41 check logged
`relaunch`, reason `missing`, counted, outcome `launched 1d3558a4`, at 10:00:21.

## What a death costs

The kill came back on its own, so it costs nothing and proves nothing. A `session end` death costs
up to one check interval, 5 minutes, until the counted relaunch. Before the next death the check
must see the new session, which costs up to another 5 minutes. A death cycle is up to 10 minutes.
A relaunched pass resumes at its next step, so the recovery fixture's step count must cover every
launch's 15 minutes of work: about 260 steps a launch at 3.5 seconds a step.

## Teardown

`session end` stopped `1d3558a4`, and the list then held no `cc-desk-scratch-` row. `cc-desk
release done cc-desk-scratch` moved the release to
`releases/done/cc-desk-scratch-probe-20261120t094250z-20261120T100044Z.json`. The scratch checkout
is back on `main`. The snapshot reads 0 needs, 0 waits, 0 waiting, 0 sessions, and 0 unread.
