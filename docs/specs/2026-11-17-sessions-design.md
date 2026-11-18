# cc-desk Sessions: the round design

This is the third round of the cc-desk route. After it, the `cc-desk` command launches every
background session, records who holds which worktree, and ends each session when its work is
done. A timer runs `cc-desk check` every five minutes with no model, and the check recovers a
session that died, stalled, or stopped. A clean pull request merges with a record.

The main design is [2026-11-17-cc-desk-design.md](2026-11-17-cc-desk-design.md), and this round
builds its section "Sessions run themselves". The state it builds on is in
[2026-11-17-state-and-rules-design.md](2026-11-17-state-and-rules-design.md). Where this file is
silent, those two hold.

## Why this round

Geoff's ruling on the route, 2026-11-17: no avalanche of permission prompts. The prototype lost
410 minutes to one lost connection and 3.5 hours to one restart, because nothing watched the
sessions. The State and rules round can say a pass is free to start, and nothing acts on that
yet. This round makes the record launch something and keeps it running.

## Goals

- A pass that is free starts with no person and no session asking for it.
- A session that dies, stalls, or is cut off is running again within 10 minutes, at most twice.
- Anything the check can't fix shows as "Needs you now", with the round's name and the one
  thing to do.
- A pass launched by cc-desk runs from its worktree to its merge with no permission prompt.
  Every prompt a session does meet is recorded with its cause.
- A clean pull request merges with a record, after its checks have registered and passed.

## What it leaves out

- The page. This round's results are read from the snapshot. The Page round draws them.
- The desk and the two skills. Sessions here launch from prompt templates that name `pass-core`.
  The Desk and skills round rewrites the templates to name the working skill.
- The conductor's judgment. The check starts free work by the rules already built. A conductor
  session is launched, watched, and relaunched here, and its tests use a stub.
- Answering a permission prompt while the session waits at it. See "Permission prompts".
- Moving cairn and dubplate. Every live test runs in `glw907/cc-desk-scratch`.

## What would make it a failure

- Two sessions hold one pass, or one worktree.
- The check relaunches without limit, or ends a session it never launched.
- A session runs with permission checks turned off.
- A pull request merges before a check has registered, or with a check failed.
- The check uses a model, or stays running between timer runs.

## Approaches

| Approach | What it trades |
| --- | --- |
| **A timer runs a check that holds nothing in memory.** Each run reads the session list and the state files, acts, and exits. Picked. | A fault is seen up to five minutes late. In return, a restart or a crash of the check loses nothing, and every decision can be tested from a fixture of two files. |
| A long-running supervisor keeps the sessions as children and reacts at once. | Faster. It is a standing process that needs its own recovery, and a restart ends it with everything it knew. That is the desk's failure again. |
| Each session reports on itself through hooks, and nothing watches from outside. | No timer. A dead or cut-off session can't report, and those are the cases that cost the prototype its hours. |

One piece of the third approach is kept. A `PermissionRequest` hook in each session's launch
settings records the prompt the moment it appears, with the tool and the command. The check
alone could see only that a prompt is open.

## What the command records

The State and rules round left session and pull request fields to this round. They go in the
release file, under the thing the session works on: a pass for a pass session, a round for a
design writer, the release for its conductor. Field names are the builder's. The record holds:

- the session's name, its kind, its model and effort, and the short id and session id that
  `claude --bg` gave it;
- the worktree's path and its branch, and the time of the launch;
- the count of relaunches, each with its time and cause;
- the last time the check saw the session working;
- each permission prompt: the time, the tool, the command or path, the rule Claude Code
  suggested, and Geoff's answer once given;
- for a pass, its pull request: the number, the last commit the session reported, the
  reviewer's verdict, the count of fix attempts, and the time of the merge;
- for a conductor or a design writer, whether it has reported its work finished.

A release also records where its repo is checked out on this machine and the repo's default
branch. `Release.Repo` holds only the GitHub name today. A verb sets both, and a start is
refused for a release that has neither.

A worktree belongs to at most one live record. The strict reader refuses a state where two
records name one worktree or one session name. That record is what replaces
`claude-worktree-busy` for work cc-desk launched.

