package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glw907/cc-desk/internal/launch"
	"github.com/glw907/cc-desk/internal/sessions"
	"github.com/glw907/cc-desk/internal/snapshot"
	"github.com/glw907/cc-desk/internal/state"
	"github.com/glw907/cc-desk/internal/usage"
	"github.com/spf13/cobra"
)

// checkFlags holds the check's own flags. The four hidden ones stand in for
// the real claude command, its jobs folder, the network probe, and the wait
// between the two list reads, so a test never reaches them.
type checkFlags struct {
	launchFlags
	dryRun   bool
	probeURL string
	ghCmd    string
	// callLimit is how long the merge call, the fetch, and the pull each may
	// run.
	callLimit time.Duration
}

func newCheckCmd(f *rootFlags) *cobra.Command {
	var cf checkFlags
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Decide what each session the state files name needs",
		Long: "Read the session list and the state files, decide what each session record needs, and\n" +
			"record the decisions in check.json. Then carry them out in the order end, confirm,\n" +
			"launch: an ended session is stopped, a relaunch waits for the old session to leave\n" +
			"the list, and a launch makes its worktree and starts claude. Each outcome is written\n" +
			"back to check.json. With --dry-run, print the decisions and write or start nothing.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error { return runCheck(cmd, f, &cf) },
	}
	cmd.Flags().BoolVar(&cf.dryRun, "dry-run", false, "print the decisions and write nothing")
	cmd.Flags().StringVar(&cf.probeURL, "probe-url", "https://api.anthropic.com/", "the address the network probe asks")
	_ = cmd.Flags().MarkHidden("probe-url")
	cmd.Flags().StringVar(&cf.ghCmd, "gh-command", "gh", "the gh command the pull request read runs")
	_ = cmd.Flags().MarkHidden("gh-command")
	cmd.Flags().DurationVar(&cf.callLimit, "call-limit", launch.DefaultLimit, "how long the merge call, the fetch, and the pull each may run")
	_ = cmd.Flags().MarkHidden("call-limit")
	cf.register(cmd)
	return cmd
}

// runCheck runs the ten steps of the check, as the design lists them.
func runCheck(cmd *cobra.Command, f *rootFlags, cf *checkFlags) error {
	dir, err := f.dir()
	if err != nil {
		return err
	}
	now, err := f.clock()
	if err != nil {
		return err
	}
	if cf.jobsDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("find home folder: %w", err)
		}
		cf.jobsDir = filepath.Join(home, ".claude", "jobs")
	}
	ctx := context.Background()

	// Step 1: the check lock, held to the end.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create state folder: %w", err)
	}
	unlock, err := state.LockFile(dir, state.CheckLockFile, state.LockWait)
	if err != nil {
		return err
	}
	defer unlock()

	// Step 2: the usage reading, taken before any lock on the state.
	u, _, _ := usage.Take(ctx, dir, f.usageCommand, now, false)

	// Step 3: load the state to learn the records.
	s, err := state.Load(dir)
	if err != nil {
		if cf.dryRun {
			return err
		}
		// The commit path writes the error file as every verb does.
		if _, cerr := state.Commit(dir, now, state.LockWait, nil, snapshot.Render); cerr != nil {
			return cerr
		}
		return err
	}
	s.FoldUsage(u, false, now)

	// Steps 4 to 8: gather the evidence.
	ev, err := gatherEvidence(ctx, s, cf, now)
	if err != nil {
		if !cf.dryRun {
			if _, cerr := state.Commit(dir, now, state.LockWait, state.RecordListFailure(), snapshot.Render); cerr != nil {
				return errors.Join(err, cerr)
			}
		}
		return err
	}

	// Steps 9 and 10: decide, record, and print.
	var results []sessions.Result
	if cf.dryRun {
		if results, err = sessions.Plan(s, ev, now); err != nil {
			return err
		}
	} else {
		record := func(s *state.State, now time.Time) error {
			var err error
			if results, err = sessions.Plan(s, ev, now); err != nil {
				return err
			}
			recs := make([]state.Recorded, len(results))
			for i, r := range results {
				recs[i] = state.Recorded{Decision: r.Decision, LastState: r.Observed.State, Seen: r.Observed.Seen, Working: r.Observed.Working, PRWait: r.Merge.Wait}
				if r.Decision.Decision == sessions.Stop {
					recs[i].Stop = state.StopReason(r.Reason)
				}
			}
			return state.RecordCheck(recs)(s, now)
		}
		if _, err := state.Commit(dir, now, state.LockWait, state.WithReading(u, false, record), snapshot.Render); err != nil {
			return err
		}
	}

	// Carry the decisions out, then write what happened beside each one.
	var carryErr error
	if !cf.dryRun {
		l, err := newLauncher(f, &cf.launchFlags)
		if err != nil {
			return err
		}
		l.gh = cf.ghCmd
		l.callLimit = cf.callLimit
		outcomes, err := l.carryOut(ctx, results)
		if outcomes == nil {
			return err
		}
		carryErr = err
		for i := range results {
			results[i].Decision.Outcome = outcomes[i]
		}
		if err := l.recordOutcomes(outcomes); err != nil {
			return errors.Join(carryErr, err)
		}
	}
	out := cmd.OutOrStdout()
	for _, r := range results {
		line, err := json.Marshal(r.Decision)
		if err != nil {
			return fmt.Errorf("encode decision: %w", err)
		}
		fmt.Fprintf(out, "%s\n", line)
	}
	return carryErr
}

