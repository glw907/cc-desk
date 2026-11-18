package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

// kick starts the check's service without waiting, so work a verb just made
// reaches the check at once instead of at the next timer run. It acts only
// on the workstation's own state folder or when --systemctl-command names a
// stand-in, so no run on a temporary folder can start the real service. A
// failed kick prints one line and leaves the verb's exit alone.
func (f *rootFlags) kick(cmd *cobra.Command) {
	if !f.kickable() {
		return
	}
	if _, err := f.systemctl("start", "--no-block", checkService); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "kick: %v\n", err)
	}
}

// kickable reports whether the state folder is the live one, both sides made
// absolute with symlinks resolved, or a stand-in is set.
func (f *rootFlags) kickable() bool {
	if f.systemctlCommand != "" {
		return true
	}
	dir, err := f.dir()
	if err != nil {
		return false
	}
	home, err := homeDir()
	if err != nil {
		return false
	}
	return resolved(dir) == resolved(filepath.Join(home, ".local", "state", "cc-desk"))
}

// resolved makes p absolute and follows its symlinks, falling back to the
// absolute path when p does not exist.
func resolved(p string) string {
	if abs, err := filepath.Abs(p); err == nil {
		p = abs
	}
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	return p
}