The snapshot gains what the page needs from this: each session with its state as the check
last saw it, each prompt, what the check did in its last run, and each "Needs you now" entry
with its round, its cause, and the one thing to do.

## Launching

The command makes the worktree, then launches. A session never makes its own.

1. `git worktree add <repo>/.claude/worktrees/<pass> -b <pass> origin/<default>` after a fetch.
   A relaunch reuses the worktree and the branch as they are.
2. The command writes the session's settings file in the state folder, fresh at every launch
   and relaunch. It holds the project's allow rules, the `PermissionRequest` hook, and deny
   rules for `gh pr merge` and `approval-queue land`. A session can't write this file: the
   proof found that Claude Code refuses a session that writes allow rules, and the file sits
   outside every worktree.
3. From the worktree, the command runs
   `claude --bg --name <project>-<pass> --model <m> --effort <e> --permission-mode auto --settings <file> "<prompt>" --disallowedTools <tools>`.
   The prompt comes before `--disallowedTools`, which takes every word after it.
4. It records the ids the launch prints, and the start of the pass if this is the first launch.

The session name is `<project>-<pass>`, and a relaunch keeps it. A conductor is
`<project>-conductor`, and a design writer is `<project>-<round>-design`.

**What starts a session.** The check launches a pass session when the settling step says the
pass is free. In this round a conductor or a design writer starts only when a desk session
runs the start verb for it, and it ends when it reports its work finished through the command.
The check never relaunches a record that reads finished. When those two kinds start by
themselves is the Desk and skills round's to settle.

**Models.** A pass session is Opus at high, because it plans the pass before it builds. A
conductor is Sonnet at medium. A design writer is Opus at high. The release file can name
another model for a pass.

**The prompt.** The command fills a template kept in the repo, one per kind. It names the
session, the design's path, the worktree, the branch, and the verbs the session reports with.
A relaunch adds one paragraph: the cause, the count, and "carry on from the branch as it
stands; the commits on it are yours".

**Allow rules.** Each project's rules are a list in its release file. They can arrive in the
input file when the release is added. After that only a desk session can change them, under
the same guard as a ruling, so a background session can't widen what it may do. The rules are exact commands where the proof showed an exact rule is enough.

**Tools left out.** Each kind launches without the tools it never uses. The proof measured about
19,000 tokens saved per request. The list per kind is the builder's, checked by a launch that
still finishes its pass.

**The scope.** Every launch goes through `systemd-run --user --scope`. Claude Code's session
supervisor lives in the control group of whatever first started it, and systemd ends a timer
service's whole group when the service exits. A session launched straight from the check would
die with it.

## The check

`cc-desk check` reads three things: the state files, `claude agents --json --all`, and, for a
blocked session whose row gives no reason, that session's `claude logs`. Then it acts on each
record. It touches only sessions the state files name.

| What it finds | How it knows | What it does |
| --- | --- | --- |
| A session working | The row reads `working` | Records the time |
| A session that died, or all of them after a restart | No live row at two reads 30 seconds apart, and the work isn't finished | Relaunches in the same worktree |
| A session cut off from the network | `blocked`, and the screen holds an API connection error | Ends it and relaunches, once the API can be reached |
| A session idle 15 minutes with its work unfinished | The row reads `done` or its status reads `idle`, the last working time is 15 minutes old, and no ready report stands | Ends it and relaunches |
| A launch that never started | No row with the recorded id one check after the launch | Launches again, once. A second failure is "Needs you now" |
| A permission prompt | `waitingFor` is `permission prompt` | "Needs you now": the prompt's command, and "answer it at the desk" |
| An expired sign-in | `blocked`, and the screen or the session's state file reads "Login expired" | "Needs you now": "sign in to Claude Code" |
| A usage limit | `blocked`, and the text reads "You've hit your ... limit" | "Needs you now", with the reset text. It relaunches once the usage read is under the stop line |
| A blocked session it can't read | None of the above | "Needs you now", with the session's last line |
| A third stop | Two relaunches are recorded and the session is down again | "Needs you now". No third relaunch |
| A pull request reported ready | The pass's record | See "Merging" |
| A merged pass | The merge is recorded | Ends the session, removes the worktree and the branch |
| Work that became free | The settling step of the State and rules round | Starts it, by launching |
| A conductor or design writer that reported finished | Its record | Ends the session. Never relaunches it |

