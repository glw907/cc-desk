package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glw907/cc-desk/internal/atomicfile"
	"github.com/glw907/cc-desk/internal/snapshot"
	"github.com/glw907/cc-desk/internal/state"
	"github.com/glw907/cc-desk/internal/usage"
	"github.com/spf13/cobra"
)

// version is the release this build reports.
const version = "0.1.0"

// rootFlags holds the persistent flags every verb reads.
type rootFlags struct {
	stateDir         string
	now              string
	usageCommand     string
	systemctlCommand string
	claudeCommand    string
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

// commit runs one event through state.Commit and returns the snapshot's
// bytes. Every verb takes a usage reading first; see commitReading.
func (f *rootFlags) commit(ev state.Event) ([]byte, error) {
	snap, _, err := f.commitReading(ev, false, "")
	return snap, err
}

// commitReading runs ev on the state folder and the clock the flags name.
// It takes the usage reading before the lock, folds it in inside the lock,
// and returns the snapshot's bytes and the reading the state kept. force
// takes a fresh reading even when the cached one is fresh. A non-empty
// rawFile receives the raw bytes the usage command printed.
func (f *rootFlags) commitReading(ev state.Event, force bool, rawFile string) ([]byte, state.Usage, error) {
	dir, err := f.dir()
	if err != nil {
		return nil, state.Usage{}, err
	}
	now, err := f.clock()
	if err != nil {
		return nil, state.Usage{}, err
	}
	u, raw, took := usage.Take(context.Background(), dir, f.usageCommand, now, force)
	if rawFile != "" && raw != nil {
		if err := atomicfile.WriteFile(rawFile, raw, 0o644); err != nil {
			return nil, state.Usage{}, fmt.Errorf("save the raw result: %w", err)
		}
	}
	var kept state.Usage
	in := state.WithReading(u, force && took, ev)
	snap, err := state.Commit(dir, now, state.LockWait, func(s *state.State, now time.Time) error {
		err := in(s, now)
		if err == nil && s.Usage != nil {
			kept = *s.Usage
		}
		return err
	}, snapshot.Render)
	return snap, kept, err
}

// group makes a command that only holds subcommands. Without a RunE, cobra
// prints help and exits 0 on an unknown child.
func group(cmd *cobra.Command) *cobra.Command {
	cmd.Args = cobra.NoArgs
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("%s needs a subcommand; run it with --help", cmd.CommandPath())
	}
	return cmd
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
	cmd.PersistentFlags().StringVar(&f.usageCommand, "usage-command", "", "the command that prints the usage result")
	_ = cmd.PersistentFlags().MarkHidden("usage-command")
	cmd.PersistentFlags().StringVar(&f.systemctlCommand, "systemctl-command", "", "the command that stands in for systemctl")
	_ = cmd.PersistentFlags().MarkHidden("systemctl-command")
	cmd.PersistentFlags().StringVar(&f.claudeCommand, "claude-command", "claude", "the command that stands in for claude")
	_ = cmd.PersistentFlags().MarkHidden("claude-command")
	cmd.AddCommand(newTimerCmd(&f), newReleaseCmd(&f), newRoundCmd(&f), newPassCmd(&f), newSnapshotCmd(&f), newUsageCmd(&f), newFreeCmd(&f), newRulingCmd(&f), newCheckCmd(&f), newSessionCmd(&f), newPromptCmd(&f), newFollowupCmd(&f), newRulesCmd(&f), newDeskCmd(&f), newQuestionCmd(&f))
	return cmd
}
