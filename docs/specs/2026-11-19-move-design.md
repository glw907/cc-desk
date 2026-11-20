# cc-desk Move: the round design

This is the sixth round of the cc-desk route. After it, cairn and dubplate each show on the
cc-desk page with their rounds and passes. cairn's first round, `checks`, then runs under
cc-desk from its desk session to its report, and Geoff launches nothing.

The main design is [2026-11-17-cc-desk-design.md](2026-11-17-cc-desk-design.md). This round closes
its done items 8 and 9. The rounds it builds on are in
[2026-11-18-desk-skills-design.md](2026-11-18-desk-skills-design.md) and the designs that file
names. Where this file is silent, those hold.

**What this asks of you.** Two things, once you have run your sitting for the Desk round:

1. Approve this design and the round plan beside it
   ([2026-11-19-move-round.md](../plans/2026-11-19-move-round.md)). That approves five passes
   planned at 535 minutes of session time, the auto mode entry in the plan, and the cleanup of
   cairn's leftover pull request, branches, and worktree in step 2 of "The run".
2. Answer the five questions under "Questions for Geoff" at the end. Each has a recommendation,
   and this design is written on those recommendations. To change anything, name it before you
   approve.

Approving starts nothing at once. The Move conductor launches when the round plan's start line
reads true: pass 6 has landed, your sitting passed, and GitHub Actions runs again.

A few names this design uses. Pass 6 is the Desk round's last pass, `desk-run`, which builds the
real release you run in your sitting: five steps at the terminal that end with one question
waiting for cc-desk on the page. 1.1 is cc-desk's next release, which works through the lessons
ledger and may build beside this round. The check is cc-desk's timed run that starts passes and
merges pull requests, and a slot is one of the two passes it lets build at once.

## Why this round

cc-desk now runs its own work: the check starts free passes, merges a clean pull request, and
starts a conductor when a report is due. It has run only cc-desk's own scratch releases. Two
efforts still sit in the old process. cairn's route has 16 rounds and 45 passes, designed and
ruled on, and every round waits for the revised desk (ruling `4135d8`). dubplate's next rounds
have two designs that stopped mid-question when the desk was retired, and their drafts sit in
`~/.local/state/desk/kept-drafts/`, a folder the Retire round removes.

The route ([2026-11-17-route.md](../plans/2026-11-17-route.md)) planned Move as a cairn release
file at 75 minutes, a dubplate release file at 45, and cairn's round 1 for as long as it runs. It
planned one decision session for each reshaped release. Under cc-desk that session is the desk,
which Geoff opens once the releases are added and their questions are waiting. Two earlier rounds
also sent work here: the Desk round moved design writers that start by the timer, and the merge
of a design writer's pull request, to this round.

## Goals

- cairn and dubplate each show on the page as a release whose rounds and passes match their
  reshaping, with every earlier answer on the round it settles.
- cairn's round 1 builds, merges, and reports with no launch and no status request from Geoff.
- A pass session in either repo finds its round's design and its plan brief by a path its
  launch prompt names.
- The timed check starts a design writer for a round that needs one, and merges the writer's
  pull request the way it merges a pass's.
- dubplate's stopped drafts have a home in the dubplate repo before Retire removes the folder.

## What it leaves out

- cairn's rounds 2 to 16 and dubplate's rounds. They run under cc-desk after this round, and
  their reports go to Geoff as any round's do.
- The 1.1 round's work. 1.1 answers the lessons ledger and runs beside this round (standing
  answer 37). Where a point below belongs to 1.1, its line says so.
- The overrun budget as a release field. cairn's ruling 9 (the ceiling is twice the plan) has no
  mechanism in cc-desk. The Move conductor watches round 1 (standing answer 45), and the
  conductor of each later cairn round watches its own until 1.1 builds one.
- The Retire round's work. The cairn and desk drafts in `kept-drafts/818e389e/` stay where they
  are, and the Retire round's encrypted archive keeps them with the rest of `~/.local/state/desk/`.

## What would make it a failure

- Geoff is asked for anything during cairn's round 1 beyond the questions the release carries.
- A pass session can't find its design or its brief and asks, guesses, or stops.
- A cairn or dubplate pass stops at a permission prompt that the release's rules should have
  allowed.
- Three passes build at once on the workstation because a pass run outside cc-desk took a slot
  the check didn't know about.
