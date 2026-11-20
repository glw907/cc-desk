# The cleanup of a release marked done

The check cleans up a merged pass's session, worktree, and branch only on a live release.
`startFree`, the launch loop, `readPulls`, `LiveSessions`, and `fillSessions` all skip a release
marked done (`internal/sessions/decide.go:190`, `cmd/cc-desk/check.go:228`,
`internal/state/check.go:505`). Pass 6's `cc-desk release done <project>` is the one verb that marks
a release done. A pass merged in the run that closes the release, or one whose cleanup left a note,
keeps its worktree and branch, and nothing removes them.

Decide in `release done` whether it refuses while a merged pass has no `cleaned` mark, or whether
the check finishes cleanup on a done release. The refusal is recommended, since it adds no
done-release path to the check. If pass 6's verb already refuses, this item is done.
