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
	"time"
)

// minutes gives the clock the given number of minutes after workedClock.
func minutes(t *testing.T, minutes int) string {
	t.Helper()
	base, err := time.Parse(time.RFC3339, workedClock)
	if err != nil {
		t.Fatal(err)
	}
	return base.Add(time.Duration(minutes) * time.Minute).Format(time.RFC3339)
}

// needs reads the needs_you_now entries of the state's snapshot.
func (r launchRig) needs(t *testing.T) []map[string]any {
	t.Helper()
	var snap struct {
		Needs []map[string]any `json:"needs_you_now"`
	}
	if err := json.Unmarshal([]byte(readFileIn(t, r.state, "snapshot.json")), &snap); err != nil {
		t.Fatal(err)
	}
	return snap.Needs
}

// conductor gives the project's conductor record.
func (r launchRig) conductor(t *testing.T) map[string]any {
	t.Helper()
	c, ok := r.release(t)["conductor"].(map[string]any)
	if !ok {
		t.Fatal("the release has no conductor record")
	}
	return c
}

// relaunchPrompt gives the prompt argument of a launch call.
func relaunchPrompt(c call) string {
	for _, a := range c.Args {
		if strings.Contains(a, "\n") {
			return a
		}
	}
	return ""
}

// reads gives how many list reads the stand-in has served.
func (r launchRig) reads(t *testing.T) int {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(r.dir, "reads"))
	if os.IsNotExist(err) {
		return 0
	}
	if err != nil {
		t.Fatal(err)
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// check runs one check and requires exit 0.
func (r launchRig) check(t *testing.T, now string) result {
	t.Helper()
	got := r.run(t, now, "check")
	requireRun(t, got)
	return got
}

func TestRecoveryConductorRowVanishes(t *testing.T) {
	r := newLaunchRig(t)
	r.addRelease(t, true, "a")
	requireRun(t, r.run(t, workedClock, "session", "start", "conductor", "proj"))
	first := r.conductor(t)["short_id"].(string)
	r.vanish(t, first)

	got := r.check(t, minutes(t, 1))
	if n := len(r.launches(t)); n != 2 {
		t.Fatalf("%d launches, want 2 (the start and the check's launch again)", n)
	}
	if slices.Contains(r.kinds(t), "stop") {
		t.Errorf("claude stop ran for a row that was never there: %v", r.kinds(t))
	}
	if second := r.conductor(t)["short_id"]; second == first {
		t.Errorf("the conductor kept short id %s", first)
	}
	if n := r.needs(t); len(n) != 0 {
		t.Errorf("needs_you_now = %v, want none", n)
	}
	if !strings.Contains(got.stdout, `"decision":"launch_again"`) {
		t.Errorf("check output %q lacks launch_again", got.stdout)
	}
}

func TestRecoveryFinishedConductorStays(t *testing.T) {
	r := newLaunchRig(t)
	r.addRelease(t, true, "a")
	requireRun(t, r.run(t, workedClock, "session", "start", "conductor", "proj"))
	requireRun(t, r.do(t, workedClock, "session", "finish", "conductor", "proj"))
	r.vanish(t, r.conductor(t)["short_id"].(string))

	for i, m := range []int{1, 2} {
		got := r.check(t, minutes(t, m))
		if n := len(r.launches(t)); n != 1 {
			t.Fatalf("check %d: %d launches, want only the first", i+1, n)
		}
		if i == 0 && !strings.Contains(got.stdout, `"decision":"end"`) {
			t.Errorf("first check output %q lacks end", got.stdout)
		}
	}
	if r.conductor(t)["ended"] == nil {
		t.Errorf("conductor record not ended: %v", r.conductor(t))
	}

	// An ended record is replaced by a fresh start in its own worktree.
	requireRun(t, r.run(t, minutes(t, 3), "session", "start", "conductor", "proj"))
	if n := len(r.launches(t)); n != 2 {
		t.Errorf("%d launches after restarting the ended conductor, want 2", n)
	}
}

func TestRecoveryPassRowVanishesThreeTimes(t *testing.T) {
	r := newLaunchRig(t)
	r.addRelease(t, true, "a")
	requireRun(t, r.do(t, workedClock, "pass", "start", "proj", "a"))
	worktree := filepath.Join(r.clone, ".claude", "worktrees", "a")

	minute := 0
	next := func() string { minute++; return minutes(t, minute) }

	r.check(t, next()) // launch
	commitFile(t, worktree, "work.txt", "work\n", "Do some work")
	tip := git(t, worktree, "rev-parse", "HEAD")
	r.check(t, next()) // working

	for n := 1; n <= 2; n++ {
		r.vanish(t, r.sessionOf(t, "a")["short_id"].(string))
		r.check(t, next())
		if got := len(r.launches(t)); got != n+1 {
			t.Fatalf("after vanish %d: %d launches, want %d", n, got, n+1)
		}
		if head := git(t, worktree, "rev-parse", "HEAD"); head != tip {
			t.Errorf("after relaunch %d the branch tip is %s, want %s", n, head, tip)
		}
		if cwd := r.launches(t)[n].Cwd; cwd != worktree {
			t.Errorf("relaunch %d ran in %s, want %s", n, cwd, worktree)
		}
		rel, _ := r.sessionOf(t, "a")["relaunches"].([]any)
		if len(rel) != n {
			t.Fatalf("after vanish %d the record holds %d relaunches: %v", n, len(rel), rel)
		}
		last := rel[n-1].(map[string]any)
		if last["cause"] == "" || last["counted"] != true {
			t.Errorf("relaunch %d = %v, want a cause and counted", n, last)
		}
		want := "This is relaunch " + strconv.Itoa(n) + " of 2"
		if p := relaunchPrompt(r.launches(t)[n]); !strings.Contains(p, want) {
			t.Errorf("relaunch %d prompt lacks %q:\n%s", n, want, p)
		}
		r.check(t, next()) // the new row works
	}

	r.vanish(t, r.sessionOf(t, "a")["short_id"].(string))
	got := r.check(t, next())
	if n := len(r.launches(t)); n != 3 {
		t.Errorf("%d launches after the third vanish, want 3", n)
	}
	if p := r.passRecord(t, "a"); p["state"] != "stopped" || p["stop_reason"] != "relaunch_limit" {
		t.Errorf("pass = %v, want stopped with relaunch_limit", p)
	}
	needs := r.needs(t)
	if len(needs) != 1 || needs[0]["round"] != "r1" || needs[0]["reason"] != "relaunch_limit" {
		t.Errorf("needs_you_now = %v, want one relaunch_limit entry naming r1", needs)
	}
	if !strings.Contains(got.stdout, `"reason":"relaunch_limit"`) {
		t.Errorf("check output %q lacks the relaunch_limit stop", got.stdout)
	}
}

// idlePass launches pass a, sees it working, then turns its row idle.
func idlePass(t *testing.T, r launchRig) {
	t.Helper()
	r.addRelease(t, true, "a")
	requireRun(t, r.do(t, workedClock, "pass", "start", "proj", "a"))
	r.check(t, workedClock)
	r.check(t, minutes(t, 1))
	rows := r.readRows(t)
	rows[0]["state"], rows[0]["status"] = "done", "idle"
	r.setRows(t, rows...)
}

func TestRecoveryStopConfirmLaunch(t *testing.T) {
	r := newLaunchRig(t)
	idlePass(t, r)
	old := r.sessionOf(t, "a")["short_id"].(string)
	before := len(r.calls(t))

	r.check(t, minutes(t, 30))
	got := r.calls(t)[before:]
	var kinds []string
	for _, c := range got {
		kinds = append(kinds, c.Args[0])
	}
	stop := slices.Index(kinds, "stop")
	if stop < 0 || got[stop].Args[1] != old {
		t.Fatalf("calls %v: want claude stop %s", kinds, old)
	}
	if stop+2 >= len(kinds) || kinds[stop+1] != "agents" || !slices.Contains(kinds[stop+2:], "--bg") {
		t.Errorf("calls %v: want stop, then a list read, then the launch", kinds)
	}
	if n := len(r.launches(t)); n != 2 {
		t.Errorf("%d launches, want 2", n)
	}
}

func TestRecoveryStopFailsNoLaunch(t *testing.T) {
	r := newLaunchRig(t)
	idlePass(t, r)
	r.control(t, "stop-fails", true)

	got := r.run(t, minutes(t, 30), "check")
	if got.code != 2 {
		t.Errorf("check exit %d, want 2; stderr: %s", got.code, got.stderr)
	}
	if n := len(r.launches(t)); n != 1 {
		t.Errorf("%d launches after a failed stop, want only the first", n)
	}
}

func TestRecoveryRowBackAtSecondRead(t *testing.T) {
	r := newLaunchRig(t)
	r.addRelease(t, true, "a")
	requireRun(t, r.do(t, workedClock, "pass", "start", "proj", "a"))
	r.check(t, workedClock)
	r.check(t, minutes(t, 1))
	rows := r.readRows(t)

	n := r.reads(t)
	r.setList(t, n+1)
	r.setList(t, n+2, rows...)
	r.check(t, minutes(t, 2))

	if got := len(r.launches(t)); got != 1 {
		t.Errorf("%d launches, want the row left alone", got)
	}
	if rel, _ := r.sessionOf(t, "a")["relaunches"].([]any); len(rel) != 0 {
		t.Errorf("relaunches = %v, want none", rel)
	}
	if slices.Contains(r.kinds(t), "stop") {
		t.Errorf("claude stop ran: %v", r.kinds(t))
	}
}

func TestRecoveryThreePassesAfterRowsVanish(t *testing.T) {
	r := newLaunchRig(t)
	r.addRelease(t, true, "a", "b", "c")
	for _, v := range [][]string{
		{"start", "proj", "a"}, {"start", "proj", "b"}, {"stop", "proj", "a", "--reason", "scope"},
		{"start", "proj", "c"}, {"start", "proj", "a"},
	} {
		requireRun(t, r.do(t, workedClock, append([]string{"pass"}, v...)...))
	}
	for _, name := range []string{"a", "b", "c"} {
		if s := r.passRecord(t, name)["state"]; s != "building" {
			t.Fatalf("pass %s is %v, want building", name, s)
		}
	}

	r.check(t, minutes(t, 1)) // launches each
	if n := len(r.launches(t)); n != 3 {
		t.Fatalf("%d launches, want 3", n)
	}
	r.check(t, minutes(t, 2)) // sees all three working
	r.setRows(t)

	usage := newStandIn(t, 0, usageResult(100))
	held := runWith(t, usage.path, r.state, minutes(t, 70), r.args(t, "check")...)
	requireRun(t, held)
	if n := len(r.launches(t)); n != 3 {
		t.Errorf("%d launches while usage reads at the stop line, want 3", n)
	}
	if !strings.Contains(held.stdout, `"decision":"wait_usage"`) {
		t.Errorf("held check output %q lacks wait_usage", held.stdout)
	}

	usage.set(t, usageResult(40))
	freed := runWith(t, usage.path, r.state, minutes(t, 140), r.args(t, "check")...)
	requireRun(t, freed)
	if n := len(r.launches(t)); n != 6 {
		t.Errorf("%d launches after usage fell, want 6: one relaunch each", n)
	}
	for _, name := range []string{"a", "b", "c"} {
		if rel, _ := r.sessionOf(t, name)["relaunches"].([]any); len(rel) != 1 {
			t.Errorf("pass %s relaunches = %v, want one", name, rel)
		}
	}
}
