package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/glw907/cc-desk/internal/snapshot"
	"github.com/glw907/cc-desk/internal/state"
	"github.com/spf13/cobra"
)

func newReportCmd(f *rootFlags) *cobra.Command {
	cmd := group(&cobra.Command{
		Use:   "report",
		Short: "Record a report and mark it read",
	})

	var round string
	add := &cobra.Command{
		Use:   "add <project> <file> [--round <round>]",
		Short: "Record the report in <file> for a round, or for the release without --round",
		Long: "Record one report from a file. The file holds a title, a summary, the calls the\n" +
			"work made and the misses the plan had. All four keys are required, and a list may\n" +
			"be empty. A round or release takes one report. A done release takes one too.",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, file := args[0], args[1]
			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("read report: %w", err)
			}
			rf, err := state.DecodeReport(data)
			if err != nil {
				return fmt.Errorf("%s: %w", file, err)
			}
			var id string
			if _, err := f.commit(state.AddReport(project, round, rf, &id)); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "recorded report %s\n", id)
			return nil
		},
	}
	add.Flags().StringVar(&round, "round", "", "the round the report is for; omit it for the release's report")
	cmd.AddCommand(add)

	cmd.AddCommand(&cobra.Command{
		Use:   "read <id>",
		Short: "Mark a report read",
		Long: "Mark a report read now. A report already read keeps its first time. This takes\n" +
			"no usage reading and starts no waiting pass.",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return err
			}
			if id := args[0]; id == "" || strings.HasPrefix(id, "-") || strings.Contains(id, "/") {
				return fmt.Errorf("report id %q is not an id", id)
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
			if _, err := state.Commit(dir, now, state.LockWait, state.ReadReport(args[0]), snapshot.Render); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "marked report %s read\n", args[0])
			return nil
		},
	})
	return cmd
}
