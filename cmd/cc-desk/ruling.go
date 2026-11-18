package main

import (
	"fmt"
	"time"

	"github.com/glw907/cc-desk/internal/desk"
	"github.com/glw907/cc-desk/internal/state"
	"github.com/spf13/cobra"
)

func newRulingCmd(f *rootFlags) *cobra.Command {
	cmd := group(&cobra.Command{
		Use:   "ruling",
		Short: "Record Geoff's rulings",
	})

	var req state.RulingRequest
	record := &cobra.Command{
		Use:   "record <words> [--project P --round R --question Q --option O] [--replaces <id>]",
		Short: "Record a ruling in Geoff's exact words",
		Long: "Append one ruling to rulings.jsonl, from a desk session only. The words are one\n" +
			"argument, taken byte for byte; words that start with - go after --. With --project,\n" +
			"--round, --question, and --option together, the ruling answers that question, closes\n" +
			"any guess on it, and frees what waited. With --replaces as well, it replaces the\n" +
			"question's current answer while no pass resting on the question has started. A session that is not a registered desk, a\n" +
			"command that does not run under the desk's process, and a desk process with no\n" +
			"terminal are each refused, with the failing test named.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			req.Words = args[0]
			// The caller is read before the lock, so the lock wait never
			// shifts what the chain shows.
			ev := state.RecordRuling(desk.Caller(), req)
			var id string
			_, err := f.commit(func(s *state.State, now time.Time) error {
				if err := ev(s, now); err != nil {
					return err
				}
				id = s.Rulings[len(s.Rulings)-1].ID
				return nil
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "recorded ruling %s\n", id)
			if req.Question != "" {
				f.kick(cmd)
			}
			return nil
		},
	}
	fl := record.Flags()
	fl.StringVar(&req.Project, "project", "", "the project of the question this ruling answers")
	fl.StringVar(&req.Round, "round", "", "the round of the question")
	fl.StringVar(&req.Question, "question", "", "the question's id")
	fl.StringVar(&req.Option, "option", "", "the option Geoff chose")
	fl.StringVar(&req.Replaces, "replaces", "", "the id of the ruling this one replaces as the question's answer")
	cmd.AddCommand(record)
	return cmd
}
