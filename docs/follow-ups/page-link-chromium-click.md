# Drop the "Copy command" button

Geoff clicked the "Open in a terminal" link in Chromium in his sitting of 2026-11-20. Chromium
asked once, he ticked "Always allow", and `~/.config/chromium/Geoff/Preferences` now holds
`"http://127.0.0.1:7908": {"cc-desk": true}`. The answer holds, so a later pass drops the "Copy
command" button from `web/src/lib/Command.svelte` and keeps the link.

The first clicks opened no terminal: the handler found no `cc-desk` on the session's `PATH`.
Pull request 26 makes it fall back to `~/.local/share/cc-desk/bin/cc-desk`.
