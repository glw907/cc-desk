package main

import (
	"fmt"
	"os"

	"github.com/glw907/cc-desk/internal/state"
	"github.com/spf13/cobra"
)

func newQuestionCmd(f *rootFlags) *cobra.Command {
	cmd := group(&cobra.Command{
		Use:   "question",
		Short: "Add a question to a round",
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "add <project> <round> <file>",
		Short: "Add the question in <file> to the round and to the not-started passes it names",
		Long: "Add one question from a file to a round. The file holds the question's id, text,\n" +
			"options, recommended option, background paragraphs, what each option causes, the\n" +
			"reason for the recommendation, whether it is never guessed, and the not-started\n" +
			"passes of the round that rest on it. A part that is missing is named in the refusal.",
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, round, file := args[0], args[1], args[2]
			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("read question: %w", err)
			}
			q, err := state.DecodeQuestion(data)
			if err != nil {
				return fmt.Errorf("%s: %w", file, err)
			}
			if _, err := f.commit(state.AddQuestion(project, round, q)); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "added question %s to %s/%s\n", q.ID, project, round)
			return nil
		},
	})
	return cmd
}
