package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// Clocks, all in UTC. Alaska is nine hours behind on 2026-11-20.
const (
	at200 = workedClock            // 2:00 pm
	at230 = "2026-11-20T23:30:00Z" // 2:30 pm
	at250 = "2026-11-20T23:50:00Z" // 2:50 pm
	at520 = "2026-11-21T02:20:00Z" // 5:20 pm
)

type estSnap struct {
	Releases []struct {
		Project  string   `json:"project"`
		Pace     float64  `json:"pace"`
		PaceLow  *float64 `json:"pace_low"`
		PaceHigh *float64 `json:"pace_high"`
		Rough    bool     `json:"rough"`
		Rounds   []struct {
			Name          string  `json:"name"`
			State         string  `json:"state"`
			MinutesBefore *int    `json:"minutes_before"`
			ForecastTime  *string `json:"forecast_time"`
			Needs         struct {
				Kind string `json:"kind"`
			} `json:"needs"`
			Passes []struct {
				Name       string  `json:"name"`
				NeededTime *string `json:"needed_time"`
			} `json:"passes"`
		} `json:"rounds"`
	} `json:"releases"`
	NextSession struct {
		State           string                      `json:"state"`
		Time            *string                     `json:"time"`
		Range           *struct{ Low, High string } `json:"range"`
		Rough           *bool                       `json:"rough"`
		QuestionCount   int                         `json:"question_count"`
		Rounds          []string                    `json:"rounds"`
		Projects        []string                    `json:"projects"`
		NotDesigned     []string                    `json:"not_designed"`
		WaitingOnDesign []string                    `json:"waiting_on_design"`
	} `json:"next_session"`
	WaitsForGeoff []struct {
		Project   string `json:"project"`
		Round     string `json:"round"`
		Questions int    `json:"questions"`
	} `json:"waits_for_geoff"`
}

func readEst(t *testing.T, dir, now string) estSnap {
	t.Helper()
	r := run(t, dir, now, "snapshot")
	if r.code != 0 {
		t.Fatalf("snapshot exit %d: %s", r.code, r.stderr)
	}
	var s estSnap
	if err := json.Unmarshal([]byte(r.stdout), &s); err != nil {
		t.Fatal(err)
	}
	return s
}

func do(t *testing.T, dir, now string, args ...string) {
	t.Helper()
	if r := run(t, dir, now, args...); r.code != 0 {
		t.Fatalf("%v exit %d: %s", args, r.code, r.stderr)
	}
}

func str(p *string) string {
	if p == nil {
		return "null"
	}
	return *p
}

func (s estSnap) release(t *testing.T, project string) int {
	t.Helper()
	for i, r := range s.Releases {
		if r.Project == project {
			return i
		}
	}
	t.Fatalf("no release %s", project)
	return 0
}

// wantSession checks the next session's time, range, rough mark, and the
// fields every row of the table shares.
func wantSession(t *testing.T, label string, s estSnap, state, tm, low, high string) {
	t.Helper()
	n := s.NextSession
	if n.State != state || str(n.Time) != tm {
		t.Errorf("%s: state %s time %s, want %s %s", label, n.State, str(n.Time), state, tm)
	}
	if low == "" {
		if n.Range != nil || n.Rough == nil || !*n.Rough {
			t.Errorf("%s: range %+v rough %v, want null and true", label, n.Range, n.Rough)
		}
		return
	}
	if n.Range == nil || n.Range.Low != low || n.Range.High != high || n.Rough == nil || *n.Rough {
		t.Errorf("%s: range %+v rough %v, want %s to %s and false", label, n.Range, n.Rough, low, high)
	}
}

func wantPace(t *testing.T, label string, s estSnap, project string, pace, low, high float64) {
	t.Helper()
	r := s.Releases[s.release(t, project)]
	if r.Pace != pace || r.PaceLow == nil || *r.PaceLow != low || r.PaceHigh == nil || *r.PaceHigh != high || r.Rough {
		t.Errorf("%s: %s pace %v ends %v %v rough %v, want %v %v %v", label, project, r.Pace, r.PaceLow, r.PaceHigh, r.Rough, pace, low, high)
	}
}

func wantList(t *testing.T, what string, got, want []string) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("%s = %v, want %v", what, got, want)
	}
}

type forecastLine struct {
	Format  int     `json:"format"`
	Made    string  `json:"made"`
	Project string  `json:"project"`
	Round   string  `json:"round"`
	Time    *string `json:"time"`
}

