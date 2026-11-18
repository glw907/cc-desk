# The terminal link: pass plan

Initiative: cc-desk
Round: page
Pass: page-link

Class `runner`: the pass registers a desktop handler on the workstation. Planned 30 minutes. The
path below is 52, under the 60-minute line (twice the planned minutes, round standing answers 14
and 17). The plan reviewer timed CI's gate at 2.7 to 3.9 minutes a run, so the path counts
three full gates, and each task otherwise runs only the specs it touches.

Design: `docs/specs/2026-11-17-page-design.md`, "The terminal link", "What each page shows" (the
command line on a round page and a pass page), "Steps only Geoff can do", and done items 9 and
10. Round plan: `docs/plans/2026-11-17-page-round.md`. Where the round design is silent,
`docs/specs/2026-11-17-cc-desk-design.md` holds. Writing a person reads (the PR body, STATUS
lines, follow-ups) follows `~/.claude/docs/voice/written-for-people.md`, loaded through the
`writing-voice` skill. Labels on the page follow Survey's writing standard, through the `survey`
skill (round standing answer 19).

Worktree `/var/home/glw907/Projects/cc-desk/.claude/worktrees/page-link`, branch `page-link`,
cut from `origin/page-surface` at `31b712f` (pull request 11, open, which carries pull requests 9
and 7, open). The branch tracks `origin/page-surface`, so every push names its target:
`git push -u origin page-link`, then `git push origin page-link`. A bare `git push` would land on
pull request 11's branch. PR base: `main`. Before the PR is ready, merge `origin/page-surface` in
again if it moved, and `origin/main` if main moved. The Sessions round runs beside this one. This
pass edits no Go file, nothing under `e2e/`, `internal/`, or `cmd/`, and neither
`scripts/gate.sh` nor the workflow, and it messages none of that round's sessions.

## Pre-flight

No step in this pass needs Geoff before it runs. His one step, the Chromium click, comes after
it, as the design's "Steps only Geoff can do" says (see `## Decisions`). The auto-mode entry
`cc-desk-sessions-page` (`~/.claude/settings.json:426`) reads: "The sessions
`cc-desk-page-surface` and `cc-desk-page-link` may install and enable the user unit
cc-desk-page.service, listening on 127.0.0.1:7908, and register a desktop handler for the
cc-desk:// scheme." Every other write is inside the worktree, a test's `mkdtemp` folder,
`$CLAUDE_JOB_DIR/tmp`, `~/.local/share/cc-desk/link/`, or the one file
`~/.local/share/applications/cc-desk.desktop` and the cache `update-desktop-database` rewrites
beside it.

- Permissions checked: `~/.local/state/claude-denials.jsonl`, read again on 2026-11-18 for this
  pass over 2026-10-19 to 2026-11-18. 130 rows, 2026-11-01T00:55:05Z to 2026-11-18T06:27:23Z,
  the same rows passes 1 to 3 read. Classes and rows: Classifier unavailable 47; Destructive or
  Unwanted Actions 14; Self-Modification 11; Interfere With Workloads 11; server-side classifier,
  no reason 10; Undermining Oversight 5; Unauthorized Access 4; Merge Without Review 3; Auto-Mode
  Bypass 3; Irreversible Local Destruction 2; Git Destructive 2; Create Public Surface 2;
  Unverifiable Deletion Scope 2; Sensitive-Source Provenance 2; Security Weaken 2; Production
  Deploy 1; Secret-Store Writes 1; DNS/Domain/Cert Changes 1; Bot Check Circumvention 1; Code
  from External 1; Shared Scratch Sweep 1; Safety Bypass Flag 1; Out-of-Place Publication 1; PII
  Data Handling 1; Modify Shared Resources 1. This pass runs `npm ci`, `npm run check`, `npm run
  build`, `npx playwright test`, `bash scripts/gate.sh`, `bash web/scripts/install-link.sh`
  (`install`, `desktop-file-validate`, `update-desktop-database`), `bash
  web/scripts/check-link.sh` (`xdg-open`, which starts the handler and through it `ptyxis -x`),
  `xdg-mime query default`, `jq` reading Chromium's `Preferences` file, `git push -u origin
  page-link`, `gh pr create --draft`, `gh pr ready`, `gh pr checks`, and `gh run list`. A search
  of the rows' targets for `xdg-open`, `xdg-mime`, `xdg-settings`, `update-desktop-database`,
  `desktop-file-validate`, `ptyxis`, `.desktop`, `applications`, and `Preferences` found none.
  The near rows are the ones pass 3 named: one `npm ci` chained and piped to `tail`, `node`
  one-liners in classifier outages, `kill` and `pkill` of test processes, and one settings `jq`
  edit (a write, unlike this pass's read). No row refuses any of this pass's commands run bare.
- Reviewed: one `standards:plan-reviewer` read before code, as the `runner` class sets
  (`pass-core` "Plan review lens"), folded on 2026-11-18; its `[review]` lines are under
  `## Decisions`. The close gets one whole-branch `standards:diff-reviewer` read.
