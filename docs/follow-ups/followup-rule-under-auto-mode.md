# Check that the follow-up rule stands under auto mode

Closed by pass 5 (sessions-timer), and the rule stands. In the live run of 2026-11-18 `timer-live-b` filed a follow-up with the allowed `followup` command at 15:19:15Z, under `--permission-mode auto`. The release file records 0 prompts for it, the denials log holds no row for the scratch worktrees, and the file reached `origin/live-2026-11-18`.

Every launched session's settings file allows `Bash(<binary> --state-dir <dir> followup <project>
<record> *)`. The permission-modes page says auto mode drops broad allow rules that grant arbitrary
code execution, and some wildcarded paths. This rule has a like shape, and no run has shown whether
it stands.

In pass 5's live run, have the scratch pass file one follow-up and watch for a permission prompt on
the command. If auto mode drops the rule, the classifier decides each follow-up, and the design
should pick between a narrower rule (one per title) and a fixed folder the session writes to
without the verb.
