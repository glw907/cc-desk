# Rerun proofs 1 to 6

The Prove pass stopped before its first launch because weekly usage read 87 percent, above the plan's
85. Proofs 1 to 6 weren't attempted. `docs/plans/2026-11-17-prove-plan.md` still holds each proof's
steps, and `spikes/own.sh` is built and tested to launch and control the sessions.

To rerun:

1. Read usage with `claude -p "/usage" --model haiku < /dev/null` and record it. The ceiling is
   now 100 percent, and the limit pauses the pass rather than closing it (Geoff, 2026-11-17).
2. Run `bash spikes/own.sh start`, then proof 5, then the others in the plan's order.
3. Write each proof's folder under `spikes/`, with a `run` script, and its `## Proof <n>` section in
   `docs/research/2026-11-17-prove-findings.md`.
4. Proof 4 is the goal Geoff set: zero prompts for a pass from worktree to merge.

Proof 1's real-drop test also stays open: on the next real `API unavailable` block, give
the session a few minutes to retry by itself, then try `own.sh respawn` and then `own.sh resume`
on that session.