- Evidence: the mechanism table and the recorded searches below. Every acceptance check was run
  on today's code before any code was written. The outputs are under `## Acceptance`.

### Refused commands and their alternatives

| Refused in the log | This pass does instead |
| --- | --- |
| `npm ci` chained with other installs and piped to `tail` (2026-11-14) | Run `npm ci` alone, bare, from `web/` |
| `node` one-liners and inline scripts during a classifier outage | A script file, run by path: `web/scripts/install-link.sh` and `web/scripts/check-link.sh` |
| `kill`, `pkill`, and `kill -9` on test processes (Interfere With Workloads) | The Playwright fixture stops each server it starts. No `ptyxis` window is closed by force: the stand-in exits at once, and a window that stays open is left for Geoff (standing answer 61) |
| `sleep` in the foreground | `check-link.sh` waits in a bounded loop on its record file. The session waits with the `Monitor` tool |
| `rm -rf` sweeps of shared temporary folders | Tests remove only their own `mkdtemp` folder. The install script replaces only `~/.local/share/cc-desk/link/` and `~/.local/share/applications/cc-desk.desktop`, each by name |
| `git stash drop` | No stash. A temporary commit sets work aside |
| `gh pr merge`, and `gh pr view` used to judge a merge (Merge Without Review) | The pass merges nothing. It reads checks with `gh pr checks <n>` and `gh run list --branch page-link` |
| A classifier outage, 47 rows | Wait 60 seconds through `Monitor`, then run the command once more |
| Any refusal of the install or of `check-link.sh` | Standing answer 59 |

## Mechanism table

| Mechanism | Source or probe, with what it printed |
| --- | --- |
| No `cc-desk://` handler exists today | `xdg-mime query default x-scheme-handler/cc-desk` printed nothing. `gio mime x-scheme-handler/cc-desk` printed `No default applications for “x-scheme-handler/cc-desk”`. `xdg-open cc-desk://alpha/r1` printed `gio: cc-desk://alpha/r1: Operation not supported` and exited 4. `ls ~/.local/share/applications` holds no `cc-desk` file (2026-11-18) |
| `update-desktop-database` alone makes a `.desktop` file the scheme's default, with no `xdg-mime default` and no write to `mimeapps.list` | Planning probe with a throwaway scheme `ccdeskprobe`: a `.desktop` file with `MimeType=x-scheme-handler/ccdeskprobe;` in `~/.local/share/applications`, then `update-desktop-database ~/.local/share/applications`. Then `xdg-mime query default` printed `ccdeskprobe.desktop`, `xdg-settings get default-url-scheme-handler ccdeskprobe` printed `ccdeskprobe.desktop`, and `gio mime` named it as the default. `~/.config/mimeapps.list` held 0 lines naming it. The file was removed and the database rebuilt; `cmp` showed `mimeinfo.cache` as before |
| `xdg-open` hands the handler the address unchanged, as one argument, with the caller's environment | Planning probe: a `ccdeskprobe` handler that recorded its arguments and environment. `alpha/../r1`, `%3B`, `;x`, `?x=1`, and `Alpha` all arrived as typed, with no decoding or folding. The caller's `PATH` prefix and a marker variable reached the handler. Every `xdg-open` exited 0, and the record was written within about 1 second |
| `ptyxis -x "<command>"` runs the command in the running `ptyxis` | Planning probe: `ptyxis -x "<tmp>/bin/standin alpha r1"`, with the stand-in named by absolute path, exited 0, and the stand-in recorded `alpha r1` within 0.5 seconds. The running instance is `/usr/bin/ptyxis --gapplication-service`, `ptyxis` 50.1. The command runs in that service, so the handler's `PATH` doesn't reach it: the handler names the command by absolute path, and the stand-in's record path is written in its body, never passed by environment |
| `cc-desk` isn't on any `PATH` yet | `command -v cc-desk` exited 1 in a login shell. A user unit's `PATH` is `/usr/lib64/ccache:/usr/local/sbin:/usr/local/bin:/usr/bin` |
| The tools the install runs are present | `desktop-file-validate`, `update-desktop-database`, `xdg-mime`, `xdg-settings`, and `/home/linuxbrew/.linuxbrew/bin/gio` are on `PATH`. The desktop is GNOME on Wayland |
| Prior art for a scheme handler | `~/.local/share/applications/claude-code-url-handler.desktop`: `Exec="<absolute path>" --handle-uri %u`, `Type=Application`, `NoDisplay=true`, `MimeType=x-scheme-handler/claude-cli;` |
| Project names may carry `.` and `_`; round names have no rule | `internal/state/events.go:13` sets `projectName` to the pattern `^[a-z0-9][a-z0-9._-]*$`, checked in `AddRelease` (`events.go:39`). No pattern checks a round's `Name` (`types.go:40`, `49`, `63`). No project or round name in `e2e/testdata` fails `^[a-z0-9-]+$` (a grep of every `"project"` and `"name"` value printed none) |
| Both item pages draw the command through one component | `web/src/lib/Command.svelte`: `<h2>Command</h2>` then `<p><code>cc-desk {project} {round}</code></p>`, props `project` and `round`. Used at `routes/[project]/[round]/+page.svelte:81` and `routes/[project]/[round]/[pass]/+page.svelte:77` as `<Command project={release.project} round={round.name} />`. The design's "What each page shows" puts the terminal link beside the command on both pages |
| Each "Waiting for you" line is already a link | `web/src/lib/Sidebar.svelte`: `{#each snapshot.waits_for_geoff as wait …}<a class="rail-row" href="/{wait.project}/{wait.round}">{wait.project}/{wait.round}<span class="sub">…</span></a>`. A link can't hold a link, so the terminal link sits beside the row's link, inside a row element |
| The row styles key on `a.rail-row` | `web/src/app.css:52` `.rail-row`, `:60` `a.rail-row:hover`, `:63` `.rail-row .sub`. The layout's Rulings link also carries `rail-row` |
| The tests that read these places | `items.spec.ts` expects `main code` to read `cc-desk alder r1`, `cc-desk birch r1`, or `cc-desk alpha r2`: one `code` in `main`. Its `main a[href]` check needs the item's own addresses among the `href` values; more links pass. `sidebar.spec.ts` expects `waits.getByRole('link')` count 0 on an empty list, and links named `north/n1 2 questions` (`href` `/north/n1`) and `south/s1 1 question` (`href` `/south/s1`) on `big` |
| The walk allows one button name | `web/tests/surface.spec.ts` counts every `getByRole('button')` whose accessible name isn't exactly `Copy command` and fails on any |
| The source search reads every file under `web/src` | `surface.spec.ts` fails on a call matched as `\bexec\s*\(` and others, anywhere under `web/src`, and on `open(` in a file that imports `node:fs`. A regex test in app code uses `.test(`, never `.exec(`. `navigator.clipboard.writeText(` matches none of its names |
| The crawl never follows a `cc-desk://` link | `web/tests/crawl.ts:25-26`: `const url = new URL(href); if (url.origin === new URL(base).origin) …`. A `cc-desk:` address has the origin `null` |
| Tests import app code by relative path | `tests/words.spec.ts:2-4` imports from `../src/lib/time`, `../src/lib/words`, and `../src/lib/snapshot` |
| CI can't run the desktop chain | `.github/workflows/gate.yml` runs on `ubuntu-latest` with no `ptyxis` and no desktop session, and the gate fails on a skipped test. The handler's test runs the script with a stub `ptyxis`; the `xdg-open` chain runs once on the workstation |
| Chromium remembers an "always allow" answer per origin and scheme | Chromium's `Preferences` file holds `protocol_handler.allowed_origin_protocol_pairs`. In `~/.config/chromium/Geoff/Preferences` it reads `{"https://accounts.google.com": {"sms": true}}`. Whether the dialog offers the checkbox to an `http://127.0.0.1` page is unconfirmed (Chromium's `external_protocol_dialog.cc` shows it unless a policy hides it; an Edge thread says the policy limits it to secure pages), so it gets no row of its own: Geoff's click settles it |
| Claude can't click in Chromium | `~/.config/chromium` holds only `Geoff` and `System Profile`, and no `Default` profile, against `~/.claude/docs/bluefin-admin/browsers.md`. `list_connected_browsers` printed `[]` |

