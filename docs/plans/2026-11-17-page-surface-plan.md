# The reading surface: pass plan

Initiative: cc-desk
Round: page
Pass: page-surface

Class `runner`: the pass installs a user unit on the workstation. Planned 60 minutes. The path
below is 105, under the 120-minute park line (twice the planned minutes, round standing answer
14).

Design: `docs/specs/2026-11-17-page-design.md`, "The app" (the server paragraph and the unit
paragraph), "Themes", "State words" (the last paragraph), and done items 4, 6, 8, and 10. Round
plan: `docs/plans/2026-11-17-page-round.md`. Where the round design is silent,
`docs/specs/2026-11-17-cc-desk-design.md` holds. Writing a person reads (the PR body, STATUS
lines, the README, follow-ups) follows `~/.claude/docs/voice/written-for-people.md`, loaded
through the `writing-voice` skill. Labels on the page follow Survey's writing standard, through
the `survey` skill (round standing answer 19).

Worktree `/var/home/glw907/Projects/cc-desk/.claude/worktrees/page-surface`, branch
`page-surface`, cut from `origin/page-items` at `f8536e5` (pull request 9, open, which carries
pull request 7, open). PR base: `main`. Push: `glw907/cc-desk`. Before the PR is ready, merge
`origin/page-items` in again if it moved, and `origin/main` if main moved. The Sessions round
runs beside this one. This pass edits no Go file, nothing under `e2e/`, `internal/`, or `cmd/`,
and neither `scripts/gate.sh` nor the workflow, and it messages none of that round's sessions.

## Pre-flight

No step in this pass needs Geoff. The auto-mode entry `cc-desk-sessions-page`
(`~/.claude/settings.json:426`) reads: "The sessions `cc-desk-page-surface` and
`cc-desk-page-link` may install and enable the user unit cc-desk-page.service, listening on
127.0.0.1:7908, and register a desktop handler for the cc-desk:// scheme." The conductor reports
it live and tracked, so round standing answer 13's gate is met. Every other write is inside the
worktree, a test's temporary folder, `$CLAUDE_JOB_DIR/tmp`, the deploy folder
`~/.local/share/cc-desk/page/`, or `~/.config/systemd/user/`.

- Permissions checked: `~/.local/state/claude-denials.jsonl`, read again on 2026-11-18 for this
  pass over 2026-10-19 to 2026-11-18. 130 rows, 2026-11-01T00:55:05Z to 2026-11-18T06:27:23Z, the
  same rows passes 1 and 2 read. Classes and rows: Classifier unavailable 47; Destructive or
  Unwanted Actions 14; Self-Modification 11; Interfere With Workloads 11; server-side classifier,
  no reason 10; Undermining Oversight 5; Unauthorized Access 4; Merge Without Review 3; Auto-Mode
  Bypass 3; Irreversible Local Destruction 2; Git Destructive 2; Create Public Surface 2;
  Unverifiable Deletion Scope 2; Sensitive-Source Provenance 2; Security Weaken 2; Production
  Deploy 1; Secret-Store Writes 1; DNS/Domain/Cert Changes 1; Bot Check Circumvention 1; Code from
  External 1; Shared Scratch Sweep 1; Safety Bypass Flag 1; Out-of-Place Publication 1; PII Data
  Handling 1; Modify Shared Resources 1. This pass runs `npm ci`, `npm install --save-dev
  --save-exact @axe-core/playwright@4.14.0`, `npm run check`, `npm run build`, `npx playwright
  test`, `bash scripts/gate.sh`, `bash web/scripts/install-page.sh`, `systemctl --user
  daemon-reload`, `systemctl --user enable --now`, `systemctl --user restart`, `systemctl --user
  show`, `systemd-analyze --user verify`, a drop-in written under
  `~/.config/systemd/user/cc-desk-page.service.d/` and removed after, `curl -s` on 127.0.0.1,
  `ss -ltn`, `git push origin page-surface`, `gh pr create`, `gh pr ready`, `gh pr checks`, and
  `gh run list`. A search of the rows' targets for `systemctl`, `chmod`, `npm`, `node`, `curl`,
  `ss -l`, `playwright`, `cc-desk`, `daemon-reload`, and `install` found these near ones: one
  `npm ci` chained with other installs and piped to `tail`; two `node` scripts run during
  classifier outages and two `node --test` runs chained after `pkill` (Interfere With Workloads);
  one `curl` of a phone API with credentials (Unauthorized Access); one settings `jq` edit naming
  `chmod` and `curl` (Self-Modification). No row names `systemctl`, the unit, `ss`, or `chmod` on
  a test folder. No row refuses any of this pass's commands run bare. The table below holds the
  alternatives.
- Reviewed: one `standards:plan-reviewer` read before code, as the `runner` class sets
  (`pass-core` "Plan review lens"), folded on 2026-11-18; its `[review]` line is under `## Decisions`. The
  close gets one whole-branch `standards:diff-reviewer` read.
- Evidence: the mechanism table and the recorded searches below. Every acceptance check was run
  on today's code before any code was written. The outputs are under `## Acceptance`.

### Refused commands and their alternatives

| Refused in the log | This pass does instead |
| --- | --- |
| `npm ci` chained with other installs and piped to `tail` (2026-11-14) | Run `npm ci` alone, bare, from `web/` |
| `node` one-liners and inline scripts during a classifier outage | A script file, run by path. The install script is `web/scripts/install-page.sh` |
| `kill`, `pkill`, and `kill -9` on test processes (Interfere With Workloads) | The Playwright fixture stops each server it starts. The unit stops with `systemctl --user stop`. A background command stops with `TaskStop` |
| `sleep 90` in the foreground (2026-11-18) | Wait with the `Monitor` tool, or a `curl --retry` loop inside the install script |
| A settings edit that adds the Playwright download hosts (Security Weaken, 2026-11-08) | No settings edit. A missing browser parks, as pass 1's standing answer 22 says |
| `rm -rf` sweeps of shared temporary folders (Shared Scratch Sweep, Unverifiable Deletion Scope) | Tests make each folder with `mkdtemp` and remove only that folder, after restoring its write bits. The install script replaces only `~/.local/share/cc-desk/page/` |
| `git stash drop` (2026-11-17) | No stash. A temporary commit sets work aside |
| `gh pr merge`, and `gh pr view` used to judge a merge (Merge Without Review) | The pass merges nothing. It reads checks with `gh pr checks <n>` and `gh run list --branch page-surface` |
| A classifier outage, 47 rows | Wait 60 seconds through `Monitor`, then run the command once more |
| Any refusal of a `systemctl --user` step, the copy to the deploy folder, or the drop-in | Standing answer 49 |

