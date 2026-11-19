# A leftover live session parks a pass, and the conductor didn't see it

Pass 5 (`desk-check`) parked at about 11:50 on 2026-11-19, with all of its code built and
reviewed. Its live script refuses to start while any scratch session is live, and one was: the
blocked hold pass from pass 5's own earlier probe, `96f2d2ee`. The script's `abort` step waits for
the check to end that session, but the check leaves a stopped pass's session alone by design
(`early` in `internal/sessions/decide.go` returns `none` for a stop with a reason). The wait could
never end. No `cc-desk` verb ends a session, and standing answer 52 keeps the executor from typing
`claude stop`, so it parked.

The same thing happened once before. Pass 3's live run left two blocked sessions, and the
conductor needed Geoff's word to stop them before pass 5 could launch.

The conductor then missed the park for about 85 minutes. Its watch waited only for the pull
request to turn ready, and a parked executor leaves the pull request a draft. Geoff found it by
asking how the round was going.

What should change:

1. **cc-desk.** A live script, or the check, needs a way to end a session that cc-desk launched
   and whose pass is stopped. The Fable review's "sessions nobody owns" item is the page half of
   this. The Move design should settle whether the desk, a verb, or the check ends one.
2. **The plan.** A plan whose live script waits on the check should name the check's rule that
   ends the wait, quoted from the code. This plan didn't, and its plan review didn't catch it.
3. **The conductor.** A watch on a building pass ends on either the pull request turning ready or
   the executor's state turning blocked or done. This is now how the conductor watches, and it
   belongs in `pass-core`'s `references/conductor.md`.

Standing answer 33 covers the rest of the desk round.
