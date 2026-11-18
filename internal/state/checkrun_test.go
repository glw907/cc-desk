//go:build unix

package state

import (
	"strings"
	"testing"
	"time"
)

func TestLockFile(t *testing.T) {
	dir := t.TempDir()
	unlock, err := LockFile(dir, CheckLockFile, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()

	other, err := Lock(dir, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("the state lock was refused while the check lock is held: %v", err)
	}
	other()

	_, err = LockFile(dir, CheckLockFile, 100*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "check.lock") {
		t.Fatalf("error = %v, want one that names check.lock", err)
	}
	unlock()
	again, err := LockFile(dir, CheckLockFile, time.Second)
	if err != nil {
		t.Fatalf("check lock after release: %v", err)
	}
	again()
}

const waitingPass = `,"wait":{"since":"2026-11-20T10:00:00Z","limits":["usage"]}`

func TestRecordListFailure(t *testing.T) {
	rel := strings.Replace(sessionRelease("waiting", "", ""), `"state":"waiting"`, `"state":"waiting"`+waitingPass, 1)
	dir := writeRelease(t, "p", rel)
	writeFile(t, dir, UsageFile, `{"format":1,"percent":40,"reset_text":"x","read_at":"2026-11-20T22:58:00Z"}`)
	before := readFile(t, dir, "releases/p.json")

	for i := 1; i <= 3; i++ {
		if _, err := Commit(dir, testClock, time.Second, RecordListFailure(), render); err != nil {
			t.Fatal(err)
		}
		if got := readFile(t, dir, CheckStateFile); !strings.Contains(got, `"list_failures": `+string(rune('0'+i))) {
			t.Fatalf("after failure %d check.json = %s", i, got)
		}
	}
	if readFile(t, dir, "releases/p.json") != before {
		t.Error("a failed read rewrote the release, so the waiting pass was settled")
	}
	if exists(dir, ForecastFile) {
		t.Error("a failed read appended a forecast line")
	}

	// The same state with an ordinary event starts the pass, which is what
	// the failed read must not do.
	if _, err := Commit(dir, testClock, time.Second, nil, render); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(readFile(t, dir, "releases/p.json"), `"state": "building"`) {
		t.Error("the fixture would not start the pass, so the test proves nothing")
	}
}

func TestRecordCheck(t *testing.T) {
	counts := `"launch_tries":1,"relaunches":[{"time":"2026-11-20T10:30:00Z","cause":"idle","counted":true}]`
	sess := strings.Replace(goodSession, `"launch_tries":1`, counts, 1)
	newDir := func(t *testing.T) string {
		dir := writeRelease(t, "p", sessionRelease("building", sess, ""))
		writeFile(t, dir, CheckStateFile, `{"format":1,"list_failures":2,"decisions":[]}`)
		return dir
	}
	d := Decision{Project: "p", Record: "pass:x", Session: "p-one", Decision: "none"}

	t.Run("a working session", func(t *testing.T) {
		dir := newDir(t)
		rec := []Recorded{{Decision: d, LastState: SeenWorking, Seen: true, Working: true}}
		if _, err := Commit(dir, testClock, time.Second, RecordCheck(rec), render); err != nil {
			t.Fatal(err)
		}
		s, err := Load(dir)
		if err != nil {
			t.Fatal(err)
		}
		got := s.Releases[0].Passes[0].Session
		if got.LastState != SeenWorking || got.LastSeen == nil || got.LastWorking == nil || !got.LastWorking.Equal(testClock) {
			t.Errorf("session = %+v", got)
		}
		if got.LaunchTries != 1 || len(got.Relaunches) != 1 {
			t.Errorf("the check changed launch_tries or relaunches: %+v", got)
		}
		if s.Check.ListFailures != 0 || s.Check.LastRun == nil || len(s.Check.Decisions) != 1 {
			t.Errorf("check.json = %+v, want the failures reset and one decision", s.Check)
		}
		if s.Releases[0].Passes[0].State != Building {
			t.Error("a working session stopped its pass")
		}
	})

	t.Run("a stop on a pass", func(t *testing.T) {
		dir := newDir(t)
		stop := d
		stop.Decision, stop.Reason, stop.Detail = "stop", "permission_prompt", "Bash: rm x"
		rec := []Recorded{{Decision: stop, LastState: SeenBlocked, Seen: true, Stop: StopPermissionPrompt}}
		if _, err := Commit(dir, testClock, time.Second, RecordCheck(rec), render); err != nil {
			t.Fatal(err)
		}
		s, err := Load(dir)
		if err != nil {
			t.Fatal(err)
		}
		p := s.Releases[0].Passes[0]
		if p.State != Stopped || p.StopReason != StopPermissionPrompt || p.Session.StopDetail != "Bash: rm x" {
			t.Errorf("pass = %+v session = %+v", p, p.Session)
		}
		if p.Session.StopReason != "" || p.Session.Stopped != nil {
			t.Error("a pass's stop landed on its session record")
		}
		if !exists(dir, ForecastFile) {
			t.Error("a stopped pass appended no forecast line")
		}
	})

	t.Run("a stop on a conductor", func(t *testing.T) {
		cond := strings.Replace(strings.Replace(goodSession, `"pass"`, `"conductor"`, 1), `"p-one"`, `"p-conductor"`, 1)
		cond = strings.Replace(cond, "/w/one", "/w/cond", 1)
		cond = strings.Replace(cond, `"branch":"one"`, `"branch":"cond"`, 1)
		rel := sessionRelease("not_started", "", `,"conductor":`+cond)
		dir := writeRelease(t, "p", rel)
		stop := Decision{Project: "p", Record: "conductor", Session: "p-conductor", Decision: "stop", Reason: "unreadable_stop", Detail: "last line"}
		rec := []Recorded{{Decision: stop, LastState: SeenBlocked, Seen: true, Stop: StopUnreadable}}
		if _, err := Commit(dir, testClock, time.Second, RecordCheck(rec), render); err != nil {
			t.Fatal(err)
		}
		s, err := Load(dir)
		if err != nil {
			t.Fatal(err)
		}
		c := s.Releases[0].Conductor
		if c.StopReason != StopUnreadable || c.Stopped == nil || c.StopDetail != "last line" {
			t.Errorf("conductor = %+v", c)
		}
	})

	t.Run("an unknown record is refused and writes nothing", func(t *testing.T) {
		dir := newDir(t)
		before := readFile(t, dir, CheckStateFile)
		gone := d
		gone.Record = "pass:nope"
		_, err := Commit(dir, testClock, time.Second, RecordCheck([]Recorded{{Decision: gone}}), render)
		if err == nil {
			t.Fatal("expected an error")
		}
		if readFile(t, dir, CheckStateFile) != before {
			t.Error("check.json changed")
		}
	})
}