**A live row.** A row is live when it carries the recorded id and its state is `working` or
`blocked`. A row that reads `done`, `failed`, or `stopped`, and a missing row, are not live.
Claude Code's supervisor brings a killed process back by itself within about 20 seconds, as
proof 6 showed, so the check reads the list a second time 30 seconds later before it calls a
session dead. A session counts as ended when its row is missing or reads `stopped`, `done`,
or `failed`.

**A ready report holds the idle rule off.** A pass session that has reported its pull request
ready sits idle while the checks run, and that is not a stall. The idle rule applies only
while no ready report stands. A fix relaunch clears the report, so the rule applies again.

**Limits on a relaunch.** No relaunch happens while the usage read is at the stop line or
can't be read. The session waits, and the wait isn't counted. A relaunched pass keeps the
slot it already held under the two-pass limit, so a restart never puts a third pass in flight.

A relaunch never wakes the old session. The proof showed `respawn` and `resume` can't wake a
blocked session, and `resume` leaves a second row. So the check always ends the old session
first, confirms its row is gone or stopped, and only then launches. That order is what keeps
two sessions off one pass.

The network case waits for the link. Before it relaunches a cut-off session the check sends
one HTTPS request to `api.anthropic.com`, and any HTTP answer counts as the link being up. The
usage read can't serve here: proof 7 showed Claude Code answers it locally. No answer leaves
the session as it is until the next run, and that wait doesn't count as a relaunch. The
relaunch itself counts toward the two.

Each "Needs you now" cause is a stop reason. A session's usage limit is the `usage` reason the
release file already has. Five are added to the fixed list: a permission prompt, an expired
sign-in, an unreadable stop, the relaunch limit, and a failed launch. For a pass, the check
stops it through the stop the last round built, so its stop interval and its actual minutes
follow the rules already there. A relaunch after the cause clears is that round's resume. A
conductor or design writer has no pass state, and its stop sits on its session record. All
six causes need Geoff.

The check works in two steps. It first decides one action for each record, then carries the
actions out. A flag makes it print the decisions and carry out none, which is how the tests
read a decision without a launch.

Two checks never run at once. The second takes the state lock's 10-second wait and exits.

## Permission prompts

Auto mode with a project's allow rules is meant to leave no prompt. A prompt is a defect in
the rules. When one appears:

1. The hook runs `cc-desk` with the prompt on its input. The command records it and exits with
   no decision, so Claude Code shows its normal prompt and the session waits.
2. The next check sees the waiting session and the snapshot reads "Needs you now".
3. Geoff answers at the desk, yes or no. A yes adds the exact rule to the project's list. The
   command ends the session and relaunches it from its branch. A no relaunches it too, with a
   line in its prompt naming the refused command. Neither counts toward the two relaunches.
4. The command files the prompt as a follow-up, as below.

Until the Desk and skills round, the answer is a verb Geoff's terminal session runs, under the
ruling guard.