func forecastLines(t *testing.T, dir string) []forecastLine {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "forecasts.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var out []forecastLine
	for line := range strings.SplitSeq(strings.TrimSuffix(string(data), "\n"), "\n") {
		var f forecastLine
		if err := json.Unmarshal([]byte(line), &f); err != nil {
			t.Fatalf("forecasts.jsonl line %q: %v", line, err)
		}
		out = append(out, f)
	}
	return out
}

// started is a copy of the fixture with beta's b3 designed at 2:00 pm and
// p13 started at 2:00 pm. The snapshot at that clock is returned with it.
func started(t *testing.T) (string, estSnap) {
	t.Helper()
	dir := newState(t)
	do(t, dir, at200, "round", "design", "beta", "b3", input(t, "design-b3.json"))
	do(t, dir, at200, "pass", "start", "alpha", "p13")
	return dir, readEst(t, dir, at200)
}

func TestEstimate(t *testing.T) {
	t.Run("table", func(t *testing.T) {
		dir, s := started(t)
		check := func(label string, s estSnap, tm, low, high string) {
			t.Helper()
			wantSession(t, label, s, "scheduled", tm, low, high)
			n := s.NextSession
			if n.QuestionCount != 2 {
				t.Errorf("%s: question_count %d, want 2", label, n.QuestionCount)
			}
			wantList(t, label+" projects", n.Projects, []string{"alpha", "beta"})
			wantList(t, label+" rounds", n.Rounds, []string{"alpha/r2", "beta/b3"})
		}

		check("2:00 pm", s, "2026-11-21T05:00:00Z", "2026-11-21T03:00:00Z", "2026-11-21T06:30:00Z")
		wantPace(t, "2:00 pm", s, "alpha", 1.2, 0.8, 1.5)
		b := s.Releases[s.release(t, "beta")]
		if b.Pace != 1.0 || !b.Rough || b.PaceLow != nil || b.PaceHigh != nil {
			t.Errorf("beta pace %v rough %v ends %v %v, want 1.0, rough, null", b.Pace, b.Rough, b.PaceLow, b.PaceHigh)
		}
		var found bool
		for _, rd := range b.Rounds {
			for _, p := range rd.Passes {
				if p.Name == "b-four" {
					found = true
					if str(p.NeededTime) != "2026-11-21T07:00:00Z" {
						t.Errorf("b-four needed_time %s, want 2026-11-21T07:00:00Z", str(p.NeededTime))
					}
				}
			}
		}
		if !found {
			t.Error("b-four is not in the snapshot")
		}

		check("2:30 pm", readEst(t, dir, at230), "2026-11-21T05:00:00Z", "2026-11-21T03:00:00Z", "2026-11-21T06:30:00Z")

		slow := newState(t)
		do(t, slow, at200, "round", "design", "beta", "b3", input(t, "design-b3.json"))
		do(t, slow, at200, "pass", "start", "alpha", "p13")

		do(t, dir, at250, "pass", "finish", "alpha", "p13")
		fast := readEst(t, dir, at250)
		check("2:50 pm", fast, "2026-11-21T03:30:00Z", "2026-11-21T01:30:00Z", "2026-11-21T04:00:00Z")
		wantPace(t, "after the fast finish", fast, "alpha", 1.1, 0.5, 1.25)

		do(t, slow, at520, "pass", "finish", "alpha", "p13")
		late := readEst(t, slow, at520)
		check("5:20 pm", late, "2026-11-21T06:30:00Z", "2026-11-21T05:00:00Z", "2026-11-21T09:00:00Z")
		wantPace(t, "after the slow finish", late, "alpha", 1.25, 0.8, 2.0)
	})

	t.Run("forecast_lines", func(t *testing.T) {
		for _, tc := range []struct{ name, finish string }{{"fast", at250}, {"slow", at520}} {
			dir, startSnap := started(t)
			do(t, dir, tc.finish, "pass", "finish", "alpha", "p13")
			s := readEst(t, dir, tc.finish)
			lines := forecastLines(t, dir)
			if len(lines) != 3 {
				t.Fatalf("%s: forecasts.jsonl has %d lines, want the fixture line and 2", tc.name, len(lines))
			}
			for i, want := range []struct {
				made string
				time string
			}{{at200, str(startSnap.NextSession.Time)}, {tc.finish, str(s.NextSession.Time)}} {
				l := lines[i+1]
				if l.Made != want.made || str(l.Time) != want.time || l.Project != "alpha" || l.Round != "r2" {
					t.Errorf("%s line %d = %+v, want made %s time %s for alpha/r2", tc.name, i+2, l, want.made, want.time)
				}
			}
		}
	})

	t.Run("variant", func(t *testing.T) {
		dir := newState(t)
		path := filepath.Join(dir, "releases", "beta.json")
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		changed := strings.Replace(string(data), `"planned_minutes": 480`, `"planned_minutes": 60`, 1)
		if changed == string(data) {
			t.Fatal("beta.json holds no pass planned at 480 minutes")
		}
		if err := os.WriteFile(path, []byte(changed), 0o644); err != nil {
			t.Fatal(err)
		}
		do(t, dir, at200, "round", "design", "beta", "b3", input(t, "design-b3.json"))
		do(t, dir, at200, "pass", "start", "alpha", "p13")
		s := readEst(t, dir, at200)
		wantSession(t, "variant", s, "scheduled", "2026-11-21T00:00:00Z", "", "")
		lines := forecastLines(t, dir)
		last := lines[len(lines)-1]
		if last.Project != "beta" || last.Round != "b3" || str(last.Time) != "2026-11-21T00:00:00Z" {
			t.Errorf("forecast line = %+v, want beta/b3 at 2026-11-21T00:00:00Z", last)
		}
	})

	t.Run("not_designed", func(t *testing.T) {
		dir := newState(t)
		do(t, dir, at200, "pass", "start", "alpha", "p13")
		s := readEst(t, dir, at200)
		b := s.Releases[s.release(t, "beta")]
		var foundB3 bool
		for _, rd := range b.Rounds {
			if rd.Name != "b3" {
				continue
			}
			foundB3 = true
			if rd.State != "not_designed" || rd.Needs.Kind != "not_designed" || rd.MinutesBefore == nil || *rd.MinutesBefore != 480 || str(rd.ForecastTime) != "2026-11-21T07:00:00Z" {
				t.Errorf("b3 = %+v, want not_designed, 480 minutes before, forecast 2026-11-21T07:00:00Z", rd)
			}
		}
		if !foundB3 {
			t.Fatal("b3 is not in the snapshot")
		}
		n := s.NextSession
		wantList(t, "not_designed", n.NotDesigned, []string{"beta/b3"})
		wantList(t, "projects", n.Projects, []string{"alpha"})
		if n.QuestionCount != 1 || str(n.Time) != "2026-11-21T05:00:00Z" {
			t.Errorf("question_count %d time %s, want 1 and 2026-11-21T05:00:00Z", n.QuestionCount, str(n.Time))
		}
	})

	t.Run("waiting_on_design", func(t *testing.T) {
		dir := newState(t)
		do(t, dir, at200, "pass", "start", "beta", "b-three")
		do(t, dir, "2026-11-20T23:10:00Z", "pass", "finish", "beta", "b-three")
		s := readEst(t, dir, "2026-11-20T23:10:00Z")
		n := s.NextSession
		wantList(t, "waiting_on_design", n.WaitingOnDesign, []string{"beta/b3"})
		wantList(t, "not_designed", n.NotDesigned, []string{})
		if str(n.Time) != "2026-11-21T05:10:00Z" {
			t.Errorf("time %s, want alpha's 2026-11-21T05:10:00Z", str(n.Time))
		}
	})

	t.Run("other_states", func(t *testing.T) {
		dir := newState(t)
		do(t, dir, at200, "pass", "start", "alpha", "p13")
		do(t, dir, "2026-11-20T23:10:00Z", "pass", "stop", "alpha", "p13", "--reason", "scope")
		if got := readEst(t, dir, "2026-11-20T23:10:00Z").NextSession.State; got != "needs_you_now" {
			t.Errorf("stopped pass: state %s, want needs_you_now", got)
		}

		dir = newState(t)
		do(t, dir, at200, "pass", "start", "alpha", "p13")
		do(t, dir, at250, "pass", "finish", "alpha", "p13")
		do(t, dir, at250, "pass", "start", "alpha", "p14")
		do(t, dir, "2026-11-21T01:00:00Z", "pass", "finish", "alpha", "p14")
		s := readEst(t, dir, "2026-11-21T01:00:00Z")
		wantSession(t, "ready", s, "ready_now", "2026-11-21T01:00:00Z", "2026-11-21T01:00:00Z", "2026-11-21T01:00:00Z")
		if len(s.WaitsForGeoff) != 1 || s.WaitsForGeoff[0].Project != "alpha" || s.WaitsForGeoff[0].Round != "r2" || s.WaitsForGeoff[0].Questions != 1 {
			t.Errorf("waits_for_geoff = %+v, want alpha r2 with 1", s.WaitsForGeoff)
		}
	})
}
