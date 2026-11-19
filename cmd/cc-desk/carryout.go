package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
		if err := l.runner.Stop(ctx, sess.ShortID); err != nil {
			return "", err
		}
		return "ended " + sess.ShortID, l.commit(state.EndSession(r.Project, r.Record))
	case sessions.Launch:
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

	if usage {
		// The pass was resumed in the record commit; a live row for the old
		// id is stopped first. A failed list read marks nothing: the next
		// check reads again.
		rows, err := sessions.List(ctx, l.runner.Claude)
		if err != nil {
			return "", err
		}
		if row, found := sessions.Find(rows, sess.ShortID); !found || !row.Live() {
			return l.launchOrStop(ctx, r, req)
		}
	}
	if !missing {
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
	return l.launchOrStop(ctx, r, req)
}

// launchOrStop runs the launch and, when it fails, stops the pass as
// launch_failed with the error, or ends the conductor's or design writer's
// record the same way. A first launch of a conductor or design writer has no
// record to mark.
func (l *launcher) launchOrStop(ctx context.Context, r sessions.Result, req launchRequest) (string, error) {
	id, err := l.Launch(ctx, r.Project, r.Record, req)
	if err != nil {
		return "", l.failedRelaunch(r, err)
	}
	return "launched " + id, nil
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
