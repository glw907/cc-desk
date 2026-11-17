# Build cc-desk to replace the Approval Queue and the desk

cc-desk is a from-scratch replacement for the Approval Queue, the `desk` session, and the round
machinery around them. It has a status page that Geoff reads and a terminal session where he
decides and discusses. It forecasts when work will next need him, gathers those decisions across
projects into as few sessions as it can, and runs every background session without his help. The
current system is a prototype to borrow from or discard. This document asks for one approval of
the whole design.

Geoff's decisions in this document are marked "Geoff". Everything else is Claude's call, and the
borderline ones are listed under [Calls Claude made](#calls-claude-made).

## Why it is being rebuilt

Two days of running rounds through the desk showed where the prototype costs time. The record is
`~/.local/state/desk/notes.md` and `efficiency-notes.md`. Geoff's summary of what he wants:

- The queue's pages are "more detailed and more formal than I usually need", and they often ask
  "questions where the answer is architecturally obvious".
- He wants software that is "clean, streamlined, idiomatic for the stack, and with an eye towards
  beauty and evenness". He decides high-level design and anything about the interface. Claude is
  the technical architect.
- "Ask all the meaningful questions up front, and then give me a good summary report when the
  work is done", in every project.
- "A key problem to solve is giving me a way to tell me in advance when an effort is going to
  need my input. And to consolidate those decision sessions as much as possible."
- "The page holds status and is a reading surface. The terminal is a perfect interface for
  approvals and discussion."
- "I don't want to have to manage a bunch of background sessions."
- The tool should be "lightweight and easy to maintain", live in a single repo, and be named
  cc-desk.

## Goals

1. The page tells Geoff, at a glance, the estimated clock time of his next decision session.
2. Decisions from every running project arrive together, in as few sessions as possible.
3. He can open any round or pass that hasn't started and see what it will do and whether it will
   need him.
4. He decides by talking with a session that can answer back. His typed words are the record.
5. He never launches, names, watches, restarts, or cleans up a background session.
6. A design page asks only what is his to decide. A report says what was done, which judgment
   calls were made, and what the plan missed.
7. The whole tool is small enough that one session can hold it in its head: one repo, one
   command, one page, one timer, two skills.

## What it leaves out

- No buttons that decide. The page's only control marks a report as read.
- No desktop notices and no way to reach Geoff away from the page.
- No model of his working hours. The page shows the real time a decision is needed.
- No sync between `pavilion` and `thinkpad-x1`. cc-desk runs on one host and its state stays there.
- No change to how a single pass is planned, built, gated, and reviewed, beyond the points where
  a pass reports to cc-desk. That stays in `pass-core`.
- No history browser for the old queue's 72 stored documents. They stay on disk, unread by cc-desk.

## What would make it a failure

- Geoff can't talk a decision through when it comes in.
- He has to ask for status, or has to tend a session.
- A page asks him something a competent architect would decide alone.
- Work built on a guessed answer costs more than one small-to-medium pass to redo.
- The tool grows back the weight it was built to shed.

## The page

A working mockup on the real stack is in `2026-11-17-cc-desk-mockups/` beside this file, with
screenshots in its `screens/` folder. Its names are real and its figures are invented. The mockup
settles the layout and the look. Where this text and the mockup differ, this text wins.

The page has a wide sidebar (about 23rem) and a reading column. The sidebar holds four things, in
this order.

**The next decision session.** A boxed widget, the only box in the sidebar. It shows the best
current estimate ("Tomorrow, about 9:30 am"), the likely range, and how many questions from how
many rounds in which projects. When a session is due it reads "Ready now" and lists what waits.
When a hard stop needs him it reads "Needs you now" and names the round and the reason. The state
shows in words and in color, and the browser tab's title carries it too.

**Waiting for you.** What is ready to decide now, one line each.

**Reports to read.** Unread reports, each with its count of judgment calls and planning misses.

**Running.** One block per release. The hierarchy is release, then round, then pass (Geoff:
"n number of rounds towards a release, and n number of passes towards a given round"). A release
shows a scale bar with one segment per round, each as wide as its share of planned minutes, then
"2 of 17 rounds done" and the percent of work done. A small triangle under a segment marks a
round that will ask him something. Under the bar, each round now running shows on an indented
rule with a bar of its passes and its state in words. Rounds that are not running don't appear
in the sidebar.

The front page of the reading column repeats nothing from the sidebar. It shows what the next
session will cover (each round, what its questions are about, and when each is needed), then a
board per release listing every round, with the running rounds opened to their passes.

**Click-ahead pages.** Every round and every pass, in the sidebar and on the boards, links to
a page about it. A page for work that hasn't started says what it will do, how many minutes are
planned, how much work stands before it, and what it needs from Geoff. That last section holds
the written questions, or "Nothing" with the reason, or "Not designed yet" with the forecast.

**A design page** has three parts (Geoff): one to three paragraphs of context, enough to decide
with; the questions that are his, each with options and a recommendation; and a short list of
borderline calls Claude made. Obvious architectural calls are not listed. The full design sits
folded below. A question is his when it is about what a person sees or types, about spend, or
about scope the record leaves open.

**A report page** has three parts (Geoff): a high-level summary of what was done, the judgment
calls made mid-effort, and what the plan missed. A fourth, folded part carries the measurements
that let the system improve: planned against actual minutes for each pass, relaunches, guesses
made and whether each held, and the models used. A release ends with one report on the whole release.

The look follows Survey's plain-app recipe: stock daisyUI 5 on Tailwind 4, tokens and never
literals, no shadows or gradients, no all-caps labels, no looping motion, WCAG 2.2 contrast, and
every color state also shown in words.

Once the page works, it gets one round of the `standards:design-refinement` skill, the
"invisible craft" polish (Geoff). He doesn't review that round. Its changes are listed in the
report.

## The desk

Geoff types `cc-desk` in a terminal. It opens an ordinary Claude Code session on Opus, loads a
short skill, and hands the session a brief built from the state files. The session opens by
saying what is waiting, what could come back to him in the next few hours, and anything the
timed check settled without him. He never has to ask.

`cc-desk <project> <round>` opens the session on one round. A click-ahead page shows that
command where a button would be.

In the session he talks. He can push back, ask what an option costs, or propose another one.
When he settles a question, the session records a ruling: one dated line with a short id and his
exact words, written through the command. No page is written and no check chain runs. The
session tells him what it recorded. The command passes each ruling to the sessions it affects
and releases any work that waited on it.

No session needs to stay open. Everything a desk session knows comes from the files, so a fresh
one knows as much as an old one. The same skill loads in any session where he asks what is
waiting.

Design and planning conversations stay ordinary brainstorming sessions (Geoff). cc-desk's part
starts when one ends: the approved design enters the state files as a release, a round, or a
change to one. A design conversation in progress shows on the page as one line with no bar.

## Forecasting decisions

**The first read.** When a release is planned, each round is marked as needing Geoff or not, by
whether it touches what a person sees or types, spend, or open scope. That puts the triangles on
the bars before any question exists.

**Designs run ahead.** While one round builds, the designs for the next two rounds of that
release are already written. A background session on Opus writes each one from the release's
approved design, and writing it produces the round's questions. A round inside an approved
release whose design raises no question of his starts without a further approval. A new release
always gets one, and it begins as a brainstorming conversation with Geoff.

**The clock estimate.** Each question has a needed time, the moment the first pass that rests on
it could start. The command works it out in three steps:

1. Add up the planned minutes of the unfinished passes in that release that must finish before
   that pass.
2. Multiply the sum by the project's pace: the ratio of actual to planned minutes over the
   project's last ten finished passes.
3. Add the result to the current time.

The range comes from the smallest and largest ratio among those ten passes. With fewer than
three finished passes the pace is 1.0 and the page labels the estimate "rough". The first
version counts that work as one line of passes. It ignores the two-pass limit, other projects,
and usage pauses, and the pace takes them up as passes finish. A round marked as needing Geoff
whose design isn't written has no questions to count yet. Its needed time is its start, and the
widget names it as "not designed yet" beside the question count. The estimate is recomputed
each time a pass starts, finishes, or stops, and each forecast is stored with the time it was
made.

**The session.** The next decision session is set by the earliest needed time across every
project. It covers every question that is written at that moment, in any project. Geoff can
start it earlier at any time.

**Stale answers.** When a round finishes, the release's conductor checks each answer already
given for a later round against what the finished round changed. Only a question whose ground
changed is asked again, with his earlier answer shown.

## What a round does at an unanswered question

Geoff: there is no fixed answer, it takes judgment, and waiting by doing other work is the first
choice. The session running the release applies this order.

1. Start other work that doesn't rest on the answer: another pass in the round, or another round.
2. Build the part that comes out the same under every answer.
3. Take the recommended answer and build on it, within the cap below, and only where the work is
   cheap to undo. The guess is listed for Geoff at the next session.
4. Stop. The page moves the session to "Ready now".

The cap is Geoff's: "One medium to small pass would be OK as a miss. A large pass or multiple
passes in sequence would not." That becomes two rules. The planned minutes of all work now
resting on a guess, added up across every project, stay at 300 or under. No pass starts if it
rests on work that itself rests on a guess.

The conductor judges steps 2 and 3: which part is the same under every answer, and whether the
work is cheap to undo. The command enforces the rest. It decides what is free to start, and it
refuses any start that breaks either rule, naming the rule.

Step 1 works only if the plan records which passes rest on which question. The plan also
records which rounds can run side by side. Passes are cut along file lines so that more of them
are free. No pass edits a file that every pass shares. Follow-up items go through the command
into one file per item. In the prototype every pass edited `BACKLOG.md`, which chained each
pass to the one before it.

Some stops never take a guess, as now: spend past the usage limit, an irreversible or
destructive act, a secret or a step only Geoff can do, an `auth-data` question, and work beyond
the approved scope.

A ruling can change what a release is meant to end up as. When one does, that release's
unmerged work stops and the release is replanned as a whole. Approved rounds are never patched
with amendments.

## Sessions run themselves

The command launches, records, and ends every background session. Each working session gets its
own worktree, made by the command, and never runs in a repo's main checkout. The command's
record of who holds which worktree replaces the text-matching busy check, so two rounds can
build in one repo when their files don't overlap.

There are four kinds of session. The desk session is Opus, in Geoff's terminal, and loads the
desk skill. The other three run in the background and load the working skill:

- A conductor, one per release, on Sonnet. It asks the command to start passes, makes the two
  judgments in the waiting order, checks stale answers, and writes reports.
- A pass session, which plans and builds one pass as `pass-core` says.
- A design writer on Opus, which writes one round's design ahead of the building.

A conductor asks Opus when a call is hard. Opus asks Fable when it can't rank the options, or
when a wrong call would cost more than one medium pass. Both are one-off questions, never
standing sessions.

At most two passes build at once on the host, so parallel gate runs can't exhaust its memory.
The usage stop stays where Geoff set it. The page shows work that waits on either limit.

A systemd user timer runs `cc-desk check` every five minutes and once at login. It uses no
model. It handles each case below, and the page shows what it did.

| What it finds | What it does |
| --- | --- |
| A session blocked on a lost connection, with the link back | Wakes it |
| A session that died, or all of them after a restart | Relaunches it in the same worktree, from the state files |
| A pass session idle 15 minutes with its pass unfinished | Relaunches it the same way |
| A launch that never started | Ends it and launches again, once |
| A session stopped on a permission prompt, an expired sign-in, or a usage limit | "Needs you now", naming the session's round and the one thing to do |
| A stopped session whose reason it can't read | "Needs you now", with the session's last line |
| A session relaunched twice that stops again | "Needs you now" |
| A finished session, and its worktree once merged | Ends and removes them |
| Work that became free to start | Starts it |

The limit of two relaunches covers every cause, a death as much as a stall, and it applies to
each pass and each conductor. A relaunch carries on from the pass's branch and never starts a
second session for the same pass.

A pull request that stays inside its approved design, with the reviewer's accept and every
check passed, merges with a record and no page for Geoff (his ruling of 2026-11-17). Any other
pull request goes back to its pass session to fix, twice at most. After that, or when the
change reaches outside the approved design, it becomes a question under "Waiting for you".

## Technical shape

Everything for cc-desk lives in one repo, `cc-desk`: the command, the page, the check, the two
skills, the short guide, the tests, and this design. It replaces pieces now spread over the
`approval-queue` repo, the workstation repo (`desk.md`, the round references in `pass-core`,
`approval-packet`, `round-watch`, `claude-worktree-busy`), and `~/.local/state/desk/`. The
workstation takes the skills from the repo at a pinned tag, as it does Survey's.

- **State files.** One JSON file per release lists its rounds and passes: planned and actual
  minutes, state, questions, answers, and what rests on what. An append-only log holds rulings.
  Follow-up items are one file each. Designs and reports are Markdown. There is no database.
- **The command.** One Go program, `cc-desk`, is the only writer of the state files. It records
  events from sessions, computes the estimate and the session, launches and ends sessions, runs
  the check, and opens the desk. Every scheduling rule is enforced here, and a conductor only
  proposes.
- **The page.** A small SvelteKit app that draws a snapshot the command produces and redraws
  when the files change. It holds no scheduling logic.
- **The skills.** One for a desk session, one for a working session. Each stays under about 150
  lines. They replace the desk rules and the packet procedure.

Only a session that `cc-desk` opened in Geoff's terminal can record a ruling. The command
records that session's id when it opens the desk and refuses a ruling from any other session,
or from one with no terminal. The guard is there to stop a mistake by a background session. It
isn't built to stop a session that sets out to forge a ruling. The desk session copies his
words, and nothing checks them against the transcript. Every ruling shows on the page, so he
can read what was recorded in his name.

## What the prototype taught

Each row is a loss recorded in the desk's notes, and the part of this design that answers it.

| What happened | What cc-desk does |
| --- | --- |
| A lost connection stopped a round for 410 minutes until Geoff typed | The check wakes blocked sessions |
| A restart ended six sessions and nothing relaunched them for 3.5 hours | The check runs at login and relaunches from the files |
| A night's run stopped on two questions 25 minutes after Geoff left, idle 480 minutes in each round | The forecast, the desk's "what could come back" brief, and the waiting order |
| Every round was one chain, because every pass edits `BACKLOG.md` | No shared file, and passes cut along file lines |
| An approved stack lost its approval to a `BACKLOG.md` conflict | The same, and clean merges need no page |
| Recording one chat instruction cost a checked document, about ten tool calls | A one-line ruling |
| Words typed mid-turn couldn't be recorded | The desk session records from the conversation it is in |
| The desk woke on Fable for changes it then said nothing about | No standing desk; the check uses no model |
| The desk ran out of context about a dozen times a day | A desk session is short and holds nothing |
| The desk heard of a finished pass up to an hour late | Sessions report to the command, and the page redraws |
| Permission prompts, expired sign-ins, and usage limits stopped sessions and woke nobody | The check shows each as "Needs you now" |
| The busy check misread a stale server, an idle planner, a language server, and a conductor's folder | The command's record of who holds each worktree |
| An idle planner, a never-started session, and a design session waiting days for a start line | The command owns every start and every end |
| A design session thrashed on three long files | Two short page shapes, designs written per round |
| Throwaway builds vanished with the job folder | The command gives each design a durable folder |
| When the target moved, approved rounds were patched until Geoff stopped it | A changed target stops and replans the release |
| Merge pages for mechanics; a no-question design sent to the browser | Clean merges recorded; no page where nothing is asked |
| A node process reached 19.4 GB while gates ran beside idle sessions | Two building passes at most; finished sessions end |
| Half of every request was the same 60,000 tokens | Sessions launch with only the tools they use, if the third proof holds |
| Planning sat on the critical path, 176 minutes in one round | Designs and plans run ahead of the building |

Two notes fall outside cc-desk and are filed as follow-ups: the Proton Pass hook refusing plain
`grep` calls, and the missing `libatomic` in the `term-headless` container.

## Prove first

Seven things in this design rest on behavior nobody has checked. The first pass proves each in a
throwaway build before anything depends on it, and reports what it found. Each has a fallback.

1. A script can wake a background session blocked on `API unavailable`. If it can't, the check
   ends that session and relaunches it from the files.
2. A script can tell these states apart from what Claude Code records for a session: working,
   idle, blocked on a lost connection, and stopped on a permission prompt, an expired sign-in,
   or a usage limit. A state it can't tell shows as "Needs you now" with the session's last
   line.
3. A session launched with `--disallowedTools` no longer carries those tools' definitions. If
   it still does, sessions launch with every tool, at the token cost the prototype paid.
4. A launch can carry permissions that rule out a prompt for the edits and worktree steps a pass
   makes. If it can't, the launch carries the narrowest allow list that works, and a prompt
   that still appears shows as "Needs you now". No session runs with permission checks turned
   off unless Geoff rules that it may.
5. A new worktree under a repo Geoff has already trusted raises no trust prompt. If it does,
   worktrees go in one folder he trusts once, at setup.
6. A session can be relaunched after a restart from what the state files hold. If resuming
   fails, a fresh session starts from the pass's plan and branch.
7. The command can read how much of the usage allowance is left without a model. If it can't,
   each conductor reads it and reports it to the command.

## Moving cairn and dubplate over

Both efforts are reshaped for cc-desk as part of this work (Geoff). Each becomes a release file.

- **cairn.** The approved route (16 rounds, 45 passes, nine answers) becomes the release "the
  first public release". Its 61 written prompts and its start-lines file are replaced by the
  command's launcher. Each round gets the first read. Passes that chain only through
  `BACKLOG.md` are re-cut. Round 1 is holding for this.
- **dubplate.** The intake, structure, and catalog designs become rounds of one release. Its two
  design sessions that were left blocked are ended, and their drafts are kept.

Every answer Geoff has already given carries over, and no answered question is asked again. Each
reshaped release comes to him once, in one decision session, with only the new questions the
reshaping raises. The reshaping writes down each of cairn's nine answers beside the question it
settles, and the full list of rounds and passes after the re-cut.

## Retiring the old system

The old system keeps running until cc-desk passes its done list. Then one pass removes the
`approval-queue` command, the queue's service, the approval-queue reminder hook, the `desk`
launcher, and the replaced workstation files. The `approval-queue` repo is archived, and
`~/.local/state/desk/` is kept as a record that nothing reads.

## Limits

| Limit | Value |
| --- | --- |
| Work resting on guesses, all projects | 300 planned minutes, one pass deep |
| Building passes at once | 2 |
| Check interval | 5 minutes, and at login |
| Idle before a pass session counts as stalled | 15 minutes |
| Relaunches per pass or conductor, any cause | 2 |
| Fix attempts on a pull request that isn't clean | 2 |
| Designs kept ahead of the building | 2 rounds |
| Passes used for the pace ratio | the last 10 per project; under 3, the ratio is 1.0 |
| Page redraw after a state change | within 5 seconds |

A state file that fails to parse stops the command with the file's name and the error, and the
page shows the last good snapshot marked stale. The command never guesses at a broken file.

## Left to the builder

The file layout and field names, where the state files live, the port, how the page watches the
files, the subcommand names beyond `cc-desk`, `cc-desk <project> <round>`, and `cc-desk check`,
the split of the work into rounds and passes, and whether any code from the prototype is reused.
Geoff: reuse or start fresh, "whichever is most efficient. I'm just looking for a good, clean
outcome."

## What done looks like

1. **Estimate.** A test fixture holds two releases with known planned and actual minutes. The
   widget and the front page show the next session's time, range, question count, and projects
   equal to figures the test states, worked from the three steps in this document. Marking a
   fixture pass finished in half its planned minutes moves the time earlier by the stated
   amount, and marking one finished in double moves it later. A project with two finished
   passes shows "rough".
2. **Click-ahead.** In the fixture, every round and pass on the page links to a page that
   loads and shows that item's name, planned minutes, and the minutes of work before it, equal
   to the fixture's. A round with written questions shows them, a round with none shows
   "Nothing" and a reason, and an undesigned round shows "Not designed yet" with its forecast
   time. A test fails on any round or pass with no link.
3. **Reading surface.** A test walks every page and finds no form control except "Mark as
   read". After the walk has followed every link, the state files are byte for byte what they
   were. The Survey gate and an automated WCAG 2.2 AA scan pass in both themes, and a test
   finds words beside every state the page shows in color. The `standards:design-refinement`
   round is one commit range, the report lists its changes, and the scan passes after it.
4. **Ruling.** In a `cc-desk` session Geoff answers a waiting question in words he types. The
   rulings log gains one line with the time and his words, which he confirms on the page are
   his. The page drops the question within 5 seconds. The pass that waited starts with no
   other action from him, or shows as waiting on the two-pass limit when two are building. The
   same record attempted from a background session is refused, and so is one from a shell
   with the desk session's id and no terminal.
5. **Waiting order.** Fixture cases run against the real command, with two projects. With a
   free pass available, the free pass starts and the dependent one doesn't. With only a
   400-minute dependent pass, nothing starts and the widget reads "Ready now". With a
   200-minute dependent pass, it starts on the recommendation and the guess is recorded. A
   150-minute dependent pass in the other project is then refused, and the refusal names the
   300-minute rule. A 50-minute pass that rests on the guessed pass is refused, and the
   refusal names the one-pass rule.
6. **Recovery.** A conductor process is killed: it is running again within 10 minutes and the
   page shows no stop. A pass session is killed three times: after the second relaunch the page
   reads "Needs you now" with the round's name, no third relaunch happens, and at no point do
   two sessions hold that pass. Each relaunch carries on from the pass's branch with its
   commits kept. After a logout and login with work in flight, every session that was running
   is running again within 10 minutes. A session stopped on a permission prompt shows as
   "Needs you now" within 10 minutes, and a session cut off from the network carries on within
   10 minutes of the link's return.
7. **No tending.** Geoff approves one real design at the desk. Its first pass starts, builds,
   merges cleanly with no page shown, and its session and worktree are gone within 10 minutes
   of the merge. He typed no command but `cc-desk`, answered no prompt, and opened no other
   session.
8. **A real round.** cairn's first round runs under cc-desk from its decision session to its
   report. Geoff launches nothing, and the rulings log and the desk transcripts show no
   request from him for status. The report has the three parts. Its folded part gives each
   session's first forecast, with the time that forecast was made, beside the real time.
9. **Both efforts moved.** cairn and dubplate each show on the page as a release whose rounds
   and passes match the list the reshaping wrote down. Each of cairn's nine answers shows on
   the round it settles, and no question on the page repeats one.
10. **One repo.** `approval-queue`, `desk`, `round-watch`, and `claude-worktree-busy` no longer
    exist as commands. The queue's service and its reminder hook are gone, and the
    `approval-queue` repo is archived. The workstation's `claude-tooling-sync verify` passes,
    and nothing in the workstation repo names a removed file.

## Calls Claude made

These are borderline. Each can be overturned at approval.

- A design page reaches Geoff only when it asks him something or starts a new release. A round
  inside an approved release with no question of his starts without one.
- The session covers every written question, with designs kept two rounds ahead. No question is
  held back for a later session.
- Rulings are guarded by where they are recorded (a `cc-desk` terminal session) and by showing
  every one on the page. The prototype's search of the transcript for his words is dropped.
- The ruling guard stops mistakes, and a deliberate forgery by a session is out of its reach.
- No background session runs with permission checks turned off, even if that costs an
  occasional "Needs you now".
- The first estimate is simple arithmetic that ignores parallel work and usage pauses.
- The widget sits above "Waiting for you", and shows a time with a range.
- 300 planned minutes is the size of "one medium to small pass".
- cc-desk owns everything above a single pass: releases, rounds, scheduling, decisions, reports,
  and sessions. `pass-core` keeps the inside of a pass.
- Project designs and reports stay in the `initiatives` records. The design and plans for
  cc-desk itself are the only ones in its repo.
- The repo starts private.
- The command Geoff types is `cc-desk`, with no shorter alias.
