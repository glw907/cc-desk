//go:build unix

package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

// signInStopped launches pass a in the rig and stops it on the sign-in: the
// working row turns blocked and its job file reads "Login expired". The check
// that stops it runs at minute 2.
func signInStopped(t *testing.T, r launchRig) {
	t.Helper()
	r.addRelease(t, true, "a")
	requireRun(t, r.do(t, workedClock, "pass", "start", "proj", "a"))
	r.check(t, workedClock)
	r.check(t, minutes(t, 1))
	r.blockOnSignIn(t)
	r.check(t, minutes(t, 2))
	if p := r.passRecord(t, "a"); p["state"] != "stopped" || p["stop_reason"] != "sign_in_expired" {
		t.Fatalf("pass = %v, want stopped on the sign-in", p)
	}
}

// blockOnSignIn turns every working row blocked, with a job file that reads
// as an expired sign-in.
func (r launchRig) blockOnSignIn(t *testing.T) {
	t.Helper()
	rows := r.readRows(t)
	for _, row := range rows {
		if row["state"] != "working" {
			continue
		}
		row["state"], row["status"] = "blocked", "blocked"
		b, err := json.Marshal(map[string]string{"detail": "Login expired · Run /login"})
		if err != nil {
			t.Fatal(err)
		}
		r.write(t, filepath.Join("jobs", row["id"].(string), "state.json"), string(b))
	}
	r.setRows(t, rows...)
}

// signIn writes the stand-in's stored sign-in at the given minute of the
// clock.
func (r launchRig) signIn(t *testing.T, minute int) {
	t.Helper()
	at, err := time.Parse(time.RFC3339, minutes(t, minute))
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(r.dir, ".credentials.json")
	r.write(t, ".credentials.json", "{}")
	if err := os.Chtimes(p, at, at); err != nil {
		t.Fatal(err)
	}
}

// authReads counts the stand-in's auth status calls.
func (r launchRig) authReads(t *testing.T) int {
	t.Helper()
	n := 0
	for _, k := range r.kinds(t) {
		if k == "auth" {
			n++
		}
	}
	return n
}

// signInTodo gives the needs_you_now line of pass a's sign-in stop.
func (r launchRig) signInTodo(t *testing.T) string {
	t.Helper()
	for _, n := range r.needs(t) {
		if n["reason"] == "sign_in_expired" {
			todo, _ := n["todo"].(string)
			return todo
		}
	}
	t.Fatalf("needs_you_now %v has no sign-in line", r.needs(t))
	return ""
}

