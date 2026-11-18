# The front page: pass plan

Initiative: cc-desk
Round: page
Pass: page-front

Class `runner`, because the pass changes `scripts/gate.sh` and `.github/workflows/gate.yml`.
Planned 90 minutes. The path below is 95, under the round's 180-minute park line. The State and
rules round's passes took about 0.6 of their planned minutes.

Design: `docs/specs/2026-11-17-page-design.md`, sections "The app", "What each page shows" (the
sidebar, the front page, times, empty fields, themes, and state words), and the first four rows of
"When something goes wrong". Round plan: `docs/plans/2026-11-17-page-round.md`. Where the round
design is silent, `docs/specs/2026-11-17-cc-desk-design.md` holds. Writing a person reads (the PR
body, STATUS lines, follow-ups) follows `~/.claude/docs/voice/written-for-people.md`, loaded
through the `writing-voice` skill. Labels on the page follow Survey's writing standard, through
the `survey` skill (round standing answer 19).

Worktree `/var/home/glw907/Projects/cc-desk/.claude/worktrees/page-front`, branch `page-front`,
cut from `origin/main` at `4e7e78e`. PR base: `main`. Push: `glw907/cc-desk`. The Sessions round
runs beside this one. This pass edits no Go file and nothing under `e2e/`, `internal/`, or
`cmd/`, and it messages none of that round's sessions.

## Pre-flight

No step in this pass needs Geoff. Survey's repo is public, so the install needs no credential.
Every write is inside the worktree, a test's temporary folder, or `$CLAUDE_JOB_DIR/tmp`.

- Permissions checked: `~/.local/state/claude-denials.jsonl`, read again on 2026-11-18 for this
  pass. 130 rows, 2026-11-01T00:55:05Z to 2026-11-18T06:27:23Z, the same rows as the round
  plan's read. Classes and rows: Classifier unavailable 47; Destructive or Unwanted Actions 14;
  Self-Modification 11; Interfere With Workloads 11; server-side classifier, no reason 10;
  Undermining Oversight 5; Unauthorized Access 4; Merge Without Review 3; Auto-Mode Bypass 3;
  Irreversible Local Destruction 2; Git Destructive 2; Create Public Surface 2; Unverifiable
  Deletion Scope 2; Sensitive-Source Provenance 2; Security Weaken 2; Production Deploy 1;
  Secret-Store Writes 1; DNS/Domain/Cert Changes 1; Bot Check Circumvention 1; Code from
  External 1; Shared Scratch Sweep 1; Safety Bypass Flag 1; Out-of-Place Publication 1; PII Data
  Handling 1; Modify Shared Resources 1. This pass runs `npm install` once, `npm ci`, `npm run
  check`, `npm run build`, `npx playwright test`, `go build`, `bash scripts/gate.sh`, `git push
  origin page-front`, `gh pr create`, `gh pr ready`, `gh pr checks`, and `gh run list`. No row
  refuses any of them run bare. The rows near them are in the table below.
- Reviewed: one `standards:plan-reviewer` read of this plan; the `[review]` line under
  `## Decisions` records its verdict and folds.
- Evidence: the mechanism table and the recorded searches below. Each acceptance check was run on
  today's code before any code was written; the outputs are under `## Acceptance`.

### Refused commands and their alternatives

| Refused in the log | This pass does instead |
| --- | --- |
| `npm ci` chained with other installs and piped to `tail` (2026-11-14) | Run `npm ci` alone, bare, from `web/` |
| `node` one-liners and inline scripts during a classifier outage | A script file in `web/scripts/` or `web/tests/`, run by path |
| `kill`, `pkill`, and `kill -9` on test processes (Interfere With Workloads, 6 rows) | The Playwright fixture stops each server it starts. A background command stops with `TaskStop`. A port still held gets a new port, and the old process is left alone |
| `sleep 90` in the foreground (2026-11-18) | Wait with the `Monitor` tool, never a foreground `sleep` |
| A settings edit that adds the Playwright download hosts (Security Weaken, 2026-11-08) | No settings edit. A missing browser parks (standing answer 22) |
| `go run <tool>@latest` (Code from External, 2026-11-09) | No tool fetched at `@latest`. Every package is pinned to an exact version |
| `rm -rf` sweeps of shared temporary folders (Shared Scratch Sweep, Unverifiable Deletion Scope) | Tests make each folder with `mkdtemp` and remove only that folder. No sweep |
| `git stash drop` (2026-11-17) | No stash. A temporary commit sets work aside |
| `gh pr merge`, and `gh pr view` used to judge a merge (Merge Without Review) | The pass merges nothing. It reads checks with `gh pr checks <n>` and `gh run list --branch page-front` |
| A classifier outage, 47 rows | Wait 60 seconds through `Monitor`, then run the command once more |

## Mechanism table

