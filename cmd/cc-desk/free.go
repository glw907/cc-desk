package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/glw907/cc-desk/internal/state"
	"github.com/spf13/cobra"
)

func newFreeCmd(f *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "free",
		Short: "List which passes are free to start",
		Long: "Print one line for each live pass that has not started or is waiting:\n" +
			"free <project>/<pass>, waiting <project>/<pass> on <limits>, or\n" +
			"not free <project>/<pass>: <rules it breaks>.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var lines []string
			_, err := f.commit(func(s *state.State, _ time.Time) error {
				lines = freeLines(s)
				return nil
			})
			if err != nil {
				return err
			}
			for _, l := range lines {
				fmt.Fprintln(cmd.OutOrStdout(), l)
			}
			return nil
		},
	}
}

// freeLines describes each live not-started or waiting pass, in release
// order then pass order.
func freeLines(s *state.State) []string {
	var lines []string
	for _, rf := range s.Releases {
		if rf.Done {
			continue
		}
		for i := range rf.Passes {
			p := &rf.Passes[i]
			name := rf.Project + "/" + p.Name
			switch p.State {
			case state.Waiting:
				lines = append(lines, "waiting "+name+" on "+strings.Join(p.Wait.Limits, ", "))
			case state.NotStarted:
				broken := s.Breaks(rf, p, "")
				if len(broken) == 0 {
					lines = append(lines, "free "+name)
					continue
				}
				names := make([]string, len(broken))
				for j, b := range broken {
					names[j] = b.Name
				}
				lines = append(lines, "not free "+name+": "+strings.Join(names, ", "))
			}
		}
	}
	return lines
}

// heldOn returns the limits a waiting pass is held on, or nil when the
// pass is not waiting.
func heldOn(s *state.State, project, pass string) []string {
	for _, rf := range s.Releases {
		if rf.Project != project {
			continue
		}
		for i := range rf.Passes {
			if p := &rf.Passes[i]; p.Name == pass && p.State == state.Waiting {
				return p.Wait.Limits
			}
		}
	}
	return nil
}
