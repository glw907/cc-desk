package main

import (
	"fmt"

	"github.com/glw907/cc-desk/internal/state"
	"github.com/spf13/cobra"
)

func newUsageCmd(f *rootFlags) *cobra.Command {
	var raw string
	cmd := &cobra.Command{
		Use:   "usage [--raw FILE]",
		Short: "Print the weekly usage figure",
		Long: "Print the weekly usage figure and when it resets. The reading is kept for 5 minutes.\n" +
			"With --raw, take a fresh reading and save the command's raw result to FILE.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, u, err := f.commitReading(nil, raw != "", raw)
			if err != nil {
				return err
			}
			if u.Error != "" {
				return fmt.Errorf("%s: %s", state.UsageUnreadable, u.Error)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%d%% used, resets %s\n", u.Percent, u.ResetText)
			return nil
		},
	}
	cmd.Flags().StringVar(&raw, "raw", "", "take a fresh reading and save the raw result to this file")
	return cmd
}
