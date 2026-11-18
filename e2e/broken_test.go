package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// verb is one command a broken file must stop. Later passes append their
// verbs here.
type verb struct {
	name string
	args func(t *testing.T) []string
}

var verbs = []verb{
	{"snapshot", func(*testing.T) []string { return []string{"snapshot"} }},
	{"pass start", func(*testing.T) []string { return []string{"pass", "start", "alpha", "p13"} }},
	{"pass finish", func(*testing.T) []string { return []string{"pass", "finish", "alpha", "p13"} }},
	{"pass stop", func(*testing.T) []string { return []string{"pass", "stop", "alpha", "p13", "--reason", "scope"} }},
	{"usage", func(*testing.T) []string { return []string{"usage"} }},
	{"free", func(*testing.T) []string { return []string{"free"} }},
	{"ruling record", func(*testing.T) []string { return []string{"ruling", "record", "go"} }},
	{"release add", func(t *testing.T) []string { return []string{"release", "add", input(t, "gamma.json")} }},
	{"round design", func(t *testing.T) []string {
		return []string{"round", "design", "beta", "b3", input(t, "design-b3.json")}
	}},
}

func TestBrokenFile(t *testing.T) {
	cases := []struct {
		name, file, find, replace string
	}{
		{"syntax error in a release", "releases/alpha.json", "{\n  \"format\": 1,", "{\n  \"format\": 1,,"},
		{"unknown field", "releases/alpha.json", `"repo": "glw907/alpha",`, `"repo": "glw907/alpha", "bogus": 1,`},
		{"missing required field", "releases/alpha.json", `"repo": "glw907/alpha",`, ``},
		{"reference to a pass that doesn't exist", "releases/alpha.json", `"after": ["p13"]`, `"after": ["zz"]`},
		{"answer that points at no ruling", "releases/beta.json", `"ruling": "r-0001"`, `"ruling": "r-9999"`},
		{"loop", "releases/alpha.json", `"planned_minutes": 100, "state": "not_started"}`, `"planned_minutes": 100, "state": "not_started", "after": ["p14"]}`},
		{"syntax error in desks.json", "desks.json", `"format": 1,`, `"format": 1,,`},
		{"syntax error in settings.json", "settings.json", `"format": 1,`, `"format": 1,,`},
		{"syntax error in one line of rulings.jsonl", "rulings.jsonl", "\n", "\n{\"format\":1,\n"},
	}
	for _, c := range cases {
		for _, v := range verbs {
			t.Run(c.name+"/"+v.name, func(t *testing.T) {
				dir := newState(t)
				path := filepath.Join(dir, c.file)
				orig, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(orig), c.find) {
					t.Fatalf("fixture %s has no %q to change", c.file, c.find)
				}
				broken := strings.Replace(string(orig), c.find, c.replace, 1)
				writeIn(t, dir, c.file, broken)

				before := folder(t, dir)
				requireSnapshot(t, before)
				got := run(t, dir, workedClock, v.args(t)...)
				if got.code != 1 {
					t.Fatalf("exit %d, want 1; stderr: %s", got.code, got.stderr)
				}
				if !strings.Contains(got.stderr, c.file) {
					t.Errorf("stderr does not name %s: %s", c.file, got.stderr)
				}
				after := folder(t, dir)
				if after["snapshot.json"] != before["snapshot.json"] {
					t.Error("snapshot.json changed")
				}
				if after["snapshot-error.json"] == "" {
					t.Fatal("snapshot-error.json is absent")
				}
				if !strings.Contains(after["snapshot-error.json"], c.file) {
					t.Errorf("snapshot-error.json does not name %s: %s", c.file, after["snapshot-error.json"])
				}
				for name, content := range after {
					if name != "snapshot-error.json" && content != before[name] {
						t.Errorf("%s changed", name)
					}
				}

				writeIn(t, dir, c.file, string(orig))
				if ok := run(t, dir, workedClock, "snapshot"); ok.code != 0 {
					t.Fatalf("snapshot after the restore exited %d: %s", ok.code, ok.stderr)
				}
				if _, err := os.Stat(filepath.Join(dir, "snapshot-error.json")); err == nil {
					t.Error("snapshot-error.json survived the restore")
				}
			})
		}
	}
}
