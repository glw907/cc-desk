package state

import (
	"slices"
	"strings"
	"testing"
	"time"
)

func logLines(t *testing.T, dir string) []CheckLogLine {
	t.Helper()
	s, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	return s.CheckLog
}

func TestCheckLogStopsThenOutcomes(t *testing.T) {
	dir, _ := loadDir(t, sessionRelease("building", goodSession, ""))
	d := Decision{Project: "p", Record: "pass:x", Session: "p-x", Decision: "stop", Reason: "usage", Detail: "at the limit"}
	recs := []Recorded{{Decision: d, Stop: StopUsage}}
	if _, err := Commit(dir, testClock, time.Second, RecordCheck(recs), render); err != nil {
		t.Fatal(err)
	}
	got := logLines(t, dir)
	if len(got) != 1 {
		t.Fatalf("check log after RecordCheck = %+v, want the one stop line", got)
	}
	l := got[0]
	if l.Outcome != "stopped: usage" || l.Decision != "stop" || l.Reason != "usage" || l.Detail != "at the limit" || l.Record != "pass:x" {
		t.Errorf("stop line = %+v", l)
	}

	if _, err := Commit(dir, testClock, time.Second, RecordOutcomes([]string{"stopped: usage"}), render); err != nil {
		t.Fatal(err)
	}
	got = logLines(t, dir)
	if len(got) != 2 || got[1].Outcome != "stopped: usage" {
		t.Errorf("check log after RecordOutcomes = %+v, want the stop line then the outcome line", got)
	}
}

func TestCheckLogOutcomeKinds(t *testing.T) {
	tests := []struct {
		name    string
		outcome string
		reason  string
		want    int
	}{
		{"failed", "failed: no tmux", "dropped connection", 1},
		{"not relaunched", "not relaunched: two tries used", "", 1},
		{"no checkout", "no checkout: the release has none", "", 1},
		{"nothing carried out", "", "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, _ := loadDir(t, sessionRelease("building", goodSession, ""))
			d := Decision{Project: "p", Record: "pass:x", Session: "p-x", Decision: "relaunch", Reason: tt.reason}
			if _, err := Commit(dir, testClock, time.Second, RecordCheck([]Recorded{{Decision: d}}), render); err != nil {
				t.Fatal(err)
			}
			if _, err := Commit(dir, testClock, time.Second, RecordOutcomes([]string{tt.outcome}), render); err != nil {
				t.Fatal(err)
			}
			got := logLines(t, dir)
			if len(got) != tt.want {
				t.Fatalf("check log = %+v, want %d lines", got, tt.want)
			}
			if tt.want == 1 && (got[0].Outcome != tt.outcome || got[0].Reason != tt.reason) {
				t.Errorf("line = %+v, want outcome %q and reason %q", got[0], tt.outcome, tt.reason)
			}
		})
	}
}

func TestCheckLogRejectsBadLine(t *testing.T) {
	dir := writeRelease(t, "p", sessionRelease("building", goodSession, ""))
	writeFile(t, dir, CheckLogFile, `{"format":1,"time":"2026-11-20T10:00:00Z","project":"p","record":"pass:x","decision":"launch"}`+"\n")
	if _, err := Load(dir); err == nil || !strings.Contains(err.Error(), "check-log.jsonl") {
		t.Errorf("Load of a line with no outcome: %v", err)
	}
}

func TestCheckLogAppendsAcrossRuns(t *testing.T) {
	dir, _ := loadDir(t, sessionRelease("building", goodSession, ""))
	d := Decision{Project: "p", Record: "pass:x", Session: "p-x", Decision: "launch"}
	var outcomes []string
	for i := range 3 {
		if _, err := Commit(dir, testClock, time.Second, RecordCheck([]Recorded{{Decision: d}}), render); err != nil {
			t.Fatal(err)
		}
		o := "launched " + string(rune('a'+i))
		outcomes = append(outcomes, o)
		if _, err := Commit(dir, testClock, time.Second, RecordOutcomes([]string{o}), render); err != nil {
			t.Fatal(err)
		}
	}
	var got []string
	for _, l := range logLines(t, dir) {
		got = append(got, l.Outcome)
	}
	if !slices.Equal(got, outcomes) {
		t.Errorf("log outcomes = %v, want %v", got, outcomes)
	}
}
