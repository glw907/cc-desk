package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/glw907/cc-desk/internal/launch"
	"github.com/glw907/cc-desk/internal/sessions"
	"github.com/glw907/cc-desk/internal/snapshot"
	"github.com/glw907/cc-desk/internal/state"
)

// carryOut carries out each decision in the order printed and gives each
// one's outcome. It runs under check.lock and outside the state lock. A
// failed action does not end the run: the rest still run, and the error that
// comes back names every record that failed, as a refusal (exit 2). A failed
// commit is the one error that ends the run, as exit 1.
func (l *launcher) carryOut(ctx context.Context, results []sessions.Result) ([]string, error) {
	outcomes := make([]string, len(results))
	var failed []string
	for i, r := range results {
		out, err := l.carry(ctx, r)
		var commitErr *commitError
		if errors.As(err, &commitErr) {
			return nil, commitErr.err
		}
		if err != nil {
			out = "failed: " + err.Error()
			failed = append(failed, fmt.Sprintf("%s/%s: %v", r.Project, r.Record, err))
		}
		outcomes[i] = out
	}
	if len(failed) > 0 {
		return outcomes, state.Refuse("a launch or an end failed after the rest ran: " + strings.Join(failed, "; "))
	}
	return outcomes, nil
}

// commitError marks a failed commit, which ends the run.
type commitError struct{ err error }

func (e *commitError) Error() string { return e.err.Error() }

// carry runs one decision. It gives the outcome, or an error when the action
// failed. Before any side effect it checks the release's checkout.
func (l *launcher) carry(ctx context.Context, r sessions.Result) (string, error) {
	if isMergeDecision(r.Decision.Decision) {
		return l.carryMerge(ctx, r)
	}
	switch r.Decision.Decision {
	case sessions.Start:
		return "started", nil
	case sessions.WaitUsage:
		if strings.HasPrefix(r.Detail, sessions.ResetWaitPrefix) {
			return "usage limit holds the relaunch; " + r.Detail, nil
		}
		return "", nil
	case sessions.Launch, sessions.LaunchAgain, sessions.Relaunch, sessions.EndAndRelaunch, sessions.End:
	default:
		return "", nil
	}
	s, rf, err := l.liveRelease(r.Project)
	if err != nil {
		return "", err
	}
	if rf.Checkout == "" {
		return noCheckout(r.Project), nil
	}
	sess := recordOf(s, r.Project, r.Record)
	d := r.Decision

	switch d.Decision {
	case sessions.End:
		if sess == nil {
			return "", fmt.Errorf("no session record to end")
		}
		e, err := l.stopSession(ctx, sess.ShortID)
		if err != nil {
			return "", err
		}
		if !e.gone {
			return "", fmt.Errorf("session %s is still live after claude stop", sess.ShortID)
		}
		out := "ended " + sess.ShortID
		if err := l.commit(state.EndSession(r.Project, r.Record)); err != nil {
			return out, err
		}
		if r.Record == "conductor" {
			note, err := l.removeConductorWorktree(ctx, r.Project, sess)
			if err != nil {
				return out, err
			}
			if note != "" {
				out += "; " + note
			}
		}
		return out, nil
	case sessions.Launch:
		if r.Record == "conductor" {
			req, ok := conductorRequest(s, rf, sess)
			if !ok {
				return "not carried out: no report is due for " + r.Project, nil
			}
			return l.launchOrStop(ctx, r, req)
		}
		return l.launchOrStop(ctx, r, launchRequest{Kind: state.LaunchFirst})
	case sessions.LaunchAgain:
		if sess == nil {
			return "", fmt.Errorf("no session record to launch again")
		}
		return l.launchOrStop(ctx, r, launchRequest{Kind: state.LaunchAgain, Replaces: sess.ShortID})
	}

	// A relaunch, from here: the old session is stopped and confirmed gone
	// before the new one starts.
	if sess == nil {
		return "", fmt.Errorf("no session record to relaunch")
	}
	counted := d.Counted != nil && *d.Counted
	req := launchRequest{Kind: state.LaunchRelaunch, Cause: d.Reason, Counted: counted, Replaces: sess.ShortID}
	usage := d.Decision == sessions.Relaunch && d.Reason == string(state.StopUsage)
	missing := d.Decision == sessions.Relaunch && !usage
	relaunch := func() (string, error) {
		out, err := l.launchOrStop(ctx, r, req)
		if err == nil && usage && counted {
			out += "; counted relaunch for usage, the stop named no reset time"
		}
		return out, err
	}

	if usage {
		// The pass was resumed in the record commit; a live row for the old
		// id is stopped first. A failed list read marks nothing: the next
		// check reads again.
		rows, err := sessions.List(ctx, l.runner.Claude)
		if err != nil {
			return "", err
		}
		if row, found := sessions.Find(rows, sess.ShortID); !found || !row.Live() {
			return relaunch()
		}
	}
	if d.Decision == sessions.EndAndRelaunch {
		e, err := l.stopSession(ctx, sess.ShortID)
		if err != nil {
			return "", err
		}
		if !e.gone {
			return "not relaunched: session " + sess.ShortID + " is still live after claude stop", nil
		}
	} else if !missing {
		if err := l.runner.Stop(ctx, sess.ShortID); err != nil {
			if usage {
				return "", l.failedRelaunch(r, err)
			}
			return "", err
		}
		rows, err := sessions.List(ctx, l.runner.Claude)
		if err != nil {
			return "", err
		}
		if row, found := sessions.Find(rows, sess.ShortID); found && row.Live() {
			if usage {
				return "", l.failedRelaunch(r, fmt.Errorf("session %s is still live after claude stop", sess.ShortID))
			}
			return "not relaunched: session " + sess.ShortID + " is still live after claude stop", nil
		}
	}
	return relaunch()
}

