//go:build unix

package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// decision is one line of the check's output, in the shape of an entry in
// check.json.
type decision struct {
	Project  string `json:"project"`
	Record   string `json:"record"`
	Session  string `json:"session"`
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
	Detail   string `json:"detail"`
	Counted  *bool  `json:"counted"`
}

// rig is the stand-in claude, jobs folder, and probe address one run of the
// check gets, so no test reaches the real claude or the network.
type rig struct {
	claude, jobs, probe string
	reads               string // the log the stand-in claude adds a nanosecond timestamp to on each list read
}

// claudeStandIn is the script a stand-in claude runs. The first list read
// prints list-1.json and every later one list-2.json. A screen read prints
// screens/<id>.txt.
const claudeStandIn = `#!/bin/sh
d=%q
case "$1" in
agents)
  date +%%s%%N >> "$d/reads.log"
  n=$(wc -l < "$d/reads.log")
  if [ "$n" -le 1 ]; then cat "$d/list-1.json"; else cat "$d/list-2.json"; fi
  ;;
logs)
  cat "$d/screens/$2.txt"
  ;;
esac
`

// newRig makes a stand-in claude that reads the evidence under testdata/<fixture>.
func newRig(t *testing.T, fixture, probe string) rig {
	t.Helper()
	evidence, err := filepath.Abs(filepath.Join("testdata", fixture, "evidence"))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	// The script reads and appends under one folder, so copy the evidence
	// next to the log and leave the fixture untouched.
	copyTree(t, evidence, filepath.Join(dir, "evidence"))
	ev := filepath.Join(dir, "evidence")
	script := filepath.Join(dir, "claude.sh")
	if err := os.WriteFile(script, fmt.Appendf(nil, claudeStandIn, ev), 0o755); err != nil {
		t.Fatal(err)
	}
	return rig{claude: script, jobs: filepath.Join(ev, "jobs"), probe: probe, reads: filepath.Join(ev, "reads.log")}
}

func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}

// flags is the check verb with every stand-in flag the rig names.
func (r rig) flags(extra ...string) []string {
	args := []string{"check", "--claude-command", r.claude, "--jobs-dir", r.jobs, "--probe-url", r.probe, "--second-read-wait", "1ms"}
	return append(args, extra...)
}

// readCount is how many times the stand-in claude listed sessions.
func (r rig) readCount(t *testing.T) int {
	t.Helper()
	b, err := os.ReadFile(r.reads)
	if os.IsNotExist(err) {
		return 0
	}
	if err != nil {
		t.Fatal(err)
	}
	return len(strings.Fields(string(b)))
}

