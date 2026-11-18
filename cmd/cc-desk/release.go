package main

import (
	"fmt"
	"os"

	"github.com/glw907/cc-desk/internal/state"
	"github.com/spf13/cobra"
)

func newReleaseCmd(f *rootFlags) *cobra.Command {
	cmd := group(&cobra.Command{
		Use:   "release",
		Short: "Add a release or record its checkout",
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
