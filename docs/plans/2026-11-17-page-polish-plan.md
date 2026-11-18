# The polish: pass plan

Initiative: cc-desk
Round: page
Pass: page-polish

Class `ordinary`: the pass changes CSS and markup under `web/src/` and one screenshot script.
Planned 90 minutes. The path below is 80. Earlier passes in this round took about 0.6 of their
planned minutes, and weekly usage read 97 percent against a 100 percent stop when this plan was
written, so the pass keeps to the path and stops early rather than late.

Design: `docs/specs/2026-11-17-page-design.md`, "The polish" and done items 7 and 10. Round plan:
`docs/plans/2026-11-17-page-round.md`, round standing answers 15, 18, 19, and 20. The method is
the `standards:design-refinement` skill, run as a settled landing: the layout is approved, and
Geoff doesn't review this pass. Writing a person reads (the PR body, STATUS lines, follow-ups)
follows `/var/home/glw907/.claude/docs/voice/written-for-people.md`, loaded through the
`writing-voice` skill. Labels follow Survey's writing standard through the `survey` skill, and
this pass changes no label (round standing answer 18).

Worktree `/var/home/glw907/Projects/cc-desk/.claude/worktrees/page-polish`, branch
`page-polish`, cut from `origin/page-link` at `d5c0435` (pull request 13, open, which carries
pull requests 11, 9, and 7, all open and waiting for Geoff). Every push names its target, `git
push origin page-polish`. PR base: `main`. Until those four land, the PR's diff carries their
commits too. Before the PR is ready, merge `origin/page-link` in again if it moved past
`d5c0435`, and `origin/main` if main moved past `2f6293e`, after the after shots (task 4). The
Sessions round runs beside this one. This pass edits no Go file, nothing under `e2e/`,
`internal/`, or `cmd/`, and neither `scripts/gate.sh` nor the workflow, and it messages none of
that round's sessions.

## Pre-flight

No step needs Geoff before or after this pass. Every write is inside the worktree, a test's
`mkdtemp` folder, or `$CLAUDE_JOB_DIR/tmp`. The pass doesn't run `web/scripts/install-page.sh`
and doesn't restart `cc-desk-page.service`: the auto-mode entry `cc-desk-sessions-page`
(`~/.claude/settings.json:426`) names only `cc-desk-page-surface` and `cc-desk-page-link` for
that unit. The live page on 127.0.0.1:7908 picks up the polish when the conductor or a later
pass installs the landed branch.

- Permissions checked: `~/.local/state/claude-denials.jsonl`, read again on 2026-11-18 for this
  pass over 2026-10-19 to 2026-11-18. 130 rows, 2026-11-01T00:55:05Z to 2026-11-18T06:27:23Z,
  the same rows pass 4 read. Classes and rows: Classifier unavailable 47; Destructive or
  Unwanted Actions 14; Self-Modification 11; Interfere With Workloads 11; server-side classifier,
  no reason 10; Undermining Oversight 5; Unauthorized Access 4; Merge Without Review 3; Auto-Mode
  Bypass 3; Irreversible Local Destruction 2; Git Destructive 2; Create Public Surface 2;
  Unverifiable Deletion Scope 2; Sensitive-Source Provenance 2; Security Weaken 2; Production
  Deploy 1; Secret-Store Writes 1; DNS/Domain/Cert Changes 1; Bot Check Circumvention 1; Code
  from External 1; Shared Scratch Sweep 1; Safety Bypass Flag 1; Out-of-Place Publication 1; PII
  Data Handling 1; Modify Shared Resources 1. This pass runs `npm ci`, `npm run check`, `npm run
  build`, `npx playwright test` (whole and by spec), `bash scripts/gate.sh`, `go build -o` into
  `$CLAUDE_JOB_DIR/tmp`, `node tests/screenshots.ts` run by path, `vale` on named files, `cmp`
  and `sha256sum` on PNGs, `git push origin page-polish`, `gh pr create --draft`, `gh pr edit`,
  `gh pr ready`, `gh pr checks`, `gh run list --branch page-polish`, and Agent dispatches. A
  search of the rows' targets for `npm`, `npx`, `node `, `playwright`, `gate.sh`, `gh pr`, `git
  push`, `screenshot`, `svelte`, and `cc-desk` matched 21 rows. None refuses one of these
  commands run bare: they are `npm ci` chained and piped to `tail`, `node` scripts in other repos
  during a classifier outage, a `pkill` of a test process, `gh pr merge` and a merge judged by
  `gh pr view`, a push to `main`, a settings `jq` edit, and a launch during a classifier outage.
  Planning ran this pass's gate and the screenshot script bare on 2026-11-18, and neither was
  refused.
- Reviewed: no review before code, as the `ordinary` class sets (`pass-core` "Plan review
  lens"). `round.md` "A plan inside a round" asks for one `standards:plan-reviewer` read whatever
  the class, and the round plan gives one only to passes 1, 3, and 4; this plan follows the round
  plan and its planner prompt, and says so under `## Decisions`. The close gets one whole-branch
  `standards:diff-reviewer` read beside one `standards:daisyui-a11y-reviewer` read.