// gatherEvidence reads the session list, and the second list when a record
// has no row, then the job file and screen of each blocked row that names no
// reason, then sends the network probe once when a lost connection could be
// relaunched. A failed list read is the only error.
func gatherEvidence(ctx context.Context, s *state.State, cf *checkFlags, now time.Time) (sessions.Evidence, error) {
	list := func() ([]sessions.Row, error) {
		rows, err := sessions.List(ctx, cf.claude)
		if err != nil {
			return nil, fmt.Errorf("read the session list: %w", err)
		}
		return rows, nil
	}
	rows, err := list()
	if err != nil {
		return sessions.Evidence{}, err
	}
	targets := sessions.Watched(s, now)
	if anyMissing(targets, rows) {
		time.Sleep(cf.secondWait)
		if rows, err = list(); err != nil {
			return sessions.Evidence{}, err
		}
	}

	ev := sessions.Evidence{Records: make(map[sessions.Key]sessions.RecordEvidence, len(targets))}
	hold := s.UsageStop()
	probe := false
	for _, t := range targets {
		row, ok := sessions.Find(rows, t.Session.ShortID)
		re := sessions.RecordEvidence{Row: row, HasRow: ok}
		if ok && row.Class() == sessions.ClassBlocked && row.WaitingFor == "" {
			re.Job = sessions.ReadJob(cf.jobsDir, t.Session.ShortID)
			re.Screen = sessions.Screen(ctx, cf.claude, t.Session.ShortID)
			lost := sessions.Classify(row, re.Job, re.Screen) == sessions.LabelLostConnection
			if lost && !hold && t.Session.Counted() < sessions.RelaunchLimit {
				probe = true
			}
		}
		ev.Records[t.Key] = re
	}
	ev.Pulls = readPulls(ctx, s, cf.ghCmd)
	if probe {
		ev.Probe = sessions.ProbeDown
		if sessions.Probe(ctx, cf.probeURL) {
			ev.Probe = sessions.ProbeUp
		}
	}
	return ev, nil
}

// readPulls makes the one pull request read for each pass that is building or
// stopped, has a session record with a branch, and has no recorded merge, on a
// live release with a repo and a checkout.
func readPulls(ctx context.Context, s *state.State, gh string) map[sessions.Key]sessions.PullRead {
	var out map[sessions.Key]sessions.PullRead
	for _, rf := range s.Releases {
		if rf.Done || rf.Repo == "" || rf.Checkout == "" {
			continue
		}
		for i := range rf.Passes {
			p := &rf.Passes[i]
			if p.State != state.Building && p.State != state.Stopped {
				continue
			}
			if p.Session == nil || p.Session.Branch == "" || (p.PullRequest != nil && p.PullRequest.Merged != nil) {
				continue
			}
			if out == nil {
				out = map[sessions.Key]sessions.PullRead{}
			}
			out[sessions.Key{Project: rf.Project, Record: "pass:" + p.Name}] = sessions.ReadPulls(ctx, gh, rf.Repo, p.Session.Branch)
		}
	}
	return out
}

// anyMissing reports whether a record that needs a row has none.
func anyMissing(targets []sessions.Target, rows []sessions.Row) bool {
	for _, t := range targets {
		if row, ok := sessions.Find(rows, t.Session.ShortID); !ok || row.Class() == sessions.ClassNone {
			return true
		}
	}
	return false
}
