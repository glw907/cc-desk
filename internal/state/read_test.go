package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const goodRelease = `{
  "format": 1, "project": "alpha", "name": "v1", "repo": "glw907/alpha",
  "rounds": [{"name": "r1", "summary": "First.", "needs_geoff": true}],
  "passes": [{"name": "p1", "summary": "Build.", "round": "r1", "planned_minutes": 60, "state": "not_started"}]
}`

func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestReadEmptyFolder(t *testing.T) {
	s, err := Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Releases) != 0 || len(s.Rulings) != 0 || s.Usage != nil {
		t.Errorf("empty folder loaded %+v", s)
	}
	if s.Settings.UsageStopPercent != 100 {
		t.Errorf("absent settings.json gave stop %d, want 100", s.Settings.UsageStopPercent)
	}
}

func TestReadGood(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "releases/alpha.json", goodRelease)
	writeFile(t, dir, "settings.json", `{"format":1,"usage_stop_percent":90}`)
	writeFile(t, dir, "rulings.jsonl", `{"format":1,"id":"a","time":"2026-11-20T23:00:00Z","desk":"d","words":"go"}`)
	s, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Releases) != 1 || s.Releases[0].Path != "releases/alpha.json" {
		t.Errorf("releases = %+v", s.Releases)
	}
	if s.Settings.UsageStopPercent != 90 || len(s.Rulings) != 1 {
		t.Errorf("settings %d, rulings %d", s.Settings.UsageStopPercent, len(s.Rulings))
	}
}