- Evidence: the mechanism table and the recorded searches below. Both acceptance checks were run
  on today's code before any code was written, and the outputs are under `## Acceptance`.

### Refused commands and their alternatives

| Refused in the log | This pass does instead |
| --- | --- |
| `npm ci` chained with other installs and piped to `tail` (2026-11-14) | Run `npm ci` alone, bare, from `web/`. The gate runs it itself |
| `node` one-liners and inline scripts during a classifier outage | A script file run by path: `node tests/screenshots.ts`. During an outage, wait 60 seconds and run once more |
| `kill` and `pkill` on test processes (Interfere With Workloads) | The screenshot script and the Playwright fixture stop each server they start. Stop a background command with TaskStop |
| `sleep` in the foreground | Wait with the Monitor tool |
| `git stash drop` | No stash. A temporary commit sets work aside |
| `gh pr merge`, and `gh pr view` used to judge a merge (Merge Without Review) | The pass merges nothing. It reads checks with `gh pr checks <n>` and `gh run list --branch page-polish` |
| Editing sandbox or auto-mode settings | Never edit a setting, an auto-mode entry, or a hook |

## Mechanism table

| Mechanism | Source or probe, with what it printed |
| --- | --- |
| The gate passes today and counts the suite | `bash scripts/gate.sh`, bare, at `d5c0435` on 2026-11-18: `gate: gofmt` to `gate: web tests`, `130 passed (1.5m)`, `gate: ok`, exit 0 |
| The screenshot script exists and shoots the front page only | `web/tests/screenshots.ts`. Usage `CC_DESK_BINARY=<binary> node tests/screenshots.ts <out> [state]`, state `started` by default. It builds the state with `buildState`, starts `node build` (the built app in `web/build`) on a free port, and for `light` and `dark` opens a 1600 by 1000 context in `Europe/London`, `en-US`, with the clock fixed by `page.clock.setFixedTime`. It waits for the region `Next decision session` and saves a full-page `<out>/<state>-<scheme>.png`. Run at `d5c0435` it printed `recorded the design of beta/b3` and `started alpha/p13` and wrote `started-dark.png` and `started-light.png`, two files |
| The script serves the last build, not the source | It runs `node build` in `web/`. A shot taken without `(cd web && npm run build)` first shows the previous build, so every run in this plan builds first |
| The script is deterministic | Two runs at `d5c0435`, then `cmp` of each pair: `started-light identical`, `started-dark identical`. SHA-256 light `ceaa0d97…f8d273`, dark `f9d2de33…bda4e6` both times |
| The suite skips the script | `web/playwright.config.ts` `testMatch: '**/*.spec.ts'`; `screenshots.ts` isn't a spec, so the gate's count doesn't change when it changes |
| The two-release fixture is `started` | Design done item 1: "the fixture of two releases from the State and rules round". `web/tests/state.ts:81`: `started` fills the folder with the `state` fixture (`e2e/testdata/state/releases/alpha.json` project `alpha` to `v1`, rounds `r1` and `r2`; `beta.json` to `v2`, rounds `b1` to `b3`), designs `beta/b3`, and starts `alpha/p13` at 23:00Z |
| The pages to shoot exist on `started` | `web/tests/items.spec.ts:64` shows `/alpha/r1`, `:91` `/alpha/r2/p14`, `:113` `/alpha/r2/p13` (on `stopped`). On `started`, `alpha/r2` is building with `p13` building and `p14` and `p15` not started (the baseline front-page shot shows them). The rulings fixture `e2e/testdata/state/rulings.jsonl` holds 1 line |
| The Svelte and CSS the polish may touch | 17 files: `web/src/app.css` and 16 `.svelte` files (list under "Recorded searches") |
| App CSS takes Survey tokens and a short list of named literals | `web/src/app.css:10-14`: the three widths, the 80rem breakpoint, line heights 1.25, 1.35, 1.5 and 1.55, and font weight 600, filed in `docs/follow-ups/survey-rail-muted-doc-measure.md` |
| No color literal, not even a mix | `web/tests/color.spec.ts:221-229`: a hex color; `rgb`, `hsl`, `hwb`, `lab`, `lch`, `oklab`, `oklch`, `color`, `color-mix`, or `light-dark` followed by `(`; an arbitrary Tailwind color; and (`:228`) CSS color names as whole words in any `.svelte` or `.css` file under `web/src`, comments too |
| No copy of Survey | `web/tests/color.spec.ts:256-275`: any trimmed line over 40 characters from Survey's `theme/*.css` found in a file under `web/` fails |
| Survey's tokens, v1.7.0 | `git -C ~/Projects/survey show v1.7.0:DESIGN.md`: no shadows, no gradients, no all-caps labels, WCAG 2.2 contrast. Spacing `--spacing-1px` to `--spacing-24px`, type `--text-step-1` to `--text-step-13` (0.64rem to 1.6rem), `--font-display` Spectral, `--font-body` Karla, `--font-weight-heading` 500, `--radius-field` 3px, `--radius-box` 4px, `--site-rule-color` (`theme/components.css:9`), `--motion-quick`, `--motion-base`, `--motion-ease` |
| Tokens the app already uses | `grep -o 'var(--[a-z0-9-]*' web/src/app.css \| sort \| uniq -c`: 41 distinct; the most used are `--spacing-12px` 17, `--color-muted` 12, `--spacing-24px` 11, `--border` 10, `--text-step-4` 9 |
| The selectors the tests read | `grep -ohE` over `web/tests/*.ts` for `locator(`, `getByRole(`, and `querySelectorAll(`: `getByRole('heading')` 11, `[data-banner=` 10, `main code` 6, `getByRole('main')` 4, `.sr-only` 3, `main .state` 3, `td` 3, `listitem` 3, `.seg` 2, `.scale` 2, `main [data-pass=` 2, `main a[href]` 2, `tbody tr` 2, the regions `Waiting for you` and `Next decision session`, the navigation `Status`, the links `Rulings`, `cc-desk`, `Open in a terminal`, `Go to the front page`, and named round and pass links, and the button `Copy command`. Class names in test strings: `.state` 11, `.seg` 7, `.sr-only` 3, `.scale` 2, and `.state-line`, `.next`, `.figures`, `.doc-meta`, `.banner`, `.asks` 1 each |
| The words a state carries sit in its colored element | `color.spec.ts:13-50` writes out every state word and which builder shows it; the words test reads the text inside each `.seg` and `.state` |
| The scan runs both themes and the 24px rule | `web/tests/scan.spec.ts:30` reads each theme's body background first; axe runs with `wcag22aa`, which runs `target-size`. `web/src/app.css` carries the 24px rules for `.rounds a.t`, `.pass-links a`, and `Command`'s link and button |
| Vale reads every tracked `.md` under `docs/` | `scripts/gate.sh:57-58`: `git ls-files 'docs/*.md'`, mock-ups left out, `vale --minAlertLevel=error`. This plan and the audit ledger are both read |
| The mock-ups are the benchmark for the settled layout | `docs/specs/2026-11-17-cc-desk-mockups/screens/front-page.png`, `round-page.png`, `design-page.png`, `report-page.png`: sidebar on `base-200`, Spectral headings, the round page as a reading column about 36rem wide with dashed rules between sections, boards in two columns with numbered rounds and state words at the right |
| Claude can't drive Chromium here | `~/.config/chromium` holds only `Geoff` and `System Profile`, no `Default` profile (pass 4's finding, unchanged). Every render is Playwright's Chromium |
| The repo is private | `gh repo view glw907/cc-desk --json visibility` printed `PRIVATE`. A committed PNG shows in the PR body through a `blob/<sha>/<path>?raw=true` link for any reader with access |
| The service grant leaves this session out | `~/.claude/settings.json:426`, the entry `cc-desk-sessions-page`: "The sessions `cc-desk-page-surface` and `cc-desk-page-link` may install and enable the user unit cc-desk-page.service" |

