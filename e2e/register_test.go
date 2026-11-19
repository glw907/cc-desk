package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/glw907/cc-desk/internal/state"
)

const registerSession = "e2e-register-session"

// registerRun is what a desk's `desk register` and a `ruling record` after
// it left behind.
type registerRun struct {
	register, ruling result
	dir              string
}

// registerUnderDesk runs `desk register` and then `ruling record` from a
// desk shell, in the desk's own order: the verbs run setsid, below the shell.
// The claude stand-in is written after the desk starts, since its row needs
// the desk's PID; rows gets that PID and returns the stand-in's output.
func registerUnderDesk(t *testing.T, terminal bool, sessionID string, rows func(deskPID int) string) registerRun {
	t.Helper()
	dir := newStateFrom(t, "order")
	work := t.TempDir()
	stand := filepath.Join(work, "claude")
	args := []string{"--state-dir", dir, "--now", at(t, 20), "--usage-command", usageStandIn(t), "--claude-command", stand}
	regArgs := append(append([]string{}, args...), "desk", "register")
	rulArgs := append(append([]string{}, args...), "ruling", "record", "--", rulingWords)

	script := "( set -- " + shellWords(regArgs) + "\n" +
		`setsid -w "$BIN" "$@" > "$OUT/reg.stdout" 2> "$OUT/reg.stderr"; echo $? > "$OUT/reg.code"` + "\n" +
		"set -- " + shellWords(rulArgs) + "\n" +
		`setsid -w "$BIN" "$@" > "$OUT/rul.stdout" 2> "$OUT/rul.stderr"; echo $? > "$OUT/rul.code" ) &`
	env := []string{"BIN=" + binary, "OUT=" + work, "CLAUDE_CODE_SESSION_ID=" + sessionID}
	pid, release, wait := startDesk(t, terminal, env, script)

	body := "#!/bin/sh\ncat <<'EOF'\n" + rows(pid) + "\nEOF\n"
	if err := os.WriteFile(stand, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	release()
	wait()

	read := func(name string) result {
		code, err := strconv.Atoi(strings.TrimSpace(readFile(t, filepath.Join(work, name+".code"))))
		if err != nil {
			t.Fatalf("the desk's %s left no exit code: %v", name, err)
		}
		return result{readFile(t, filepath.Join(work, name+".stdout")), readFile(t, filepath.Join(work, name+".stderr")), code}
	}
	return registerRun{register: read("reg"), ruling: read("rul"), dir: dir}
}

// agentRow is one `claude agents --json` row.
func agentRow(sessionID, kind string, pid int) string {
	return fmt.Sprintf(`[{"id":"ab12cd34","sessionId":%q,"name":"desk","kind":%q,"pid":%d,"state":"working"}]`, sessionID, kind, pid)
}

func TestDeskRegister(t *testing.T) {
	refused := func(t *testing.T, r registerRun, test string) {
		t.Helper()
		if r.register.code != 2 || !strings.Contains(r.register.stderr, "cannot register the desk: ") || !strings.Contains(r.register.stderr, test) {
			t.Errorf("register exit %d, stderr %q; want a refusal naming %q", r.register.code, r.register.stderr, test)
		}
		if got := readFile(t, filepath.Join(r.dir, "desks.json")); strings.Contains(got, registerSession) {
			t.Errorf("desks.json holds the session after a refusal:\n%s", got)
		}
		if r.ruling.code == 0 {
			t.Error("ruling record passed the guard after a refused registration")
		}
	}

	t.Run("interactive row under a terminal registers", func(t *testing.T) {
		r := registerUnderDesk(t, true, registerSession, func(pid int) string { return agentRow(registerSession, "interactive", pid) })
		if r.register.code != 0 || !strings.Contains(r.register.stderr, "registered desk "+registerSession) {
			t.Fatalf("register exit %d, stderr %q", r.register.code, r.register.stderr)
		}
		var d state.Desks
		if err := json.Unmarshal([]byte(readFile(t, filepath.Join(r.dir, "desks.json"))), &d); err != nil {
			t.Fatal(err)
		}
		var got *state.Desk
		for i := range d.Desks {
			if d.Desks[i].SessionID == registerSession {
				got = &d.Desks[i]
			}
		}
		if got == nil || got.PID <= 0 || got.Started.IsZero() {
			t.Fatalf("desks.json = %+v, want an entry for %s with a PID and a start time", d, registerSession)
		}
		if r.ruling.code != 0 || !strings.Contains(r.ruling.stderr, "recorded ruling") {
			t.Errorf("ruling record after registering: exit %d, stderr %q", r.ruling.code, r.ruling.stderr)
		}
	})

	t.Run("bg row is refused", func(t *testing.T) {
		r := registerUnderDesk(t, true, registerSession, func(pid int) string { return agentRow(registerSession, "bg", pid) })
		refused(t, r, "not interactive")
	})

	t.Run("pid outside the chain is refused", func(t *testing.T) {
		other := exec.Command("sleep", "60")
		if err := other.Start(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { other.Process.Kill(); other.Wait() })
		r := registerUnderDesk(t, true, registerSession, func(int) string {
			return agentRow(registerSession, "interactive", other.Process.Pid)
		})
		refused(t, r, "isn't above this command")
	})

	t.Run("desk with no terminal is refused", func(t *testing.T) {
		r := registerUnderDesk(t, false, registerSession, func(pid int) string { return agentRow(registerSession, "interactive", pid) })
		refused(t, r, "has no terminal")
	})

	t.Run("empty session id is refused", func(t *testing.T) {
		r := registerUnderDesk(t, true, "", func(pid int) string { return agentRow(registerSession, "interactive", pid) })
		refused(t, r, "CLAUDE_CODE_SESSION_ID is empty")
	})

	t.Run("no matching row is refused", func(t *testing.T) {
		r := registerUnderDesk(t, true, registerSession, func(pid int) string { return agentRow("another-session", "interactive", pid) })
		refused(t, r, "lists no row for this session")
	})
}

func TestDeskRegisterListWait(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", registerSession)
	dir := newStateFrom(t, "order")
	before := folder(t, dir)
	stand := filepath.Join(t.TempDir(), "claude")
	if err := os.WriteFile(stand, []byte("#!/bin/sh\nsleep 5\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	r := run(t, dir, at(t, 20), "--claude-command", stand, "desk", "register", "--list-wait", "1s")
	if took := time.Since(start); took > 3*time.Second {
		t.Errorf("register took %s, want it to give up within 3s", took)
	}
	if r.code == 0 || !strings.Contains(r.stderr, "no list within 1s") {
		t.Errorf("register exit %d, stderr %q; want a failure naming the wait", r.code, r.stderr)
	}
	sameFolder(t, before, folder(t, dir))
}