- A Move pass edits a file that a 1.1 pass in flight also edits.

## Approaches

The main question was where each release's papers live. Three shapes were weighed.

1. **Each repo holds its own papers (chosen).** A one-time pass copies cairn's route and its 45
   planner prompts into `docs/internal/plans/route/` in cairn-cms, and dubplate's designs and
   prompts into dubplate's `docs/specs/` and `docs/plans/`. Each round's `design` field names a
   path inside the checkout. The page's design view already refuses a path outside the checkout
   (`web/src/lib/server/design.ts:60`), so this is the only shape the page can show. dubplate's
   designs and prompts are also untracked files in the initiatives repo today, and this commits
   them.
2. **The rounds point at the initiatives repo.** No copy, but the page refuses every design, and
   a pass session reads files from a repo it doesn't work in. dubplate's files there aren't even
   committed.
3. **The launch prompt carries the design text.** It avoids a path, but a 1,300-word brief and a
   design several thousand words long don't belong in a command line, and nothing records which
   text a session was given.

The second question was how Move's own passes run. They run the old way, under a Move conductor,
as the Desk round's did. cc-desk can't run them as a release: `AddRelease` refuses a second live
release for one project, and 1.1 shares the project `cc-desk`. The cost is that the check can't
count a Move pass against its two building slots, which "The run" handles by order.

## Where each release's papers live

`move-cairn-docs` opens one pull request on `glw907/cairn-cms`. It copies from
`~/Projects/initiatives/theme-garden/` at commit `f461b3f`:

- `specs/2026-11-17-route-rounds.md` to `docs/internal/plans/route/route-rounds.md`.
- The 45 pass planner prompts (`plans/2026-11-17-route-<pass>-planner-prompt.txt`) to
  `docs/internal/plans/route/prompts/<pass>.txt`, one per pass of rounds 1 to 16.
- A new file, `docs/internal/plans/route/README.md`, the regime map. Its content is the next
  section, written for a pass session.

The same pull request does two cleanups the checks conductor left. It copies the `foundation-7`
plan, `2026-11-17-foundation-7-routing-plan.md`, an untracked file in that worktree, to
`docs/internal/plans/`, and it replaces the foundation block in cairn's `docs/STATUS.md`
with a pointer to the cc-desk release. `docs/internal/` is outside cairn's published-docs gates
(`scripts/checks/check-stable-source.mjs:83`). The pass runs cairn's gate and lists anything it
flags.

`move-dubplate-docs` opens one pull request on `glw907/dubplate`:

- From `~/Projects/initiatives/dubplate-server/specs/`: the structure design, the intake design,
  and the catalog architecture design, into `docs/specs/` under their names.
- From `plans/`: the five intake planner prompts and the intake conductor prompt, and the
  `2026-11-17-structure/` folder without its `zz_*_test.go` probe, into `docs/plans/`.
- From `~/.local/state/desk/kept-drafts/`: folders `93d87e73` and `b3c9cad7` into
  `docs/archive/kept-drafts/`, without `93d87e73/probe/dubplate-server` (a 23 MB binary the
  repo builds). Also `818e389e/dubplate-server-data.queue.md`, the one dubplate draft in the
  cairn folder. A `README.md` there says these are unsent drafts of the stopped design sessions,
  not rulings.

The two kept design drafts are byte-identical to the initiatives copies (checked 2026-11-20), so
`docs/specs/` holds the one copy that counts and the archive keeps the drafts as they were left.
The pass edits nothing it copies.

## What binds a cairn pass

The regime map tells a cairn pass session which parts of the old papers still bind. Its rules:

- The route's round sections, pass tables, acceptance runs, and assumptions bind, with Geoff's
  rulings 1 to 9 (`a03b21`).
- A pass's prompt file is `prompts/<pass>.txt` beside the map, and it binds in its "For this
  pass" and "Rules the plan carries" paragraphs. Its
  hand-over, its executor launch, its approval-packet step, and every `decide` page are void.
  The pass plans and builds in one session and reports by `cc-desk-work`.
- Standing answers 15 to 19, 23, and 24 hold as written.
- Answers 20, 25, and 26 are void: cc-desk cuts the worktrees, `pass ready` and the check replace
  the merge document, and the check and the reports replace the round-end conductor.