## Recorded searches

Run in the worktree at `d5c0435`.

| Search | Hits | What it settles |
| --- | --- | --- |
| `find web/src -name '*.svelte' -o -name '*.css'` | 17: `app.css`; `lib/` `Banners`, `Command`, `Figures`, `NextSession`, `NotFound`, `Questions`, `ReleaseBlock`, `ReleaseBoard`, `ScaleBar`, `Sidebar`; `routes/` `+error`, `+layout`, `+page`, `[project]/[round]/+page`, `[project]/[round]/[pass]/+page`, `rulings/+page` | The whole surface the three passes cover |
| `grep -c 'class='` per `.svelte` file | `ReleaseBoard` 17, `+page` 13, `NextSession` 9, `ReleaseBlock` 9, `[project]/[round]/+page` 9, `[pass]/+page` 8, `Sidebar` 8, `rulings/+page` 7, `+layout` 5, `Command` 4, `NotFound` 4, `+error` 4, `Banners` 3, `ScaleBar` 3, `Figures` 2, `Questions` 2 | Where the markup's classes live; most changes belong in `app.css` |
| Tests reading computed style or geometry: `getComputedStyle\|toHaveCSS\|boundingBox\|evaluate(` | `scan.spec.ts`, `color.spec.ts`, `link.spec.ts`, `surface.spec.ts` | None reads a spacing or size value, so a token swap breaks no geometry check |
| Tests reading class names | 12 files: `items`, `estimate`, `sidebar`, `surface`, `color`, `words`, `order`, `rulings`, `redraw`, `server`, `scan`, `page.ts` | Rename or drop no class the selector row names |
| Tests reading `app.css` by name | 0 | The source checks read every file under `web/src`, not `app.css` by name |
| `fullPage\|toHaveScreenshot` in `web/tests` | `screenshots.ts` only | No visual snapshot test pins the render, so before and after come from the script alone |
| `ls docs/design-benchmark` | No such folder | The ledger and the captures are new |
| `git ls-files docs/follow-ups` | 15 files, among them `survey-rail-muted-doc-measure.md` and `survey-table-header-contrast.md` | A new Survey gap joins the first by edit, or gets its own file |

