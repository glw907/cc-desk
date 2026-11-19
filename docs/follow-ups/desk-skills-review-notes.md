# Notes from pass 2 (desk-skills)'s whole-branch review

The review accepted the code and escalated one call. The two skills and the guide describe the
desk the design says pass 3 builds, and several of those behaviors aren't built yet. The
call kept them as written, with one caveat line at the top of the guide. Pass 3 should check
each item below against its code and remove the guide's caveat once they all hold.

1. **Typing `cc-desk` opens the desk.** Today the root command has no `RunE` and prints help
   (`cmd/cc-desk/root.go`).
2. **A merge question's answers act.** Nothing reads the answer yet, and `state.ClosePass` has no
   caller outside tests. The skill (`skills/cc-desk/SKILL.md`, "A merge question") and the guide
   say what merge, fix again, and close do.
3. **A ruling starts a free pass.** The skill's "The build starts within a minute" and its
   replace table's "Hasn't started" row assume the check starts a pass a ruling frees. Today it
   launches only a pass a command already started.
4. **A replacement restarts a building pass.** `checkReplace` in `internal/state/ruling.go`
   refuses once the pass has started. The skill's table and the guide say its session restarts
   with the new ruling.
5. **Where a permission prompt's details come from.** `cc-desk prompt yes|no` needs a project, a
   record, and a follow-ups folder, and the brief gives the desk none of them. The brief or the
   skill should say where the desk reads them.
6. **The `$(cat <file>)` form under the desk's allow rule.** No run has shown that the desk's
   allow rule matches `ruling record ... -- "$(cat <file>)"`. Pass 3's live desk run should.
