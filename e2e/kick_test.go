//go:build unix

package e2e

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/glw907/cc-desk/internal/desk"
	"github.com/glw907/cc-desk/internal/state"
)

const kickLine = "--user start --no-block cc-desk-check.service"

// kickStandIn writes a systemctl stand-in that logs each call's arguments and
// whether the release file already holds a building pass and a pull request, so a test can see
// the kick came after the commit. It returns the script's folder, its path,
// and its log's path.
func kickStandIn(t *testing.T, stateDir string, fail bool) (dir, script, log string) {
	t.Helper()
	dir = t.TempDir()
	script = filepath.Join(dir, "systemctl")
	log = filepath.Join(dir, "calls.log")
	body := "#!/bin/sh\necho \"$* building=$(grep -c building " + filepath.Join(stateDir, "releases", "proj.json") + " 2>/dev/null) pr=$(grep -c pull_request " + filepath.Join(stateDir, "releases", "proj.json") + " 2>/dev/null)\" >> " + log + "\n"
	if fail {
		body += "echo 'no bus' >&2\nexit 1\n"
	}
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir, script, log
}

// runKick runs the binary with exactly the environment given plus a PATH of
// pathDirs then /usr/bin and /bin, and the state folder and clock named. It
// sets --usage-command and nothing else, so a kick happens only by the
// folder or a flag the test passes in args.
func runKick(t *testing.T, home, pathDir, dir, now string, args ...string) result {
	t.Helper()
	full := append([]string{"--state-dir", dir, "--now", now, "--usage-command", usageStandIn(t)}, args...)
	cmd := exec.Command(binary, full...)
	path := "/usr/bin:/bin"
	if pathDir != "" {
		path = pathDir + ":" + path
	}
	cmd.Env = append(os.Environ(), "PATH="+path, "HOME="+home)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	code := 0
	if err := cmd.Run(); err != nil {
		ee, ok := errors.AsType[*exec.ExitError](err)
		if !ok {
			t.Fatalf("run %v: %v", args, err)
		}
		code = ee.ExitCode()
	}
	return result{out.String(), errOut.String(), code}
}

func logLines(t *testing.T, log string) []string {
	t.Helper()
	b, err := os.ReadFile(log)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimSpace(string(b)), "\n")
}

// kickRelease adds the rig's release with passes a and b to the rig's state.
func kickRelease(t *testing.T) launchRig {
	t.Helper()
	r := newLaunchRig(t)
	r.addRelease(t, false, "a", "b")
	return r
}

var readyFlags = []string{"--pr", "7", "--commit", "abc123", "--verdict", "accept", "--design", "inside"}

// rulingUnderDesk runs `ruling record` for the order fixture's north-q1
// from a desk that has a terminal, with the extra global arguments given,
// and returns the verb's result.
func rulingUnderDesk(t *testing.T, dir, now string, global []string, withQuestion bool) result {
	t.Helper()
	const deskSession = "e2e-desk-session"
	out := t.TempDir()
	args := append([]string{"--state-dir", dir, "--now", now, "--usage-command", usageStandIn(t)}, global...)
	if withQuestion {
		args = append(args, rulingArgs("north-q1", "yes")...)
	} else {
		args = append(args, "ruling", "record", "--", rulingWords)
	}
	script := "set -- " + shellWords(args) + "\n" +
		`( setsid -w "$BIN" "$@" > "$OUT/stdout" 2> "$OUT/stderr"; echo $? > "$OUT/code" ) &`
	env := []string{"BIN=" + binary, "OUT=" + out, "CLAUDE_CODE_SESSION_ID=" + deskSession}
	pid, release, wait := startDesk(t, true, env, script)
	entry, err := desk.Entry(deskSession, pid)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(state.Desks{Format: state.Format, Desks: []state.Desk{entry}})
	if err != nil {
		t.Fatal(err)
	}
	writeIn(t, dir, "desks.json", string(data)+"\n")
	writeIn(t, dir, "rulings.jsonl", earlierRuling)
	release()
	wait()
	code, err := strconv.Atoi(strings.TrimSpace(readFile(t, filepath.Join(out, "code"))))
	if err != nil {
		t.Fatalf("the desk's verb left no exit code: %v", err)
	}
	return result{readFile(t, filepath.Join(out, "stdout")), readFile(t, filepath.Join(out, "stderr")), code}
}

