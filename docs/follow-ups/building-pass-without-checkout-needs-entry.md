# A building pass on a release with no checkout is silent in "Needs you now"

A pass that settles to building on a release with no recorded checkout gets a `launch` decision
each check, and `carry` in `cmd/cc-desk/carryout.go` answers it with the outcome "not carried out:
project <p> has no checkout recorded". The pass has no session record and no stop, so
`fillSessions` in `internal/snapshot/sessions.go` adds nothing to `needs_you_now`. The pass sits
building with no session, and the only sign is the outcome text in `check.json`.

That is deliberate for pass 1's fixtures, whose release has no checkout, but a real release in this
state would wait forever. Decide whether the snapshot should name a building pass that has no
session and no checkout, and which `release checkout` command is the fix.

Recommended: add one `needs_you_now` entry with the reason `no_checkout` and the todo "run cc-desk
release checkout <project> <path>", written for a release that has a building pass and no checkout.
