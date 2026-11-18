# Page: the round plan

This file is what `pass-core` asks a round of several passes to carry on top of its design. The
design is [2026-11-17-page-design.md](../specs/2026-11-17-page-design.md). Where the two differ,
the design wins. A planner writes each pass's plan just before the pass runs.

- Initiative: `cc-desk`
- Round id: `page`
- Repo: `/var/home/glw907/Projects/cc-desk`, pushing to `glw907/cc-desk`, default branch `main`
- Conductor: `conductor-cc-desk`, a session Geoff starts in his terminal on Opus. It conducts
  this round and the Sessions round together
- Conductor prompt: `docs/plans/2026-11-17-sessions-page-conductor-prompt.txt`
- Chain: 345 planned minutes, or 315 without the terminal link. The round starts nothing new
  past 690.

## Pre-flight

**Permissions checked:** I read `~/.local/state/claude-denials.jsonl` on 2026-11-17 for
2026-10-18 to 2026-11-17. The log holds 130 rows, from 2026-11-01T00:55Z to 2026-11-18T06:27Z.
These are the round's commands that the log shows refused, or has never seen.

| Command | Refusals in the log | Alternative |
| --- | --- | --- |
| `npm ci` chained with other commands and piped to `tail` | 1 row | Run `npm ci` alone, bare, from `web/` |
| A background browser started with `nohup` | 1 row | Playwright starts and ends its own browser through `npx playwright test` |
| An edit that adds Playwright's download hosts to a sandbox setting | 1 row | No pass edits a setting. Chromium is already at `~/.cache/ms-playwright/`. A missing browser parks the pass |
| `node` one-liners | 5 rows | A script file in `web/scripts/`, run by path |
| `systemctl --user enable --now cc-desk-page.service` | None in the log | Covered by the auto-mode entry in the Sessions round plan. A refusal parks pass 3 with the commands in the pull request, as standing answer 3 says |
| `xdg-mime default` and `update-desktop-database` | None in the log | The same, for pass 4 |
| Any command, in a classifier outage | 47 rows | Wait 60 seconds and run it once more |

No row refuses `npm run build`, `npm run check`, `npx playwright test`, `bash scripts/gate.sh`,
`git push` of a pass branch, or `gh pr create`.

**Reviewed:** the design had its builder check on 2026-11-17, and every finding is folded.
Geoff's start of the conductor is his approval of the design, as `docs/STATUS.md` says. Pass 1
edits the gate and the CI workflow, so its plan gets one `standards:plan-reviewer` read. So do
passes 3 and 4.

**Evidence:** the shared assumptions and the launch checks below.

**Steps for Geoff:** none of his own. The Sessions round plan holds the one auto-mode entry
for both rounds. Passes 1, 2, and 5 run without it.

## The passes

Each pass's acceptance list is the named items of the design's "What done looks like", which
its planner copies word for word. Every pass also carries item 10, the gate, and writes its
pull request body by `~/.claude/docs/voice/written-for-people.md`.

| Pass | Slug | Outcome | Acceptance | Files | Class | Minutes | Depends on | Shares a file with |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | `page-front` | The app exists, reads the snapshot, draws the sidebar and the front page, and redraws on a change | Items 1, 3, 5 | `web/` (new: the app, its config, the snapshot type, the words table, the tests), `scripts/gate.sh`, `.github/workflows/gate.yml`, `.gitignore`, `.vale.ini` if `web/` needs leaving out | `runner` | 90 | Nothing | 2, 3, 4, 5 |
| 2 | `page-items` | Every round and pass has a page, and the rulings have one | Item 2 | `web/src/routes/`, `web/src/lib/`, `web/tests/` | `ordinary` | 75 | 1 | 1, 3, 4, 5 |
| 3 | `page-surface` | The walk, the scan, and the words tests pass, and the service runs | Items 4, 6, 8 | `web/tests/`, `web/systemd/`, `web/scripts/`, the README | `runner` | 60 | 2 | 1, 2, 4, 5 |
| 4 | `page-link` | The link opens a terminal, or the copy button stands in | Item 9 | `web/src/`, `web/desktop/`, `web/scripts/`, `web/tests/` | `runner` | 30 | 3 | 1, 2, 3, 5 |
| 5 | `page-polish` | One `standards:design-refinement` round | Item 7 | `web/src/` | `ordinary` | 90 | 4, or 3 if 4 is dropped | 1 to 4 |

Pass 1 is `runner` because it changes the gate and the CI workflow. Passes 3 and 4 are
`runner` because each installs something on the workstation.

No pass in this round edits a Go file or anything under `e2e/`, `internal/`, or `cmd/`. Those
are the Sessions round's. A pass that finds it must edit one stops. After pass 1, no pass in
either round edits `scripts/gate.sh` or the workflow.

## Shared assumptions