## Mechanism table

| Mechanism | Source or probe, with what it printed |
| --- | --- |
| Every page draws one `input`: the drawer toggle | `web/src/routes/+layout.svelte:29-30`: `<div class="drawer drawer-open">` then `<input id="rail" type="checkbox" class="drawer-toggle" />`. A planning walk of the 26 pages reached from `/` on `started` found `input` count 1 and `role=button` count 0 on every page |
| Removing the toggle hides the sidebar, so the layout needs a plain grid | `web/node_modules/daisyui/daisyui.css`: `.drawer-open{…&>.drawer-toggle{display:none;&~.drawer-side{pointer-events:auto;visibility:visible;…opacity:1;…display:block;position:sticky…}`. The side panel is shown only through `.drawer-toggle ~ .drawer-side`. Survey's `DESIGN.md:251-259` plain-app recipe names no drawer. The mockup (`src/routes/+layout.svelte:11-12`) has the same toggle |
| The server already answers 405 to every method but `GET` and `HEAD` | `web/src/hooks.server.ts` returns `new Response('Method Not Allowed', { status: 405, headers: { allow: 'GET, HEAD' } })`. It reaches the hook only because `web/vite.config.ts` sets `csrf: { trustedOrigins: ['*'] }` (pull request 7's call 1). A planning probe of the copied build printed 405 to `POST /` and `OPTIONS /` |
| The built app runs from a read-only copy of `build/` with no `node_modules` | Planning probe: `build/` copied to a temporary folder, `chmod -R a-w` on it and on a state folder, then `HOST=127.0.0.1 PORT=7918 CC_DESK_STATE_DIR=<folder> ~/.local/share/mise/installs/node/lts/bin/node <copy>` from `cd /`. It printed only `Listening on http://127.0.0.1:7918`, with an empty stderr. `GET /` was 200, `/api/state` held `"project":"alpha"`, and `ss` showed `127.0.0.1:7918` |
| `GET /` holds no project name, because the page is drawn in the browser | `web/src/routes/+layout.ts`: `export const ssr = false;`. The probe's `GET /` body is the SvelteKit shell. Done item 8 is amended for this (see `## Acceptance` and `## Decisions`) |
| The server source holds one file call, a read | `grep -rnE "from 'node:(fs\|fs/promises\|child_process)'\|writeFile\|appendFile\|createWriteStream\|mkdir\|rename\|unlink\|rmSync\|copyFile\|spawn\|exec" web/src` printed one line: `src/lib/server/state.ts:3: import { readFile } from 'node:fs/promises';` |
| The state folder comes from `CC_DESK_STATE_DIR`, then `$XDG_STATE_HOME/cc-desk`, then `~/.local/state/cc-desk` | `web/src/lib/server/state.ts`, `stateDir()` |
| The live state folder doesn't exist | `ls ~/.local/state/cc-desk` failed, and `XDG_STATE_HOME` is empty, on 2026-11-18 |
| adapter-node binds every address unless `HOST` is set | `web/build/server/adapter-node-env.js` holds `env("HOST", "0.0.0.0")`. The unit sets `HOST=127.0.0.1` |
| A user unit's `PATH` lacks the mise Node | `systemctl --user show-environment` printed `PATH=/usr/lib64/ccache:/usr/local/sbin:/usr/local/bin:/usr/bin` and `HOME=/var/home/glw907` |
| The mise `lts` symlink is a stable Node path | `command -v node` and `mise which node` both print `/var/home/glw907/.local/share/mise/installs/node/lts/bin/node`, a symlink to 26.13.1. Homebrew's `/home/linuxbrew/.linuxbrew/bin/node` is also 26.13.1 and isn't used |
| `%h` works in `ExecStart`, and `systemd-analyze` checks the path | A probe unit with `ExecStart=%h/.local/share/mise/installs/node/lts/bin/node %h/.local/share/cc-desk/page` passed `systemd-analyze --user verify` with exit 0. The same unit with `node/nope/` printed `Command /var/home/glw907/.local/share/mise/installs/node/nope/bin/node is not executable` and exit 1 |
| Port 7908 is free, and nothing is installed | `ss -ltn` showed 7907 held by the approval queue (node, 127.0.0.1) and nothing on 7908. `curl -s http://127.0.0.1:7908/` exited 7. `systemctl --user status cc-desk-page.service` printed `Unit cc-desk-page.service could not be found.` `~/.local/share/cc-desk` doesn't exist |
| Prior art runs a build from a checkout | The approval queue on 7907 runs as `node /var/home/glw907/Projects/approval-queue/approval-queue/build` from a terminal scope, not a unit |
| The two themes give different backgrounds | Planning probe with Playwright's `colorScheme`: the body background read `rgb(247, 247, 243)` in light and `rgb(17, 23, 30)` in dark. Survey ships `survey` (default) and `survey-dark` (`prefersdark: true`) |
| axe-core finds two kinds of violation today | Planning probe with `@axe-core/playwright` 4.14.0 (it pulls `axe-core` 4.14.1) and the tags `wcag2a`, `wcag2aa`, `wcag21a`, `wcag21aa`, `wcag22aa`, on every page reached from `/` on `started`, in both themes. On `/`: `target-size`, 7 nodes in both themes, the round links `a.t` (`r1`, `b1`, `b3`) and the closed-round pass links in `.pass-links` (`p01`, `b-one`, `b-three`, `b-four`), 10 by 21.6px to 41 by 14px against 24px. On `/rulings`: `color-contrast`, 3 nodes in light only, the `th` cells When, Question, and Ruling, `#777a7b` on `#f7f7f3`, 4.02 against 4.5. Every other page: none |
| The tokens the fixes can use | `--spacing-24px: 24px` is in `survey.css`, `survey-vars.css`, and `components.css`. `--color-muted` is `#5d5f58` in light and `#a2a9b1` in dark |
| A sidebar bar's segments carry color and no words | `web/src/lib/ScaleBar.svelte`: `<div class="scale" class:slim role="img" aria-label={label}>` around `<span class="seg {part.state}" class:asks={part.asks} …></span>`, empty. The triangle is `.seg.asks::after` in `app.css`. Used by `ReleaseBlock.svelte` (twice) and `ReleaseBoard.svelte` (once) |
| The board's state words already sit in the colored element | `ReleaseBlock.svelte` and `ReleaseBoard.svelte` draw `<span class="state {state}">word</span>`; the item pages draw `.state-line` with `.state {state}` |
| The widget has no state color | `web/src/lib/NextSession.svelte`: the words are in `[data-testid=session-state]` inside `.session-time`. `app.css:78-82` gives `.session-time` a size and no color. Standing answer 51 says what the test checks |
| Each banner's words sit in the colored box | `web/src/lib/Banners.svelte`: `.banner` with `data-banner` set to `not-connected`, `no-state`, or `stale`, its words in a `<strong>`. `app.css` gives `.banner` the `--color-error` border and `--cairn-error-ink` |
| The app's CSS has no color literal today | A grep of `web/src` `.css` and `.svelte` files for hex, `rgb(`, `hsl(`, `oklch(`, and color keywords printed three lines: `transparent` at `app.css:41` and `app.css:132`, and `white-space` at `app.css:261` (a false hit on `white`) |
| The app imports Survey and copies none of it | `web/src/app.css` imports `survey/theme/fonts.css`, `survey/theme/survey.css`, and `survey/theme/components.css`. `web/package.json` pins `github:glw907/survey#v1.7.0` |
| The test fixture spawns the app with no log and from `web/build` | `web/tests/app.ts`: `spawn('node', [buildDir], { env: { HOST: '127.0.0.1', PORT, CC_DESK_STATE_DIR, SHUTDOWN_TIMEOUT: '1' }, stdio: 'ignore' })`, with `buildDir = path.resolve(import.meta.dirname, '..', 'build')` |
| `server.spec.ts` covers `POST`, `PUT`, `PATCH`, and `DELETE` on `/` and `/api/state`, and `HEAD` | `web/tests/server.spec.ts` |

## Recorded searches

Run with `git grep -c` over `web/src` and `web/tests` at `f8536e5` unless marked.

| Search | Hits | What it settles |
| --- | --- | --- |
| `drawer` | `src/routes/+layout.svelte` 4 | The drawer lives in the layout alone; no test names it |
| `<input\|<form\|<select\|<textarea\|contenteditable\|<button` | `src/routes/+layout.svelte` 1 | The toggle is the one control |
| `role="img"` | `src/lib/ScaleBar.svelte` 1 | The one bar component |
| `\.seg` (in `tests/`) | `sidebar.spec.ts` 6 | The triangle tests read `.seg` and `asks`; segment text must not break them |
| `pass-links` | `src/lib/ReleaseBoard.svelte` 2, `src/app.css` 2 | Pull request 9's closed-round pass links, one of the scan's targets |
| `class="t"` | `src/lib/ReleaseBoard.svelte` 2 | The round and pass name links, the other target |
| `<th` | `src/routes/rulings/+page.svelte` 3 | The rulings table headers, the contrast target |
| `stdio` | `tests/app.ts` 1 | The one spawn to extend |
| `@axe-core` | 0, and 0 in `web/package.json` | The scan package is new |
| `cc-desk-page\|7908` (repo-wide `git grep -c`) | `docs/` only: the page design, the round plan, and planner prompts | No unit, script, or README line yet |
| `ls-files web/systemd web/scripts` | `web/scripts/check-playwright-report.sh` 1 | Both new files are new; `web/systemd/` doesn't exist |

## File list

| File | Change |
| --- | --- |
| `web/src/routes/+layout.svelte` | The drawer and its toggle give way to a plain two-column layout, the rail first in the source as today's drawer-side is not |
| `web/src/app.css` | The layout's grid in tokens; the scan's two fixes; `.sr-only` use only if Tailwind's isn't enough |
| `web/src/lib/ScaleBar.svelte` | A list of segments, each holding its state's words, and an `asks` segment the triangle's words, as visually hidden text; a prop names the words table |
| `web/src/lib/ReleaseBlock.svelte` | Passes the bar its words table; maybe a scan fix |
| `web/src/lib/ReleaseBoard.svelte` | Passes the bar its words table; maybe a scan fix |
| `web/src/routes/rulings/+page.svelte` | Only if the scan fix can't be made from `app.css` alone |
| `web/package.json`, `web/package-lock.json` | `@axe-core/playwright` 4.14.0, exact, as a dev dependency |
| `web/tests/app.ts` | The fixture captures the server's output and can start from another app folder |
| `web/tests/crawl.ts` (or a name the executor picks) | New. The crawl helper with its draw wait, shared by the walk and the scan |
| `web/tests/state.ts` | Only for a builder standing answer 56 adds |
| `web/tests/sidebar.spec.ts`, and `order.spec.ts` or `redraw.spec.ts` if the grep finds them | A state-word assert scoped to the state span, so a bar's words can't satisfy it |
| `web/tests/surface.spec.ts` | New. The walk, the methods, the read-only run, the log, and the source search |
| `web/tests/scan.spec.ts` | New. The theme check and the axe-core scan |
| `web/tests/color.spec.ts` | New. Words with color, no color literal, no copy of Survey |
| `web/systemd/cc-desk-page.service` | New. The user unit |
| `web/scripts/install-page.sh` | New. Builds, deploys, installs, enables, and restarts |
| `README.md` | How to install the page, and to run the script again after a change lands |
| `docs/follow-ups/survey-table-header-contrast.md` | New. The Survey gap the rulings scan found |

## Acceptance

The four items, word for word from the page design.

4. **Reading surface.** A test walks every page, following every link on this site, and finds
   no `form`, `input`, `select`, `textarea`, or `contenteditable`, and nothing with the role
   of a button but "Copy command". A `POST` to any address returns 405. The walk and the poll
   run with the state folder and the app's folder both read-only, the server's log holds no
   error, and afterward every file in both is byte for byte what it was. A test searches the
   server's source and finds no call that writes a file.
   - Met by task 1. Shown by `npx playwright test tests/surface.spec.ts tests/server.spec.ts` in
     `web/`, run by the gate, and by the planted failure in task 1.
   - Before: `npx playwright test tests/surface.spec.ts` printed `Error: No tests found.` The
     planning walk found one `input` on each of the 26 pages reached from `/` on `started`. The
     source search above printed its one `readFile` line. The app fixture starts the app only
     from `web/build` and drops its output.
6. **Words and contrast.** An axe-core scan with the WCAG 2.2 AA rules passes on every page in
   both themes. The test first reads the page's background color in each theme and fails if
   the two are equal. A test finds the words for a state inside the element that carries its
   color: each pass state, each widget state, the triangle, and each banner. A test finds no
   color literal in any `.css` or `.svelte` file under `web/src`, in a stylesheet, a `style`
   attribute, or a class such as `bg-[#fff]`, and no copy of Survey's files in `web/`.
   - Met by tasks 2 and 3. Shown by `npx playwright test tests/scan.spec.ts tests/color.spec.ts`
     in `web/`, run by the gate, and by the planted failures in task 3.
   - Before: both files are new, so `Error: No tests found.` The planning scan printed
     `target-size` (7 nodes, `/`, both themes) and `color-contrast` (3 nodes, `/rulings`, light),
     with the two backgrounds `rgb(247, 247, 243)` and `rgb(17, 23, 30)`. The sidebar bar's
     segments are empty `span`s. The color grep printed the three lines above, none a literal.
8. **The service.** After the install script runs, `curl -s http://127.0.0.1:7908/` returns a
   page holding the name of a project in the live state folder, and it does so again after
   `systemctl --user restart cc-desk-page.service`. `ss -ltn` shows port 7908 bound to
   `127.0.0.1` and to no other address.
   - Amended check (`Call:` under `## Decisions`). The page is drawn in the browser
     (`ssr = false`), so `GET /` holds no project name, and the live state folder doesn't exist
     yet. The check runs against a copy of `e2e/testdata/state`, pointed at by a drop-in that is
     removed afterward. After the install script runs,
     `curl -s http://127.0.0.1:7908/` returns 200 with the app's shell, and `curl -s
     http://127.0.0.1:7908/api/state` returns the snapshot holding `"project":"alpha"`. Both
     hold again after `systemctl --user restart cc-desk-page.service`. `ss -ltn` shows 7908 on
     `127.0.0.1` only. With the drop-in removed and the unit restarted, `systemctl --user show
     -p Environment cc-desk-page.service` names no `CC_DESK_STATE_DIR`, and `/api/state` reads
     `"noState":true` from the live folder, or the live snapshot if the folder exists by then.
   - Met by task 4. Shown by the run in task 4, pasted into the PR. CI has no user manager, so
     this check runs once on the workstation and isn't in the gate.
   - Before: `curl -s http://127.0.0.1:7908/` exited 7. `systemctl --user status
     cc-desk-page.service` printed `Unit cc-desk-page.service could not be found.` `ss -ltn`
     showed no 7908. `git ls-files web/systemd web/scripts` printed only
     `web/scripts/check-playwright-report.sh`.
10. **The gate.** `bash scripts/gate.sh` passes, and it now runs the page's checks. A
    Playwright run with a skipped test, with no browser, or with fewer tests than the suite lists
    fails the gate.
    - Met by task 5. Shown by `bash scripts/gate.sh` printing the page steps, the new count, and
      `gate: ok`, and by the three failing runs in task 5.
    - Before: `bash scripts/gate.sh`, run bare at `f8536e5` on 2026-11-18, printed `gate:
      gofmt`, `gate: go vet`, `gate: go test`, `gate: vale`, `gate: web install`, `gate: web
      check` (`348 FILES 0 ERRORS 0 WARNINGS`), `gate: web build`, `gate: web tests`, `60 passed
      (37.5s)`, and `gate: ok`.

## The strings the tests write out

Every expected string below is written out in the test that reads it, never imported from
`web/src/`.

| Where | Words |
| --- | --- |
| A pass segment, and `.state.<code>` for a pass | `not_started` Not started, `waiting` Waiting, `building` Building, `stopped` Stopped, `done` Done |
| A round segment, and `.state.<code>` for a round | `not_designed` Not designed yet, `not_started` Not started, `building` Building, `done` Done |
| The triangle, added to an `asks` segment's words | Needs Geoff (the design's phrase at "a round that needs Geoff"; checked through the `survey` skill's writing standard before the first commit, and changed in the plan's place in the PR if Survey's standard says otherwise) |
| The widget, `[data-testid=session-state]` | `needs_you_now` Needs you now, `ready_now` Ready now, `nothing_waiting` Nothing waiting, and for `scheduled` the formatted time, such as `Today, about 8:00 pm` on `started` (`estimate.spec.ts:8`), never the word Scheduled |
| The banners, in the `<strong>` | `not-connected` Not connected, `no-state` No state yet, `stale` Stale |

## Tasks

The path is the sum, 105 minutes, since the tasks run in order. Each task commits by path, runs
the gate bare before the commit, and pushes. Each task's report names the gate run's conclusion
on its pushed commit from `gh run list --branch page-surface`, or "no run yet" with the link,
and also the conclusion on the previous task's head. The close names the last head's.

### Task 1: the walk and the read-only run (25 minutes)

Outcome: every page on the site is walked with both folders read-only, holds no form control,
refuses every write method, and leaves both folders as they were; the server's source writes no
file.

- `+layout.svelte`: drop the daisyUI drawer and its toggle. Lay the rail and the content out as
  two columns with CSS grid in `app.css`, the rail `var(--rail-width)` wide and sticky to the
  viewport as it is today, in tokens. The page must look as it does now; compare a screenshot of
  `/` on `started` at 1440 by 900 before and after, and put both in the task's report. The
  sidebar and front-page tests must pass unchanged.
- `tests/app.ts`: the fixture keeps the server's stdout and stderr (piped, kept in memory) and
  exposes them, and `start` can take another app folder than `web/build`. Existing tests use it
  unchanged.
- A crawl helper in a new file under `web/tests/`, shared with task 2. It pins the browser clock
  as `show()` does, opens a page, and waits for it to draw before reading it: the first
  `/api/state` response, then `nav [data-release]` or the page's `h1`, or
  `[data-banner="no-state"]` on the no-state page. It collects every `a[href]` on this site and
  visits each once.
- `tests/surface.spec.ts`, one test for the walk:
  1. Build `started` into a state folder. Copy `web/build` into a `mkdtemp` folder. Hash every
     file in both folders (SHA-256 by relative path). `chmod -R a-w` both. Then write a file
     into each folder and expect `EACCES`, so a run as root, where `a-w` does nothing, fails here
     instead of passing the read-only claim with nothing proved.
  2. Start the app from the copy and crawl from `/`. On every page, count `form`, `input`,
     `select`, `textarea`, and `[contenteditable]`, and every `getByRole('button')` whose name
     isn't "Copy command"; each count must be 0, and the failure names the page and the element.
  3. On every address visited, plus `/api/state`, an address that names nothing, and one
     `/_app/immutable/` asset taken from the page's own `script[src]`, send `POST`, `PUT`,
     `PATCH`, `DELETE`, and `OPTIONS` with Playwright's `request`; each must be 405 with
     `allow: GET, HEAD`. The asset's 405 comes from adapter-node's handler
     (`build/server/adapter-node-handler.js`), the rest from `hooks.server.ts`.
  4. Keep one page open until it has sent at least three `/api/state` requests (count them with
     `page.on('request')`), so the poll has run against the read-only folders.
  5. Stop the app. Its stderr is empty, and no stdout line matches `/error/i`. Hash again and
     compare: every file the same, no file added or gone.
  6. In a `finally`, whatever failed: restore the write bits on both folders, then remove them.
     A failing walk otherwise leaves read-only folders the fixture can't remove (`rmSync` throws
     `EACCES`).
  The test asserts it visited every address the snapshot names (`/`, `/rulings`, and each
  release's rounds and passes at `/<project>/<round>` and `/<project>/<round>/<pass>`), built
  from the snapshot in the test, so a broken crawl can't pass by visiting one page and a fixture
  change from the Sessions round moves the floor with it. Planning reached 26 pages on `started`.
- `tests/surface.spec.ts`, one test for the source: read every `.ts`, `.js`, and `.svelte` file
  under `web/src` and fail on any of these, each written out in the test:
  - an import or `require` of `(node:)?(fs|fs/promises|child_process)` that names anything but
    `readFile`, `stat`, or `access`;
  - a default or namespace import of `fs` or `fs/promises` (`import fs from`, `import * as fs`);
  - a call, matched as the name then `\s*\(`, of `writeFile`, `writeFileSync`, `appendFile`,
    `createWriteStream`, `mkdir`, `mkdirSync`, `rename`, `unlink`, `rm`, `rmSync`, `rmdir`,
    `copyFile`, `cp`, `open`, `spawn`, `exec`, `execSync`, `execFile`, or `fork`, with a word
    boundary before the name so `format(` and `execute(` don't hit.
  The `open(` match is allowed only where the file holds no `node:fs` import, since `window.open`
  isn't a file call; the test says so in a comment.

Files: `web/src/routes/+layout.svelte`, `web/src/app.css`, `web/tests/app.ts`, the crawl
helper, `web/tests/surface.spec.ts`. Class `ordinary`.

Test: `tests/surface.spec.ts`, with `tests/server.spec.ts` and `tests/sidebar.spec.ts` unchanged.

Acceptance check: `npx playwright test tests/surface.spec.ts tests/server.spec.ts
tests/sidebar.spec.ts` passes. Then, once each and reverted before the commit: put an `<input>`
back in the layout, and the walk fails naming it and leaves no folder behind; add `import {
writeFile } from 'node:fs/promises'` to `state.ts`, and the source test fails naming the file.
Paste both failing lines into the task's report.

### Task 2: the scan in both themes (25 minutes)

Outcome: every page passes the WCAG 2.2 AA scan in light and in dark, and the test proves the
two themes differ.

- `npm install --save-dev --save-exact @axe-core/playwright@4.14.0` in `web/`, alone and bare.
  Commit `package.json` and `package-lock.json`.
- `tests/scan.spec.ts`:
  1. First, one test opens `/` on `started` in light and in dark (`page.emulateMedia({
     colorScheme })`), reads the computed `background-color` of `body` in each, and fails if they
     are equal.
  2. Then one test per state and theme from the start, so no test holds dozens of scans: scan with
     `new AxeBuilder({ page }).withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa',
     'wcag22aa'])`, after the crawl helper's draw wait, on every page the crawl reaches on
     `started` and on `twin`; on `/` for `stopped`, `two`, and `guess`; on `/` with a
     `snapshot-error.json` (the stale banner); on `/` with no snapshot ("No state yet"); and on
     `/alpha/zz` and `/a/b/c/d`. The failure lists each violation's id, the page, and the first
     target of each node.
- Fix what the scan finds, within Survey's tokens:
  - `target-size` on `a.t` and `.pass-links a`: give those links a 24px minimum target from
    `var(--spacing-24px)` (a minimum block size with inline-flex, or padding, as reads best).
    Check the board and the sidebar still look as they did; a screenshot of `/` before and after
    goes in the report.
  - `color-contrast` on the rulings `th`: set their ink to `var(--color-muted)` (`#5d5f58` on
    `#f7f7f3` is 6.03). Write `docs/follow-ups/survey-table-header-contrast.md` naming what was
    measured (`#777a7b` on `#f7f7f3`, 4.02), the fix used, and that Survey's daisyUI table header
    needs it, for the conductor to file in Survey's backlog.
  - Any other finding takes standing answer 47.

Files: `web/package.json`, `web/package-lock.json`, `web/tests/scan.spec.ts`, the crawl helper,
`web/src/app.css`, maybe `web/src/lib/ReleaseBoard.svelte` and
`web/src/routes/rulings/+page.svelte`, `docs/follow-ups/survey-table-header-contrast.md`. Class
`ordinary`.

Acceptance check: `npx playwright test tests/scan.spec.ts` passes. Before the fixes it fails
listing `target-size` and `color-contrast`; paste that failing list into the report.

### Task 3: words with color, no literal, no copy (20 minutes)

Outcome: every state shown in color carries its words in the same element, a screen reader
hears them, and the app's CSS holds no color literal and no copy of Survey.

- `ScaleBar.svelte`: the bar becomes a `role="list"` with its `aria-label`, and each segment a
  `role="listitem"` holding its state's words, and for an `asks` segment the triangle's words
  too, as visually hidden text. Children of `role="img"` are presentational, so words inside
  today's bar would reach no screen reader. A new prop names which table the parts' codes come
  from (`passState` for `passParts`, `roundState` for `roundParts`), since a `Part` carries only
  its code. `ReleaseBlock.svelte` and `ReleaseBoard.svelte` pass it. The `@component` comment
  and the `label` prop's comment change to match.
- `sidebar.spec.ts:35`: `round.toContainText('Building')` reads the whole round item, which will
  then hold p13's segment word "Building" too. Scope it to the round's own state span
  (`round.locator('.figures .state')`, or whatever selector the markup gives). Grep every other
  `toContainText` of a state word over a release block or a board row (`order.spec.ts:29`,
  `order.spec.ts:37`, `redraw.spec.ts:34` today) and scope each the same way where a bar now sits
  inside the located element.
- `tests/color.spec.ts`:
  1. **Words with color.** For each state in "The strings the tests write out", find a builder in
     `tests/state.ts` that shows it, and check each colored element's `textContent` holds its
     words: `.seg.<code>` in the sidebar and on the board for every pass and round code; an
     `asks` segment for the triangle; `.state.<code>` on the board and the item pages;
     `.banner[data-banner=…]` for each banner (stop the server for `not-connected`); and the
     widget by standing answer 51. The test lists, written out, which builder shows which state,
     and fails on any listed state no builder shows, rather than skipping it. Planning found
     `ready_now` on `big` and `scheduled` on `started`; standing answer 56 covers a state no
     builder reaches.
  2. **No color literal.** Read every `.css` and `.svelte` file under `web/src` and fail on any
     of these, each written out in the test, in a stylesheet, a `<style>` block, a `style`
     attribute, or a `class`:
     - a hex color, `#([0-9a-fA-F]{8}|[0-9a-fA-F]{6}|[0-9a-fA-F]{3,4})(?![\w-])`, with the
       trailing boundary so `{#each` doesn't hit;
     - a color function: `rgb(`, `rgba(`, `hsl(`, `hsla(`, `hwb(`, `lab(`, `lch(`, `oklab(`,
       `oklch(`, `color(`, `color-mix(`, or `light-dark(`;
     - one of the 148 CSS named colors (the list in `node_modules/tailwindcss/dist/lib.js`),
       minus `transparent` and `currentcolor`, matched case-blind with the hyphen-aware boundary
       `(?<![\w-])(names)(?![\w-])`, so `white-space` doesn't hit. CSS system colors are left
       out, since `mark` hits two comments today;
     - a Tailwind arbitrary color such as `bg-[#fff]` or `text-[rgb(…)]`;
     - a Tailwind palette class in a `class` attribute,
       `(bg|text|border|ring|outline|fill|stroke|decoration|accent|caret|shadow|from|via|to)-(<named>|slate|gray|zinc|neutral|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose|black|white)(-\d+)?`,
       since Survey resets no `--color-*` and those classes would compile.
     Each failure names the file, the line, and the literal. Planning ran the hex and keyword
     patterns with these boundaries over today's tree: 0 hits.
  3. **No copy of Survey.** Hash every file under `web/node_modules/survey/` and every file in
     `web/` outside `node_modules/`, `build/`, `.svelte-kit/`, and `test-results/`; fail on any
     match. Also fail on any file there that holds a trimmed line of Survey's `theme/*.css`
     longer than 40 characters, so an edited copy is caught too. The thresholds are written out
     in the test. Planning ran both over today's 55 files: 0 hits.

Files: `web/src/lib/ScaleBar.svelte`, `web/src/lib/ReleaseBlock.svelte`,
`web/src/lib/ReleaseBoard.svelte`, `web/tests/sidebar.spec.ts`, and any of
`web/tests/order.spec.ts` and `web/tests/redraw.spec.ts` the grep finds, `web/tests/color.spec.ts`.
Class `ordinary`.

Acceptance check: `npx playwright test tests/color.spec.ts tests/sidebar.spec.ts tests/scan.spec.ts`
passes. Then, once each and reverted before the commit: empty a segment's words, and the words
test fails naming it; add `color: #fff;` to `app.css`, and the literal test fails naming the
line; copy `node_modules/survey/theme/survey-vars.css` into `web/src/`, and the copy test fails
naming it. Paste the three failing lines into the report.

### Task 4: the service (15 minutes)

Outcome: a user unit runs the built page on 127.0.0.1:7908, and a script installs it.

- `web/systemd/cc-desk-page.service`: `Description`; `Environment=HOST=127.0.0.1 PORT=7908`;
  `ExecStart=%h/.local/share/mise/installs/node/lts/bin/node %h/.local/share/cc-desk/page`;
  `Restart=on-failure`; `WantedBy=default.target`. A comment in the unit says why Node is named
  by its full path: a user unit's `PATH` has no mise shims, and the `lts` symlink is what `mise
  which node` prints. No state folder is set, so the server reads the live one.
- `web/scripts/install-page.sh`, Bash with `set -euo pipefail`, safe to run again:
  1. In `web/`: `npm ci` if `node_modules` is missing, then `npm run build`.
  2. Build every path from `share="${HOME:?}/.local/share/cc-desk"`. Copy `build/` to
     `$share/page.new`, move any `$share/page` to `$share/page.old`, move `page.new` to `page`,
     and remove only `$share/page.old` by that name.
  3. Copy the unit to `~/.config/systemd/user/`, run `systemd-analyze --user verify` on it, and
     `systemctl --user daemon-reload`.
  4. `systemctl --user enable --now cc-desk-page.service`, then `systemctl --user restart
     cc-desk-page.service` so a re-run serves the new build.
  5. Wait with `curl -sf --retry 10 --retry-connrefused --retry-delay 1 -o /dev/null
     http://127.0.0.1:7908/`, and print the address.
- `README.md`: a short section on the page: what it is, `bash web/scripts/install-page.sh` to
  install it or to serve a newer build, the address, and that the Move round records the
  install in the workstation repo.
- Run done item 8's amended check on the workstation, in this order, and paste every command
  with what it printed into the PR. Any stop after step 3 first removes the drop-in, as
  standing answer 57 says.
  1. `bash web/scripts/install-page.sh`.
  2. `cp -r e2e/testdata/state "$CLAUDE_JOB_DIR/tmp/page-check-state"`, from the worktree's
     root. Its `snapshot.json` already holds `"project": "alpha"`, and the server reads only
     `snapshot.json`, so no binary run is needed for a check that isn't a test.
  3. Write `~/.config/systemd/user/cc-desk-page.service.d/zz-check.conf` holding `[Service]` and
     `Environment=CC_DESK_STATE_DIR=<that folder, written out in full>`. Then `systemctl --user
     daemon-reload` and `systemctl --user restart cc-desk-page.service`.
  4. `curl -s -o /dev/null -w '%{http_code}' --retry 10 --retry-connrefused --retry-delay 1
     http://127.0.0.1:7908/` prints `200`, and `curl -s http://127.0.0.1:7908/api/state` holds
     `"project":"alpha"`.
  5. `systemctl --user restart cc-desk-page.service`. Then `curl -s -o /dev/null -w
     '%{http_code}' --retry 10 --retry-connrefused --retry-delay 1 http://127.0.0.1:7908/`
     prints `200` again, and `curl -s http://127.0.0.1:7908/api/state` holds `"project":"alpha"`
     again.
  6. `ss -ltn` shows `127.0.0.1:7908`, and `ss -ltn | grep -c ':7908 '` prints `1`.
  7. `ls -la ~/.local/state/cc-desk`, recorded. Remove `zz-check.conf` and its folder, then
     `systemctl --user daemon-reload` and `systemctl --user restart cc-desk-page.service`.
     `systemctl --user show -p Environment cc-desk-page.service` prints `Environment=HOST=127.0.0.1
     PORT=7908` with no `CC_DESK_STATE_DIR`. If the live folder was missing, `curl -s
     http://127.0.0.1:7908/api/state` holds `"noState":true`. If it exists by then (the Sessions
     round runs beside this one), the snapshot's `generated_at` in that answer equals the one in
     `~/.local/state/cc-desk/snapshot.json`, and the PR says which case ran.
  8. `systemctl --user is-enabled cc-desk-page.service` prints `enabled`, and `systemctl --user
     is-active cc-desk-page.service` prints `active`. The unit stays installed.

Files: `web/systemd/cc-desk-page.service`, `web/scripts/install-page.sh`, `README.md`. Class
`runner`.

Acceptance check: the eight steps above print what they name.

### Task 5: the gate's evidence, the review, and the PR (20 minutes)

Outcome: done item 10 has its evidence, the branch is reviewed, and the PR is ready. The new
suite adds a 26-page crawl and about 60 scans, so each gate run is longer than pass 2's.

1. `bash scripts/gate.sh`, bare: paste the page lines, the new passed count, and `gate: ok`.
2. With one test changed to `test.skip`, uncommitted: `bash scripts/gate.sh` fails naming the
   skipped figure. Revert.
3. `PLAYWRIGHT_BROWSERS_PATH=/nonexistent bash scripts/gate.sh` fails at the Playwright run.
4. A short run: take the full count from `npx playwright test --list --reporter=list` in `web/`,
   run `npx playwright test tests/color.spec.ts`, then `bash
   web/scripts/check-playwright-report.sh web/test-results/report.json <that count>` fails naming
   the count.
5. Merge `origin/page-items` in if it moved and `origin/main` if main moved, then run the gate
   again. If either brought a change to `web/`, run `bash web/scripts/install-page.sh` again.
   Then `curl -s -o /dev/null -w '%{http_code}' --retry 10 --retry-connrefused --retry-delay 1
   http://127.0.0.1:7908/` prints `200`, and `curl -s http://127.0.0.1:7908/api/state` holds what
   task 4 step 7 found (`"noState":true`, or the live `generated_at`).
6. One whole-branch `standards:diff-reviewer` read, folded. Then the PR body and ready.

Paste each command and the lines it printed under `## Acceptance` in the PR.

Files: none beyond fixes. Class `ordinary`.

## Standing answers

The twelve in `round-run.md` hold, with the round plan's 13 to 20, pull request 9's plan's 34 to
45 where they apply (41, 42, and 43 with `page-items` for `page-front`), and these.

- **46. The path runs past 120 minutes.** The pass parks under `## Stop` in `docs/STATUS.md` on
  the branch, as round standing answer 14 says.
- **47. The scan finds a violation the plan doesn't name.** Fix it in the app's CSS with
  Survey's tokens. A color that fails contrast is tuned within its hue to another Survey token,
  and the PR lists the old and new values. If no token passes, write one follow-up for Survey
  under `docs/follow-ups/` and use the nearest passing token. Never disable an axe rule, and
  never exclude an element from the scan.
- **48. A scan or the crawl still times out.** The scan is already one test per state and theme.
  Split further by page if one still runs long. Never raise the global timeout or add retries. A
  single test may call `test.slow()`.
- **49. Auto mode refuses a step of the install or the check.** Use the alternatives table once.
  If it still refuses, stop running the check: the PR carries the install script's command, the
  eight steps, and their expected output under `## Before merge`, as round standing answer 13
  says, and waits for Geoff's approve. Never edit a settings file or an auto-mode entry.
- **50. The list form of the bar draws an axe or `svelte-check` finding.** Keep the words as
  visually hidden text in each segment and fix the markup the finding names. If the list roles
  can't pass, use `role="img"` with an `aria-label` that names each segment's state and the
  triangle in order, keep the words in the segments for the `textContent` test, and record the
  change under `## Decisions`.
- **51. The widget's state has no color.** The test checks that each widget state's words (the
  word, or the formatted time for `scheduled`) sit in `[data-testid=session-state]`, and that
  its computed color equals the page's base-content ink, so a state color added later without
  words fails.
- **52. The removed drawer changes the look.** Match the old layout by computed widths and the
  screenshot. If a difference remains that the eye can see, keep the plain grid (the walk needs
  it) and list the difference under `## Decisions` for pass 5's polish.
- **53. The install leaves a server that won't start.** Read `journalctl --user -u
  cc-desk-page.service -n 50`. A path or environment fault is fixed in the unit or the script.
  Anything else stops, by standing answer 57, with the journal lines under `## Stop`.
- **54. Something already listens on 7908 when task 4 starts.** Stop; never stop a process you
  didn't start. Write what `ss -ltnp` shows under `## Stop`.
- **55. `@axe-core/playwright` 4.14.0 won't install, or its `axe-core` 4.14.1 breaks.** Pin
  `axe-core` to 4.14.0 (published 2026-10-05) with an `overrides` entry in `package.json` and try
  once more. Failing that, stop.
- **56. No builder in `tests/state.ts` shows a state the words test lists.** Find the verbs that
  reach it on a fixture copy and add a builder for it, as `variant` and `stopped` do; the
  builders are in `web/tests/`, inside this pass's files. If the binary can't reach it from any
  fixture, keep the state in the test's list marked as unreachable with the reason in a comment,
  and record it in `## Planning misses` as `[design-gap]`. Never write a `snapshot.json` by hand.
- **57. A stop after task 4 step 3.** First remove `zz-check.conf` and its folder, then
  `systemctl --user daemon-reload` and `systemctl --user restart cc-desk-page.service`, so the
  unit never points at a folder in `$CLAUDE_JOB_DIR`. Then write `## Stop`.

## Decisions

- [advisor] Done item 8's check is amended, and its words stay as the design has them above it.
  `GET /` returns the shell, and `/api/state` holds the project name, both after install and
  after restart, against a copy of `e2e/testdata/state`, since the live folder doesn't exist. The
  readings passed over were turning on server rendering, which overturns pass 1's `ssr = false`,
  and a headless browser on 7908, which adds a browser to a check the design wrote as `curl`.
  Call: done item 8 reads its project name from `/api/state`, not from the page's HTML, and runs
  against a test state folder. It goes under `## Calls for Geoff` in the PR.
- [advisor] The unit runs a copy of the build in `~/.local/share/cc-desk/page`, never a checkout's
  `web/build`. The worktree is removed after landing, main has no `web/` until pull requests 7
  and 9 land, and every gate run rebuilds `web/build` under a live server. The script redeploys on
  every run, and the README says to run it again after a change lands.
- [advisor] Node is the mise `lts` symlink, named through `%h` so the committed unit holds no home
  path and needs no substitution. `systemd-analyze --user verify` checks the path at install.
- [advisor] The unit sets `HOST=127.0.0.1` itself, since adapter-node binds `0.0.0.0` without it.
- [advisor] `transparent` and `currentColor` aren't color literals: neither names a color, and
  `app.css` uses `transparent` twice. The two words are written out in the test as its only
  allowed keywords.
- [advisor] `@axe-core/playwright` is pinned exactly at 4.14.0 (published 2026-10-15). It pulls
  `axe-core` 4.14.1, published 2026-11-16; no rule bars a new transitive release, and standing
  answer 55 holds the fallback. `standards:dependency-upgrade` covers bumps, not a new
  dependency.
- [planner] The daisyUI drawer goes. Its toggle is an `input` on every page, which done item 4
  forbids, and daisyUI shows the side panel only through that toggle. The mockup has the same
  toggle, so the layout departs from the mockup's markup and keeps its look.
- [planner] The bar becomes a list of segments with their words, in place of `role="img"`.
  Children of an image role are presentational, so words there would pass a `textContent` test
  and reach no screen reader, and axe wouldn't flag it. The triangle reads "Needs Geoff".
- [planner] The scan's two fixes touch pull request 9's call (closed rounds list their passes as
  muted links): those links get a 24px target and stay muted and small in type. If Geoff
  overturns that call, the `.pass-links` rule goes with it.
- [planner] The 405 test builds on pull request 7's call 1 (`csrf.trustedOrigins: ['*']`): without
  it, SvelteKit answers a cross-site form `POST` with 403 before the hook runs. If Geoff
  overturns that call, the walk's method checks fail and the hook needs another route to 405.
- [planner] The words tests build on pull request 7's call 2: a missing `snapshot.json` reads "No
  state yet", which the scan and the banner check use.
- [planner] The file list goes past the round plan's (`web/tests/`, `web/systemd/`,
  `web/scripts/`, the README): the walk needs `+layout.svelte` and `app.css`, the words test needs
  `ScaleBar.svelte` and the two components that pass it a table, the scan's fixes need `app.css`,
  and axe-core needs `package.json` and the lockfile. Each is named in the file list.
