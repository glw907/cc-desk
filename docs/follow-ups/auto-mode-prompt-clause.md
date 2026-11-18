# Choose what replaces item 6's hook clause under auto mode

Done item 6 of the sessions design says "a third session launched with one rule left out records
its prompt." That can't happen as built. The launcher starts every session with
`--permission-mode auto` (`internal/launch/run.go:62`). In auto mode, a command that no allow
rule covers runs with no permission prompt, so the `PermissionRequest` hook never fires and
`cc-desk prompt hook` records nothing.

Three probe sessions on 2026-11-18 show it. Their settings, logs, and the one hook input are in
`docs/research/2026-11-18-auto-mode-hook-probe/`. Two auto-mode runs ran the command with no
prompt and no hook call. A default-mode control blocked on a permission prompt, and the hook wrote
its input.

With the hook unreachable, a count of zero prompts no longer says anything about the allow list.
It says only that no session stopped on a prompt. Pass 5 (`sessions-timer`) marks the clause unmet
and runs item 6 without it.

The choice is Geoff's. There are three options:

1. Record each auto-mode denial the way a prompt is recorded now, through Claude Code's
   `PermissionDenied` hook. That hook fires after the auto-mode classifier denies a call. Its
   payload doesn't carry the denial reason, and no probe here has tried it yet.
2. Drop the clause for good, and rely on the denials log for what auto mode refuses.
3. Launch sessions in another permission mode. That brings back the prompts auto mode removed,
   and a session that hits one waits until someone answers it.

Recommended: option 1, as a later pass. It keeps item 6's intent, which is to show the record
fires before trusting a zero count, and it keeps the sessions unattended.

Pass 5's live run adds its result below: whether the left-out rule raised a prompt, and any
denial row a fixture session caused.
