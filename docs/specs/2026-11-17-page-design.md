# cc-desk Page: the round design

This is the fourth round of the cc-desk route. After it, a page on the workstation shows the
next decision session, what waits for Geoff, and every release with its rounds and passes. The
page reads the snapshot the command writes. It writes nothing.

The main design is [2026-11-17-cc-desk-design.md](2026-11-17-cc-desk-design.md), and this round
builds its section "The page". The mockup beside that file,
`2026-11-17-cc-desk-mockups/`, settles the layout and the look. The snapshot is described in
[2026-11-17-state-and-rules-design.md](2026-11-17-state-and-rules-design.md). Where this file is
silent, those hold.

## Why this round

Geoff stopped reading the Approval Queue on 2026-11-17. Until this page exists, the only way
to see where the work stands is to ask a session. The command already computes everything the
page shows, so this round draws it and changes no rule.

The round runs beside the Sessions round. It touches no Go file, so the two share nothing but
one line of the gate script.

## Goals

- Geoff opens one address and sees when he is next needed, with a range, and what for.
- Every round and pass on the page links to a page about it.
- The page shows the state within 5 seconds of a change, with no reload.
- Every state shown in color is also shown in words, in both themes.
- The page can't change anything.

## What it leaves out

- Design pages, report pages, "Reports to read", and "Mark as read". The reports and the
  design-conversation line arrive as files in the Desk and skills round, and that round draws
  them. The sidebar has no reports section until then.
- What the check did, and each session's state. The Sessions round adds those to the snapshot,
  and the Desk and skills round draws them (the route gives it "the page half of 6").
- Any change to the snapshot. The page draws what the command writes today. A field the page
  doesn't know is ignored, so the Sessions round can add fields while this round runs.
- Sign-in. The page listens on the loopback address only.

## What would make it a failure

- The page works out a time, a count, or an order itself, and it differs from the command's.
- A state is told apart by color alone.
- A link, a button, or a page load changes a state file.
- The page shows an old snapshot as if it were current.

## Approaches

**How the page learns of a change.**

| Approach | What it trades |
| --- | --- |
| **The browser asks the server every 2 seconds whether the snapshot changed, and reloads the data when it has.** Picked. | One tiny request every 2 seconds per open tab. It has no long-lived connection to drop, and the test is one file write and one wait. |
| The server watches the state folder and pushes an event to the browser. | Nothing is sent while nothing changes. The command replaces `snapshot.json` by renaming a new file over it, which a file watch can miss, and a dropped connection needs its own retry. |
| The command writes a finished HTML page after every change. | No server. It puts the drawing in Go, and the main design fixes a SvelteKit app. |

**Where the page gets its types.** The page has a hand-written TypeScript type for the
snapshot, and a test decodes a snapshot printed by the real `cc-desk` binary and fails on a
field the type doesn't name or a named field that is missing. Generating the type from Go was
the other choice. It needs a Go change, and this round touches no Go file.

## The app

The app lives in `web/` in the cc-desk repo. It is SvelteKit 3 on Svelte 5 with the Node
adapter, Tailwind 4, and daisyUI 5, as the mockup is. It takes Survey from `glw907/survey` at
tag `v1.7.0` and follows Survey's plain-app recipe. The mockup carries a copy of Survey's CSS;
the app imports the package and carries none.

The server reads two files from the state folder on each request: `snapshot.json` and
`snapshot-error.json`. The folder is `$XDG_STATE_HOME/cc-desk`, else `~/.local/state/cc-desk`,
and an environment variable overrides it for tests. The server never opens a file for
writing, anywhere, and it never runs the command. It answers `GET` and `HEAD` and refuses
every other method with 405.

An item's address is `/<project>/<round>` or `/<project>/<round>/<pass>`, so two projects can
each have a round of one name.

A user unit, `cc-desk-page.service`, runs the built app on `127.0.0.1:7908`. The old queue
holds 7907 until the Retire round. The unit file lives in `web/systemd/`, and a script in
`web/scripts/` installs and enables it. The Move round records the install in the workstation
repo.

