//go:build unix

package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// loggedRun is the part of check.json the log must agree with.
type loggedRun struct {
	Decisions []struct {
		Project  string `json:"project"`
		Record   string `json:"record"`
		Decision string `json:"decision"`
		Outcome  string `json:"outcome"`
		Reason   string `json:"reason"`
	} `json:"decisions"`
}

// logOutcomes returns the outcomes of check-log.jsonl lines in order, each
// line decoded strictly enough to fail on a missing field.
func logOutcomes(t *testing.T, dir string) []string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, "check-log.jsonl"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for line := range strings.SplitSeq(strings.TrimSpace(string(b)), "\n") {
		var l struct {
			Format   int    `json:"format"`
			Time     string `json:"time"`
			Project  string `json:"project"`
			Record   string `json:"record"`
			Decision string `json:"decision"`
			Outcome  string `json:"outcome"`
		}
		if err := json.Unmarshal([]byte(line), &l); err != nil {
			t.Fatalf("check-log.jsonl line %q: %v", line, err)
		}
		if l.Format != 1 || l.Time == "" || l.Project == "" || l.Record == "" || l.Decision == "" || l.Outcome == "" {
			t.Errorf("check-log.jsonl line %q is missing a field", line)
		}
		out = append(out, l.Outcome)
	}
	return out
}

// expectedFromCheck gives what one run adds to the log: a stop line for each
// decision that stopped a record, then each non-empty outcome, in decision
// order.
func expectedFromCheck(t *testing.T, dir string) []string {
	t.Helper()
	var c loggedRun
	if err := json.Unmarshal([]byte(readFileIn(t, dir, "check.json")), &c); err != nil {
		t.Fatal(err)
	}
	var stops, outcomes []string
	for _, d := range c.Decisions {
		if d.Decision == "stop" {
			stops = append(stops, "stopped: "+d.Reason)
		}
		if d.Outcome != "" {
			outcomes = append(outcomes, d.Outcome)
		}
	}
	return append(stops, outcomes...)
}

func TestCheckLog(t *testing.T) {
	r := newLaunchRig(t)
	r.addRelease(t, true, "a")
	requireRun(t, r.do(t, workedClock, "pass", "start", "proj", "a"))

	r.check(t, workedClock)
	want := expectedFromCheck(t, r.state)
	if len(want) == 0 {
		t.Fatal("run 1 recorded no outcome, so the fixture does not launch")
	}
	if got := logOutcomes(t, r.state); !slices.Equal(got, want) {
		t.Fatalf("after run 1 the log holds %v, want %v", got, want)
	}

	rows := r.readRows(t)
	rows[0]["state"], rows[0]["status"], rows[0]["waitingFor"] = "blocked", "blocked", "permission prompt"
	r.setRows(t, rows...)
	r.check(t, minutes(t, 1))
	want = append(want, expectedFromCheck(t, r.state)...)
	got := logOutcomes(t, r.state)
	if !slices.Equal(got, want) {
		t.Fatalf("after run 2 the log holds %v, want %v", got, want)
	}
	if p := r.passRecord(t, "a"); p["state"] != "stopped" {
		t.Errorf("pass = %v, want stopped by run 2", p)
	}
	if last := got[len(got)-1]; last != "stopped: permission_prompt" {
		t.Errorf("last log line = %q, want the stop run 2 applied", last)
	}

	var snap struct {
		CheckLog []struct {
			Outcome string `json:"outcome"`
		} `json:"check_log"`
	}
	if err := json.Unmarshal([]byte(readFileIn(t, r.state, "snapshot.json")), &snap); err != nil {
		t.Fatal(err)
	}
	if len(snap.CheckLog) != len(want) {
		t.Errorf("snapshot check_log has %d lines, want %d", len(snap.CheckLog), len(want))
	}
}