## File list

| File | Change |
| --- | --- |
| `web/tests/screenshots.ts` | Shoots four pages, not one (task 1). Outside the round plan's `web/src/` row (see `## Decisions`) |
| `docs/design-benchmark/page-polish/before/*.png` | New. Eight shots at the commit before the range |
| `docs/design-benchmark/page-polish/ledger.md` | New. The audit ledger, the adjudication, and the decisions log |
| `web/src/app.css` | The polish, in tokens. Most of the range lands here |
| `web/src/lib/*.svelte`, `web/src/routes/**/*.svelte` | Only where a class or a wrapper must move; no word changes |
| `docs/follow-ups/survey-rail-muted-doc-measure.md` | Only if a new literal is named |
| `docs/design-benchmark/page-polish/after/*.png` | New. Eight shots at the range's last commit |
| `docs/follow-ups/<name>.md` | One file for each change the audit names that needs a new element or a new word |

## Acceptance

The two items, word for word from the page design.

7. **The polish.** The `standards:design-refinement` pass is one commit range. The round report
   lists its changes. Items 1 to 6 pass after it. The pull request holds screenshots of the front
   page, a round page, and a pass page in both themes, taken by one script from the two-release
   fixture at the commit before the range and at its last commit.
   - Met by tasks 1 to 5. Task 1's commit is the commit before the range. Its SHA goes in the PR
     body as the range's parent. The range runs from task 2's first commit to task 3's last (or
     the last fix from task 4 or 5 that touches `web/src`). The PR names both ends.
   - The shots: `(cd web && npm run build && CC_DESK_BINARY=$CLAUDE_JOB_DIR/tmp/cc-desk node
     tests/screenshots.ts ../docs/design-benchmark/page-polish/<before|after>)`, with the
     binary from `go build -o $CLAUDE_JOB_DIR/tmp/cc-desk ./cmd/cc-desk` at the repo root.
     `ls` of each folder prints eight files: `started-<front|round|pass|rulings>-<light|dark>.png`.
     The PR body shows each pair through links pinned by SHA.
   - "Items 1 to 6 pass after it": the gate at the range's last commit runs their specs
     (`estimate`, `items`, `order`, `redraw`, `color`, `words`, `scan`, `sidebar`, `rulings`,
     `surface`), and the PR names its run.
   - The round report is the conductor's (round standing answer 20). The PR's `## Changes`
     section lists each change by axis, with its ledger line, so the report can copy it.
   - Before: `CC_DESK_BINARY=… node tests/screenshots.ts <tmp>/shots-baseline` at `d5c0435`
     printed `recorded the design of beta/b3` and `started alpha/p13` and wrote
     `started-dark.png` and `started-light.png`. The script shoots no round page and no pass
     page, so item 7 fails today. `ls docs/design-benchmark` printed `No such file or
     directory`.
