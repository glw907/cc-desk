# The ready verb does not start the page's service

Closed by pass 5 (sessions-timer). `pass start`, `pass ready`, and a ruling that answers a question now start `cc-desk-check.service` with `systemctl --user start --no-block`, only against the live state folder (`TestKick`, seven cases). The design means the check's service here, not the Page round's service on port 7908, so the ready verb needs no page-service start.

`pass ready` records the pull request, its commit, and the session's verdict, and nothing more. A
round whose pass ends on a running service (the Page round's `http://127.0.0.1:7908/`) has no step
that starts it once the pull request is ready.

Decide in pass 5 whether the ready verb starts the service, or whether the pass's own script does,
and where the start is recorded so the check can tell a service that is down from one never started.
