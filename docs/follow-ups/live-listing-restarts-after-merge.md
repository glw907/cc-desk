# The live listing restarts after the merge

In pass 5's run 5 (2026-11-20, `e6b7641`), `scripts/live-repairs.sh` started its listing reader
at 06:44:30Z, 8 seconds after the fixture merge at 06:44:22Z. The conductor phase then took about
12 minutes, and that reader finished first. The listing phase called `start_listing` again at
06:59:35Z. Its guard skips the restart only while the reader's PID is alive, so it started a
second reader and truncated `listing.out`, which erased the reads taken from the merge. The ten
reads in the pull request ran from 07:00:35Z to 07:09:37Z, 16 to 25 minutes after the merge.

The design's item 7 ("a listing each minute for 10 minutes shows the session gone") is met as
written, and pull request 21 landed on it under standing answer 26. The plan's check, "one a
minute from the merge", wasn't met, so the run never showed the merged session's row in the
first 10 minutes. The check's log line shows the stop step ran and `claude stop` exited 0, but a
done row can stay listed after a clean stop, so that line alone doesn't show the row gone.

Fix: make the listing phase keep a finished reader's output (skip the restart when `listing.out`
holds its ten reads, and never truncate it once the merge phase has written it). Pass 6's done
item 9 checks the session, worktree, and branch gone within 10 minutes of a merge; its check
reads from the merge, so a script it reuses needs this fix first.
