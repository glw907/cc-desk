package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const goodSession = `{"name":"p-one","kind":"pass","short_id":"ab12","session_id":"s-1","worktree":"/w/one","branch":"one","launched":"2026-11-20T10:00:00Z","launch_tries":1}`

// sessionRelease builds a release with one pass in the given state carrying
// session JSON, plus extra top-level fields.
func sessionRelease(state, session, extra string) string {
	s := ""
	if session != "" {
		s = `,"session":` + session
	}
	return `{"format":1,"project":"p","name":"n","repo":"r"` + extra + `,"rounds":[{"name":"a","summary":"s","needs_geoff":true}],` +
		`"passes":[{"name":"x","summary":"s","round":"a","planned_minutes":10,"state":"` + state + `"` + started(state) + s + `}]}`
}

// started gives a building pass the start time the checks need.
func started(state string) string {
	if state == "building" {
		return `,"started":"2026-11-20T10:00:00Z"`
	}
	return ""
}

func TestCheckSessions(t *testing.T) {
	tests := []struct {
		name string
		json string
		want string // empty means no error
	}{
		{"building with a session", sessionRelease("building", goodSession, ""), ""},
		{"not started with a session", sessionRelease("not_started", goodSession, ""), "carries a session record"},
		{"wrong kind", sessionRelease("building", strings.Replace(goodSession, `"pass"`, `"design"`, 1), ""), `session kind "design"`},
		{"relative worktree", sessionRelease("building", strings.Replace(goodSession, "/w/one", "w/one", 1), ""), "not an absolute path"},
		{"missing branch", sessionRelease("building", strings.Replace(goodSession, `"branch":"one",`, "", 1), ""), `missing "branch"`},
		{"launch tries 3", sessionRelease("building", strings.Replace(goodSession, `"launch_tries":1`, `"launch_tries":3`, 1), ""), "not 1 or 2"},
		{"unknown key", sessionRelease("building", strings.Replace(goodSession, `"kind"`, `"extra":1,"kind"`, 1), ""), "unknown field"},
		{"stop reason without time", sessionRelease("building", strings.Replace(goodSession, `"launch_tries":1`, `"launch_tries":1,"stop_reason":"usage"`, 1), ""), "set together"},
		{"stop on a pass session", sessionRelease("building", strings.Replace(goodSession, `"launch_tries":1`, `"launch_tries":1,"stop_reason":"usage","stopped":"2026-11-20T11:00:00Z"`, 1), ""), "stays on the pass"},
		{"checkout alone", sessionRelease("not_started", "", `,"checkout":"/c"`), "set together"},
		{"checkout relative", sessionRelease("not_started", "", `,"checkout":"c","default_branch":"main"`), "not an absolute path"},
		{"checkout pair", sessionRelease("not_started", "", `,"checkout":"/c","default_branch":"main"`), ""},
		{"pull request zero", strings.Replace(sessionRelease("building", goodSession, ""), `"session"`, `"pull_request":{"number":0},"session"`, 1), `needs a "number"`},
		{"relaunch without cause", sessionRelease("building", strings.Replace(goodSession, `"launch_tries":1`, `"launch_tries":1,"relaunches":[{"time":"2026-11-20T10:30:00Z","cause":"","counted":true}]`, 1), ""), `needs "time" and "cause"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeRelease([]byte(tt.json))
			switch {
			case tt.want == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)):
				t.Fatalf("error = %v, want it to hold %q", err, tt.want)
			}
		})
	}
}

func TestStopReasonsAreTen(t *testing.T) {
	if len(StopReasons) != 10 {
		t.Fatalf("StopReasons has %d entries, want 10", len(StopReasons))
	}
	for _, r := range []StopReason{StopPermissionPrompt, StopSignInExpired, StopUnreadable, StopRelaunchLimit, StopLaunchFailed} {
		if !ValidStopReason(r) {
			t.Errorf("%q is not valid", r)
		}
	}
	if ValidStopReason("nope") {
		t.Error(`"nope" is valid`)
	}
}

// writeRelease puts json at releases/<project>.json under a new folder.
func writeRelease(t *testing.T, project, json string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "releases"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "releases", project+".json"), []byte(json), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestLoadSessionsUnique(t *testing.T) {
	dir := writeRelease(t, "p", sessionRelease("building", goodSession, ""))
	other := strings.Replace(sessionRelease("building", goodSession, ""), `"project":"p"`, `"project":"q"`, 1)
	if err := os.WriteFile(filepath.Join(dir, "releases", "q.json"), []byte(other), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(dir)
	if err == nil || !strings.Contains(err.Error(), "name one worktree") {
		t.Fatalf("error = %v, want two records on one worktree", err)
	}
}

func TestNameUniqueAndEndedExempt(t *testing.T) {
	second := strings.Replace(strings.Replace(goodSession, "/w/one", "/w/two", 1), `"branch":"one"`, `"branch":"two"`, 1)
	tests := []struct {
		name string
		edit func(string) string
		want string
	}{
		{"one session name", func(s string) string { return s }, "name one session"},
		{"ended record is exempt", func(s string) string {
			return strings.Replace(s, `"launch_tries":1`, `"launch_tries":1,"ended":"2026-11-20T11:00:00Z"`, 1)
		}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := writeRelease(t, "p", sessionRelease("building", tt.edit(goodSession), ""))
			other := strings.Replace(sessionRelease("building", second, ""), `"project":"p"`, `"project":"q"`, 1)
			if err := os.WriteFile(filepath.Join(dir, "releases", "q.json"), []byte(other), 0o644); err != nil {
				t.Fatal(err)
			}
			_, err := Load(dir)
			switch {
			case tt.want == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)):
				t.Fatalf("error = %v, want it to hold %q", err, tt.want)
			}
		})
	}
}

func TestConductorAndDesignRecords(t *testing.T) {
	conductor := strings.Replace(goodSession, `"kind":"pass"`, `"kind":"conductor"`, 1)
	design := strings.Replace(goodSession, `"kind":"pass"`, `"kind":"design"`, 1)
	stopped := `,"stop_reason":"scope","stopped":"2026-11-20T11:00:00Z"`
	tests := []struct {
		name string
		json string
		want string
	}{
		{"conductor", sessionRelease("not_started", "", `,"conductor":`+conductor), ""},
		{"conductor with a stop", sessionRelease("not_started", "", `,"conductor":`+strings.Replace(conductor, `"launch_tries":1`, `"launch_tries":1`+stopped, 1)), ""},
		{"conductor of the wrong kind", sessionRelease("not_started", "", `,"conductor":`+goodSession), `session kind "pass"`},
		{"design writer", strings.Replace(sessionRelease("not_started", "", ""), `"needs_geoff":true`, `"needs_geoff":true,"session":`+design, 1), ""},
		{"design writer of the wrong kind", strings.Replace(sessionRelease("not_started", "", ""), `"needs_geoff":true`, `"needs_geoff":true,"session":`+conductor, 1), `session kind "conductor"`},
		{"unknown stop reason", sessionRelease("not_started", "", `,"conductor":`+strings.Replace(conductor, `"launch_tries":1`, `"launch_tries":1,"stop_reason":"bored","stopped":"2026-11-20T11:00:00Z"`, 1)), "not a stop reason"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeRelease([]byte(tt.json))
			switch {
			case tt.want == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)):
				t.Fatalf("error = %v, want it to hold %q", err, tt.want)
			}
		})
	}
}

func TestNewReleaseCarriesNoSessionFields(t *testing.T) {
	conductor := strings.Replace(goodSession, `"kind":"pass"`, `"kind":"conductor"`, 1)
	tests := []struct {
		name  string
		extra string
	}{
		{"checkout", `,"checkout":"/c","default_branch":"main"`},
		{"conductor", `,"conductor":` + conductor},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := DecodeRelease([]byte(sessionRelease("not_started", "", tt.extra)))
			if err != nil {
				t.Fatal(err)
			}
			err = AddRelease(r)(&State{}, time.Now())
			if err == nil || !strings.Contains(err.Error(), "carries no checkout") {
				t.Fatalf("error = %v, want a refusal", err)
			}
		})
	}
	t.Run("a pass with a session record", func(t *testing.T) {
		r, err := DecodeRelease([]byte(sessionRelease("building", goodSession, "")))
		if err != nil {
			t.Fatal(err)
		}
		r.Passes[0].State, r.Passes[0].Started = NotStarted, nil
		err = AddRelease(r)(&State{}, time.Now())
		if err == nil || !strings.Contains(err.Error(), "not_started with no events") {
			t.Fatalf("error = %v, want a refusal", err)
		}
	})
	t.Run("a designed pass with a pull request", func(t *testing.T) {
		rf := &ReleaseFile{Release: Release{Project: "p", Rounds: []Round{{Name: "a"}}}}
		s := &State{Releases: []*ReleaseFile{rf}}
		d := &Design{Passes: []Pass{{Name: "y", Round: "a", State: NotStarted, PullRequest: &PullRequest{Number: 1}}}}
		err := DesignRound("p", "a", d)(s, time.Now())
		if err == nil || !strings.Contains(err.Error(), "not_started with no events") {
			t.Fatalf("error = %v, want a refusal", err)
		}
	})
}

func TestLoadCheckFile(t *testing.T) {
	rel := sessionRelease("not_started", "", "")
	tests := []struct {
		name string
		file string // empty means no check.json
		want string
	}{
		{"absent", "", ""},
		{"good", `{"format":1,"list_failures":2,"decisions":[]}`, ""},
		{"no format", `{"list_failures":0}`, `"format"`},
		{"negative failures", `{"format":1,"list_failures":-1}`, "below 0"},
		{"unknown key", `{"format":1,"other":1}`, "unknown field"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := writeRelease(t, "p", rel)
			if tt.file != "" {
				if err := os.WriteFile(filepath.Join(dir, CheckStateFile), []byte(tt.file), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			s, err := Load(dir)
			switch {
			case tt.want == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)):
				t.Fatalf("error = %v, want it to hold %q", err, tt.want)
			case tt.want == "" && s.Check.Decisions == nil:
				t.Fatal("Decisions is nil, want an empty list")
			}
		})
	}
}
