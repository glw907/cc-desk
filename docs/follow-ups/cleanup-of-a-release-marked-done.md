# The cleanup of a release marked done

The cleanup of a merged pass's session, worktree, and branch runs from a live release file. A
release marked done is skipped by the check, so a pass merged in the same run that closes the
release, or one whose cleanup was left with a note, may never have its worktree removed.

Decide whether the check finishes cleanup on a done release until each pass records `cleaned`, or
whether closing a release refuses while any cleanup is open.
