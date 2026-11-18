# The snapshot shows a stop reason but no stop time

A pass in the snapshot carries `stop_reason` and no stop times, so a stopped pass has no "stopped
since" time. The design table asks for event times. Expose the stop intervals, or the open stop's
start, in `internal/snapshot/snapshot.go` and check it in `e2e/events_test.go`.

Also for pass 2: a designed round with no questions and `needs_geoff: true` shows needs `nothing`
with a `null` reason (alpha r1 in the fixture).