- An `auth-data` pass's merge document is void. It merges on its reviews and checks like any
  other pass inside its design.
- Answers 21 and 22 become questions on the release: a step only Geoff can take is a question
  its pass rests on.
- Answer 13 (usage) is the global usage stop. Answer 14 (overrun) has no mechanism, and each
  round's conductor watches it until 1.1.
- Ruling 7 (rounds run one at a time) is the chain of `after` edges in the release, so the check
  holds a round until the one before it is done.

## The two releases

`move-releases` writes two input files for `cc-desk release add` and a test that loads them.

**`docs/releases/cairn-cms.json`.** Project `cairn-cms`, release name `theme garden`, repo
`glw907/cairn-cms`. It holds:

- The 16 rounds `checks` to `release`, each with `design_written` true and `design` set to
  `docs/internal/plans/route/README.md`. Its one `context` paragraph names the round's section
  in `route-rounds.md` and the folder of its pass prompts.
- Each round's `after` edges: the round before it in route order, which is ruling 7, plus its
  own builds-on rounds from the route table.
- The 45 passes, with the route's minutes and depends-on, model and effort from the route's
  class (Opus at high for `runner` and `auth-data`, Opus at medium otherwise).
- The nine answers as carried questions, each on the round the route says it settles and resting
  on that round's first pass. Ruling 8 becomes the global usage stop (Question 2). Ruling 9 is
  carried with a note that the conductor watches it.
- The steps only Geoff can take, as unanswered questions: the terminal screens (`access-3`),
  the Cloudflare sign-in and the npm trusted publisher (`setup-3`), the test site and 907.life
  (the `owner` leg, on the pass the route names), and "cut" (`release-3`).
- Allow rules for what cairn's gate and lanes run, from cairn's `CLAUDE.md`, its
  `package.json` scripts, and the denials log at `~/.local/state/claude-denials.jsonl` for rows whose `cwd` is under
  `~/Projects/cairn-cms`.

Round 1, `checks`, rests on no unanswered question. Its `no_need_reason` cites `a03b21`.

**`docs/releases/dubplate.json`.** Project `dubplate`, release name `0.8`, repo `glw907/dubplate`.
Its rounds are the "Rounds to 0.8" section of `BACKLOG.md` from round 3, with the round the
structure design adds, in a chain:

| Round | Design | Passes |
| --- | --- | --- |
| `structure` | `docs/specs/2026-11-17-structure-design.md` | `structure-base`, `structure-keys`, `structure-doctor`, from its table |
| `intake` | `docs/specs/2026-11-17-intake-round-design.md` | The five passes `intake-ground` to `intake-look`, one round, as its question 9 recommends |
| `catalog` | Not written. `context` names the catalog architecture design | None until its writer's design merges |
| `playback`, `queue`, `library`, `launch` | Not written | None |

The order of the first two is Question 4; the table shows the recommendation. BACKLOG prices no
round, so each unwritten round carries 600 planned minutes until its writer's design replaces them.
The structure passes carry 390, 360, and 340 path minutes, and intake's five carry 2,610.

The questions the two designs asked that cc-desk now settles are dropped, each with its reason
written as a call (`calls`) on the round whose design asked it. Structure's 2 and intake's 9 asked round size: under cc-desk
Geoff reads a design only when it asks something. Structure's 3 asked the merge rule: the check
merges a change inside its design. Intake's 10 asked the same for passes that touch sign-in and
data, which Geoff's merge ruling of 2026-11-17 settles ("Calls Claude made"). Structure's 4 and intake's 12 asked the usage
stop, now global. Structure's 5 asked overrun, which the conductor watches. The order question,
structure's 1 and intake's 13, is Question 4. The product questions go in unanswered for the
desk, each resting on the pass its design names: intake's 1 to 8 and 11, and structure's 6.
Catalog's five come later, in its writer's design, so none sits on a round with no pass.

Allow rules come from dubplate's `CLAUDE.md`, `scripts/check.sh`, the `Makefile`, and the denials
log rows under `~/Projects/dubplate`.

**The test.** `e2e/release_inputs_test.go` runs `release add` for each file against a temporary
state folder through `--state-dir`, then reads the snapshot. It checks that every question has a
pass (so no question is an orphan, `orphan-question-forecast.md`), that every key in the file is
a known field with its exact case (the gap in `strict-reader-key-case.md`, checked by a token scan
in the test), and that cairn has 16 rounds and 45 passes. The conductor checks the rest live:
each `design` path exists on the repo's `origin/main` before it adds the release.