| Mechanism | Source or probe, with what it printed |
| --- | --- |
| Survey installs from a public repo at tag `v1.7.0`, and the plain-app recipe takes the admin recipe's imports | `gh repo view glw907/survey --json visibility` printed `PUBLIC`. `~/Projects/survey/package.json` is `1.7.0`. `~/Projects/survey/DESIGN.md:226` gives the install line `npm install github:glw907/survey#v<version>` and the three Fontsource packages. `DESIGN.md:232` gives the entry sheet's order: Tailwind, the daisyUI plugin, then `survey/theme/fonts.css`, `survey/theme/survey.css`, `survey/theme/components.css`. `DESIGN.md:251-257` is the plain-app recipe |
| `fonts.css` needs three Fontsource packages | `~/Projects/survey/theme/fonts.css:3-9` imports `@fontsource/spectral` (400, 400 italic, 500, 600), `@fontsource-variable/karla` with `wght-italic.css`, and `@fontsource/monaspace-neon/400.css` |
| The lockfile resolves Survey over `git+ssh`, which a runner can't fetch, and one rewrite fixes it | `~/Projects/phonebook/package-lock.json:5298` reads `"resolved": "git+ssh://git@github.com/glw907/survey.git#..."`. `~/Projects/phonebook/.github/workflows/ci.yml:27-29` runs `git config --global url."https://github.com/".insteadOf "ssh://git@github.com/"` before `npm ci`. `gh run list -R glw907/phonebook -L 3` showed three `success` runs, the latest 2026-11-18T06:20Z |
| The mockup's Survey CSS is v1.7.0's, so the app can import the package and copy nothing | `diff -q` of the mockup's `src/lib/survey/` sheets against `~/Projects/survey/theme/` (`survey.css`, `components.css`, `fonts.css`, `survey-vars.css`) printed nothing |
| Playwright 1.65.1 drives the installed `chromium-1255` headless shell | `ls ~/.cache/ms-playwright/` printed `chromium-1248`, `chromium-1255`, and a headless shell for each. A probe under `$CLAUDE_JOB_DIR/tmp/pw` on `@playwright/test` 1.65.1 with no browser path printed `Executable doesn't exist at /nonexistent/chromium_headless_shell-1255/...`, so 1.65.1 wants 1255. `npm view @playwright/test version` printed `1.65.1` |
| A skipped Playwright test exits 0; the JSON report counts it | The same probe with three tests, one `test.skip`: `--list` printed `Total: 3 tests in 1 file`; the run printed `1 skipped`, `2 passed`, and exited 0. The JSON report's `stats` was `{"expected":2,"skipped":1,"unexpected":0,"flaky":0}` |
| A run with no browser fails | The probe with `PLAYWRIGHT_BROWSERS_PATH=/nonexistent` printed `1 failed` and exited 1. The test that opened no page passed, so only tests that open a page fail |
| Pinned versions | `npm view`: `@sveltejs/kit` 3.3.1, `@sveltejs/adapter-node` 6.0.4. `~/Projects/dubplate/web/package.json` pins `@sveltejs/vite-plugin-svelte` 7.3.1, `@tailwindcss/vite` and `tailwindcss` 4.3.5, `daisyui` 5.8.1, `svelte` 5.57.3, `svelte-check` 4.7.8, `typescript` 6.0.3, `vite` 8.3.8, and `playwright` 1.65.1. `node --version` printed `v26.13.1` |
| `actions/setup-node` v7.1.0 is commit `949feb2413d6458794dcd2491c4babbbce0c15c1` | `gh api repos/actions/setup-node/git/ref/tags/v7.1.0` printed that sha, type `commit`. `cairn-cms` and `dubplate` workflows pin the same sha |
| Server rendering would use the real clock, which Playwright can't pin | SvelteKit page options (Svelte MCP, `kit/page-options`, "ssr"): "If you add `export const ssr = false` to your root `+layout.js`, your entire app will only be rendered on the client". So the app renders in the browser only, and every time word comes from the browser's clock |
| `page.clock.setFixedTime` pins `Date` and keeps timers running, so the 2-second poll still fires | <https://playwright.dev/docs/api/class-clock#clock-set-fixed-time>: "Makes `Date.now` and `new Date()` return fixed fake time at all times, keeps all the timers running." `test.use({ timezoneId })` sets the browser's zone (<https://playwright.dev/docs/emulation#locale--timezone>) |
| The real binary writes both files the server reads | `internal/state/commit.go:46-55`: a load failure writes `snapshot-error.json` and leaves every other file alone; a clean load removes it (`:74`); `snapshot.json` is rewritten last (`:140`). `:40-44` and `:232-237`: the error file is `{format, file, message}` |
| A test can print every snapshot it needs by running verbs on a copy of a fixture | `e2e/main_test.go:69-71`: the binary runs as `--state-dir <dir> --now <RFC3339> --usage-command <abs path to e2e/testdata/usage40.sh>`, with `PATH=/usr/bin:/bin`. The probes below ran that way, from a binary built with `go build -o $CLAUDE_JOB_DIR/tmp/probe/cc-desk ./cmd/cc-desk` |
| The state folder default | `cmd/cc-desk/root.go:27-35`: the flag, else `$XDG_STATE_HOME/cc-desk`, else `~/.local/state/cc-desk` |
| `e2e/gate_test.go` runs a copy of `gate.sh` in a scratch repo with no `web/` | `e2e/gate_test.go:27-31` copies `scripts/gate.sh`, `.vale.ini`, and `styles/` only; `:101-106` wants exit 0 and `gate: ok` for the "no skip" case. The page block in the gate must run only when `web/package.json` is tracked, or this Go test fails, and `e2e/` is not this pass's to change |
| A TypeScript type rejects an extra field and a missing field in a JSON value checked with `satisfies` | A probe under `$CLAUDE_JOB_DIR/tmp/sat` with TypeScript 6.0.3 and `--strict`: an unknown field inside an array element gave `TS2353 ... 'bogus' does not exist in type 'P'`; a missing field gave `TS2741 Property 'actual' is missing`; an unknown field in a nested object gave `TS2353`. So a test that writes each printed snapshot into a `.ts` file as `<json> satisfies Snapshot` and runs `tsc --noEmit` fails on both |
| Go tooling would see `.go` files under `web/node_modules` | `find ~/Projects/dubplate/web/node_modules -name '*.go'` printed 0 for a tree with the same packages. Task 1 runs the same `find` on this one |
| The gate passes today, in about 29 seconds | `bash scripts/gate.sh` printed `gate: ok`, exit 0, `real 0m28.863s`, on `4e7e78e` |

