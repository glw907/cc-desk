package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/glw907/cc-desk/internal/sessions"
	"github.com/glw907/cc-desk/internal/state"
	"github.com/spf13/cobra"
)

func newReleaseCmd(f *rootFlags) *cobra.Command {
	cmd := group(&cobra.Command{
		Use:   "release",
		Short: "Add a release, record its checkout, or end it",
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "add <file>",
		Short: "Validate the release in <file> and install it as releases/<project>.json",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read release: %w", err)
			}
			r, err := state.DecodeRelease(data)
			if err != nil {
				return fmt.Errorf("%s: %w", args[0], err)
			}
			if _, err := f.commit(state.AddRelease(r)); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "added release %s of %s\n", r.Name, r.Project)
			return nil
		},
	})
	cmd.AddCommand(newReleaseDoneCmd(f))
	var branch string
	checkout := &cobra.Command{
		Use:   "checkout <project> <path>",
		Short: "Record the repo's main checkout and default branch on the project's release",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := f.commit(state.SetCheckout(args[0], args[1], branch)); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "recorded the checkout of %s\n", args[0])
			return nil
		},
	}
	checkout.Flags().StringVar(&branch, "default-branch", "", "the repo's default branch")
	_ = checkout.MarkFlagRequired("default-branch")
	cmd.AddCommand(checkout)
	return cmd
}

func newReleaseDoneCmd(f *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "done <project>",
		Short: "Move the project's live release to releases/done/",
		Long: "End the project's live release: its file moves to releases/done/ and is marked done,\n" +
			"so the project can take a new release. The command refuses when the project has no live\n" +
			"release, or when a session one of the release's records holds is still live. It takes\n" +
			"check.lock first, so no check run can launch into the release as it moves. It prints\n" +
			"the new file's path.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project := args[0]
			dir, err := f.dir()
			if err != nil {
				return err
			}
			if dir, err = filepath.Abs(dir); err != nil {
				return fmt.Errorf("resolve state folder: %w", err)
			}
			unlock, err := state.LockFile(dir, state.CheckLockFile, state.LockWait)
			if err != nil {
				return err
			}
			defer unlock()
			s, err := state.Load(dir)
			if err != nil {
				return err
			}
			var rf *state.ReleaseFile
			for _, o := range s.Releases {
				if !o.Done && o.Project == project {
					rf = o
				}
			}
			if rf == nil {
				return state.Refuse(fmt.Sprintf("project %q has no live release", project))
			}
			rows, err := sessions.List(context.Background(), f.claudeCommand)
			if err != nil {
				return fmt.Errorf("read the session list: %w", err)
			}
			var held string
			s.LiveSessions(func(o *state.ReleaseFile, record string, sess *state.Session) {
				if row, ok := sessions.Find(rows, sess.ShortID); o == rf && ok && holdsRelease(row) && held == "" {
					held = fmt.Sprintf("%s/%s holds a live session, %s", project, record, sess.Name)
				}
			})
			if held != "" {
				return state.Refuse(held)
			}
			var to string
			if _, err := f.commit(state.ReleaseDone(project, &to)); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), filepath.Join(dir, to))
			return nil
		},
	}
}

// holdsRelease reports whether a listed row keeps its release live. Only a
// stopped, done, or failed row lets the release end: a state Claude Code adds
// later, such as one for a session still starting, holds it.
func holdsRelease(row sessions.Row) bool {
	switch row.State {
	case "stopped", "done", "failed":
		return false
	}
	return true
}
