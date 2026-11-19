//go:build unix

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSkillBriefs keeps the skills' live briefs in step with the brief verb:
// each fixture's brief, printed on a copy at the worked clock, must equal the
// file the live script hands its session.
func TestSkillBriefs(t *testing.T) {
	for _, tc := range []struct {
		name  string
		wants []string
	}{
		{"none", []string{"Next question forecast: 2026-11-21T07:00:00Z\n", "Waiting: nothing\n"}},
		{"one", []string{"Waiting: 1 question\n", "question q-juneau-7, holds up 85 planned minutes\n"}},
		{"two", []string{"Waiting: 2 questions\n", "1. dock / d1, merge question merge-dock-2 on pull request 4817", "2. dock / d2, question q-ferry-3"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := newStateFrom(t, "skills-"+tc.name)
			r := run(t, dir, workedClock, "brief")
			if r.code != 0 {
				t.Fatalf("brief exit %d: %s", r.code, r.stderr)
			}
			want, err := os.ReadFile(filepath.Join("..", "scripts", "live", "skills", "brief-"+tc.name+".txt"))
			if err != nil {
				t.Fatal(err)
			}
			if r.stdout != string(want) {
				t.Errorf("the brief differs from brief-%s.txt; print it again from a copy of the fixture.\ngot:\n%s", tc.name, r.stdout)
			}
			for _, w := range tc.wants {
				if !strings.Contains(r.stdout, w) {
					t.Errorf("brief lacks %q", w)
				}
			}
		})
	}
}
