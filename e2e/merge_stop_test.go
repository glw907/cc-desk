//go:build unix

package e2e

import (
	"os"
	"slices"
	"strings"
	"testing"
)

// mergeOne reports one clean pass ready and runs the check that merges it,
// with every session row set to rowState first. It returns the pass's
// short id and the run, which the caller checks.
func (m mergeRig) mergeOne(t *testing.T, pass, rowState string) (string, result) {
	t.Helper()
	id := m.sessionOf(t, pass)["short_id"].(string)
	head := m.push(t, pass, pass+".txt")
	m.report(t, pass, 1, head, "inside")
	m.answer(t, pass, 1, head, passing)
	rows := m.readRows(t)
	for _, r := range rows {
		r["state"], r["status"] = rowState, rowState
	}
	m.setRows(t, rows...)
	return id, m.check(t, mergeClock, "")
}

// stops returns the ids the stand-in claude was asked to stop.
func (m mergeRig) stops(t *testing.T) []string {
	t.Helper()
	var ids []string
	for _, c := range m.calls(t) {
		if c.Args[0] == "stop" {
			ids = append(ids, c.Args[1])
		}
	}
	return ids
}

// lastOutcome is the outcome of the last line of the check's log.
func lastOutcome(t *testing.T, dir string) string {
	t.Helper()
	out := logOutcomes(t, dir)
	if len(out) == 0 {
		t.Fatal("the check log is empty")
	}
	return out[len(out)-1]
}

func TestMergeStopsSession(t *testing.T) {
	t.Run("a done row is stopped", func(t *testing.T) {
		m := newMergeRig(t, "sd")
		id, run := m.mergeOne(t, "sd", "done")
		requireRun(t, run)
		if got := m.stops(t); !slices.Equal(got, []string{id}) {
			t.Errorf("stops = %v, want [%s]", got, id)
		}
		want := "merged pull request 1; stopped session " + id + "; cleaned up"
		if got := lastOutcome(t, m.state); !strings.HasPrefix(got, want) {
			t.Errorf("outcome = %q, want one starting %q", got, want)
		}
		if m.pullOf(t, "sd")["cleaned"] == nil {
			t.Error("the pass has no cleaned time")
		}
	})
	t.Run("a stopped row takes no stop", func(t *testing.T) {
		m := newMergeRig(t, "sg")
		id, run := m.mergeOne(t, "sg", "stopped")
		requireRun(t, run)
		if got := m.stops(t); len(got) != 0 {
			t.Errorf("stops = %v, want none", got)
		}
		want := "merged pull request 1; session " + id + " already gone; cleaned up"
		if got := lastOutcome(t, m.state); !strings.HasPrefix(got, want) {
			t.Errorf("outcome = %q, want one starting %q", got, want)
		}
	})
	t.Run("a failed stop on a session that is gone", func(t *testing.T) {
		m := newMergeRig(t, "sf")
		m.control(t, "stop-fails", true)
		m.control(t, "stop-marks", true)
		id, run := m.mergeOne(t, "sf", "done")
		requireRun(t, run)
		if m.pullOf(t, "sf")["cleaned"] == nil {
			t.Error("a stop that failed on a gone session left the pass uncleaned")
		}
		want := "merged pull request 1; session " + id + " already gone; cleaned up"
		if got := lastOutcome(t, m.state); !strings.HasPrefix(got, want) {
			t.Errorf("outcome = %q, want one starting %q", got, want)
		}
	})
	t.Run("a failed stop on a live session", func(t *testing.T) {
		m := newMergeRig(t, "sl")
		m.control(t, "stop-fails", true)
		id, run := m.mergeOne(t, "sl", "done")
		if run.code != 2 || !strings.Contains(run.stderr, "claude stop "+id) {
			t.Errorf("run = exit %d, %q; want exit 2 naming claude stop %s", run.code, run.stderr, id)
		}
		pr := m.pullOf(t, "sl")
		if pr["merged"] == nil || pr["cleaned"] != nil {
			t.Errorf("pull request = %v, want merged and not cleaned", pr)
		}
		if got := m.stops(t); !slices.Equal(got, []string{id}) {
			t.Errorf("stops = %v, want [%s]", got, id)
		}
		if _, err := os.Stat(m.worktree("sl")); err != nil {
			t.Errorf("the worktree went with its session still listed: %v", err)
		}
	})
}

// TestEndStopsSession ends a finished conductor whose row is in each state.
func TestEndStopsSession(t *testing.T) {
	for _, tc := range []struct {
		name, row  string
		fails, got bool
		wantStops  int
		wantEnded  bool
	}{
		{"done row", "done", false, false, 1, true},
		{"stopped row", "stopped", false, false, 0, true},
		{"failed stop, session gone", "done", true, true, 1, true},
		{"failed stop, done row stays listed", "done", true, false, 1, true},
		{"failed stop, live row stays listed", "working", true, false, 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := newLaunchRig(t)
			r.addRelease(t, true, "a")
			requireRun(t, r.run(t, workedClock, "session", "start", "conductor", "proj"))
			requireRun(t, r.do(t, workedClock, "session", "finish", "conductor", "proj"))
			rows := r.readRows(t)
			for _, row := range rows {
				row["state"], row["status"] = tc.row, tc.row
			}
			r.setRows(t, rows...)
			r.control(t, "stop-fails", tc.fails)
			r.control(t, "stop-marks", tc.got)
			got := r.run(t, minutes(t, 1), "check")
			wantCode := 2
			if tc.wantEnded {
				wantCode = 0
			}
			if got.code != wantCode {
				t.Errorf("exit = %d, want %d; stderr %q", got.code, wantCode, got.stderr)
			}
			if !strings.Contains(got.stdout, `"decision":"end"`) {
				t.Fatalf("check output lacks end:\n%s%s", got.stdout, got.stderr)
			}
			stops := 0
			for _, k := range r.kinds(t) {
				if k == "stop" {
					stops++
				}
			}
			if stops != tc.wantStops {
				t.Errorf("stops = %d, want %d (%v)", stops, tc.wantStops, r.kinds(t))
			}
			if ended := r.conductor(t)["ended"] != nil; ended != tc.wantEnded {
				t.Errorf("conductor ended = %v, want %v", ended, tc.wantEnded)
			}
		})
	}
}
