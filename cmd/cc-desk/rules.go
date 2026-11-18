package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/glw907/cc-desk/internal/desk"
	"github.com/glw907/cc-desk/internal/state"
)

func newRulesCmd(f *rootFlags) *cobra.Command {
	cmd := group(&cobra.Command{
		Use:   "rules",
		Short: "Change a project's allow rules",
	})
	for _, verb := range []string{"add", "remove"} {
		cmd.AddCommand(newRulesVerbCmd(f, verb))
	}
	return cmd
}

func newRulesVerbCmd(f *rootFlags, verb string) *cobra.Command {
	var lf launchFlags
	cmd := &cobra.Command{
		Use:   verb + " <project> <rule>",
		Short: map[string]string{"add": "Add a rule to a project's allow rules", "remove": "Remove a rule from a project's allow rules"}[verb],
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// The caller is read before the lock, so the lock wait never
			// shifts what the chain shows.
			caller := desk.Caller()
			ev := state.AddRule(caller, args[0], args[1])
			if verb == "remove" {
				ev = state.RemoveRule(caller, args[0], args[1])
			}
			if _, err := f.commit(ev); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "%s %s\n", verb, args[1])
			return nil
		},
	}
	lf.register(cmd)
	return cmd
}
