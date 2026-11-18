# Put `cc-desk` where the terminal link's handler finds it

The handler for `cc-desk://` looks for `cc-desk` on the `PATH` it starts with, which is the GNOME
session's when Chromium launches it. `cc-desk` isn't on that `PATH` today, so the link opens no
terminal.

The Desk and skills round should install `cc-desk` somewhere the session's `PATH` reaches, such
as `~/.local/bin`, and check it with `bash web/scripts/check-link.sh` run from a session that
started after the install.

The handler takes only names that match `[a-z0-9][a-z0-9-]*`. The state accepts project names
with `.` and `_`, so such a project shows the command and the button on its pages and no link.
Widen the handler's rule and `web/src/lib/link.ts` together if a project like that needs one.