10. **The gate.** `bash scripts/gate.sh` passes, and it now runs the page's checks. A Playwright
    run with a skipped test, with no browser, or with fewer tests than the suite lists fails the
    gate.
    - Met by task 5. Shown by `bash scripts/gate.sh` printing `gate: web tests`, `130 passed`
      (the polish adds no spec, so the count holds), and `gate: ok`; and by `git diff d5c0435 --
      scripts/gate.sh web/scripts/check-playwright-report.sh .github` printing nothing, so the
      skip, no-browser, and count failures pull requests 11 and 13 showed still hold.
    - Before: `bash scripts/gate.sh`, run bare at `d5c0435` on 2026-11-18, printed `gate:
      gofmt`, `gate: go vet`, `gate: go test`, `gate: vale`, `gate: web install`, `gate: web
      check`, `gate: web build`, `gate: web tests`, `130 passed (1.5m)`, and `gate: ok`, and
      exited 0.

## The rules every change keeps

These come from round standing answer 18 ("It runs without Geoff. It changes spacing, type, and
hierarchy within Survey, and changes no word the tests read and no layout the mockup settles.")
and the tests above.

1. Spacing, type, and hierarchy only. No new color, surface, shadow, gradient, border style, or
   radius beyond what Survey's tokens already give a selector. A rule may swap one Survey color
   token for another when the audit's hierarchy finding needs it, as long as the scan passes.
2. Survey tokens only. A value Survey lacks is one of the named literals in `app.css:10-14`, or a
   new one added to that comment and to `docs/follow-ups/survey-rail-muted-doc-measure.md`.
   Never `color-mix(` or any color function: `color.spec.ts` bans it in `web/src`.
3. No word changes, added, or removed, visible or for a screen reader. No class a test selects on
   is renamed or dropped (the selector row).
4. No layout the mock-ups settle changes: the rail on the left at `--rail-width`, the reading
   column at `--doc-measure`, the boards' two columns at 80rem and up, the order of sections.
5. Every target the scan checks stays at least 24px. Contrast stays at WCAG 2.2 AA in both themes.
6. No dependency, no stack change, no Go change (round standing answer 15).
7. Each commit in the range passes `npm run check`, `npm run build`, and the specs task 3 names on
   its own, so a stop at any commit leaves a branch that lands.

## Tasks

The path is 80 minutes, since the tasks run in order. The executor's first act, before task 1,
is `gh pr create --draft --base main --head page-polish` with the start time from `date -u` in
the body; the plan's commit is already pushed. Every command names its working folder: `(cd web
&& …)` for `npm`, `npx`, and `node`, and the repo root for `bash scripts/gate.sh`, `go build`,
and `git`.

Each task commits by path and pushes with `git push origin page-polish`. The full gate, `bash
scripts/gate.sh` bare, runs once after task 3 and once in task 5. CI runs on pull requests, so the
draft PR comes first. Each task's report names the gate run's conclusion on its pushed commit
from `gh run list --branch page-polish`, or "no run yet" with the link.

### Task 1: the script and the before shots (10 minutes)

Outcome: one script shoots the front page, the round page `/alpha/r2`, the pass page
`/alpha/r2/p13`, and `/rulings` in both themes from `started`, and the before shots are committed
with it. This commit is the commit before the range.

- `web/tests/screenshots.ts`: keep the usage line, the state build, the server, the contexts, the
  clock, and the 1600 by 1000 viewport. Shoot a list of pages, each with an address, a name, and
  what to wait for: `front` `/` waits for the region `Next decision session` (as today); `round`
  `/alpha/r2`, `pass` `/alpha/r2/p13`, and `rulings` `/rulings` wait for the `main` heading of
  level 1. Each saves a full-page `<out>/<state>-<name>-<scheme>.png`. The top comment says which pages it shoots and that design done item 7 reads its output. Comments
  follow `standards:ts-conventions`.
- Build and shoot from the repo root: `go build -o $CLAUDE_JOB_DIR/tmp/cc-desk ./cmd/cc-desk`, then
  `(cd web && npm run build)`, then `(cd web && CC_DESK_BINARY=$CLAUDE_JOB_DIR/tmp/cc-desk node
  tests/screenshots.ts ../docs/design-benchmark/page-polish/before)`.
- Run it a second time into `$CLAUDE_JOB_DIR/tmp/before-2` and `cmp` each of the eight pairs. All
  eight match, or standing answer 67 applies.
- Read all eight shots. Each shows its page drawn, not an error page or "No state yet".

Files: `web/tests/screenshots.ts`, `docs/design-benchmark/page-polish/before/` (8 PNGs). Class
`ordinary`.

Acceptance check: `ls docs/design-benchmark/page-polish/before` prints the eight names; the eight
`cmp` runs print nothing; `(cd web && npm run check)` passes. The commit message names it the
commit before the range, and the task report gives its SHA.

### Task 2: the audit (15 minutes)

Outcome: a committed ledger grades the four pages against the mock-ups, with every knob marked
`already-right`, `recommend`, or `not-applicable`, and the executor's ruling on each `recommend`.

Two read-only lenses run in parallel, each a fresh context, each given the paths of the eight
before shots, the four mock-up screens, `web/src/app.css`, the 16 `.svelte` files, Survey's
`DESIGN.md` at v1.7.0 (`git -C ~/Projects/survey show v1.7.0:DESIGN.md`), and "The rules every
change keeps" above, pasted whole. Neither edits a file.

1. Lens A, typography, rhythm, and hierarchy. `subagent_type: general-purpose`, `model: opus`,
   effort `medium`. Its brief, word for word: "You are a read-only typography and rhythm
   reviewer for a settled page. The layout is approved; you grade the last fifteen percent.
   The burden of proof is on the change, and a mostly already-right ledger is the expected
   outcome of auditing good work. Never manufacture recommendations to look useful. The owner's
   dose: 'It runs without Geoff. It changes spacing, type, and hierarchy within Survey, and
   changes no word the tests read and no layout the mockup settles.' Grade these knobs on each
   page and in both themes: the type scale against Survey's steps, heading sizes and weights,
   leading tuned to size and measure, the measure of the reading column, heading margins (tighter
   above, looser below), vertical rhythm between sections, spacing that follows hierarchy,
   alignment of labels, numbers, and state words, tabular figures where digits column, and
   text-wrap on headings. Also answer one question: does anything here read as AI-default?
   (cream, serif, and terracotta; one acid accent on near-black; hairline broadsheet rules with
   zero radius; identical rounded cards with one shadow; all-caps eyebrows over every heading,
   middle-dot meta strings, monospace data labels, or an arrow on link text; one colored accent
   word in a headline; 01/02/03 markers on content that is not a sequence; pill buttons.) For
   each knob return one line: the page, the selector, `already-right`, `recommend`, or
   `not-applicable`; for `recommend`, the exact rule change in Survey tokens and a felt-impact
   rating of low, medium, or high. Name no new word, element, or layout."
2. Lens B, measured. `subagent_type: general-purpose`, `model: sonnet`, effort `medium`. It
   serves the built app from `started` through the same steps the script uses (a script file
   under `$CLAUDE_JOB_DIR/tmp`, run by path, never a `node -e` one-liner) and measures, at
   viewports 1600 and 1200 wide (both sides of the 80rem breakpoint) in both themes, with
   `getComputedStyle` and `getBoundingClientRect`: font size, line height, and weight of every
   heading and body run; the gaps between sections and rows; whether state words, numbers, and
   bar ends share an edge; the focus ring on every link family and the button (by pressing Tab);
   each target's size; the reading column's width in characters; and horizontal overflow. Its
   brief carries the same burden-of-proof sentence, the dose quote, and the rule that every
   finding cites a measured number. It returns the same one-line ledger form.
3. Write `docs/design-benchmark/page-polish/ledger.md` yourself from both returns. A section per
   page, a table per section (knob, lens, verdict, change, felt impact, ruling), and a closing
   "Decisions log" list. Rule on every `recommend`: `apply` only when it keeps all seven rules
   and its felt impact is medium or high, or it is low and costs one line; `decline` with a
   reason otherwise. A `recommend` that needs a new element, a new word, or a layout change is
   declined and filed (standing answer 69). Tokens, selectors, and values go in code spans.
   Then `vale --minAlertLevel=error docs/design-benchmark/page-polish/ledger.md` prints nothing.

Files: `docs/design-benchmark/page-polish/ledger.md`. Class `ordinary`.

Acceptance check: the ledger names every page and both themes, every `recommend` has a ruling,
and vale prints nothing on it. The commit is the range's first.

### Task 3: the three passes (25 minutes)

Outcome: every `apply` ruling lands, grouped by the method's passes, one commit each, in order:
(a) spacing, alignment, and hierarchy; (b) type and token conformance; (c) states, focus, and
accessibility. A pass with no `apply` gets no commit, and the ledger says so.

- Each pass covers every page the rulings reach, so local fixes compose. State the treatment of
  every element a pass touches in the commit body before the diff; never patch one element after
  a render shows it off.
- Load the `survey` skill before the first rule, `standards:svelte-conventions` before any
  `.svelte` comment, and `standards:daisyui` before any daisyUI class. Run the Svelte MCP
  server's `svelte-autofixer` on every `.svelte` file a pass changes until it reports no issue.
- After each pass: `(cd web && npm run check && npm run build)`, then `(cd web && npx playwright
  test tests/color.spec.ts tests/scan.spec.ts tests/surface.spec.ts tests/words.spec.ts
  tests/sidebar.spec.ts tests/items.spec.ts)`, then shoot the four pages into
  `$CLAUDE_JOB_DIR/tmp/pass-<a|b|c>` and read the shots against the before shots. That read is the
  builder's own check, never the gate (task 4 holds the gate read).
- After pass (c): `bash scripts/gate.sh`, bare.

Files: `web/src/app.css`, and only the `.svelte` files a ruling names; the ledger's ruling
column marks each `apply` as `landed in <sha>`. Class `ordinary`.

Acceptance check: the named specs pass after each pass, the gate passes after pass (c), and
`git diff <task 1 sha> -- web/src | grep -E '^\+.*(#[0-9a-fA-F]{3}|rgb\(|oklch\(|color-mix\()'`
prints nothing.

### Task 4: the fresh read and the after shots (12 minutes)

Outcome: a context that didn't build the polish grades before against after, every regression it
names is reverted, and the after shots are committed.

1. Build and shoot into `$CLAUDE_JOB_DIR/tmp/after-candidate` with task 1's commands.
2. Dispatch `standards:visual-verifier` (`model: opus`, effort `medium`; it takes its model from
   its frontmatter where that differs) with the eight before shots, the eight candidates, and the
   four mock-up screens as separate labeled blocks, the ledger's `apply` list, and "The rules
   every change keeps". It returns a verdict per change: kept, `COSMETIC` regression, or
   `STRUCTURAL` regression.
3. Revert each change it calls a regression (standing answer 72), by a new commit, never a
   rebase. If anything reverted, shoot the candidates again.
4. Shoot into `docs/design-benchmark/page-polish/after` at the range's last commit, with that
   commit's SHA in the task report, and commit the eight PNGs. This commit sits after the range.

Files: `docs/design-benchmark/page-polish/after/` (8 PNGs), reverts if any. Class `ordinary`.

Acceptance check: `ls` prints the eight names; the task report holds each verdict from `standards:visual-verifier`;
the after commit changes no file under `web/`.

### Task 5: the gate, the reviews, and the PR (18 minutes)

Outcome: done items 7 and 10 have their evidence, the branch is reviewed, and the PR is ready.

1. Merge `origin/page-link` in if it moved past `d5c0435`, and `origin/main` if main moved past
   `2f6293e`. Never rebase a pushed branch. If a merge changes a file under `web/src`, it doesn't
   change the after shots, which belong to the range's last commit; say so in the PR.
2. `bash scripts/gate.sh`, bare: paste `gate: web tests`, the passed count, and `gate: ok`. Then
   `git diff d5c0435 -- scripts/gate.sh web/scripts/check-playwright-report.sh .github` prints
   nothing.
3. Dispatch one whole-branch `standards:diff-reviewer` and one `standards:daisyui-a11y-reviewer`
   in parallel, each with the range (task 1's SHA to the range's last commit), this plan, and the
   design's done items 7 and 10. Fold their findings. A fix under `web/src` joins the range: shoot
   the after shots again, commit them over the old ones, and run the gate once more.
4. The PR body, per `round-run.md` "Evidence" and `pass-core` "Close": the opening paragraph per
   written-for-people; `Round: page`, `Pass: page-polish`, `Review: …`, `Calls: …`;
   `## Changes` (each applied change by axis with its ledger line, for the round report);
   `## Screenshots` (the four pages, each light and dark, before beside after, linked as
   `https://github.com/glw907/cc-desk/blob/<sha>/docs/design-benchmark/page-polish/<before|after>/<file>?raw=true`
   with the SHA of the commit that holds the file); `## Acceptance` with items 7 and 10 word
   for word and their evidence, including the before outputs above; `## Decisions` (this plan's
   lines, then yours); `## Planning misses`; `## Calls for Geoff`; `## Review`; and the score with
   its `pass-score` comment. `gh pr view <n> --json body` shows every link with a full SHA.