func TestReadFailures(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		content string
		want    string // the file name and a fragment the error must hold
	}{
		{"syntax", "releases/alpha.json", `{"format":1,`, "releases/alpha.json: unexpected EOF"},
		{"empty file", "releases/alpha.json", ``, "releases/alpha.json: syntax error: the file is empty"},
		{"trailing value", "releases/alpha.json", goodRelease + `}`, "releases/alpha.json: syntax error: data after"},
		{"trailing second value", "releases/alpha.json", goodRelease + goodRelease, "releases/alpha.json: syntax error: data after"},
		{"unknown field", "releases/alpha.json", strings.Replace(goodRelease, `"format": 1,`, `"format": 1, "extra": 1,`, 1), `releases/alpha.json: json: unknown field "extra"`},
		{"missing format", "releases/alpha.json", strings.Replace(goodRelease, `"format": 1,`, ``, 1), `releases/alpha.json: missing required field "format"`},
		{"wrong format", "releases/alpha.json", strings.Replace(goodRelease, `"format": 1`, `"format": 2`, 1), "releases/alpha.json: format 2 is not supported"},
		{"missing project", "releases/alpha.json", strings.Replace(goodRelease, `"project": "alpha",`, ``, 1), `releases/alpha.json: missing required field "project"`},
		{"missing pass minutes", "releases/alpha.json", strings.Replace(goodRelease, `"planned_minutes": 60,`, ``, 1), `releases/alpha.json: pass "p1": missing required field "planned_minutes"`},
		{"name mismatch", "releases/beta.json", goodRelease, `releases/beta.json: the file name does not match project "alpha"`},
		{"unknown pass", "releases/alpha.json", strings.Replace(goodRelease, `"state": "not_started"`, `"state": "not_started", "after": ["nope"]`, 1), `pass "nope", which does not exist`},
		{"unknown question", "releases/alpha.json", strings.Replace(goodRelease, `"state": "not_started"`, `"state": "not_started", "questions": ["q9"]`, 1), `question "q9", which does not exist`},
		{"bad state", "releases/alpha.json", strings.Replace(goodRelease, `"not_started"`, `"paused"`, 1), `state "paused"`},
		{"done without finish", "releases/alpha.json", strings.Replace(goodRelease, `"not_started"`, `"done"`, 1), "needs start, finish, and actual minutes"},
		{"desks syntax", "desks.json", `{"format":1,"desks":[`, "desks.json: unexpected EOF"},
		{"settings syntax", "settings.json", `[]`, "settings.json: json: cannot unmarshal array"},
		{"settings range", "settings.json", `{"format":1,"usage_stop_percent":101}`, "settings.json: usage_stop_percent 101"},
		{"usage missing read_at", "usage.json", `{"format":1,"percent":10,"reset_text":"x"}`, `usage.json: missing required field "read_at"`},
		{"usage error with percent", "usage.json", `{"format":1,"error":"x","percent":10,"read_at":"2026-11-20T22:30:00Z"}`, `usage.json: an "error" reading carries no`},
		{"usage without reset text or error", "usage.json", `{"format":1,"percent":10,"read_at":"2026-11-20T22:30:00Z"}`, `usage.json: needs "reset_text"`},
		{"rulings blank line", "rulings.jsonl", `{"format":1,"id":"a","time":"2026-11-20T23:00:00Z","desk":"d","words":"go"}` + "\n\n", "rulings.jsonl line 2: syntax error: blank line"},
		{"rulings syntax on line 2", "rulings.jsonl", `{"format":1,"id":"a","time":"2026-11-20T23:00:00Z","desk":"d","words":"go"}` + "\n{bad\n", "rulings.jsonl line 2:"},
		{"rulings duplicate id", "rulings.jsonl", `{"format":1,"id":"a","time":"2026-11-20T23:00:00Z","desk":"d","words":"go"}` + "\n" + `{"format":1,"id":"a","time":"2026-11-20T23:00:00Z","desk":"d","words":"go"}`, `rulings.jsonl line 2: ruling id "a" appears twice`},
		{"forecasts incomplete", "forecasts.jsonl", `{"format":1}`, "forecasts.jsonl line 1:"},
		{"forecast project without time", "forecasts.jsonl", `{"format":1,"made":"2026-11-20T23:00:00Z","project":"a","round":"r","time":null}`, "forecasts.jsonl line 1:"},
		{"forecast time without round", "forecasts.jsonl", `{"format":1,"made":"2026-11-20T23:00:00Z","project":"a","time":"2026-11-21T00:00:00Z"}`, "forecasts.jsonl line 1:"},
		{"answer without ruling", "releases/alpha.json", strings.Replace(goodRelease, `"passes"`, `"questions": [{"id":"q1","text":"t","options":["a"],"recommended":"a","answer":{"ruling":"nope"}}], "passes"`, 1), `releases/alpha.json: question "q1" is answered by ruling "nope"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, dir, tt.file, tt.content)
			_, err := Load(dir)
			if err == nil {
				t.Fatal("Load accepted a broken file")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %q, want it to hold %q", err, tt.want)
			}
			if fe, ok := err.(*FileError); !ok || fe.File != tt.file {
				t.Errorf("error = %#v, want a FileError for %s", err, tt.file)
			}
		})
	}
}

func TestReadFirstFailureInLoadOrder(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "releases/alpha.json", `{bad`)
	writeFile(t, dir, "desks.json", `{bad`)
	writeFile(t, dir, "rulings.jsonl", `{bad`)
	_, err := Load(dir)
	if fe, ok := err.(*FileError); !ok || fe.File != "releases/alpha.json" {
		t.Fatalf("error = %v, want releases/alpha.json first", err)
	}
	writeFile(t, dir, "releases/alpha.json", goodRelease)
	_, err = Load(dir)
	if fe, ok := err.(*FileError); !ok || fe.File != "desks.json" {
		t.Fatalf("error = %v, want desks.json next", err)
	}
}

func TestReadSortsReleases(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "releases/beta.json", strings.Replace(goodRelease, `"alpha"`, `"beta"`, 1))
	writeFile(t, dir, "releases/alpha.json", goodRelease)
	writeFile(t, dir, "releases/done/old.json", strings.Replace(goodRelease, `"alpha"`, `"old"`, 1))
	s, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, rf := range s.Releases {
		got = append(got, rf.Path)
	}
	want := "releases/alpha.json releases/beta.json releases/done/old.json"
	if strings.Join(got, " ") != want {
		t.Errorf("order = %v, want %s", got, want)
	}
	if !s.Releases[2].Done || s.Releases[0].Done {
		t.Error("Done flags are wrong")
	}
}