## Recorded searches

Run with `git grep -c` over `web/src` and `web/tests` at `31b712f` unless marked.

| Search | Hits | What it settles |
| --- | --- | --- |
| `<Command` | `routes/[project]/[round]/+page.svelte` 1, `routes/[project]/[round]/[pass]/+page.svelte` 1 | One component carries the link and the button to both item pages |
| `waits_for_geoff` | `src/lib/Sidebar.svelte` 1, `src/lib/snapshot.ts` 1 | The one place the waiting lines are drawn |
| `rail-row` | `src/app.css:52`, `:60`, `:63`; `src/lib/Sidebar.svelte:27` (the waiting row's `a`), `:32` (`Nothing`); `src/routes/+layout.svelte:32` (the Rulings link, which needs `a.rail-row:hover`) | The row styles to keep working for the new row shape |
| `main code` | `tests/items.spec.ts` 3, `tests/notfound.spec.ts` 1 | One `code` in `main` stays |
| `Copy command` | `tests/surface.spec.ts` 2 | The walk already allows the button by name |
| `cc-desk://`, `clipboard`, `navigator\.` | 0 each | The link, the copy, and the browser call are new |
| `\.exec(` | `tests/color.spec.ts` 1, `tests/surface.spec.ts` 2, none in `src/` | App code has no `.exec(` today, and the new code adds none |
| `ls web/desktop` | No such folder | The handler and the `.desktop` file are new |
| `git ls-files web/scripts` | `check-playwright-report.sh`, `install-page.sh` | Both new scripts are new |
| `git ls-files docs/follow-ups` | 14 files, counting `.gitkeep`, none on the link | Both follow-ups are new |
| Specs that pin the changed names, from the plan review | `tests/color.spec.ts:228`, `:239` (148 CSS color names as whole words, any case, in every `.svelte` and `.css` file under `web/src`, comments too); `tests/empty.spec.ts:6`; `tests/notfound.spec.ts` | Task 2 runs them too, and new comments avoid words such as "red", "tan", or "white" |
| The 24px target rule | `src/app.css:301-307`, `.rounds a.t, .pass-links a` only | `Command`'s link and button join it, since the scan's `wcag22aa` tag runs `target-size` on every item page |

## File list

| File | Change |
| --- | --- |
| `web/desktop/cc-desk-open` | New. The handler, Bash, executable |
| `web/desktop/cc-desk.desktop` | New. The `.desktop` file, with a placeholder the install script fills with the handler's installed path |
| `web/scripts/install-link.sh` | New. Copies the handler, writes the `.desktop` file, validates, registers, and checks |
| `web/scripts/check-link.sh` | New. The workstation check through `xdg-open` and the real `ptyxis` |
| `web/src/lib/link.ts` | New. Builds the `cc-desk://` address from a project and a round, or nothing when a name fails the handler's rule |
| `web/src/lib/Command.svelte` | The command as text, the terminal link, and the "Copy command" button |
| `web/src/lib/Sidebar.svelte` | Each waiting line gets a row element holding today's row link and the terminal link beside it |
| `web/src/app.css` | Only what the new row shape and the button need, in tokens |
| `web/tests/link.spec.ts` | New. The handler test, the address test, the page test, and the copy test |
| `README.md` | How to register the link, and that it runs nothing until `cc-desk` is on `PATH` |
| `docs/follow-ups/page-link-chromium-click.md` | New. Geoff's click, its check, and which of the link and the button to drop |
| `docs/follow-ups/cc-desk-on-browser-path.md` | New. The Desk and skills round puts `cc-desk` where the handler finds it |

## Acceptance

The two items, word for word from the page design.

9. **The terminal link, if it fits.** Opening `cc-desk://alpha/r1` through `xdg-open` runs the
   stand-in command with `alpha r1`. Each of these runs nothing: `cc-desk://alpha/r1;x`,
   `cc-desk://alpha/r1%3Bx`, `cc-desk://Alpha/r1`, `cc-desk://alpha/r1/x`,
   `cc-desk://alpha/../r1`, and `cc-desk://alpha/r1?x=1`. A click on the link in Chromium is
   tried by hand. The report says what Chromium asked, whether one answer held, and which of the
   link and the button shipped.
   - Met by tasks 1 to 3. The handler's reading of every address is shown in the gate by
     `npx playwright test tests/link.spec.ts` in `web/`. The `xdg-open` chain is shown once on
     the workstation by `bash web/scripts/check-link.sh` after `bash
     web/scripts/install-link.sh`, pasted into the PR. CI has no desktop, so that run isn't in
     the gate.
   - The Chromium click can't be tried by this pass (the last two mechanism rows). The PR says
     so: Chromium wasn't tried, so what it asks and whether one answer holds are open, and both
     the link and the button shipped. The click is Geoff's step, in the follow-up
     `page-link-chromium-click.md`.
   - Before: `xdg-mime query default x-scheme-handler/cc-desk` printed nothing. `xdg-open
     cc-desk://alpha/r1` printed `gio: cc-desk://alpha/r1: Operation not supported` and exited 4.
     `ls web/desktop` printed `No such file or directory`. `npx playwright test
     tests/link.spec.ts` found no test (`Make sure that arguments are regular expressions matching
     test files.`). `git grep -c 'cc-desk://' -- web/src web/tests` printed nothing.
10. **The gate.** `bash scripts/gate.sh` passes, and it now runs the page's checks. A Playwright
    run with a skipped test, with no browser, or with fewer tests than the suite lists fails the
    gate.
    - Met by task 4. Shown by `bash scripts/gate.sh` printing `gate: web tests`, a passed count
      above 91 that holds `link.spec.ts`'s tests, and `gate: ok`; by `git diff 31b712f --
      scripts/gate.sh web/scripts/check-playwright-report.sh .github` printing nothing, so the
      skipped-test and no-browser failures pull request 11 showed still hold; and by the short
      run in task 4 step 2.
    - Before: `bash scripts/gate.sh`, run bare at `31b712f` on 2026-11-18, printed `gate:
      gofmt`, `gate: go vet`, `gate: go test` (seven `ok` packages), `gate: vale`, `gate: web
      install`, `gate: web check`, `gate: web build`, `gate: web tests`, `91 passed (1.5m)`, and
      `gate: ok`, and exited 0.

## The strings the tests write out

Every expected string is written out in the test that reads it, never imported from `web/src/`.
The labels are checked through the `survey` skill's writing standard before the first commit,
and if Survey's standard changes one, the PR names the old and new words.

| Where | Words |
| --- | --- |
| The terminal link on an item page | `Open in a terminal` |
| The terminal link on a waiting line | `Open <project>/<round> in a terminal`, such as `Open north/n1 in a terminal` |
| The button | `Copy command`, the same before and after a click |
| The status after a copy | `Copied` |
| The clipboard after a copy on `/alder/r1` | `cc-desk alder r1` |

## Tasks

The path is the sum, 52 minutes, since the tasks run in order. Before task 1, write the start
time from `date -u` into the draft PR's body, so standing answer 58 can be measured. Every
command names its working folder: `(cd web && …)` for `npx`, and the repo root for `bash
scripts/…` and `bash web/scripts/…`.

Each task runs its named specs, commits by path, and pushes with `git push origin page-link`.
The full gate, `bash scripts/gate.sh` bare, runs before task 2's commit and twice in task 4.
Task 1's first push is `git push -u origin page-link`; then `gh pr create --draft --base main
--head page-link` opens the PR, since CI runs on pull requests only. Each task's report names
the gate run's conclusion on its pushed commit from `gh run list --branch page-link`, or "no
run yet" with the link.

### Task 1: the handler and its test (10 minutes)

Outcome: a script takes one `cc-desk://<project>/<round>` address, runs `ptyxis -x "<absolute
path of cc-desk> <project> <round>"` when both names match `[a-z0-9][a-z0-9-]*`, and runs
nothing otherwise.

- `web/desktop/cc-desk-open`, Bash, `set -euo pipefail`, and `export LC_ALL=C` at the top so
  `[a-z]` can't match an upper-case or accented letter under a UTF-8 locale. A `refuse`
  function prints `cc-desk-open: refused: <reason>` to stderr and exits 2. The guards run in
  this order, so every refusal is exit 2 and never `set -u`'s or `set -e`'s exit 1:
  1. `(( $# == 1 )) || refuse "takes one address"`, before anything reads `$1`.
  2. The pattern sits in a variable, `re='^cc-desk://([a-z0-9][a-z0-9-]*)/([a-z0-9][a-z0-9-]*)$'`,
     and `[[ $1 =~ $re ]] || refuse "…"`. A quoted right-hand side would match as plain text.
     A name can't start with `-`, so it can't reach `cc-desk` as a flag (see `## Decisions`).
  3. `cc=$(command -v cc-desk) || refuse "cc-desk is not on PATH"`, then refuse a path that
     doesn't match `^/[A-Za-z0-9._/-]+$`, since `ptyxis` splits the `-x` string like a shell.
  4. `ptyxis -x "$cc ${BASH_REMATCH[1]} ${BASH_REMATCH[2]}"`, with `ptyxis` found by name on
     `PATH`, never as `/usr/bin/ptyxis`, so the test's stub stands in.

  No `eval`, and no decoding of the address. A comment says why the command is named by
  absolute path: `ptyxis` runs it in its service, whose `PATH` isn't the handler's.
- `web/desktop/cc-desk.desktop`: `[Desktop Entry]`, `Type=Application`, `Name=cc-desk terminal
  link`, `NoDisplay=true`, `Exec=@HANDLER@ %u`, `MimeType=x-scheme-handler/cc-desk;`. The
  install script replaces `@HANDLER@`.
- `web/tests/link.spec.ts`, a `describe` for the handler. Each test makes a `mkdtemp` folder
  with a `bin/` holding an executable stub `ptyxis`, which appends its arguments, one per line,
  to a record file whose path is written in the stub's body, and a stub `cc-desk`. It runs
  `bash <handler> <address>` with `spawnSync`, where `<handler>` is
  `path.resolve(import.meta.dirname, '..', 'desktop', 'cc-desk-open')` (the form `tests/app.ts:9`
  uses; Playwright runs from `web/`). Its environment is only `PATH=<tmp>/bin:/usr/bin:/bin`, and
  the folder is removed in a `finally`.
  1. `cc-desk://alpha/r1`: exit 0, and the record is exactly the two lines `-x` and
     `<tmp>/bin/cc-desk alpha r1`.
  2. Each of the six from done item 9, and `cc-desk://alpha/r1/`, `cc-desk://alpha`,
     `cc-desk:alpha/r1`, `cc-desk:///alpha/r1`, `cc-desk://alpha/r1#x`, `cc-desk://alpha/r1 x`,
     `cc-desk://alph%C3%A4/r1`, `cc-desk://ALPHA/R1`, `cc-desk://-x/r1`, and
     `cc-desk://alpha/--help`: exit 2, no record file, and stderr holding `cc-desk-open:
     refused`. One test per address, named by it.
  3. `cc-desk://alphä/r1`, run with `LANG=en_US.UTF-8` added to the environment, the locale
     under which `[a-z]` matches `ä` unless the handler sets `LC_ALL=C`: exit 2, no record. On a
     machine without that locale bash falls back to C and the test still passes, so it never
     skips.
  4. No argument, and two arguments: exit 2, no record.
  5. No `cc-desk` in `bin/`: exit 2, no record, and stderr names `cc-desk`.

Files: `web/desktop/cc-desk-open`, `web/desktop/cc-desk.desktop`, `web/tests/link.spec.ts`.
Class `ordinary`.

Acceptance check: `(cd web && npx playwright test tests/link.spec.ts)` passes. Then two planted
failures on the workstation, each reverted before the commit, with the failing lines pasted
into the report: drop the `$` from the pattern, and the tests for `cc-desk://alpha/r1;x` and
`cc-desk://alpha/r1/x` fail; remove the `LC_ALL=C` line, and the `alphä` test fails.

### Task 2: the link and the button on the page (17 minutes)

Outcome: a round page, a pass page, and each line under "Waiting for you" show the terminal
link, and both item pages show the "Copy command" button beside the command.

- Before any edit, take the "before" screenshots from the `31b712f` code: build the binary with
  `go build -o "$CLAUDE_JOB_DIR/tmp/cc-desk" ./cmd/cc-desk` at the repo root (a build, not a Go
  edit), run `(cd web && npm run build)`, then `(cd web && CC_DESK_BINARY="$CLAUDE_JOB_DIR/tmp/cc-desk"
  node tests/screenshots.ts "$CLAUDE_JOB_DIR/tmp/shots-before" big)`. The tool's viewport is
  1600 by 1000. After the change, the same command into `shots-after`. Both go in the report.
- `web/src/lib/link.ts`: one function that takes a project and a round and returns
  `cc-desk://<project>/<round>`, or `null` when either name fails `^[a-z0-9][a-z0-9-]*$` (tested
  with `.test(`), with a doc comment saying the rule is the handler's.
- `Command.svelte`: keeps `<h2>Command</h2>` and the one `<code>cc-desk {project} {round}</code>`.
  Beside it, the terminal link from `link.ts` when it isn't `null`, and a `<button
  type="button">Copy command</button>` that writes the command text with
  `navigator.clipboard.writeText`. A visually hidden `role="status"` element reads `Copied`
  after a copy, so the button's name never changes. The `@component` comment says what the
  component now draws. Add `Command`'s link and button to the 24px rule at `app.css:301-307`.
- `Sidebar.svelte`: each waiting line becomes a `div` with class `rail-row` holding today's row
  link first, its text and `.sub` unchanged, and the terminal link after it, named `Open
  <project>/<round> in a terminal`. The empty list still draws `<p class="rail-row
  muted">Nothing</p>`. In `app.css`, keep `a.rail-row:hover` for the Rulings link
  (`+layout.svelte:32`). Add rules for the new row: it's a flex row, its first link grows and
  takes the hover background (for example `.rail-row > a:first-child:hover`), and the terminal
  link is at least 24px by 24px, as pass 3 set for links.
- `web/tests/link.spec.ts`, more `describe`s:
  1. The address: `link.ts` gives `cc-desk://alpha/r1` for `alpha` and `r1`, and `null` for
     `Alpha`, `a.b`, `a_b`, `a/b`, `-x`, an empty name, and `alpha;x`, each pair written out.
  2. The pages, on `twin`: in `main` on `/alder/r1` and on one pass page under it,
     `getByRole('link', { name: 'Open in a terminal', exact: true })` has count 1 and the `href`
     `cc-desk://alder/r1`, and `main code` reads `cc-desk alder r1`. On `big`, inside
     `getByRole('region', { name: 'Waiting for you' })`, the links named `Open north/n1 in a
     terminal` and `Open south/s1 in a terminal` have the `href` values `cc-desk://north/n1` and
     `cc-desk://south/s1`, and the row links `north/n1 2 questions` and `south/s1 1 question`
     keep theirs.
  3. The copy, on `/alder/r1`: grant `clipboard-read` and `clipboard-write` on the context,
     click `Copy command`, and read `cc-desk alder r1` back from `navigator.clipboard.readText()`.
     The status reads `Copied`, and the button is still named `Copy command`.
- Run the Svelte MCP server's `svelte-autofixer` on both components until it reports nothing.

Files: `web/src/lib/link.ts`, `web/src/lib/Command.svelte`, `web/src/lib/Sidebar.svelte`,
`web/src/app.css`, `web/tests/link.spec.ts`. Class `ordinary`.

Acceptance check: `(cd web && npx playwright test tests/link.spec.ts tests/items.spec.ts
tests/sidebar.spec.ts tests/surface.spec.ts tests/scan.spec.ts tests/color.spec.ts
tests/empty.spec.ts tests/notfound.spec.ts)` passes. The walk finds the new button allowed by
its name, and the scan finds no violation on the changed pages. Then the full gate, bare,
before the commit.

### Task 3: the install and the workstation check (9 minutes)

Outcome: the handler is registered for `cc-desk://`, each of the six bad addresses reaches the
handler through `xdg-open` and is refused, the good one runs the stand-in once with `alpha r1`,
and the live page shows the link for Geoff's click.

- `web/scripts/install-link.sh`, Bash with `set -euo pipefail`, safe to run again. It builds
  every path from `share="${HOME:?}/.local/share"`.
  1. `install -D -m 0755 web/desktop/cc-desk-open "$share/cc-desk/link/cc-desk-open"`, run from
     the repo root it finds from its own path.
  2. Write `$share/applications/cc-desk.desktop` from the template with `@HANDLER@` replaced by
     that path, through a `.new` file moved into place.
  3. `desktop-file-validate` on it, then `update-desktop-database "$share/applications"`.
  4. `xdg-mime query default x-scheme-handler/cc-desk` must print `cc-desk.desktop`, or the
     script exits 1. No `xdg-mime default`, and no write to `~/.config/mimeapps.list` (see
     `## Decisions`).
- `web/scripts/check-link.sh`, Bash with `set -euo pipefail`. It makes a folder with `mktemp -d`
  under `$CLAUDE_JOB_DIR/tmp` (or `$TMPDIR`), writes a stand-in `cc-desk` there that appends
  `$*` to a record file named by absolute path in its body, and puts that folder first on
  `PATH`. `xdg-open` exits 0 whatever the handler does, and the handler's stderr reaches
  `xdg-open`'s caller (plan review probe), so the proof is the refusal line. For each of the six
  bad addresses it runs `xdg-open <address> 2> <folder>/<n>.err`, then waits up to 5 seconds in
  a bounded loop until that file holds `cc-desk-open: refused`, and fails if it never does or if
  the record exists. It then runs `xdg-open cc-desk://alpha/r1`, waits up to 10 seconds for the
  record, and requires it to hold exactly one line, `alpha r1`. It prints each address with the
  refusal line it got, then the record, then `check-link: ok`. It removes its folder at the end.
- `README.md`: a short section beside the page's: `bash web/scripts/install-link.sh` registers
  the link; it runs nothing until `cc-desk` is on `PATH`, which the Desk and skills round does;
  Chromium asks before it opens the link.
- Run on the workstation, in this order, and paste each command with what it printed into the
  PR:
  1. `bash web/scripts/install-link.sh`.
  2. `xdg-mime query default x-scheme-handler/cc-desk` prints `cc-desk.desktop`.
  3. `bash web/scripts/check-link.sh` prints `check-link: ok`.
  4. `jq '.protocol_handler.allowed_origin_protocol_pairs' ~/.config/chromium/Geoff/Preferences`,
     a read only, recorded as the value before Geoff's click.
  5. `bash web/scripts/install-page.sh`, which the same auto-mode entry covers and which pass 3
     ran from its branch the same way, so the live page at `http://127.0.0.1:7908` carries the
     link for Geoff's click. Then `curl -s -o /dev/null -w '%{http_code}\n'
     http://127.0.0.1:7908/` prints `200`, and `grep -rl 'Open in a terminal'
     ~/.local/share/cc-desk/page` names a file. The page draws on the client, so `curl` alone
     can't show the link.

Files: `web/scripts/install-link.sh`, `web/scripts/check-link.sh`, `README.md`. Class `runner`.

Acceptance check: the five steps print what they name. The handler stays registered.

### Task 4: the gate's evidence, the follow-ups, the review, and the PR (16 minutes)

Outcome: done items 9 and 10 have their evidence, the follow-ups are filed, the branch is
reviewed, and the PR is ready.

1. `docs/follow-ups/page-link-chromium-click.md`: Geoff opens a round page on
   `http://127.0.0.1:7908` whose command shows `Open in a terminal` (a name with `.` or `_`
   shows none), clicks the link, notes what the dialog asks, ticks "Always allow" if it's
   offered, and clicks once more. The check is the `jq` line from task 3 step 4, read after
   Chromium is closed or a minute later, since Chromium writes `Preferences` late. It then names
   `http://127.0.0.1:7908` with `"cc-desk": true` if the answer held. If it held, a later pass
   drops the button; if Chromium asks every time, a later pass drops the link and the handler,
   as round standing answer 17 says. Until `cc-desk` is on `PATH`, the click opens no terminal,
   so it tests the dialog only.
