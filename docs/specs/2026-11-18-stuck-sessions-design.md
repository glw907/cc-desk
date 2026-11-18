# cc-desk: a stopped pass after Geoff fixes its cause

This is a small change to the timed check and to four lines the page shows. It is the first
design cc-desk builds for itself: Geoff answers its one question at the desk, and that answer
is his approval. The round that sets this up is
[2026-11-18-desk-skills-design.md](2026-11-18-desk-skills-design.md).

## What is wrong today

When a background session stops on something only Geoff can fix, the page lists the pass under
"Needs you now" with one thing to do. Four stops work this way:

| Stop | The page says |
| --- | --- |
| The sign-in to Claude Code expired | "sign in to Claude Code" |
| The session is blocked and the check can't read why | "read the session's log, then start the pass again" |
| The pass has used its two relaunches | "read the session's log, then start the pass again" |
| The launch failed | "read the launch error, then start the pass again" |

A stopped pass leaves that state only when someone runs `pass start` for it. After Geoff signs
in again, the line stays and the work doesn't resume. The other three lines say "start the
pass again" and never name the command.

## The question

This question is never guessed. The pass waits for Geoff's answer.

**Background.** As above. The check can tell when one of the four causes is gone: `claude auth
status` reports whether the sign-in is good, with no model call. For the other three the cause
is something Geoff read and judged, and the check has nothing to look at.

**Should cc-desk notice by itself that you fixed the problem?**

- **Yes.** After a sign-in stop, the check reads the sign-in every five minutes. Once it is
  good, the check removes the line and resumes the work. Geoff signs in and walks away. The
  other three lines name the command, because the check can't see those fixes.
- **No.** All four lines name the command. The work resumes when that command is run at the
  desk, so nothing restarts without a deliberate act.

**Recommended: yes.** Geoff shouldn't have to tell the system about a fix it can see. A wrong
resume costs little, since a session that meets the expired sign-in again stops again.

**What waits.** This one pass, planned at 60 minutes.

## What gets built

Under either answer, the three lines that say "start the pass again" name the command, with
the project and the pass filled in. The desk session can run it when Geoff says to.

Under "yes", the check also reads the sign-in on each run where a pass is stopped for it. A
good read resumes the pass: the stop gets its end time, the pass reads as building, and a new
session starts from the pass's branch. That relaunch isn't counted toward the limit of two.
If Geoff's words set a condition, such as more than one good read in a row, the pass builds
that.

Under "no", the sign-in line names the command as well, and the check resumes nothing.

## What it leaves out

- A stop on a permission prompt, which already clears when Geoff answers it at the desk.
- A stop on a usage limit, which already relaunches after the reset.
- Any new stop reason.

## When something goes wrong

| What happens | What the check does |
| --- | --- |
| `claude auth status` fails, or runs past 30 seconds | Treats the sign-in as not good and changes nothing |
| The resumed session stops on the sign-in again | The pass stops again, and the line returns |
| The same pass has been resumed twice for the sign-in | No third resume. The line names the command |

| Limit | Value |
| --- | --- |
| Good reads before a resume | 1, unless Geoff's answer sets another number |
| Resumes by the check for one pass's sign-in stops | 2 |
| Time allowed for the sign-in read | 30 seconds |

## Files

The decision in `internal/sessions/`, its carrying out in `cmd/cc-desk/`, the four lines in
`internal/snapshot/sessions.go`, and their tests. The page changes only by what the snapshot
says. The change touches no sign-in secret: it reads a yes or no.

## What done looks like

1. In the snapshot, each of the three other stops shows a line holding `pass start`, the
   project, and the pass.
2. Under "yes": a fixture pass stopped for the sign-in stays stopped while a stand-in for
   `claude auth status` reports no sign-in. On the first run after it reports a good one, the
   pass reads building, its stop has an end time, one launch is recorded, and the relaunch
   count is unchanged. After two such resumes a third stop stays, and its line names the
   command. A stand-in that sleeps is ended at 30 seconds and nothing changes.
3. Under "no": the same fixture stays stopped under a good sign-in, no launch is recorded, and
   all four lines name the command.
4. `bash scripts/gate.sh` passes with no skipped test.

Only the items for Geoff's answer are built and tested.

## Calls Claude made

- Only the sign-in stop clears itself under "yes". The other three have no cause the check can
  read.
- One good read is enough unless Geoff says otherwise, because a wrong resume only stops
  again.
- Two resumes, then the command, so an account that keeps losing its sign-in can't loop.
