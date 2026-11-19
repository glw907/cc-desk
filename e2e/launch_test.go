//go:build unix

package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// launchAllow is the allow rule every launch test's release carries, so the
// settings file can be read back for it.
const launchAllow = "Bash(go test ./...)"

// addRelease installs a release named proj with one round, r1, and the named
// passes, none started. With a checkout it records the rig's clone.
func (r launchRig) addRelease(t *testing.T, checkout bool, passes ...string) {
	t.Helper()
	var ps []string
	for _, p := range passes {
		ps = append(ps, fmt.Sprintf(`{"name": %q, "summary": "A pass.", "round": "r1", "planned_minutes": 30, "state": "not_started"}`, p))
	}
	release := fmt.Sprintf(`{
  "format": 1, "project": "proj", "name": "v1", "repo": "glw907/proj",
  "allow_rules": [%q],
  "rounds": [{"name": "r1", "summary": "A round.", "needs_geoff": false, "no_need_reason": "A fixture.", "design_written": true}],
  "passes": [%s]
}`, launchAllow, strings.Join(ps, ","))
	file := filepath.Join(r.dir, "release.json")
	if err := os.WriteFile(file, []byte(release), 0o644); err != nil {
		t.Fatal(err)
	}
	requireRun(t, run(t, r.state, workedClock, "release", "add", file))
	if checkout {
		requireRun(t, run(t, r.state, workedClock, "release", "checkout", "proj", r.clone, "--default-branch", "main"))
	}
}

// do runs a release or pass verb on the rig's state at the clock.
func (r launchRig) do(t *testing.T, now string, verb ...string) result {
	t.Helper()
	return run(t, r.state, now, verb...)
}

