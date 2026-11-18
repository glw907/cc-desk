# The item pages: pass plan

Initiative: cc-desk
Round: page
Pass: page-items

Class `ordinary`: the pass adds routes, components, and tests under `web/` and changes no gate,
workflow, or Go file. Planned 75 minutes. The path below is 85, under the 150-minute park line in
the launch prompt and the round's 180. Recent passes took about 0.6 of their planned minutes, so
it is likely to run near 50.

Design: `docs/specs/2026-11-17-page-design.md`, the paragraphs on a round page, a pass page, and
the rulings page; the address paragraph (`/<project>/<round>` and `/<project>/<round>/<pass>`);
"Times", "Empty fields", and "State words"; and the not-found row of "When something goes wrong".
Its "Backlog" verdict on `docs/follow-ups/snapshot-stop-times.md` holds: a stopped pass shows its
reason and no "since", and a round whose needs are `nothing` with a null reason shows "Nothing"
with "No reason recorded". The command shows as text; the terminal link beside it is pass 4. Round
plan: `docs/plans/2026-11-17-page-round.md`. Where the round design is silent,
`docs/specs/2026-11-17-cc-desk-design.md` holds. Writing a person reads (the PR body, STATUS
lines, follow-ups) follows `~/.claude/docs/voice/written-for-people.md`, loaded through the
`writing-voice` skill. Labels on the page follow Survey's writing standard, through the `survey`
skill (round standing answer 19).

Worktree `/var/home/glw907/Projects/cc-desk/.claude/worktrees/page-items`, branch `page-items`,
cut from `origin/page-front` at `893ceab` (pull request 7, open). PR base: `main`. Push:
`glw907/cc-desk`. Before the PR is ready, merge `origin/page-front` in again if it moved, and
`origin/main` if main moved. The Sessions round runs beside this one. This pass edits no Go file,
nothing under `e2e/`, `internal/`, or `cmd/`, and neither `scripts/gate.sh` nor the workflow, and
it messages none of that round's sessions.

## Pre-flight

No step in this pass needs Geoff. Every write is inside the worktree, a test's temporary folder,
or `$CLAUDE_JOB_DIR/tmp`.

- Permissions checked: `~/.local/state/claude-denials.jsonl`, read again on 2026-11-18 for this
  pass over 2026-10-19 to 2026-11-18. 130 rows, 2026-11-01T00:55:05Z to 2026-11-18T06:27:23Z, the
  same rows as pass 1's read. Classes and rows: Classifier unavailable 47; Destructive or Unwanted
  Actions 14; Self-Modification 11; Interfere With Workloads 11; server-side classifier, no reason
  10; Undermining Oversight 5; Unauthorized Access 4; Merge Without Review 3; Auto-Mode Bypass 3;
  Irreversible Local Destruction 2; Git Destructive 2; Create Public Surface 2; Unverifiable
  Deletion Scope 2; Sensitive-Source Provenance 2; Security Weaken 2; Production Deploy 1;
  Secret-Store Writes 1; DNS/Domain/Cert Changes 1; Bot Check Circumvention 1; Code from External
  1; Shared Scratch Sweep 1; Safety Bypass Flag 1; Out-of-Place Publication 1; PII Data Handling
  1; Modify Shared Resources 1. This pass runs `npm ci`, `npm run check`, `npm run build`, `npx
  playwright test`, `bash scripts/gate.sh`, `git push origin page-items`, `gh pr create`, `gh pr
  ready`, `gh pr checks`, and `gh run list`. A search of the rows' targets for `npm`, `npx`,
  `node`, `gate.sh`, `gh run`, and `git push` found four near ones: three classifier outages and
  one `npm ci` chained with other installs and piped to `tail` (Destructive or Unwanted Actions,
  2026-11-14). No row refuses any of this pass's commands run bare. The table below holds the
  alternatives.
