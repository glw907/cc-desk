# Click the terminal link in Chromium and report what it asks

The page now carries an "Open in a terminal" link and a "Copy command" button, and nothing has
yet tried the link in Chromium. This pass couldn't: no Chromium profile exists for Claude to
drive, and Claude never opens yours. What Chromium asks, and whether one answer holds, are still
open.

To try it:

1. Open a round page on <http://127.0.0.1:7908> whose command shows "Open in a terminal". A
   project whose name holds `.` or `_` shows no link.
2. Click the link and note what the dialog asks.
3. If the dialog offers "Always allow", tick it, and click the link once more.
4. Close Chromium, or wait a minute, since Chromium writes `Preferences` late. Then run:

   ```bash
   jq '.protocol_handler.allowed_origin_protocol_pairs' ~/.config/chromium/Geoff/Preferences
   ```

The answer held if the output names `http://127.0.0.1:7908` with `"cc-desk": true`. Before the
click it held only `{"https://accounts.google.com": {"sms": true}}`.

If it held, a later pass drops the "Copy command" button. If Chromium asks every time, a later
pass drops the link and the handler, as round standing answer 17 says, and moves a button to each
line under "Waiting for you".

Until `cc-desk` is on a `PATH` the handler sees, the click opens no terminal, so it tests the
dialog only. See `cc-desk-on-browser-path.md`.
