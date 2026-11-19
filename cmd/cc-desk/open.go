package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"github.com/glw907/cc-desk/internal/brief"
	"github.com/glw907/cc-desk/internal/desk"
	"github.com/glw907/cc-desk/internal/state"
	"github.com/spf13/cobra"
)

const (
	openLine   = "What is waiting for me?"
	openModel  = "claude-opus-5-5"
	openEffort = "high"
)

// sessionEnv names the variables that make a new session count as a child of
// the one that started it, so the session list leaves it out.
var sessionEnv = []string{"CLAUDECODE", "CLAUDE_CODE_CHILD_SESSION", "CLAUDE_CODE_SESSION_ID", "CLAUDE_CODE_ENTRYPOINT"}

func openArgs(cmd *cobra.Command, args []string) error {
	if len(args) != 0 && len(args) != 2 {
		return fmt.Errorf("%s takes no arguments or a project and a round, not %d", cmd.CommandPath(), len(args))
	}
	return nil
}

// liveRelease finds the live release of project and checks its round, naming
// the names it knows when either is wrong.
func liveRelease(s *state.State, project, round string) (*state.ReleaseFile, error) {
	var names []string
	for _, rf := range s.Releases {
		if rf.Done {
			continue
		}
		if rf.Project == project {
			rounds := make([]string, len(rf.Rounds))
			for i, r := range rf.Rounds {
				rounds[i] = r.Name
			}
			if !slices.Contains(rounds, round) {
				return nil, state.Refuse(fmt.Sprintf("project %q has no round %q; its rounds are %s", project, round, strings.Join(rounds, ", ")))
			}
			return rf, nil
		}
		names = append(names, rf.Project)
	}
	return nil, state.Refuse(fmt.Sprintf("no project %q; the projects are %s", project, strings.Join(names, ", ")))
}

// runOpen opens the desk: it builds the brief, registers a session under a
// new id, writes the session's settings, and replaces this process with
// claude, so the registered process is the listed one.
func runOpen(cmd *cobra.Command, f *rootFlags, args []string) error {
	if !desk.HasTerminal() {
		return cmd.Help()
	}
	dir, err := f.dir()
	if err != nil {
		return err
	}
	if dir, err = filepath.Abs(dir); err != nil {
		return fmt.Errorf("find state folder: %w", err)
	}
	now, err := f.clock()
	if err != nil {
		return err
	}
	claude, err := exec.LookPath(f.claudeCommand)
	if err != nil {
		return fmt.Errorf("find %s: %w", f.claudeCommand, err)
	}
	binary, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find this command: %w", err)
	}

	unlock, err := state.Lock(dir, state.LockWait)
	if err != nil {
		return err
	}
	s, err := state.Load(dir)
	if err != nil {
		unlock()
		return state.Refuse("cannot open the desk: " + err.Error())
	}
	var text string
	var rf *state.ReleaseFile
	if len(args) == 2 {
		if rf, err = liveRelease(s, args[0], args[1]); err == nil {
			text, err = brief.BuildRound(s, now, args[0], args[1])
		}
	} else {
		rf = brief.First(s)
		text = brief.Build(s, now)
	}
	unlock()
	if err != nil {
		return err
	}

	prefix := binary
	if f.stateDir != "" {
		prefix += " --state-dir " + dir
	}
	wordsDir := filepath.Join(dir, "desk", "words")
	if err := os.MkdirAll(wordsDir, 0o755); err != nil {
		return fmt.Errorf("make the words folder: %w", err)
	}
	skillFile := ""
	if f.skillFile != "" {
		if skillFile, err = filepath.Abs(f.skillFile); err != nil {
			return fmt.Errorf("find the skill file: %w", err)
		}
	}
	id, err := desk.NewID()
	if err != nil {
		return err
	}
	settings, err := desk.WriteSettings(dir, id, prefix, wordsDir, skillFile)
	if err != nil {
		return err
	}
	entry, err := desk.Entry(id, os.Getpid())
	if err != nil {
		return err
	}
	if _, err := f.commit(state.RegisterDesk(entry)); err != nil {
		return err
	}

	name := "desk"
	folder := ""
	if rf != nil {
		name = rf.Project + "-desk"
		folder = rf.Checkout
	}
	if folder != "" {
		if err := os.Chdir(folder); err != nil {
			return fmt.Errorf("open the project's checkout: %w", err)
		}
	}
	argv := []string{claude, "--session-id", id, "--model", openModel, "--effort", openEffort,
		"--permission-mode", "default", "--settings", settings,
		"--append-system-prompt", systemPrompt(prefix, wordsDir, skillFile, text),
		"--name", name, openLine}
	return syscall.Exec(claude, argv, scrubbed(os.Environ()))
}

// scrubbed drops the session variables from env.
func scrubbed(env []string) []string {
	return slices.DeleteFunc(slices.Clone(env), func(kv string) bool {
		k, _, _ := strings.Cut(kv, "=")
		return slices.Contains(sessionEnv, k)
	})
}

// systemPrompt tells the session what it is, where its skill is, the exact
// command to type for each verb, and gives it the brief.
func systemPrompt(prefix, wordsDir, skillFile, text string) string {
	load := "Load the cc-desk skill with the Skill tool"
	if skillFile != "" {
		load = "Read the file at " + skillFile + " with the Read tool"
	}
	return "You are the desk session of cc-desk, the one place where Geoff answers what waits for him.\n" +
		load + " before your first reply, and follow it.\n\n" +
		"Type each cc-desk verb as this command, never as a bare cc-desk: " + prefix + "\n" +
		"For a ruling, write his words with the Write tool to a file in " + wordsDir + ", named <number>.txt, " +
		"then record them with --words-file and that path.\n\n" +
		"The brief follows. Geoff does not see it.\n\n" + text
}
