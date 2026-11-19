# Remove the desk entry when `cc-desk` fails to exec `claude`

`cc-desk` registers the desk session in the state folder and then replaces itself with `claude`.
When the exec fails (no `claude` on PATH, or a refused launch), the process exits with an
error but the desk entry it registered stays, so the next `cc-desk` finds a session that
never started.

Decide whether the command removes its entry before it returns the exec error, or whether the
next run treats an entry with no live session as stale and replaces it.
