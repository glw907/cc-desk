# cc-desk Desk and skills: the round design

This is the fifth round of the cc-desk route. After it, Geoff types `cc-desk` in a terminal and
a session opens that already knows what is waiting for him. He answers in his words, the
session records the ruling, and the work that waited starts. The page gains its report pages
and design pages. The round ends with one real change to cc-desk that goes from his answer to
a merge with nothing from him in between.

The main design is [2026-11-17-cc-desk-design.md](2026-11-17-cc-desk-design.md), and this round
builds its section "The desk" and closes its done items 4 and 7 and the page half of item 6.
The rounds it builds on are in [2026-11-17-sessions-design.md](2026-11-17-sessions-design.md)
and [2026-11-17-page-design.md](2026-11-17-page-design.md). Where this file is silent, those
three hold.

**What this asks of you.** One approval, of this design and the round plan beside it
([2026-11-18-desk-skills-round.md](../plans/2026-11-18-desk-skills-round.md)). Approving starts
six passes that build without you, planned at 575 minutes of session time. It also approves the
auto mode entry in that plan. At the end you sit down once, for the five steps under "Steps
only Geoff can do". To change anything, name it before you approve. "Calls Claude made" lists
what was decided without you. Its first three calls move work out of this round, and its
fourth changes when work starts.

## Why this round

The command can already record a guarded ruling, start a session, and merge a clean pull
request on a timer. Nothing opens the desk. Typing `cc-desk` prints help, no skill exists, and
every launch prompt still tells its session to follow the old process and write its questions
into a status file. Geoff still reads reports in a terminal, and the page has nowhere to show
one.

