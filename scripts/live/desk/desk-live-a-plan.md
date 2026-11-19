# desk-live-a plan

This file is the plan for the pass `desk-live-a`. No other plan is written for it, so do not
plan the pass again, run a review, or update STATUS. The steps below are the whole pass.

Run each step from the pass's worktree, in order, and stop after the last one.

1. Write the file `notes-desk-live-a.md` with the one line `desk-live-a ran.`
2. Run `bash scripts/gate.sh`.
3. Run `git add notes-desk-live-a.md`.
4. Run `git commit -m "Add desk-live-a notes"`.
5. Run `git push origin desk-live-a`.
6. Run `gh pr create --repo glw907/cc-desk-scratch --head desk-live-a --title "Add desk-live-a notes" --body "Fixture pass desk-live-a."` and note the pull request number it prints.
7. Run `git rev-parse HEAD`.
8. Run the ready command your launch prompt names, adding `--pr <number> --commit <sha> --verdict accept --design inside`.

The pass is done once the ready command has run.
