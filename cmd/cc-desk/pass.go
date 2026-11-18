package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/glw907/cc-desk/internal/state"
	"github.com/spf13/cobra"
)

func newPassCmd(f *rootFlags) *cobra.Command {
	cmd := group(&cobra.Command{
		Use:   "pass",
		Short: "Record a pass's start, finish, or stop",
	})

	var guess string
	start := &cobra.Command{
		Use:   "start <project> <pass> [--guess QUESTION]",
		Short: "Start a pass, or let it wait, or resume a stopped one",
		Long: "Start a pass that is free to start. A free pass that a limit holds is recorded as\n" +
			"waiting, and starts when the limit lifts. A pass that breaks a rule is refused with\n" +
			"every broken rule named. With --guess, start a pass that rests on the unanswered\n" +
			"question, recording the question's recommended option as the guess.",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ev := state.StartPass(args[0], args[1])
			if guess != "" {
				ev = state.StartOnGuess(args[0], args[1], guess)
			}
			var held []string
			_, err := f.commit(func(s *state.State, now time.Time) error {
				if err := ev(s, now); err != nil {
					return err
				}
				held = heldOn(s, args[0], args[1])
				return nil
			})
			if err != nil {
				return err
			}
			if len(held) > 0 {
				fmt.Fprintf(cmd.ErrOrStderr(), "waiting %s/%s on %s\n", args[0], args[1], strings.Join(held, ", "))
				return nil
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "started %s/%s\n", args[0], args[1])
			return nil
		},
	}
	start.Flags().StringVar(&guess, "guess", "", "start on a guess at this unanswered question")
	cmd.AddCommand(start)

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
		Short: "Record the stop of a pass with one of the stop reasons",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := f.commit(state.StopPass(args[0], args[1], state.StopReason(reason))); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "stopped %s/%s: %s\n", args[0], args[1], reason)
			return nil
		},
	}
	stop.Flags().StringVar(&reason, "reason", "", "why the pass stopped: usage, irreversible, secret, auth_data, scope, permission_prompt, sign_in_expired, unreadable_stop, relaunch_limit, or launch_failed")
	_ = stop.MarkFlagRequired("reason")
	cmd.AddCommand(stop)
	return cmd
}
