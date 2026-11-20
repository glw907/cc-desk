package state

import (
	"strings"
	"testing"
	"time"
)

func TestReleaseDoneMovesTheFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "releases/alpha.json", goodRelease)
	var path string
	if _, err := Commit(dir, testClock, time.Second, ReleaseDone("alpha", &path), render); err != nil {
		t.Fatal(err)
	}
	want := "releases/done/v1-20261120T230000Z.json"
	if path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
	if exists(dir, "releases/alpha.json") {
		t.Error("releases/alpha.json survived the move")
	}
	if !strings.Contains(readFile(t, dir, want), `"project": "alpha"`) {
		t.Errorf("%s does not hold the release", want)
	}

	s, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Releases) != 1 || !s.Releases[0].Done || s.Releases[0].Path != want {
		t.Fatalf("loaded releases = %+v, want the one done release", s.Releases)
	}
	if _, err := s.liveRelease("alpha"); err == nil {
		t.Error("the project still has a live release")
	}
}

func TestReleaseDoneRefuses(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, dir string)
		want  string
	}{
		{"no release", func(t *testing.T, dir string) {}, "no live release"},
		{"only a done release", func(t *testing.T, dir string) {
			writeFile(t, dir, "releases/done/v1-20261101T000000Z.json", goodRelease)
		}, "no live release"},
		{"the target exists", func(t *testing.T, dir string) {
			writeFile(t, dir, "releases/alpha.json", goodRelease)
			writeFile(t, dir, "releases/done/v1-20261120T230000Z.json", goodRelease)
		}, "already exists"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)
			_, err := Commit(dir, testClock, time.Second, ReleaseDone("alpha", nil), render)
			if !isRefusal(err) || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %v, want a refusal naming %q", err, tt.want)
			}
		})
	}
}

// twoReleases loads a done release v1 and a live release v2 of one project,
// both holding round r1.
func twoReleases(t *testing.T) *State {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, dir, "releases/done/v1-20261101T000000Z.json", goodRelease)
	writeFile(t, dir, "releases/alpha.json", strings.Replace(goodRelease, `"v1"`, `"v2"`, 1))
	s, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestAddReportTagsItsRelease(t *testing.T) {
	rf, err := DecodeReport([]byte(goodReport))
	if err != nil {
		t.Fatal(err)
	}
	s := twoReleases(t)
	for _, round := range []string{"r1", ""} {
		if err := AddReport("alpha", round, rf, nil)(s, at(9, 0)); err != nil {
			t.Fatal(err)
		}
	}
	for i, r := range s.Reports.Reports {
		if r.Release != "v2" {
			t.Errorf("report %d release = %q, want v2", i, r.Release)
		}
	}

	t.Run("a release that ended takes its round report", func(t *testing.T) {
		s := twoReleases(t)
		s.Releases[0].Done, s.Releases[1].Done = true, true
		s.Releases[0].Path = "releases/done/v2-20261120T000000Z.json"
		if err := AddReport("alpha", "r1", rf, nil)(s, at(9, 0)); err != nil {
			t.Fatal(err)
		}
		if got := s.Reports.Reports[0].Release; got != s.Releases[0].Name {
			t.Errorf("release = %q, want the newest done release, %q", got, s.Releases[0].Name)
		}
	})

	t.Run("a second release takes its own report", func(t *testing.T) {
		s := twoReleases(t)
		s.Reports.Reports = []Report{{ID: "rep-0001", Project: "alpha", Release: "v1", Title: "T", Summary: "S"}}
		if err := AddReport("alpha", "", rf, nil)(s, at(9, 0)); err != nil {
			t.Errorf("a v2 release report with a v1 one recorded: %v", err)
		}
		if err := AddReport("alpha", "", rf, nil)(s, at(9, 1)); !isRefusal(err) {
			t.Errorf("a second v2 release report: error = %v, want a refusal", err)
		}
	})

	t.Run("a report with no release keeps the project and round match", func(t *testing.T) {
		s := twoReleases(t)
		s.Reports.Reports = []Report{{ID: "rep-0001", Project: "alpha", Round: "r1", Title: "T", Summary: "S"}}
		if err := AddReport("alpha", "r1", rf, nil)(s, at(9, 0)); !isRefusal(err) {
			t.Errorf("error = %v, want a refusal", err)
		}
	})
}