The page's tests run in the gate. The gate fails when the Playwright run reports a skipped
test, finds no browser, or runs fewer tests than the count the suite lists.

## What each page shows

Everything is drawn from the snapshot. The page formats a time ("Tomorrow, about 9:30 am",
"20 minutes ago") and does no other arithmetic.

**Times.** Every time is shown in `America/Anchorage`, whatever zone the browser is in, and to
the minute with no rounding. "About" is a word, never a rounding. "Today", "Tomorrow", and
"20 minutes ago" are measured from the browser's clock. The tests pin that clock to the
fixture's time and compare the page with strings written out in the test, never with the
app's formatter.

**Empty fields.** A field the snapshot holds as null shows an en dash, with "None" as its
text for a screen reader.

**Themes.** The page follows the browser's light or dark setting and has no theme switch.

**The sidebar** is on every page, about 23rem wide, in this order.

1. *The next decision session.* The only box in the sidebar. It shows the state in words,
   then the time and its range, then the count of questions, rounds, and projects, and any
   round named as not designed yet. The four states read "Needs you now", "Ready now", a time
   for a scheduled session, and "Nothing waiting". "Needs you now" names each reason.
   "Rough" shows beside the time when `next_session.rough` is set. The browser tab's title is
   the state and then "cc-desk". For a scheduled session the state there is the time as the
   widget words it, as in "Tomorrow, about 9:30 am · cc-desk".
2. *Waiting for you.* One line per round with questions needed now, linking to the round's
   page. Empty, it reads "Nothing".
3. *Running.* One block per release: a scale bar with a segment per round as wide as its share
   of planned minutes, "2 of 17 rounds done", and the percent of planned minutes done. "Rough"
   shows beside the release's name here, and on its board, when the release's `rough` is set. A small
   triangle under a segment marks a round that needs Geoff. Each round with a pass building
   shows under the bar with a bar of its passes and its state in words.

**The front page** repeats nothing from the sidebar. It shows what the next session will
cover: each round, what its questions ask, and when each is needed. Then it shows a board per
release that lists every round, with the running rounds opened to their passes. Passes waiting
on the two-pass limit or on usage are listed with the reset text. Open guesses are listed with
their minutes and the total against 300.

**A round page and a pass page** show the item's name and summary, its state, its planned and
actual minutes, the planned minutes of unfinished work before it, its needed time, and what it
needs from Geoff. That last part is the written questions, or "Nothing" with the reason, or
"Not designed yet" with the forecast time. A round page lists its passes, each a link. Both
pages show the command `cc-desk <project> <round>`, with the terminal link beside it.

**A rulings page** lists every ruling, newest first: the time, the question, and Geoff's
words. The main design says every ruling shows on the page, so he can read what was recorded
in his name.

**State words.** A pass reads "Not started", "Waiting", "Building", "Stopped", or "Done", as
the snapshot has it. A stopped pass names its reason in words. The mockup's "In review" and
"Stalled" don't exist in the state and aren't used.

Every code the snapshot can carry has its words in one table in the app: pass and round
states, stop reasons, the limits a pass waits on, the kinds of need, and guess outcomes. The
five pass words above are fixed. The builder writes the rest to Survey's writing standard. A
test fails when any fixture snapshot holds a code with no entry. A code the table doesn't
know, as the Sessions round's new stop reasons will be, shows as the code with its
underscores turned to spaces. `next_session.reasons` is already prose and shows as written.

The words for a state sit in the same element as its color, so a test can look for them
there.

## The terminal link

The route calls this not critical, and the round can close without it.

A round page, and each line under "Waiting for you", shows a link of the form
`cc-desk://<project>/<round>`. A `.desktop` file registers a handler for that scheme, and the
handler runs `ptyxis -x "cc-desk <project> <round>"`. The handler takes the project and the
round only when each matches `[a-z0-9-]+`, and it runs nothing otherwise.

Chromium asks before it opens an outside program. If it asks every time, with no "always
allow" that holds, the link is dropped and a "Copy command" button takes its place. The pass
tries the link first in Chromium and records what it saw.

