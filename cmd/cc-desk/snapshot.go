package main

import (
	"github.com/spf13/cobra"
)

func newSnapshotCmd(f *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "snapshot",
		Short: "Print the snapshot as JSON",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			snap, err := f.commit(nil)
			if err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write(snap)
			return err
		},
	}
}
