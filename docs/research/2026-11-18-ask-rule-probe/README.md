# An `ask` rule stops a session in auto mode

A session in auto mode stops at a permission prompt when its settings name the command in an
`ask` rule. One background session, run on 2026-11-18 on Claude Code 2.1.327 in the
cc-desk-scratch checkout, shows it.

The session was `cc-desk-scratch-ask-probe`, on Sonnet, started with `--permission-mode auto`
and `settings.json` from this folder. Its task was to run `date -u` and reply `done`. The
settings name `Bash(date *)` and `Bash(date)` as `ask` rules and set a `PermissionRequest` hook
that saves its input.

Within 15 seconds the session's row in `claude agents --json` read `state: blocked`,
`status: waiting`, and `waitingFor: permission prompt`. The hook wrote `hook-input.json`. Its
`permission_mode` is `auto`, its `tool_input.command` is `date -u`, and its
`permission_suggestions` is null, so the input suggests no rule. `claude stop` ended the
session.

The earlier probe in `../2026-11-18-auto-mode-hook-probe/` named a whole command line with a
redirect in its `ask` rule and saw no prompt. Its record says a rule that fails to match would
look the same. A command that no rule names still runs with no prompt, as that probe found.
