package main

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

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
			return "", fmt.Errorf("merge: %w", err)
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

// callMerge runs the one merge call: pinned to the commit the check read, a
// merge commit, with no delete, auto, or admin flag.
func (l *launcher) callMerge(ctx context.Context, repo string, number int, commit string) error {
	cmd := exec.CommandContext(ctx, l.gh, "pr", "merge", strconv.Itoa(number), "-R", repo, "--merge", "--match-head-commit", commit)
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
	var text, recommended string
	switch r.Decision.Reason {
	case sessions.ReasonOutsideDesign:
		text = fmt.Sprintf("Pull request %d is outside the approved design. Merge it, send it back, or close it?", m.Number)
		recommended = sessions.OptionFixAgain
	case sessions.ReasonHeadMoved:
		text = fmt.Sprintf("Pull request %d has a new head, %s, since you answered. Merge it, send it back, or close it?", m.Number, shortCommit(m.Commit))
		recommended = sessions.OptionMerge
	case sessions.ReasonChecksFailed:
		text = fmt.Sprintf("Pull request %d now fails %s. Merge it anyway, send it back, or close it?", m.Number, strings.Join(m.Failures, ", "))
		recommended = sessions.OptionFixAgain
	default:
		text = fmt.Sprintf("Pull request %d has failed %d times (%s). Merge it, send it back again, or close it?", m.Number, m.Attempts+1, strings.Join(m.Failures, ", "))
		recommended = sessions.OptionClose
	}
	options := []string{sessions.OptionMerge, sessions.OptionFixAgain, sessions.OptionClose}
	if err := l.commit(state.AskMerge(r.Project, pass, text, m.Commit, options, recommended)); err != nil {
		return "", err
	}
	return "asked about pull request " + strconv.Itoa(m.Number), nil
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
	if sess != nil {
		live, got, err := l.sessionLive(ctx, sess.ShortID)
		if err != nil {
			return "", err
		}
		rows = got
		if live {
			if err := l.runner.Stop(ctx, sess.ShortID); err != nil {
				return "", err
			}
		}
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
		wt := launch.Worktree{Checkout: rf.Checkout, Default: rf.DefaultBranch, Path: sess.Worktree, Branch: sess.Branch}
		pulled, note, err := wt.Pull(cwds)
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
	if len(notes) == 0 {
		return "cleaned up", nil
	}
	return "cleaned up; " + strings.Join(notes, "; "), nil
}
