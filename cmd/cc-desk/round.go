package main

import (
	"fmt"
	"os"

	"github.com/glw907/cc-desk/internal/state"
	"github.com/spf13/cobra"
)

func newRoundCmd(f *rootFlags) *cobra.Command {
	cmd := group(&cobra.Command{
		Use:   "round",
		Short: "Record a round's design",
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "design <project> <round> <file>",
		Short: "Add the passes and questions in <file> to the round and mark its design written",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, round, file := args[0], args[1], args[2]
			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("read design: %w", err)
			}
			d, err := state.DecodeDesign(data)
			if err != nil {
				return fmt.Errorf("%s: %w", file, err)
			}
			if _, err := f.commit(state.DesignRound(project, round, d)); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "recorded the design of %s/%s\n", project, round)
			return nil
		},
	})
	return cmd
}
