package snapshot

import (
	"slices"
	"testing"

	"github.com/glw907/cc-desk/internal/state"
)

func TestReportPassesComeFromTheirOwnRelease(t *testing.T) {
	rounds := []state.Round{{Name: "r", DesignWritten: true}}
	v1 := releaseOf("a", rounds, donePass("old", "r", 10, 10, utcTime(18, 22, 0)))
	v1.Path, v1.Done = "releases/done/v1-20261119T000000Z.json", true
	v2 := releaseOf("a", rounds, donePass("new1", "r", 20, 20, utcTime(19, 22, 0)), donePass("new2", "r", 30, 30, utcTime(19, 23, 0)))
	v2.Name = "v2"
	s := &state.State{
		Releases: []*state.ReleaseFile{v2, v1},
		Reports: state.Reports{Reports: []state.Report{
			{ID: "rep-0001", Project: "a", Release: "v1", Title: "T", Summary: "S"},
			{ID: "rep-0002", Project: "a", Release: "v2", Round: "r", Title: "T", Summary: "S"},
			{ID: "rep-0003", Project: "a", Release: "v2", Title: "T", Summary: "S"},
			{ID: "rep-0004", Project: "a", Round: "r", Title: "T", Summary: "S"},
		}},
	}
	tests := []struct {
		id   string
		want []string
	}{
		{"rep-0001", []string{"old"}},
		{"rep-0002", []string{"new1", "new2"}},
		{"rep-0003", []string{"new1", "new2"}},
		// A report with no release name reads the project's first release.
		{"rep-0004", []string{"new1", "new2"}},
	}
	got := map[string][]string{}
	for _, e := range buildReports(s) {
		for _, p := range e.Passes {
			got[e.ID] = append(got[e.ID], p.Name)
		}
	}
	for _, tc := range tests {
		if !slices.Equal(got[tc.id], tc.want) {
			t.Errorf("%s passes = %v, want %v", tc.id, got[tc.id], tc.want)
		}
	}
}