The `cc-desk <project> <round>` form itself is built in the Desk and skills round. Until then
the handler is tested with a stand-in command, and the page shows the command as text.

## When something goes wrong

| What happens | What the page shows |
| --- | --- |
| `snapshot-error.json` exists | The last good snapshot, with a banner: "Stale", the file's message, and the time of the last good snapshot |
| `snapshot.json` is missing | "No state yet", and nothing else |
| `snapshot.json` isn't valid JSON | The same banner as stale, with the parse error, over the last snapshot the server read. With none, "No state yet" and the error |
| The server stops answering | A banner in the open tab within 10 seconds: "Not connected", with the time of the last answer |
| An address names a project, round, or pass that doesn't exist | A not-found page with a link home |

| Limit | Value |
| --- | --- |
| Time from a state change to the redraw | 5 seconds |
| Poll interval | 2 seconds |
| Time to "Not connected" | 10 seconds |
| Sidebar width | About 23rem |

## The polish

The last pass is one round of `standards:design-refinement` on the working page. The main
design says Geoff doesn't review it, and the route says "the polish works with him". This
design follows the main design: the pass runs without him, lists its changes in the report,
and puts before and after screenshots of each page in the pull request. He can ask for more
after he has used the page.

## The passes

Five passes in one chain, since each builds on the app the first one makes.

| Pass | Outcome | Done items | Minutes |
| --- | --- | --- | --- |
| 1. The front page | The app exists, reads the snapshot, draws the sidebar and the front page, and redraws on a change | 1, 3, 5 | 90 |
| 2. Item pages | Every round and pass has a page, and the rulings have one | 2 | 75 |
| 3. The reading surface | The walk, the scan, and the words tests pass, and the service runs | 4, 6, 8 | 60 |
| 4. The terminal link | The link opens a terminal, or the copy button stands in | 9 | 30 |
| 5. The polish | One `standards:design-refinement` round | 7 | 90 |

Every pass carries item 10. Pass 4 can be dropped, and pass 5 then follows pass 3.

## Left to the builder

The route layout under `web/src/routes/`, the component split, the test tools beyond
Playwright and axe-core, the wording of the poll endpoint, and the words for every code but
the five pass states and the four widget states.

## Steps only Geoff can do

None before the round. After the terminal-link pass, Chromium may ask him once to allow the
`cc-desk` scheme. The pass records whether that one answer holds.

## What done looks like

Every check runs the built app against a state folder the test makes. The snapshots in it are
printed by the real `cc-desk` binary from the fixtures in `e2e/testdata/`, never written by
hand. Every expected string is written out in the test.

1. **Estimate, on the page.** The browser's clock is pinned to the fixture's time and its
   zone set to one that isn't Alaska's. With the fixture of two releases from the State and
   rules round, the widget shows the next session's time and range as the Alaska times that
   round's test states (8:00 pm, 6:00 pm to 9:30 pm), with its question count and projects.
   The front page shows the same figures. With the fixture where a pass finished in half its
   planned minutes, the widget reads 6:30 pm with 4:30 pm to 7:00 pm, and the slow one reads
   9:30 pm with 8:00 pm to midnight. "Rough" shows beside the release whose `rough` is
   set and beside no other, and the widget shows it only in the variant where
   `next_session.rough` is set.
2. **Click-ahead.** A test collects every round and pass in the snapshot and fails when one
   has no link on the front page, or when its link's address isn't that item's. Each linked
   page shows the item's name in its heading, and its planned minutes and minutes of work
   before it in elements labeled for those fields, equal to the snapshot's. A fixture where
   two projects each have a round named `r1` gives two pages with different figures. A round
   with written questions shows them. A round with none shows "Nothing" and its reason. An
   undesigned round shows "Not designed yet" with its forecast time. A null field shows the
   en dash.
3. **Waiting order, on the page.** Against the State and rules round's order fixtures: with
   only the 400-minute dependent pass, the widget reads "Ready now". After the 200-minute
   guess start, the front page lists the guess with "200 of 300 minutes", and a snapshot
   edited by the test to a total of 250 reads "250 of 300", so the page isn't adding the
   figures itself. A pass held by the two-pass limit reads "Waiting" with that reason, in
   that pass's row.
