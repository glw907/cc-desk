# How four stops clear once Geoff acts

`cc-desk check` records a stop for an expired sign-in (`sign_in_expired`), an unreadable block
(`unreadable_stop`), a relaunch limit (`relaunch_limit`), and a failed launch (`launch_failed`), and
the snapshot lists each under "Needs you now" with one thing to do. The sessions design says how a
stop is raised. It says nothing about how one clears after Geoff signs in, reads the log, raises the
limit, or fixes the launch.

Today a stopped pass leaves `stopped` only through `pass start`, which resumes it. The check never
reads the session list to learn that the cause is gone, so a stop for a sign-in that Geoff has
already renewed stays in "Needs you now" until someone resumes the pass.

Two ways to close it. Let the check clear a stop whose cause no longer reads (a live row that is no
longer blocked, a fresh sign-in), or keep clearing a manual act and name the command in each stop's
"one thing to do". Recommended: name the command now, and let pass 2 decide whether the check
should clear a stop itself, since clearing needs the same list read that relaunching does.

Geoff's call on whether the check clears stops by itself. Pass 2 of the sessions round carries out
launches and relaunches, so it is the pass that meets this first.
