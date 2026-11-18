# A crash can leave a ruling that answers nothing

`cc-desk ruling record` appends the ruling line to `rulings.jsonl` before it writes the release
files that name it. A crash between the two leaves a ruling no question points at. The state still
loads, which is why the append comes first.

A retry after the crash records a second ruling with the same words and a new id, and that one
answers the question. The first stays in the log as an orphan. Nothing reads it wrongly, but the
log then holds two lines for one decision.

Two fixes are open. Have the retry notice an orphan with the same words, the same desk, and no
question pointing at it, and reuse its id. Or leave the log append-only and mark the orphan in the
snapshot so a reader can tell it from a ruling that answers something. Recommended: the second,
since it never rewrites a line.

The code is `RecordRuling` in `internal/state/ruling.go` and the append order in
`internal/state/commit.go`.
