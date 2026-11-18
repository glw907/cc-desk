# Notes from pass 2's whole-branch review

The review accepted the branch once the unrecorded-folder refusal was fixed. These items stayed open.

1. **A stop on a row that is already gone.** `end` in `cmd/cc-desk/carryout.go` always runs `claude
   stop`. If the real command exits nonzero for an id that has left the list, a finished conductor is
   never marked ended and every check exits 2. Probe the real command, then treat a failed stop
   followed by a list with no live row as ended.
2. **The launch's known rows.** `launchRequest` passes only the old record's short id as `Known`.
   `Request.Known` means the rows read before the launch. When `claude --bg` prints no id, an older
   stopped row with the same name can make the new row ambiguous. Pass the ids of a list read taken
   just before the launch.
3. **Test gaps.** No test covers a counted conductor relaunch after its row was seen, a failed
   conductor relaunch (`StopSessionOnLaunchFailure`), or a row still live after `claude stop`
   (`carryout.go`, ("still live after `claude stop`"). `idlePass` turns the row idle before the check, so the
   confirm read would pass even if the stop did nothing.
4. **Exit code of a failed list read inside an action.** The plan's table gives exit 1; the usage
   path and the confirm read after a stop give exit 2 with the action's failure.
5. **Prompt wording.** An uncounted usage relaunch with no counted relaunches reads "This is relaunch
   1 of 2". Reword it for an uncounted relaunch.