### What the binary prints for each test state

Each row ran on a fresh copy of the named fixture, verbs in order, with the clock shown. Times are
UTC in the snapshot; the Alaska column is what the page shows (AKST, UTC-9, on these dates).

| State | Verbs | Printed | Alaska |
| --- | --- | --- | --- |
| `started` | `state`: `round design beta b3 inputs/design-b3.json` at 23:00Z, `pass start alpha p13` at 23:00Z | `next_session` scheduled, time `2026-11-21T05:00:00Z`, range `03:00Z` to `06:30Z`, `rough` false, `question_count` 2, rounds `alpha/r2` and `beta/b3`, projects `alpha` and `beta`. Release `rough`: alpha false, beta true | 8:00 pm, 6:00 pm to 9:30 pm |
| `fast` | `started`, then `pass finish alpha p13` at 23:50Z | time `03:30Z`, range `01:30Z` to `04:00Z` | 6:30 pm, 4:30 pm to 7:00 pm |
| `slow` | `started`, then `pass finish alpha p13` at `2026-11-21T02:20:00Z` | time `06:30Z`, range `05:00Z` to `09:00Z` | 9:30 pm, 8:00 pm to midnight |
| `variant` | `state` with `releases/beta.json`'s one `"planned_minutes": 480` set to 60 in the copy, then as `started` | scheduled, time `2026-11-21T00:00:00Z`, range null, `rough` true | 3:00 pm, no range |
| `big` | `order`: `snapshot` at 23:00Z | `ready_now`, time `2026-11-20T23:00:00Z`, range null, `rough` true, `question_count` 3 | Ready now |
| `guess` | `order`: `pass start north n-free` at 23:00Z, `pass start north n-guess --guess north-q2` at 23:10Z | `open_guesses` `[{north, n-guess, north-q2, 200}]`, `guess_minutes_total` 200, `guess_minutes_limit` 300 | 200 of 300 minutes |
| `two` | `order`: start `north n-free` 23:00Z, `south s-free` 23:01Z, `north n-two` 23:02Z, `south s-two` 23:03Z | `waiting` lists `n-two` then `s-two`, each `limits ["two_pass_limit"]`, `reset_text` null | Waiting, in each pass's row |
| `stale` | `state`: `snapshot` at 23:00Z; `releases/alpha.json`'s `{\n  "format": 1,` made `{\n  "format": 1,,`; `pass start alpha p13` at 23:05Z | Exit 1. `snapshot-error.json` message `releases/alpha.json: invalid character ',' looking for beginning of object key string`. `snapshot.json` keeps `generated_at` `2026-11-20T23:00:00Z` | Last good at 2:00 pm |

Codes across these snapshots: round states `building`, `done`, `not_started`; pass states
`building`, `done`, `not_started`, `waiting`; needs kinds `nothing`, `questions`; answer kinds
`carried`, `ruling`; limit `two_pass_limit`; next-session states `scheduled`, `ready_now`. No
`stop_reason` or `guess_outcome` is set in any of them, so the words test also prints a stopped
pass (`pass stop alpha p13 --reason scope` after `started`).

## Recorded searches

Run with `git grep -l -F` on `4e7e78e`.

