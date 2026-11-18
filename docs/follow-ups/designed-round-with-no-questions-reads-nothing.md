# A designed round with no questions reads "nothing" and gives no reason

A round that is designed, has `needs_geoff: true`, and holds no questions prints `needs.kind` as
`nothing` with a `null` reason. Round `r1` of `alpha` in the `desk` fixture is the case. The page
then says Nothing with no reason, which reads as a gap where the design meant to ask.

Decide which it is. If the round truly needs nothing, the design should record a reason, and
`round design` should refuse a round with `needs_geoff: true` and no questions. If it needs Geoff
but the questions are still to come, the snapshot should say so with its own kind. Either change
belongs in `internal/snapshot/snapshot.go` and `internal/state/events.go`, checked in
`e2e/events_test.go`.
