package e2e

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/glw907/cc-desk/internal/state"
)

// openRig is a state folder with a checkout, and a stand-in claude that
// records what it was launched with.
type openRig struct {
	dir, checkout, stand, out string
}

// scrubbedVars are the variables the open must drop before it starts claude.
var scrubbedVars = []string{"CLAUDECODE", "CLAUDE_CODE_CHILD_SESSION", "CLAUDE_CODE_SESSION_ID", "CLAUDE_CODE_ENTRYPOINT"}

func newOpenRig(t *testing.T) openRig {
	t.Helper()
	for _, v := range scrubbedVars {
		t.Setenv(v, "leaked")
	}
	dir := newStateFrom(t, "desk-open")
	checkout, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	rel := filepath.Join(dir, "releases", "north.json")
	data := strings.ReplaceAll(readFile(t, rel), "/CHECKOUT", checkout)
	if err := os.WriteFile(rel, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	stand := filepath.Join(out, "claude")
	body := "#!/bin/sh\n" +
		`o=$(dirname "$0")` + "\n" +
		`echo $$ > "$o/pid"` + "\n" +
		`pwd -P > "$o/folder"` + "\n" +
		`printf '%s\0' "$@" > "$o/argv"` + "\n" +
		`env -0 > "$o/env"` + "\n"
	if err := os.WriteFile(stand, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return openRig{dir, checkout, stand, out}
}

// cmdline is the binary's command line with the rig's flags, then args.
func (r openRig) cmdline(t *testing.T, args ...string) []string {
	t.Helper()
	return append([]string{binary, "--state-dir", r.dir, "--now", at(t, 0), "--usage-command", usageStandIn(t), "--claude-command", r.stand}, args...)
}

// onTerminal runs the command line on a pty, as a person's terminal does, and
// returns what it printed.
func onTerminal(t *testing.T, line []string, setsid bool) result {
	t.Helper()
	inner := shellWords(line)
	if setsid {
		inner = "setsid -w " + inner
	}
	cmd := exec.Command("script", "-qec", inner, "/dev/null")
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	return finish(t, cmd, &out)
}

func finish(t *testing.T, cmd *exec.Cmd, out *bytes.Buffer) result {
	t.Helper()
	code := 0
	if err := cmd.Run(); err != nil {
		ee, ok := errors.AsType[*exec.ExitError](err)
		if !ok {
			t.Fatal(err)
		}
		code = ee.ExitCode()
	}
	return result{stdout: out.String(), code: code}
}

// launched is what the stand-in recorded.
type launched struct {
	pid, folder string
	argv, env   []string
}

func (r openRig) launched(t *testing.T) launched {
	t.Helper()
	split := func(name string) []string {
		s := readFile(t, filepath.Join(r.out, name))
		return strings.Split(strings.TrimSuffix(s, "\x00"), "\x00")
	}
	return launched{
		pid:    strings.TrimSpace(readFile(t, filepath.Join(r.out, "pid"))),
		folder: strings.TrimSpace(readFile(t, filepath.Join(r.out, "folder"))),
		argv:   split("argv"),
		env:    split("env"),
	}
}

// flag gives the value after name in argv.
func (s launched) flag(t *testing.T, name string) string {
	t.Helper()
	i := slices.Index(s.argv, name)
	if i < 0 || i+1 >= len(s.argv) {
		t.Fatalf("argv has no %s: %q", name, s.argv)
	}
	return s.argv[i+1]
}

func (r openRig) ran(t *testing.T) bool {
	t.Helper()
	_, err := os.Stat(filepath.Join(r.out, "argv"))
	return err == nil
}

func (r openRig) noDesks(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(r.dir, "desks.json")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("desks.json exists after a run that registers nothing: %v", err)
	}
}

func TestOpen(t *testing.T) {
	t.Run("bare opens the desk", func(t *testing.T) {
		r := newOpenRig(t)
		res := onTerminal(t, r.cmdline(t), false)
		if res.code != 0 || !r.ran(t) {
			t.Fatalf("exit %d, output %q; want the stand-in to run", res.code, res.stdout)
		}
		s := r.launched(t)
		id := s.flag(t, "--session-id")

		var d state.Desks
		if err := json.Unmarshal([]byte(readFile(t, filepath.Join(r.dir, "desks.json"))), &d); err != nil {
			t.Fatal(err)
		}
		if len(d.Desks) != 1 || d.Desks[0].SessionID != id || strconv.Itoa(d.Desks[0].PID) != s.pid {
			t.Errorf("desks.json = %+v, want one desk with id %s and PID %s", d.Desks, id, s.pid)
		}

		if got := s.flag(t, "--model"); got != "claude-opus-5-5" {
			t.Errorf("model = %q", got)
		}
		if got := s.flag(t, "--effort"); got != "high" {
			t.Errorf("effort = %q", got)
		}
		if got := s.flag(t, "--permission-mode"); got != "default" {
			t.Errorf("permission mode = %q", got)
		}
		if got := s.argv[len(s.argv)-1]; got != "What is waiting for me?" {
			t.Errorf("first line = %q", got)
		}
		prefix := binary + " --state-dir " + r.dir
		prompt := s.flag(t, "--append-system-prompt")
		for _, want := range []string{prefix, "Should the harbor chart keep its old scale?", "Should the tide table list the Sitka station?", "Load the cc-desk skill", filepath.Join(r.dir, "desk", "words")} {
			if !strings.Contains(prompt, want) {
				t.Errorf("system prompt lacks %q", want)
			}
		}
		if s.folder != r.checkout {
			t.Errorf("folder = %q, want the checkout %q", s.folder, r.checkout)
		}
		for _, kv := range s.env {
			k, _, _ := strings.Cut(kv, "=")
			if slices.Contains(scrubbedVars, k) {
				t.Errorf("environment still holds %s", k)
			}
		}

		settings := s.flag(t, "--settings")
		if want := filepath.Join(r.dir, "launch", "desk-"+id+".json"); settings != want {
			t.Errorf("settings path = %q, want %q", settings, want)
		}
		raw := readFile(t, settings)
		var cfg map[string]json.RawMessage
		if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
			t.Fatal(err)
		}
		if len(cfg) != 1 || cfg["permissions"] == nil {
			t.Errorf("settings keys = %v, want permissions only (no hook)", cfg)
		}
		for _, want := range []string{
			"Bash(" + prefix + " ruling record *)",
			"Bash(" + prefix + " brief*)",
			"Bash(" + prefix + " prompt *)",
			"Bash(" + prefix + " pass start *)",
			"Edit(/" + filepath.Join(r.dir, "desk", "words") + "/**)",
		} {
			if !strings.Contains(raw, want) {
				t.Errorf("settings lack %q", want)
			}
		}
	})

	t.Run("a second open registers a second desk", func(t *testing.T) {
		r := newOpenRig(t)
		onTerminal(t, r.cmdline(t), false)
		onTerminal(t, r.cmdline(t), false)
		var d state.Desks
		if err := json.Unmarshal([]byte(readFile(t, filepath.Join(r.dir, "desks.json"))), &d); err != nil {
			t.Fatal(err)
		}
		if len(d.Desks) != 2 || d.Desks[0].SessionID == d.Desks[1].SessionID {
			t.Errorf("desks = %+v, want two with different ids", d.Desks)
		}
	})

	t.Run("a skill file replaces the skill load", func(t *testing.T) {
		r := newOpenRig(t)
		skill := filepath.Join(t.TempDir(), "SKILL.md")
		if err := os.WriteFile(skill, []byte("skill\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		onTerminal(t, r.cmdline(t, "--skill-file", skill), false)
		s := r.launched(t)
		prompt := s.flag(t, "--append-system-prompt")
		if !strings.Contains(prompt, "Read the file at "+skill+" with the Read tool") || strings.Contains(prompt, "Load the cc-desk skill") {
			t.Errorf("system prompt = %q", prompt)
		}
		if !strings.Contains(readFile(t, s.flag(t, "--settings")), "Read(/"+skill+")") {
			t.Error("settings lack a Read rule for the skill file")
		}
	})

	t.Run("a round opens on its questions only", func(t *testing.T) {
		r := newOpenRig(t)
		res := onTerminal(t, r.cmdline(t, "north", "n1"), false)
		if res.code != 0 || !r.ran(t) {
			t.Fatalf("exit %d, output %q", res.code, res.stdout)
		}
		prompt := r.launched(t).flag(t, "--append-system-prompt")
		if !strings.Contains(prompt, "Should the harbor chart keep its old scale?") || strings.Contains(prompt, "Sitka") {
			t.Errorf("system prompt holds the wrong questions:\n%s", prompt)
		}
	})

	t.Run("a mistyped name is refused with the names it knows", func(t *testing.T) {
		for _, c := range []struct {
			args []string
			want string
		}{
			{[]string{"norht", "n1"}, "the projects are "},
			{[]string{"north", "n9"}, "its rounds are n1, n2, n3"},
			{[]string{"relase"}, "takes no arguments or a project and a round"},
		} {
			r := newOpenRig(t)
			res := onTerminal(t, r.cmdline(t, c.args...), false)
			if res.code == 0 || !strings.Contains(res.stdout, c.want) {
				t.Errorf("%v: exit %d, output %q; want a refusal with %q", c.args, res.code, res.stdout, c.want)
			}
			if r.ran(t) {
				t.Errorf("%v: the stand-in ran", c.args)
			}
			r.noDesks(t)
		}
	})

	t.Run("no terminal prints help and registers nothing", func(t *testing.T) {
		r := newOpenRig(t)
		cmd := exec.Command(binary, r.cmdline(t)[1:]...)
		cmd.Stdin = nil
		var out bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &out
		res := finish(t, cmd, &out)
		if res.code != 0 || !strings.Contains(res.stdout, "Usage:") || r.ran(t) {
			t.Errorf("</dev/null: exit %d, output %q, ran %v", res.code, res.stdout, r.ran(t))
		}
		r.noDesks(t)
	})

	t.Run("a pty with no controlling terminal prints help and registers nothing", func(t *testing.T) {
		r := newOpenRig(t)
		res := onTerminal(t, r.cmdline(t), true)
		if res.code != 0 || !strings.Contains(res.stdout, "Usage:") || r.ran(t) {
			t.Errorf("setsid: exit %d, output %q, ran %v", res.code, res.stdout, r.ran(t))
		}
		r.noDesks(t)
	})

	t.Run("a broken release file prints the error and starts nothing", func(t *testing.T) {
		r := newOpenRig(t)
		if err := os.WriteFile(filepath.Join(r.dir, "releases", "north.json"), []byte("{ not json"), 0o644); err != nil {
			t.Fatal(err)
		}
		res := onTerminal(t, r.cmdline(t), false)
		if res.code != 2 || !strings.Contains(res.stdout, "north.json") || r.ran(t) {
			t.Errorf("exit %d, output %q, ran %v; want exit 2 naming the file", res.code, res.stdout, r.ran(t))
		}
		r.noDesks(t)
	})
}

func TestRulingWordsFile(t *testing.T) {
	const deskSession = "e2e-desk-session"
	words := "it's \"fine\", $HOME stays\nsecond line"
	verb := func(file string) []string {
		return []string{"ruling", "record", "--project", "north", "--round", "n1", "--question", "north-q1", "--option", "yes", "--words-file", file}
	}

	t.Run("keeps the words byte for byte", func(t *testing.T) {
		dir := newStateFrom(t, "order")
		file := filepath.Join(t.TempDir(), "words.txt")
		if err := os.WriteFile(file, []byte(words+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		r := recordRulingArgs(t, inTerminalDesk, deskSession, dir, at(t, 20), verb(file), nil)
		if r.code != 0 {
			t.Fatalf("exit %d: %s", r.code, r.stderr)
		}
		lines := strings.Split(strings.TrimSpace(readFile(t, filepath.Join(dir, "rulings.jsonl"))), "\n")
		var got state.Ruling
		if err := json.Unmarshal([]byte(lines[len(lines)-1]), &got); err != nil {
			t.Fatal(err)
		}
		if got.Words != words {
			t.Errorf("words = %q, want %q", got.Words, words)
		}
	})

	t.Run("refuses the words twice", func(t *testing.T) {
		dir := newStateFrom(t, "order")
		file := filepath.Join(t.TempDir(), "words.txt")
		if err := os.WriteFile(file, []byte(words), 0o644); err != nil {
			t.Fatal(err)
		}
		r := recordRulingArgs(t, inTerminalDesk, deskSession, dir, at(t, 20), append(verb(file), "--", "and more"), nil)
		if r.code == 0 || !strings.Contains(r.stderr, "not both or neither") {
			t.Errorf("exit %d, stderr %q; want a refusal", r.code, r.stderr)
		}
	})
}
