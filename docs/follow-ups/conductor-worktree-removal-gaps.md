# Three gaps in removing a finished conductor's worktree

When a conductor session ends, `removeConductorWorktree` in `cmd/cc-desk/carryout.go` pulls the
default branch and removes the conductor's worktree and branch. Three cases fall short.

1. **A skipped pull leaves the branch.** When the pull is skipped, the `conductor` branch stays
   behind. The next conductor's `Make` refuses a leftover branch, so the next report that falls due
   cannot start one. Remove the branch when the pull is skipped and the branch holds nothing the
   default branch lacks, or make `Make` take over a branch with no worktree.
2. **No retry after a failed removal.** The session's record is committed as ended before the
   removal runs. When the removal fails, the check reports the error once, and the next check sees
   an ended record and does not try again. Record the removal as pending, or run it on every check
   until it succeeds.
3. **The failed-launch record carries a placeholder id.** A conductor launch that fails records its
   session with the short id `none`, which stands for no session. Any reader of short ids must know
   that `none` is not one. Store an empty id instead.
