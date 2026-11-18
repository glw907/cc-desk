package e2e

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

type snap struct {
	Releases []struct {
		Project string `json:"project"`
		Rounds  []struct {
			Name   string `json:"name"`
			State  string `json:"state"`
			Passes []struct {
				Name          string  `json:"name"`
				State         string  `json:"state"`
				Started       *string `json:"started"`
				Finished      *string `json:"finished"`
				ActualMinutes *int    `json:"actual_minutes"`
				StopReason    *string `json:"stop_reason"`
			} `json:"passes"`
		} `json:"rounds"`
	} `json:"releases"`
}

func readSnap(t *testing.T, r result) snap {
	t.Helper()
	if r.code != 0 {
		t.Fatalf("exit %d: %s", r.code, r.stderr)
	}
	var s snap
	if err := json.Unmarshal([]byte(r.stdout), &s); err != nil {
		t.Fatal(err)
	}
	return s
}

// apply runs a mutating verb, which prints status to stderr, then reads the
// snapshot at the same clock.
func apply(t *testing.T, dir, now string, args ...string) snap {
	t.Helper()
	if r := run(t, dir, now, args...); r.code != 0 {
		t.Fatalf("%v exit %d: %s", args, r.code, r.stderr)
	}
	return readSnap(t, run(t, dir, now, "snapshot"))
}

func passOf(t *testing.T, s snap, project, name string) (state string, started, finished *string, actual *int, reason *string) {
	t.Helper()
	for _, rel := range s.Releases {
		if rel.Project != project {
			continue
		}
		for _, rd := range rel.Rounds {
			for _, p := range rd.Passes {
				if p.Name == name {
					return p.State, p.Started, p.Finished, p.ActualMinutes, p.StopReason
				}
			}
		}
	}
	t.Fatalf("no pass %s/%s in the snapshot", project, name)
	return
}

func TestFixtureSnapshotIsCurrent(t *testing.T) {
	dir := newState(t)
	got := run(t, dir, workedClock, "snapshot")
	if got.code != 0 {
		t.Fatalf("exit %d: %s", got.code, got.stderr)
	}
	want, err := os.ReadFile("testdata/state/snapshot.json")
	if err != nil {
		t.Fatal(err)
	}
	if got.stdout != string(want) {
		t.Error("testdata/state/snapshot.json is out of date; rerun the snapshot verb on the fixture at the worked clock")
	}
}

func TestEvents(t *testing.T) {
	dir := newState(t)

	s := apply(t, dir, workedClock, "release", "add", input(t, "gamma.json"))
	var gamma bool
	for _, rel := range s.Releases {
		if rel.Project == "gamma" {
			gamma = true
			if len(rel.Rounds) != 1 || rel.Rounds[0].State != "not_started" || len(rel.Rounds[0].Passes) != 2 {
				t.Errorf("gamma rounds = %+v", rel.Rounds)
			}
		}
	}
	if !gamma {
		t.Fatal("gamma is not in the snapshot")
	}
	if _, err := os.Stat(dir + "/releases/gamma.json"); err != nil {
		t.Error(err)
	}

	steps := []struct {
		now   string
		args  []string
		state string
	}{
		{"2026-11-21T10:00:00Z", []string{"pass", "start", "gamma", "g-one"}, "building"},
		{"2026-11-21T10:20:00Z", []string{"pass", "stop", "gamma", "g-one", "--reason", "scope"}, "stopped"},
		{"2026-11-21T10:50:00Z", []string{"pass", "start", "gamma", "g-one"}, "building"},
		{"2026-11-21T11:30:00Z", []string{"pass", "finish", "gamma", "g-one"}, "done"},
	}
	for _, st := range steps {
		state, started, finished, actual, reason := passOf(t, apply(t, dir, st.now, st.args...), "gamma", "g-one")
		if state != st.state {
			t.Errorf("after %v: state %s, want %s", st.args, state, st.state)
		}
		switch st.args[1] {
		case "start":
			if started == nil || *started != "2026-11-21T10:00:00Z" {
				t.Errorf("started = %v", started)
			}
		case "stop":
			if reason == nil || *reason != "scope" {
				t.Errorf("stop_reason = %v", reason)
			}
		case "finish":
			if finished == nil || *finished != st.now {
				t.Errorf("finished = %v, want %s", finished, st.now)
			}
			// 10:00 to 11:30 is 90 minutes, less the 30 stopped.
			if actual == nil || *actual != 60 {
				t.Errorf("actual_minutes = %v, want 60", actual)
			}
		}
	}
}

