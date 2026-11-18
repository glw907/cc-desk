package main

import (
	"fmt"
	"os"

	"github.com/glw907/cc-desk/internal/state"
	"github.com/spf13/cobra"
)

func newReleaseCmd(f *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release",
		Short: "Add a release",
	}
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
	return cmd
}
