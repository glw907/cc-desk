# A pull request closed without a merge

When GitHub reports a pull request as CLOSED, the check clears the pass's wait and decides nothing
further: the pass stays building with a ready report that points at a closed pull request. Nothing
names the case to Geoff or lets the pass reopen, rebuild, or end.

Decide what a closed pull request means for the pass (abandon, reopen, or a new pull request) and
whether it belongs in `needs_you_now`.
