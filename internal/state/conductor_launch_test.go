package state

import (
	"testing"
	"time"
)

func conductorSession(report string) Session {
	return Session{
		Name: "p-conductor", Kind: KindConductor, Model: "sonnet", Effort: "medium",
		ShortID: "cd01", SessionID: "s-cd01", Worktree: "/w/conductor", Branch: "conductor", ForReport: report,
	}
}

func TestConductorLaunchRecords(t *testing.T) {
	dir, _ := loadDir(t, sessionRelease("not_started", "", ""))
	commit := func(ev Event) error {
		t.Helper()
		_, err := Commit(dir, testClock, time.Second, ev, render)
		return err
	}
	conductor := func() *Session {
		t.Helper()
		s, err := Load(dir)
		if err != nil {
			t.Fatal(err)
		}
		return s.Releases[0].Conductor
	}

	// A failed first launch leaves an ended record that names the report.
	failed := Launch{Kind: LaunchFirst, Session: conductorSession("r1")}
	if err := commit(RecordConductorLaunchFailure("p", failed, "claude: could not start")); err != nil {
		t.Fatal(err)
	}
	c := conductor()
	if c == nil || c.Ended == nil || c.StopReason != StopLaunchFailed || c.StopDetail != "claude: could not start" || c.ForReport != "r1" || c.ShortID != "none" || c.Counted() != 0 {
		t.Fatalf("record after a failed first launch = %+v", c)
	}

	// A continuation replaces the ended record, counts, and keeps the report.
	next := conductorSession("")
	next.ShortID = "cd02"
	if err := commit(RecordLaunch("p", "conductor", Launch{Kind: LaunchContinue, Replaces: "none", Cause: "launch_failed", Counted: true, Session: next})); err != nil {
		t.Fatal(err)
	}
	c = conductor()
	if c.Ended != nil || c.ForReport != "r1" || c.Counted() != 1 || c.LaunchTries != 1 || c.StopReason != "" {
		t.Fatalf("record after a continuation = %+v", c)
	}

	// A relaunch of the live record carries the report too.
	again := conductorSession("")
	again.ShortID = "cd03"
	if err := commit(RecordLaunch("p", "conductor", Launch{Kind: LaunchRelaunch, Replaces: "cd02", Cause: "missing", Counted: true, Session: again})); err != nil {
		t.Fatal(err)
	}
	if c = conductor(); c.ForReport != "r1" || c.Counted() != 2 {
		t.Fatalf("record after a relaunch = %+v", c)
	}

	// A continuation needs an ended record.
	live := conductorSession("")
	live.ShortID = "cd04"
	if err := commit(RecordLaunch("p", "conductor", Launch{Kind: LaunchContinue, Replaces: "cd03", Session: live})); err == nil {
		t.Error("a continuation of a live record was accepted")
	}
	if err := commit(RecordConductorLaunchFailure("p", failed, "x")); err == nil {
		t.Error("a failed launch wrote over a live record")
	}
}
