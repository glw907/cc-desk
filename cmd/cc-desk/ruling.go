package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
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
	var wordsFile string
	record := &cobra.Command{
		Use:   "record (<words> | --words-file <path>) [--project P --round R --question Q --option O] [--replaces <id>]",
		Short: "Record a ruling in Geoff's exact words",
		Long: "Append one ruling to rulings.jsonl, from a desk session only. The words are one\n" +
			"argument, taken byte for byte; words that start with - go after --. Or they are the\n" +
			"whole of the file --words-file names, less its trailing newlines. With --project,\n" +
			"--round, --question, and --option together, the ruling answers that question, closes\n" +
			"any guess on it, and frees what waited. With --replaces as well, it replaces the\n" +
			"question's current answer until a pass resting on it has finished, or, for a merge\n" +
			"question, until its merge is called or its answer is carried out. A session that is\n" +
			"not a registered desk, a\n" +
			"command that does not run under the desk's process, and a desk process with no\n" +
			"terminal are each refused, with the failing test named.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if (len(args) == 1) == (wordsFile != "") {
				return errors.New("give the words as one argument or with --words-file, not both or neither")
			}
			if wordsFile != "" {
				data, err := os.ReadFile(wordsFile)
				if err != nil {
					return fmt.Errorf("read the words file: %w", err)
				}
				req.Words = strings.TrimRight(string(data), "\n")
			} else {
				req.Words = args[0]
			}
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
	fl.StringVar(&wordsFile, "words-file", "", "a file that holds the words, in place of the argument")
	fl.StringVar(&req.Project, "project", "", "the project of the question this ruling answers")
	fl.StringVar(&req.Round, "round", "", "the round of the question")
	fl.StringVar(&req.Question, "question", "", "the question's id")
	fl.StringVar(&req.Option, "option", "", "the option Geoff chose")
	fl.StringVar(&req.Replaces, "replaces", "", "the id of the ruling this one replaces as the question's answer")
	cmd.AddCommand(record)
	return cmd
}
