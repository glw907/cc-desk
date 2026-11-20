# Notes from pass 2's whole-branch review

The whole-branch review returned fix on one blocking finding. A first launch could take over a
folder no record named, and the last commit of the branch refuses it. These items stayed open.

1. **The launch's known rows.** `launchRequest` passes only the old record's short id as `Known`.
   `Request.Known` means the rows read before the launch. When `claude --bg` prints no id, an older
   stopped row with the same name can make the new row ambiguous. Pass the ids of a list read taken
   just before the launch.
2. **Test gaps.** No test covers a counted conductor relaunch after its row was seen, a failed
   conductor relaunch (`StopSessionOnLaunchFailure`), or a row still live after `claude stop`
   (`carryout.go`, "still live after `claude stop`"). `idlePass` turns the row idle before the check, so the
   confirm read would pass even if the stop did nothing.
3. **Exit code of a failed list read inside an action.** The plan's table gives exit 1; the usage
   path and the confirm read after a stop give exit 2 with the action's failure.
4. **The merge deny rules against a bare command.** The settings deny `Bash(gh pr merge *)` and
   `Bash(approval-queue land *)`. Check, against Claude Code's rule matching, whether each also
   denies the command with no argument, since a bare `gh pr merge` merges the current branch's pull
   request. If not, the design must name the two bare rules; this pass does not change the deny list.
5. **An ended record whose folder was removed by hand.** `Launch` in `cmd/cc-desk/launcher.go` sends
   a `session start` over an ended record to `Reuse`, which fails with "the recorded worktree is
   missing" as a plain error. When the folder is absent, fall back to `Make`, which still refuses a
   leftover branch.
6. **The restart test's reach.** The restart in `TestRecoveryFinishedConductorStays` checks only the
   launch count. Assert the new short id and a cleared `ended` too.