func TestBadRequests(t *testing.T) {
	t.Run("finish twice", func(t *testing.T) {
		dir := newState(t)
		apply(t, dir, "2026-11-21T10:00:00Z", "pass", "start", "alpha", "p13")
		apply(t, dir, "2026-11-21T10:30:00Z", "pass", "finish", "alpha", "p13")
		before := folder(t, dir)
		requireSnapshot(t, before)
		got := run(t, dir, "2026-11-21T11:00:00Z", "pass", "finish", "alpha", "p13")
		if got.code != 2 {
			t.Fatalf("exit %d, want 2; stderr: %s", got.code, got.stderr)
		}
		sameFolder(t, before, folder(t, dir))
	})

	refusals := []struct {
		name string
		prep func(t *testing.T, dir string)
		args func(t *testing.T) []string
	}{
		{"finish before a start", nil, func(*testing.T) []string { return []string{"pass", "finish", "alpha", "p13"} }},
		{"stop before a start", nil, func(*testing.T) []string { return []string{"pass", "stop", "alpha", "p13", "--reason", "usage"} }},
		{"start a done pass", nil, func(*testing.T) []string { return []string{"pass", "start", "alpha", "p01"} }},
		{"release for a live project", nil, func(t *testing.T) []string {
			return []string{"release", "add", input(t, "gamma.json")}
		}},
		{"design already written", nil, func(t *testing.T) []string {
			return []string{"round", "design", "alpha", "r2", input(t, "design-b3.json")}
		}},
	}
	for _, c := range refusals {
		t.Run(c.name, func(t *testing.T) {
			dir := newState(t)
			if c.name == "release for a live project" {
				apply(t, dir, workedClock, "release", "add", input(t, "gamma.json"))
			}
			before := folder(t, dir)
			requireSnapshot(t, before)
			got := run(t, dir, workedClock, c.args(t)...)
			if got.code != 2 {
				t.Fatalf("exit %d, want 2; stderr: %s", got.code, got.stderr)
			}
			sameFolder(t, before, folder(t, dir))
		})
	}

	names := []struct {
		name string
		args func(t *testing.T) []string
		want string
	}{
		{"project", func(*testing.T) []string { return []string{"pass", "start", "nope", "p1"} }, "nope"},
		{"pass", func(*testing.T) []string { return []string{"pass", "start", "alpha", "p99"} }, "p99"},
		{"round", func(t *testing.T) []string {
			return []string{"round", "design", "beta", "b9", input(t, "design-b3.json")}
		}, "b9"},
		{"question", func(t *testing.T) []string {
			return []string{"round", "design", "beta", "b3", input(t, "design-undefined-question.json")}
		}, "beta-q9"},
	}
	for _, c := range names {
		t.Run("unknown "+c.name, func(t *testing.T) {
			dir := newState(t)
			before := folder(t, dir)
			requireSnapshot(t, before)
			got := run(t, dir, workedClock, c.args(t)...)
			if got.code != 1 {
				t.Fatalf("exit %d, want 1; stderr: %s", got.code, got.stderr)
			}
			if !strings.Contains(got.stderr, c.want) {
				t.Errorf("stderr does not name %q: %s", c.want, got.stderr)
			}
			sameFolder(t, before, folder(t, dir))
		})
	}
}
