package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glw907/cc-desk/internal/launch"
	"github.com/glw907/cc-desk/internal/snapshot"
	"github.com/glw907/cc-desk/internal/state"
	"github.com/spf13/cobra"
)

// launchFlags holds the hidden flags every verb that launches shares. They
// stand in for the real claude and systemd-run commands, Claude Code's jobs
// folder, and the wait for a new session's row, so a test never reaches them.
type launchFlags struct {
	claude     string
	systemdRun string
	jobsDir    string
	rowWait    time.Duration
	// secondWait is the check's wait before its second list read. The
	// session verbs take the flag and ignore it, so one stand-in command line
	// serves every verb.
	secondWait time.Duration
}

func (lf *launchFlags) register(cmd *cobra.Command) {
	cmd.Flags().StringVar(&lf.claude, "claude-command", "claude", "the command that runs claude")
	cmd.Flags().StringVar(&lf.systemdRun, "systemd-run-command", "systemd-run", "the command that runs systemd-run")
	cmd.Flags().StringVar(&lf.jobsDir, "jobs-dir", "", "the folder of Claude Code's job state (default ~/.claude/jobs)")
	cmd.Flags().DurationVar(&lf.rowWait, "launch-row-wait", 10*time.Second, "how long a launch waits for its session's row")
	cmd.Flags().DurationVar(&lf.secondWait, "second-read-wait", 30*time.Second, "the wait before the check's second list read")
	for _, name := range []string{"claude-command", "systemd-run-command", "jobs-dir", "launch-row-wait", "second-read-wait"} {
		_ = cmd.Flags().MarkHidden(name)
	}
}

// noCheckout is the outcome of an action on a release with no checkout.
func noCheckout(project string) string {
	return "not carried out: project " + project + " has no checkout recorded"
}

// launcher makes a worktree and starts a session in it, for the check and
// for the session verbs.
type launcher struct {
	dir    string
	binary string
	now    time.Time
	runner launch.Runner
}

// newLauncher resolves the absolute state folder, the running binary, and
// the jobs folder the settings and the launch use.
func newLauncher(f *rootFlags, lf *launchFlags) (*launcher, error) {
	dir, err := f.dir()
	if err != nil {
		return nil, err
	}
	if dir, err = filepath.Abs(dir); err != nil {
		return nil, fmt.Errorf("resolve state folder: %w", err)
	}
	now, err := f.clock()
	if err != nil {
		return nil, err
	}
	binary, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("find this command: %w", err)
	}
	if lf.jobsDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("find home folder: %w", err)
		}
		lf.jobsDir = filepath.Join(home, ".claude", "jobs")
	}
	return &launcher{
		dir: dir, binary: binary, now: now,
		runner: launch.Runner{Claude: lf.claude, SystemdRun: lf.systemdRun, JobsDir: lf.jobsDir, RowWait: lf.rowWait},
	}, nil
}

// placement is where a record's session runs and what it runs as.
type placement struct {
	name, worktree, branch, model, effort string
	promptKind                            string
	sessionKind                           state.SessionKind
	pass, round                           string
}

// place gives the name, worktree, branch, model, and effort a record's
// session has, by the design's table.
func place(rf *state.ReleaseFile, record string) (placement, error) {
	at := func(branch string) string { return filepath.Join(rf.Checkout, ".claude", "worktrees", branch) }
	switch {
	case record == "conductor":
		return placement{rf.Project + "-conductor", at("conductor"), "conductor", "sonnet", "medium", "conductor", state.KindConductor, "", ""}, nil
	case len(record) > 6 && record[:6] == "round:":
		round := record[6:]
		return placement{rf.Project + "-" + round + "-design", at(round + "-design"), round + "-design", "opus", "high", "design", state.KindDesign, "", round}, nil
	case len(record) > 5 && record[:5] == "pass:":
		name := record[5:]
		for _, p := range rf.Passes {
			if p.Name != name {
				continue
			}
			model, effort := "opus", "high"
			if p.Model != "" {
				model = p.Model
			}
			if p.Effort != "" {
				effort = p.Effort
			}
			return placement{rf.Project + "-" + name, at(name), name, model, effort, "pass", state.KindPass, name, p.Round}, nil
		}
	}
	return placement{}, fmt.Errorf("no record %s in project %q", record, rf.Project)
}

// liveRelease loads the state and finds the project's live release.
func (l *launcher) liveRelease(project string) (*state.State, *state.ReleaseFile, error) {
	s, err := state.Load(l.dir)
	if err != nil {
		return nil, nil, err
	}
	for _, rf := range s.Releases {
		if !rf.Done && rf.Project == project {
			return s, rf, nil
		}
	}
	return nil, nil, fmt.Errorf("project %q not found", project)
}