- Reviewed: no review before code, as the `ordinary` class sets (`pass-core` "Plan review
  lens"). The close gets one whole-branch `standards:diff-reviewer` read.
- Evidence: the mechanism table and the recorded searches below. Both acceptance checks were run
  on today's code before any code was written: `bash scripts/gate.sh` and `npx playwright test
  tests/items.spec.ts`. Their outputs, with the built-app probe, are under `## Acceptance`. One
  mechanism is quoted but not probed, `getByLabel` on a `dd`; standing answer 44 holds its
  fallback.

### Refused commands and their alternatives

| Refused in the log | This pass does instead |
| --- | --- |
| `npm ci` chained with other installs and piped to `tail` (2026-11-14) | Run `npm ci` alone, bare, from `web/` |
| `node` one-liners and inline scripts during a classifier outage | A script file in `web/tests/`, run by path |
| `kill`, `pkill`, and `kill -9` on test processes (Interfere With Workloads) | The Playwright fixture stops each server it starts. A background command stops with `TaskStop` |
| `sleep 90` in the foreground (2026-11-18) | Wait with the `Monitor` tool, never a foreground `sleep` |
| A settings edit that adds the Playwright download hosts (Security Weaken, 2026-11-08) | No settings edit. A missing browser parks, as pass 1's standing answer 22 says |
| `rm -rf` sweeps of shared temporary folders (Shared Scratch Sweep, Unverifiable Deletion Scope) | Tests make each folder with `mkdtemp` and remove only that folder |
| `git stash drop` (2026-11-17) | No stash. A temporary commit sets work aside |
| `gh pr merge`, and `gh pr view` used to judge a merge (Merge Without Review) | The pass merges nothing. It reads checks with `gh pr checks <n>` and `gh run list --branch page-items` |
| A classifier outage, 47 rows | Wait 60 seconds through `Monitor`, then run the command once more |

## Mechanism table

| Mechanism | Source or probe, with what it printed |
| --- | --- |
| An item's address is `/<project>/<round>` or `/<project>/<round>/<pass>` | Page design, "The app": "An item's address is `/<project>/<round>` or `/<project>/<round>/<pass>`, so two projects can each have a round of one name." `ReleaseBoard.svelte:44` already links rounds as `/{release.project}/{round.name}`; `Sidebar.svelte` and `ReleaseBlock.svelte` link `/{project}/{round}` |
| A pass's name is unique within its release, so the pass page finds it by project and name and checks its round | `internal/snapshot/snapshot.go:345-373` builds each pass from its release; `ReleaseBoard.svelte:57` keys a row as `data-pass="{project}/{pass}"` with no round |
| Today the built app answers an item address with 404 and the shell | A probe ran `node build` from `web/` on port 7999 with the twin state below: `/` printed 200; `/alpha/r1`, `/a/b/c/d`, and `/rulings` printed 404, each with the app's `<body class="font-body bg-base-100 text-base-content">` |
| With `ssr = false`, no `load` can throw a 404 for a missing item, because the data arrives later from the poll | `web/src/routes/+layout.ts` is `export const ssr = false;`. `poll.svelte.ts` fetches `/api/state` every 2 seconds after mount. So a missing project, round, or pass is drawn as a not-found view inside the page once a snapshot is in, and an address that matches no route falls to `+error.svelte` |
| `/rulings` can't clash with an item | Item addresses have two or three segments, and no route draws `/<project>` alone. The static `api/state` route does shadow a project named `api` with a round named `state`; standing answer 39 covers it |
| Project names are lower-case letters, digits, dot, dash, or underscore | `internal/state/events.go:13`: `^[a-z0-9][a-z0-9._-]*$` |
| The binary prints rulings oldest first, so the page sorts them | `snapshot.go:187` appends `s.Rulings` in file order. The twin probe printed `r-0001` at `2026-11-19T18:00:00Z` before `r-0002` at `2026-11-20T20:00:00Z` |
| A ruling names its question by id, so the page looks the text up | `internal/state/types.go:134-144`: `Question string json:"question,omitempty"`. The `started` snapshot's only ruling has `"question":"beta-q1"`, whose text is in `beta/b2`'s `needs.questions` |
| A round's needs come in three kinds, and `nothing` can carry a null reason | `snapshot.go:327-341`: `not_designed` when undesigned and `needs_geoff`; `nothing` when undesigned or with no question ids, with `Reason` set only from `no_need_reason`; else `questions`, answered ones included. `started` prints `alpha/r1` as `nothing` with `reason: null` |
| A pass has questions but no needs kind or reason | `snapshot.go:345-373` and `web/src/lib/snapshot.ts`'s `Pass`: `questions` only. Standing answer 36 settles what the pass page shows |
| A stopped pass carries a reason and no stop time | `snapshot.go:357-360` sets `StopReason`; `docs/follow-ups/snapshot-stop-times.md` says no stop time is in the snapshot |
| `ruling record` runs only from a registered desk, so a fixture's rulings are input files | `cc-desk ruling record --help`: "A session that is not a registered desk, a command that does not run under the desk's process, and a desk process with no terminal are each refused." `e2e/testdata/state/` holds `rulings.jsonl` and `desks.json` as input |
| The doc-page classes live in the mockup, not in Survey or the app | `.page`, `.page-main`, `.doc-meta`, `.doc-title`, and `.prose-doc` are at `docs/specs/2026-11-17-cc-desk-mockups/src/lib/survey/app.css:120-205`, `.calls li` at `src/lib/mock.css:208`. `git grep` finds none of them in `web/src` or in Survey's `theme/` sheets. `web/src/app.css:17` already defines `--doc-measure` |
| Labeled figures can be read by label | Quoted, not probed. `web/node_modules/playwright-core/types/types.d.ts:3100`: "Allows locating input elements by the text of the associated `<label>` or `aria-labelledby` element, or by the `aria-label` attribute." The docs speak of input elements, so a `dd` with `aria-labelledby` is unproven; task 2 checks it first, and standing answer 44 holds the fallback |
| Existing tests the new links can break | `sidebar.spec.ts:42` expects no link inside the "Waiting for you" region on `started`, so the rulings link sits outside that region. `empty.spec.ts:6` finds the link named "cc-desk", which a "Rulings" link doesn't match. `redraw.spec.ts:28` expects no `data-pass` row for `alpha/p13` before it starts. No board test counts links or matches a round row's whole text |

### What the binary prints for each test state

The clock is 2026-11-20T23:00:00Z, 2:00 pm in Alaska, in every row. `twin` is the new fixture
of task 1, probed with these files under `$CLAUDE_JOB_DIR/tmp/probe/twin3`. `started` and
`stopped` are pass 1's builders.

| State | Item | Printed |
| --- | --- | --- |
| `twin` | `alder/r1` | `not_started`, planned 100, actual null, before 0, needed 23:00Z, needs `nothing`, reason "Routine upkeep.", passes `a-one` (60, before 0) and `a-two` (40, before 60) |
| `twin` | `birch/r0` | `not_started`, planned 120, before 0, needs `questions` with "Start with the setup round?" answered, `answer_kind` `ruling` |
| `twin` | `birch/r1` | `not_started`, planned 90, before 120, needed 01:00Z on 11-21, needs `questions` with "Keep the long names?" unanswered, pass `b-one` |
| `twin` | `birch/r2` | `not_designed`, planned 30, before 210, forecast 02:30Z on 11-21, needs `not_designed`, no passes |
| `twin` | rulings | `r-0001` at 2026-11-19T18:00Z, "Yes, setup first.", question `birch-q0`; `r-0002` at 2026-11-20T20:00Z, "Hold the release a day.", no question |
| `started` | `alpha/r1` | `done`, planned 1200, actual 1800, before null, needed null, needs `nothing`, reason null |
| `started` | `alpha/r2` | `building`, planned 360, before null, needs "Which storage should the report use?" unanswered; passes `p13` (`building`, 100, before null), `p14` (200, before 100, needed 01:00Z), `p15` (60, before 300, needed 05:00Z, question `alpha-q1`) |
| `started` | `beta/b1` | `done`, needs `nothing`, reason "Routine cleanup." |
| `started` | `beta/b2` | needs "Keep the old format?" (`ruling`) and "Name the tool?" (`carried`), both answered |
| `started` | rulings | one: `r-0001` at 2026-11-18T16:00Z, "Drop the old format.", question `beta-q1` |
| `started` | all | 5 rounds and 19 passes |
| `stopped` | `alpha/p13` | `stopped`, `stop_reason` `scope` |

## Recorded searches

Run with `git grep -c` over `web/src` and `web/tests` at `893ceab`.

| Search | Hits | What it settles |
| --- | --- | --- |
| `href=` | 1 each in `ReleaseBlock.svelte`, `ReleaseBoard.svelte`, `Sidebar.svelte`, `+layout.svelte`, `+page.svelte` | Every link the page draws today; none points at a pass or at rulings |
| `class="t"` | 2, both `ReleaseBoard.svelte` (lines 44 and 58) | Line 58 is the unlinked pass name |
| `isOpen` | `ReleaseBoard.svelte` 2, `sums.ts` 1 | Only open rounds draw pass rows |
| `findRound` | `sums.ts` 1, `+page.svelte` 2 | The helper the round page reuses |
| `+error` | 0 | No error page yet |
| `rulings` | `snapshot.ts` 1 | Nothing draws rulings yet |
| `no reason recorded` | 0 | The words are new |
| `cc-desk <project>\|cc-desk {` | 0 | The command is new |
| `stop_reason\|stopReason` | `ReleaseBoard.svelte` 2, `snapshot.ts` 1, `words.ts` 2, `words.spec.ts` 2 | The stopped reason's words exist |
| `allStates` | `snapshot-type.spec.ts` 2, `state.ts` 1, `words.spec.ts` 2 | A new builder reaches the type and words tests with no other change |
| `testdata` | `state.ts` 6 | `copyFixture` reads only `e2e/testdata/` today |
| `data-pass` | `ReleaseBoard.svelte` 1, `redraw.spec.ts` 1 | `redraw.spec.ts:25-27` expects no `data-pass` row for `alpha/p13` before it starts |

## File list

| File | Change |
| --- | --- |
| `web/tests/fixtures/twin/releases/alder.json`, `birch.json`, `rulings.jsonl`, `desks.json` | New. The twin-`r1` fixture of task 1 |
| `web/tests/state.ts` | `copyFixture` can copy from `web/tests/fixtures/`; a `twin` builder |
| `web/tests/page.ts` | `show` takes an optional address |
| `web/src/routes/[project]/[round]/+page.svelte` | New. The round page |
| `web/src/routes/[project]/[round]/[pass]/+page.svelte` | New. The pass page |
| `web/src/routes/rulings/+page.svelte` | New. The rulings page |
| `web/src/routes/+error.svelte` | New. Not found, with a link home |
| `web/src/lib/*.svelte` | New components for the shared parts; the split is the executor's |
| `web/src/lib/sums.ts` | A pass finder and a question-text finder |
| `web/src/lib/ReleaseBoard.svelte` | Pass names become links, and closed rounds list their passes as links |
| `web/src/routes/+layout.svelte` or `web/src/lib/Sidebar.svelte` | A link to the rulings page in the rail |
| `web/src/app.css` | The doc-page rules ported from the mockup, in tokens |
| `web/tests/items.spec.ts`, `web/tests/rulings.spec.ts`, `web/tests/notfound.spec.ts` | New. The tests named in the tasks |

## Acceptance

The two items, word for word from the page design.

2. **Click-ahead.** A test collects every round and pass in the snapshot and fails when one
   has no link on the front page, or when its link's address isn't that item's. Each linked
   page shows the item's name in its heading, and its planned minutes and minutes of work
   before it in elements labeled for those fields, equal to the snapshot's. A fixture where
   two projects each have a round named `r1` gives two pages with different figures. A round
   with written questions shows them. A round with none shows "Nothing" and its reason. An
   undesigned round shows "Not designed yet" with its forecast time. A null field shows the
   en dash.
   - Met by tasks 1 to 3. Shown by `npx playwright test tests/items.spec.ts` in `web/`, run by
     the gate.
   - Before: `npx playwright test tests/items.spec.ts`, run in `web/` at `893ceab`, printed
     `Error: No tests found.` and exited 1. The probe above printed 404 for `/alpha/r1`.
     `ReleaseBoard.svelte:58` draws a pass name as `<span class="t">`, with no link, and a
     closed round draws no pass at all.
10. **The gate.** `bash scripts/gate.sh` passes, and it now runs the page's checks. A
    Playwright run with a skipped test, with no browser, or with fewer tests than the suite lists
    fails the gate.
    - Met by task 5. Shown by `bash scripts/gate.sh` printing the page steps, the new count, and
      `gate: ok`, and by the three failing runs in task 5.
    - Before: `bash scripts/gate.sh`, run bare at `893ceab`, printed `gate: web install`, `gate:
      web check` (`0 ERRORS 0 WARNINGS`), `gate: web build`, `gate: web tests`, `40 passed
      (30.4s)`, and `gate: ok`, in 1 minute 7 seconds.

### The strings the tests write out

The browser zone is `Europe/London` and the clock is pinned to 2026-11-20T23:00:00Z in every row.
A figure reads as a number with a thousands comma and "minutes", as in "1,800 minutes". A null
figure or time reads as the en dash with "None" as screen-reader text. The labels are the
executor's under Survey's writing standard; the test finds each figure with `getByLabel` and
writes out the label that lands.

| Address | State | The page shows |
| --- | --- | --- |
| `/alder/r1` | `twin` | Heading "r1"; planned "100 minutes"; before "0 minutes"; needed "Today, 2:00 pm"; actual the en dash; "Not started"; "Nothing" with "Routine upkeep."; links to `a-one` and `a-two`; the command `cc-desk alder r1` |
| `/birch/r1` | `twin` | Heading "r1"; planned "90 minutes"; before "120 minutes"; needed "Today, 4:00 pm"; the question "Keep the long names?"; the command `cc-desk birch r1` |
| `/birch/r0` | `twin` | The question "Start with the setup round?" with "Ruled on" |
| `/birch/r2` | `twin` | "Not designed yet" with "Today, about 5:30 pm"; planned "30 minutes"; before "210 minutes" |
| `/alpha/r1` | `started` | "Done"; actual "1,800 minutes"; before and needed the en dash, read as "None"; "Nothing" with "No reason recorded" |
| `/beta/b1` | `started` | "Nothing" with "Routine cleanup." |
| `/beta/b2` | `started` | "Keep the old format?" with "Ruled on", "Name the tool?" with "Carried over" |
| `/alpha/r2/p14` | `started` | Heading "p14"; planned "200 minutes"; before "100 minutes"; needed "Today, 4:00 pm"; "Not started"; a link to `/alpha/r2`; the command `cc-desk alpha r2` |
| `/alpha/r2/p15` | `started` | The question "Which storage should the report use?" |
| `/alpha/r2/p13` | `stopped` | "Stopped" and "A change of scope", and no "since" |
| `/rulings` | `twin` | Two rows, `r-0002` first: "Today, 11:00 am", no question (the en dash), "Hold the release a day."; then "Thu, Nov 19, 9:00 am", "Start with the setup round?", "Yes, setup first." |
| `/rulings` | `started` | One row: "Wed, Nov 18, 7:00 am", "Keep the old format?", "Drop the old format." |
| `/alpha/zz`, `/zz/r1`, `/alpha/r1/p13` | `started` | The not-found view naming the address, with a link to `/` |
| `/a/b/c/d` | `started` | `+error.svelte`'s not-found text with a link to `/` |

The executor may reword any phrase the design doesn't quote. The design's words stay exact:
"Nothing", "Not designed yet", the five pass words, and the en dash. "No reason recorded" is the
design's "no reason recorded" at a sentence start. The test writes out whatever wording lands.

## Tasks

The path is the sum, 85 minutes, since the tasks run in order. Each task commits by path, runs
the gate bare before the commit, and pushes. Each task's report names the gate run's conclusion
on its pushed commit from `gh run list --branch page-items`, or "no run yet" with the link.

### Task 1: the twin fixture and the helpers (15 minutes)

Outcome: a fixture where two projects each have a round named `r1` exists under `web/tests/`, and
the tests can build its state through the real binary and open any address.

- `web/tests/fixtures/twin/` holds the four files below, exactly as the planning probe ran them
  (`cc-desk snapshot --now 2026-11-20T23:00:00Z --usage-command e2e/testdata/usage40.sh` exited
  0 and printed the table above). No `snapshot.json`, `usage.json`, or `lock` is committed; the
  binary writes them.

<details>
<summary>The twin fixture's four files</summary>

`releases/alder.json`:

```json
{
  "format": 1,
  "project": "alder",
  "name": "v1",
  "repo": "glw907/alder",
  "rounds": [
    {"name": "r1", "summary": "Alder's first round.", "needs_geoff": false, "no_need_reason": "Routine upkeep.", "design_written": true}
  ],
  "passes": [
    {"name": "a-one", "summary": "Alder pass one.", "round": "r1", "planned_minutes": 60, "state": "not_started"},
    {"name": "a-two", "summary": "Alder pass two.", "round": "r1", "planned_minutes": 40, "state": "not_started", "after": ["a-one"]}
  ]
}
```

`releases/birch.json`:

```json
{
  "format": 1,
  "project": "birch",
  "name": "v1",
  "repo": "glw907/birch",
  "rounds": [
    {"name": "r0", "summary": "Birch's opening round.", "needs_geoff": false, "no_need_reason": "Setup only.", "design_written": true},
    {"name": "r1", "summary": "Birch's first round.", "needs_geoff": true, "after": ["r0"], "design_written": true},
    {"name": "r2", "summary": "Birch's round with no design yet.", "needs_geoff": true, "after": ["r1"], "planned_minutes": 30}
  ],
  "passes": [
    {"name": "b-zero", "summary": "Birch pass zero.", "round": "r0", "planned_minutes": 120, "state": "not_started", "questions": ["birch-q0"]},
    {"name": "b-one", "summary": "Birch pass one.", "round": "r1", "planned_minutes": 90, "state": "not_started", "after": ["b-zero"], "questions": ["birch-q1"]}
  ],
  "questions": [
    {"id": "birch-q0", "text": "Start with the setup round?", "options": ["yes", "no"], "recommended": "yes", "answer": {"ruling": "r-0001"}},
    {"id": "birch-q1", "text": "Keep the long names?", "options": ["keep", "shorten"], "recommended": "keep"}
  ]
}
```

`rulings.jsonl`:

```json
{"format":1,"id":"r-0001","time":"2026-11-19T18:00:00Z","desk":"desk-1","words":"Yes, setup first.","project":"birch","round":"r0","question":"birch-q0","option":"yes"}
{"format":1,"id":"r-0002","time":"2026-11-20T20:00:00Z","desk":"desk-1","words":"Hold the release a day."}
```

`desks.json`:

```json
{"format":1,"desks":[{"session_id":"desk-1","pid":4242,"started":"2026-11-19T17:00:00Z"}]}
```

</details>

- `state.ts`: `copyFixture` can copy a folder from `web/tests/fixtures/` as well as from
  `e2e/testdata/`. A `twin` builder copies the fixture, runs no verb, and returns
  `WORKED_CLOCK`, so `buildState` prints its snapshot through the binary. `allStates()` then
  includes it, and the type and words tests check it with no other change.
- `page.ts`: `show(app, page, name, address = '/')` opens that address.

Files: `web/tests/fixtures/twin/*`, `web/tests/state.ts`, `web/tests/page.ts`. Class `ordinary`.

Test: `tests/snapshot-type.spec.ts` and `tests/words.spec.ts`, now over the `twin` state too.

Acceptance check: `npx playwright test tests/snapshot-type.spec.ts tests/words.spec.ts` passes.
A temporary print of `allStates().twin` shows `alder/r1` planned 100 before 0 and `birch/r1`
planned 90 before 120; paste those two lines into the task's report and remove the print.

### Task 2: the round page and the pass page (30 minutes)

Outcome: every round and pass has a page at its address showing what the design lists.

- Both pages read `poll.answer.snapshot` and the route's params. With no snapshot yet they draw
  nothing beyond the layout's banners. With a snapshot and no such item they draw the not-found
  view: a heading, the address, and a link to `/`. A pass is found by project and name, and a
  pass whose round isn't the address's round is not found.
- The round page, after the mockup's `r/[project]/[n]/+page.svelte` and `screens/round-page.png`:
  a meta line linking to `/` with the project and release; an `h1` with the round's name; its
  summary; its state in words (`word('roundState', …)`); the figures; what it needs; its passes,
  each a link to `/<project>/<round>/<pass>` with its state in words; and the command
  `cc-desk <project> <round>` in a `code` element. No button and no "decide" section: the walk
  in done item 4 allows only "Copy command", which is pass 4.
- The pass page: a meta line linking to its round's page; an `h1` with the pass's name; its
  summary; its state in words, and for a stopped pass the reason in words
  (`word('stopReason', …)`) and no "since"; the figures; what it needs; and the same command,
  with the pass's round.
- The figures are a `dl`. Each `dd` itself carries `aria-labelledby` pointing at its `dt`'s id,
  so a test reads it with `getByLabel`. Never put the label on an inner `span` or `div`: a
  generic element can't be named, and pass 3's accessibility walk would flag it. Before writing
  the rest of the tests, check with one figure that `getByLabel` finds the `dd`; if it doesn't,
  standing answer 44. A round shows planned minutes, actual minutes, minutes of work
  before it, and needed time; a not-designed round also shows its forecast time. A pass shows
  planned, actual, before, and needed. A null value is the en dash with "None" as `sr-only` text,
  as the front page does. Figures take a thousands comma. Times use `formatWhen` from
  `time.ts`; the forecast passes `about`.
- What it needs, for a round: `questions` lists every question's text, an answered one with
  `word('answerKind', …)` beside it; `nothing` reads "Nothing" with the reason, or "No reason
  recorded" when the reason is null; `not_designed` reads "Not designed yet" with the forecast
  time. For a pass, standing answer 36.
- Words for a state sit in the element that carries its color, as on the board.
- `app.css` gains the doc-page rules from the mockup's `app.css:120-205` and `mock.css:208`, in
  tokens, with no literal. A gap takes pass 1's standing answer 31.
- `sums.ts` gains a pass finder (project, round, pass) and a question-text finder (project and
  question id, over every round's `needs.questions` and every pass's `questions`).

Files: `web/src/routes/[project]/[round]/+page.svelte`,
`web/src/routes/[project]/[round]/[pass]/+page.svelte`, `web/src/lib/*.svelte`,
`web/src/lib/sums.ts`, `web/src/app.css`, `web/tests/items.spec.ts`. Class `ordinary`.

Test: `tests/items.spec.ts` gets one test per address row of the strings table, from `/alder/r1`
to `/alpha/r2/p13`, each opened with `show(app, page, state, address)`. The `/alder/r1` and
`/birch/r1` tests together are the twin check: same name, different figures. The not-found rows
go to task 4.

Acceptance check: `npx playwright test tests/items.spec.ts` passes.

### Task 3: every item linked from the front page (15 minutes)

Outcome: every round and every pass has a link on the front page whose address is its own.

- `ReleaseBoard.svelte`: a pass name in an open round's row becomes a link to
  `/<project>/<round>/<pass>`. A closed round lists its passes under its line as plain links,
  with no state row and no `data-pass`, so `redraw.spec.ts:25-27` still finds no row for
  `alpha/p13` before it starts. No `details` or `summary`: done item 4's walk allows only "Copy
  command" in a button's role.
- The click-ahead test in `tests/items.spec.ts` runs on `started` and on `twin`. It builds the
  expected addresses from the snapshot itself, as `/${project}/${round}` and
  `/${project}/${round}/${pass}`, with no import from `web/src/`. It collects the front page's
  `a[href]` values and fails naming each item with no link. For each item it then opens the
  link, checks the `h1` is the item's name, and checks the planned and before figures by label
  against the snapshot's numbers (with the en dash for null). On `started` that is 5 rounds and
  19 passes; on `twin`, 4 rounds and 4 passes.

Files: `web/src/lib/ReleaseBoard.svelte`, `web/tests/items.spec.ts`. Class `ordinary`.

Acceptance check: `npx playwright test tests/items.spec.ts tests/redraw.spec.ts` passes. Then,
once and reverted before the commit, turn one pass link back into a `span`: the click-ahead test
fails naming that pass. Paste the failing line into the task's report.

### Task 4: the rulings page and not-found (10 minutes)

Outcome: the rulings have a page, and an address that names nothing shows not-found with a link
home.

- `web/src/routes/rulings/+page.svelte`: every ruling, newest first by `time`, each with its time
  (`formatWhen`), its question's text from the question-text finder or the en dash with "None"
  when it has no question, and Geoff's words. With no rulings it reads "No rulings yet".
- The rail gains a "Rulings" link to `/rulings`, after the "Waiting for you" and "Running"
  sections or beside the app's name, as the executor finds reads best. It never sits inside the
  "Waiting for you" region, which `sidebar.spec.ts:42` expects to hold no link on `started`.
- `web/src/routes/+error.svelte`: for a 404, a heading that reads not found and a link to `/`;
  any other status shows its message with the same link.
- Tests: `tests/rulings.spec.ts` (the two rulings rows of the strings table, and the order) and
  `tests/notfound.spec.ts` (the four not-found addresses, each with a link to `/` that leads to
  the front page).

Files: `web/src/routes/rulings/+page.svelte`, `web/src/routes/+error.svelte`,
`web/src/routes/+layout.svelte` or `web/src/lib/Sidebar.svelte`, `web/tests/rulings.spec.ts`,
`web/tests/notfound.spec.ts`. Class `ordinary`.

Acceptance check: `npx playwright test tests/rulings.spec.ts tests/notfound.spec.ts` passes.

### Task 5: the gate's evidence, the review, and the PR (15 minutes)

Outcome: done item 10 has its evidence, the branch is reviewed, and the PR is ready.

1. `bash scripts/gate.sh`, bare: paste the page lines, the new passed count, and `gate: ok`.
2. With one test changed to `test.skip`, uncommitted: `bash scripts/gate.sh` fails naming the
   skipped figure. Revert.
3. `PLAYWRIGHT_BROWSERS_PATH=/nonexistent bash scripts/gate.sh` fails at the Playwright run.
4. A short run: take the full count from `npx playwright test --list --reporter=list` in `web/`,
   run `npx playwright test tests/items.spec.ts`, then `bash
   web/scripts/check-playwright-report.sh web/test-results/report.json <that count>` fails naming
   the count.
5. Merge `origin/page-front` in if it moved and `origin/main` if main moved, then run the gate
   again. A Sessions snapshot change takes round standing answer 16.
6. One whole-branch `standards:diff-reviewer` read, folded. Then the PR body and ready.

Paste each command and the lines it printed under `## Acceptance` in the PR.

Files: none beyond fixes. Class `ordinary`.

## Standing answers

The twelve in `round-run.md` hold, with the round plan's 13 to 20, and these.

- **34. The path runs past 150 minutes.** The pass parks under `## Stop` in `docs/STATUS.md` on
  the branch, as round standing answer 14 says for 180.
- **35. A field the page wants isn't in the snapshot.** Round standing answer 15: draw what the
  snapshot has and file one follow-up under `docs/follow-ups/` for the Desk and skills round.
- **36. What a pass page shows it needs.** The pass's questions, each with its answered word.
  With none, "Nothing" with the round's reason when the round's needs are `nothing` with a
  reason, else "No reason recorded". File `docs/follow-ups/pass-needs-reason.md` asking the
  snapshot for a pass's own needs kind and reason.
- **37. A stopped pass and "since".** The snapshot has no stop time, so the page shows "Stopped"
  and the reason in words, and no time. `docs/follow-ups/snapshot-stop-times.md` stays as the
  design's Backlog rewords it; this pass doesn't edit it.
- **38. A closed round's pass links crowd the board.** Keep them, smaller and muted, under the
  round line, in tokens. Done item 2 asks for every pass on the front page, and the main design
  says every pass on the boards links to its page. The choice is listed under `## Calls` for
  Geoff (see `## Decisions`).
- **39. A project named `api` with a round named `state`.** The static `/api/state` route wins
  and that round's page can't be reached. File one follow-up naming it; no fixture has one, and
  the server's address stays.
- **40. The not-found view flashes before the first poll answers.** Draw the not-found view only
  once a snapshot is in. While `poll.answer` is empty, draw nothing beyond the banners.
- **41. A test reads a figure before the poll answers.** Wait on the heading first, as pass 1's
  tests wait on the widget. `retries` stays 0, and a flaky test takes pass 1's standing answer 29.
- **42. `svelte-autofixer` and `svelte-check` disagree.** `npm run check` in the gate wins;
  record the `svelte-autofixer` note under `## Decisions`.
- **43. `page-front` lands or moves mid-pass.** Merge `origin/page-front` (or `origin/main` once
  it holds pull request 7) into `page-items`, run the gate, and go on. Never rebase a pushed
  branch.
- **44. `getByLabel` doesn't find a `dd`.** Keep `aria-labelledby` on the `dd` and read the
  figure with `getByRole('definition', { name: <label> })`. If that finds nothing either, put
  `aria-label` with the same words on the `dd` and try both locators once more. Never move the
  label to an inner element, and never fall back to a test id. Record which locator landed under
  `## Decisions`.
- **45. The click-ahead test times out.** Split it into one test per release and state, so no
  single test opens all 24 pages of `started`. Never raise the global timeout or add retries.

## Decisions

- [planner] Done item 2 is read literally: every round and every pass has a link on the front
  page. A closed round's passes show as plain links under its line, and only an open round draws
  state rows. The main design's "Every round and every pass, in the sidebar and on the boards,
  links to a page about it" agrees. The reading passed over was linking a closed round's passes
  only from its round page, which the done item's "on the front page" rules out.
  Call: the board now lists every pass of every round as a link, where pass 1 showed passes only
  under running rounds. Geoff sees the change on the front page.
- [planner] The twin fixture lives in `web/tests/fixtures/twin/`, as the launch prompt says, and
  its rulings are input files, because `ruling record` runs only from a registered desk. Its
  snapshot is printed by the real binary at the worked clock, never written by hand.
- [planner] Not-found for an item is drawn in the page, because with `ssr = false` the snapshot
  arrives from the poll after the route has loaded. `+error.svelte` covers an address that
  matches no route.
- [planner] The rulings page is at `/rulings`, linked from the rail. Item addresses have two or
  three segments, so it can't clash with one.
- [planner] The click-ahead test builds each expected address from the snapshot, not from an app
  helper, so a wrong helper can't pass its own test.
- [planner] Builds on pull request 7's call 2: a missing `snapshot.json` reads "No state yet",
  and the item pages draw nothing under that banner. Call 1 (`csrf.trustedOrigins: ['*']`) is
  untouched; the pages add no write.
- [planner] Builds on pull request 7's `[advisor]`-free decisions: the poll's connection state,
  the en dash with "None", and the weekday-and-date form for a time past tomorrow, which the
  rulings page uses for "Thu, Nov 19, 9:00 am".
- [planner] Pull request 7's planning misses hold here: SvelteKit 3 has no `svelte.config.js`;
  `#lib` imports carry the extension (`#lib/sums.ts`); `Handle` comes from
  `@sveltejs/kit/hooks`; TypeScript 6 reports TS1360 under `satisfies`, so a type test uses a
  typed `const`.
- [planner] No review before code. The class table and the round plan set none for an `ordinary`
  pass. `round.md` asks for one plan-reviewer read whatever the class; the launch prompt and the
  round plan win here, and the close gets the whole-branch read.
- [planner] The sidebar's pass segments stay unlinked. Done item 2 tests the front page, and the
  sidebar's rounds already link to their pages.

## Planning misses

None yet.
