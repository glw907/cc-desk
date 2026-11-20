package main

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/glw907/cc-desk/internal/launch"
	"github.com/glw907/cc-desk/internal/sessions"
	"github.com/glw907/cc-desk/internal/state"
)

// isMergeDecision reports whether the decision is one of the six the merge
// table adds.
func isMergeDecision(decision string) bool {
	switch decision {
	case sessions.Merge, sessions.Merged, sessions.Fix, sessions.Ask, sessions.Cleanup, sessions.Close:
		return true
	}
	return false
}

// carryMerge carries out a merge, merged, fix, ask, or cleanup decision.
func (l *launcher) carryMerge(ctx context.Context, r sessions.Result) (string, error) {
	_, rf, err := l.liveRelease(r.Project)
	if err != nil {
		return "", err
	}
	if rf.Checkout == "" {
		return noCheckout(r.Project), nil
	}
	pass, ok := strings.CutPrefix(r.Record, "pass:")
	if !ok {
		return "", fmt.Errorf("record %s is not a pass", r.Record)
	}
	m := r.Merge
	switch r.Decision.Decision {
	case sessions.Merge:
		if err := l.commit(state.MarkMerge(r.Project, pass)); err != nil {
			return "", err
		}
		if err := l.callMerge(ctx, rf.Repo, m.Number, m.Commit); err != nil {
			return l.mergeFailed(r, pass, err)
		}
		if err := l.commit(state.MergePass(r.Project, pass, m.Number, l.now, false)); err != nil {
			return "", err
		}
		out, err := l.cleanup(ctx, r.Project, pass)
		return fmt.Sprintf("merged pull request %d; %s", m.Number, out), err
	case sessions.Merged:
		if err := l.commit(state.MergePass(r.Project, pass, m.Number, m.Merged, m.Outside)); err != nil {
			return "", err
		}
		out, err := l.cleanup(ctx, r.Project, pass)
		return fmt.Sprintf("pull request %d was merged; %s", m.Number, out), err
	case sessions.Cleanup:
		return l.cleanup(ctx, r.Project, pass)
	case sessions.Fix:
		return l.sendBack(ctx, r, pass)
	case sessions.Close:
		return l.closePass(ctx, r, pass)
	}
	return l.ask(r, pass)
}

// mergeCalls is how many merge calls in a row may fail before the check asks.
const mergeCalls = 3

// mergeFailed counts a failed merge call. The outcome is plain, so the check
// exits 0 and the log keeps the line; the third failure in a row asks in the
// same commit, which clears the count.
func (l *launcher) mergeFailed(r sessions.Result, pass string, callErr error) (string, error) {
	_, rf, err := l.liveRelease(r.Project)
	if err != nil {
		return "", err
	}
	n := 1
	for _, p := range rf.Passes {
		if p.Name == pass && p.PullRequest != nil {
			n = p.PullRequest.MergeFailures + 1
		}
	}
	msg := lastLine(callErr.Error())
	ev := state.RecordMergeFailure(r.Project, pass)
	if n >= mergeCalls {
		text, recommended := mergeQuestion(sessions.ReasonMergeFailed, r.Merge, msg)
		options := []string{sessions.OptionMerge, sessions.OptionFixAgain, sessions.OptionClose}
		ev = state.Sequence(ev, state.AskMerge(r.Project, pass, text, r.Merge.Commit, options, recommended))
	}
	if err := l.commit(ev); err != nil {
		return "", err
	}
	return fmt.Sprintf("merge call failed (%d of %d): %s", n, mergeCalls, msg), nil
}

