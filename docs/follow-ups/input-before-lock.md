# Input files are parsed before the lock

`release add` and `round design` read and decode their input file before `state.Commit` runs. A
bad input exits 1 before the lock, leaves a stale `snapshot-error.json` in place, and hides a
broken state file behind the input error. The design says any verb removes the error file after a
clean load.

Defensible, since the input is not state. Decide whether to move the decode inside the commit.
Also minor: `internal/atomicfile` does not sync the parent folder after the rename.
