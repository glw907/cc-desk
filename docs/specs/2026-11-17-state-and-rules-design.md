# Build the state files and the scheduling rules

This round makes `cc-desk` the command that holds the state of every release and enforces every
scheduling rule. It reads and writes one file per release, works out when the next decision
session is due, decides which passes are free to start, refuses a start that breaks a rule, and
keeps the log of rulings. It launches no session and draws no page. Those come in the Sessions
and Page rounds, which both build on what this round writes. This document asks for one approval.

It is the second round of the route in [2026-11-17-route.md](../plans/2026-11-17-route.md), and
it follows the approved design in [2026-11-17-cc-desk-design.md](2026-11-17-cc-desk-design.md),
called "the main design" below. Where this document is silent, the main design holds. Geoff's
decisions are marked "Geoff". Everything else is Claude's call, and the borderline ones are
listed under [Calls Claude made](#calls-claude-made).

## Why this round comes first

Every later round reads these files. The Sessions round launches what this round says is free.
The Page round draws the snapshot this round produces. The desk records rulings through the
guard built here. If the rules live in one tested place before any session or screen exists,
the later rounds can't grow a second copy of them.

## Goals

1. The command is the only writer of the state files, and two commands running at once can't
   corrupt them.
2. The next decision session's time, range, question count, and projects come out of the
   command equal to figures worked by hand.
3. A start that breaks a rule is refused with the rule's name, and a refused start changes
   nothing on disk.
4. A ruling is recorded as one line with Geoff's exact words, and only from a desk session.
5. The command reads the usage figure with no model and holds new work at the stop Geoff set.
6. The snapshot carries everything the page will show about releases, rounds, passes,
   questions, rulings, and the next session, so the Page round only draws those. Reports and
   the line for a design conversation in progress arrive with the Desk and skills round as new
   files, and they don't reshape these.

## What it leaves out

- Launching, checking, relaunching, or ending a session. A "start" in this round is a record
  in the release file. The Sessions round makes that record launch something.
- The page, the widget, and the reports' read marks.
- Opening the desk. This round builds the guard a desk session passes. The Desk and skills
  round builds the session.
- Passing a ruling to running sessions, and asking a question again after its ground changed.
- Writing follow-up files through the command. The route names no round for this. It belongs
  in the Sessions round, whose working sessions are the ones that file them.
- Starting a pass the moment a ruling frees it. Here the pass shows as free. The Sessions
  round's timed check starts it, which closes the main design's done item 4.
- Merging pull requests. One rule for that merge is decided here and built in Sessions (see
  [A decision carried to the Sessions round](#a-decision-carried-to-the-sessions-round)).

## What would make it a failure

- A rule a session can get past through the command. Every session runs as Geoff's user and
  could edit a file by hand, and this round doesn't try to stop that. The working skill tells
  sessions never to, and the strict reader catches an edit that breaks a reference.
- An estimate nobody can check by hand.
- A refusal that doesn't say which rule refused.
- A broken state file that the command reads past or repairs by guessing.
- A data model the Sessions or Page round has to reshape before it can start.

## Geoff's ruling on usage

Asked where cc-desk stops for usage, Geoff answered on 2026-11-17:

<!-- vale Vale.Spelling = NO -->
> You can go to 100%. I have a limit reset that I'd like to use this eveing.
<!-- vale Vale.Spelling = YES -->

The stop is 100 percent of the weekly
allowance. It replaces the 85 percent that `pass-core` names.

## Approaches

Three ways to hold the state were weighed. Geoff left the choice to Claude under his standing
rule that method and architecture calls are Claude's.

| Approach | What it trades |
| --- | --- |
| **Current-state files with pure rules (chosen).** Each release is one JSON file that says what is true now. Every change goes through a verb that loads the files, checks the rules, and writes back. The rules are plain functions over the loaded state and a clock. | It is the shape the main design names and the smallest of the three. History that the current state can't hold (forecasts, rulings) needs append-only logs beside the files. |
| An event log as the source of truth. Sessions append events and the state is rebuilt by replaying them. | History comes free. Every read replays the log, one bad event damages everything after it, and "a state file that fails to parse stops the command" gets harder to keep. |
| SQLite. | Questions across projects get easy. The main design rules a database out, and the page would need a second way to read. |

## The state files

The state lives in `~/.local/state/cc-desk/`, or `$XDG_STATE_HOME/cc-desk/` when that is set. A
flag points the command at another folder, which is how the tests run it on fixtures.

| File | What it holds | How it is written |
| --- | --- | --- |
| `releases/<project>.json` | One live release: its rounds, passes, questions, answers, and what rests on what | Replaced whole, atomically |
| `rulings.jsonl` | One line per ruling | Appended only |
| `forecasts.jsonl` | One line per forecast, with the time it was made | Appended only |
| `desks.json` | The desk sessions that may record a ruling | Replaced whole, atomically |
| `usage.json` | The last usage reading and when it was taken | Replaced whole, atomically |
| `settings.json` | The usage stop. Optional; absent means 100 | Edited by hand, read strictly |
| `snapshot.json` | What the page draws | Replaced whole after every change |

A project has one live release at a time. When a release ends, its file moves to
`releases/done/`, where the pace still reads its finished passes.

A release file holds these things. The field names are the builder's.

- **The release:** its project, its name, and the repo it builds in.
- **Each round:** its name, a sentence or two on what it will do, whether the first read marked it as needing Geoff, the reason when
  it needs nothing from him, the rounds that must finish before it, its planned minutes while
  it has no passes yet, and whether its design is written.
- **Each pass:** its name, what it will do, its planned minutes, the passes that must finish before it, the
  questions it rests on, its state, when it started and finished, and its actual minutes. A
  pass started on a guess also records the question, the recommendation taken, and the time.
- **Each question:** its text, its options, the recommended option, and its answer. An answer
  is one of two kinds. One points at a line in `rulings.jsonl`, and the strict reader checks
  that the line exists. The other is carried over: an answer Geoff gave before cc-desk, held
  in the release file with his words, their date, and where they were recorded. Only the verb
  that adds a release can write a carried-over answer, which is how cairn's nine answers come
  in during the Move round without a desk session.

A pass is in one of five states: not started, waiting on a limit, building, stopped, or done. A
stopped pass records its reason from a fixed list. Five reasons need Geoff, the main design's
stops that never take a guess: spend past the usage stop, an irreversible or destructive act,
a secret or a step only he can do, an `auth-data` question, and work beyond the approved
scope. The Sessions round adds the reasons its check finds. The Sessions round also adds what it needs about sessions and pull
requests to the same file.

Every file carries a format version. The command decodes strictly: an unknown field, a missing
required field, a reference to a pass or question that doesn't exist, or a loop in what rests
on what is an error that names the file. One lock on the state folder covers each load, check,
and write, so cross-project rules always see every release as it is.

## What the command does in this round

The subcommand names are the builder's. The main design fixes only `cc-desk`,
`cc-desk <project> <round>`, and `cc-desk check`, and this round builds none of those three.

| Verb | What it does |
| --- | --- |
| Add a release | Reads a release written by a planning session, validates it, and installs it |
| Record a round's design | Adds the round's passes and questions and marks the design written |
| Start a pass | Applies every rule below, then records the start, records a wait, or refuses |
| Finish or stop a pass | Records the event and its time, then settles the waiting passes |
| Record a ruling | Passes the guard, appends the line, and answers the question it names |
| List what is free | Prints the passes that could start now, and why each other pass can't |
| Read the usage | Prints the weekly figure and the reset text |
| Print the snapshot | Prints what `snapshot.json` holds, computed fresh |

**Settling.** Every verb that writes, and the snapshot verb, ends by settling the passes that
wait on a limit. It takes them oldest request first and starts each one that now fits under
both the two-pass limit and the usage stop. Nothing in this round runs on a timer, so a pass
waiting on usage starts at the first verb run after the figure drops. The Sessions round's
timed check runs the same step every five minutes.

Every verb that takes the time takes it from one clock, which the tests set. Every verb exits
with a distinct code for a refusal, and prints the refusal as one line.

## The estimate

Each written question has a needed time: the moment the first pass that rests on it could
start. The command works it out from the main design's three steps.

1. Take every unfinished pass in that release that must finish before that pass, directly or
   through another pass or an earlier round. A round with no passes yet counts as its
   round-level planned minutes.
2. Multiply each one's planned minutes by the project's pace.
3. Add the total to the current time.

One refinement keeps the estimate from sliding while a pass builds. A pass that is building
counts as its planned minutes times the pace, less the minutes since it started, and never less
than zero. At the moment a pass starts, finishes, or stops, this gives the same figure as the
three steps. A question whose earlier work is all done is needed now.

**The pace** is the project's actual minutes divided by its planned minutes, summed over its
last ten finished passes. The range uses the smallest and largest single-pass ratio among those
ten. With fewer than three finished passes the pace is 1.0, the estimate is marked rough, and
there is no range. A pass's actual minutes run from its start to its finish, less any time it
spent stopped.

A worked case, which the fixture reuses. A project has twelve finished passes, each planned at
100 minutes. The two oldest took 300 each and fall outside the last ten. The ten that count
took, oldest first, 150, 80, 120, 120, 120, 120, 120, 120, 125, and 125 minutes. That is 1,200
actual against 1,000 planned, so the pace is 1.2, the fastest ratio is 0.8, and the slowest is
1.5. A question rests on a pass with two unfinished passes before it, planned at 100 and 200
minutes.

| Moment | Pace | Needed time | Range |
| --- | --- | --- | --- |
| 2:00 pm, the 100-minute pass starts | 1.2 | 8:00 pm (300 times 1.2 is 360 minutes) | 6:00 pm to 9:30 pm |
| 2:30 pm, that pass still building | 1.2 | 8:00 pm (120 less 30, plus 240) | 6:00 pm to 9:30 pm |
| It finishes at 2:50 pm, in 50 minutes | 1.1 | 6:30 pm (200 times 1.1 is 220 minutes) | 4:30 pm to 7:00 pm |
| Or it finishes at 5:20 pm, in 200 minutes | 1.25 | 9:30 pm (200 times 1.25 is 250 minutes) | 8:00 pm to midnight |

Each finish drops the 150-minute pass from the ten and adds the new one. After the fast finish
the ratios run from 0.5 to 1.25, and after the slow one from 0.8 to 2.0.

**A round not designed yet.** A round marked as needing Geoff with no design has no questions
to count. Its needed time is its start, worked the same way, and the snapshot names it as not
designed yet. Such a round can set the session's time. If its start arrives and it still has
no design, it stops setting the time and the snapshot lists it as waiting on its design. The
session's time then comes from the next earliest item.

**The next session** is set by the earliest needed time across every project. It covers every
written, unanswered question in any project. The snapshot gives its time, the count of
questions, and the rounds and projects they come from. The range and the rough mark are those
of the one item that sets the time, so a rough estimate in another project doesn't mark it. Its
state is one of four, and the first that applies wins:

- **Needs you now:** a pass is stopped for one of the five reasons that need Geoff, or the
  command can't read the usage figure.
- **Ready now:** a written question is needed now.
- **Scheduled:** the earliest needed time is ahead.
- **Nothing waiting:** no written question is open and no round marked as needing Geoff lacks
  a design.

A pass that waits because usage is at the stop is not "Needs you now". The main design gives
that state to a session that a usage limit cut off, which the Sessions round detects.

The command appends a forecast to `forecasts.jsonl` each time a pass starts, finishes, or
stops, with the time it was made.

## What is free to start

A pass is free when all of these hold: its round's design is written, every pass and round
before it is done, every question it rests on is answered, and nothing before it rests on an
open guess.

A request to start a pass ends one of four ways.

| Outcome | When |
| --- | --- |
| Started | The pass is free, fewer than two passes are building on the host, and usage is under the stop |
| Waiting on the two-pass limit | The pass is free and two are building anywhere on the host, in any projects |
| Waiting on usage | The pass is free and the weekly figure is at or over the stop |
| Refused | Anything else. The refusal names every rule the start breaks |

**Starting on a guess.** A conductor that judges the work cheap to undo asks to start a pass on
the recommended answer to a question nobody has answered. The request says so outright. A
plain start request for such a pass is refused, and the refusal names the question. The
command then applies the two caps from the main design.

- **The 300-minute rule.** The planned minutes of every pass now resting on an open guess, in
  every project, plus this pass, stay at 300 or under.
- **The one-pass rule.** No pass starts, on a guess or not, while any pass before it rests on
  an open guess.

A wait records every limit that holds the pass, and settling starts it once none does. A guess
start that waits on a limit counts against the 300 minutes from the moment it is accepted.

A guess is open from that moment until the question is answered, whether or not the
pass has finished. A ruling closes it. The command records whether the guess held (the ruling
chose the recommended option) or missed, and either way its minutes leave the 300-minute
total. What to do with work built on a missed guess is the conductor's call in a later round.

Refusals use four fixed names, and a test matches the exact set of names in each refusal: "the 300-minute rule", "the one-pass
rule", "an unanswered question", and "unfinished work before it".

## Rulings and the guard

A ruling is one line in `rulings.jsonl`: a short id, the time, the desk session that recorded
it, Geoff's exact words, and, when it answers a question, the project, round, question, and
the option chosen. A ruling need not answer a question. The usage ruling above is that kind.
Recording a ruling that answers a question marks the question answered, closes any guess on
it, and frees the passes that waited on it alone.

The guard passes a ruling only when all three of these hold.

1. The caller's session id, which Claude Code puts in `CLAUDE_CODE_SESSION_ID`, is listed in
   `desks.json`.
2. The caller is a descendant of the process `desks.json` records for that session, matched by
   process id and start time.
3. That process has a controlling terminal.

Anything else is refused, the refusal says which test failed, and the log is unchanged. A
background session fails the first test. A shell that copies the desk session's id fails the
second.

This changes one case in the main design's done item 4. That item refuses a ruling "from a
shell with the desk session's id and no terminal". The probe below shows that a desk session's
commands never have a terminal themselves, so the test moves to the desk's process. The case
becomes a ruling from a shell with the desk session's id that isn't under the desk's process. As the main design says, the guard
stops a mistake and isn't built to stop a forgery.

These tests rest on a probe run in a terminal session on 2026-11-17. A command run by Claude
Code's Bash tool had `CLAUDE_CODE_SESSION_ID` set, had no terminal, and had the `claude`
process as its parent. That parent had a terminal.

Nothing in this round writes `desks.json` from the command line, since a verb that registers
any session would undo the guard. The round builds the function that writes an entry, and the
Desk and skills round calls it when `cc-desk` opens a desk. The tests seed the file.

## The usage stop

The command reads usage with `claude -p "/usage" --model haiku --output-format json`, the
source proof 7 found. It takes the percent and the reset text from the line "Current week (all
models)". It keeps each reading for five minutes, so a burst of start requests costs one read. The read
happens before the command takes the state lock, so a slow read holds nobody up.

The stop is the `usage_stop_percent` value in `settings.json`, 100 when the file is absent. At
or over it, free passes wait on usage and the snapshot shows the reset text as read. Passes
already building carry on.

That output is written for people, so a wording change can break the reader. A reading the
command can't parse counts as a stop: free passes wait, and the snapshot reads "Needs you now"
with "the usage figure can't be read". The command never assumes a figure.

## The snapshot

`snapshot.json` is rewritten after every change, and the print verb computes the same thing
fresh. Nothing is stored only there. It carries:

- the next session: its state, time, range, the rough mark, the question count, the rounds and
  projects, and the rounds named as not designed yet;
- what waits for Geoff now, one entry per round with written questions that are needed;
- each release, with every round and pass: name, state, planned and actual minutes, the planned
  minutes of unfinished work before it, its needed time, and what it needs from Geoff (the
  questions, or nothing with the reason, or not designed yet with the forecast time);
- every pass waiting on the two-pass limit or on usage, with the reset text;
- every open guess, with its question and its minutes, and the total against 300;
- every ruling.

Times in the snapshot are absolute. The page turns them into "20 minutes ago".

## When something goes wrong

| What happens | What the command does |
| --- | --- |
| A state file fails to parse or fails a strict check | Exits with the file's name and the error, writes nothing, leaves `snapshot.json` as it was, and writes `snapshot-error.json` with the same message. The next clean run removes that file. The page will read it as "stale" |
| Two commands run at once | The second waits on the lock, up to 10 seconds, then exits with an error |
| A start breaks a rule | Refused, with every broken rule named. Nothing is written |
| A ruling fails the guard | Refused, saying which of the three tests failed. Nothing is written |
| The usage read fails or can't be parsed | Free passes wait, and the snapshot reads "Needs you now" |
| A verb names a project, round, pass, or question that doesn't exist | Exits with the name it couldn't find |
| A pass is finished twice, or finished before it started | Refused. The state doesn't change |

| Limit | Value |
| --- | --- |
| Work resting on open guesses, all projects | 300 planned minutes |
| Depth of guessed work | One pass |
| Passes building at once | 2 |
| Usage stop (Geoff) | 100 percent of the week |
| Age of a usage reading before it is read again | 5 minutes |
| Passes used for the pace | The last 10 per project; under 3, the pace is 1.0 |
| Wait for the state lock | 10 seconds |

## The passes

The route's cut stands: four passes, 315 planned minutes.

| Pass | What it builds | Planned minutes |
| --- | --- | --- |
| State files and fixtures | The types, the strict reader, the atomic writer, the lock, the verbs that add a release and record a design, start, finish, and stop, the snapshot's shape, and the fixtures | 90 |
| The estimate | The pace, the needed times, the next session, and the forecast log | 75 |
| The waiting order and caps | What is free, the four outcomes of a start, both caps, the two-pass limit, and the usage stop | 90 |
| The rulings log and its guard | The log, the desk record, the three tests, and what a ruling frees | 60 |

The first pass goes first and lays down the command's wiring and the snapshot's type, so the
other three each work in separate files. Which of those three run side by side is the
plan's call.

## A decision carried to the Sessions round

The Prove pass found that `approval-queue land` merged a pull request three seconds before its
check had registered. The open-tests file asks this round to decide whether a merge should
wait. It should. cc-desk's merge waits until at least one check has registered on the pull
request's last commit and every registered check has passed. The Sessions round builds that in
its "merge with a record" pass.

## Left to the builder

The Go package layout, the field names, the subcommand and flag names, the exit codes, the
format of the short ruling id, how the tests give a process a controlling terminal, and how the
fixtures replace the clock and the usage read. The command line uses cobra, as the house Go
conventions set, and adds no other dependency without a reason the plan records.

## What done looks like

Each item runs the built `cc-desk` binary against a fixture state folder, and each can fail.

1. **Estimate.** A fixture holds two releases in two projects, one of them the worked case
   above with all twelve finished passes. The snapshot's next session equals each row of that
   case's table, as literal times in the test: 8:00 pm with 6:00 pm to 9:30 pm at the start and
   again at 2:30 pm, then 6:30 pm with 4:30 pm to 7:00 pm after the fast finish, or 9:30 pm
   with 8:00 pm to midnight after the slow one. The question count and the project list equal
   stated figures. The second project has two finished passes and a later needed time: its
   estimate is marked rough with a pace of 1.0 and no range, and the session carries the first
   project's range and no rough mark. A variant makes the second project the earliest, and the
   session is then rough. A round marked as needing Geoff with no design is named as not
   designed yet, with a start time equal to a stated figure. Each line that `forecasts.jsonl`
   gains holds the time it was made and the same needed time the snapshot showed.
2. **Waiting order.** Five cases run with two projects, as the main design's done item 5
   lists them. The set of names in each refusal is exactly the set given here.
   - A free pass starts. A plain start of the pass resting on a question is refused with "an
     unanswered question".
   - With only a 400-minute pass resting on a question, a guess start is refused with "the
     300-minute rule", nothing starts, and the snapshot reads "Ready now".
   - A 200-minute pass starts on a guess, and the file records the question, the
     recommendation, and the time.
   - A 150-minute guess start in the other project is refused with "the 300-minute rule".
   - The guessed pass finishes with its question still open. A 50-minute pass after it is
     then refused with "the one-pass rule" alone.

   After each refusal the state folder is byte for byte what it was.
3. **Limits.** With one pass building in each project, two more free passes are requested and
   both are recorded as waiting on the two-pass limit. Finishing one building pass starts the
   earlier request and leaves the later one waiting. With a recorded usage reading of 99 a
   free pass starts. With 100 it waits and the snapshot shows the reset text, and a later
   reading of 99 followed by any verb starts it. With `usage_stop_percent` set to 85 and a
   reading of 87, it waits. With a reading the command can't parse, it waits and the snapshot
   reads "Needs you now".
4. **Ruling.** A ruling recorded from a process under a registered desk process that has a
   terminal adds exactly one line to `rulings.jsonl`. Its words equal the typed words byte for
   byte, in a case that includes quotation marks and a non-ASCII character. The question shows
   as answered in the snapshot and the pass that waited shows as free. One case rules for the
   recommended option on the 200-minute guess: the guess reads held and the open-guess total
   drops from 200 to 0. Another rules for a different option, and the guess reads missed. Three
   requests are refused, each naming its test: a session id that isn't in `desks.json`, the
   desk's id from a process that isn't under the desk's process, and a registered desk process
   that has no terminal. After each refusal the log is byte for byte what it was.
5. **A broken file.** Each of these makes every verb exit with an error that names the file: a
   syntax error, an unknown field, a missing required field, a reference to a pass that doesn't
   exist, an answer that points at no ruling, and a loop in what rests on what, each in a
   release file; then a syntax error in `desks.json`, in `settings.json`, and in one line of
   `rulings.jsonl`. Each time `snapshot.json` is unchanged and `snapshot-error.json` exists.
   Restoring the file and running any verb removes it.
6. **Two writers.** Twenty event verbs launched at once, spread over both releases and
   including rulings, all succeed. Both release files parse and together hold every event,
   `forecasts.jsonl` holds one line for each pass event, and `snapshot.json` equals what the
   snapshot verb then prints. The usage read in this test is a stand-in that takes 15 seconds,
   longer than the wait for the lock.
7. **The usage reader.** Given a recorded result whose first percent is a different line, the
   reader returns the figure on the "Current week (all models)" line. Run live on the
   workstation, the usage verb prints the same percent as that line of the raw result it
   saved, and the result reports a cost of zero and no model turn. The live run is outside CI,
   and the round's report gives its output.
8. **The gate.** `bash scripts/gate.sh` passes, and so does CI on the pull request. The gate
   fails if any test was skipped, so a runner without a terminal can't pass item 4 by skipping
   it.

## Calls Claude made

These are borderline. Each can be overturned at approval.

- The usage stop is a value in `settings.json`, so a later ruling changes it without a code
  change. Every other limit is a constant in the code.
- A usage figure the command can't read counts as a stop and shows as "Needs you now". At a
  stop of 100 this costs little, and it keeps the rule safe if the stop is lowered later.
- The stop reads the "all models" line only. The separate Fable line isn't a stop, since
  background sessions run on Sonnet and Opus.
- A building pass counts as its remaining minutes in the estimate, a small step past the main
  design's three steps, so the time doesn't slide later while a pass builds.
- "Ready now" means a written question is needed now, even while other work in that release
  carries on.
- A guess stays open, and counts against 300, until Geoff rules on its question, even after
  its pass has finished.
- A missed guess leaves the 300-minute total at the ruling. The redo is new work.
- One live release per project, named by the project.
- The guard ties a ruling to the desk's process and its terminal, which is stronger than the
  main design's "no terminal" test and costs nothing more.
- A merge waits for a check to register before it lands.
- An answer Geoff gave before cc-desk is carried in the release file with his words and their
  source. It doesn't go in the rulings log, which stays desk-only.
- The guard's terminal test changes one refusal case in the main design's done item 4, as
  "Rulings and the guard" says.
- Hand edits to the state files are forbidden by the working skill and not blocked by the
  command. The guard already accepts that it stops mistakes, not forgery.
- Writing follow-up files through the command moves into the Sessions round, since the route
  gave it no round.
- A pass waiting at the usage stop shows as waiting, and only an unreadable figure or a
  cut-off session shows as "Needs you now".
