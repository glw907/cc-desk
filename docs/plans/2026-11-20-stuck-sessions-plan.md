# Stuck sessions: pass plan

Initiative: cc-desk
Round: stuck-sessions
Pass: stuck-sessions

Class `runner`: the change is in the timed check, which launches sessions with nobody watching.
Planned 60 minutes. Design: `docs/specs/2026-11-18-stuck-sessions-design.md`, which wins over
this plan. Geoff's ruling is `r-0021` in the state folder's `rulings.jsonl`, question `ss1`,
words "yes". It sets no condition, so one good read resumes a pass, and the design's limits
stand: two resumes for one pass's sign-in stops, 30 seconds for the read.

Worktree `/var/home/glw907/Projects/cc-desk/.claude/worktrees/stuck-sessions`, branch
`stuck-sessions`, cut from `origin/main` at `822e155`. PR base `main`, pushed to `glw907/cc-desk`.

Reviewed: one `standards:plan-reviewer` read, folded below.

## Pre-flight

No step needs Geoff. The pass runs `go`, `git`, `gh pr create`, `gh pr checks`, and the gate.
It reads `claude auth status` once to learn its output shape, and runs no live check against
the real state folder.

## Mechanism

The usage stop already resumes a pass by itself, and the sign-in resume reuses that path.

| Step | Where | What changes |
| --- | --- | --- |
| Read | `internal/sessions/read.go` | `SignedIn(ctx, claude, limit)` runs `<claude> auth status --json` in its own process group and kills the group at the limit. It reports good only on exit 0 with `"loggedIn": true`. Any failure, bad JSON, or timeout reads as not good |
| Gather | `cmd/cc-desk/check.go` `gatherEvidence` | The read runs once per check, and only when a live release holds a pass stopped on `sign_in_expired` whose session record is live and has resumes left. A hidden `--sign-in-limit` flag (default `sessions.SignInLimit`, 30 seconds) lets a test shorten the wait |
| Decide | `internal/sessions/decide.go` | A new row in `Decide`, before the early rows, for a pass stopped on `sign_in_expired` with a live record. With a good read and fewer than two resumes, it decides `relaunch`, reason `sign_in_expired`, uncounted. Under the usage hold it waits (`wait_usage`). Otherwise the early rows decide `none` as today |
| Resume | `internal/sessions/decide.go` `Plan` | The loop that resumes a usage relaunch through `StartPass` takes the sign-in relaunch too, so the stop gets its end time and the pass reads building in the record commit |
| Carry out | `cmd/cc-desk/carryout.go` | The sign-in relaunch takes the usage relaunch's branch: a live old row is stopped and confirmed gone, then the new session launches. A failed launch stops the pass as `launch_failed` |
| Count | `internal/state/types.go` | `Session.Resumes(cause)` counts the relaunches with a cause. `state.SignInResumeLimit` is 2, beside the stop reasons, so the snapshot and the decision read one number |
| Lines | `internal/snapshot/sessions.go` | `todoFor` takes the project and the pass. For a pass, the unreadable, relaunch-limit, and failed-launch lines end "then run cc-desk pass start <project> <pass>". The sign-in line reads "sign in to Claude Code, and the pass resumes by itself" while resumes are left, and names the command once they are spent. A conductor's or design writer's line keeps its wording, since `pass start` can't restart it |
| Guide | `docs/guide.md` | The sign-in row says the work resumes by itself; the other rows name the command the desk runs |

## Tasks

1. The read and its unit tests: good, signed out, bad output, and a stand-in that sleeps past a
   short limit and is ended at it. A test pins `SignInLimit` at 30 seconds. (10 minutes)
2. The decision row, the `Plan` resume, and `Resumes`, with table tests in
   `internal/sessions/decide_test.go`: not good stays `none`; good resumes uncounted; two spent
   resumes stay `none`; the hold waits; an ended record stays `none`. (15 minutes)
3. The carry-out branch and the e2e test on the launch rig, whose stand-in claude learns
   `auth status` with three controls (`auth-out`, `auth-bad`, `auth-sleeps`). (20 minutes)
4. The four lines, their unit tests, the e2e strings, and the guide. (10 minutes)
5. Gate, PR, whole-branch review, ready report. (15 minutes, off the 60)

## Acceptance

From the design's "What done looks like", for the answer "yes":

1. In the snapshot, each of the three other stops shows a line holding `pass start`, the
   project, and the pass.
2. A fixture pass stopped for the sign-in stays stopped while a stand-in for `claude auth status`
   reports no sign-in. On the first run after it reports a good one, the pass reads building,
   its stop has an end time, one launch is recorded, and the relaunch count is unchanged. After
   two such resumes a third stop stays, and its line names the command. A stand-in that sleeps
   is ended at 30 seconds and nothing changes.
3. `bash scripts/gate.sh` passes with no skipped test.

The design's item 3 belongs to the answer "no" and is not built.
