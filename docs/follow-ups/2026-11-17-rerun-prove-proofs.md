# Open tests from the Prove pass

The Prove pass ran all six proofs and closed each with a result in
`docs/research/2026-11-17-prove-findings.md`. Seven items stay open. None blocks the State and rules
round, which starts from the two fallbacks.

1. **A real connection drop.** Proof 1 modeled a lost connection with a proxy that refused the
   tunnel. A dropped link (WiFi lost, lid shut) may behave otherwise. On the next real `API
   unavailable` block, give the session a few minutes to retry by itself, then try `own.sh respawn`
   and then `own.sh resume` on it, and record the row after each.
2. **Resume after stopping the original.** Proof 6's `resume` made a copy because the original,
   idle after its `respawn`, was still registered. Run it again with `own.sh stop` on the original first, and check
   whether the copy carries on as the only row. If it does, proof 6's fallback can narrow to "stop,
   then resume".
3. **A live expired sign-in and usage limit.** Proof 2 labeled both from recorded text in
   `state.json`, an undocumented file. When one happens, save `claude logs` for that session and see
   whether the documented command shows the same words, so the check needn't read the file.
4. **The sign-in expiry warning.** `loginExpiresAt` from `claude auth status --json` was not read
   in proof 2. Read it once and record whether it gives notice before a sign-in lapses.
5. **A supervisor restart after a kill.** In proof 6 the row had a new process id and read `working` 20
   seconds after `own.sh kill`, with no command from the pass. The saved rows don't show whether the
   supervisor started it again. Repeat the kill and save the process list too.
6. **A landing before the check registers.** In proof 4, `approval-queue land` merged scratch PR 1
   at 03:28:32, three seconds before its `gate` check started. Land refuses a failing check, and
   this shows it doesn't wait for one that hasn't appeared yet. Decide in the State and rules round
   whether `land` should wait for a check to register.
7. **A fresh launch after a failed wake.** Proof 6's fallback rests on a fresh session that starts
   from the plan and branch. The run never tried it, because the `resume` copy carried on. Try it
   once, and record what the new session reads and how many commits it repeats.