| Search | Hits | What each hit is |
| --- | --- | --- |
| `scripts/gate.sh` | 40 | `.github/workflows/gate.yml` (changed here); `README.md:9` ("runs the checks, and fails on any skipped test", still true, left alone); `e2e/gate_test.go` (copies the gate into a scratch repo; the page block is skipped there, see the mechanism table); 37 plans, prompts, specs, research, and spike files that quote the gate, history, left alone |
| `gate.yml` | 10 | Plans and prompts that name the workflow. History, left alone |
| `web/` | 7 | The round's plans, prompts, and the page design. None is code |
| `node_modules` | 1 | `.gitignore`, which already ignores `node_modules/` at any depth |
| `playwright` | 1 | `docs/plans/2026-11-17-page-round.md`. No config or code yet |
| `7908` | 5 | Plans and the design. The port is pass 3's |
| `CC_DESK_STATE` | 0 | The environment variable this pass adds is new |
| `snapshot-error.json` | 10 | The Go writer (`internal/state/commit.go`, its test, `e2e/broken_test.go`), the designs, plans, and one follow-up. This pass reads the file and writes nothing |

`.gitignore` holds `node_modules/` and `.svelte-kit/`, which cover `web/` at any depth. It lacks
`web/build/`, `web/test-results/`, and `web/playwright-report/`. `.vale.ini` lints `[*.md]`, but
the gate passes `git ls-files 'docs/*.md'` only, so nothing in `web/` reaches vale and `.vale.ini`
needs no change.

## File list

| File | Change |
| --- | --- |
| `web/package.json`, `web/package-lock.json` | New. Exact versions from the mechanism table, `@playwright/test` 1.65.1, Survey from `github:glw907/survey#v1.7.0`, the three Fontsource packages, and `@lucide/svelte` if an icon is drawn |
| `web/svelte.config.js`, `web/vite.config.ts`, `web/tsconfig.json`, `web/src/app.html`, `web/src/app.css`, `web/src/app.d.ts` | New. The Node adapter; `plugins: [tailwindcss(), sveltekit()]` as the mockup's `vite.config.ts`; the entry sheet in the recipe's order, then the app's own token-only rules |
| `web/src/hooks.server.ts` | New. Every method but `GET` and `HEAD` gets 405 |
| `web/src/lib/server/state.ts` | New. Finds the folder, reads both files read-only on each call, keeps the last good snapshot in memory |
| `web/src/routes/api/state/+server.ts` | New. The poll endpoint: `GET` answers with the snapshot, the error, and the last good time |
| `web/src/lib/snapshot.ts` | New. The snapshot type, mirroring `internal/snapshot/snapshot.go` and the ruling in `internal/state/types.go` |
| `web/src/lib/words.ts` | New. The one table of words for every code |
| `web/src/lib/time.ts` | New. Formats a time in `America/Anchorage` from the browser's clock |
| `web/src/lib/poll.svelte.ts` | New. The 2-second poll, the connection state, and the time of the last answer |
| `web/src/routes/+layout.ts`, `+layout.svelte`, `+page.svelte` | New. `ssr = false`; the sidebar; the front page; the tab title |
| `web/src/lib/*.svelte` | New. The components; the split is the executor's |
| `web/playwright.config.ts` | New. `retries: 0`, `forbidOnly: true`, the JSON reporter to a file, a global setup |
| `web/tests/global-setup.ts`, `web/tests/state.ts`, `web/tests/app.ts` | New. Build the binary once; copy a fixture and run verbs on it; start and stop the built app per test |
| `web/tests/*.spec.ts` | New. The tests named in the tasks |
| `web/scripts/check-playwright-report.sh` | New. Fails unless the report's passed count equals the listed count and skipped, unexpected, and flaky are 0 |
| `scripts/gate.sh` | The page block after vale and before `gate: ok`, run only when `web/package.json` is tracked. The Go checks are unchanged |
| `.github/workflows/gate.yml` | `setup-node`, the `insteadOf` rewrite, `npm ci` in `web/`, the Chromium install, and a longer timeout |
| `.gitignore` | `web/build/`, `web/test-results/`, `web/playwright-report/` |

## Acceptance

The four items, word for word from the page design. Before any code was written, `ls web` in the
worktree printed `ls: cannot access 'web': No such file or directory`, so the checks for items 1,
3, and 5 can't run yet. They are `npx playwright test` in `web/`.

1. **Estimate, on the page.** The browser's clock is pinned to the fixture's time and its zone
   set to one that isn't Alaska's. With the fixture of two releases from the State and rules
   round, the widget shows the next session's time and range as the Alaska times that round's
   test states (8:00 pm, 6:00 pm to 9:30 pm), with its question count and projects. The front
   page shows the same figures. With the fixture where a pass finished in half its planned
   minutes, the widget reads 6:30 pm with 4:30 pm to 7:00 pm, and the slow one reads 9:30 pm
   with 8:00 pm to midnight. "Rough" shows beside the release whose `rough` is set and beside no
   other, and the widget shows it only in the variant where `next_session.rough` is set.
   - Met by task 4. Shown by `npx playwright test tests/estimate.spec.ts` in `web/`, run by the
     gate.
   - Before: `web/` doesn't exist. The binary prints the times in the table above.
3. **Waiting order, on the page.** Against the State and rules round's order fixtures: with only
   the 400-minute dependent pass, the widget reads "Ready now". After the 200-minute guess
   start, the front page lists the guess with "200 of 300 minutes", and a snapshot edited by the
   test to a total of 250 reads "250 of 300", so the page isn't adding the figures itself. A pass
   held by the two-pass limit reads "Waiting" with that reason, in that pass's row.
   - Met by task 4. Shown by `npx playwright test tests/order.spec.ts`.
   - Before: `web/` doesn't exist. The binary prints the `big`, `guess`, and `two` rows above.