- [planner] Done item 8 runs once on the workstation and isn't in the gate: CI has no user
  manager, and no pass after pass 1 edits `scripts/gate.sh`.
- [planner] The unit stays installed and enabled when the pass ends, serving this branch's build
  until a later run of the script replaces it. Call: until the command writes the live state
  folder, the page at http://127.0.0.1:7908/ reads "No state yet". It goes under `## Calls for
  Geoff` in the PR.
- [planner] The Survey gap (the table header's contrast) goes to `docs/follow-ups/` for the
  conductor to carry to Survey's backlog, as `survey-rail-muted-doc-measure.md` did, rather than
  this executor committing to Survey's `main`.
- [planner] Pull request 7's planning misses hold: SvelteKit 3 has no `svelte.config.js`; `#lib`
  imports carry the extension; `Handle` comes from `@sveltejs/kit/hooks`; TypeScript 6 reports
  TS1360 under `satisfies`, so a type test uses a typed `const`.
- [review] One `standards:plan-reviewer` read on 2026-11-18 returned 23 findings and a format
  note, and all are folded: the hex and keyword boundaries, the `scheduled` widget words, the
  rerun after a merge, the bar's roles, the scoped sidebar assert, the draw wait, the `finally`
  and the `EACCES` self-check, the wider source search, the live-folder case, the drop-in on a
  stop, task 5's minutes, one scan test per state and theme, the missing lines below, the call
  channel, the plain `cp` of the fixture, the state-to-builder list, commands written out, the
  previous head's CI verdict, the guarded install paths, an asset in the method checks, a floor
  from the snapshot, and the strings table.

## Planning misses

- [design-gap] Done item 8 says `curl /` returns a page holding a project's name, but pass 1's
  `ssr = false` draws the page in the browser, so the HTML holds none. The check is amended above.
- [design-gap] Done item 4 forbids every `input`, and the mockup's daisyUI drawer puts one on
  every page. The drawer goes.
- [file-list] The round plan's file list for this pass (`web/tests/`, `web/systemd/`,
  `web/scripts/`, the README) leaves out the app files the walk, the scan, and the words test
  need.
- [estimate] The round plan gave this pass 60 minutes; the plan's path is 105, under the
  120-minute park line.
