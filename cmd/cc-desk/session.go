package main

import (
	"context"
	"fmt"
	"slices"

	"github.com/glw907/cc-desk/internal/state"
	"github.com/spf13/cobra"
)

func newSessionCmd(f *rootFlags) *cobra.Command {
	cmd := group(&cobra.Command{
		Use:   "session",
		Short: "Start or finish a conductor or design writer session",
	})
	cmd.AddCommand(newSessionStartCmd(f), newSessionFinishCmd(f))
	return cmd
}

// recordFor gives the record place a kind and its names stand for.
func recordFor(kind string, args []string) (project, record string, err error) {
	switch kind {
	case "conductor":
		return args[0], "conductor", nil
	case "design":
		if len(args) < 2 {
			return "", "", fmt.Errorf("session design needs a project and a round")
		}
		return args[0], "round:" + args[1], nil
	}
	return "", "", fmt.Errorf("unknown session kind %q: use conductor or design", kind)
}

// sessionArgs checks the arguments of a verb that names a session: the kind,
// the project, and for a design writer the round.
func sessionArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("%s needs a kind, conductor or design", cmd.CommandPath())
	}
	want := map[string]int{"conductor": 2, "design": 3}[args[0]]
	if want == 0 {
		return fmt.Errorf("unknown session kind %q: use conductor or design", args[0])
	}
	return cobra.ExactArgs(want)(cmd, args)
}

func newSessionStartCmd(f *rootFlags) *cobra.Command {
	var lf launchFlags
	cmd := &cobra.Command{
		Use:   "start conductor <project> | design <project> <round>",
		Short: "Launch a conductor or a design writer in a worktree",
		Long: "Make the session's worktree, write its settings and prompt, and start claude in it.\n" +
			"The command refuses when the record already has a live session, or when the project's\n" +
			"release records no checkout. A pass is started by the check, never by this verb.",
		Args: sessionArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			project, record, err := recordFor(args[0], args[1:])
			if err != nil {
				return err
			}
			l, err := newLauncher(f, &lf)
			if err != nil {
				return err
			}
			unlock, err := state.LockFile(l.dir, state.CheckLockFile, state.LockWait)
			if err != nil {
				return err
			}
			defer unlock()
			s, rf, err := l.liveRelease(project)
			if err != nil {
				return err
			}
			if rf.Checkout == "" {
				return state.Refuse(noCheckout(project))
			}
			if record != "conductor" && !slices.ContainsFunc(rf.Rounds, func(r state.Round) bool { return "round:"+r.Name == record }) {
				return fmt.Errorf("no round %q in project %q", args[2], project)
			}
			if old := recordOf(s, project, record); old != nil && old.Ended == nil {
				return state.Refuse(fmt.Sprintf("%s/%s already has a live session, %s", project, record, old.Name))
			}
			id, err := l.Launch(context.Background(), project, record, launchRequest{Kind: state.LaunchFirst})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "launched %s\n", id)
			return nil
		},
	}
	lf.register(cmd)
	return cmd
}

func newSessionFinishCmd(f *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "finish conductor <project> | design <project> <round>",
		Short: "Mark a conductor or design writer finished",
		Long: "Mark the session finished, so the next check ends it and never relaunches it.\n" +
			"A session calls this as its last act.",
		Args: sessionArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			project, record, err := recordFor(args[0], args[1:])
			if err != nil {
				return err
			}
			_, err = f.commit(state.FinishSession(project, record))
			return err
		},
	}
}
