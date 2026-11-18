# cc-desk

cc-desk shows what Claude Code is building, forecasts when the work will next need a decision,
and takes each decision as a conversation in a terminal.

The `cc-desk` command so far keeps the state files. It reads and strictly checks every one under one
lock, records a pass's start, finish, and stop, adds a release or a round's design, and prints a
snapshot of it all as JSON. It also records Geoff's rulings, one line each, and only from a desk
session. `bash scripts/gate.sh` runs the checks, and fails on any skipped test.

The approved design is
[`docs/specs/2026-11-17-cc-desk-design.md`](docs/specs/2026-11-17-cc-desk-design.md), with a
working mockup of the page beside it. [`docs/STATUS.md`](docs/STATUS.md) says what happens next.

## The page

The page in `web/` is a status page that only reads. It shows the releases, rounds, and passes in
the state folder and the forecast of the next decision session. It holds no form and no button
but "Copy command", and its server answers every request but GET and HEAD with 405.

Run `bash web/scripts/install-page.sh` to install it, or to serve a newer build. The script builds
the page, copies the build to `~/.local/share/cc-desk/page`, and installs and restarts the user
unit `cc-desk-page.service`. The page is at <http://127.0.0.1:7908/>. It reads "No state yet"
until the `cc-desk` command writes the live state folder. The Move round records the install in the
workstation repo.
