package main

import (
	"fmt"

	"github.com/glw907/cc-desk/internal/brief"
	"github.com/glw907/cc-desk/internal/state"
	"github.com/spf13/cobra"
)

func newBriefCmd(f *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "brief [<project> <round>]",
		Short: "Print the questions that wait, most held-up time first",
		Long: "Print the brief: the next question forecast, each question that waits with its\n" +
			"background, answers, recommendation, and what waits on it, what could come back\n" +
			"soon, and what the check did since the last desk session. With a project and a\n" +
			"round, print that round's design and its questions only. It reads and writes\n" +
			"nothing, and takes no usage reading.",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 && len(args) != 2 {
				return fmt.Errorf("brief takes no arguments or a project and a round, not %d", len(args))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := f.dir()
			if err != nil {
				return err
			}
			now, err := f.clock()
			if err != nil {
				return err
			}
			unlock, err := state.Lock(dir, state.LockWait)
			if err != nil {
				return err
			}
			defer unlock()
			s, err := state.Load(dir)
			if err != nil {
				return err
			}
			var out string
			if len(args) == 2 {
				out, err = brief.BuildRound(s, now, args[0], args[1])
				if err != nil {
					return err
				}
			} else {
				out = brief.Build(s, now)
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), out)
			return err
		},
	}
}