func TestSignInResume(t *testing.T) {
	r := newLaunchRig(t)
	signInStopped(t, r)
	if got, want := r.signInTodo(t), "sign in to Claude Code, and the pass resumes by itself"; got != want {
		t.Errorf("todo = %q, want %q", got, want)
	}

	// Signed out: the read runs and the pass stays stopped.
	r.control(t, "auth-out", true)
	r.check(t, minutes(t, 3))
	if n := r.authReads(t); n != 1 {
		t.Errorf("%d sign-in reads, want 1", n)
	}
	if p := r.passRecord(t, "a"); p["state"] != "stopped" {
		t.Errorf("pass while signed out = %v, want stopped", p["state"])
	}

	// Signed in, but the stored sign-in predates the stop: still stopped.
	r.control(t, "auth-out", false)
	r.signIn(t, 1)
	r.check(t, minutes(t, 4))
	if p := r.passRecord(t, "a"); p["state"] != "stopped" {
		t.Errorf("pass with a sign-in from before the stop = %v, want stopped", p["state"])
	}
	if n := len(r.launches(t)); n != 1 {
		t.Fatalf("%d launches before a new sign-in, want 1", n)
	}

	// A new sign-in: the first check after it resumes the pass.
	old := r.sessionOf(t, "a")["short_id"].(string)
	r.signIn(t, 4)
	r.check(t, minutes(t, 5))
	p := r.passRecord(t, "a")
	if p["state"] != "building" {
		t.Fatalf("pass after a new sign-in = %v, want building", p)
	}
	stops, _ := p["stops"].([]any)
	if len(stops) != 1 || stops[0].(map[string]any)["end"] != minutes(t, 5) {
		t.Errorf("stops = %v, want one ended at %s", stops, minutes(t, 5))
	}
	if n := len(r.launches(t)); n != 2 {
		t.Errorf("%d launches after the resume, want 2", n)
	}
	if !slices.ContainsFunc(r.calls(t), func(c call) bool { return c.Args[0] == "stop" && c.Args[1] == old }) {
		t.Errorf("calls %v: want claude stop %s before the relaunch", r.kinds(t), old)
	}
	rel := relaunchesOf(t, r, "a")
	if len(rel) != 1 || rel[0]["cause"] != "sign_in_expired" || rel[0]["counted"] == true {
		t.Errorf("relaunches = %v, want one sign-in relaunch, uncounted", rel)
	}

	// The second stop resumes the same way.
	r.blockOnSignIn(t)
	r.check(t, minutes(t, 6))
	if p := r.passRecord(t, "a"); p["state"] != "stopped" {
		t.Fatalf("pass after the second block = %v, want stopped", p["state"])
	}
	r.signIn(t, 7)
	r.check(t, minutes(t, 8))
	if p := r.passRecord(t, "a"); p["state"] != "building" {
		t.Fatalf("pass after the second sign-in = %v, want building", p["state"])
	}
	if n := len(r.launches(t)); n != 3 {
		t.Errorf("%d launches after the second resume, want 3", n)
	}

	// Two resumes are spent, so the third stop stays and names the command.
	r.blockOnSignIn(t)
	r.check(t, minutes(t, 9))
	reads := r.authReads(t)
	r.signIn(t, 10)
	r.check(t, minutes(t, 11))
	if p := r.passRecord(t, "a"); p["state"] != "stopped" || p["stop_reason"] != "sign_in_expired" {
		t.Errorf("pass after the third stop = %v, want stopped on the sign-in", p)
	}
	if n := len(r.launches(t)); n != 3 {
		t.Errorf("%d launches after the third stop, want 3", n)
	}
	if n := r.authReads(t); n != reads {
		t.Errorf("%d sign-in reads after the resumes ran out, want none", n-reads)
	}
	if got, want := r.signInTodo(t), "sign in to Claude Code, then run cc-desk pass start proj a"; got != want {
		t.Errorf("todo = %q, want %q", got, want)
	}
	for _, e := range relaunchesOf(t, r, "a") {
		if e["counted"] == true {
			t.Errorf("relaunch %v is counted, want no counted relaunch", e)
		}
	}
}

func TestSignInReadThatSleeps(t *testing.T) {
	r := newLaunchRig(t)
	signInStopped(t, r)
	r.control(t, "auth-sleeps", true)
	r.signIn(t, 3)

	start := time.Now()
	got := r.run(t, minutes(t, 4), "check", "--sign-in-limit", "500ms")
	requireRun(t, got)
	if took := time.Since(start); took > 10*time.Second {
		t.Errorf("the check took %s, want the read ended near its 500ms limit", took)
	}
	if n := r.authReads(t); n != 1 {
		t.Errorf("%d sign-in reads, want 1", n)
	}
	if p := r.passRecord(t, "a"); p["state"] != "stopped" || p["stop_reason"] != "sign_in_expired" {
		t.Errorf("pass after a read that slept = %v, want stopped on the sign-in", p)
	}
	if n := len(r.launches(t)); n != 1 {
		t.Errorf("%d launches, want 1", n)
	}
}

// TestSignInResumeStopFails resumes a pass whose old row won't stop. The
// relaunch takes the usage relaunch's branch, so nothing launches over the
// live row and the pass stops as a failed launch.
func TestSignInResumeStopFails(t *testing.T) {
	r := newLaunchRig(t)
	signInStopped(t, r)
	r.control(t, "stop-fails", true)
	r.signIn(t, 3)
	if got := r.run(t, minutes(t, 4), "check"); got.code != 2 {
		t.Errorf("check exit %d, want 2 for the stop that failed; stderr: %s", got.code, got.stderr)
	}

	if !slices.Contains(r.kinds(t), "stop") {
		t.Errorf("calls %v: want a claude stop", r.kinds(t))
	}
	if n := len(r.launches(t)); n != 1 {
		t.Errorf("%d launches, want none over the live row", n)
	}
	if p := r.passRecord(t, "a"); p["state"] != "stopped" || p["stop_reason"] != "launch_failed" {
		t.Errorf("pass = %v, want stopped as launch_failed", p)
	}
}
