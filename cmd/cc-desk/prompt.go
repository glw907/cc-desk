package main

import (
	"io"

	"github.com/spf13/cobra"
)

func newPromptCmd(f *rootFlags) *cobra.Command {
	cmd := group(&cobra.Command{
		Use:   "prompt",
		Short: "Handle the permission prompts a session raises",
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "hook",
		Short: "Answer a PermissionRequest hook (a stub until the prompt pass)",
		Long: "The settings file of every launched session names this verb as its PermissionRequest\n" +
			"hook. It reads the hook's input, prints nothing, and exits 0, so the session's own\n" +
			"prompt stands. A later pass records the prompt.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := io.Copy(io.Discard, cmd.InOrStdin())
			return err
		},
	})
	return cmd
}
