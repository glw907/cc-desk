//go:build unix

package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// addAskRelease installs release proj with one pass, a, and the given
// ask_rules text ("" for none), then starts the pass.
func (r launchRig) addAskRelease(t *testing.T, ask string) {
	t.Helper()
	release := fmt.Sprintf(`{
  "format": 1, "project": "proj", "name": "v1", "repo": "glw907/proj",
  "allow_rules": [%q],%s
  "rounds": [{"name": "r1", "summary": "A round.", "needs_geoff": false, "no_need_reason": "A fixture.", "design_written": true}],
  "passes": [{"name": "a", "summary": "A pass.", "round": "r1", "planned_minutes": 30, "state": "not_started"}]
}`, launchAllow, ask)
	file := filepath.Join(r.dir, "release.json")
	if err := os.WriteFile(file, []byte(release), 0o644); err != nil {
		t.Fatal(err)
	}
	requireRun(t, run(t, r.state, workedClock, "release", "add", file))
	requireRun(t, run(t, r.state, workedClock, "release", "checkout", "proj", r.clone, "--default-branch", "main"))
	requireRun(t, r.do(t, workedClock, "pass", "start", "proj", "a"))
}

func TestAskRule(t *testing.T) {
	t.Run("a release with an ask rule launches with it", func(t *testing.T) {
		r := newLaunchRig(t)
		r.addAskRelease(t, `"ask_rules": ["Bash(date *)"],`)
		if got := r.run(t, workedClock, "check"); got.code != 0 {
			t.Fatalf("check exit %d, stderr: %s", got.code, got.stderr)
		}
		var s struct{ Permissions struct{ Allow, Ask []string } }
		if err := json.Unmarshal([]byte(readFileIn(t, r.state, "launch/proj-a.json")), &s); err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(s.Permissions.Ask, []string{"Bash(date *)"}) {
			t.Errorf("ask = %v, want the release's list", s.Permissions.Ask)
		}
		if slices.Contains(s.Permissions.Allow, "Bash(date *)") {
			t.Errorf("allow %v holds the ask rule", s.Permissions.Allow)
		}
	})

	t.Run("a release with none writes no ask key", func(t *testing.T) {
		r := newLaunchRig(t)
		r.addAskRelease(t, "")
		if got := r.run(t, workedClock, "check"); got.code != 0 {
			t.Fatalf("check exit %d, stderr: %s", got.code, got.stderr)
		}
		if got := readFileIn(t, r.state, "launch/proj-a.json"); strings.Contains(got, `"ask"`) {
			t.Errorf("settings hold an ask key: %s", got)
		}
	})

	t.Run("a rule in both lists is refused", func(t *testing.T) {
		r := newLaunchRig(t)
		file := filepath.Join(r.dir, "release.json")
		body := `{"format": 1, "project": "proj", "name": "v1", "repo": "glw907/proj",
  "allow_rules": ["Bash(ls)"], "ask_rules": ["Bash(ls)"]}`
		if err := os.WriteFile(file, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		got := run(t, r.state, workedClock, "release", "add", file)
		if got.code == 0 || !strings.Contains(got.stderr, "in both allow_rules and ask_rules") {
			t.Errorf("exit %d, stderr %q, want a refusal naming both lists", got.code, got.stderr)
		}
	})
}