No session runs with permission checks off (the main design's call, kept).

## Merging

A pass session ends its work by reporting its pull request ready: the number, the last commit,
its reviewer's verdict, and whether the change stayed inside the approved design. Only the
check merges. No session runs a merge command, so the merge can't depend on a session's
permissions. Proof 4's session merged its own pull request with `approval-queue land`, so each
launch's settings deny that command and `gh pr merge`. A pull request the check finds merged
by another hand is recorded as merged outside cc-desk, cleaned up the same way, and counted
in the round report.

The check merges when all of these hold:

- the report says the reviewer accepted and the change is inside the design;
- the pull request's last commit is the one reported;
- at least one check has registered on that commit, and every registered check has passed.

It merges with a plain merge commit, pinned to that commit. Then it records the merge,
finishes the pass at the merge time, pulls the main checkout when that is safe, ends the
session, and removes the worktree and the branch with no force flag. A worktree with
uncommitted changes stays, and the snapshot says so.

| Case | What the check does |
| --- | --- |
| No check has registered yet | Waits for the next run. After 30 minutes with none, "Needs you now" |
| A check is still running | Waits |
| A check failed, or the reviewer asked for a fix | Relaunches the pass session with the failures named, twice at most. The relaunch clears the ready report, and the check decides nothing more for that pull request until a new report arrives, so one failed commit spends one attempt |
| A third failure | A question under "Waiting for you", naming the pull request and the failures |
| The report says the change reaches outside the design | The same question, at once |
| The last commit moved after the report | Waits for a new report |

## Follow-ups through the command

A session files a follow-up with a verb: a title, a body, and the project. The command writes
one file in that project's `docs/follow-ups/`, in the session's worktree, named from the
title. It refuses a name that exists. The file rides the pass's pull request. The command also
uses this for each permission prompt, written into the cc-desk repo's folder.

## The timer

Two unit files live in the repo: `cc-desk-check.service`, a one-shot that runs `cc-desk check`,
and `cc-desk-check.timer`, with `OnStartupSec=1min` and `OnUnitActiveSec=5min`. The user's
systemd starts at boot here, since lingering is on, so "at login" is at boot. The service sets
`PATH`, because a user unit's default has neither `claude` nor `gh`. It also sets
`KillMode=process`. The scope keeps a launched session alive, and this line covers a
supervisor that Claude Code starts in the service's group all the same.

One case can't be tested on this workstation: a boot where the timer's check is the first
thing to start Claude Code's supervisor. Testing it means ending the supervisor, which ends
every session on the machine. The round files it as a follow-up for the Move round's first
restart.

`cc-desk timer install` copies the two files to `~/.config/systemd/user/` and enables the
timer. `cc-desk timer status` prints the last run and its result. The Move round records the
install in the workstation repo.

A verb that makes work for the check (a ready report, a ruling that frees a pass) also starts
the service, without waiting for it. The five minutes is then the longest wait, not the usual.

## When something goes wrong

| What happens | What the command does |
| --- | --- |
| `claude agents --json` fails or isn't JSON | The check exits with the error and changes nothing. Three failed runs in a row read "Needs you now" |
| `git worktree add` fails | The pass stops as a failed launch. Nothing is left half-made |
| The worktree exists and no record names it | Refused. The command never takes over a folder it didn't make |
| `claude stop` fails for a session being replaced | No relaunch this run. The next run tries again |
| GitHub can't be reached | Merging waits. Nothing else is held |
| A state file fails to parse | As the State and rules design says: exit, write nothing |

| Limit | Value |
| --- | --- |
| Check interval | 5 minutes, and 1 minute after boot |
| Wait between the two reads that call a session dead | 30 seconds |
| Relaunches per pass or conductor, any cause | 2 |
| Idle before a session counts as stalled | 15 minutes |
| Tries for a launch that never started | 2 |
| Fix attempts on a pull request | 2 |
| Wait for a first check to register | 30 minutes |
| Failed session-list reads before "Needs you now" | 3 |

## The passes

The round is five passes in one chain. They all add verbs to one command and fields to one
release file, so none runs beside another. The route listed four; the prompts and the
follow-up verb get a pass to themselves, and the timer moves last because the live run needs
everything before it.

| Pass | Outcome | Done items | Minutes |
| --- | --- | --- | --- |
| 1. The check | The command records sessions and pull requests in the release file, reads the session list, and decides and records what each session needs | 3 | 90 |
| 2. Launching | A start launches a session in a worktree the command made, and the check ends and relaunches sessions | 1, 2 | 90 |
| 3. Prompts and follow-ups | A prompt is recorded, answered from a desk session, and filed, and a session files a follow-up through the command | 4 | 60 |
| 4. Merging | The check merges a clean pull request and sends back or raises any other | 5 | 60 |
| 5. The timer and a real pass | The timer runs the check, and a pass in the scratch repo runs to its merge with no prompt | 6, 7 | 75 |

Every pass carries item 8. The check goes first because it rests on the two proofs that fell
back, as the route asks.

## Left to the builder

The package layout, field and verb names, the settings file's place in the state folder, the
template wording, the tools each kind leaves out, how a test shortens the 30-second wait, and
how tests stand in for `claude`, `gh`, and the network probe. The stand-in follows the usage stand-in the last round built: a script the test
writes, named by a hidden flag.

## Steps only Geoff can do

Two, both before the round starts. The round plan gives the exact text of each.

- An entry in his Claude Code auto-mode settings that lets this round's sessions launch and end
  `cc-desk-scratch-*` sessions through the built `cc-desk` binary, install the timer, and let
  the check merge in the scratch repo. No session can write that entry for him.
- Accepting Claude Code's trust prompt once in `~/Projects/cc-desk-scratch`, the clone the live
  pass runs in. The proof found that a folder he hasn't trusted stops a launch.

## What done looks like

Every check runs the built `cc-desk` binary. Items 1 to 5 run in CI against a fixture state
folder, with stand-ins for `claude`, `gh`, and the network probe that log each call. `git` is
real, in a repo the test makes. Items 6 and 7 run on the workstation from one script in the
repo, and the round report holds its output.

1. **Launch.** Starting a free fixture pass fetches, makes one worktree on a new branch cut
   from the default branch's remote head, and writes a settings file holding the project's
   allow rules, the hook, and the two merge deny rules. The `claude` stand-in's log shows one
   call. That call runs inside `systemd-run --user --scope` from the worktree, carries
   `--permission-mode auto` and the settings file, puts the prompt before `--disallowedTools`,
   and holds neither `bypassPermissions` nor `--dangerously-skip-permissions`. The release
   file holds the session's name, ids, worktree, and branch. A second start of that pass is
   refused with an error naming the live record, and a launch whose worktree another record
   holds is refused with an error naming that record. A start for a release with no checkout
   recorded is refused.
2. **Recovery.** A stub conductor's row vanishes, and the next check launches it again with no
   stop in the snapshot. A conductor that reported finished is not launched again. A pass
   session's row vanishes three times. The first two checks relaunch it in the same worktree,
   the commit at the branch's tip is the same before and after each, the release file's count
   reads 1 and then 2 with a cause each time, and each relaunch prompt holds the relaunch
   paragraph. The third check writes "Needs you now" with the round's name and launches
   nothing. With a row still present, the log shows `claude stop` for the old id, a list read
   that shows it ended, and only then the launch. When the stop fails, no launch follows. A
   row missing at the first read and back at the second is left alone. After all rows vanish
   at once with three passes recorded as building, one check relaunches each, and none while
   the usage stand-in reads at the stop line.
3. **Stops the check reads.** Fixture rows, one each: a permission prompt, "Login expired", a
   usage limit, an unreadable block, a lost connection with the probe failing, the same with
   the probe answering, a session idle 16 minutes, one idle 14, one idle 16 with a ready
   report standing, a launch with no row after one check, and the same after two. The
   snapshot names the first four as "Needs you now", each with its own stop reason and the one
   thing to do. The fifth is left alone, and its relaunch count doesn't change. The sixth and
   seventh are ended and relaunched. The eighth and ninth are left alone. The tenth launches
   again, and the eleventh reads "Needs you now". A usage-limit session relaunches when the
   usage stand-in drops under the stop line, and its count doesn't change. Three failed list
   reads in a row read "Needs you now".
4. **Prompts.** The hook command, given a `PermissionRequest` input recorded from a real
   session, adds the prompt to the release file with its tool and command, and prints nothing.
   A yes from a desk session adds the rule to the release file. The log then shows the old
   session ended, a settings file written after the yes that holds the new rule, and a launch
   naming that file. The relaunch count is unchanged. A no does the same with no new rule,
   and the relaunch prompt names the refused command. The same yes from a background session
   is refused. A follow-up file for the prompt exists.
5. **Merge.** Six fixture pull requests: clean; no check registered; a check running; a check
   failed; outside the design; last commit moved. Only the first merges, with a merge call
   pinned to its commit, and its pass reads done with the merge time. Its worktree and branch
   are gone, and the log holds no force flag. The failed one relaunches its session once, and
   a second check with no new report does nothing. After a new report on a new failed commit
   it relaunches again, and after a third it becomes a question. A pull request with no check
   after 31 minutes reads "Needs you now". The `gh` stand-in's log shows no merge call for
   the other five.
6. **A real pass, no prompt.** On the workstation, with the timer installed, a desk session
   adds a fixture pass to a release for the scratch repo, and the timer's check launches it.
   The session edits a file, runs the gate, opens a pull request, and reports ready. The
   check merges it after its check passes: the release file's merge record names the check,
   and the session's log holds no merge command. Within 10 minutes of the merge the session,
   the worktree, and the branch are gone. The hook is shown to fire first: a third session
   launched with one rule left out records its prompt. Then the release file records zero
   prompts for the first pass, and the denials log holds no row for it. Nobody typed anything
   after the start. A second such pass has its session ended with `claude stop` mid-run. It
   is running again within 10 minutes, and the release file shows one relaunch with its
   cause. The round report gives each run's times and the prompt count. Any prompt that did
   appear is in the state file with its cause, counted in the report, and filed as a
   follow-up.
7. **The timer.** On the workstation, `systemd-analyze verify` passes on both unit files, and
   `cc-desk timer status` names a run in the last 6 minutes. A stand-in launch that starts a
   long sleep, run by the service through the scope, is still running after the service
   exits. A Go test reads the unit files in the repo and finds the `PATH` line and
   `KillMode=process`.
8. **The gate.** `bash scripts/gate.sh` passes, with no skipped test. Items 6 and 7 are not Go
   tests. Their script's output is in the round report, and a pass that leaves it out isn't
   done.

The main design's done item 6 also names a logout and login, and its "carries on within 10
minutes of the link's return". Item 2's all-rows case and item 3's sixth row cover those from
fixtures. A live restart of the workstation is not part of this round's checks.

## Backlog

The repo's follow-ups, read against this round:

| File | Verdict |
| --- | --- |
| `2026-11-17-rerun-prove-proofs.md` | Fold two tests: a fresh launch after a failed wake (item 2 shows it), and the supervisor after a kill (item 6's second run). Keep the rest |
| `snapshot-stop-times.md` | Keep. It is the Page round's |
| `resume-skips-the-limits.md` | Keep. Ruled "Yes as built"; a relaunch here is not a resume and takes no slot |
| `orphan-ruling-after-a-crash.md`, `orphan-question-forecast.md`, `reading-errors-and-guess-on-resume.md`, `input-before-lock.md`, `desk-start-time-after-a-clock-step.md`, `strict-reader-key-case.md` | Keep. None touches this round's files |

## Calls Claude made

Geoff said on 2026-11-17 that session behavior is Claude's to design. Each call can be
overturned at approval.

- A prompt is recorded and left open, and Geoff's answer relaunches the session. Answering it
  in flight through the hook is left for later, since it needs the hook to wait for hours.
- Only the check merges. A session reports, and the command verifies what GitHub says.
- The reviewer's verdict and "inside the design" are the session's report. The command checks
  the commit and the checks, and takes those two on the session's word.
- A relaunch after Geoff's answer to a prompt, and a relaunch to fix a pull request, don't
  count toward the limit of two.
- A pass session is Opus at high, since it plans as well as builds.
- A live logout or restart is not tested in this round.
- The unit files install by a `cc-desk` verb. The workstation repo learns of them in the Move
  round.
- A usage-limit stop relaunches by itself once the usage read is under the stop line. It
  shows as "Needs you now" meanwhile, because Geoff may want to reset early.
- A conductor or a design writer starts and ends by a verb in this round. Starting them
  without a person is left to the Desk and skills round.
- The link is tested with a plain HTTPS request, with no model and no tokens.
- A boot where the check is the first to start Claude Code's supervisor is not tested. The
  service carries `KillMode=process` against it, and a follow-up asks for the test at the
  Move round's first restart.
