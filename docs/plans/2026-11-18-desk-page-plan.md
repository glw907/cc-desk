# Desk page: pass plan

Initiative: cc-desk
Round: desk
Pass: desk-page

Class `runner`. Planned 90 minutes. The path below is 170 minutes, gate and CI time included,
under twice the plan (180),
so the pass runs and doesn't park. Design: `docs/specs/2026-11-18-desk-skills-design.md`,
sections "The page", "When something goes wrong" (the page's rows), "Backlog", and done items 6
and 11. The design wins over this plan. Round plan: `docs/plans/2026-11-18-desk-skills-round.md`.
Where the round is silent, the main, State-and-rules, Sessions, and Page designs under
`docs/specs/` hold, and the mockups in `docs/specs/2026-11-17-cc-desk-mockups/screens/` set the
look. Anything a person reads (the PR body, STATUS lines, README lines, follow-ups) follows
`/var/home/glw907/.claude/docs/voice/written-for-people.md`, loaded through the `writing-voice`
skill before drafting.

Worktree `/var/home/glw907/Projects/cc-desk/.claude/worktrees/desk-page`, branch `desk-page`,
cut from `origin/main` at `0d992b2`. PR base: `main`. Push: `glw907/cc-desk`. Pass 1
(`desk-records`, PR #17) built every record and verb this pass reads. Its `## Decisions` hand
this pass one item: the pass page reads its own `needs_kind` and `needs_reason`.

This pass writes only under `web/`, `README.md`, `docs/follow-ups/`, and `docs/STATUS.md`.
Everything under `cmd/`, `internal/`, and `skills/` belongs to passes 2, 3, and 5, which may run
beside it. A need for a verb or a snapshot field pass 1 didn't build parks that one task
(round plan standing answer 17).

## Pre-flight

No step needs Geoff. Every write is inside the worktree, a test's temp folder,
`$CLAUDE_JOB_DIR/tmp`, or the page's install: `~/.local/share/cc-desk/page`,
`~/.config/systemd/user/cc-desk-page.service`, and a `mkdir -p` of `~/.local/state/cc-desk`
(which exists today). The round plan's auto mode entry lets `cc-desk-desk-page` reinstall and
restart `cc-desk-page.service` with write access to `~/.local/state/cc-desk` and nothing else.
This pass installs no binary: standing answer 22 has the conductor install the binary and the
page from `main` after the round.

- Permissions checked: `~/.local/state/claude-denials.jsonl`, read on 2026-11-18. It holds 135
  rows, from 2026-11-01 to 2026-11-18T21:28:10Z, all inside 30 days. Since pass 1's read it
  gained two: a classifier outage, and a `claude stop` refused as Destructive. No row refuses
  `bash scripts/gate.sh`, `npm ci` or `npm run build` run bare, `npx playwright test`,
  `systemctl --user`, `bash web/scripts/install-page.sh`, a push of a pass branch,
  `gh pr create`, or `curl` to `127.0.0.1`. One old row refused `npm ci ... | tail` as
  Destructive. The rows that touch this pass's commands are in "Refused commands and their
  alternatives".
- Reviewed: one `standards:plan-reviewer` read of this plan (class `runner`), findings folded
  under `## Decisions` ("From the plan review").
- Evidence: the mechanism table, the recorded searches, and the checks under `## Acceptance`,
  each run on today's code (`0d992b2`) before any code was written. The round plan names the
  unit `systemd/cc-desk-page.service`; it is `web/systemd/cc-desk-page.service`
  (`[stale-fact]`). The installed binary predates pass 1: `~/.local/share/cc-desk/bin/cc-desk
  report read --help` prints `unknown command "report" for "cc-desk"`. So the live check after
  the install can show that the unit reaches the binary and the folder, and only the tests show
  a real mark.

### Refused commands and their alternatives

| Refused in the log | This pass does instead |
| --- | --- |
| `npm ci ... \| tail` (Destructive) | Run `npm ci`, the gate, and Playwright bare |
| `git commit -qam` and `git commit --no-verify` | `git add <path> ...` by name, then `git commit -m`, with hooks |
| `git push origin HEAD:main` | Push only `desk-page` |
| A chained `claude stop` | Never needed. If it is, run `claude stop <id>` alone |
| A write to a file named `settings.json` (Self-Modification) | No file of that name is written |

## Mechanism table

Each row is read from source at `0d992b2` or probed in `$CLAUDE_JOB_DIR/tmp`.

| Mechanism | Where | What it does today | This pass |
| --- | --- | --- | --- |
| Kit's origin check | `@sveltejs/kit` 3.3.1, `src/runtime/server/respond.js` and `csrf.js` | Runs only in a production build. On when `kit.csrf.trustedOrigins` lacks `'*'` (`checkOrigin` is gone in Kit 3). Refuses with 403 a POST, PUT, PATCH, or DELETE whose content type is absent or a form type and whose `Origin` is missing or not the server's own. It runs before `handle`. A form action sent with another content type gets 415 | Turn it on by dropping `trustedOrigins: ['*']` |
| The page's own origin | adapter-node 6.0.4 `get_origin` | Takes the protocol from `PROTOCOL_HEADER`, default `https`. With no `ORIGIN` set, the URL reads `https://127.0.0.1:7908`, so every real click would get 403 | The unit and `tests/app.ts` set `ORIGIN` to the page's `http` address |
| The method guard | `web/src/hooks.server.ts` | Answers every method but GET and HEAD with 405 and `allow: GET, HEAD` | Also lets through one POST: path `^/reports/rep-\d{4,}$`, search exactly `?/read` |
| Origin setting | `web/vite.config.ts:9` | `sveltekit({ adapter: adapter(), csrf: { trustedOrigins: ['*'] } })`, with a comment saying the check is off | Drop the option, rewrite the comment |
| Bundling | adapter-node `index.js` | Leaves out every package under `dependencies`. The install copies only `build/` | `markdown-it` goes under `devDependencies`, so it is bundled |
| Rendering | markdown-it 15.0.3 (probe `t.mjs`) | With `html: false`, `linkify: false`, `disable('image')`, and a `validateLink` that takes only `^https?://`, a `<script>` and an `<img onerror>` come out as text, and a `javascript:`, `data:`, or relative link stays plain text. A heading draws as `h1`, and an `https` link draws | The one renderer, on the server |
| Marking a report | `cmd/cc-desk` `report read <id>` | Refuses an empty id, a leading `-`, and any `/`. Runs `state.Commit` with `state.ReadReport(id)`: no usage reading, no session list, no settle. Takes a flock on `lock` (wait 10 s), writes `reports.json` and `snapshot.json` atomically, removes `snapshot-error.json`. A second read changes nothing. Prints `marked report <id> read` to stderr; a missing one is `report "<id>" not found` | Run from `src/lib/server/command.ts` by its full path |
| Report ids | `internal/state/report.go` | `fmt.Sprintf("rep-%04d", n)` | The route matcher and the hook use `^rep-\d{4,}$` |
| A round's checkout | `release checkout <project> <path> --default-branch <b>` | Sets `checkout` on the release file. The path must be absolute and exist. Runs no git | The fixture sets one; the round page's server load reads it |
| Pass sessions | `internal/state/types.go:258`, a pass's `session` | Written only by the check's launch. `prompt hook` reads hook JSON on stdin and appends a prompt to the live record whose `session_id` matches | The fixture writes a record by hand in pass 1's shape, then runs `prompt hook` |
| Words | `web/src/lib/words.ts` | Tables for pass and round states, five stop reasons, limits, needs kinds, guess outcomes, answer kinds, next-session states. `word()` falls back to the code with spaces | Adds the other stop reasons, session states, check decisions, and check reasons |
| The unit | `web/systemd/cc-desk-page.service` | `Environment=HOST=127.0.0.1 PORT=7908`, `ProtectSystem=strict`, `PrivateTmp`, `NoNewPrivileges`. `systemctl --user show` gives `ReadWritePaths=` empty and `ProtectHome=no`. `/var/home` is its own mount, so `ProtectSystem=strict` leaves it writable: the live unit's mount table shows `/home /var/home rw`, and its comment "the whole file system is read-only" is already false. The user manager's PATH is `/usr/lib64/ccache:/usr/local/sbin:/usr/local/bin:/usr/bin`. Node is named by its full path | Adds `ORIGIN`, `CC_DESK_BINARY`, `ProtectHome=read-only`, and `ReadWritePaths=-%h/.local/state/cc-desk`. Probed by the plan reviewer on a transient user unit: the folder is `rw`, the rest of home refuses a write, and the binary still reads |
| The install | `web/scripts/install-page.sh` | `npm ci` if needed, build, copy to `page.new`, swap, copy the unit, `systemd-analyze --user verify`, `daemon-reload`, `enable --now`, `restart`, then `curl -sf --retry 10` | Adds `mkdir -p ~/.local/state/cc-desk` before the restart |

The unit's PATH needs nothing new. `node` and `cc-desk` are both named by full path, and
`report read` starts no command (no file under `internal/state`, `internal/snapshot`, or
`internal/atomicfile` imports `os/exec`, and `root.go` has no `PersistentPreRun`).

## Recorded searches

Each ran in the worktree at `0d992b2`, over `web/src`, `web/tests`, `web/scripts`,
`web/systemd`, `web/desktop`, `README.md`, and `docs/`. A hit in an old plan or spec under
`docs/plans/` or `docs/specs/` is history and is dismissed.

| Search | Hits | Each hit |
| --- | --- | --- |
| `word(` | 8 files | `session.ts`, the round page, `ScaleBar`, `ReleaseBoard`, the pass page, `Questions`, `ReleaseBlock`, `words.ts`. Each new code reaches the page through these, so none changes for the words alone |
| `stopReason` | 6 | `ReleaseBoard.svelte:62`, `words.ts:20` and `:60`, the pass page `:49`, `words.spec.ts:11` and `:41`. The table and the spec change |
| `terminalLink` | 8 | `Sidebar.svelte:9` and `:28`, `Command.svelte:8` and `:21`, `link.ts:11`, `link.spec.ts:5`, `:120`, `:142`. Only `link.ts` and `link.spec.ts` change |
| `Copy command` | 5 | `Command.svelte:40`, `link.spec.ts:182` and `:185`, `surface.spec.ts:60` and `:68`. The surface lines change |
| `writes nothing`, `never opens a file for writing`, `no form` | 4 | `cc-desk-page.service:12`, `src/lib/server/state.ts:1`, `surface.spec.ts:1`, `README.md:21`. All four change |
| `trustedOrigins` | 1 | `vite.config.ts:9`. Changes |
| `CC_DESK_BINARY` | 3 files | `tests/global-setup.ts:13`, `tests/screenshots.ts:5`, `tests/state.ts:32-33`. None changes; the server reads the same name |
| `Questions` | 4 | The round page `:12` and `:57`, the pass page `:13` and `:68`. Both pages take the upgraded component |
| `'form'` selector | 1 | `surface.spec.ts:62`. Changes |
| `cc-desk-page.service` | 3 live | `README.md`, `install-page.sh`, `web/package.json` (the name only). The first two change; `package.json` doesn't |
| `405` in `web/tests` | 2 files | `server.spec.ts:10-15` and `surface.spec.ts:83-95`. Both send their writes with no `Origin`, which now gets 403. Both change |
| `exec\|execFile\|child_process` in `web/src` | 0 | The source scan's exception for `command.ts` is the only one |
| `connection_dropped` | 2 | `e2e/testdata/desk/check-log.jsonl` (two lines) and `e2e/brief_test.go:41`. The binary never writes it. Both are pass 1's and stay; the words test names it as its one exception |
| `[a-z0-9][a-z0-9-]*` | 4 | `link.ts`, `web/desktop/cc-desk-open`, `check-link.sh`'s cases, `README.md:30-45`. All widen together |

## File list

| File | Task | Change |
| --- | --- | --- |
| `web/tests/state.ts` | 1 | `run` takes stdin; a `page` builder |
| `web/tests/fixtures/page/` | 1 | New: the design file for the fixture's checkout, and the hook input if it differs from pass 1's |
| `web/src/lib/words.ts`, `web/tests/words.spec.ts` | 2 | New tables and their checks |
| `web/src/params/report.ts` | 3 | New matcher |
| `web/src/routes/reports/[id=report]/+page.svelte`, `+page.server.ts` | 3, 4 | New report page and its action |
| `web/src/lib/Sidebar.svelte` | 3 | "Reports to read" |
| `web/src/lib/Report.svelte` or inline in the page | 3 | The report's parts |
| `web/src/routes/[project]/[round]/+page.svelte` | 3, 5 | The round's reports; the design page |
| `web/tests/sidebar.spec.ts`, `web/tests/notfound.spec.ts` | 3 | "Reports to read"; an unknown report |
| `web/src/hooks.server.ts` | 4 | The one POST |
| `web/vite.config.ts` | 4 | Origin check on |
| `web/src/lib/server/command.ts` | 4 | New: runs `report read` |
| `web/src/lib/server/state.ts` | 4, 5 | Header; `stateDir` reused; the release file read |
| `web/tests/app.ts` | 4 | `ORIGIN`, and an environment override for the stub binary |
| `web/tests/server.spec.ts` | 4 | 405 with the page's origin, 403 without, the mark, the refusals |
| `web/tests/surface.spec.ts` | 4, 8 | The source scan's two exceptions; the walk on `page` |
| `web/src/routes/[project]/[round]/+page.server.ts` | 5 | New: reads the design file |
| `web/src/lib/server/design.ts` | 5 | New: find, confine, cap, render |
| `web/src/lib/Design.svelte` | 5 | New: the one `{@html}` |
| `web/src/lib/Questions.svelte` | 5 | Background, option cards, causes, reason |
| `web/package.json`, `web/package-lock.json` | 5 | `markdown-it` under `devDependencies` |
| `web/src/routes/+page.svelte` | 6 | "Since the last check", sessions |
| `web/src/routes/[project]/[round]/[pass]/+page.svelte` | 6 | Prompts, refusals, stop time, the pass's needs |
| `web/src/lib/ReleaseBoard.svelte` | 6 | A stop's time beside its reason |
| `web/tests/items.spec.ts` | 6 | The pass page's "Nothing." line reads the pass's own reason |
| `web/systemd/cc-desk-page.service`, `web/scripts/install-page.sh` | 7 | The unit's grant |
| `web/tests/scan.spec.ts`, `web/tests/color.spec.ts`, `web/tests/crawl.ts` | 8 | The `page` builder; `SHOWN_BY`; `<details>` opened before a scan |
| `README.md` | 7, 9 | Line 21; the link rule |
| `web/src/lib/link.ts`, `web/tests/link.spec.ts`, `web/desktop/cc-desk-open`, `web/scripts/check-link.sh` | 9 | The wider name rule |
| `docs/follow-ups/` | 10 | See task 10 |

`README.md` is shared with pass 3's install lines; standing answer 15 covers it.

## The page

### A report

`/reports/<id>`, matched by `src/params/report.ts` (`^rep-\d{4,}$`). A static `reports` segment
outranks `[project]`, so the matcher only shadows a project named `reports`. The page draws from
the polled snapshot, as every page does (`ssr = false`), and finds the report by id; an unknown
id draws `NotFound`. In order, as the design's sketch:

1. The meta line "A report from <project>." with the round when it has one, and "Sent <when>."
   from `time.ts`.
2. The title as `h1`, then the summary.
3. "Calls made along the way": each call, what was chosen, the other way, and why. "What the plan
   missed": each miss and what it cost. A report with none says so in a sentence.
4. The form: `method="POST"`, `action="?/read"`, `use:enhance`, one button "Mark as read", no
   input, and beside it "The report stays on the round's page afterward." A report already read
   shows "Read <when>." in place of the form.
5. `<details>` with `<summary>`The detail: each pass, its pull request, and its minutes: one row
   per `passes` entry, with the pass's state in words.

The sidebar gains "Reports to read" between "Waiting for you" and "Running", one row per unread
report, newest first: a link to the report reading "<project>: <title>", and under it "<n
judgment calls>, <n planning misses>. <when>" with `count()`, so 0 reads "no planning miss" as the
design's sketch. With none it shows "Nothing", as the other sections. The round page lists its
round's reports, read or not, each a link, so a report marked read stays there.

### Mark as read

`+page.server.ts` holds one action, `read`. It takes the id from `params` (the matcher has
already checked it), checks it again against the same pattern, and calls
`markRead(id)` from `src/lib/server/command.ts`. That calls `execFile` (never `exec`, never a
shell) on `process.env.CC_DESK_BINARY` with the arguments `--state-dir`, `stateDir()`, `report`,
`read`, and the id, each its own array member, with a timeout of 30 s (above the 10 s lock wait).
An unset or relative `CC_DESK_BINARY` is a failure, not a lookup on `PATH`. On a non-zero exit
the action returns `fail(500, { message })`, the message being "Mark as read couldn't be
recorded:" and the first line of stderr. The page shows it in a `role="alert"` line and the
report stays listed. On success the action returns nothing and the next poll draws the report as
read; `use:enhance` with no callback keeps the page in place.

The hook lets through exactly one write: POST, path matching `^/reports/rep-\d{4,}$`, and search
exactly `?/read`. Everything else that isn't GET or HEAD gets 405 and `allow: GET, HEAD`, as
today. Kit's origin check runs before the hook, so the codes a request gets are:

| Request | Answer |
| --- | --- |
| POST `?/read` on a report, from the page's origin, form body | The action runs |
| Any POST, PUT, PATCH, or DELETE with a form body or none, and a missing or foreign `Origin` | 403 from Kit |
| The same, from the page's origin, to any other address | 405 from the hook |
| OPTIONS, or a write with a JSON body, to any other address | 405 from the hook |
| POST `?/read` on a report with a JSON body from any origin, or with no body from the page's origin | 415 from Kit's action call (`page/actions.js:245`); no command runs |
| POST `?/read` on `/reports/-x` or `/reports/a/b`, from the page's origin | 405 from the hook; no command runs |

Item 6's "A write to any other address gets 405" holds for a write from the page's origin. A
write from elsewhere gets 403 first; see `## Decisions`.

### A design that waits

The round page is the design page when `round.design` is set. No new address: the sidebar's
"Waiting for you" rows already link there. A new `+page.server.ts` load reads
`releases/<project>.json` read-only for `checkout`, joins `round.design` to it, takes the real
path of both, and refuses a file outside the checkout, one larger than 256 KiB, or one that
isn't a regular file. It renders the file with markdown-it, set as the mechanism table says, and
returns the rendered HTML, or a sentence saying why there is none. Kit decodes params, so
`/..%2F..%2Fx/r1` gives `project` `../../x`: the load checks `params.project` against
`^[a-z0-9][a-z0-9._-]*$` (`internal/state/events.go:15`) before any read, and so a
`/reports/<id>` the matcher refuses, which falls through with `project` `reports`, reads
nothing. A round with no `design` reads no file. An unknown project or round, a release file
that isn't JSON or has no `checkout`, a missing file, or a refused one gets the sentence "The
design file isn't readable here: <why>." and the rest of the page draws. The read goes real
path, then stat, then read; a file swapped between those steps could pass the check. The source
scan bans `open`, so that small local race is accepted and recorded under `## Decisions`.

The round page then reads, in order: the context paragraphs; "Yours to decide" with each
unanswered question through the upgraded `Questions`; "Already decided" from `round.calls`; the
full design in `<details>` through `Design.svelte`, the one component with `{@html}`; and
`Command` last, which holds the command and the terminal link. A round with no design keeps its
page as today.

`Questions` shows each question's text, its background paragraphs, its options as cards, and its
reason. A card is a plain `<div class="option-card">` holding the option, its cause from `causes`,
and the word "Recommended" on the recommended one, never color alone. A card holds no input,
button, or link, and no `cursor: pointer` or hover style. An answered question keeps today's
answer-kind line.

### The front page and the item pages

"Since the last check" sits on the front page after the next-session section: the newest ten
lines of `check_log`, newest first, each a sentence with its time: the record, what the check
chose in words, the reason in words when there is one, and the line's `outcome`, which says
what happened. The desk fixture's fourth line is a relaunch the check chose with the outcome
"not relaunched: two tries used", so the sentence can't say the session was relaunched. With no lines it says "The check has done nothing yet." It
reads `check_log` and not `last_check`, because `check.json` is written only by `cc-desk check`,
which no test folder runs; see `## Decisions`.

A "Sessions" list follows it: each entry of `snapshot.sessions`, its name linked to its pass or
round when it has one, its state in words from `last_state`, and its relaunches as "restarted
once" or "restarted <n> times" with the last relaunch reason from `check_log` for that record
when one is there. A stopped session shows its reason in words and `stopped_at` through `time.ts`.

The pass page shows the pass's stop reason and `stopped_at`, reads "Nothing." from the pass's own
`needs_reason`, and gains "Permission prompts" and "Refused commands" for the sessions whose
project, round, and pass match: each prompt's time, tool, command, the rule it would add, and its
answer or "Not answered yet"; each refusal's time, tool, target, and reason. A pass with neither
says so in one sentence. `ReleaseBoard` adds a stop's time beside its reason.

## Fixtures

The new builder `page` in `web/tests/state.ts` starts from `copyFixture('desk')`, pass 1's
fixture, and runs pass 1's verbs on it. It edits no old fixture (round plan standing answer 16).

1. Make a checkout folder with `mkdtempSync` in the system temp folder, and copy
   `web/tests/fixtures/page/alpha-r1-design.md` to
   `<checkout>/docs/specs/2026-11-18-alpha-r1-design.md`, the path the desk fixture's round
   `alpha/r1` names. The file holds a heading, a paragraph with an `https://` link, a
   `<script>` element, an `<img src=x onerror=...>`, and a `[here](javascript:...)` link.
2. Run `release checkout alpha <checkout> --default-branch main`.
3. Write a session record by hand on beta's building pass `b-build`, through `writeAtomic` on
   `releases/beta.json`, in pass 1's `Session` shape (`internal/state/types.go:74`): name
   `beta-b-build`, kind the pass kind, a non-empty `short_id`, `session_id` `abc123`,
   `launch_tries` 1, a worktree, a branch, `launched` before the worked clock, `last_state`
   `working`, and one refusal `{time, tool: "Bash", target, reason}`. `validateSession`
   (`internal/state/check.go:313-354`) refuses a record without `short_id` or with
   `launch_tries` outside 1 and 2. This is the hand-written refused command the round plan
   allows.
4. Run `prompt hook` with `e2e/testdata/prompts/hook-bash.json` on stdin, through `spawnSync`
   with `input`, at the worked clock (after `launched`), so the prompt is recorded the way the
   binary records one. `prompt hook` exits 0 even when it fails (`cmd/cc-desk/prompt.go:36-41`),
   so the builder asserts its stderr is empty and no `snapshot-error.json` exists. Step 2 also
   runs at the worked clock.
5. The desk fixture already holds an unread report `rep-0001`, a designed round `alpha/r1` that
   needs Geoff with question `qa` (background and causes), a stopped pass, four check-log
   lines, and two sessions already (`alpha-conductor` with its own refusal, and
   `beta-b1-design`). The report `rep-0001` belongs to round `beta/b2`.
6. Remove the checkout folder when the test's state folder goes: make it under the same
   temp parent as the state folder, or register it for cleanup, so `allStates()` leaves none.

The builder's check: its snapshot has one unread report, `alpha/r1` with `design` set and an
unanswered question, a session on `beta/b-build` with one prompt and one refusal, and a
non-empty `check_log`. The plan reviewer built it by hand with these steps: `beta-b-build` on
`beta/b1/b-build` came out `working` with one prompt and one refusal. If the hook still finds no
record, read `validateSession` before anything else; this is a fixture fix, not standing
answer 17.

## Acceptance

Done items 6 and 11, word for word from the design:

> 6. **The page.** The test's fixture holds an unread report, a design that waits, and a pass
> with a permission prompt and a refused command. A test walks every page and finds no form
> control except "Mark as read". After the walk the state files are byte for byte what they
> were. One click on "Mark as read" changes the report's record and the snapshot, changes no
> other file in the state folder, and takes no usage reading. The report leaves "Reports to
> read" and stays on its round's page. A write to any other address gets 405, and a "Mark as
> read" sent from another origin gets 403. A report name that starts with a dash or holds a
> slash is refused before any command runs. A design page shows each question's background,
> answers, and recommendation, and an option card holds no input, button, or link. A design
> file holding a script tag, an `onerror` attribute, and a `javascript:` link draws none of the
> three. The pass page shows the prompt and the refused command with their causes. The WCAG 2.2
> AA scan passes in both themes, and the words test finds words beside every state shown in
> color.

> 11. **The gate.** `bash scripts/gate.sh` passes with no skipped test. Items 2 to 5 and 7 to 10
> have parts that run live and aren't Go tests. Each live script prints one line for each check
> and ends with its result. A pass's pull request holds that output from a run on the pass's
> last commit, and a pass that leaves it out isn't done.

Each part of item 6, its task, and the command that shows it. Every Playwright command runs from
`web/` after `npm run build`, bare.

| Part | Task | Command | Test |
| --- | --- | --- | --- |
| The fixture holds the three | 1 | `npx playwright test tests/words.spec.ts` | `the page fixture holds a report, a waiting design, a prompt, and a refusal` |
| The walk finds no control but "Mark as read"; files unchanged | 8 | `npx playwright test tests/surface.spec.ts` | The walk on `page`: one form, on report pages only, one button "Mark as read"; every other button "Copy command"; both trees and the checkout hash the same |
| One click changes only the report's record and the snapshot; no usage reading | 4 | `npx playwright test tests/server.spec.ts -g "Mark as read"` | Hash every file before and after: only `reports.json` and `snapshot.json` differ, and `usage.json` is byte for byte the same |
| Leaves "Reports to read", stays on the round page | 4 | the same | After the click the form is gone and the sidebar row is gone. `report read` runs at the wall clock against a fixture dated 2026-11-20, so the test asserts nothing drawn from a time: not the "Read" line's words, not an order by time. The report's round, `/beta/b2`, still links it |
| Another address gets 405 | 4 | `npx playwright test tests/server.spec.ts tests/surface.spec.ts` | Each write method from the page's origin to every visited address, `/api/state`, an unknown address, and an asset gets 405 and `allow: GET, HEAD` |
| From another origin gets 403 | 4 | the same | POST `?/read` with a form content type and `Origin: http://evil.example`, and with no content type and no `Origin`, gets 403, and `reports.json` is unchanged |
| A dash or slash is refused before a command | 4 | the same | With `CC_DESK_BINARY` set to a stub that touches a marker and exits 1: first a good POST `?/read` on `rep-0001` makes the marker and shows the alert line (the control); then, with the marker removed, POST `/reports/-x?/read` and `/reports/a/b?/read` get 405 with `allow: GET, HEAD` (Kit's own 405 would carry `allow: GET`) and no marker exists |
| Background, answers, recommendation; cards hold no control | 5 | `npx playwright test tests/server.spec.ts -g "design"` | `/alpha/r1` shows `qa`'s two background paragraphs, both options with their causes, and "Recommended" on `keep`; `.option-card :is(input,button,a,select,textarea)` counts 0 |
| The design file draws none of the three | 5 | the same | No `script` element in the design's `<details>`, no `[onerror]` anywhere, no `a[href^="javascript:"]`; the file's heading and its `https` link do draw |
| The pass page shows the prompt and the refusal with causes | 6 | `npx playwright test tests/server.spec.ts -g "pass page"` | `/beta/b1/b-build` shows `rm -rf node_modules`, the rule the snapshot gives as its `suggested_rule`, "Not answered yet", and the refusal's target and reason |
| The scan in both themes | 8 | `npx playwright test tests/scan.spec.ts` | `every page on page passes the scan in light` and `in dark` |
| Words beside every colored state | 2, 8 | `npx playwright test tests/words.spec.ts tests/color.spec.ts` | Every session state, stop reason, check decision, and check reason in every built state has a written word; `page` is walked by `color.spec` |

Item 11: `bash scripts/gate.sh`, bare, on the last commit, with no skipped test. Item 6 has no
live part in item 11's list, so this pass adds no live script. The unit's live check in task 7
is recorded in the PR body all the same.

Checks run now, on `0d992b2` before any code (background task, exit 0): `bash scripts/gate.sh`
printed gofmt and `go vet` clean; `ok` for `cmd/cc-desk`, `internal/brief`,
`internal/atomicfile`, `internal/desk`, `internal/snapshot`, `internal/sessions`,
`internal/systemd`, `internal/launch`, `internal/state`, `internal/usage`, and `e2e` (46.5 s);
Vale clean; `found 0 vulnerabilities`; `svelte-check found 0 errors and 0 warnings` on 356
files; the web build; `130 passed (1.5m)`; and `gate: ok`. Every item 6 row above fails today:
no report page, no form, no builder named `page`.

## Tasks

Commit after tasks 2, 4, 6, 8, 9, and 10, each with the gate green; a gate run is about four
minutes at `0d992b2`, counted in the minutes below. Each commit's report names its CI run's
conclusion.

### Task 1: the `page` fixture (15 minutes)

- Outcome: `buildState('page', dir)` builds the state in "Fixtures", and `allStates()` includes it.
- Files: `web/tests/state.ts`, `web/tests/fixtures/page/alpha-r1-design.md`, `web/tests/words.spec.ts`.
  `web/tests/fixtures/` holds only `twin/` today, with no README.
- Class: `runner`.
- Check: `npx playwright test tests/words.spec.ts`.
- Test: a new test in `words.spec.ts` asserts the builder's check from "Fixtures".

### Task 2: words (10 minutes)

- Outcome: every code the binary can write for a session state, a stop reason, a check decision,
  and a check reason has a word.
- Files: `web/src/lib/words.ts`, `web/tests/words.spec.ts`.
- Class: `runner`.
- Check: `npx playwright test tests/words.spec.ts`.
- Test: the spec lists the codes from "Codes" below and checks each table holds each; it walks
  every built state's `sessions[].last_state`, `sessions[].stop_reason`, `check_log[].decision`,
  and `check_log[].reason` and finds a word for each. `connection_dropped` is the one named
  exception: pass 1's desk fixture carries it and the binary never writes it.

Codes, from `internal/state` at `0d992b2`: stop reasons `usage`, `irreversible`, `secret`,
`auth_data`, `scope`, `permission_prompt`, `sign_in_expired`, `unreadable_stop`,
`relaunch_limit`, `launch_failed`; session states `working`, `blocked`, `idle`, `missing`; check
decisions `none`, `end`, `relaunch`, `end_and_relaunch`, `launch_again`, `launch`, `stop`,
`wait_usage`, `wait_link`; check reasons `lost_connection`, `idle`, `missing`, and the stop
reasons. If `internal/state` on `main` gains a code before this pass lands, add its word.

### Task 3: the report page and "Reports to read" (18 minutes)

- Outcome: the report page, the sidebar section, and the round page's reports, as "A report"
  says, with no form yet.
- Files: `web/src/params/report.ts`, `web/src/routes/reports/[id=report]/+page.svelte`,
  `web/src/lib/Sidebar.svelte`, `web/src/routes/[project]/[round]/+page.svelte`,
  `web/tests/sidebar.spec.ts`, `web/tests/notfound.spec.ts`, `web/tests/server.spec.ts`.
- Class: `runner`.
- Check: `npx playwright test tests/sidebar.spec.ts tests/notfound.spec.ts tests/server.spec.ts`.
- Test: on `page`, the sidebar lists `rep-0001` under "Reports to read" with its counts; the
  report page shows its parts in order; `/reports/rep-9999` draws `NotFound`; the round page
  `/beta/b2` links the report.

### Task 4: Mark as read (32 minutes)

- Outcome: the one write, the hook, the origin check, the command, and their tests, as "Mark as
  read" says.
- Files: `web/src/routes/reports/[id=report]/+page.server.ts` and its page, `web/src/hooks.server.ts`,
  `web/vite.config.ts`, `web/src/lib/server/command.ts`, `web/src/lib/server/state.ts` (header),
  `web/tests/app.ts`, `web/tests/server.spec.ts`, `web/tests/surface.spec.ts`.
- Class: `runner`.
- Check: `npx playwright test tests/server.spec.ts tests/surface.spec.ts`.
- Test: the item 6 rows for task 4. `app.ts` passes `ORIGIN` as its own URL and takes an environment
  override per test, used for the stub binary (a shell script in the test's temp folder that
  touches a marker and exits 1). `ORIGIN` is built from `wanted` inside the port-retry loop
  (`tests/app.ts:64-74`), not from `app.url`. The source scan in `surface.spec.ts` allows `execFile` and a
  `node:child_process` import in `src/lib/server/command.ts` only, adds `realpath` to the
  allowed reads, and still bans every other name it bans today; no file in `src` may call
  `.exec(` on a regex, so use `.test` or `.match`. The 405 sweeps send the page's own `Origin`;
  a new loop sends POST, PUT, PATCH, and DELETE with none and expects 403. The click test
  hashes the state tree before and after and waits for the poll to draw the report as read.
  A failure test with the stub binary shows the alert line and the report still listed.

### Task 5: the design page (30 minutes)

- Outcome: the round page as "A design that waits" says, with the renderer set to keep scripts,
  raw HTML, and every link scheme but `http` and `https` off the page.
- Files: `web/src/routes/[project]/[round]/+page.server.ts`, `web/src/lib/server/design.ts`,
  `web/src/lib/server/state.ts`, `web/src/lib/Design.svelte`, `web/src/lib/Questions.svelte`,
  `web/src/routes/[project]/[round]/+page.svelte`, `web/package.json`, `web/package-lock.json`,
  `web/tests/server.spec.ts`.
- Class: `runner`.
- Check: `npm run build`, then `grep -rlE "from ['\"]markdown-it" build --include='*.js'`
  prints nothing (built chunks quote imports with `"`, and the `.map` files carry the source,
  so the grep is limited to `.js`), then `npx playwright test tests/server.spec.ts -g "design"`.
  The stronger proof is task 8's walk, which serves a copy of `build/` with no `node_modules`
  and visits `/alpha/r1`.
- Test: the item 6 rows for task 5, plus: a release with no checkout draws the sentence and the
  rest of the page; a `design` path with `..` that leaves the checkout draws the sentence.
  `npm install --save-dev markdown-it@15.0.3` adds it; name the version as
  `package-lock.json` records it in the PR body, with the reason: no package already in `web/`
  renders Markdown, and a hand-written renderer would be the riskier way to keep HTML out.

### Task 6: the front page, sessions, and the pass page (18 minutes)

- Outcome: "Since the last check", "Sessions", and the pass page's prompts, refusals, stop time,
  and own needs, as "The front page and the item pages" says.
- Files: `web/src/routes/+page.svelte`, `web/src/routes/[project]/[round]/[pass]/+page.svelte`
  (its `@component` block changes: the snapshot now holds a stop time and the pass's own
  needs), `web/src/lib/ReleaseBoard.svelte`, `web/tests/server.spec.ts`, and
  `web/tests/items.spec.ts`, whose lines 28, 70, 75, and 109 pin the pass page's "Nothing."
  line to the round's reason and change wherever the pass's own reason differs.
- Class: `runner`.
- Check: `npx playwright test tests/server.spec.ts -g "pass page|since the last check"`.
- Test: on `page`, the front page lists the four check-log lines newest first with words (the
  fourth line's sentence says the check chose a relaunch and that it wasn't relaunched), and
  the sessions list shows `beta-b-build` as working; the pass page row from item 6; on `stopped`
  the board shows the stop's time.

### Task 7: the unit (12 minutes)

- Outcome: the unit gains `Environment=ORIGIN=http://127.0.0.1:7908`,
  `Environment=CC_DESK_BINARY=%h/.local/share/cc-desk/bin/cc-desk`, `ProtectHome=read-only`,
  and `ReadWritePaths=-%h/.local/state/cc-desk`. Without `ProtectHome=read-only` the home
  mount stays writable (see the mechanism table), so the grant would confine nothing. Its
  comment drops "the whole file system is read-only" and says the page writes one thing, a
  report's read time, through the binary. `install-page.sh` makes the state folder before the
  restart. `README.md:21` says the page holds one form, "Mark as read", and answers every other
  write with 405, or 403 from another origin.
- Files: `web/systemd/cc-desk-page.service`, `web/scripts/install-page.sh`, `README.md`.
- Class: `runner`.
- Check: task 7's edits ride on task 8's commit and its gate run. After task 9's commit (task 9
  changes `link.ts`, which the page bundles), run `bash web/scripts/install-page.sh` once, then
  the live check below, so the live page holds every page change.
- Test: the live check.

The live check, each command run bare and its output pasted under `## Acceptance`:

1. `systemctl --user show cc-desk-page -p Environment -p ReadWritePaths -p ActiveState` names
   `ORIGIN`, `CC_DESK_BINARY`, the state folder, and `active`.
2. With `m=/proc/$(systemctl --user show -p MainPID --value cc-desk-page)/mountinfo`, run
   `grep -E ' / / | /var/home | /var/home/glw907/.local/state/cc-desk ' "$m"`: the root and
   `/var/home` lines show `ro`, and the state folder's line shows `rw`.
3. `curl -s -o /dev/null -w '%{http_code}\n' -X POST http://127.0.0.1:7908/` prints `403`;
   with `-H 'Origin: http://127.0.0.1:7908' -H 'content-type: application/x-www-form-urlencoded'`
   it prints `405`.
4. `curl -s -X POST -H 'Origin: http://127.0.0.1:7908' -H 'content-type:
   application/x-www-form-urlencoded' -H 'x-sveltekit-action: true'
   'http://127.0.0.1:7908/reports/rep-9999?/read'` answers a failure whose message holds the
   installed binary's words: today `unknown command "report"`, which shows the unit ran the
   binary by its full path. After the conductor installs from `main`, the same probe reads
   `report "rep-9999" not found`. Neither changes this pass's files. Once the new binary is
   installed, a clean load can remove a `snapshot-error.json` another session wrote
   (`internal/state/commit.go:77`); that is the binary's normal behavior.

The PR body also says that, until the conductor installs the binary from `main`, the live page
shows a "Mark as read" whose click fails with the old binary's words.
5. `curl -sf http://127.0.0.1:7908/ -o /dev/null && echo ok` prints `ok`.

The unit's PATH (`/usr/lib64/ccache:/usr/local/sbin:/usr/local/bin:/usr/bin`) resolves nothing
this pass needs: `node` and `cc-desk` are full paths, and `report read` runs no command.

### Task 8: the walk, the scan, and the color test on `page` (12 minutes)

- Outcome: the item 6 rows for the walk, the scan, and the words beside color run on `page`.
- Files: `web/tests/surface.spec.ts`, `web/tests/scan.spec.ts`, `web/tests/color.spec.ts`.
- Class: `runner`.
- Check: `npx playwright test tests/surface.spec.ts tests/scan.spec.ts tests/color.spec.ts`.
- Test: the walk on `started` stays as it is, apart from the `Origin` its 405 sweep now sends.
  A second walk runs on `page`: it hashes the state tree, the app copy, and the checkout, and
  asserts the control rule from the item 6 row. Its `named` list adds `/reports/rep-0001`, and
  its floor is its own count, since `page` names 14 addresses where `started` names more than
  20. The walk never clicks. `page` joins the full-crawl list in `scan.spec.ts`; the crawl
  opens every `<details>` before the scan so the design body and the report's detail table are
  scanned. In `color.spec.ts`, `page` gets an entry in `SHOWN_BY` (line 38) and joins the
  crawl condition (line 112).

### Task 9: the name rule (6 minutes)

- Outcome: the terminal link takes names matching `[a-z0-9][a-z0-9._-]*`, the state's project
  rule (`internal/state/events.go:15`), in `link.ts` and `cc-desk-open` together. A name still
  can't start with `-` or `.`, so `ptyxis -x` can't read one as a flag.
- Files: `web/src/lib/link.ts`, `web/tests/link.spec.ts`, `web/desktop/cc-desk-open`,
  `web/scripts/check-link.sh`, `README.md`.
- Class: `runner`.
- Check: `npx playwright test tests/link.spec.ts` and `bash web/scripts/check-link.sh`.
- Test: `a.b` and `a_b` link; `.a`, `-a`, `a/b`, `a;b`, and `A` don't. The installed handler is
  not reinstalled: the auto mode entry doesn't name `install-link.sh`, and no project with `.` or
  `_` exists today. Say so in the PR body.

### Task 10: close (17 minutes)

- Outcome: the PR is ready, with the whole-branch review folded and CI green on the last commit.
- Files: `docs/follow-ups/cc-desk-on-browser-path.md` (drop its last paragraph, the name rules,
  which task 9 does; if pass 3 already removed the file on `main`, nothing to do),
  `docs/follow-ups/blocked-session-reads-as-unreadable-stop.md` (reword: the stop stays and the
  page now says it in words; what stays open is showing the transcript's last message as the
  cause), a new `docs/follow-ups/desk-fixture-connection-dropped.md` (the desk fixture and
  `e2e/brief_test.go:41` use a reason the binary never writes), `docs/STATUS.md` if the executor
  reference asks.
- Class: `runner`.
- Check: the PR body's `## Acceptance`; the `gate` check on the last commit.
- Test: none of its own.

Path: tasks 1 to 10 in order, 170 minutes. Tasks 3 and 4 share the report page, and 5 and 6
share the round and pass pages, so the pass runs no pair.

## The gate and CI

The gate is `bash scripts/gate.sh`, run bare before every commit, never piped. The CI watch,
after each push:

1. `gh run list --branch desk-page --limit 3`
2. `gh run watch <id> --exit-status` on the newest run for the pushed commit, run bare with
   `run_in_background`.

Before the next push, read that watch's result. A red run is fixed before any other task.

## Standing answers

The twelve in `round-run.md` hold, with the round plan's changes and its answers 13 to 25. This
pass adds:

- **26. Kit answers a code the table doesn't predict.** Read `respond.js` and `csrf.js` in
  `node_modules/@sveltejs/kit` before changing the hook. Keep the rule: the one form works from
  the page's origin, nothing else writes. Record the real code in `## Decisions`.
- **27. A click gets 403 in the test or live.** The server's origin and the browser's differ:
  check `ORIGIN` in `app.ts` or the unit before anything else. Never turn the check off again.
- **28. markdown-it lets something through.** Tighten its settings (a rule disabled, a stricter
  `validateLink`). Never add an HTML sanitizer package, and never let `{@html}` take anything
  but its output.
- **29. The source scan flags a call this plan needs.** Allow it in the one named file only and
  list it. A write call (`writeFile`, `rename`, and the rest) is never allowed: the binary
  writes.
- **30. The live check's `curl` gets an answer the plan doesn't name.** Paste it as it is and
  say what it shows. Don't install a binary to change it.
- **31. A test needs a session state or a code no verb produces.** Write it by hand in pass 1's
  shape inside the `page` builder, as step 3 of "Fixtures" does, and list it. Never edit `desk`.
- **32. The path runs past 180 minutes.** Finish the task in flight, push, and write the stop,
  naming the tasks left.
- **33. A review after the whole-branch read finds only low findings.** Fold them once or file
  them as follow-ups. They never start another round of review.
- **34. The unit fails to start, or the action fails, under `ProtectHome=read-only`.** That
  setting was probed on a transient unit, not with node serving the page. Paste the journal's
  line (`journalctl --user -u cc-desk-page -n 20`) and write the stop. Never drop the setting
  or widen `ReadWritePaths` to get past it.

## Decisions

- [advisor] Item 6's "A write to any other address gets 405" holds for a write from the page's
  origin. With the origin check on, Kit refuses a form or bodiless write from elsewhere with 403
  before the hook runs, so a bare `curl -X POST` now gets 403. The tests send their 405 sweeps
  with the page's `Origin` and add the 403 cases.
- [advisor] The unit and `tests/app.ts` set `ORIGIN`. Without it adapter-node reads the origin as
  `https://`, and every real click would get 403. A fixed origin also refuses a page reached by
  another host name; the failure line covers that 403.
- [advisor] The round page's server load reads `releases/<project>.json` for `checkout`. The
  field exists and item 6 can't pass without the file, so the pass doesn't park on standing
  answer 17.
- [advisor] `markdown-it` goes under `devDependencies`, so adapter-node bundles it into
  `build/`, the only folder the install copies.
- [advisor] The source scan keeps its bans and allows `execFile` in one named file.
- The design page is the round page when the round has a design; no new address.
- "Since the last check" reads `check_log`, not `last_check`: no test folder has a `check.json`,
  and the log names what the check did. The advisor preferred `last_check`; the log is what the
  fixture can hold, and it shows the same runs.
- "No usage reading" is shown by the byte comparison: only `reports.json` and `snapshot.json`
  change, and `usage.json` is byte for byte the same. The fixture's `release checkout` takes a
  reading, so `usage.json` already exists before the click.
- The terminal link's name rule widens to the state's project rule in this pass, from the
  Backlog table. The handler on disk keeps the old rule until the next `install-link.sh`.
- From the plan review (one `standards:plan-reviewer` read, Opus at medium), each folded:
  - The hand-written session needs `short_id` and `launch_tries` 1, and `prompt hook` exits 0
    on failure, so the builder checks stderr. The reviewer built the fixture to confirm it.
  - `ProtectSystem=strict` leaves `/var/home` writable on this host; the unit adds
    `ProtectHome=read-only`, probed on a transient unit. The live check reads the root and
    home lines.
  - The `started` walk stays, and `page` gets a second walk with its own floor.
  - The bundle grep reads `.js` files only, with either quote.
  - `rep-0001` is on `beta/b2`, not `alpha/r1`.
  - The request table adds the 415 rows; the 403 rows name their content type.
  - The round load checks `params.project` before it builds a path.
  - "Since the last check" draws each line's `outcome`.
  - `items.spec.ts` and `color.spec.ts`'s `SHOWN_BY` join the file lists.
  - The stub rows gain a control that shows the stub runs, and the `allow` header shows the
    hook ran.
  - The minutes count gate and CI time: 170, from 148.
  - Notes folded: `ORIGIN` from `wanted` in the retry loop, the click test checks a "Read" line
    rather than its words, the live page's failing click until the install, the scan opens
    `<details>`, the checkout is cleaned up, and the design read's small race is accepted.
- The design read's real-path-then-read race is accepted: a swap needs a local writer inside
  the checkout, and closing it would need `open`, which the source scan bans.