5. **Redraw.** A test replaces `snapshot.json` while a page is open. The page shows the new
   state within 5 seconds with no reload. Adding `snapshot-error.json` shows a banner that reads
   "Stale" with the message and the last good snapshot's time, with that snapshot still drawn
   under it, and removing the file clears the banner. Replacing `snapshot.json` with text that
   isn't JSON shows the same banner over the last snapshot. A server started with no snapshot
   reads "No state yet". Stopping the server shows "Not connected" within 10 seconds, and
   starting it again clears that.
   - Met by task 5. Shown by `npx playwright test tests/redraw.spec.ts`.
   - Before: `web/` doesn't exist. The `stale` row above is what the binary writes.
10. **The gate.** `bash scripts/gate.sh` passes, and it now runs the page's checks. A
    Playwright run with a skipped test, with no browser, or with fewer tests than the suite lists
    fails the gate.
    - Met by tasks 1 and 6. Shown by `bash scripts/gate.sh` printing the page steps and `gate:
      ok`, and by the three failing runs in task 6.
    - Before: `bash scripts/gate.sh` printed `gate: ok`, exit 0, with no page step.

### The strings the tests write out

The browser zone is `Europe/London` in every page test. At these instants London's date differs
from Alaska's, so a page that takes "Today" or the hour from the browser's zone fails. The clock
is pinned to the snapshot's `generated_at` unless the row says otherwise.

| State | Widget | Tab title | Elsewhere |
| --- | --- | --- | --- |
| `started` | "Today, about 8:00 pm", "6:00 pm to 9:30 pm", 2 questions, 2 rounds, alpha and beta; no "Rough" | "Today, about 8:00 pm · cc-desk" | The front page shows the same time, range, count, and projects. "Rough" beside beta in the sidebar and on its board, and not beside alpha |
| `fast` | "Today, about 6:30 pm", "4:30 pm to 7:00 pm" | | |
| `slow` | "Today, about 9:30 pm", "8:00 pm to midnight" | | |
| `variant` | "Today, about 3:00 pm", "Rough", the range as an en dash with "None" for a screen reader | | |
| `big` | "Ready now" | "Ready now · cc-desk" | |
| `guess` | | | The front page: n-guess with "200 of 300 minutes"; after the test sets `guess_minutes_total` to 250, "250 of 300 minutes" |
| `two` | | | The rows of n-two and s-two: "Waiting" and the two-pass limit's words |
| `stale` | | | "Stale", `releases/alpha.json: invalid character ',' looking for beginning of object key string`, and "2:00 pm", over the `started`-era snapshot |

## Tasks

The path is the sum, 95 minutes, since the tasks run in order. Each task commits by path, runs
the gate bare before the commit, and pushes. Each task's report names the gate run's conclusion
on its pushed commit from `gh run list --branch page-front`, or "no run yet" with the link.

### Task 1: the empty app and the gate (25 minutes)

Outcome: `web/` builds an empty SvelteKit app that follows Survey's plain-app recipe, and the
gate and CI run its checks before anything is drawn.

- Write `web/package.json` with the exact versions, then run `npm install` once, bare, from
  `web/`. Survey goes in as the recipe's line, `github:glw907/survey#v1.7.0`. Every later install
  is `npm ci`.
