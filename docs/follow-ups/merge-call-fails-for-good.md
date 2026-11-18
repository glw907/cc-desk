# A merge call that fails for good is retried every run

When `gh pr merge` fails for a lasting reason (a conflict, branch protection), the check records the
marker, exits 2 with the error, and tries again on the next run. The failure never reaches
`needs_you_now`, and neither the merge call nor the cleanup's `git fetch` and `pull` has a timeout,
so a hung call holds `check.lock`.

Decide how many failed calls make a merge question for Geoff, and set a timeout on both calls as
`ReadPulls` has one.