// launchOrStop runs the launch and, when it fails, stops the pass as
// launch_failed with the error, or ends the conductor's or design writer's
// record the same way. A first launch of a conductor or design writer has no
// record to mark.
func (l *launcher) launchOrStop(ctx context.Context, r sessions.Result, req launchRequest) (string, error) {
	id, err := l.Launch(ctx, r.Project, r.Record, req)
	if err != nil {
		if r.Record == "conductor" && (req.Kind == state.LaunchFirst || req.Kind == state.LaunchContinue) {
			return "", l.failedConductorStart(r, req, err)
		}
		return "", l.failedRelaunch(r, err)
	}
	return "launched " + id, nil
}

// conductorRequest gives the launch for a due report. An ended record for the
// same report is continued, as a counted relaunch; any other record, or none,
// starts a new one. It reports false when no report is due now, as when one
// was recorded between the check's record step and this one.
func conductorRequest(s *state.State, rf *state.ReleaseFile, sess *state.Session) (launchRequest, bool) {
	due, ok := sessions.ReportDue(s, rf)
	if !ok {
		return launchRequest{}, false
	}
	req := launchRequest{Kind: state.LaunchFirst, ForReport: due}
	if sess != nil && sess.Ended != nil && sess.ForReport == due {
		req.Kind, req.Replaces, req.Counted, req.Cause = state.LaunchContinue, sess.ShortID, true, "no_report"
		if sess.StopReason == state.StopLaunchFailed {
			req.Cause = "launch_failed"
		}
	}
	return req, true
}

// failedConductorStart writes the ended conductor record a failed start for a
// report leaves, so the next check counts it, and gives back the error or the
// commit's error when that fails too.
func (l *launcher) failedConductorStart(r sessions.Result, req launchRequest, cause error) error {
	_, rf, err := l.liveRelease(r.Project)
	if err != nil {
		return err
	}
	pl, err := place(rf, "conductor")
	if err != nil {
		return err
	}
	ev := state.RecordConductorLaunchFailure(r.Project, state.Launch{
		Kind: req.Kind, Cause: req.Cause, Counted: req.Counted,
		Session: state.Session{Name: pl.name, Kind: pl.sessionKind, Worktree: pl.worktree, Branch: pl.branch, ForReport: req.ForReport},
	}, cause.Error())
	if err := l.commit(ev); err != nil {
		return err
	}
	return cause
}

// removeConductorWorktree takes the conductor's worktree and branch away with
// the helper a pass's cleanup uses, once the report the conductor was started
// for is recorded. It gives the note of what it left, or "removed the
// conductor worktree and branch" when nothing stayed.
func (l *launcher) removeConductorWorktree(ctx context.Context, project string, sess *state.Session) (string, error) {
	s, rf, err := l.liveRelease(project)
	if err != nil {
		return "", err
	}
	if sess.Worktree == "" || !sessions.ReportRecorded(s, project, sess.ForReport) {
		return "", nil
	}
	rows, err := sessions.List(ctx, l.runner.Claude)
	if err != nil {
		return "", err
	}
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
	notes := wt.Remove(pulled)
	if note != "" {
		notes = append([]string{note}, notes...)
	}
	if len(notes) == 0 {
		return "removed the conductor worktree and branch", nil
	}
	return "removed the conductor worktree; " + strings.Join(notes, "; "), nil
}

// failedRelaunch records a failed launch on the record and gives back the
// error, or the commit's error when that fails too.
func (l *launcher) failedRelaunch(r sessions.Result, cause error) error {
	var ev state.Event
	if name, ok := strings.CutPrefix(r.Record, "pass:"); ok {
		ev = state.StopOnLaunchFailure(r.Project, name, cause.Error())
	} else {
		ev = state.StopSessionOnLaunchFailure(r.Project, r.Record, cause.Error())
	}
	if err := l.commit(ev); err != nil {
		return err
	}
	return cause
}

// commit runs one event and marks a failure as the run-ending kind.
func (l *launcher) commit(ev state.Event) error {
	if _, err := state.Commit(l.dir, l.now, state.LockWait, ev, snapshot.Render); err != nil {
		var refused *state.RefusedError
		if errors.As(err, &refused) {
			return err
		}
		return &commitError{err}
	}
	return nil
}

// recordOutcomes writes the outcomes into check.json in one last commit.
func (l *launcher) recordOutcomes(outcomes []string) error {
	_, err := state.Commit(l.dir, l.now, state.LockWait, state.RecordOutcomes(outcomes), snapshot.Render)
	return err
}