// callMerge runs the one merge call: pinned to the commit the check read, a
// merge commit, with no delete, auto, or admin flag.
func (l *launcher) callMerge(ctx context.Context, repo string, number int, commit string) error {
	_, errOut, err := launch.Run(ctx, l.limit(), l.gh, "pr", "merge", strconv.Itoa(number), "-R", repo, "--merge", "--match-head-commit", commit)
	if err != nil {
		if strings.HasPrefix(err.Error(), "stopped at") {
			return err
		}
		msg := lastLine(errOut)
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}

// limit is the limit of one merge call, fetch, or pull.
func (l *launcher) limit() time.Duration {
	if l.callLimit > 0 {
		return l.callLimit
	}
	return launch.DefaultLimit
}

// lastLine is the last non-empty line of text.
func lastLine(text string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	return strings.TrimSpace(lines[len(lines)-1])
}

// sessionLive reports whether the session's row is live in a fresh list read.
func (l *launcher) sessionLive(ctx context.Context, shortID string) (bool, []sessions.Row, error) {
	rows, err := sessions.List(ctx, l.runner.Claude)
	if err != nil {
		return false, nil, err
	}
	row, found := sessions.Find(rows, shortID)
	return found && row.Live(), rows, nil
}

// ended is what stopSession found: the list read after the stop, and whether
// the stop ran.
type ended struct {
	rows    []sessions.Row
	stopped bool
	gone    bool
	// listedDone is true when the stop failed and the row stays listed done.
	listedDone bool
}

// step names the stop in a carry-out's outcome.
func (e ended) step(id string) string {
	if e.stopped {
		return "stopped session " + id
	}
	if e.listedDone {
		return "session " + id + " stop failed, row stays listed done"
	}
	return "session " + id + " already gone"
}

// stopSession reads the list, stops the session's row when it is working,
// blocked, or done, and reads the list again. The session is gone when the
// second read has no row for it, or a stopped, failed, or done one: a done
// row runs nothing, so it counts as ended whether the stop exited 0 or
// failed. A row still working or blocked after a failed stop returns the
// stop's error together with the list read after the stop. A row that is
// already stopped or failed takes no stop.
func (l *launcher) stopSession(ctx context.Context, shortID string) (ended, error) {
	rows, err := sessions.List(ctx, l.runner.Claude)
	if err != nil {
		return ended{}, err
	}
	row, found := sessions.Find(rows, shortID)
	if !found || row.Class() == sessions.ClassNone {
		return ended{rows: rows, gone: true}, nil
	}
	stopErr := l.runner.Stop(ctx, shortID)
	rows, err = sessions.List(ctx, l.runner.Claude)
	if err != nil {
		return ended{}, err
	}
	row, found = sessions.Find(rows, shortID)
	gone := !found || row.Class() == sessions.ClassNone || row.Class() == sessions.ClassDone
	if stopErr != nil && !gone {
		return ended{rows: rows}, stopErr
	}
	listedDone := stopErr != nil && found && row.Class() == sessions.ClassDone
	return ended{rows: rows, stopped: stopErr == nil, gone: gone, listedDone: listedDone}, nil
}

// sendBack stops the pass's session when live and relaunches it, uncounted,
// with the fix note. One commit holds the fix and the launch, so a failed
// stop or launch spends no attempt.
func (l *launcher) sendBack(ctx context.Context, r sessions.Result, pass string) (string, error) {
	s, _, err := l.liveRelease(r.Project)
	if err != nil {
		return "", err
	}
	sess := recordOf(s, r.Project, r.Record)
	if sess == nil {
		return "", fmt.Errorf("no session record to send the pull request back to")
	}
	live, _, err := l.sessionLive(ctx, sess.ShortID)
	if err != nil {
		return "", err
	}
	if live {
		if err := l.runner.Stop(ctx, sess.ShortID); err != nil {
			return "", err
		}
		if still, _, err := l.sessionLive(ctx, sess.ShortID); err != nil {
			return "", err
		} else if still {
			return "not relaunched: session " + sess.ShortID + " is still live after claude stop", nil
		}
	}
	return l.launchOrStop(ctx, r, launchRequest{
		Kind: state.LaunchRelaunch, Cause: sessions.ReasonFix, Counted: false, Replaces: sess.ShortID,
		Fix:  &launch.FixNote{Number: r.Merge.Number, Commit: r.Merge.Commit, Failures: r.Merge.Failures, Words: r.Merge.Words},
		Also: state.RecordFix(r.Project, pass, r.Merge.Failures),
	})
}

// ask adds the merge question for an outside-design or fix-limit decision.
func (l *launcher) ask(r sessions.Result, pass string) (string, error) {
	m := r.Merge
	text, recommended := mergeQuestion(r.Decision.Reason, m, "")
	options := []string{sessions.OptionMerge, sessions.OptionFixAgain, sessions.OptionClose}
	if err := l.commit(state.AskMerge(r.Project, pass, text, m.Commit, options, recommended)); err != nil {
		return "", err
	}
	return "asked about pull request " + strconv.Itoa(m.Number), nil
}

// mergeQuestion is the text of a merge question for the reason, and the
// option it recommends. The detail is the last line of a failed merge call.
func mergeQuestion(reason string, m sessions.MergeInfo, detail string) (text, recommended string) {
	switch reason {
	case sessions.ReasonOutsideDesign:
		return fmt.Sprintf("Pull request %d is outside the approved design. Merge it, send it back, or close it?", m.Number), sessions.OptionFixAgain
	case sessions.ReasonCallsForGeoff:
		what := "calls"
		if m.Calls == 1 {
			what = "call"
		}
		return fmt.Sprintf("Pull request %d leaves %d %s for you. Merge it, send it back, or close it?", m.Number, m.Calls, what), sessions.OptionMerge
	case sessions.ReasonHeadMoved:
		return fmt.Sprintf("Pull request %d has a new head, %s, since you answered. Merge it, send it back, or close it?", m.Number, shortCommit(m.Commit)), sessions.OptionMerge
	case sessions.ReasonChecksFailed:
		return fmt.Sprintf("Pull request %d now fails %s. Merge it anyway, send it back, or close it?", m.Number, strings.Join(m.Failures, ", ")), sessions.OptionFixAgain
	case sessions.ReasonMergeFailed:
		return fmt.Sprintf("The merge call for pull request %d has failed %d times (%s). Try the merge again, send it back, or close it?", m.Number, mergeCalls, detail), sessions.OptionFixAgain
	}
	return fmt.Sprintf("Pull request %d has failed %d times (%s). Merge it, send it back again, or close it?", m.Number, m.Attempts+1, strings.Join(m.Failures, ", ")), sessions.OptionClose
}

// shortCommit abbreviates a commit id for a question's text.
func shortCommit(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}

// closePass carries out a "close" answer, and is safe to run twice. It closes
// the pull request unless it already reads closed, stops the session and
// confirms it gone, then records the session ended and the pass closed in one
// commit. A stop that fails leaves the pass building, so the next run skips
// the close and finishes.
func (l *launcher) closePass(ctx context.Context, r sessions.Result, pass string) (string, error) {
	s, rf, err := l.liveRelease(r.Project)
	if err != nil {
		return "", err
	}
	m := r.Merge
	if !m.AlreadyClosed {
		if err := l.callClose(ctx, rf.Repo, m.Number); err != nil {
			return "", fmt.Errorf("close: %w", err)
		}
	}
	ended := []state.Event{}
	if sess := recordOf(s, r.Project, r.Record); sess != nil {
		live, _, err := l.sessionLive(ctx, sess.ShortID)
		if err != nil {
			return "", err
		}
		if live {
			if err := l.runner.Stop(ctx, sess.ShortID); err != nil {
				return "", err
			}
			if still, _, err := l.sessionLive(ctx, sess.ShortID); err != nil {
				return "", err
			} else if still {
				return "not closed in the record: session " + sess.ShortID + " is still live after claude stop", nil
			}
		}
		if sess.Ended == nil {
			ended = append(ended, state.EndSession(r.Project, r.Record))
		}
	}
	ended = append(ended, state.ClosePass(r.Project, pass, m.Ruling))
	if err := l.commit(state.Sequence(ended...)); err != nil {
		return "", err
	}
	return fmt.Sprintf("closed pull request %d and pass %s", m.Number, pass), nil
}

// callClose runs the one close call, with no comment and no branch delete.
func (l *launcher) callClose(ctx context.Context, repo string, number int) error {
	cmd := exec.CommandContext(ctx, l.gh, "pr", "close", strconv.Itoa(number), "-R", repo)
	var errOut bytes.Buffer
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		msg := lastLine(errOut.String())
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}

// cleanup ends the merged pass's session, pulls the checkout when it is safe,
// removes the worktree and branch with no force flag, and records the notes
// of whatever it left. A note is a report line; only a git error before a
// refusal can be read is a failure.
func (l *launcher) cleanup(ctx context.Context, project, pass string) (string, error) {
	s, rf, err := l.liveRelease(project)
	if err != nil {
		return "", err
	}
	sess := recordOf(s, project, "pass:"+pass)
	var notes []string
	var rows []sessions.Row
	step := ""
	if sess != nil {
		e, err := l.stopSession(ctx, sess.ShortID)
		if err != nil {
			return "", err
		}
		if !e.gone {
			return "", fmt.Errorf("session %s is still live after claude stop", sess.ShortID)
		}
		rows = e.rows
		step = e.step(sess.ShortID)
		if sess.Ended == nil {
			if err := l.commit(state.EndSession(project, "pass:"+pass)); err != nil {
				return "", err
			}
		}
	} else if rows, err = sessions.List(ctx, l.runner.Claude); err != nil {
		return "", err
	}
	if sess == nil || sess.Worktree == "" {
		notes = append(notes, "worktree left: the pass has no worktree on record")
	} else {
		var cwds []string
		for _, row := range rows {
			if row.Live() && row.ID != sess.ShortID {
				cwds = append(cwds, row.Cwd)
			}
		}
		wt := launch.Worktree{Checkout: rf.Checkout, Default: rf.DefaultBranch, Path: sess.Worktree, Branch: sess.Branch, Limit: l.limit()}
		pulled, note, err := wt.Pull(ctx, cwds)
		if err != nil {
			return "", err
		}
		if note != "" {
			notes = append(notes, note)
		}
		notes = append(notes, wt.Remove(pulled)...)
	}
	if err := l.commit(state.RecordCleanup(project, pass, notes)); err != nil {
		return "", err
	}
	out := "cleaned up"
	if step != "" {
		out = step + "; " + out
	}
	if len(notes) == 0 {
		return out, nil
	}
	return out + "; " + strings.Join(notes, "; "), nil
}
