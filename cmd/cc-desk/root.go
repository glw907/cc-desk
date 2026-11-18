package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// version is the release this build reports.
const version = "0.1.0"

// rootFlags holds the persistent flags every verb reads.
type rootFlags struct {
	stateDir string
	now      string
}

// dir returns the state folder: the flag, else $XDG_STATE_HOME/cc-desk,
// else ~/.local/state/cc-desk.
func (f *rootFlags) dir() (string, error) {
	if f.stateDir != "" {
		return f.stateDir, nil
	}
	if x := os.Getenv("XDG_STATE_HOME"); x != "" {
		return filepath.Join(x, "cc-desk"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find home folder: %w", err)
	}
	return filepath.Join(home, ".local", "state", "cc-desk"), nil
}

// clock returns the time every verb uses: the hidden --now flag, else the
// system time.
func (f *rootFlags) clock() (time.Time, error) {
	if f.now == "" {
		return time.Now().UTC(), nil
	}
	t, err := time.Parse(time.RFC3339, f.now)
	if err != nil {
		return time.Time{}, fmt.Errorf("--now %q is not an RFC 3339 time", f.now)
	}
	return t.UTC(), nil
}

func newRootCmd() *cobra.Command {
	var f rootFlags
	cmd := &cobra.Command{
		Use:           "cc-desk",
		Short:         "Keep the state of Claude Code's work and print its snapshot",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.PersistentFlags().StringVar(&f.stateDir, "state-dir", "",
		"state folder (default $XDG_STATE_HOME/cc-desk, else ~/.local/state/cc-desk)")
	cmd.PersistentFlags().StringVar(&f.now, "now", "", "the time to use, in RFC 3339")
	_ = cmd.PersistentFlags().MarkHidden("now")
	return cmd
}
