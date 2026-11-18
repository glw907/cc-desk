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
	got, fc, err := Render(&state.State{}, clock)
	if err != nil {
		t.Fatal(err)
	}
	if fc != (state.Forecast{}) {
		t.Errorf("forecast of an empty state = %+v, want the zero line", fc)
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
	got, _, err := Render(&state.State{}, clock)
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
		"sessions":     "[]", "needs_you_now": "[]", "last_check": "null",
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

// ordered is a release with a free pass, a pass behind unfinished work and
// an unanswered question, a waiting pass on both limits, and a building and
// a waiting pass that each carry an open guess.
func ordered(t *testing.T) *state.ReleaseFile {
	return mustRelease(t, `{"format":1,"project":"p","name":"v1","repo":"r",
	  "rounds":[{"name":"r1","summary":"s","needs_geoff":true,"design_written":true}],
	  "passes":[
	    {"name":"free","summary":"s","round":"r1","planned_minutes":30,"state":"not_started"},
	    {"name":"late","summary":"s","round":"r1","planned_minutes":30,"state":"not_started","after":["free"],"questions":["q1"]},
	    {"name":"held","summary":"s","round":"r1","planned_minutes":40,"state":"waiting",
	     "wait":{"since":"2026-11-20T22:00:00Z","limits":["two_pass_limit","usage"]}},
	    {"name":"g1","summary":"s","round":"r1","planned_minutes":200,"state":"building","started":"2026-11-20T22:30:00Z",
	     "questions":["q2"],"guess":{"question":"q2","recommendation":"x","at":"2026-11-20T22:30:00Z"}},
	    {"name":"g2","summary":"s","round":"r1","planned_minutes":90,"state":"waiting",
	     "wait":{"since":"2026-11-20T21:00:00Z","limits":["two_pass_limit"]},
	     "questions":["q3"],"guess":{"question":"q3","recommendation":"x","at":"2026-11-20T21:00:00Z"}}],
	  "questions":[
	    {"id":"q1","text":"t","options":["x","y"],"recommended":"x"},
	    {"id":"q2","text":"t","options":["x","y"],"recommended":"x"},
	    {"id":"q3","text":"t","options":["x","y"],"recommended":"x"}]}`)
}

func passNamed(t *testing.T, s *Snapshot, name string) Pass {
	t.Helper()
	for _, p := range s.Releases[0].Rounds[0].Passes {
		if p.Name == name {
			return p
		}
	}
	t.Fatalf("no pass %s", name)
	return Pass{}
}

func TestSnapshotFree(t *testing.T) {
	st := &state.State{
		Releases: []*state.ReleaseFile{ordered(t)},
		Usage:    &state.Usage{Format: 1, Percent: 100, ResetText: "Nov 20, 10:59pm", ReadAt: clock},
	}
	snap := Build(st, clock)

	if p := passNamed(t, snap, "free"); p.Free == nil || !*p.Free || len(p.WhyNot) != 0 {
		t.Errorf("free pass: free %v why_not %v", p.Free, p.WhyNot)
	}
	p := passNamed(t, snap, "late")
	want := []string{"unfinished work before it", "an unanswered question"}
	if p.Free == nil || *p.Free || strings.Join(p.WhyNot, "|") != strings.Join(want, "|") {
		t.Errorf("late pass: free %v why_not %v, want %v", p.Free, p.WhyNot, want)
	}
	if p := passNamed(t, snap, "held"); p.Free != nil {
		t.Errorf("a waiting pass reads free %v, want null", *p.Free)
	}

	if len(snap.Waiting) != 2 {
		t.Fatalf("waiting = %+v", snap.Waiting)
	}
	if w := snap.Waiting[0]; w.Pass != "g2" || w.ResetText != nil || strings.Join(w.Limits, ",") != "two_pass_limit" {
		t.Errorf("oldest waiting pass = %+v", w)
	}
	w := snap.Waiting[1]
	if w.Pass != "held" || w.ResetText == nil || *w.ResetText != "Nov 20, 10:59pm" ||
		strings.Join(w.Limits, ",") != "two_pass_limit,usage" || !w.Since.Equal(clock.Add(-time.Hour)) {
		t.Errorf("held waiting entry = %+v", w)
	}
}

func TestSnapshotUsageUnreadable(t *testing.T) {
	st := &state.State{
		Releases: []*state.ReleaseFile{ordered(t)},
		Usage:    &state.Usage{Format: 1, Error: "no line", ReadAt: clock},
	}
	snap := Build(st, clock)
	n := snap.NextSession
	if n.State == nil || *n.State != "needs_you_now" || len(n.Reasons) != 1 || n.Reasons[0] != "the usage figure can't be read" {
		t.Errorf("next_session = state %v reasons %v", n.State, n.Reasons)
	}
	for _, w := range snap.Waiting {
		if w.ResetText != nil {
			t.Errorf("%s has reset text %q on an unreadable reading", w.Pass, *w.ResetText)
		}
	}

	none := Build(&state.State{Releases: []*state.ReleaseFile{ordered(t)}}, clock)
	if len(none.NextSession.Reasons) != 0 || none.NextSession.Reasons == nil {
		t.Errorf("a missing reading gives reasons %v, want []", none.NextSession.Reasons)
	}
}

func TestSnapshotOpenGuesses(t *testing.T) {
	snap := Build(&state.State{Releases: []*state.ReleaseFile{ordered(t)}}, clock)
	if len(snap.OpenGuesses) != 2 || snap.GuessMinutesTotal != 290 {
		t.Fatalf("open_guesses %+v total %d, want two entries and 290", snap.OpenGuesses, snap.GuessMinutesTotal)
	}
	if g := snap.OpenGuesses[0]; g != (OpenGuess{Project: "p", Pass: "g1", Question: "q2", Minutes: 200}) {
		t.Errorf("first guess = %+v", g)
	}
	if g := snap.OpenGuesses[1]; g != (OpenGuess{Project: "p", Pass: "g2", Question: "q3", Minutes: 90}) {
		t.Errorf("second guess = %+v", g)
	}
}

func TestSnapshotGuessOutcome(t *testing.T) {
	const rel = `{"format":1,"project":"p","name":"v1","repo":"r",
	  "rounds":[{"name":"r1","summary":"s","needs_geoff":true,"design_written":true}],
	  "passes":[
	    {"name":"g","summary":"s","round":"r1","planned_minutes":50,"state":"building","started":"2026-11-20T22:30:00Z",
	     "questions":["q"],"guess":{"question":"q","recommendation":"x","at":"2026-11-20T22:30:00Z"}},
	    {"name":"plain","summary":"s","round":"r1","planned_minutes":50,"state":"not_started"}],
	  "questions":[{"id":"q","text":"t","options":["x","y"],"recommended":"x"%s}]}`
	withAnswer := func(answer string, rulings ...state.Ruling) *state.State {
		return &state.State{
			Releases: []*state.ReleaseFile{mustRelease(t, strings.Replace(rel, "%s", answer, 1))},
			Rulings:  rulings,
		}
	}
	ruling := `,"answer":{"ruling":"r-0001"}`
	tests := []struct {
		name string
		s    *state.State
		want string // "" is null
	}{
		{"held", withAnswer(ruling, state.Ruling{ID: "r-0001", Option: "x"}), Held},
		{"missed", withAnswer(ruling, state.Ruling{ID: "r-0001", Option: "y"}), Missed},
		{"open", withAnswer(""), ""},
		{"carried over", withAnswer(`,"answer":{"carried":{"words":"w","date":"2026-11-01","source":"s"}}`), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := passNamed(t, Build(tt.s, clock), "g").GuessOutcome
			switch {
			case tt.want == "" && got != nil:
				t.Errorf("guess_outcome = %q, want null", *got)
			case tt.want != "" && (got == nil || *got != tt.want):
				t.Errorf("guess_outcome = %v, want %q", got, tt.want)
			}
		})
	}
	held := withAnswer(ruling, state.Ruling{ID: "r-0001", Option: "x"})
	if got := passNamed(t, Build(held, clock), "plain").GuessOutcome; got != nil {
		t.Errorf("a pass with no guess has outcome %q", *got)
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
