//go:build unix

package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// finishPass marks the named pass done in the release file, as a merged pass
// is, so a report falls due for its round.
func (r launchRig) finishPass(t *testing.T, name string) {
	t.Helper()
	rel := r.release(t)
	for _, p := range rel["passes"].([]any) {
		if m := p.(map[string]any); m["name"] == name {
			m["state"] = "done"
			m["started"] = "2026-11-20T21:00:00Z"
			m["finished"] = "2026-11-20T21:30:00Z"
			m["actual_minutes"] = 30
		}
	}
	b, err := json.MarshalIndent(rel, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(r.state, "releases", "proj.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
}

// addRound adds a second round with one pass not started, so a check has a
// decision to make beside the first round's report.
func (r launchRig) addRound(t *testing.T, round, pass string) {
	t.Helper()
	rel := r.release(t)
	rel["rounds"] = append(rel["rounds"].([]any), map[string]any{
		"name": round, "summary": "A round.", "needs_geoff": false, "no_need_reason": "A fixture.", "design_written": true,
	})
	rel["passes"] = append(rel["passes"].([]any), map[string]any{
		"name": pass, "summary": "A pass.", "round": round, "planned_minutes": 30, "state": "not_started",
	})
	b, err := json.MarshalIndent(rel, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(r.state, "releases", "proj.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
}

// conductorLaunches gives the launches that started the project's conductor.
func (r launchRig) conductorLaunches(t *testing.T) []call {
	t.Helper()
	var out []call
	for _, c := range r.launches(t) {
		if strings.HasSuffix(c.Cwd, filepath.Join("worktrees", "conductor")) {
			out = append(out, c)
		}
	}
	return out
}

// finishConductor has the conductor's session finish with no report, and
// runs the check that ends it.
func (r launchRig) finishConductor(t *testing.T, now string) {
	t.Helper()
	requireRun(t, r.do(t, now, "session", "finish", "conductor", "proj"))
	r.check(t, now)
}

func TestConductorStart(t *testing.T) {
	worktree := func(r launchRig) string { return filepath.Join(r.clone, ".claude", "worktrees", "conductor") }

	t.Run("a round's report falls due and the conductor starts for it", func(t *testing.T) {
		r := newLaunchRig(t)
		r.addRelease(t, true, "a")
		r.finishPass(t, "a")
		got := r.check(t, minutes(t, 1))

		if n := len(r.conductorLaunches(t)); n != 1 {
			t.Fatalf("%d conductor launches, want 1: %v", n, r.kinds(t))
		}
		c := r.conductor(t)
		if c["for_report"] != "r1" || c["kind"] != "conductor" || c["worktree"] != worktree(r) {
			t.Errorf("conductor record = %v, want kind conductor for r1 in %s", c, worktree(r))
		}
		if branch := git(t, worktree(r), "rev-parse", "--abbrev-ref", "HEAD"); branch != "conductor" {
			t.Errorf("worktree branch %q, want conductor", branch)
		}
		if !strings.Contains(got.stdout, `"reason":"report_due"`) {
			t.Errorf("check output %q lacks the report_due launch", got.stdout)
		}
		prompt := relaunchPrompt(r.conductorLaunches(t)[0])
		if !strings.Contains(prompt, "Write the report for round r1") || !strings.Contains(prompt, "do nothing else") {
			t.Errorf("prompt lacks the report block:\n%s", prompt)
		}
		if strings.Contains(prompt, "Conduct the rounds") {
			t.Errorf("prompt still says to conduct the rounds:\n%s", prompt)
		}
		outcomes := logOutcomes(t, r.state)
		if len(outcomes) != 1 || !strings.HasPrefix(outcomes[0], "launched ") {
			t.Errorf("check log = %v, want one launched line", outcomes)
		}

		// A later check leaves the live conductor alone.
		r.check(t, minutes(t, 2))
		if n := len(r.conductorLaunches(t)); n != 1 {
			t.Errorf("%d conductor launches after a second check, want 1", n)
		}
	})

	t.Run("no report is due while a pass is not done", func(t *testing.T) {
		r := newLaunchRig(t)
		r.addRelease(t, true, "a", "b")
		r.finishPass(t, "a")
		r.check(t, minutes(t, 1))
		if n := len(r.conductorLaunches(t)); n != 0 {
			t.Errorf("%d conductor launches, want none", n)
		}
		if _, ok := r.release(t)["conductor"]; ok {
			t.Error("the release has a conductor record")
		}
	})

	t.Run("a release with no checkout waits", func(t *testing.T) {
		r := newLaunchRig(t)
		r.addRelease(t, false, "a")
		r.finishPass(t, "a")
		r.check(t, minutes(t, 1))
		if n := len(r.launches(t)); n != 0 {
			t.Errorf("%d launches with no checkout, want none", n)
		}
	})

	t.Run("a recorded report ends the conductor and removes its worktree", func(t *testing.T) {
		r := newLaunchRig(t)
		r.addRelease(t, true, "a")
		r.finishPass(t, "a")
		r.check(t, minutes(t, 1))
		first := r.conductor(t)["short_id"].(string)
		requireRun(t, r.do(t, minutes(t, 2), "report", "add", "proj", input(t, "report.json"), "--round", "r1"))
		requireRun(t, r.do(t, minutes(t, 3), "session", "finish", "conductor", "proj"))
		got := r.check(t, minutes(t, 3))

		if !strings.Contains(got.stdout, `"decision":"end"`) {
			t.Errorf("check output %q lacks the end", got.stdout)
		}
		if c := r.conductor(t); c["ended"] == nil {
			t.Errorf("conductor record = %v, want ended", c)
		}
		if _, err := os.Stat(worktree(r)); !os.IsNotExist(err) {
			t.Errorf("the conductor worktree still stands: %v", err)
		}
		if out := git(t, r.clone, "branch", "--list", "conductor"); out != "" {
			t.Errorf("branch conductor still exists: %q", out)
		}
		outcomes := logOutcomes(t, r.state)
		if last := outcomes[len(outcomes)-1]; last != "ended "+first+"; removed the conductor worktree and branch" {
			t.Errorf("last log line = %q", last)
		}
		if want := expectedFromCheck(t, r.state); !slices.Equal(outcomes[len(outcomes)-len(want):], want) {
			t.Errorf("the run's log lines = %v, want %v", outcomes[len(outcomes)-len(want):], want)
		}

		// With every round reported, the release's report falls due next.
		r.check(t, minutes(t, 4))
		if n := len(r.conductorLaunches(t)); n != 2 {
			t.Fatalf("%d conductor launches, want 2", n)
		}
		if c := r.conductor(t); c["for_report"] != "release" || c["ended"] != nil {
			t.Errorf("conductor record = %v, want a live one for the release", c)
		}
		if prompt := relaunchPrompt(r.conductorLaunches(t)[1]); !strings.Contains(prompt, "Write the report for the release") {
			t.Errorf("prompt lacks the release block:\n%s", prompt)
		}
		if _, err := os.Stat(worktree(r)); err != nil {
			t.Errorf("the new conductor has no worktree: %v", err)
		}
	})

	t.Run("a finish with no report relaunches twice and then stops", func(t *testing.T) {
		r := newLaunchRig(t)
		r.addRelease(t, true, "a")
		r.finishPass(t, "a")
		r.check(t, minutes(t, 1))

		for n := 1; n <= 2; n++ {
			r.finishConductor(t, minutes(t, 10*n))
			if _, err := os.Stat(worktree(r)); err != nil {
				t.Fatalf("relaunch %d: the worktree went with no report: %v", n, err)
			}
			r.check(t, minutes(t, 10*n+1))
			if got := len(r.conductorLaunches(t)); got != n+1 {
				t.Fatalf("after finish %d: %d conductor launches, want %d", n, got, n+1)
			}
			rel, _ := r.conductor(t)["relaunches"].([]any)
			if len(rel) != n || rel[n-1].(map[string]any)["counted"] != true || rel[n-1].(map[string]any)["cause"] != "no_report" {
				t.Errorf("relaunches = %v, want %d counted no_report", rel, n)
			}
			prompt := relaunchPrompt(r.conductorLaunches(t)[n])
			for _, want := range []string{"This is relaunch " + strconv.Itoa(n) + " of 2", "Write the report for round r1", "write no second one"} {
				if !strings.Contains(prompt, want) {
					t.Errorf("relaunch %d prompt lacks %q:\n%s", n, want, prompt)
				}
			}
			if c := r.conductor(t); c["for_report"] != "r1" {
				t.Errorf("relaunch %d lost for_report: %v", n, c)
			}
		}

		r.finishConductor(t, minutes(t, 40))
		got := r.check(t, minutes(t, 41))
		if n := len(r.conductorLaunches(t)); n != 3 {
			t.Errorf("%d conductor launches, want 3", n)
		}
		if c := r.conductor(t); c["stop_reason"] != "relaunch_limit" || c["ended"] == nil {
			t.Errorf("conductor record = %v, want ended with relaunch_limit", c)
		}
		if !strings.Contains(got.stdout, `"reason":"relaunch_limit"`) {
			t.Errorf("check output %q lacks the relaunch_limit stop", got.stdout)
		}
		needs := r.needs(t)
		if len(needs) != 1 || needs[0]["reason"] != "relaunch_limit" {
			t.Errorf("needs_you_now = %v, want one relaunch_limit entry", needs)
		}

		// The stop holds: a later check neither launches nor stops again.
		r.check(t, minutes(t, 50))
		if n := len(r.conductorLaunches(t)); n != 3 {
			t.Errorf("%d conductor launches after the stop, want 3", n)
		}
	})

	t.Run("a conductor that dies counts, keeps its report, and reaches the limit with a finish", func(t *testing.T) {
		r := newLaunchRig(t)
		r.addRelease(t, true, "a")
		r.finishPass(t, "a")
		r.check(t, minutes(t, 1))
		r.check(t, minutes(t, 2)) // the check sees the conductor working
		r.vanish(t, r.conductor(t)["short_id"].(string))
		r.check(t, minutes(t, 3))

		rel, _ := r.conductor(t)["relaunches"].([]any)
		if len(rel) != 1 || rel[0].(map[string]any)["counted"] != true || rel[0].(map[string]any)["cause"] != "missing" {
			t.Fatalf("relaunches = %v, want one counted missing", rel)
		}
		if r.conductor(t)["for_report"] != "r1" {
			t.Errorf("the relaunch lost for_report: %v", r.conductor(t))
		}
		launches := r.conductorLaunches(t)
		if len(launches) != 2 {
			t.Fatalf("%d conductor launches, want 2", len(launches))
		}
		prompt := relaunchPrompt(launches[1])
		for _, want := range []string{"This is relaunch 1 of 2", "Write the report for round r1", "Run `", " brief", "write no second one"} {
			if !strings.Contains(prompt, want) {
				t.Errorf("relaunch prompt lacks %q:\n%s", want, prompt)
			}
		}

		// A finish with no report is the second counted relaunch, and the
		// next one reads the limit.
		r.finishConductor(t, minutes(t, 10))
		r.check(t, minutes(t, 11))
		if n := len(r.conductorLaunches(t)); n != 3 {
			t.Fatalf("%d conductor launches after the finish, want 3", n)
		}
		r.finishConductor(t, minutes(t, 20))
		r.check(t, minutes(t, 21))
		if c := r.conductor(t); c["stop_reason"] != "relaunch_limit" {
			t.Errorf("conductor record = %v, want relaunch_limit", c)
		}
		if n := len(r.conductorLaunches(t)); n != 3 {
			t.Errorf("%d conductor launches, want 3", n)
		}
	})

	t.Run("a hand start after an ended report conductor takes no report", func(t *testing.T) {
		r := newLaunchRig(t)
		r.addRelease(t, true, "a")
		r.finishPass(t, "a")
		r.check(t, minutes(t, 1))
		requireRun(t, r.do(t, minutes(t, 2), "report", "add", "proj", input(t, "report.json"), "--round", "r1"))
		r.finishConductor(t, minutes(t, 3))
		requireRun(t, r.run(t, minutes(t, 4), "session", "start", "conductor", "proj"))
		if c := r.conductor(t); c["for_report"] != nil && c["for_report"] != "" {
			t.Errorf("a hand start inherited for_report: %v", c)
		}
		launches := r.conductorLaunches(t)
		if prompt := relaunchPrompt(launches[len(launches)-1]); strings.Contains(prompt, "Write the report for") {
			t.Errorf("a hand start's prompt holds a report block:\n%s", prompt)
		}
	})

	t.Run("the log keeps the run's other decisions beside the conductor's", func(t *testing.T) {
		r := newLaunchRig(t)
		r.addRelease(t, true, "a")
		r.finishPass(t, "a")
		r.addRound(t, "r2", "b")
		r.check(t, minutes(t, 1))
		if n := len(r.conductorLaunches(t)); n != 1 {
			t.Fatalf("%d conductor launches, want 1", n)
		}
		if n := len(r.launches(t)); n != 2 {
			t.Fatalf("%d launches, want the conductor's and pass b's", n)
		}
		outcomes := logOutcomes(t, r.state)
		launched := 0
		for _, o := range outcomes {
			if strings.HasPrefix(o, "launched ") {
				launched++
			}
		}
		if launched != 2 || len(outcomes) != 3 {
			t.Fatalf("check log = %v, want the start and two launched lines", outcomes)
		}
		if want := expectedFromCheck(t, r.state); !slices.Equal(outcomes, want) {
			t.Errorf("check log = %v, want %v", outcomes, want)
		}
	})

	t.Run("a failed first launch counts and the next check relaunches", func(t *testing.T) {
		r := newLaunchRig(t)
		r.addRelease(t, true, "a")
		r.finishPass(t, "a")
		r.control(t, "fail-bg", true)
		got := r.run(t, minutes(t, 1), "check")
		if got.code != 2 {
			t.Fatalf("check exit %d, want 2; stderr: %s", got.code, got.stderr)
		}
		c := r.conductor(t)
		if c["stop_reason"] != "launch_failed" || c["ended"] == nil || c["for_report"] != "r1" || c["short_id"] != "none" {
			t.Fatalf("conductor record = %v, want an ended launch_failed record for r1", c)
		}
		if _, err := os.Stat(worktree(r)); !os.IsNotExist(err) {
			t.Errorf("the failed launch left its worktree: %v", err)
		}
		needs := r.needs(t)
		if len(needs) != 1 || needs[0]["reason"] != "launch_failed" {
			t.Errorf("needs_you_now = %v, want one launch_failed entry", needs)
		}

		r.control(t, "fail-bg", false)
		r.check(t, minutes(t, 2))
		c = r.conductor(t)
		rel, _ := c["relaunches"].([]any)
		if c["ended"] != nil || len(rel) != 1 || rel[0].(map[string]any)["cause"] != "launch_failed" || rel[0].(map[string]any)["counted"] != true {
			t.Errorf("conductor record = %v, want a live relaunch counted after the failed launch", c)
		}
		if n := len(r.conductorLaunches(t)); n != 2 {
			t.Errorf("%d conductor launches, want the failed one and the one that worked", n)
		}
		if needs := r.needs(t); len(needs) != 0 {
			t.Errorf("needs_you_now = %v, want none once the launch worked", needs)
		}
	})
}