// recordOf gives the record's session, or nil.
func recordOf(s *state.State, project, record string) *state.Session {
	var found *state.Session
	s.LiveSessions(func(rf *state.ReleaseFile, rec string, sess *state.Session) {
		if rf.Project == project && rec == record {
			found = sess
		}
	})
	return found
}

// causeText is how a relaunch prompt names its cause.
func causeText(cause string) string {
	switch cause {
	case "usage":
		return "a usage stop"
	case "lost_connection":
		return "a lost connection"
	case "idle":
		return "sitting idle with its work unfinished"
	case "missing":
		return "its session vanishing from the session list"
	}
	return cause
}

// launchRequest says which launch of a record to make.
type launchRequest struct {
	Kind     state.LaunchKind
	Cause    string
	Counted  bool
	Replaces string
}

// Launch carries out one launch, in the design's order: refuse before any
// side effect, make or reuse the worktree, write the settings, fill the
// prompt, run claude, and commit the record. It gives the new short id. A
// failure after the worktree was made undoes what this call made, and one
// after claude started a session stops it first. A release with no checkout
// is the caller's to check, since its outcome is no launch at all.
func (l *launcher) Launch(ctx context.Context, project, record string, req launchRequest) (shortID string, err error) {
	s, rf, err := l.liveRelease(project)
	if err != nil {
		return "", err
	}
	if rf.Checkout == "" {
		return "", fmt.Errorf("%s", noCheckout(project))
	}
	pl, err := place(rf, record)
	if err != nil {
		return "", err
	}
	old := recordOf(s, project, record)
	if req.Kind == state.LaunchFirst && old != nil && old.Ended == nil {
		return "", state.Refuse(fmt.Sprintf("%s/%s already has a live session record, %s", project, record, old.Name))
	}
	var conflict string
	s.LiveSessions(func(orf *state.ReleaseFile, rec string, sess *state.Session) {
		if sess.Ended == nil && !(orf.Project == project && rec == record) && (sess.Worktree == pl.worktree || sess.Name == pl.name) {
			conflict = fmt.Sprintf("%s/%s", orf.Project, rec)
		}
	})
	if conflict != "" {
		return "", state.Refuse(fmt.Sprintf("%s/%s would use the worktree %s and the name %s, which the live record %s holds", project, record, pl.worktree, pl.name, conflict))
	}

	wt := launch.Worktree{Checkout: rf.Checkout, Default: rf.DefaultBranch, Path: pl.worktree, Branch: pl.branch}
	var undo func() error
	// A first launch takes over a folder only when this record's own ended
	// session made it; any other existing folder or branch is refused.
	if req.Kind == state.LaunchFirst && !(old != nil && old.Ended != nil && old.Worktree == pl.worktree) {
		if undo, err = wt.Make(); err != nil {
			return "", state.Refuse(err.Error())
		}
	} else if err = wt.Reuse(); err != nil {
		return "", err
	}
	defer func() {
		if err != nil && undo != nil {
			err = errors.Join(err, undo())
		}
	}()

	verbs := launch.Verbs(l.binary, l.dir, pl.promptKind, project, pl.pass, pl.round)
	settings, err := launch.WriteSettings(l.dir, l.binary, pl.name, rf.AllowRules, verbs)
	if err != nil {
		return "", err
	}
	p := launch.Prompt{
		Kind: pl.promptKind, Name: pl.name, Project: project, Round: pl.round, Pass: pl.pass,
		Checkout: rf.Checkout, Worktree: pl.worktree, Branch: pl.branch, Base: rf.DefaultBranch, Verbs: verbs,
	}
	var known []string
	if old != nil {
		known = []string{old.ShortID}
	}
	if req.Kind == state.LaunchRelaunch {
		count := 1
		if old != nil {
			count = max(old.Counted(), 1)
			if req.Counted {
				count = old.Counted() + 1
			}
		}
		p.Relaunch = &launch.RelaunchNote{Count: count, Cause: causeText(req.Cause)}
	}
	text, err := launch.Fill(p)
	if err != nil {
		return "", err
	}
	got, err := l.runner.Run(ctx, launch.Request{
		Name: pl.name, Model: pl.model, Effort: pl.effort, Worktree: pl.worktree, Settings: settings, Prompt: text, Known: known,
	})
	if err != nil {
		return "", err
	}
	_, err = state.Commit(l.dir, l.now, state.LockWait, state.RecordLaunch(project, record, state.Launch{
		Kind: req.Kind, Replaces: req.Replaces, Cause: req.Cause, Counted: req.Counted,
		Session: state.Session{
			Name: pl.name, Kind: pl.sessionKind, ShortID: got.ShortID, SessionID: got.SessionID,
			Worktree: pl.worktree, Branch: pl.branch,
		},
	}), snapshot.Render)
	if err != nil {
		return "", errors.Join(err, l.runner.Stop(ctx, got.ShortID))
	}
	return got.ShortID, nil
}