| Assumption | Passes | Probe or source |
| --- | --- | --- |
| Survey installs by package at a tag, and has a recipe for a plain app | 1 | `git -C ~/Projects/survey tag` shows `v1.7.0` as the latest on 2026-11-17. `~/Projects/survey/DESIGN.md` holds the plain-app recipe and the install line `npm install github:glw907/survey#v<version>`. Not run here for this app: pass 1's first task installs it and builds |
| The mockup is SvelteKit with Tailwind 4 and daisyUI 5, and carries a copy of Survey's CSS | 1 | `docs/specs/2026-11-17-cc-desk-mockups/` source, read 2026-11-17 |
| The snapshot's shape is what `internal/snapshot/snapshot.go` writes, with times in UTC | all | Read 2026-11-17: `generated_at`, `releases`, `rulings`, `next_session`, `waits_for_geoff`, `waiting`, `open_guesses`, and the two guess totals |
| The built binary prints a snapshot from a fixture state folder at a fixed clock | all | The last round's tests do it with `--state-dir` and the hidden `--now`, in `e2e/` |
| Port 7908 is free, and 7907 is held by the old queue | 3 | `ss -ltn` on 2026-11-17 showed 7907 and no 7908. The conductor reads it again before pass 3 |
| A user unit's `PATH` lacks the mise Node | 3 | A user unit's `PATH` read `/usr/lib64/ccache:/usr/local/sbin:/usr/local/bin:/usr/bin` on 2026-11-17. The unit names Node by its full path |
| Node is 26 locally | all | `node --version` printed `v26.13.1`. CI needs a `setup-node` step pinned by commit, which pass 1 adds |
| `ptyxis -x "<command>"` opens a terminal that runs it, and no handler for `cc-desk://` exists | 4 | `ptyxis --help` lists `-x`; `xdg-mime query default x-scheme-handler/cc-desk` printed nothing, 2026-11-17 |
| Chromium is the browser Playwright drives, and it is installed for Playwright | 1, 3 | `ls ~/.cache/ms-playwright/` showed `chromium-1248` and `chromium-1255` on 2026-11-17. Pass 1's first task runs one empty test to prove the version its Playwright wants is there |

## Order table

Passes run one at a time within this round, in this order, and merge in this order. One pass
of the Sessions round runs beside each.

| Pass | Planner prompt | Depends on | Planned minutes | Waits for Geoff |
| --- | --- | --- | --- | --- |
| 1 `page-front` | `2026-11-17-page-1-planner-prompt.txt` | Nothing | 90 | no |
| 2 `page-items` | `2026-11-17-page-2-planner-prompt.txt` | 1 | 75 | no |
| 3 `page-surface` | `2026-11-17-page-3-planner-prompt.txt` | 2 | 60 | no, once the auto-mode entry reads done |
| 4 `page-link` | `2026-11-17-page-4-planner-prompt.txt` | 3 | 30 | no |
| 5 `page-polish` | `2026-11-17-page-5-planner-prompt.txt` | 4 `(order)` | 90 | no |

## Usage ceiling

100 percent of the weekly allowance (Geoff, 2026-11-17). The figure read 91 percent on
2026-11-17, with the reset at Nov 20, 10:59pm.

## Standing answers

The twelve in `round-run.md` hold, with the Sessions round plan's changes to 1, 4, 6, and 9,
and these.

- **13. The auto-mode entry isn't in place when pass 2 lands.** Pass 3 builds everything but
  the install. Its pull request carries the install commands and their check under
  `## Before merge`, and waits for Geoff's approve.
- **14. A pass's plan comes out over twice its minutes.** It parks.
- **15. The page needs a field the snapshot lacks.** The pass doesn't change Go. It shows what
  the snapshot has, and files the field as a follow-up for the Desk and skills round.
- **16. A Sessions pass lands a snapshot change while this round runs.** Expected. The type
  test reads the snapshot the binary prints from the last round's fixtures, and an unknown
  field is ignored at run time. A type test that fails after a rebase gets the new field
  added to the type, and nothing drawn for it.
- **17. Pass 4 doesn't fit.** If its plan comes out over 60 minutes, or Chromium asks at every
  click, it ships the "Copy command" button and no handler. That is the design's fallback and
  no stop.
- **18. The polish.** It runs without Geoff. It changes spacing, type, and hierarchy within
  Survey, and changes no word the tests read and no layout the mockup settles.
- **19. Wording.** Every label follows Survey's writing standard, loaded through the `survey`
  skill. Any wording Geoff corrects is logged as that skill says.
- **20. The report.** The conductor writes it at
  `~/Projects/initiatives/cc-desk/reports/page-report.md`, with the screenshots' paths, and
  gives Geoff its first paragraph in the terminal.

## Start line

The same five states as the Sessions round plan's "Start line". The two rounds start
together.

### Launch checks, run on 2026-11-17

The Sessions round plan holds the shared checks. These are this round's.

| Check | Output |
| --- | --- |
| `node --version` | `v26.13.1` |
| `git -C ~/Projects/survey describe --tags --abbrev=0` | `v1.7.0` |
| `ls web` in the checkout | No such folder. Pass 1 makes it |
| `xdg-mime query default x-scheme-handler/cc-desk` | Empty |

## Decisions

- Five passes, with the terminal link fourth so that it can be dropped without moving another.
- Pass 1 owns the one change to the gate and the workflow for both rounds.
- The auto-mode entry is shared with the Sessions round, so Geoff approves one.