// probeServer answers the network probe. A server that is down hijacks each
// connection and closes it, so the client reads an error.
func probeServer(t *testing.T, down bool) (string, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if down {
			conn, _, err := w.(http.Hijacker).Hijack()
			if err == nil {
				conn.Close()
			}
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	return srv.URL, &calls
}

// checkRun runs the check on dir at now with the default usage stand-in.
func checkRun(t *testing.T, r rig, dir, now string, extra ...string) result {
	t.Helper()
	return run(t, dir, now, r.flags(extra...)...)
}

// decisions parses the lines the check printed, keyed by the record.
func decisions(t *testing.T, out string) map[string]decision {
	t.Helper()
	got := map[string]decision{}
	for line := range strings.SplitSeq(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		var d decision
		if err := json.Unmarshal([]byte(line), &d); err != nil {
			t.Fatalf("output line %q: %v", line, err)
		}
		got[d.Record] = d
	}
	return got
}

func requireEqual(t *testing.T, name string, got, want any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

// requireRun fails the test unless the run exited 0.
func requireRun(t *testing.T, got result) {
	t.Helper()
	if got.code != 0 {
		t.Fatalf("exit %d; stderr: %s", got.code, got.stderr)
	}
}

func TestCheckStops(t *testing.T) {
	// Run A reads the lost connection with the probe failing.
	urlA, callsA := probeServer(t, true)
	dirA := newStateFrom(t, "check")
	rigA := newRig(t, "check", urlA)
	runA := checkRun(t, rigA, dirA, workedClock)
	requireRun(t, runA)
	a := decisions(t, runA.stdout)

	// Run B reads it with the probe answering.
	urlB, callsB := probeServer(t, false)
	dirB := newStateFrom(t, "check")
	rigB := newRig(t, "check", urlB)
	runB := checkRun(t, rigB, dirB, workedClock)
	requireRun(t, runB)
	b := decisions(t, runB.stdout)

	if len(a) != 9 || len(b) != 9 {
		t.Fatalf("decisions: run A %d, run B %d, want 9 each", len(a), len(b))
	}

	stop := func(t *testing.T, d decision, reason string) {
		t.Helper()
		if d.Decision != "stop" || d.Reason != reason || d.Counted != nil {
			t.Errorf("decision = %+v, want stop with reason %s", d, reason)
		}
	}

	// stopped checks the snapshot's needs_you_now entry for a session and
	// the pass state the release file holds.
	stopped := func(t *testing.T, dir, pass, session, reason, todo string) need {
		t.Helper()
		n, ok := snapshotNeeds(t, dir)[session]
		if !ok {
			t.Fatalf("needs_you_now names no %s", session)
		}
		if n.Reason != reason || n.Todo != todo {
			t.Errorf("%s: reason %q, todo %q, want %q and %q", session, n.Reason, n.Todo, reason, todo)
		}
		if got := passState(t, dir, pass); got != "stopped" {
			t.Errorf("%s reads %s, want stopped", pass, got)
		}
		return n
	}

	t.Run("1 a permission prompt", func(t *testing.T) {
		stop(t, b["pass:s01"], "permission_prompt")
		stopped(t, dirB, "s01", "chk-s01", "permission_prompt", "answer it at the desk")
	})
	t.Run("2 login expired", func(t *testing.T) {
		stop(t, b["pass:s02"], "sign_in_expired")
		stopped(t, dirB, "s02", "chk-s02", "sign_in_expired", "sign in to Claude Code")
	})
	t.Run("3 a usage limit", func(t *testing.T) {
		d := b["pass:s03"]
		stop(t, d, "usage")
		if !strings.Contains(d.Detail, "You've hit your weekly limit") || !strings.Contains(d.Detail, "resets Nov 6, 11pm (America/Anchorage)") {
			t.Errorf("detail = %q, want the limit and its reset time", d.Detail)
		}
		n := stopped(t, dirB, "s03", "chk-s03", "usage", "wait for the reset, or reset early")
		if n.Cause != d.Detail {
			t.Errorf("cause = %q, want %q", n.Cause, d.Detail)
		}
	})
	t.Run("4 an unreadable block", func(t *testing.T) {
		d := b["pass:s04"]
		stop(t, d, "unreadable_stop")
		if d.Detail != unreadableCause {
			t.Errorf("detail = %q, want the last 200 runes of the screen, %q", d.Detail, unreadableCause)
		}
		n := stopped(t, dirB, "s04", "chk-s04", "unreadable_stop", "read the session's log, then start the pass again")
		if n.Cause != unreadableCause {
			t.Errorf("cause = %q", n.Cause)
		}
	})
	t.Run("5 lost connection, probe failing", func(t *testing.T) {
		for _, rec := range []string{"pass:s05", "pass:s06"} {
			if d := a[rec]; d.Decision != "wait_link" || d.Counted != nil {
				t.Errorf("%s = %+v, want wait_link with no count", rec, d)
			}
		}
		if n := callsA.Load(); n != 1 {
			t.Errorf("probe calls = %d, want 1", n)
		}
		if got := readFileIn(t, dirA, "releases/chk.json"); strings.Contains(got, `"relaunches"`) {
			t.Error("a waiting link counted a relaunch")
		}
		for _, p := range []string{"s05", "s06"} {
			if got := passState(t, dirA, p); got != "building" {
				t.Errorf("%s reads %s, want building", p, got)
			}
		}
	})
	t.Run("6 lost connection, probe answering", func(t *testing.T) {
		for _, rec := range []string{"pass:s05", "pass:s06"} {
			d := b[rec]
			if d.Decision != "end_and_relaunch" || d.Reason != "lost_connection" || d.Counted == nil || !*d.Counted {
				t.Errorf("%s = %+v, want a counted end_and_relaunch for a lost connection", rec, d)
			}
		}
		if n := callsB.Load(); n != 1 {
			t.Errorf("probe calls = %d, want 1 although two sessions read lost", n)
		}
	})
	t.Run("7 done and idle for 16 minutes", func(t *testing.T) {
		d := b["pass:s07"]
		if d.Decision != "end_and_relaunch" || d.Reason != "idle" {
			t.Errorf("decision = %+v", d)
		}
	})
	t.Run("8 done and idle for 14 minutes", func(t *testing.T) {
		if d := b["pass:s08"]; d.Decision != "none" {
			t.Errorf("decision = %+v", d)
		}
	})
	t.Run("9 idle with a ready report", func(t *testing.T) {
		if d := b["pass:s09"]; d.Decision != "none" {
			t.Errorf("decision = %+v", d)
		}
	})

	// A usage-limit stop relaunches once the hold lifts, and the relaunch
	// does not count. Each run needing a fresh reading calls its stand-in once.
	t.Run("a usage-limit relaunch", func(t *testing.T) {
		hold := newStandIn(t, 0, usageResult(100))
		first := runWith(t, hold.path, dirB, "2026-11-20T23:05:00Z", rigB.flags()...)
		requireRun(t, first)
		held := decisions(t, first.stdout)
		if hold.calls(t) != 1 {
			t.Errorf("usage stand-in calls = %d, want 1", hold.calls(t))
		}
		if d := held["pass:s03"]; d.Decision != "wait_usage" || d.Reason != "usage" {
			t.Errorf("s03 at 100 percent = %+v, want wait_usage", d)
		}
		if d := held["pass:s04"]; d.Decision != "none" || d.Reason != "unreadable_stop" {
			t.Errorf("s04 at 100 percent = %+v, want its cause to stand", d)
		}
		held3, held4 := snapshotNeeds(t, dirB)["chk-s03"], snapshotNeeds(t, dirB)["chk-s04"]
		if held3.Reason != "usage" || !strings.Contains(held3.Cause, "You've hit your weekly limit") {
			t.Errorf("needs_you_now for s03 = %+v, want its cause to stand", held3)
		}
		if held4.Reason != "unreadable_stop" || held4.Cause != unreadableCause {
			t.Errorf("needs_you_now for s04 = %+v, want its cause to stand", held4)
		}

		lifted := newStandIn(t, 0, usageResult(40))
		second := runWith(t, lifted.path, dirB, "2026-11-20T23:10:00Z", rigB.flags()...)
		requireRun(t, second)
		cleared := decisions(t, second.stdout)
		if lifted.calls(t) != 1 {
			t.Errorf("usage stand-in calls = %d, want 1", lifted.calls(t))
		}
		d := cleared["pass:s03"]
		if d.Decision != "relaunch" || d.Reason != "usage" || d.Counted == nil || *d.Counted {
			t.Errorf("s03 at 40 percent = %+v, want an uncounted relaunch", d)
		}
		if got := readFileIn(t, dirB, "releases/chk.json"); strings.Contains(got, `"relaunches"`) {
			t.Error("a relaunch decision wrote to the record's relaunches")
		}
	})
	launchClauses(t)
}

// unreadableCause is the last 200 runes of the screen in
// evidence/screens/c0000004.txt, a copy of spikes/2-states/out/logs-idle-1.txt.
// It reads "used", so the usage-limit pattern does not match it.
const unreadableCause = "ng… 3 Brewing… ✶ * ✢ Brewing… · Brewing… Brewing… Brewing… Brewing… ✢ Brewing… 4 ✻ Crunched for 4s · done 6:34 PM ❯ You've used 87% of your weekly limit · resets Nov 20, 11pm (America/Anchorage) 2 s 2"

// need is one entry of the snapshot's needs_you_now.
type need struct {
	Session *string `json:"session"`
	Reason  string  `json:"reason"`
	Cause   string  `json:"cause"`
	Todo    string  `json:"todo"`
}

// snapshotNeeds reads needs_you_now from the folder's snapshot, by session.
func snapshotNeeds(t *testing.T, dir string) map[string]need {
	t.Helper()
	var snap struct {
		NeedsYouNow []need `json:"needs_you_now"`
	}
	if err := json.Unmarshal([]byte(readFileIn(t, dir, "snapshot.json")), &snap); err != nil {
		t.Fatal(err)
	}
	got := map[string]need{}
	for _, n := range snap.NeedsYouNow {
		if n.Session != nil {
			got[*n.Session] = n
		}
	}
	return got
}

// passState is the state the release file holds for a pass of chk.
func passState(t *testing.T, dir, pass string) string {
	t.Helper()
	var rel struct {
		Passes []struct {
			Name  string `json:"name"`
			State string `json:"state"`
		} `json:"passes"`
	}
	if err := json.Unmarshal([]byte(readFileIn(t, dir, "releases/chk.json")), &rel); err != nil {
		t.Fatal(err)
	}
	for _, p := range rel.Passes {
		if p.Name == pass {
			return p.State
		}
	}
	t.Fatalf("no pass %s", pass)
	return ""
}

func readFileIn(t *testing.T, dir, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, rel))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// launchClauses runs clauses 10 and 11 on the check-launch fixture, as
// subtests of TestCheckStops.
// TestCheckSecondReadWait runs the launch folder with the wait set to 200
// milliseconds and reads the gap between the stand-in's two list calls.
func TestCheckSecondReadWait(t *testing.T) {
	url, _ := probeServer(t, false)
	dir := newStateFrom(t, "check-launch")
	r := newRig(t, "check-launch", url)
	requireRun(t, checkRun(t, r, dir, workedClock, "--second-read-wait", "200ms"))
	b, err := os.ReadFile(r.reads)
	if err != nil {
		t.Fatal(err)
	}
	stamps := strings.Fields(string(b))
	if len(stamps) != 2 {
		t.Fatalf("list reads = %d, want 2", len(stamps))
	}
	first, err1 := strconv.ParseInt(stamps[0], 10, 64)
	second, err2 := strconv.ParseInt(stamps[1], 10, 64)
	if err1 != nil || err2 != nil {
		t.Fatalf("read log holds %q", stamps)
	}
	if gap := time.Duration(second - first); gap < 200*time.Millisecond {
		t.Errorf("gap between the two list reads = %v, want at least 200ms", gap)
	}
}

func launchClauses(t *testing.T) {
	url, _ := probeServer(t, false)
	dir := newStateFrom(t, "check-launch")
	r := newRig(t, "check-launch", url)
	got := checkRun(t, r, dir, workedClock)
	requireRun(t, got)
	d := decisions(t, got.stdout)

	t.Run("10 no row after one launch", func(t *testing.T) {
		if d["pass:s10"].Decision != "launch_again" {
			t.Errorf("decision = %+v", d["pass:s10"])
		}
		if !strings.Contains(readFileIn(t, dir, "releases/chk.json"), `"launch_tries": 1`) {
			t.Error("launch_tries changed")
		}
		if n := r.readCount(t); n != 2 {
			t.Errorf("list reads = %d, want 2", n)
		}
	})
	t.Run("11 no row after two launches", func(t *testing.T) {
		x := d["pass:s11"]
		if x.Decision != "stop" || x.Reason != "launch_failed" {
			t.Errorf("decision = %+v", x)
		}
		snap := readFileIn(t, dir, "snapshot.json")
		var parsed struct {
			NeedsYouNow []struct {
				Session *string `json:"session"`
				Reason  string  `json:"reason"`
			} `json:"needs_you_now"`
		}
		if err := json.Unmarshal([]byte(snap), &parsed); err != nil {
			t.Fatal(err)
		}
		found := false
		for _, n := range parsed.NeedsYouNow {
			if n.Session != nil && *n.Session == "chk-s11" && n.Reason == "launch_failed" {
				found = true
			}
		}
		if !found {
			t.Errorf("needs_you_now = %+v, want chk-s11 with launch_failed", parsed.NeedsYouNow)
		}
		if !strings.Contains(readFileIn(t, dir, "releases/chk.json"), `"launch_tries": 2`) {
			t.Error("launch_tries changed")
		}
	})
}

func TestCheckFailedReads(t *testing.T) {
	dir := newStateFrom(t, "check-fail")
	failing := filepath.Join(t.TempDir(), "claude.sh")
	if err := os.WriteFile(failing, []byte("#!/bin/sh\necho 'agents: not signed in' >&2\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	bad := rig{claude: failing, jobs: t.TempDir(), probe: "http://127.0.0.1:1/"}
	release := readFileIn(t, dir, "releases/chk.json")

	for i := 1; i <= 3; i++ {
		got := checkRun(t, bad, dir, workedClock)
		if got.code != 1 {
			t.Fatalf("read %d: exit %d, want 1; stderr: %s", i, got.code, got.stderr)
		}
		if readFileIn(t, dir, "releases/chk.json") != release {
			t.Fatalf("read %d rewrote the release", i)
		}
		if _, err := os.Stat(filepath.Join(dir, "forecasts.jsonl")); err == nil {
			t.Fatalf("read %d appended a forecast line", i)
		}
		snap := readFileIn(t, dir, "snapshot.json")
		got3 := strings.Contains(snap, `"session_list_unreadable"`)
		if (i == 3) != got3 {
			t.Fatalf("read %d: session_list_unreadable present = %v", i, got3)
		}
		if !strings.Contains(readFileIn(t, dir, "check.json"), fmt.Sprintf(`"list_failures": %d`, i)) {
			t.Fatalf("read %d: check.json = %s", i, readFileIn(t, dir, "check.json"))
		}
	}

	url, _ := probeServer(t, false)
	good := newRig(t, "check-fail", url)
	requireRun(t, checkRun(t, good, dir, workedClock))
	if !strings.Contains(readFileIn(t, dir, "check.json"), `"list_failures": 0`) {
		t.Errorf("a good read left check.json = %s", readFileIn(t, dir, "check.json"))
	}
	if strings.Contains(readFileIn(t, dir, "snapshot.json"), `"session_list_unreadable"`) {
		t.Error("session_list_unreadable survived a good read")
	}
}

func TestCheckLock(t *testing.T) {
	dir := newStateFrom(t, "check-launch")
	r := newRig(t, "check-launch", "http://127.0.0.1:1/")
	before := folder(t, dir)

	f, err := os.OpenFile(filepath.Join(dir, "check.lock"), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	got := checkRun(t, r, dir, workedClock)
	elapsed := time.Since(start)
	if got.code != 1 {
		t.Fatalf("exit %d, want 1; stderr: %s", got.code, got.stderr)
	}
	if elapsed < 10*time.Second || elapsed >= 15*time.Second {
		t.Errorf("waited %v, want at least 10 and under 15 seconds", elapsed)
	}
	if !strings.Contains(got.stderr, "check.lock") {
		t.Errorf("stderr does not name check.lock: %s", got.stderr)
	}
	sameFolder(t, before, folder(t, dir))
	if r.readCount(t) != 0 {
		t.Error("the check read the session list without the lock")
	}
}

func TestCheckDryRun(t *testing.T) {
	url, _ := probeServer(t, false)
	dir := newStateFrom(t, "check")
	r := newRig(t, "check", url)
	before := folder(t, dir)

	dry := checkRun(t, r, dir, workedClock, "--dry-run")
	requireRun(t, dry)
	sameFolder(t, before, folder(t, dir))

	realRun := checkRun(t, r, dir, workedClock)
	requireRun(t, realRun)
	var file struct {
		Decisions []decision `json:"decisions"`
	}
	if err := json.Unmarshal([]byte(readFileIn(t, dir, "check.json")), &file); err != nil {
		t.Fatal(err)
	}
	var printed []decision
	for line := range strings.SplitSeq(strings.TrimSpace(dry.stdout), "\n") {
		var d decision
		if err := json.Unmarshal([]byte(line), &d); err != nil {
			t.Fatalf("dry-run line %q: %v", line, err)
		}
		printed = append(printed, d)
	}
	requireEqual(t, "the dry run's decisions", printed, file.Decisions)
	if len(printed) != 9 {
		t.Errorf("printed %d decisions, want 9", len(printed))
	}
}