5. Mark the PR ready.

Files: the PR body, and review fixes. Class `ordinary`.

Acceptance check: the gate prints `gate: ok`; both reviews are folded with no open finding;
`gh pr checks <n>` is green on the last commit.

## Standing answers

The twelve in `round-run.md` hold, with the round plan's 13 to 20, pass 4's 58 to 66 where they
apply, and these.

- **67. The two runs of the script differ.** Find the moving part (a clock, an animation, a font
  still loading, a scroll position), fix the script so it waits for it or freezes it, and shoot
  again until two runs match. If no fix holds in 10 minutes, ship the shots from one run and name
  the difference under `## Decisions`.
- **68. The ledger comes back all `already-right`, or every `recommend` is declined.** That is the
  expected outcome of auditing good work. The range is the ledger commit alone, the after shots
  match the before shots, and the PR says so. `Call:` the polish found nothing to change; Geoff
  can ask for more after he has used the page, as the design says.
- **69. A finding needs a new element, a new word, or a layout change** (the mock-up's mark beside
  the name `cc-desk` in the rail is one). Decline it in the ledger and file one follow-up for it under
  `docs/follow-ups/`. It isn't polish.
- **70. A change needs a color value Survey lacks.** Use an existing Survey token, or drop the
  change. `color.spec.ts` bans every color function in `web/src`, and a new color isn't in this
  pass's dose. File the gap through the `survey` skill's `/log-issue` step.