2. `docs/follow-ups/cc-desk-on-browser-path.md`: the Desk and skills round installs `cc-desk`
   on a `PATH` the handler sees when Chromium starts it (the GNOME session's). It also notes
   that project names with `.` or `_` show no terminal link, because the handler takes only
   `[a-z0-9][a-z0-9-]*`.
3. Merge `origin/page-surface` in if it moved past `31b712f` and `origin/main` if main moved.
   Never rebase a pushed branch.
4. `bash scripts/gate.sh`, bare: paste `gate: web tests`, the passed count, and `gate: ok`.
5. A short run: take the full count from `(cd web && npx playwright test --list
   --reporter=list)`, run `(cd web && npx playwright test tests/link.spec.ts)`, then at the repo
   root `bash web/scripts/check-playwright-report.sh web/test-results/report.json <that count>`
   fails naming the count. Then `git diff 31b712f -- scripts/gate.sh
   web/scripts/check-playwright-report.sh .github` prints nothing.
6. One whole-branch `standards:diff-reviewer` read, folded. If the fold changes any file, run
   the gate once more, bare. Then the PR body and ready.

Files: the two follow-ups, and fixes. Class `ordinary`.

## Standing answers

The twelve in `round-run.md` hold, with the round plan's 13 to 20 and pass 3's plan's 46 to 57
where they apply, and these.

- **58. The pass runs long.** Measure from the start time in the PR body, at two points only.
  Before task 2, if the minutes so far plus 42 pass 75, and before task 3, if the minutes so far
  plus 25 pass 75, take round standing answer 17's road: finish the "Copy command" button and
  its tests, take the terminal link out of the page, and ship no handler. If the handler was
  registered, remove `~/.local/share/applications/cc-desk.desktop` and
  `~/.local/share/cc-desk/link/` by name and run `update-desktop-database` again. Record it under
  `## Decisions`. Once task 3's check passes, finish the pass whatever the clock says. If the
  button alone still can't land by 90 minutes, park under `## Stop`.
- **59. Auto mode refuses a step of the install or the check.** Use the alternatives table once.
  If it still refuses, ship the code, and put the install and check commands with their
  expected output under `## Before merge`, as round standing answer 13 does for pass 3. Never
  edit a settings file or an auto-mode entry.
- **60. `desktop-file-validate` reports an error or a warning.** Fix the template. A hint is
  recorded and left.
- **61. A `ptyxis` window stays open after the check.** Leave it; never close a window or a process
  by force. Say so in the PR.
- **62. The record doesn't appear, or appears twice, or holds other words.** Run
  `bash -x ~/.local/share/cc-desk/link/cc-desk-open cc-desk://alpha/r1` with the stand-in's folder
  first on `PATH` and read where it stops. A fault in the handler or the scripts is fixed and
  the check run again. If the chain through `xdg-open` still fails, the link is dead on this
  desktop: take standing answer 58's road and record what `xdg-open` printed.
- **63. Headless Chromium won't read the clipboard back.** Replace `navigator.clipboard.writeText`
  with a recorder through `page.addInitScript` in that one test, and assert the recorded text.
  The PR says why.
- **64. A sidebar or items test fails on the new markup.** Change the markup, never the test's
  expected name, `href`, or count. If the two can't both hold, go through the advisor.
- **65. The scan or `svelte-autofixer` flags the new link or button.** Fix the markup within
  Survey's tokens. Never disable an axe rule or exclude an element.
- **66. `origin/page-surface` moved and the merge conflicts in a file this pass changes.** Keep
  both sides' intent, run the gate, and say so under `## Decisions`. A conflict in a file this
  pass doesn't own goes through the advisor.

## Decisions

- Call: Claude can't click in Chromium (no `Default` profile exists, against `browsers.md`, and
  no browser is connected), so only Geoff's click can settle the dialog. The pass ships both
  the link and the button, as the conductor's note for this pass allows, and leaves the click to
  Geoff in `page-link-chromium-click.md`. The pass doesn't make a Chromium profile or open
  Geoff's. The click opens no terminal until `cc-desk` is on `PATH`.
