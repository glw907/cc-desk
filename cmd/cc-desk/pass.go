package main

import (
	"fmt"

	"github.com/glw907/cc-desk/internal/state"
	"github.com/spf13/cobra"
)

func newPassCmd(f *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pass",
		Short: "Record a pass's start, finish, or stop",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "start <project> <pass>",
		Short: "Record the start of a pass, or the resume of a stopped one",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := f.commit(state.StartPass(args[0], args[1])); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "started %s/%s\n", args[0], args[1])
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "finish <project> <pass>",
		Short: "Record the finish of a pass and its actual minutes",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := f.commit(state.FinishPass(args[0], args[1])); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "finished %s/%s\n", args[0], args[1])
			return nil
		},
	})

	var reason string
	stop := &cobra.Command{
		Use:   "stop <project> <pass> --reason <reason>",
		Short: "Record the stop of a pass with one of the five reasons",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := f.commit(state.StopPass(args[0], args[1], state.StopReason(reason))); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "stopped %s/%s: %s\n", args[0], args[1], reason)
			return nil
		},
	}
	stop.Flags().StringVar(&reason, "reason", "", "why the pass stopped: usage, irreversible, secret, auth_data, or scope")
	_ = stop.MarkFlagRequired("reason")
	cmd.AddCommand(stop)
	return cmd
}
