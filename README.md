# cc-desk

cc-desk shows what Claude Code is building, forecasts when the work will next need a decision,
and takes each decision as a conversation in a terminal.

The `cc-desk` command keeps the state files. It reads and strictly checks every one under one
lock, records a pass's start, finish, and stop, adds a release or a round's design, and prints a
snapshot of it all as JSON. It also records Geoff's rulings, one line each, and only from a desk
session. Typed alone on a terminal, `cc-desk` opens that desk session, and `cc-desk <project>
<round>` opens it on one round. `cc-desk check` reads the sessions, launches or relaunches work, and merges ready pull
requests. A pass reports itself ready with `cc-desk pass ready --verdict accept|fix|escalate
--calls <n>`. The review gives the verdict, and `--calls` is the number of points the pull request
leaves for Geoff to decide. The check merges a pull request only when that number is 0, and
otherwise asks Geoff. The `--call-limit` flag on `cc-desk check` sets how long the merge call, the
fetch, and the pull each may run. The default is 60 seconds.

Every session the check launches carries a `PermissionDenied` hook that runs `cc-desk refusal
hook`. The hook records each command that auto mode refuses on the session's record. It prints
nothing and exits 0, so the session carries on. A failure is one line in `refusal-hook.log` under
the state folder.

`cc-desk release done <project>` ends a project's live release. Its file moves to
`releases/done/`, so the project can take a new release, and each report names the release it
belongs to. The command refuses while one of the release's records holds a live session.
`cc-desk session end pass <project> <name>` and `cc-desk session end conductor <project>` end the
record's open session. Both commands wait for the check's lock first.

`bash scripts/setup-release.sh` adds the cc-desk release that the real run starts from, and
`bash scripts/setup-release-check.sh` checks it on a copy of the state folder. Two scripts run
cc-desk for real against the scratch repo. `scripts/live-run.sh` rehearses a whole release, from
the desk's ruling to the release report. `scripts/live-recover.sh` checks that the check brings
back a conductor and a pass after each death, stops the pass at its relaunch limit, and stops a
pass at an `ask` rule. Each runs one phase per call; `live-run.sh` with no phase runs them all.

`cc-desk timer install` puts the binary and a systemd user timer in place to run the
check every five minutes, and `cc-desk timer status` says when it last ran. `bash scripts/gate.sh`
runs the checks, and fails on any skipped test.

## Install

`cc-desk timer install` copies the binary to `~/.local/share/cc-desk/bin/cc-desk`. To run it by
name from a shell, link it onto your `PATH`:

```bash
cc-desk timer install
ln -sf ~/.local/share/cc-desk/bin/cc-desk ~/.local/bin/cc-desk
```

The approved design is
[`docs/specs/2026-11-17-cc-desk-design.md`](docs/specs/2026-11-17-cc-desk-design.md), with a
working mockup of the page beside it. [`docs/STATUS.md`](docs/STATUS.md) says what happens next.

## The page

The page in `web/` is a status page that only reads. It shows the releases, rounds, and passes in
the state folder and the forecast of the next decision session. It holds one form, "Mark as read" on a
report's page, and no button but "Copy command". Its server answers every other write with 405,
or 403 from another origin.

Run `bash web/scripts/install-page.sh` to install it, or to serve a newer build. The script builds
the page, copies the build to `~/.local/share/cc-desk/page`, and installs and restarts the user
unit `cc-desk-page.service`. The page is at <http://127.0.0.1:7908/>. It reads "No state yet"
until the `cc-desk` command writes the live state folder. The Move round records the install in the
workstation repo.

### The terminal link

A round page, a pass page, and each line under "Waiting for you" carry an "Open in a terminal"
link to `cc-desk://<project>/<round>`. A handler registered for that scheme opens a terminal that
runs `cc-desk <project> <round>`.

Run `bash web/scripts/install-link.sh` to register the handler. The script copies it to
`~/.local/share/cc-desk/link/`, writes `~/.local/share/applications/cc-desk.desktop`, and rebuilds
the desktop database. It leaves `~/.config/mimeapps.list` alone. Then run
`bash web/scripts/check-link.sh` to confirm that `xdg-open` reaches the handler.

The handler takes an address only when the project and the round each start with a lower-case
letter or a digit and hold only those, `.`, `_`, and `-`, which is the rule for a project's name in the
state. It refuses every other address and runs nothing. A name that breaks the rule shows the
command and the button, and no link.

The handler looks for `cc-desk` on the `PATH` it starts with, then at
`~/.local/share/cc-desk/bin/cc-desk`, where `cc-desk timer install` copies it. A browser started
from the desktop has no `~/.local/bin` on its `PATH`, so the second place is the one it finds. The
handler refuses when neither holds `cc-desk`. Chromium asks before it opens the link the first
time, and remembers an "Always allow".