- Call: the link shows only when the project and the round both match the handler's rule. The
  state accepts project names with `.` and `_` (`events.go:13`), so such a project shows the
  command and the button with no link. No fixture has such a name. [advisor]
- The handler's rule is `[a-z0-9][a-z0-9-]*`, which narrows the design's `[a-z0-9-]+` so a name
  can't start with `-`. The design's own rule takes `cc-desk://--help/r1` and would hand
  `cc-desk` a flag from any page that links to it. Every name the design's rule refuses is still
  refused, and the first-character rule is the one `events.go:13` sets for project names.
  [review]
- The handler test runs the `alphä` address under `LANG=en_US.UTF-8`, and task 1 plants the
  missing `LC_ALL=C` to show the test catches it. [review]
- `check-link.sh` proves each refusal by the handler's stderr line, since `xdg-open` exits 0
  whatever the handler does. [review]
- Task 3 runs `install-page.sh` so the live page carries the link for Geoff's click, as pass 3
  ran it from its branch. [review]
- The path is 52 minutes, not 38: three full gates at 3 to 4 minutes each, the screenshots,
  and the review. Standing answer 58 measures at task boundaries, so a finished, checked handler
  is never torn out late. [review]
- The `.desktop` file names an installed copy of the handler, `~/.local/share/cc-desk/link/cc-desk-open`,
  whose source is `web/desktop/cc-desk-open` in the repo. This builds on pull request 11's
  `[advisor]` decision to run the page from an installed copy, since a worktree is removed after
  it lands. The copy sits apart from `~/.local/share/cc-desk/page/`, which `install-page.sh`
  replaces whole.
- The install runs `update-desktop-database` and no `xdg-mime default`, against the round plan's
  pre-flight line. The probe showed the cache alone makes the handler the default, and `xdg-mime
  default` would write `~/.config/mimeapps.list`. [advisor]
- The handler finds `cc-desk` on its own `PATH` and hands `ptyxis` its absolute path. That lets
  the test and the check stand in with no environment seam. A missing `cc-desk` runs nothing.
- The pass page shows the link too, through `Command`, as "What each page shows" says. The
  "Waiting for you" lines show the link only; the button sits on the item pages, one click away.
  If the link is dropped later, the follow-up moves a button to each line.
- `~/.claude/docs/bluefin-admin/browsers.md` says Claude drives a `Default` Chromium profile, and
  none exists on pavilion. The PR lists it under `## Planning misses` as `[stale-fact]` for the
  conductor to correct.