- **71. A change drops a target under 24px, fails contrast, or fails a spec.** Revert that change,
  mark it `declined` in the ledger with the measurement, and go on. Never edit a test, an axe
  rule, or an excluded element to make it pass.
- **72. The fresh reader calls a change a regression.** Revert it by a new commit. A
  `STRUCTURAL` call on a change the ledger rates high may go through the advisor once; the
  advisor's choice stands and is an `[advisor]` line.
- **73. Usage reaches the ceiling, or the pass passes 120 minutes, mid-range.** Every range commit
  passes on its own (rule 7), so stop at the last one that passed: shoot the after shots there,
  commit them, write the stop under `## Stop` in `docs/STATUS.md` with what landed and what
  didn't, push, and end. The PR stays a draft.
- **74. A spec fails on a change that touches no word and no named class.** Read the failure. If
  the test reads a layout fact the polish moved, revert the change. Never change the test.
- **75. `origin/page-link` or `origin/main` moved and the merge conflicts in a file this pass
  changes.** Keep both sides' intent, run the gate, and say so under `## Decisions`. A conflict in
  a file this pass doesn't own goes through the advisor.
- **76. The before outputs recorded here no longer match** (another commit landed under the worktree
  before task 1). Run both checks again, record the new output in the PR's `## Acceptance`, and
  list it under `## Planning misses` as `[stale-fact]`.

