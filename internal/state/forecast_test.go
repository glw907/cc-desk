package state

import (
	"strings"
	"testing"
	"time"
)

func forecastRender(project, round string, at time.Time, setTime bool) Render {
	return func(*State, time.Time) ([]byte, Forecast, error) {
		fc := Forecast{}
		if setTime {
			fc.Project, fc.Round, fc.Time = project, round, &at
		}
		return []byte("{}\n"), fc, nil
	}
}

func TestPassEventMark(t *testing.T) {
	running := func(s *State, now time.Time) error { return StartPass("alpha", "p1")(s, now) }
	tests := []struct {
		name string
		ev   func(*State) Event
		want bool
	}{
		{"start", func(*State) Event { return StartPass("alpha", "p1") }, true},
		{"refused start", func(*State) Event { return StartPass("alpha", "nope") }, false},
		{"design", func(*State) Event { return func(*State, time.Time) error { return nil } }, false},
	}
	for _, tc := range tests {
		s := loaded(t)
		_ = tc.ev(s)(s, at(9, 0))
		if s.passEvent != tc.want {
			t.Errorf("%s: passEvent = %v, want %v", tc.name, s.passEvent, tc.want)
		}
	}

	s := loaded(t)
	if err := running(s, at(9, 0)); err != nil {
		t.Fatal(err)
	}
	s.passEvent = false
	if err := StopPass("alpha", "p1", StopScope)(s, at(9, 10)); err != nil || !s.passEvent {
		t.Errorf("stop: err %v passEvent %v", err, s.passEvent)
	}
	s.passEvent = false
	if err := StartPass("alpha", "p1")(s, at(9, 20)); err != nil || !s.passEvent {
		t.Errorf("resume: err %v passEvent %v", err, s.passEvent)
	}
	s.passEvent = false
	if err := FinishPass("alpha", "p1")(s, at(9, 40)); err != nil || !s.passEvent {
		t.Errorf("finish: err %v passEvent %v", err, s.passEvent)
	}
	s.passEvent = false
	if err := FinishPass("alpha", "p1")(s, at(9, 50)); err == nil || s.passEvent {
		t.Errorf("refused finish: err %v passEvent %v", err, s.passEvent)
	}
}

func TestCommitForecast(t *testing.T) {
	forecast := at(23, 0).Add(48 * time.Hour)
	start := StartPass("alpha", "p1")
	nothing := func(*State, time.Time) error { return nil }

	t.Run("pass event appends one line", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "releases/alpha.json", goodRelease)
		if _, err := Commit(dir, testClock, time.Second, start, forecastRender("alpha", "r1", forecast, true)); err != nil {
			t.Fatal(err)
		}
		want := `{"format":1,"made":"2026-11-20T23:00:00Z","project":"alpha","round":"r1","time":"2026-11-22T23:00:00Z"}` + "\n"
		if got := readFile(t, dir, ForecastFile); got != want {
			t.Errorf("forecasts.jsonl = %q, want %q", got, want)
		}
	})

	t.Run("no time writes null and no names", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "releases/alpha.json", goodRelease)
		if _, err := Commit(dir, testClock, time.Second, start, forecastRender("", "", forecast, false)); err != nil {
			t.Fatal(err)
		}
		want := `{"format":1,"made":"2026-11-20T23:00:00Z","time":null}` + "\n"
		if got := readFile(t, dir, ForecastFile); got != want {
			t.Errorf("forecasts.jsonl = %q, want %q", got, want)
		}
		if _, err := Load(dir); err != nil {
			t.Errorf("the line does not read back: %v", err)
		}
	})

	t.Run("other verbs append nothing", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "releases/alpha.json", goodRelease)
		if _, err := Commit(dir, testClock, time.Second, nothing, forecastRender("alpha", "r1", forecast, true)); err != nil {
			t.Fatal(err)
		}
		if exists(dir, ForecastFile) {
			t.Error("a verb that is not a pass event created forecasts.jsonl")
		}
	})

	t.Run("a refusal appends nothing", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "releases/alpha.json", goodRelease)
		if _, err := Commit(dir, testClock, time.Second, FinishPass("alpha", "p1"), forecastRender("alpha", "r1", forecast, true)); err == nil {
			t.Fatal("finishing a pass that has not started succeeded")
		}
		if exists(dir, ForecastFile) {
			t.Error("a refused event created forecasts.jsonl")
		}
	})

	t.Run("a missing final newline is added", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "releases/alpha.json", goodRelease)
		first := `{"format":1,"made":"2026-11-19T23:00:00Z","time":null}`
		writeFile(t, dir, ForecastFile, first)
		if _, err := Commit(dir, testClock, time.Second, start, forecastRender("alpha", "r1", forecast, true)); err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(strings.TrimSuffix(readFile(t, dir, ForecastFile), "\n"), "\n")
		if len(lines) != 2 || lines[0] != first {
			t.Errorf("lines = %q", lines)
		}
		if s, err := Load(dir); err != nil || len(s.Forecasts) != 2 {
			t.Errorf("load: %v, %d forecasts", err, len(s.Forecasts))
		}
	})
}
