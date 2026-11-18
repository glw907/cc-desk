package main

import (
	"fmt"
	"os/exec"
	"slices"

	"github.com/glw907/cc-desk/internal/desk"
	"github.com/glw907/cc-desk/internal/sessions"
	"github.com/glw907/cc-desk/internal/state"
	"github.com/spf13/cobra"
)

const registerPrefix = "cannot register the desk: "

func newDeskCmd(f *rootFlags) *cobra.Command {
	cmd := group(&cobra.Command{
		Use:   "desk",
		Short: "Register a desk session",
	})
	register := &cobra.Command{
		Use:   "register",
		Short: "Register this terminal session as a desk",
		Long: "Find this session's row in `claude agents --json` and record its process as a\n" +
			"desk. A session whose id is empty or has no row, a row that is not interactive,\n" +
			"a row whose process is not above this command, and a process with no terminal\n" +
			"are each refused, with the failing test named.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := desk.Caller()
			if c.SessionID == "" {
				return state.Refuse(registerPrefix + "CLAUDE_CODE_SESSION_ID is empty")
			}
			rows, err := f.agentRows()
			if err != nil {
				return err
			}
			i := slices.IndexFunc(rows, func(r sessions.Row) bool { return r.SessionID == c.SessionID })
			if i < 0 {
				return state.Refuse(registerPrefix + "claude agents lists no row for this session")
			}
			row := rows[i]
			if row.Kind != "interactive" {
				return state.Refuse(fmt.Sprintf("%sthe session's kind is %q, not interactive", registerPrefix, row.Kind))
			}
			j := slices.IndexFunc(c.Chain, func(p state.Proc) bool { return p.PID == row.PID })
			if j < 0 {
				return state.Refuse(registerPrefix + "the session's process isn't above this command")
			}
			if c.Chain[j].TTY == 0 {
				return state.Refuse(registerPrefix + "the session's process has no terminal")
			}
			entry, err := desk.Entry(c.SessionID, row.PID)
			if err != nil {
				return err
			}
			if _, err := f.commit(state.RegisterDesk(entry)); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "registered desk %s\n", c.SessionID)
			return nil
		},
	}
	cmd.AddCommand(register)
	return cmd
}

// agentRows runs the claude command the flags name and decodes its list.
func (f *rootFlags) agentRows() ([]sessions.Row, error) {
	out, err := exec.Command(f.claudeCommand, "agents", "--json").Output()
	if err != nil {
		return nil, fmt.Errorf("run %s agents --json: %w", f.claudeCommand, err)
	}
	rows, err := sessions.ParseList(out)
	if err != nil {
		return nil, fmt.Errorf("read the output of %s agents --json: %w", f.claudeCommand, err)
	}
	return rows, nil
}
