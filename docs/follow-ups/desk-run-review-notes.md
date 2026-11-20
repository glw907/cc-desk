# Six low findings from pass 6's closing review

The conductor's whole-branch read of pass 6 (`desk-run`, pull request 22) at `3c29181` returned
`accept`. It raised six Low findings, and none blocks the pass. Each one is a small fix for a
later pass, and the 1.1 round can take them together.

1. **A check that can't fail.** In `scripts/live-run.sh` (lines 791 to 796), the check "nothing
   typed after the answer" reads `typed.log`, and only the script's own `type_in` writes that
   file. The plan's condition was one typed entry in the pane log. The script counts the
   answer's words in `pane.log` (line 793) but never checks that count. Add the check on the
   pane log or the desk transcript.
2. **A check that passes on an error.** In `scripts/live-recover.sh` (lines 712 and 725),
   `check_two_rows` counts a list that jq can't read as 0 rows (`|| echo 0`). If `work-ids` is
   missing, every list reads as 0 and the check passes. Fail the check when jq fails or
   `work-ids` is empty.
3. **An untagged report matches every release.** `Report.Matches`
   (`internal/state/types.go:482`) lets a report with no release name match every release of
   its project. An old untagged release report would make each later release of that project
   look already reported. Today's six untagged reports are scratch round reports, so nothing
   breaks yet. Match untagged reports only to rounds, or tag the old ones.
4. **A crash during `release done` leaves two copies.** `release done` writes the done file
   (`internal/state/events.go:95`) and then removes the live file (`commit.go:120`). A crash
   between the two steps leaves both on disk, and a rerun writes a second done file. The load
   should notice a live and a done copy of one release, and the rerun should finish the
   removal.
5. **The setup check takes no lock.** `scripts/setup-release-check.sh` (lines 35 to 53) waits
   for the check's service to read inactive, then copies the folder without taking
   `check.lock`. A check can start in between. Take the lock for the copy.
6. **One install ignores the state folder.** `scripts/live-recover.sh:248` runs `timer install`
   without `--state-dir`, while every other call there passes `$STATE`. If
   `CC_DESK_STATE_DIR` is overridden, the units would point at the default folder.
   `live-run.sh:271` passes it, and `live-recover.sh` should too.

The executor left one more Low as it is, because a change would have meant a rerun. The comment
on the fast reads in `scripts/live-recover.sh` says they run every 3 seconds, but each read also
spends about a second on the listing, so they run about every 4. Fix the comment, or time the
reads from their start.
