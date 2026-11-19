# Remove a closed pass's worktree and branch

When Geoff answers a merge question with "close", the check closes the pull request, ends the
session, and records the pass as closed. The pass's worktree and branch stay on disk, because
cleanup runs only for a merged pass (`internal/state/merge.go`, the `Merged` test).

Decide whether a closed pass is cleaned up the way a merged one is, and what it keeps if its branch
holds work Geoff may want back.
