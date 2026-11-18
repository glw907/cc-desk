# The snapshot gives a pass no needs kind or reason

A round in the snapshot carries `needs.kind` and, for `nothing`, a `reason`. A pass carries only
its questions. A pass page with no questions therefore says "Nothing" and borrows the round's
reason, or reads "No reason recorded" when the round has none. The pass can have its own reason.
Expose a pass's needs kind and reason in `internal/snapshot/snapshot.go`, check it in
`e2e/events_test.go`, and have `web/src/routes/[project]/[round]/[pass]/+page.svelte` read it in
place of the round's.