func TestKick(t *testing.T) {
	t.Run("start kicks once, after the commit", func(t *testing.T) {
		r := kickRelease(t)
		_, stand, log := kickStandIn(t, r.state, false)
		got := runKick(t, t.TempDir(), "", r.state, workedClock, "--systemctl-command", stand, "pass", "start", "proj", "a")
		if got.code != 0 {
			t.Fatalf("start: exit %d: %s", got.code, got.stderr)
		}
		lines := logLines(t, log)
		if len(lines) != 1 || lines[0] != kickLine+" building=1 pr=0" {
			t.Fatalf("the log is %q, want one kick after the event", lines)
		}
	})

	t.Run("ready kicks once, after the commit", func(t *testing.T) {
		m := newMergeRig(t, "a")
		_, stand, log := kickStandIn(t, m.state, false)
		args := append([]string{"--systemctl-command", stand, "pass", "ready", "proj", "a"}, readyFlags...)
		got := runKick(t, t.TempDir(), "", m.state, workedClock, args...)
		if got.code != 0 {
			t.Fatalf("ready: exit %d: %s", got.code, got.stderr)
		}
		if lines := logLines(t, log); len(lines) != 1 || lines[0] != kickLine+" building=1 pr=1" {
			t.Fatalf("the log is %q, want one kick", lines)
		}
		if pr := m.passRecord(t, "a")["pull_request"]; pr == nil {
			t.Error("the pass holds no pull request after the kick")
		}
	})

	t.Run("a ruling kicks only with a question", func(t *testing.T) {
		_, stand, log := kickStandIn(t, "/nonexistent", false)
		flag := []string{"--systemctl-command", stand}
		without := newStateFrom(t, "order")
		if r := rulingUnderDesk(t, without, at(t, 20), flag, false); r.code != 0 {
			t.Fatalf("ruling without a question: exit %d: %s", r.code, r.stderr)
		}
		if lines := logLines(t, log); lines != nil {
			t.Errorf("a ruling with no question logged %q", lines)
		}
		with := newStateFrom(t, "order")
		if r := rulingUnderDesk(t, with, at(t, 20), flag, true); r.code != 0 {
			t.Fatalf("ruling with a question: exit %d: %s", r.code, r.stderr)
		}
		if lines := logLines(t, log); len(lines) != 1 || !strings.HasPrefix(lines[0], kickLine) {
			t.Errorf("a ruling with a question logged %q, want one kick", lines)
		}
	})

	t.Run("a refused start logs nothing", func(t *testing.T) {
		r := kickRelease(t)
		_, stand, log := kickStandIn(t, r.state, false)
		got := runKick(t, t.TempDir(), "", r.state, workedClock, "--systemctl-command", stand, "pass", "start", "proj", "zzz")
		if got.code == 0 {
			t.Fatal("start of an unknown pass exited 0")
		}
		if lines := logLines(t, log); lines != nil {
			t.Errorf("a refused start logged %q", lines)
		}
	})

	t.Run("a failed kick leaves the verb at exit 0", func(t *testing.T) {
		r := kickRelease(t)
		_, stand, _ := kickStandIn(t, r.state, true)
		got := runKick(t, t.TempDir(), "", r.state, workedClock, "--systemctl-command", stand, "pass", "start", "proj", "a")
		if got.code != 0 {
			t.Fatalf("exit %d, want 0: %s", got.code, got.stderr)
		}
		if strings.Count(got.stderr, "kick:") != 1 || !strings.Contains(got.stderr, "no bus") {
			t.Errorf("stderr %q, want one kick: line naming the failure", got.stderr)
		}
	})

	t.Run("a temporary state folder never kicks", func(t *testing.T) {
		r := kickRelease(t)
		pathDir, _, log := kickStandIn(t, r.state, false)
		got := runKick(t, t.TempDir(), pathDir, r.state, workedClock, "pass", "start", "proj", "a")
		if got.code != 0 {
			t.Fatalf("exit %d: %s", got.code, got.stderr)
		}
		if lines := logLines(t, log); lines != nil {
			t.Errorf("a temporary folder kicked: %q", lines)
		}
	})

	t.Run("the live state folder kicks", func(t *testing.T) {
		home := t.TempDir()
		live := filepath.Join(home, ".local", "state", "cc-desk")
		r := newLaunchRig(t)
		r.state = live
		r.addRelease(t, false, "a")
		pathDir, _, log := kickStandIn(t, live, false)
		got := runKick(t, home, pathDir, live, workedClock, "pass", "start", "proj", "a")
		if got.code != 0 {
			t.Fatalf("exit %d: %s", got.code, got.stderr)
		}
		if lines := logLines(t, log); len(lines) != 1 || !strings.HasPrefix(lines[0], kickLine) {
			t.Errorf("the live folder logged %q, want one kick", lines)
		}
	})
}
