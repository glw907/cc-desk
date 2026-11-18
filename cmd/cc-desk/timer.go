package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/glw907/cc-desk/internal/atomicfile"
	"github.com/glw907/cc-desk/internal/state"
	"github.com/glw907/cc-desk/systemd"
	"github.com/spf13/cobra"
)

const (
	checkService = "cc-desk-check.service"
	checkTimer   = "cc-desk-check.timer"
)

// systemctl runs the systemctl command the flags name against the user
// manager and returns what it printed. A failure carries the command's own
// message.
func (f *rootFlags) systemctl(args ...string) (string, error) {
	name := f.systemctlCommand
	if name == "" {
		name = "systemctl"
	}
	out, err := exec.Command(name, append([]string{"--user"}, args...)...).CombinedOutput()
	text := strings.TrimSpace(string(out))
	if err != nil {
		if text == "" {
			text = err.Error()
		}
		return text, fmt.Errorf("systemctl %s: %s", strings.Join(args, " "), text)
	}
	return text, nil
}

// homeDir is $HOME. The units name %h, so the install never follows the XDG
// variables.
func homeDir() (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		return "", fmt.Errorf("HOME is empty")
	}
	return home, nil
}

func newTimerCmd(f *rootFlags) *cobra.Command {
	cmd := group(&cobra.Command{
		Use:   "timer",
		Short: "Install or read the timer that runs the check",
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "install",
		Short: "Copy this binary and the two units into place and enable the timer",
		Long: "Copy this binary to ~/.local/share/cc-desk/bin/cc-desk, write the check's service and\n" +
			"timer to ~/.config/systemd/user/, reload the user manager, and enable the timer.\n" +
			"Paths come from $HOME, never the XDG variables, because the units name %h.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := homeDir()
			if err != nil {
				return err
			}
			self, err := os.Executable()
			if err != nil {
				return fmt.Errorf("find this binary: %w", err)
			}
			bin, err := os.ReadFile(self)
			if err != nil {
				return fmt.Errorf("read this binary: %w", err)
			}
			binPath := filepath.Join(home, ".local", "share", "cc-desk", "bin", "cc-desk")
			unitDir := filepath.Join(home, ".config", "systemd", "user")
			for _, dir := range []string{filepath.Dir(binPath), unitDir} {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					return fmt.Errorf("create %s: %w", dir, err)
				}
			}
			if err := atomicfile.WriteFile(binPath, bin, 0o755); err != nil {
				return fmt.Errorf("write %s: %w", binPath, err)
			}
			for name, data := range map[string][]byte{checkService: systemd.Service, checkTimer: systemd.Timer} {
				if err := atomicfile.WriteFile(filepath.Join(unitDir, name), data, 0o644); err != nil {
					return fmt.Errorf("write %s: %w", name, err)
				}
			}
			if _, err := f.systemctl("daemon-reload"); err != nil {
				return err
			}
			if _, err := f.systemctl("enable", "--now", checkTimer); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "installed %s and the check's units; the timer is enabled\n", binPath)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Print when the check last ran and what the timer and service report",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := f.dir()
			if err != nil {
				return err
			}
			now, err := f.clock()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			s, err := state.Load(dir)
			if err != nil {
				return err
			}
			if s.Check.LastRun == nil {
				fmt.Fprintln(out, "last_run: no run recorded")
			} else {
				ago := int(now.Sub(*s.Check.LastRun) / time.Minute)
				fmt.Fprintf(out, "last_run: %s (%d minutes ago)\n", s.Check.LastRun.UTC().Format(time.RFC3339), ago)
			}
			for _, q := range []struct {
				unit  string
				props []string
			}{
				{checkTimer, []string{"ActiveState", "LastTriggerUSec", "NextElapseUSecRealtime"}},
				{checkService, []string{"Result", "ExecMainStatus"}},
			} {
				args := []string{"show", "--timestamp=utc"}
				for _, p := range q.props {
					args = append(args, "-p", p)
				}
				text, err := f.systemctl(append(args, q.unit)...)
				if err != nil {
					return err
				}
				got := map[string]string{}
				for line := range strings.SplitSeq(text, "\n") {
					if k, v, ok := strings.Cut(line, "="); ok {
						got[k] = v
					}
				}
				for _, p := range q.props {
					fmt.Fprintf(out, "%s: %s\n", p, got[p])
				}
			}
			return nil
		},
	})
	return cmd
}
