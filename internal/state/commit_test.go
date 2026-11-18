package state

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var testClock = time.Date(2026, 11, 20, 23, 0, 0, 0, time.UTC)

func render(*State, time.Time) ([]byte, Forecast, error) {
	return []byte("{\"fresh\":true}\n"), Forecast{}, nil
}

func readFile(t *testing.T, dir, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, rel))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func exists(dir, rel string) bool {
	_, err := os.Stat(filepath.Join(dir, rel))
	return err == nil
}

func TestCommitWritesSnapshotAndDirtyRelease(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "releases/alpha.json", goodRelease)
	writeFile(t, dir, SnapshotErrorFile, "stale")
	ev := func(s *State, now time.Time) error {
		s.Releases[0].Name = "v2"
		s.Releases[0].Dirty = true
		return nil
	}
	got, err := Commit(dir, testClock, time.Second, ev, render)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != readFile(t, dir, SnapshotFile) {
		t.Error("returned bytes differ from snapshot.json")
	}
	if !strings.Contains(readFile(t, dir, "releases/alpha.json"), `"name": "v2"`) {
		t.Error("the dirty release was not written")
	}
	if exists(dir, SnapshotErrorFile) {
		t.Error("snapshot-error.json survived a clean run")
	}
	if !exists(dir, "lock") {
		t.Error("lock file missing")
	}
}

func TestCommitCleanReleaseIsNotRewritten(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "releases/alpha.json", goodRelease)
	if _, err := Commit(dir, testClock, time.Second, nil, render); err != nil {
		t.Fatal(err)
	}
	if readFile(t, dir, "releases/alpha.json") != goodRelease {
		t.Error("an untouched release was rewritten")
	}
}

func TestCommitLoadFailure(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "releases/alpha.json", `{"format":1,`)
	writeFile(t, dir, SnapshotFile, "old snapshot")
	_, err := Commit(dir, testClock, time.Second, nil, render)
	if err == nil {
		t.Fatal("Commit accepted a broken file")
	}
	if readFile(t, dir, SnapshotFile) != "old snapshot" {
		t.Error("snapshot.json changed on a load failure")
	}
	e := readFile(t, dir, SnapshotErrorFile)
	if !strings.Contains(e, `"file": "releases/alpha.json"`) || !strings.Contains(e, "unexpected EOF") {
		t.Errorf("snapshot-error.json = %s", e)
	}
	if !strings.Contains(e, strings.ReplaceAll(err.Error(), `"`, `\"`)) {
		t.Errorf("error file message differs from the returned error %q:\n%s", err, e)
	}
}

func TestCommitRefusalClearsErrorFileAndWritesNothingElse(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "releases/alpha.json", goodRelease)
	writeFile(t, dir, SnapshotFile, "old snapshot")
	writeFile(t, dir, SnapshotErrorFile, "stale")
	ev := func(s *State, now time.Time) error {
		s.Releases[0].Name = "changed in memory"
		s.Releases[0].Dirty = true
		return Refuse("no")
	}
	_, err := Commit(dir, testClock, time.Second, ev, render)
	if _, ok := errors.AsType[*RefusedError](err); !ok {
		t.Fatalf("error = %v, want a refusal", err)
	}
	if exists(dir, SnapshotErrorFile) {
		t.Error("a refusal after a clean load left snapshot-error.json")
	}
	if readFile(t, dir, SnapshotFile) != "old snapshot" || readFile(t, dir, "releases/alpha.json") != goodRelease {
		t.Error("a refusal wrote a file")
	}
}

func TestCommitLockTimeoutWritesNothing(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "releases/alpha.json", `{"format":1,`)
	unlock, err := Lock(dir, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	_, err = Commit(dir, testClock, 100*time.Millisecond, nil, render)
	if err == nil || !strings.Contains(err.Error(), "locked") {
		t.Fatalf("error = %v, want a lock timeout", err)
	}
	if exists(dir, SnapshotErrorFile) || exists(dir, SnapshotFile) {
		t.Error("a lock timeout wrote a file")
	}
}

func TestCommitCreatesStateFolder(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "new", "state")
	if _, err := Commit(dir, testClock, time.Second, nil, render); err != nil {
		t.Fatal(err)
	}
	if !exists(dir, SnapshotFile) {
		t.Error("snapshot.json missing in the created folder")
	}
}

func TestCommitReading(t *testing.T) {
	loadedReading := `{"format":1,"percent":40,"reset_text":"old","read_at":"2026-11-20T22:58:00Z"}`
	fresh := func(at time.Time) Usage { return Usage{Percent: 55, ResetText: "new", ReadAt: at} }

	tests := []struct {
		name    string
		file    string // usage.json before; "" for none
		reading Usage
		forced  bool
		want    string // substring of usage.json after
	}{
		{"a fresh reading is written", "", fresh(testClock), false, `"reset_text": "new"`},
		{"a newer reading replaces the loaded one", loadedReading, fresh(testClock), false, `"reset_text": "new"`},
		{"an older reading leaves usage.json alone", loadedReading, fresh(testClock.Add(-10 * time.Minute)), false, `"reset_text":"old"`},
		{"a tie keeps the loaded reading", loadedReading, fresh(testClock.Add(-2 * time.Minute)), false, `"reset_text":"old"`},
		{"a forced reading always wins", loadedReading, fresh(testClock.Add(-10 * time.Minute)), true, `"reset_text": "new"`},
		{"a loaded reading from the future loses", `{"format":1,"percent":40,"reset_text":"old","read_at":"2026-11-21T22:58:00Z"}`, fresh(testClock.Add(-time.Hour)), false, `"reset_text": "new"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, dir, "releases/alpha.json", goodRelease)
			if tt.file != "" {
				writeFile(t, dir, UsageFile, tt.file)
			}
			if _, err := Commit(dir, testClock, time.Second, WithReading(tt.reading, tt.forced, nil), render); err != nil {
				t.Fatal(err)
			}
			if got := readFile(t, dir, UsageFile); !strings.Contains(got, tt.want) {
				t.Errorf("usage.json = %s, want it to hold %s", got, tt.want)
			}
		})
	}

	t.Run("a refused event leaves the folder as it was", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "releases/alpha.json", goodRelease)
		writeFile(t, dir, UsageFile, loadedReading)
		refuse := func(*State, time.Time) error { return Refuse("no") }
		_, err := Commit(dir, testClock, time.Second, WithReading(fresh(testClock), true, refuse), render)
		if !isRefusal(err) {
			t.Fatalf("error = %v, want a refusal", err)
		}
		if readFile(t, dir, UsageFile) != loadedReading {
			t.Error("a refusal rewrote usage.json")
		}
	})
}
