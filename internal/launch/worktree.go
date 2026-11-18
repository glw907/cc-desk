// Package launch starts a background Claude Code session in a worktree the
// command made: it prepares the worktree, writes the session's settings and
// prompt, runs the launch, and reads the ids the launch gave. It owns no
// state; the caller records the result.
package launch

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// git runs one git command in dir and gives its trimmed output. An error
// carries the command's own text.
func git(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errOut.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return strings.TrimSpace(out.String()), nil
}

// Worktree is a worktree and branch a launch uses.
type Worktree struct {
	Checkout string
	Default  string
	Path     string
	Branch   string
}

// Make adds the worktree for a first launch. It refuses, before any git
// command, a path that already exists, and refuses a branch that already
// exists. It returns the undo for what it made. The undo removes the
// worktree with no force flag, and deletes the branch only while its tip is
// still the commit the branch began at.
func (w Worktree) Make() (undo func() error, err error) {
	if _, err := os.Lstat(w.Path); err == nil {
		return nil, fmt.Errorf("%s already exists and no record names it", w.Path)
	}
	if _, err := git(w.Checkout, "rev-parse", "--verify", "--quiet", "refs/heads/"+w.Branch); err == nil {
		return nil, fmt.Errorf("branch %s already exists in %s", w.Branch, w.Checkout)
	}
	if _, err := git(w.Checkout, "fetch", "origin", w.Default); err != nil {
		return nil, err
	}
	base := "origin/" + w.Default
	tip, err := git(w.Checkout, "rev-parse", base)
	if err != nil {
		return nil, err
	}
	if _, err := git(w.Checkout, "worktree", "add", "--no-track", w.Path, "-b", w.Branch, base); err != nil {
		// A failed add can still have made the branch; remove only that.
		return nil, errors.Join(err, w.dropBranch(tip))
	}
	return func() error {
		if _, err := git(w.Checkout, "worktree", "remove", w.Path); err != nil {
			return fmt.Errorf("the worktree %s is left in place: %w", w.Path, err)
		}
		return w.dropBranch(tip)
	}, nil
}

// dropBranch deletes the branch when it still points at tip.
func (w Worktree) dropBranch(tip string) error {
	at, err := git(w.Checkout, "rev-parse", "--verify", "--quiet", "refs/heads/"+w.Branch)
	if err != nil || at != tip {
		return nil
	}
	if _, err := git(w.Checkout, "branch", "-D", w.Branch); err != nil {
		return fmt.Errorf("the branch %s is left in place: %w", w.Branch, err)
	}
	return nil
}

// Reuse checks that the recorded worktree still exists and is on the
// recorded branch, for a relaunch or a launch again.
func (w Worktree) Reuse() error {
	info, err := os.Stat(w.Path)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("the recorded worktree %s is missing", w.Path)
	}
	got, err := git(w.Path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return err
	}
	if got != w.Branch {
		return fmt.Errorf("the worktree %s is on %q, not the recorded branch %s", w.Path, got, w.Branch)
	}
	return nil
}