// release reads the project's release file.
func (r launchRig) release(t *testing.T) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal([]byte(readFileIn(t, r.state, "releases/proj.json")), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

// passRecord gives the named pass's JSON object from the release file.
func (r launchRig) passRecord(t *testing.T, name string) map[string]any {
	t.Helper()
	for _, p := range r.release(t)["passes"].([]any) {
		if m := p.(map[string]any); m["name"] == name {
			return m
		}
	}
	t.Fatalf("no pass %q in the release", name)
	return nil
}

// sessionOf gives a pass's session record.
func (r launchRig) sessionOf(t *testing.T, pass string) map[string]any {
	t.Helper()
	s, ok := r.passRecord(t, pass)["session"].(map[string]any)
	if !ok {
		t.Fatalf("pass %q has no session record: %v", pass, r.passRecord(t, pass))
	}
	return s
}

func TestLaunchFreePass(t *testing.T) {
	r := newLaunchRig(t)
	r.addRelease(t, true, "a")
	pushed := r.pushAnother(t, "newer.txt")
	requireRun(t, r.do(t, workedClock, "pass", "start", "proj", "a"))

	got := r.run(t, workedClock, "check")
	if got.code != 0 {
		t.Fatalf("check exit %d, stderr: %s", got.code, got.stderr)
	}

	launches := r.launches(t)
	if len(launches) != 1 {
		t.Fatalf("%d claude launches, want 1: %v", len(launches), r.kinds(t))
	}
	worktree := filepath.Join(r.clone, ".claude", "worktrees", "a")
	if head := git(t, worktree, "rev-parse", "HEAD"); head != pushed {
		t.Errorf("worktree HEAD %s, want the commit pushed after the clone, %s", head, pushed)
	}
	if branch := git(t, worktree, "rev-parse", "--abbrev-ref", "HEAD"); branch != "a" {
		t.Errorf("worktree branch %q, want a", branch)
	}

	scope := r.scopeLog(t)
	if n := strings.Count(scope, "\n"); n != 1 || !strings.Contains(scope, "flags=--user --scope -q") || !strings.Contains(scope, "cwd="+worktree) {
		t.Errorf("scope.log = %q, want one launch from %s", scope, worktree)
	}

	c := launches[0]
	if c.Cwd != worktree {
		t.Errorf("claude ran in %s, want %s", c.Cwd, worktree)
	}
	settings := filepath.Join(r.state, "launch", "proj-a.json")
	if i := slices.Index(c.Args, "--permission-mode"); i < 0 || c.Args[i+1] != "auto" {
		t.Errorf("args lack --permission-mode auto: %v", c.Args)
	}
	if i := slices.Index(c.Args, "--settings"); i < 0 || c.Args[i+1] != settings {
		t.Errorf("args lack --settings %s: %v", settings, c.Args)
	}
	disallowed := slices.Index(c.Args, "--disallowedTools")
	prompt := slices.IndexFunc(c.Args, func(a string) bool { return strings.Contains(a, "proj-a") && strings.Contains(a, "\n") })
	if prompt < 0 || disallowed < 0 || prompt > disallowed {
		t.Errorf("the prompt (arg %d) must come before --disallowedTools (arg %d): %v", prompt, disallowed, c.Args)
	}
	for _, a := range c.Args {
		if strings.Contains(a, "bypassPermissions") || a == "--dangerously-skip-permissions" {
			t.Errorf("args hold %q", a)
		}
	}

	var s struct {
		Permissions struct{ Allow, Deny []string }
		Hooks       struct {
			PermissionRequest []struct{ Hooks []struct{ Command string } }
		}
	}
	if err := json.Unmarshal([]byte(readFileIn(t, r.state, "launch/proj-a.json")), &s); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(s.Permissions.Allow, launchAllow) {
		t.Errorf("settings allow %v lack the release's rule", s.Permissions.Allow)
	}
	for _, d := range []string{"Bash(gh pr merge *)", "Bash(approval-queue land *)"} {
		if !slices.Contains(s.Permissions.Deny, d) {
			t.Errorf("settings deny %v lack %s", s.Permissions.Deny, d)
		}
	}
	if len(s.Hooks.PermissionRequest) != 1 || !strings.HasSuffix(s.Hooks.PermissionRequest[0].Hooks[0].Command, "prompt hook") {
		t.Errorf("settings hooks = %+v, want one PermissionRequest entry ending in prompt hook", s.Hooks)
	}

	sess := r.sessionOf(t, "a")
	if sess["name"] != "proj-a" || sess["branch"] != "a" || sess["worktree"] != worktree || sess["short_id"] != "b0000001" || sess["session_id"] != "b0000001-0000" {
		t.Errorf("session record = %v", sess)
	}
}

func TestLaunchSecondStartRefused(t *testing.T) {
	r := newLaunchRig(t)
	r.addRelease(t, true, "a")
	requireRun(t, r.do(t, workedClock, "pass", "start", "proj", "a"))
	requireRun(t, r.run(t, workedClock, "check"))

	got := r.do(t, workedClock, "pass", "start", "proj", "a")
	if got.code != 2 || !strings.Contains(got.stderr, "proj-a") {
		t.Errorf("second start: exit %d, stderr %q; want exit 2 naming proj-a", got.code, got.stderr)
	}
}

func TestLaunchWorktreeHeldByAnotherRecord(t *testing.T) {
	r := newLaunchRig(t)
	r.addRelease(t, true, "r1-design")
	requireRun(t, r.run(t, workedClock, "session", "start", "design", "proj", "r1"))
	requireRun(t, r.do(t, workedClock, "pass", "start", "proj", "r1-design"))
	before := len(r.launches(t))

	got := r.run(t, workedClock, "check")
	if got.code != 2 || !strings.Contains(got.stderr, "round:r1") {
		t.Errorf("check: exit %d, stderr %q; want exit 2 naming the design writer's record", got.code, got.stderr)
	}
	if n := len(r.launches(t)); n != before {
		t.Errorf("%d launches after the conflict, want %d", n, before)
	}
	p := r.passRecord(t, "r1-design")
	if p["state"] != "stopped" || p["stop_reason"] != "launch_failed" {
		t.Errorf("pass = %v, want stopped with launch_failed", p)
	}
}

func TestLaunchRefusesAnUnrecordedFolder(t *testing.T) {
	r := newLaunchRig(t)
	r.addRelease(t, true, "a")
	folder := filepath.Join(r.clone, ".claude", "worktrees", "a")
	git(t, r.clone, "worktree", "add", "--no-track", folder, "-b", "a", "origin/main")
	requireRun(t, r.do(t, workedClock, "pass", "start", "proj", "a"))

	got := r.run(t, workedClock, "check")
	if got.code != 2 || !strings.Contains(got.stderr, folder) {
		t.Errorf("check: exit %d, stderr %q; want exit 2 naming %s", got.code, got.stderr, folder)
	}
	if n := len(r.launches(t)); n != 0 {
		t.Errorf("%d launches into a folder no record names", n)
	}
	if p := r.passRecord(t, "a"); p["state"] != "stopped" || p["stop_reason"] != "launch_failed" {
		t.Errorf("pass = %v, want stopped with launch_failed", p)
	}
}

func TestLaunchNoCheckout(t *testing.T) {
	r := newLaunchRig(t)
	r.addRelease(t, false, "a")

	got := r.run(t, workedClock, "session", "start", "conductor", "proj")
	if got.code != 2 || !strings.Contains(got.stderr, "no checkout") {
		t.Errorf("session start: exit %d, stderr %q; want exit 2 naming the missing checkout", got.code, got.stderr)
	}

	requireRun(t, r.do(t, workedClock, "pass", "start", "proj", "a"))
	got = r.run(t, workedClock, "check")
	requireRun(t, got)
	if !strings.Contains(got.stdout, `"decision":"launch"`) || !strings.Contains(got.stdout, "not carried out") {
		t.Errorf("check output %q; want a launch decision not carried out", got.stdout)
	}
	if n := len(r.launches(t)); n != 0 {
		t.Errorf("%d launches with no checkout", n)
	}
}

// TestLaunchPrompts builds each kind's prompt the way a launch does and reads
// the prompt argument the stand-in claude was given.
func TestLaunchPrompts(t *testing.T) {
	firstPrompt := func(t *testing.T, r launchRig) string {
		t.Helper()
		l := r.launches(t)
		if len(l) != 1 {
			t.Fatalf("%d launches, want 1", len(l))
		}
		return relaunchPrompt(l[0])
	}
	requireWorkPrompt := func(t *testing.T, p, line string) {
		t.Helper()
		for _, want := range []string{"`cc-desk-work`", line} {
			if !strings.Contains(p, want) {
				t.Errorf("prompt lacks %q:\n%s", want, p)
			}
		}
		for _, bad := range []string{"approval-queue", "Until cc-desk runs"} {
			if strings.Contains(p, bad) {
				t.Errorf("prompt holds %q:\n%s", bad, p)
			}
		}
	}
	// launched starts pass a and lets the check launch it and see it work.
	launched := func(t *testing.T) launchRig {
		t.Helper()
		r := newLaunchRig(t)
		r.addRelease(t, true, "a")
		requireRun(t, r.do(t, workedClock, "pass", "start", "proj", "a"))
		r.check(t, minutes(t, 1))
		r.check(t, minutes(t, 2))
		return r
	}

	t.Run("pass", func(t *testing.T) {
		r := launched(t)
		requireWorkPrompt(t, firstPrompt(t, r), "--state-dir "+r.state+" question add proj r1 <file>")
	})
	t.Run("design", func(t *testing.T) {
		r := newLaunchRig(t)
		r.addRelease(t, true, "a")
		requireRun(t, r.run(t, workedClock, "session", "start", "design", "proj", "r1"))
		requireWorkPrompt(t, firstPrompt(t, r), "--state-dir "+r.state+" question add proj r1 <file>")
	})
	t.Run("conductor", func(t *testing.T) {
		r := newLaunchRig(t)
		r.addRelease(t, true, "a")
		requireRun(t, r.run(t, workedClock, "session", "start", "conductor", "proj"))
		requireWorkPrompt(t, firstPrompt(t, r), "--state-dir "+r.state+" report add proj <file> [--round <round>]")
	})
	t.Run("uncounted_relaunch", func(t *testing.T) {
		r := launched(t)
		requireRun(t, r.do(t, minutes(t, 3), "pass", "stop", "proj", "a", "--reason", "usage"))
		r.check(t, minutes(t, 4))
		l := r.launches(t)
		if len(l) != 2 {
			t.Fatalf("%d launches after the usage stop, want 2", len(l))
		}
		p := relaunchPrompt(l[1])
		if strings.Contains(p, "relaunch 1 of 2") || !strings.Contains(p, "doesn't count toward the two") {
			t.Errorf("an uncounted relaunch's prompt reads:\n%s", p)
		}
	})
	t.Run("counted_relaunch", func(t *testing.T) {
		r := launched(t)
		r.vanish(t, r.sessionOf(t, "a")["short_id"].(string))
		r.check(t, minutes(t, 3))
		l := r.launches(t)
		if len(l) != 2 {
			t.Fatalf("%d launches after the row vanished, want 2", len(l))
		}
		if p := relaunchPrompt(l[1]); !strings.Contains(p, "This is relaunch 1 of 2") {
			t.Errorf("a counted relaunch's prompt lacks its number:\n%s", p)
		}
	})
}
