//go:build unix

package e2e

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const (
	refusalFixture   = "testdata/refusal/hook-input.json"
	refusalFixtureID = `"session_id":"615f5327-829c-4425-949e-5b94376bb769"`
	refusalCommand   = "cd /var/home/glw907/Projects/cc-desk-scratch && git stash drop stash@{0}"
)

// refusalInput gives the kept PermissionDenied input with its session_id replaced.
func refusalInput(t *testing.T, sessionID string) string {
	t.Helper()
	b, err := os.ReadFile(refusalFixture)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(b, []byte(refusalFixtureID)) {
		t.Fatalf("%s no longer holds %s", refusalFixture, refusalFixtureID)
	}
	return strings.Replace(string(b), refusalFixtureID, `"session_id":`+strconv.Quote(sessionID), 1)
}

// runRefusalHook runs the PermissionDenied hook command line the pass's
// settings file names, as claude would.
func (r launchRig) runRefusalHook(t *testing.T, input string) result {
	t.Helper()
	var s struct {
		Hooks struct {
			PermissionDenied []struct{ Hooks []struct{ Command string } }
		}
	}
	if err := json.Unmarshal([]byte(readFileIn(t, r.state, "launch/proj-a.json")), &s); err != nil {
		t.Fatal(err)
	}
	if len(s.Hooks.PermissionDenied) != 1 || len(s.Hooks.PermissionDenied[0].Hooks) != 1 {
		t.Fatalf("settings hooks = %+v, want one PermissionDenied command", s.Hooks)
	}
	cmd := exec.Command("sh", "-c", s.Hooks.PermissionDenied[0].Hooks[0].Command)
	cmd.Env = append(os.Environ(), "PATH=/usr/bin:/bin")
	cmd.Stdin = strings.NewReader(input)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	code := 0
	if err := cmd.Run(); err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}
		code = ee.ExitCode()
	}
	return result{out.String(), errOut.String(), code}
}

func (r launchRig) refusalsOf(t *testing.T) []any {
	t.Helper()
	rs, _ := r.sessionOf(t, "a")["refusals"].([]any)
	return rs
}

func TestRefusalHook(t *testing.T) {
	r := newLaunchRig(t)
	r.addRelease(t, true, "a")
	requireRun(t, r.do(t, workedClock, "pass", "start", "proj", "a"))
	r.check(t, workedClock)
	sessionID := r.sessionOf(t, "a")["session_id"].(string)

	got := r.runRefusalHook(t, refusalInput(t, sessionID))
	if got.code != 0 || got.stdout != "" {
		t.Fatalf("hook: exit %d, stdout %q, stderr %q; want exit 0 and nothing on stdout", got.code, got.stdout, got.stderr)
	}
	rs := r.refusalsOf(t)
	if len(rs) != 1 {
		t.Fatalf("refusals = %v, want one", rs)
	}
	ref := rs[0].(map[string]any)
	if ref["tool"] != "Bash" || ref["target"] != refusalCommand {
		t.Errorf("refusal = %v, want Bash and %q", ref, refusalCommand)
	}
	if reason, _ := ref["reason"].(string); reason == "" {
		t.Errorf("refusal = %v, want a reason", ref)
	}

	got = r.runRefusalHook(t, refusalInput(t, "no-such-session"))
	if got.code != 0 || got.stdout != "" {
		t.Errorf("unknown session: exit %d, stdout %q; want exit 0 and nothing on stdout", got.code, got.stdout)
	}
	entries, _ := os.ReadDir(filepath.Join(r.state, "hook-unmatched"))
	if len(entries) != 1 || !strings.Contains(entries[0].Name(), "refusal") {
		t.Errorf("hook-unmatched = %v, want one refusal input", entries)
	}
	if n := len(r.refusalsOf(t)); n != 1 {
		t.Errorf("%d refusals after the unknown session, want 1", n)
	}
	requireRun(t, r.do(t, workedClock, "snapshot"))
}

func TestRefusalHookLogsAFailure(t *testing.T) {
	r := newLaunchRig(t)
	r.addRelease(t, true, "a")
	requireRun(t, r.do(t, workedClock, "pass", "start", "proj", "a"))
	r.check(t, workedClock)
	sessionID := r.sessionOf(t, "a")["session_id"].(string)

	file := filepath.Join(r.state, "releases", "proj.json")
	if err := os.WriteFile(file, []byte("{ not a release"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := r.runRefusalHook(t, refusalInput(t, sessionID))
	if got.code != 0 || got.stdout != "" {
		t.Fatalf("hook: exit %d, stdout %q, stderr %q; want exit 0 and nothing on stdout", got.code, got.stdout, got.stderr)
	}
	lines := strings.Split(strings.TrimSpace(readFile(t, filepath.Join(r.state, "refusal-hook.log"))), "\n")
	if len(lines) != 1 || !strings.Contains(lines[0], sessionID) {
		t.Errorf("refusal-hook.log = %q, want one line naming the session", lines)
	}
}

func TestRelaunchKeepsRefusal(t *testing.T) {
	r := newLaunchRig(t)
	r.addRelease(t, true, "a")
	requireRun(t, r.do(t, workedClock, "pass", "start", "proj", "a"))
	r.check(t, workedClock)
	r.check(t, minutes(t, 1))
	first := r.sessionOf(t, "a")
	got := r.runRefusalHook(t, refusalInput(t, first["session_id"].(string)))
	if got.code != 0 || len(r.refusalsOf(t)) != 1 {
		t.Fatalf("hook: exit %d, refusals %v, want exit 0 and one refusal", got.code, r.refusalsOf(t))
	}

	r.vanish(t, first["short_id"].(string))
	r.check(t, minutes(t, 2))
	r.check(t, minutes(t, 3))
	next := r.sessionOf(t, "a")
	if next["short_id"] == first["short_id"] {
		t.Fatalf("the session was not relaunched: %v", next)
	}
	rs := r.refusalsOf(t)
	if len(rs) != 1 || rs[0].(map[string]any)["target"] != refusalCommand {
		t.Errorf("refusals after the relaunch = %v, want the one from before", rs)
	}
}