## Decisions

- The polish runs as a settled landing of `standards:design-refinement`, at the dose round
  standing answer 18 sets. The dose names spacing, type, and hierarchy, so the audit runs two
  lenses (typography, rhythm, and hierarchy on Opus; a measured lens on Sonnet) and no color,
  surface, and depth lens. A third lens would recommend outside the dose, and usage reads 97
  percent.
- `web/tests/screenshots.ts` changes in task 1, though the round plan's row for this pass names
  `web/src/` only. Item 7 needs one script to shoot a round page and a pass page, and the script
  exists only in `web/tests/`. The suite skips it (`testMatch: '**/*.spec.ts'`), so the gate's
  count holds.
- The script change and the before shots land in one commit, the commit before the range. The
  `web/src` tree at that commit is `d5c0435`'s, so "taken at the commit before the range" holds.
- The shots are committed under `docs/design-benchmark/page-polish/`, where the skill keeps its
  captures, and the PR links them by SHA. Pass 4 kept its shots in the job folder, which goes
  away with the job; item 7 asks the PR to hold them.
- The audit renders at 1600 and 1200 wide, both sides of the boards' 80rem breakpoint, in both
  themes. The design names no narrow width and no phone, so the screenshots stay at the script's
  1600.
- Pages shot: `/alpha/r2` (a building round with a question, three passes, and the command) and
  `/alpha/r2/p13` (a building pass) on `started`, with `/rulings` as the fourth page the
  planner prompt names.
- No `install-page.sh` run. The service grant names other sessions, and nothing in item 7 needs
  the live page.
- Any change to `.command-actions`, the terminal link, or the "Copy command" button builds on
  pull request 13's calls (both the link and the button shipped). Any change to the muted pass
  links under a closed round builds on pull request 9's call. Each such ruling in the ledger
  says so, and the PR repeats it under `## Decisions`.
- Plan review: none before code, as the class and the round plan set, against `round.md` "A plan
  inside a round". The close adds a `standards:daisyui-a11y-reviewer` read beside the
  whole-branch `standards:diff-reviewer`, as the planner prompt asks.
