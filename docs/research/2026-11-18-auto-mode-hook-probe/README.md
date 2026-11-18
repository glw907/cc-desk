# Auto mode and the prompt hook: findings

Under `--permission-mode auto`, a command that no allow rule covers raises no permission
request, so the `PermissionRequest` hook never fires. The launcher starts every session in that
mode (`internal/launch/run.go:62`). Done item 6's clause "a third session launched with one rule
left out records its prompt" can't be met as the design writes it. Three background sessions,
run on 2026-11-18 on Claude Code v2.1.327, show it.

| Run | Mode and settings | What happened |
| --- | --- | --- |
| `04eb1db9` | Auto, an empty allow list, the logging hook | The command ran with no prompt. The hook wrote nothing. The row read `done`. |
| `7a406e12` | Auto, plus an `ask` rule naming the command | The command ran with no prompt. The hook wrote nothing. |
| `6f7ef48d` | Default mode, the same settings as `04eb1db9` | The row went `blocked`, waiting on `permission prompt`. The hook wrote its input, and the command never ran. |

Each session got the same one-line task: run `date -u +%s` into a probe file and reply `done`.
No settings file on the workstation allows `date`. `settings-unlisted.json` and
`settings-ask.json` are the two settings files. `log-hook.sh` is the hook, which saves its input
and exits 0. The two auto-mode screens are `log-04eb1db9.txt` and `log-7a406e12.txt`. The
default-mode run's log was gone after `claude stop 6f7ef48d`, so its row read is the record.

## What the `ask` run shows

The `ask` rule named the whole command, with its redirect and path. A rule that fails to match
would give the same result as a rule auto mode ignores. This finding claims nothing about `ask`.

## The hook input

`hook-input-default-mode.json` is the input the hook got in the default-mode run. Its
`session_id` equals the row's `sessionId` and the job's `state.json` `sessionId`. That answers
the session-id half of `docs/follow-ups/hook-input-from-a-real-session.md`.

The input carries `session_id`, `transcript_path`, `cwd`, `scratchpad_dir`, `prompt_id`,
`permission_mode`, `effort`, `hook_event_name`, `tool_name`, `tool_input`, and
`permission_suggestions`. Its one suggestion has the type `addDirectories`, with a list of
directories and no rules. `parseHookPrompt` (`cmd/cc-desk/prompt.go:70`) reads only `addRules`
suggestions, so this input gives a prompt with no rule. No test covers that type yet.

## A hook that does fire in auto mode

Claude Code has a `PermissionDenied` hook that fires after the auto-mode classifier denies a
call. It came in v2.1.88 or v2.1.89. A hook that returns `{"retry": true}` lets the model try
again, and the payload doesn't carry the denial reason. These pages describe it:

- <https://claudeissues.com/issue/41261-docs-hooks-reference-missing-permissiondenied-event-and-retry-contract>
- <https://claudeissues.com/issue/72589-feature-request-expose-auto-mode-denial-reason-in-permissiondenied-hook-payload>
- <https://hookstack.vercel.app/hook/permission-denied-auto-mode-log>
- <https://claudeissues.com/issue/92279-auto-mode-let-a-classifier-block-fall-back-to-a-permission-prompt-instead-of-a-h>

These are secondary sources, and none of them was probed here. The follow-up
`docs/follow-ups/auto-mode-prompt-clause.md` carries the choice for Geoff.