4. **Reading surface.** A test walks every page, following every link on this site, and finds
   no `form`, `input`, `select`, `textarea`, or `contenteditable`, and nothing with the role
   of a button but "Copy command". A `POST` to any address returns 405. The walk and the poll
   run with the state folder and the app's folder both read-only, the server's log holds no
   error, and afterward every file in both is byte for byte what it was. A test searches the
   server's source and finds no call that writes a file.
5. **Redraw.** A test replaces `snapshot.json` while a page is open. The page shows the new
   state within 5 seconds with no reload. Adding `snapshot-error.json` shows a banner that
   reads "Stale" with the message and the last good snapshot's time, with that snapshot still
   drawn under it, and removing the file clears the banner. Replacing `snapshot.json` with
   text that isn't JSON shows the same banner over the last snapshot. A server started with
   no snapshot reads "No state yet". Stopping the server shows "Not connected" within 10
   seconds, and starting it again clears that.
6. **Words and contrast.** An axe-core scan with the WCAG 2.2 AA rules passes on every page in
   both themes. The test first reads the page's background color in each theme and fails if
   the two are equal. A test finds the words for a state inside the element that carries its
   color: each pass state, each widget state, the triangle, and each banner. A test finds no
   color literal in any `.css` or `.svelte` file under `web/src`, in a stylesheet, a `style`
   attribute, or a class such as `bg-[#fff]`, and no copy of Survey's files in `web/`.
7. **The polish.** The `standards:design-refinement` pass is one commit range. The round
   report lists its changes. Items 1 to 6 pass after it. The pull request holds screenshots of
   the front page, a round page, and a pass page in both themes, taken by one script from the
   two-release fixture at the commit before the range and at its last commit.
8. **The service.** After the install script runs, `curl -s http://127.0.0.1:7908/` returns a
   page holding the name of a project in the live state folder, and it does so again after
   `systemctl --user restart cc-desk-page.service`. `ss -ltn` shows port 7908 bound to
   `127.0.0.1` and to no other address.
9. **The terminal link, if it fits.** Opening `cc-desk://alpha/r1` through `xdg-open` runs the
   stand-in command with `alpha r1`. Each of these runs nothing: `cc-desk://alpha/r1;x`,
   `cc-desk://alpha/r1%3Bx`, `cc-desk://Alpha/r1`, `cc-desk://alpha/r1/x`,
   `cc-desk://alpha/../r1`, and `cc-desk://alpha/r1?x=1`. A click on the link in Chromium is
   tried by hand. The report says what Chromium asked, whether one answer held, and which of
   the link and the button shipped.
10. **The gate.** `bash scripts/gate.sh` passes, and it now runs the page's checks. A
    Playwright run with a skipped test, with no browser, or with fewer tests than the suite
    lists fails the gate.

What Geoff looks at is the mockup's `screens/front-page.png` and `screens/round-page.png`,
less the reports section and with the real state words.

## Backlog

| File | Verdict |
| --- | --- |
| `snapshot-stop-times.md` | Keep, reworded as a snapshot change for the Desk and skills round. This round shows a stopped pass with its reason and no "since", and shows "Nothing" with "no reason recorded" for the round the file names |
| `2026-11-17-rerun-prove-proofs.md` and the other seven | Keep. None touches the page |

## Calls Claude made

Each can be overturned at approval.

- The page polls every 2 seconds, with no pushed events.
- The port is 7908.
- Times show in Alaska time whatever the browser's zone, and the page has no theme switch.
- An item's address carries its project, as in `/alpha/r1`.
- The polish runs without Geoff, as the main design says.
- "The Survey gate" in the main design's done item 3 is read as: Survey at a pinned tag, no
  copy of its files, no color literal, and the contrast scan in both themes.
- "No form control except Mark as read" is read as item 4 words it. "Mark as read" arrives
  with reports, and "Copy command" is the only button until then.
- A rulings page is added. The main design asks that rulings show on the page and names no
  place for them.
