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
