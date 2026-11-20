package sessions

import (
	"testing"

	"github.com/glw907/cc-desk/internal/state"
)

func TestReportDueBelongsToItsRelease(t *testing.T) {
	done := state.Done
	report := func(release, round string) state.Report {
		return state.Report{Project: "p", Release: release, Round: round}
	}
	tests := []struct {
		name    string
		release string
		reports []state.Report
		wantDue string
	}{
		{"v1 owes its round report", "v1", nil, "r1"},
		{"v1 owes its release report", "v1", []state.Report{report("v1", "r1"), report("v1", "r2")}, ReleaseReport},
		{"v1 is square", "v1", []state.Report{report("v1", "r1"), report("v1", "r2"), report("v1", "")}, ""},
		{"v2 owes its round report after v1's", "v2",
			[]state.Report{report("v1", "r1"), report("v1", "r2"), report("v1", "")}, "r1"},
		{"v2 owes its release report after v1's", "v2",
			[]state.Report{report("v1", "r1"), report("v1", "r2"), report("v1", ""), report("v2", "r1"), report("v2", "r2")}, ReleaseReport},
		{"v2 is square", "v2", []state.Report{report("v2", "r1"), report("v2", "r2"), report("v2", "")}, ""},
		{"reports with no release match by round", "v2", []state.Report{report("", "r1"), report("", "r2"), report("", "")}, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := reportWorld([]state.PassState{done}, []state.PassState{done}, tc.reports...)
			s.Releases[0].Name = tc.release
			if due, _ := ReportDue(s, s.Releases[0]); due != tc.wantDue {
				t.Errorf("report due %q, want %q", due, tc.wantDue)
			}
		})
	}
}

func TestReportRecordedMatchesItsRelease(t *testing.T) {
	s := reportWorld(nil, nil,
		state.Report{Project: "p", Release: "v1", Round: "r1"},
		state.Report{Project: "p", Release: "v1"},
		state.Report{Project: "p", Round: "r2"})
	tests := []struct {
		release, forReport string
		want               bool
	}{
		{"v1", "r1", true},
		{"v2", "r1", false},
		{"v1", ReleaseReport, true},
		{"v2", ReleaseReport, false},
		{"v2", "r2", true},
		{"v1", "", false},
	}
	for _, tc := range tests {
		if got := ReportRecorded(s, "p", tc.release, tc.forReport); got != tc.want {
			t.Errorf("ReportRecorded(%q, %q) = %v, want %v", tc.release, tc.forReport, got, tc.want)
		}
	}
}
