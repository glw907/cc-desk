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