- The entry sheet imports Tailwind, the daisyUI plugin, then Survey's `fonts.css`, `survey.css`,
  and `components.css`. The app carries no copy of Survey's CSS. The body takes `font-body
  bg-base-100 text-base-content` as the mockup's `app.html` does.
- The root `+layout.ts` sets `ssr = false`. The page reads "cc-desk" and nothing else.
- `playwright.config.ts`: `retries: 0`, `forbidOnly: true`, one Chromium project, a JSON report
  to a file in `test-results/`, and `webServer` left out, since the fixture in task 2 starts the
  app.
- One test, `tests/empty.spec.ts`, opens the built app and finds "cc-desk".
- `scripts/gate.sh` gains the page block after vale and before `gate: ok`, run only when
  `git ls-files --error-unmatch web/package.json` succeeds. It runs, in `web/`, `npm ci`,
  `npm run check`, `npm run build`, `npx playwright test --list` (keeping the count from its
  `Total:` line, singular or plural), then `npx playwright test`, then
  `web/scripts/check-playwright-report.sh <report> <count>`. A missing or zero count fails. The
  header comment names the page steps.
- `web/scripts/check-playwright-report.sh` fails, naming the figure, unless `stats.expected`
  equals the count and `stats.skipped`, `stats.unexpected`, and `stats.flaky` are 0. It uses
  `jq`, which the gate already requires.
- `gate.yml`: after `setup-go`, `actions/setup-node@949feb2413d6458794dcd2491c4babbbce0c15c1 #
  v7.1.0` with `node-version: 26`; the step `git config --global url."https://github.com/".insteadOf
  "ssh://git@github.com/"` with phonebook's comment; `npm ci` in `web/`; then `npx playwright
  install --with-deps chromium` in `web/`. `timeout-minutes` goes to 30.
- `.gitignore` gains the three `web/` paths.
- Run `find web/node_modules -name '*.go' | wc -l` and record it. A count above 0 takes
  standing answer 24.

Files: `web/` config files, `web/src/app.*`, `web/src/routes/+layout.ts`, `+page.svelte`,
`web/playwright.config.ts`, `web/tests/empty.spec.ts`, `web/scripts/check-playwright-report.sh`,
`scripts/gate.sh`, `.github/workflows/gate.yml`, `.gitignore`. Class `runner`.

Test: `tests/empty.spec.ts`, and `go test -count=1 ./e2e -run TestGateFailsOnSkip`, which must
still pass with the new gate.

Acceptance check: `bash scripts/gate.sh` prints `gate: web install`, `gate: web check`, `gate:
web build`, `gate: web tests`, and `gate: ok`. The draft PR exists, and the gate run on this
task's commit is `success` in `gh run list --branch page-front`. That green run proves the
install, the browser, and the rewrite on a runner before anything is drawn.

### Task 2: the server, the snapshot type, and the test harness (20 minutes)

Outcome: the server reads the state folder read-only and answers the poll, and the snapshot type
matches what the real binary prints.

- `src/lib/server/state.ts` takes the folder from `CC_DESK_STATE_DIR`, else
  `$XDG_STATE_HOME/cc-desk`, else `~/.local/state/cc-desk`. On each call it reads
  `snapshot.json` and `snapshot-error.json` with a read-only call. It keeps the last snapshot it
  parsed in memory. It opens no file for writing and runs no command.
- The poll endpoint at `/api/state` answers `GET` with four parts: the snapshot to draw (the
  file's, or the last good one), the error file's message or the app's parse-error text, the last
  good snapshot's `generated_at`, and whether no snapshot has ever been read. The app's
  parse-error text starts with a fixed phrase the app writes, such as "snapshot.json isn't valid
  JSON:", followed by the parser's message. Tests match the fixed phrase only, since the parser's
  wording changes with Node.
- `hooks.server.ts` answers 405 to every method but `GET` and `HEAD`.
- `src/lib/snapshot.ts` mirrors every field of `internal/snapshot/snapshot.go`'s `Snapshot`,
  `Release`, `Round`, `Pass`, `Question`, `NextSession`, `WaitEntry`, `WaitingPass`, and
  `OpenGuess`, and the ruling in `internal/state/types.go`, with `null` where Go has a pointer and
  `omitempty` fields optional. Codes are `string`, so a new code from the Sessions round still
  type-checks.
- The harness: `tests/global-setup.ts` runs `go build -o <temp>/cc-desk ./cmd/cc-desk` from the
  repo root once. `tests/state.ts` copies a fixture from `e2e/testdata/` into a fresh `mkdtemp`
  folder and runs verbs on it as the mechanism table says, with the usage stand-in and `PATH` of
  `/usr/bin:/bin`. `tests/app.ts` is a Playwright fixture that starts `node build` on a free port
  with `HOST=127.0.0.1` and `CC_DESK_STATE_DIR` set to the test's folder, waits for it to answer,
  and offers `stop()` and `start()` on the same port. Teardown stops it and removes the folder.
- `tests/snapshot-type.spec.ts` prints every state in the binary table, plus a stopped pass,
  writes each as a `.ts` file holding `<json> satisfies Snapshot` under a temporary folder, and
  runs `tsc --noEmit` on them. It fails with tsc's output.
- `tests/server.spec.ts`: `POST` and `PUT` to `/` and `/api/state` get 405, `HEAD` gets 200.

Files: `web/src/lib/server/state.ts`, `web/src/routes/api/state/+server.ts`,
`web/src/hooks.server.ts`, `web/src/lib/snapshot.ts`, `web/tests/global-setup.ts`,
`web/tests/state.ts`, `web/tests/app.ts`, `web/tests/snapshot-type.spec.ts`,
`web/tests/server.spec.ts`, `web/playwright.config.ts`. Class `ordinary`.

Acceptance check: `npx playwright test tests/snapshot-type.spec.ts tests/server.spec.ts` passes.
Then, once each and reverted before the commit, remove one field from the type and add one field
to it. Each run fails with tsc's `TS2353` and `TS2741` lines; paste both into the task's report.

### Task 3: the words and the times (10 minutes)

Outcome: one table words every code, and one formatter words every time.

- `src/lib/words.ts` holds every code from "Key codes" below. The five pass words and four widget
  states are the design's, exact. The rest follow Survey's writing standard. An unknown code
  shows with its underscores turned to spaces. `next_session.reasons` shows as written.
- `src/lib/time.ts` formats in `America/Anchorage` to the minute, as "8:00 pm", with 12:00 am as
  "midnight". "Today" and "Tomorrow" come from comparing Alaska dates of the browser's clock and
  the time. A range shows its two times, bare, joined by "to". A null shows an en dash with
  "None" as its screen-reader text.
- `tests/words.spec.ts` collects every code from every snapshot task 2's type test prints and
  fails naming each code with no entry. It also checks the five pass words exactly.

Key codes, from `internal/snapshot/snapshot.go` and `internal/state/types.go`: pass states
`not_started`, `waiting`, `building`, `stopped`, `done`; round states `not_designed`,
`not_started`, `building`, `done`; stop reasons `usage`, `irreversible`, `secret`, `auth_data`,
`scope`; limits `two_pass_limit`, `usage`; needs kinds `questions`, `nothing`, `not_designed`;
guess outcomes and answer kinds `ruling`, `carried`; next-session states `needs_you_now`,
`ready_now`, `scheduled`, `nothing_waiting`. The executor checks this list against the Go source
as it stands and adds any code it finds.

Files: `web/src/lib/words.ts`, `web/src/lib/time.ts`, `web/tests/words.spec.ts`. Class
`ordinary`.

Acceptance check: `npx playwright test tests/words.spec.ts` passes. Remove one entry and it fails
naming that code; paste the line into the task's report and restore the entry.

### Task 4: the sidebar and the front page (25 minutes)

Outcome: the sidebar and the front page draw the snapshot as the design's "What each page shows"
says, laid out as the mockup is.

- Read the mockup first: `docs/specs/2026-11-17-cc-desk-mockups/src/routes/+layout.svelte`,
  `+page.svelte`, `src/lib/ReleaseBlock.svelte`, `ReleaseBoard.svelte`, `ScaleBar.svelte`,
  `mock.css`, `app.css`, and `screens/front-page.png`. The mockup settles the layout. Its rules
  that use tokens only may be carried over; its copy of Survey stays behind, and so do "Reports
  to read", "In review", and "Stalled".
- The sidebar, about 23rem wide, in the design's order: the next-session widget (the only box),
  "Waiting for you" (each line links to `/<project>/<round>`, which pass 2 builds; empty, it
  reads "Nothing"), and "Running" with one block per release. The scale bar's segments grow by
  planned minutes. "Rough" sits beside the release name when its `rough` is set.
- The front page: what the next session covers (each round, its question texts, and its needed
  time), then a board per release with running rounds opened to their passes. Waiting passes
  show their limits' words and the reset text when there is one. Open guesses show their minutes
  and "N of 300 minutes" from `guess_minutes_total` and `guess_minutes_limit`.
- The tab title is the widget's state words, then " · cc-desk".
- State words sit in the same element as their color.
- `tests/estimate.spec.ts` and `tests/order.spec.ts` assert the strings in the table above,
  written out in the test. The `variant` state edits `releases/beta.json` in the test's copy of
  the fixture, never in `e2e/`. The `guess` test edits `guess_minutes_total` in the copy's
  `snapshot.json`, then waits for the redraw.

Files: `web/src/routes/+layout.svelte`, `web/src/routes/+page.svelte`, `web/src/lib/*.svelte`,
`web/src/app.css`, `web/tests/estimate.spec.ts`, `web/tests/order.spec.ts`. Class `ordinary`.

Acceptance check: `npx playwright test tests/estimate.spec.ts tests/order.spec.ts` passes. Take
one screenshot of the `started` front page in light and one in dark, and put them in the PR.

### Task 5: the poll and the banners (10 minutes)

Outcome: the page redraws within 5 seconds of a change and shows the four banners.

- `src/lib/poll.svelte.ts` fetches `/api/state` every 2 seconds with a 5-second timeout. A
  failed fetch shows "Not connected" with the time of the last answer and keeps the last state
  drawn. The next good answer clears it.
- The banners, from the design's table: "Stale" with the message and the last good time over the
  last good snapshot; "No state yet" alone; the same "Stale" banner for text that isn't JSON; and
  "Not connected".
- `tests/redraw.spec.ts`, one test per sentence of done item 5. The replace step runs `pass
  start alpha p13` on the open `started`-era copy and waits for the new state. The stale step
  breaks `releases/alpha.json` as the `stale` row says and runs a verb. The clear step restores
  the file and runs `snapshot`, which removes the error file. The non-JSON step writes text over
  `snapshot.json`. "No state yet" starts the app on a copy of `order`, which has no
  `snapshot.json`. The stop step calls the fixture's `stop()` and `start()`.
- The 5-second and 10-second limits are the tests' timeouts for those assertions.

Files: `web/src/lib/poll.svelte.ts`, `web/src/routes/+layout.svelte`, `web/src/lib/*.svelte`,
`web/tests/redraw.spec.ts`. Class `ordinary`.

Acceptance check: `npx playwright test tests/redraw.spec.ts` passes.

### Task 6: the gate's failures, the review, and the PR (5 minutes)

Outcome: done item 10 has its evidence, the branch is reviewed, and the PR is ready.

1. `bash scripts/gate.sh`, bare: paste the page lines and `gate: ok`.
2. With one test changed to `test.skip`, uncommitted: `bash scripts/gate.sh` fails naming the
   skipped figure. Revert.
3. `PLAYWRIGHT_BROWSERS_PATH=/nonexistent bash scripts/gate.sh` fails at the Playwright run.
4. A short run: `npx playwright test --grep redraw` from `web/`, then
   `bash web/scripts/check-playwright-report.sh <its report> <the full listed count>` fails naming
   the count.
5. Rebase on `origin/main`; a Sessions snapshot change takes standing answer 16.
6. One whole-branch `standards:diff-reviewer` read, folded. Then the PR body and ready.

Paste each command and the lines it printed under `## Acceptance` in the PR.

Files: none beyond fixes. Class `runner`.

## Standing answers

The twelve in `round-run.md` hold, with the round plan's 13 to 20, and these.

- **21. Survey won't install.** If `npm install` is refused, write the dependency in
  `package.json` by hand and run `npm install` alone once more. If CI can't fetch it, the
  `insteadOf` step is missing or misplaced: it goes before every `npm ci`. Survey is public, so
  this is never a credential stop.
- **22. No browser for the pinned Playwright.** Pin `@playwright/test` to the version whose
  Chromium is in `~/.cache/ms-playwright/` (1.65.1 wants 1255, 1.64.1 wants 1248). If neither
  works, the pass parks under `## Stop` in `docs/STATUS.md`: no settings edit, and no browser
  download outside CI.
- **23. The Go gate test fails on the new gate.** The page block must be skipped when
  `web/package.json` isn't tracked, as in `e2e/gate_test.go`'s scratch repo. Fix the gate;
  never change `e2e/`.
- **24. `go vet ./...` or `go test ./...` reaches `.go` files under `web/node_modules`.** Feed the
  Go checks `go list ./...` with any path under `web/` taken out, and record it under
  `## Decisions`. The Go checks are otherwise unchanged.
- **25. The binary prints a figure other than the table's.** The design's strings win for items
  1 and 3. Rebase first; if the Sessions round changed the estimate, record the new figure, put
  it in the test, and write it under `## Decisions` with the commit that changed it.
- **26. A field the page wants isn't in the snapshot.** Round standing answer 15: draw what the
  snapshot has and file a follow-up for the Desk and skills round.
- **27. The design's sentences pull two ways.** "The front page repeats nothing from the sidebar"
  and item 1's "The front page shows the same figures" are both met by the front page stating
  the time, range, count, and projects in its heading prose, as the mockup's "Your next decision
  session: tomorrow, about 9:30 am" does, and repeating no widget, list, or bar.
- **28. A `ready_now` session with `rough` set.** The widget reads "Ready now", shows the time
  only if the state is `scheduled`, and shows "Rough" whenever `next_session.rough` is true.
  Only the `variant` test asserts "Rough" in the widget; the `started` test asserts its absence.
- **29. A timing test is flaky.** The 5-second, 10-second, and 2-second figures are the design's
  and don't change. Find the cause (a fixed port, a missed `await`, a server not ready) and fix
  it. `retries` stays 0.
- **30. CI runs past 30 minutes.** Raise `timeout-minutes` to 45 and say so under
  `## Decisions`. Past 45, the cause is a hang: trace it.
- **31. Survey lacks a token or state the mockup needs.** Use the nearest token, name the gap in
  a CSS comment, and file one follow-up under `docs/follow-ups/` naming the rule and what was
  seen. No literal values.
- **32. The path runs past 180 minutes.** Round standing answer 14: the pass parks under
  `## Stop` in `docs/STATUS.md`.

## Decisions

- [planner] The app renders in the browser only (`ssr = false`), so every time word comes from a
  clock Playwright can pin. The page is a local single-reader status page, where server rendering
  buys nothing.
- [planner] The poll is a JSON endpoint fetched every 2 seconds, never `invalidateAll`, so a
  stopped server shows a banner rather than SvelteKit's error page.
- [planner] The test zone is `Europe/London`, where the date differs from Alaska's at the fixture
  instants.
- [planner] Test state comes from running the real binary on copies of `e2e/testdata/` fixtures.
  The only hand edits are the two the design names: the variant's planned minutes, and the
  `guess_minutes_total` of 250. The non-JSON text and the broken release file are the inputs
  item 5 names.
- [planner] The count check lives in `web/scripts/check-playwright-report.sh`, so task 6 can feed
  it a short report directly. The gate's page block runs only when `web/package.json` is tracked,
  because `e2e/gate_test.go` runs the gate in a scratch repo without `web/`.
- [planner] The stale banner's message is the binary's text, which names a path relative to the
  state folder, so the test writes it out whole. The non-JSON message is the app's fixed phrase
  plus Node's parser text, and the test matches the phrase only.
- [planner] No `[advisor]` line in the State and rules PRs (#3 to #6) changes this pass. Two it
  builds on: `ready_now` can carry a time and a null range (#4), and `next_session.reasons` is
  `[]` when empty (#5).

## Planning misses

None yet.