The route (`docs/plans/2026-11-17-route.md`, the list of rounds) planned three passes for this
round, at 225 minutes. Two earlier rounds sent work here. The Page round left the report pages,
the design pages, and "Mark as read". The Sessions round left three repairs to the timed check
that Geoff ruled on (2026-11-18). The round is six passes (Geoff, 2026-11-18: "six passes is
fine").

## Goals

- Geoff types `cc-desk` and reads what is waiting without asking for it.
- Every question he is asked can be answered by someone who has just sat down, with no memory
  of the work.
- His typed answer becomes a ruling in his exact words, and the work it frees starts within
  5 minutes with no other action from him.
- A report reaches him on the page, and one click marks it read.
- One real design runs from his answer to its merge and cleanup while he is away.

## What it leaves out

- **Design writers that start by themselves.** A design writer still starts when a desk session
  runs the start verb for it, and nothing merges its pull request. Both move to the Move round,
  where cairn's release has several rounds to prove them on.
- **Building on a guess without a person.** A pass that rests on an unanswered question waits
  for the desk. The command can already start one on the recommended answer
  (`pass start --guess`), and a desk session can still ask for that. The check doesn't.
- **Checking earlier answers after a round finishes.** No verb reopens a question yet. The
  round files a follow-up.
- **Moving cairn and dubplate.** That is the Move round's, with the old system's removal after
  it. The workstation's old desk rules and its `approval-packet` skill stay until then.
- **A logout or restart with work in flight.** The Move round has the first planned restart.

## What would make it a failure

- The desk speaks in shorthand. Geoff's correction of 2026-11-18 is the test: "The verbiage
  needs to be useful for a human without immediate context, not agent shorthand."
- The desk asks "shall I record that?" after a clear answer, or records words he didn't type.
- Anything asks for him between his answer and the report in the real run: a permission
  prompt, a folder trust question, a stop, or a second session.
- The page gains a second control that writes.
- A ruling can be recorded from a session that `cc-desk` didn't open in his terminal.

## Approaches

Three ways to open the desk were weighed.

| Approach | What it trades |
| --- | --- |
| **The command replaces itself with Claude Code (picked).** `cc-desk` builds the brief, registers the desk, and then becomes the `claude` process | The process Geoff started is the session, so the guard's record of it is exact. It rests on one fact, probed on 2026-11-18 and given under "The desk opens" |
| The command starts Claude Code as a child and waits | Simple, but a second process sits between the terminal and the session for hours, and the guard would have to match a child |
| Geoff starts `claude` himself and the skill registers the desk | Nothing to build, but he types more than `cc-desk`, and the session can't speak first |

The real design for the last pass was also Geoff's pick (2026-11-18, "Option 1"): a small real
change to cc-desk itself, over a change in another repo or a fixture in the scratch repo.

## The desk opens

Bare `cc-desk`, typed in a terminal, opens the desk. With no terminal it prints help, as today.
`cc-desk <project> <round>` opens the same session on one round.

Before it starts Claude Code, the command does three things. It builds the brief from the state
files. It picks a new session id and registers the desk under that id, the process id it runs
under, and its start time. It writes a settings file that allows the command's verbs, so the
session meets no permission prompt when it records a ruling. Then it replaces itself with
`claude` on Opus at high effort, passing the session id.

The probe of 2026-11-18 on Claude Code 2.1.327 showed the facts this rests on. A session
started as `claude --session-id <id> "<first prompt>"` in a terminal is listed as
`interactive` under that id and the launched process id. Inside it, a shell command reads the
same id from `CLAUDE_CODE_SESSION_ID`, and its parent is that process.

The session opens in the main checkout of the project whose question comes first, or of the
project named. Claude Code asks a trust question the first time it runs in a folder, and the
same probe showed that it asks none in a project's main checkout it has run in before.

Geoff sees the opening and never the brief. A Claude Code session speaks only after a first
message, and the probe showed that a first message passed on the command line appears on
screen. So the command passes one short line, such as "What is waiting for me?", and the brief
reaches the session in a file or in its system prompt. The screen shows that line and then the
session's opening.

No flag loads a skill when a session starts. The session's system prompt names the `cc-desk`
skill, and the session loads it before it answers. Until the workstation has the skill
installed, the desk pass's tests give the session the skill by its path in the repo.

### What the session says first

The session says what is waiting, what could come back to him in the next few hours, and what
the timed check settled since he was last at the desk. Then it takes the first question. The
wording and figures below are an illustration.

```text
$ cc-desk

Two questions are waiting for you. A third may arrive this afternoon. Since you
were last here, yesterday at 4:10 pm, one change merged and one session
restarted after a dropped connection. Neither needs anything from you.

The first question is about cc-desk's status page.

When a background session gets stuck on something only you can fix, such as an
expired sign-in, the page lists it under "Needs you now" with the one thing to
do. Today that line stays after you have done the thing. The work resumes only
when someone runs a command. A small change to fix this is designed and ready
to build, and it needs one decision first.

Should cc-desk notice by itself that you fixed the problem?

  Yes  After a sign-in stop it looks again every five minutes. Once the
       sign-in is good it removes the line and resumes the work. You sign in
       and walk away. For a problem it can't see, the line names a command.
  No   Every such line names a command. The work resumes when that command
       is run here at the desk, so nothing restarts without a deliberate act.

I recommend yes. You shouldn't have to tell the system about a fix it can see.
The build waits on your answer and is planned at 60 minutes.
```

The rules behind that opening:

- One question at a time, in the order of how much planned work each one holds up.
- Each question carries its background, the plain question, what each answer causes, the
  recommendation with its reason, and what waits on it.
- Everything follows `~/.claude/docs/voice/written-for-people.md`, with Geoff's rules of
  2026-11-12: no made-up ids, work named in ordinary words, the plain question.
- A ruling's id stays in the log and on the page's rulings table. The desk never speaks one.
- When nothing waits, the session says so in one sentence and says when the next question is
  forecast.

The same skill loads in any session where Geoff asks what is waiting. There it reads the same
brief and answers. It can't record a ruling, and it says that `cc-desk` opens the session that
can.

### A round's desk

`cc-desk <project> <round>` gives the session that round's design, its questions, and the calls
already made in it. If the round has no written design, the session is the start of a design
conversation, and the page shows it as one line with no bar until the session ends or the
design is recorded.

## Rulings

Geoff answers in his words and can argue first. When his words settle the question, the
session records them through the command and then tells him.

```text
> yes, but only after two clean looks in a row. one could be luck

Recorded at 2:43 pm as your ruling, in your words:
  "yes, but only after two clean looks in a row. one could be luck"
It answers "Should cc-desk notice by itself that you fixed the problem?"
The build starts within a minute. The second question is about cairn.
```

The session records without asking first. It asks only when his words leave the choice between
the answers open. His words reach the command unchanged, whatever characters they hold.

The guard is the main design's. A ruling is refused from a background session, and from a
shell that holds the desk session's id with no terminal. `desk register` stays for a session
already open in his terminal, with the same guard, and its read of the session list gains a
30-second limit.

A ruling already starts the timed check without waiting for the next five minutes. Today that
run starts nothing, because the check launches only a pass that a command has already started.
The main design says otherwise: "A round inside an approved release whose design raises no
question of his starts without a further approval."

So the check gains one step. It starts each pass that is free to start, by the rules the start
command already applies. A pass is free when its round's design is written, every pass before
it is done, and none of its questions waits for an answer. The pass that waited starts on the
run his ruling set off, unless two passes are already building. The check starts no pass on a
guess, and none in a release that records no checkout.

### A wrong record

If the session recorded the wrong thing, Geoff says so, and a new ruling replaces the old one.
Both lines stay in the log, and the page marks the first as replaced. What happens to the work
depends on how far it got:

| The pass that waited | What the replacement does |
| --- | --- |
| Hasn't started | It starts on the new ruling |
| Is building | Its session is ended, and a new one starts from the pass's branch with the new ruling in its prompt. The relaunch isn't counted toward the pass's limit of two relaunches |
| Has merged | The replacement is refused with that reason, and the desk says a new change is needed |

### A question about a pull request

The check asks a merge question when a pull request lies outside its design or has used its
two fix attempts. The question names the pull request by what it does. It has three answers:

| Answer | What the check does |
| --- | --- |
| Merge | Merges on its next run if every check has passed and the last commit is the one he was asked about. Otherwise it asks again |
| Fix again | Sends the pass back with his words for one more attempt. If that attempt comes back the same way, the question is asked again |
| Close | Closes the pull request, ends the session, and records the pass as closed by his ruling. The pass leaves "Needs you now" |

A closed pass didn't merge, so a pass that comes after it doesn't start, and the page says
what it waits on. The round's report counts the closed pass as finished and says it was
closed. Work to replace it takes a new design, and this round builds no verb for that.

The desk also takes the two things "Needs you now" sends Geoff to it for. It answers a
session's permission prompt with his yes or no, and it starts a stopped pass again when he
says the cause is fixed. Both verbs exist today.

## What the records gain

The desk and the page read the same records, so the first pass widens them.

- **A question** gains its background, what each answer causes, the reason for the
  recommendation, and a mark that it is never guessed. A working session can add one to its
  round from a file.
- **A ruling** can name the ruling it replaces.
- **A pass** can be recorded as closed by a ruling.
- **A release** can name commands that always ask Geoff, beside the commands it allows. A probe
  on 2026-11-18 showed that such a rule stops a session at a permission prompt in auto mode,
  and that a command no rule names runs with no prompt
  (`docs/research/2026-11-18-ask-rule-probe/`).
- **A session's record** gains the commands auto mode refused, each with its cause. The repairs
  pass fills it.
- **A report** is new: a title, a summary, the calls made, what the plan missed, and the time
  Geoff marked it read. A conductor records one from a file. The folded detail (each pass, its
  pull request, its planned and actual minutes) comes from the release file.
- **A round** gains its design's path, its context paragraphs, and the calls already made.
- **A design conversation in progress** is one record: the project, the round, and the desk
  session.
- **A log of what the check carried out** is new and only grows. Today the check keeps its last
  run and nothing before it, so "since you were last here" has nothing to read.
- **A stop** gains its time, and a pass gains a reason apart from its round's.
- **An ended record** never reads "Needs you now". A failed launch still does.

A release file written before this round still reads after it.

**The brief** is a new read of these records, printed by one verb. It holds what waits, in the
order the desk asks it, what could come back in the next few hours, and what the check carried
out since the last desk session. The desk opens from it, and the skills' fixtures are made
with it.

The records pass builds every new verb the two skills name: the brief, a question from a file,
a report from a file, "mark as read", and a ruling that replaces another. The verb behind "Mark
as read" takes no usage reading and reads no session list.

## The two skills and the guide

The repo gains `skills/cc-desk/` and `skills/cc-desk-work/`, each under about 150 lines, and a
short guide for Geoff at `docs/guide.md`.

**`cc-desk`** is the desk session's. It holds how the session opens, the order of questions,
when to record and when to ask, how to say what it recorded, the three answers to a merge
question, how a wrong record is replaced, and how to answer a permission prompt or start a
stopped pass again when Geoff says to.

**`cc-desk-work`** is for every session the command launches: a pass, a conductor, a design
writer. It holds how to report a pull request ready, file a follow-up, write a question, stop,
and record a report. A session that writes a question gives it to a reader with no context
before it records it. `pass-core` keeps the inside of a pass. Both skills name the verbs as the
records pass built them.

The three launch prompts name `cc-desk-work` and drop the lines that send a question to a
status file. A relaunch that isn't counted no longer says "This is relaunch 1 of 2".

**The guide** tells Geoff what he types, what each part of the page means, what "Needs you now"
asks of him, and how to fix a wrong ruling.

The workstation takes the two skills from this repo at a pinned tag, the way it takes Survey's
skill. Once the first five passes have landed, the round's conductor tags the repo `v0.1.0`,
adds the two entries to the workstation's vendored-skills manifest, and lands that change. The
pin moves again only if the last pass changes a skill.

## The page

The mockups in `docs/specs/2026-11-17-cc-desk-mockups/screens/` set the look. This section is
about what the page says.

The sidebar gains "Reports to read":

```text
Reports to read
  cc-desk: a stuck session now clears itself
    1 judgment call, no planning miss. 1 hour ago
```

A report page has three parts, a control, and a folded detail:

```text
A report from cc-desk. Sent 1 hour ago.

A stuck session now clears itself

  A few sentences on what works now that didn't before.

Calls made along the way
  Each call, what was chosen, the other way, and why.

What the plan missed
  Each miss and what it cost.

  [ Mark as read ]  The report stays on the round's page afterward.

> The detail: each pass, its pull request, and its minutes
```

A design that waits for Geoff gets a page in the same manner: the context paragraphs, "Yours to
decide" with each question written as it is at the desk, "Already decided", and the full design
folded below. It ends with the command that opens the desk on that design, with the terminal
link beside it.

The front page gains a short "Since the last check" list in sentences. Each running session
shows its state in words, such as "building" or "restarted once after a dropped connection". A
stop shows its reason in words and its time. A pass page shows each permission prompt and each
refused command its sessions met, with the cause.

Three rules hold the page to the main design:

- The option cards on a design page are for reading. Geoff answers at the desk, so nothing on a
  card looks clickable. The mockup drew radio buttons.
- "Mark as read" is the only thing the page writes. The click runs the command, which records
  the time. The page accepts that one form and refuses every other write as it does now, and
  SvelteKit's origin check is turned back on for it.
- No script and no raw HTML from a design file reaches the page.

The page's service gains write access to the state folder and nothing else, and it names the
command by its full path.

## The check's repairs

Five repairs, three of them ruled by Geoff on 2026-11-18.

**Refused commands are recorded.** Auto mode refuses a command without asking anyone, and
today no record shows it. Each launched session's settings file gains a `PermissionDenied`
hook that records the refusal on the session's record, the way a prompt is recorded now. The
workstation's hook of the same kind has logged refusals with their reasons under Claude Code
2.1.327, so the event fires in auto mode. A refusal is counted in the report and filed as a
follow-up. It doesn't stop the session.

**A model limit holds the relaunch.** A session stopped on a usage limit is relaunched once the
reset time in the stop's text has passed. When the text gives no time the check can read, the
first usage relaunch isn't counted, and each one after it counts toward the pass's two
relaunches.

**The cleanup stops the session.** In the Sessions round a session stayed listed after its
pass merged. The cleanup gains a step that stops the merged pass's session. A stop that fails
on a session already gone counts as ended. The pass also merges a fixture pass in the scratch
repo with the timer on and lists the sessions each minute for 10 minutes, as Geoff's ruling
asks.

**A merge call can't hang or fail forever.** The merge call and the cleanup's fetch and pull
each get a 60-second limit. Three failed merge calls in a row on one pull request make a
question for Geoff.

**The conductor starts when a report is due.** A round's report is due when its last pass
merges, and a release's when its last round does. The check starts the release's conductor if a
report is due and none runs, and relaunches it twice at most. The conductor writes the report,
records it, and reports itself finished. A relaunched conductor that finds the report recorded
writes no second one. It has no other job in this round.

## The rehearsal and the real run

The last pass proves the whole path twice: once in the scratch repo with nobody watching, and
once on cc-desk itself with Geoff.

**The rehearsal.** A script opens the desk on a terminal it makes, reads the opening, types an
answer to a fixture question, and checks the ruling. The pass that waited starts, makes a small
Go change behind a gate of the same shape as cc-desk's, and merges by the timed check. The
session, worktree, and branch go. The check starts the conductor, which records the report,
and the script marks it read with the same request the page's button sends. The scratch
release carries the allow rules the real release will carry, so the real run meets no command
the rehearsal didn't.

**The recovery checks.** With the real conductor and the live page, the script shows the page
half of the main design's item 6. A killed conductor is running again within 10 minutes and the
page shows no stop. A pass session killed three times reads "Needs you now" with the round's
name after its second relaunch. A session launched under a release that names one command as
always asking stops at that prompt and reads "Needs you now" within 10 minutes.

**The real run.** The change is small and real: today a stop stays on the page after Geoff
fixes its cause, and the work resumes only when someone runs a command. Its design is
[2026-11-18-stuck-sessions-design.md](2026-11-18-stuck-sessions-design.md), and it carries one
question that is never guessed. Once the last pass has landed, the round's conductor adds a
release for cc-desk to the real state folder with that one round, one pass, and the question.
The page then shows one question waiting. Geoff's answer at the desk is his approval of that
design.

## When something goes wrong

| What happens | What the command does |
| --- | --- |
| `cc-desk` runs with no terminal | Prints help, as today |
| A state file fails to parse when the brief is built | Prints the error and opens nothing |
| Geoff's words leave the choice open | The session asks which answer he means and records nothing yet |
| A ruling answers no waiting question | It is recorded with its project and round, and nothing starts |
| A replacement names a pass that has merged | Refused, with the reason |
| "Merge" is answered and the commit has moved or a check has failed | The question is asked again |
| A report file lacks one of its parts | Refused, naming the part |
| "Mark as read" can't be recorded | The page says so, and the report stays listed |
| A write other than "Mark as read" reaches the page | Refused with 405, as now. One from another origin is refused with 403 |
| The conductor dies while a report is due | Relaunched twice at most, then "Needs you now" |
| The refusal hook can't write its record | The hook returns no decision and the session carries on. The error goes to the hook's log |
| A usage stop's text gives no reset time | The first usage relaunch isn't counted. Each later one counts toward the pass's two relaunches |
| "Mark as read" names no report, or a name that starts with a dash or holds a slash | Refused before any command runs |
| A pass comes after a closed pass | It doesn't start, and the page says what it waits on |
| A merge call runs past its limit | It is ended and counted as a failure |
| `claude stop` fails for a session already gone | The session counts as ended |
| Anything asks for Geoff during the real run | The run has failed. See "Steps only Geoff can do" |

| Limit | Value |
| --- | --- |
| From a ruling to the freed pass starting | 5 minutes at most, about 1 in practice |
| From a ruling to the page dropping the question | 5 seconds |
| Read of the session list by any verb | 30 seconds |
| A skill's length | About 150 lines |
| Conductor relaunches | 2 |
| Merge call, and the cleanup's fetch and pull | 60 seconds each |
| Failed merge calls in a row before a question | 3 |
| Further attempts after "Fix again" | 1 |
| From a merge to the session, worktree, and branch gone | 10 minutes |
| Runs of the real change after a failed one | 1 |
| Usage ceiling for the round | 100 percent of the week, as in the last two rounds |

## The passes

The round is six passes. The records come first and alone. The skills and the page start
together when the records land, and no more than two passes build at once.

| Pass | Outcome | Done items | Minutes |
| --- | --- | --- | --- |
| 1. The records | Questions, rulings, reports, rounds, and stops carry what the desk and the page say, the check keeps a log, and one verb prints the brief | 1 | 120 |
| 2. The skills and the guide | The two skills and the guide exist, and the launch prompts name the working skill | 2 | 75 |
| 3. The desk | `cc-desk` opens the session, the check starts free passes, a wrong record is replaced, and a merge question has answers | 3, 4, 5 | 100 |
| 4. The page | Reports, designs, "Mark as read", and the check and sessions in words | 6 | 90 |
| 5. The check's repairs | Refusals are recorded, a model limit holds, the cleanup stops the session, a merge call is bounded, and the conductor starts for a report | 7 | 100 |
| 6. The rehearsal and the real run | The whole path runs in the scratch repo, the recovery checks pass on the live page, and the real release is ready for Geoff | 8, 9, 10 | 90 |

Every pass carries item 11. The skills follow the records, because they name its verbs and
their fixtures come from its brief. The page follows the records too. The desk follows the
skills, since the session it opens loads one. The repairs follow the desk, with which they
share the check's code. The last pass follows all five. The longest chain is the records, the
skills, the desk, the repairs, and the last pass, at 485 planned minutes. The round plans 575
minutes, against the route's 225.

## Left to the builder

- The names of the new verbs and of the new fields in the state files.
- Whether the brief reaches the desk session in a file or in its system prompt, within the
  rule that Geoff sees only the opening.
- How his words reach the command unchanged: a file, standard input, or an argument.
- Whether a second `cc-desk`, typed while a desk is open, opens a second working desk or says
  where the first is. Either way no desk opens that can't record.
- How the folded design is drawn, within the three rules under "The page". The snapshot
  doesn't carry the design's text.
- The tool that drives the desk on a terminal in the rehearsal.
- How the two skills are checked against fixtures, within what item 2 asks.
- Whether a closed pass is a new state or a mark on a finished one. Either way an older file
  still reads, and a closed pass never reads as merged.
- Which command the repairs pass uses to draw a refusal from auto mode. The workstation's log
  of refusals shows what auto mode has refused here.
- Where `cc-desk` goes on Geoff's `PATH`. `~/.local/bin` holds links into other repos' tools
  already. The desktop handler falls back to the installed path, because the user services'
  `PATH` has no `~/.local/bin`.

## Steps only Geoff can do

Before the round starts, one step: approve the auto mode entry in the round plan. It lets this
round's sessions run the command, and it lets the round's conductor install the two skills on
the workstation. It also lets the timed check merge one pull request into `glw907/cc-desk` for
the real run. Claude adds it to the settings after his approval.

At the end, one sitting. The page shows one question waiting for cc-desk.

1. Type `cc-desk`. The desk reads out the small design about stuck sessions and its question.
   Answer in your words.
2. Look at the page and confirm the ruling shown is what you said.
3. Walk away. The build starts, merges, and cleans up after itself. This is the first time the
   timed check merges into a real repo with no approve from you, as your ruling of 2026-11-17
   allows for a change inside its approved design.
4. When the report appears under "Reports to read", read it and click "Mark as read".
5. Click the terminal link once in Chromium, the click still owed from the Page round. What to
   look for is in `docs/follow-ups/page-link-chromium-click.md`.

If anything else asks for you during step 3, the run has failed its test. The next planning
session reads the record, fixes the cause, and sets up one more run with another small
follow-up. That session also adds the run's result to this round's report.

## What done looks like

1. **Records.** A fixture question with every new field reaches the snapshot unchanged. A
   report recorded from a file shows in the snapshot as unread, and marking it read records
   the time. A report file missing its title, its summary, its list of calls, or its list of
   misses is refused, and each refusal names the part. An empty list is accepted. The check's
   log holds one line for each thing two fixture runs carried out, in order. A copy of the real
   state folder taken before this round reads with no error, and after one verb rewrites it,
   every field it held is still there. An ended record with a stop is absent from "Needs you
   now", and a failed launch is present. The brief verb, run on a fixture with two waiting
   questions, puts the one that holds more planned minutes first. It prints each one's
   background, answers, recommendation, and what waits. A session launched under a fixture
   release that names a command as always asking has that rule in its settings file.
2. **Skills.** Each skill is under 160 lines and passes `vale` and `tellgrader`. The brief verb
   makes three fixture briefs: nothing waiting, one question, and a question with a merge
   question. For each, a fresh session given the desk skill and the brief opens without being
   asked. With nothing waiting, the opening is one sentence and holds the fixture's forecast
   time. With one question, the opening holds a marked phrase from each part of the fixture
   question: its background, what each answer causes, the reason for the recommendation, and
   what waits. It puts the plain question on a separate line and holds no ruling id. A reader
   with no context, given only that opening, states the question and what each answer causes.
   With two, the question that holds more planned minutes comes first, and the merge question
   names its pull request by what it does and gives the three answers. A Go test builds the
   three launch prompts the way a launch does. It finds `cc-desk-work` in each, and finds
   neither `approval-queue` nor "Until cc-desk runs". The prompt of a usage relaunch that isn't
   counted holds no "relaunch 1 of 2", and a counted relaunch's prompt still gives its number.
3. **The desk opens.** On a terminal a test makes, in the scratch repo's main checkout as it
   stands, `cc-desk` starts a session that the session list shows as `interactive` under the
   id and process the command registered. The process's arguments name Opus and high effort.
   The screen holds the one line the command passed and then the opening. It holds neither
   the brief as the command printed it nor the session's instructions. No permission prompt
   and no trust question appears. `cc-desk <project> <round>` opens on that round's questions
   only. With no terminal, `cc-desk` prints help and registers nothing. With a state file that
   fails to parse, it prints the error, starts no session, and registers nothing.
4. **Ruling.** In that session a typed answer to a fixture question adds one line to the
   rulings log with the time and the typed words, byte for byte. The test types an apostrophe,
   a double quote, a dollar sign, and a second line. The snapshot drops the question within 5
   seconds. The pass that waited starts within 5 minutes with nothing else typed: the check's
   log holds the start, and the desk session ran no start command. The same record from a
   background session is refused. So is one from a shell that the desk session started with
   no terminal, and that refusal names the missing terminal.
5. **Replacing and merging.** A replacement before the pass starts leaves both lines in the log
   with the first marked replaced, and the pass starts on the second. A replacement while the
   pass builds ends its session and starts one with the new words in its prompt. The relaunch
   count is unchanged, and the session list never shows two sessions for the pass. A
   replacement after the merge is refused. For a fixture merge question, "merge" merges at
   the commit asked about, and asks again when the commit has moved. "Fix again" relaunches
   the pass once with the words, and asks again when that attempt comes back the same way.
   "Close" closes the pull request, ends the session, records the pass as closed with the
   ruling, and leaves "Needs you now" empty. A pass that comes after the closed one doesn't
   start.
6. **The page.** The test's fixture holds an unread report, a design that waits, and a pass
   with a permission prompt and a refused command. A test walks every page and finds no form
   control except "Mark as read". After the walk the state files are byte for byte what they
   were. One click on "Mark as read" changes the report's record and the snapshot, changes no
   other file in the state folder, and takes no usage reading. The report leaves "Reports to
   read" and stays on its round's page. A write to any other address gets 405, and a "Mark as
   read" sent from another origin gets 403. A report name that starts with a dash or holds a
   slash is refused before any command runs. A design page shows each question's background,
   answers, and recommendation, and an option card holds no input, button, or link. A design
   file holding a script tag, an `onerror` attribute, and a `javascript:` link draws none of
   the three. The pass page shows the prompt and the refused command with their causes. The
   WCAG 2.2 AA scan passes in both themes, and the words test finds words beside every state
   shown in color.
7. **Repairs.** A session launched in the scratch repo on a command auto mode refuses has the
   refusal on its record with a cause. A fixture usage stop that names a reset time isn't
   relaunched before it and is after it. A fixture pass whose usage stops name no time is
   relaunched after the first stop with its count unchanged, and after the second and third
   with the count at 1 and 2. At the fourth stop it reads "Needs you now". After a fixture
   merge in the scratch repo with the timer on, the check's log shows the stop step, and a
   listing each minute for 10 minutes shows the session gone. The listing is in the pull
   request. Stand-ins that sleep for the merge call, the fetch, and the pull are each ended at
   60 seconds. The third failed merge call in a row makes a question, and a success in
   between starts the count again. With a report due and no conductor, the next check starts
   one. It starts none when no report is due, and none while one runs. A conductor ended
   three times with its report due is relaunched twice and then reads "Needs you now".
8. **Recovery on the page.** On the workstation, with the real conductor: its process is
   killed, it is running again within 10 minutes, and the page shows no stop. A pass session is
   killed three times. After the second relaunch the page reads "Needs you now" with the
   round's name, and the session list never shows two sessions for that pass. A session
   launched under a release that names one command as always asking reads "Needs you now"
   within 10 minutes, and the line names the command.
9. **The rehearsal.** In the scratch repo, from one typed answer on a made terminal: the pass
   starts, opens a pull request, and the timed check merges it. Within 10 minutes of the merge
   the session, worktree, and branch are gone. The conductor starts by itself and records one
   report, which the page lists. The release file records no permission prompt for the pass
   and lists each command auto mode refused. The transcripts hold no merge command. Nothing
   was typed after the answer.
10. **No tending.** The main design's item 7, on cc-desk itself. Geoff types `cc-desk`, answers
    the question about stuck sessions, and confirms the ruling on the page. The pass starts,
    builds, and merges with no page shown to him, and its session and worktree are gone within
    10 minutes of the merge. He typed no command but `cc-desk`, answered no prompt, and opened
    no other session. The records agree: one desk was registered in that time, the rulings log
    gained one line, and the state file records no permission prompt for the pass. The report
    reaches "Reports to read" with its count of refused commands, and he marks it read.
11. **The gate.** `bash scripts/gate.sh` passes with no skipped test. Items 2 to 5 and 7 to 10
    have parts that run live and aren't Go tests. Each live script prints one line for each
    check and ends with its result. A pass's pull request holds that output from a run on the
    pass's last commit, and a pass that leaves it out isn't done.

The main design's item 4 is this file's items 4 and 10. Its item 6 also names a logout and a
lost network link, which the Sessions round covered from fixtures.

## Backlog

The repo's 35 follow-ups, read against this round. A file marked "Fold" is removed by the pass
named. The four marked "Done" are removed in the commit that lands this design.

| File | Verdict |
| --- | --- |
| `auto-mode-prompt-clause.md` | Fold into the repairs: the refusal record |
| `cc-desk-on-browser-path.md` | Fold into the desk. Its note on the two name rules goes to the page |
| `how-a-stop-clears.md` | Fold. It is the real run's design |
| `live-check-cleanup-gaps.md` | Fold into the last pass. Its item 6, the read with no time limit, goes to the desk |
| `merge-call-fails-for-good.md` | Fold into the repairs |
| `merged-pass-session-still-listed.md` | Fold into the repairs |
| `merge-question-answer.md` | Fold into the desk |
| `page-link-chromium-click.md` | Fold. It is step 5 of Geoff's sitting |
| `pass-needs-reason.md`, `snapshot-stop-times.md` | Fold into the records, shown by the page |
| `usage-relaunch-loop-on-a-model-limit.md` | Fold into the repairs |
| `ended-record-stop-in-needs-you-now.md` | Fold into the records. A failed launch stays listed |
| `blocked-session-reads-as-unreadable-stop.md` | Reword. The stop stays, and the page says it in words |
| `launch-review-notes.md` | Reword. Item 5 goes to the skills and item 1 to the repairs. The rest waits for Move |
| `2026-11-17-rerun-prove-proofs.md` | Reword for Move |
| `desk-registration-gap.md`, `followup-rule-under-auto-mode.md`, `hook-input-from-a-real-session.md`, `ready-verb-service-start.md` | Done. Each file says the Sessions round's last pass closed it |
| `api-project-state-round.md`, `building-pass-without-checkout-needs-entry.md`, `cleanup-of-a-release-marked-done.md`, `cold-boot.md`, `desk-start-time-after-a-clock-step.md`, `input-before-lock.md`, `live-sample-of-limit-texts.md`, `orphan-question-forecast.md`, `orphan-ruling-after-a-crash.md`, `pull-request-closed-without-merge.md`, `reading-errors-and-guess-on-resume.md`, `resume-skips-the-limits.md`, `strict-reader-key-case.md` | Keep for Move. None touches this round's outcome |
| `rulings-table-reflow.md` | Keep for later page work |
| `survey-rail-muted-doc-measure.md`, `survey-table-header-contrast.md` | Keep. They are Survey's to fix |

One new file lands with this design, `earlier-answers-after-a-round.md`. The Page report named
8 high-severity `npm audit` warnings in the page's build packages and nobody filed them. On
2026-11-18 `npm audit` in `web/` found none, so no file is added for them.

## Calls Claude made

Each call can be overturned at approval. The first three change what the round covers.

- Design writers don't start by the timer in this round, and nothing merges a design writer's
  pull request. The Sessions design left that to this round, and the process caps a round at
  six passes. It moves to the Move round.
- The check doesn't start a pass on a guess. A pass resting on an unanswered question waits for
  the desk.
- Checking earlier answers when a round finishes is left out and filed.
- The check starts every pass that is free to start, in any release that records a checkout.
  Both earlier designs say it does, and the code never did. Once this lands, adding a release
  starts its free work at the next check with no further word from Geoff.
- The round plans 575 minutes. Geoff was told about 510 when he chose six passes. The builder
  check added the last 20, by moving the brief into the first pass.
- The conductor runs only while a report is due. With the check starting free passes, a
  conductor has nothing else to do in this round.
- A release can name commands that always ask Geoff. It is the one way a session cc-desk
  launches stops at a permission prompt in auto mode, and the recovery check needs a real one.
- The cleanup always stops a merged pass's session. Geoff's ruling asked for the step only if
  the 10-minute listing showed the session staying, and one clean listing would prove little.
- A closed pass holds the passes after it, and replacing its work takes a new design.
- The desk answers permission prompts and starts a stopped pass again, since the page already
  sends Geoff to the desk for both.
- The desk opens in a project's main checkout, so Claude Code asks no trust question. The
  session reads that project's instructions as any session there does.
- `desk register` stays, for a session Geoff already has open in a terminal.
- A refused command is counted and filed, and it doesn't fail the real run. Only something
  that asks for Geoff does.
- The report Geoff reads in his sitting is the real run's, written by cc-desk's conductor. This
  round's report goes the old way, to the initiatives repo and his terminal, because the
  round's passes don't run under cc-desk.
- The real change's design is written now, as a separate short file, so the desk has a real
  design to read out.
- The skills are named `cc-desk` and `cc-desk-work`, and the first tag is `v0.1.0`.
- The numbers under "When something goes wrong" that the main design didn't set: 30 seconds,
  60 seconds, 3 failed merge calls, and 1 further attempt after "Fix again".
- The rehearsal's pass makes a Go change behind a gate like cc-desk's, so the allow rules are
  proven before Geoff sits down.