## Design writers on the timer

Today a writer starts only through `cc-desk session start design <project> <round>`, it marks its
round written with `round design` before its pull request merges, and nothing merges that pull
request. `move-writers` changes all three.

**When a writer starts.** A new step in the check's plan starts a writer for a live release that
records a checkout. It takes the first round in `after` order whose design isn't written and that
has no writer. It starts one only when that round is at most two rounds ahead of the earliest
round with a pass not done, matching the main design's "designs kept two rounds ahead". A release
has at most one writer at a time. A writer counts as one of the two building slots, and the usage
stop holds it as it holds a pass.

**What a writer finishes with.** The writer commits its design, its pass plans, and a design file
for `round design`'s format at `docs/designs/<round>.json`. It then runs a new verb, `cc-desk round
ready <project> <round> --pr <n> --commit <sha> --verdict <v> --design-file <path>`. The verb
reads the design file at that commit and refuses one that `DecodeDesign` refuses. A design that
can't load never merges. The writer's questions travel in the design file, and `DesignRound`
already adds them. The `question add` line leaves the writer's prompt: `AddQuestion` needs passes
the round doesn't have yet (`internal/state/question.go:78`).

**The merge.** A round gains a pull request record like a pass's. The check's merge table reads
it the same way: checks passed, the verdict, no open finding, and no call for Geoff. When the
pull request is on `main`, the check reads the design file from the merged commit and applies
`DesignRound`, and the round's passes become free on the usual rules. `round design` stays for a
hand fix, and the desk skill says so.

**What it folds.** The same pass takes three follow-ups that sit in the lines it rewrites:

- `pull-request-closed-without-merge.md`. A pull request closed without a merge stops its pass
  or round with the reason `pull_request_closed` (`internal/sessions/merge.go:125`).
- `conductor-worktree-removal-gaps.md`, its last paragraph. A failed merge clears `MergeCalled`
  (`cmd/cc-desk/merge.go:88`).
- `designed-round-with-no-questions-reads-nothing.md`. `DesignRound` refuses a design for a round
  marked `needs_geoff` that brings no question and has none.

**Where it is proven.** cairn's rounds all arrive designed, so cairn doesn't prove it. The pass's
live run proves it in the scratch repo, and dubplate's `catalog` round is its first real use.

## The other points this round settles

**One slot for each project.** `startFree` walks releases in name order, so `cairn-cms` takes both
building slots whenever it has two free passes. `move-writers` changes the loop: while two
releases each have a free pass, each gets at most one slot. This is Question 3. A pass resumed
after a usage stop skips the limits (`resume-skips-the-limits.md`), so a resume can put two cairn
passes building for one cycle. That is accepted and stays filed for 1.1.

**The launch prompt names the design.** `move-prompt` adds the round's `design` path to
`pass.txt` when the round records one, in place of the line "the round's design is in the repo
under docs/". The round's `context` paragraphs follow it. They are the paragraphs a desk reads
before it asks (`internal/state/types.go:236`), and for cairn each one names the round's section
of the route. A cairn pass then reads the regime map, which says where its prompt file is.

**Earlier answers after a round** (`earlier-answers-after-a-round.md`). `move-prompt` adds one
step to the conductor's prompt for a report. Before it records the report, the conductor reads
each answered question of the release's later rounds. Where the round's findings undercut an
answer, it writes a new question with `question add` that quotes the earlier ruling and says what
changed. The follow-up is then done.

**`session end` belongs to the desk.** Pass 6 adds `cc-desk session end pass <project> <name>`
and `session end conductor <project>`. The
desk skill offers it when Geoff asks to end a stuck session, and the page shows the command in
words, as it shows every other. Main design done item 3 keeps the page free of controls, so the
page gets no button. Pass 6's rewording of `leftover-live-session-parks-a-pass.md` points here.

**The follow-ups the Desk round sent here.** Each has a verdict under "Backlog" below. None of
them blocks this round's outcome.

## The passes

Five passes, in two lines of work that never share a file. The measured Desk passes ran 906
minutes against 630 planned, a ratio of 1.44, and the figures below already carry it.

| Pass | Outcome | Repo | Main design items | Minutes |
| --- | --- | --- | --- | --- |
| 1. `move-prompt` | The pass prompt names the round's design and context, and the report conductor checks earlier answers | cc-desk | 8 | 60 |
| 2. `move-releases` | The two release input files and the test that loads them | cc-desk | 9 | 120 |
| 3. `move-cairn-docs` | The route, its prompts, and the regime map are in cairn-cms | cairn-cms | 8, 9 | 60 |
| 4. `move-dubplate-docs` | dubplate's designs, prompts, and kept drafts are in dubplate | dubplate | 9 | 45 |
| 5. `move-writers` | Writers start on the timer, their pull requests merge through the check, and each project gets a slot | cc-desk | none directly | 250 |

`move-prompt` and `move-writers` both change the launcher, so the writers follow the prompt. The
other three share nothing and run in the second slot. The longest line is `move-prompt` then
`move-writers`, at 310 planned minutes.

`move-writers` is priced from the Desk round's repairs pass, planned at 175 and taken at about
425, with its live run. Its base of 175 at the 1.44 ratio is 250, the three folds included. The
route's 75 and 45 for the release files become 120 for the two together with their test. The
two docs passes copy files and run a gate.

## The run

The Move conductor runs these steps after the passes it waits on have landed. The round plan
gives each one's check.

Its first act, before any pass, lands cairn's `tool-delete` pull request (#260) on Question 1.
That pull request edits cairn's `docs/STATUS.md` beside the block `move-cairn-docs` replaces, so
the docs pass waits for it. If you answer No, `move-cairn-docs` leaves cairn's STATUS alone, and
the report lists the foundation block as stale.

1. When passes 2, 3, and 4 have landed, it runs `cc-desk release add` for both files with no checkout. Nothing
   starts: the check skips a release with no checkout (`internal/sessions/decide.go:321`). Both
   releases and their questions now show on the page, which is done item 9.
2. It closes cairn pull request #259 and deletes branch `foundation-ci-edits`. It removes the
   `foundation-7` worktree and its branch, now that pass 3 has kept the plan. It lists the stash
   `foundation3-task3-wip` and the 110 `agent-*` worktrees in the round report for Geoff, and
   removes neither.
3. When pass 5 has landed and is installed, it writes the usage figure to `settings.json`
   (Question 2) and runs `cc-desk release checkout` for both projects. From the next check,
   cairn's `checks-1` starts. dubplate's passes rest on questions, so they wait for the desk.
4. It watches round 1 to its report, reading `cc-desk snapshot` at each wait. It starts and
   ends no cairn session itself, outside the round plan's standing answers. The check starts a
   conductor for cairn when round 1's report is due, and that conductor writes the report.
5. It writes the Move report, with what Geoff's desk session settled and the evidence for done
   item 8, and ends.

The order keeps three passes from building at once. While a Move pass builds, neither release has
a checkout, so the check starts nothing for them.

If a 1.1 pass in flight holds a file `move-writers` needs when it would start, the conductor
doesn't wait. It runs step 3 without the writers, and `move-writers` goes to 1.1 with its plan,
written under "Open decisions" in `docs/STATUS.md`. dubplate's `catalog` round then waits for a
writer started by hand from the desk.

## Steps only Geoff can take

Before the round starts: approve, which approves the auto mode entry in the round plan and the
cleanup in step 2 of "The run". Claude adds the entry after the approval.

Once the releases are added, one desk session, separate from your Desk round sitting. The page shows the questions waiting for cairn and
dubplate. Type `cc-desk` and answer them in your words. cairn's round 1 doesn't wait for this
sitting.

Later, as cairn's rounds reach them, the steps only you can take come to the desk as questions:
the terminal screens, the Cloudflare sign-in, the npm trusted publisher, the test site and
907.life, and "cut".

## What done looks like

1. **The prompt.** A pass launched from a fixture release whose round names a design and two
   context paragraphs has the path and both paragraphs in its prompt. The conductor template for a report holds
   the earlier-answers step. `bash scripts/gate.sh` passes.
2. **The release files.** `e2e/release_inputs_test.go` loads both files into a temporary state
   folder with no error. The snapshot shows cairn with 16 rounds and 45 passes whose names,
   minutes, and depends-on match `route-rounds.md`, and dubplate with the rounds in the table
   above. Every question rests on a pass. No key in either file differs in case from its field.
3. **The papers.** cairn's `main` holds `docs/internal/plans/route/` with the route, 45 prompt
   files, and the regime map, and cairn's gate passed on the pull request. dubplate's `main`
   holds the three designs, the prompts, and `docs/archive/kept-drafts/` with no binary.
4. **The writers.** In a scratch release, the check starts a writer for an undesigned round with
   no command from anyone. The writer's pull request merges by the check, and the round's
   passes appear only after the merge. A design file that doesn't decode is refused at
   `round ready`, and a closed pull request stops its record with `pull_request_closed`. With
   two scratch releases each holding two free passes, each gets one building slot.
5. **Both efforts moved** (main item 9). The page shows cairn and dubplate as releases matching
   their reshaping. Each of cairn's nine answers shows on the round it settles, and no question on
   the page repeats one.
6. **A real round** (main item 8). cairn's round 1 runs from Geoff's approval to its report.
   Geoff launches nothing, and the rulings log and the desk transcripts show no request from him
   for status. The report has the three parts, and its folded part gives each session's first
   forecast, with the time it was made, beside the real time.
7. **The gate.** Every cc-desk pass passes `bash scripts/gate.sh` with no skipped test. Item 4's
   live run prints one line for each check, and its pull request holds that output from a run
   on its last commit.

## Backlog

A scoped sweep read the follow-ups at `d0805fb` against this round, on 2026-11-20. It left out
the ledger, `scheduled-check-ins.md` (a 1.2 feature), and the four files pass 6 changes.

| File | Verdict |
| --- | --- |
| `pull-request-closed-without-merge.md` | Fold into `move-writers` |
| `conductor-worktree-removal-gaps.md` | Fold its last paragraph into `move-writers`. Gaps 1 to 3 stay |
| `designed-round-with-no-questions-reads-nothing.md` | Fold into `move-writers` |
| `earlier-answers-after-a-round.md` | Fold into `move-prompt` |
| `orphan-question-forecast.md`, `strict-reader-key-case.md` | Keep. `move-releases`' test keeps both from biting the two release files |
| `resume-skips-the-limits.md` | Keep for 1.1. A resume can break the new slot rule for one cycle |
| `cleanup-of-a-release-marked-done.md` | Reword: pass 6's `release done` makes it reachable, and that verb is where it is decided |
| `launch-review-notes.md` | Reword item 5: the fallback to `Make` now covers only conductors, and it should cover pass and design records too |
| `2026-11-17-rerun-prove-proofs.md` | Reword: item 6 is done, and item 3 points at `live-sample-of-limit-texts.md` |
| `live-sample-of-limit-texts.md`, `cold-boot.md` | Keep. The real timer is the chance to see them; the conductor keeps any capture |
| `reading-errors-and-guess-on-resume.md` | Keep. A failed usage read holds every start for one cycle |
| `api-project-state-round.md`, `building-pass-without-checkout-needs-entry.md`, `desk-start-time-after-a-clock-step.md`, `input-before-lock.md`, `orphan-ruling-after-a-crash.md`, `round-estimates-priced-from-evidence.md` | Keep. None touches this round's files or outcome |

The three rewords land with this design. A fold is removed by the pass named.

## Pass 6's state, as read

Read on 2026-11-20 at 09:40 UTC, just before this design was pushed to `main`. `origin/main`
was at `f265ad6`, with nothing of pass 6 on it. Pass 6's branch, `desk-run`, was at `f2b0c18`
with seven commits, as draft pull request 22. Its executor, `cc-desk-desk-run`, and its
conductor were busy. The branch already has `cc-desk release done <project>` and the shared live
library `scripts/live/lib.sh`. Its plan's Task 2 still has to build `cc-desk session end pass
<project> <name>` and `session end conductor <project>`, which this design relies on.

GitHub Actions has run no job on glw907/cc-desk since 09:12 UTC that day, because of a billing
block that only you can clear (`docs/STATUS.md`, "Open decisions"). Pass 6 builds on the local
gate until then. Every Move pass and every cairn and dubplate merge needs CI, so the round plan's
start line waits for Actions to run again.

No Move pass races pass 6 for a file: every Move pass starts after your sitting, and your
sitting follows pass 6's landing. `move-writers` changes `internal/sessions/decide.go` and
`internal/state/` after pass 6 has changed them, starting from `main`.

## Calls Claude made

Each call can be overturned at approval.

- cairn's papers are copied into cairn-cms and dubplate's into dubplate, so each design path sits
  inside its checkout where the page can show it.
- Move's passes run under a Move conductor in the old way, since cc-desk can't hold a second live
  cc-desk release beside 1.1.
- Design writers are built here, not in 1.1, because the Desk design placed them here and
  dubplate's `catalog` needs one. They are built last, off the path to cairn's round 1, and go to
  1.1 if a 1.1 pass holds their files.
- No release gets a checkout until `move-writers` lands, so no pass the check starts shares the
  host with a Move pass.
- A writer counts as a building slot.
- The release input files live in the cc-desk repo, so they are reviewed and tested before they
  are added.
- A pass that touches sign-in or data code merges like any other change inside its design, under
  Geoff's ruling of 2026-11-17 (main design, "Sessions run themselves"). The old merge document
  for an `auth-data` pass is void, and the regime map says so. A question in such a pass still
  stops, as does a closing review short of accept on an `auth-data` finding (`advisor.md`).
- cairn's project name is `cairn-cms`, matching its repo and its old session names.
- dubplate's intake is one round of five passes, the design's recommendation for its
  question 9.
- The cairn and desk drafts in `kept-drafts/818e389e/` stay for the Retire round's archive.

## Questions for Geoff

**1. Land cairn's `tool-delete` pull request?** Pull request #260 removes the Go `cairn` tool and
the `create-cairn-site` launcher from cairn's `main`, with their workflows, checks, and doc lines.
cairn's round 1 builds on it. It is open at `1fa506cc2` with an accept from its review and its
security read. Its session has waited since 2026-11-17 on a document in the old approval queue,
which this round retires. It comes first because it edits cairn's `docs/STATUS.md` beside the
block the cairn docs pass replaces.

- **Yes** (recommended). The conductor withdraws queue document `743a55`, stops the waiting
  session, and lands #260 by `approval-queue land`. Round 1 can start once `move-writers` lands.
- **No.** cairn's release is added with no checkout, and round 1 waits for your word.

**2. Which usage figure stops new work?** cc-desk has one figure for every release: past it, the
check starts no new session. Your ruling 8 for cairn says 85 percent. cc-desk's own rounds have
run to 100. The week stood at 15 percent on 2026-11-20.

- **85 percent for everything** (recommended). It keeps your latest ruling. cc-desk's own work,
  1.1 included, also stops starting at 85.
- **100 percent for everything.** cairn runs past its ruling.
- **A figure for each release.** That is a new field, and the work goes to 1.1.

**3. How do cairn and dubplate share the two building slots?**

- **One each while both have free work** (recommended). Neither starves the other.
- **First release in name order.** cairn takes both slots whenever it has two free passes, and
  dubplate waits.

**4. Which dubplate round runs first?** Both designs are written. The structure design
recommends itself first, since it moves the store and the settings that intake then builds on.

- **Structure first** (recommended).
- **Intake first.** The uploader page, where you add music to dubplate, comes sooner, and
  structure's moves come after it.

**5. Approve the round plan's proposed standing answers 40 to 46?** They cover the stops this
round can foresee that no earlier answer covers. The round plan gives each in full.

- 40: a 1.1 pass holds a file a Move pass needs. The Move pass waits; `move-writers` goes to 1.1
  rather than hold up cairn's round 1.
- 41: auto mode refuses a listed step. The conductor tries once more, then writes it under "Open
  decisions" and goes on.
- 42: a cairn or dubplate session stops at a permission prompt. The conductor adds the rule if
  the pass's steps need it and restarts the session; otherwise the pass parks for you.
- 43: the busy check names a 1.1 session. It's no stop unless the two passes share a file.
- 44: the first step finds you were asked for more than your sitting's five steps. The
  conductor launches nothing and writes what asked you.
- 45: cairn's round 1 runs past twice its plan, your ruling 9. The conductor writes the overrun
  for you and stops nothing.
- 46: CI stops running mid-round. Nothing lands without its checks; every other step goes on.

- **Yes, as written** (recommended). The conductor acts on them and names each use in the report.
- **No.** Each stop waits for you under "Open decisions".
