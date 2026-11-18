package snapshot

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glw907/cc-desk/internal/state"
)

var update = flag.Bool("update", false, "rewrite the golden files")

var clock = time.Date(2026, 11, 20, 23, 0, 0, 0, time.UTC)

func TestSnapshotEmptyGolden(t *testing.T) {
	got, err := Render(&state.State{}, clock)
	if err != nil {
		t.Fatal(err)
	}
	golden := filepath.Join("testdata", "empty.golden.json")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(golden, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Errorf("snapshot of an empty state differs from %s:\n%s", golden, got)
	}
}

// TestSnapshotEveryFieldPresent checks the shape the golden hides behind
// its bytes: no top-level field encodes as null where it should be a list
// or object.
func TestSnapshotEveryFieldPresent(t *testing.T) {
	got, err := Render(&state.State{}, clock)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(got, &m); err != nil {
		t.Fatal(err)
	}
	wantJSON := map[string]string{
		"format": "1", "releases": "[]", "rulings": "[]", "waits_for_geoff": "[]",
		"waiting": "[]", "open_guesses": "[]", "guess_minutes_total": "0", "guess_minutes_limit": "300",
		"generated_at": `"2026-11-20T23:00:00Z"`,
	}
	for k, v := range wantJSON {
		if string(m[k]) != v {
			t.Errorf("%s = %s, want %s", k, m[k], v)
		}
	}
	var next map[string]json.RawMessage
	if err := json.Unmarshal(m["next_session"], &next); err != nil {
		t.Fatalf("next_session is not an object: %s", m["next_session"])
	}
	for _, k := range []string{"time", "range", "rough"} {
		if string(next[k]) != "null" {
			t.Errorf("next_session.%s = %s, want null", k, next[k])
		}
	}
	if string(next["state"]) != `"nothing_waiting"` || string(next["question_count"]) != "0" {
		t.Errorf("next_session state = %s, question_count = %s; want nothing_waiting, 0", next["state"], next["question_count"])
	}
	for _, k := range []string{"rounds", "projects", "not_designed", "waiting_on_design"} {
		if string(next[k]) != "[]" {
			t.Errorf("next_session.%s = %s, want []", k, next[k])
		}
	}
}

func mustRelease(t *testing.T, j string) *state.ReleaseFile {
	t.Helper()
	r, err := state.DecodeRelease([]byte(j))
	if err != nil {
		t.Fatal(err)
	}
	return &state.ReleaseFile{Path: "releases/p.json", Release: *r}
}

func TestSnapshotRounds(t *testing.T) {
	rf := mustRelease(t, `{"format":1,"project":"p","name":"v1","repo":"r",
	  "rounds":[
	    {"name":"done","summary":"s","needs_geoff":true},
	    {"name":"ask","summary":"s","needs_geoff":true,"after":["done"]},
	    {"name":"quiet","summary":"s","needs_geoff":false,"no_need_reason":"routine","planned_minutes":50},
	    {"name":"later","summary":"s","needs_geoff":true,"planned_minutes":80}],
	  "passes":[
	    {"name":"a","summary":"s","round":"done","planned_minutes":100,"state":"done",
	     "started":"2026-11-20T20:00:00Z","finished":"2026-11-20T21:30:00Z","actual_minutes":90},
	    {"name":"b","summary":"s","round":"ask","planned_minutes":60,"state":"not_started","questions":["q1"]}],
	  "questions":[{"id":"q1","text":"t","options":["x","y"],"recommended":"x",
	    "answer":{"carried":{"words":"w","date":"2026-11-14","source":"chat"}}}]}`)
	snap := Build(&state.State{Releases: []*state.ReleaseFile{rf}}, clock)
	if len(snap.Releases) != 1 || len(snap.Releases[0].Rounds) != 4 {
		t.Fatalf("snapshot = %+v", snap.Releases)
	}
	rounds := snap.Releases[0].Rounds

	if r := rounds[0]; r.State != "done" || r.ActualMinutes == nil || *r.ActualMinutes != 90 || r.PlannedMinutes != 100 {
		t.Errorf("done round = %+v", r)
	}
	if r := rounds[1]; r.State != "not_started" || r.Needs.Kind != "questions" ||
		len(r.Needs.Questions) != 1 || !r.Needs.Questions[0].Answered ||
		*r.Needs.Questions[0].AnswerKind != "carried" {
		t.Errorf("ask round = %+v", r)
	}
	if r := rounds[2]; r.State != "not_designed" || r.Needs.Kind != "nothing" || *r.Needs.Reason != "routine" || r.PlannedMinutes != 50 {
		t.Errorf("quiet round = %+v", r)
	}
	if r := rounds[3]; r.State != "not_designed" || r.Needs.Kind != "not_designed" || r.Needs.Reason != nil {
		t.Errorf("later round = %+v", r)
	}
	if p := rounds[0].Passes[0]; p.Free != nil || len(p.WhyNot) != 0 || p.WhyNot == nil || p.GuessOutcome != nil {
		t.Errorf("pass placeholders = %+v", p)
	}

	out, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), `"why_not":null`) || strings.Contains(string(out), `"questions":null`) {
		t.Errorf("a list encoded as null: %s", out)
	}
}

func TestSnapshotLeavesOutDoneReleases(t *testing.T) {
	rf := mustRelease(t, `{"format":1,"project":"p","name":"v1","repo":"r"}`)
	rf.Done = true
	snap := Build(&state.State{Releases: []*state.ReleaseFile{rf}}, clock)
	if len(snap.Releases) != 0 {
		t.Errorf("a done release appeared: %+v", snap.Releases)
	}
}
