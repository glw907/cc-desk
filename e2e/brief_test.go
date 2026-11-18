//go:build unix

package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const briefClock = "2026-11-20T23:00:00Z"

func TestBrief(t *testing.T) {
	t.Run("order_and_blocks", func(t *testing.T) {
		dir := newStateFrom(t, "desk")
		r := run(t, dir, briefClock, "brief")
		if r.code != 0 {
			t.Fatalf("brief exit %d: %s", r.code, r.stderr)
		}
		qb, merge, qa := strings.Index(r.stdout, "question qb"), strings.Index(r.stdout, "merge question merge-b4-1"), strings.Index(r.stdout, "question qa")
		if qb < 0 || merge < qb || qa < merge {
			t.Fatalf("blocks are not in order qb, merge, qa:\n%s", r.stdout)
		}
		for _, want := range []string{
			"holds up 70 planned minutes",
			"     A dropped connection looks the same as a stopped session.\n",
			"     restart: The session comes back by itself.\n",
			"     stop: The pass stops and waits for you.\n",
			"   Recommended: restart. A restart costs nothing and the pass is cheap to redo.\n",
			"   Never guessed.\n",
			"   Waits: b-after (30 minutes), b-tail (40 minutes)\n",
			`on pull request 12, "Add the report verbs", at commit 3f2a9c1, holds up 60 planned minutes`,
			"     merge: Merges on the check's next run if every check has passed and the last commit is still 3f2a9c1. Otherwise it asks again.\n",
			"     fix again: Sends the pass back with your words for one more attempt.",
			"     close: Closes the pull request, ends the session, and records the pass as closed by your ruling.\n",
			"   Recommended: merge.\n",
			"   Waits: a-free (10 minutes)\n",
			"Since the last desk session, 2026-11-20T21:30:00Z:\n",
			"relaunch: not relaunched: two tries used (reason: connection_dropped)\n",
		} {
			if !strings.Contains(r.stdout, want) {
				t.Errorf("brief lacks %q", want)
			}
		}
		if strings.Contains(r.stdout, "launched a1b2c3d4") {
			t.Errorf("brief holds a log line from before the desk started:\n%s", r.stdout)
		}
	})

	t.Run("forecast_order_differs", func(t *testing.T) {
		dir := newStateFrom(t, "desk")
		r := run(t, dir, briefClock, "snapshot")
		if r.code != 0 {
			t.Fatalf("snapshot exit %d: %s", r.code, r.stderr)
		}
		var snap struct {
			NextSession struct {
				Rounds []string `json:"rounds"`
			} `json:"next_session"`
		}
		if err := json.Unmarshal([]byte(r.stdout), &snap); err != nil {
			t.Fatal(err)
		}
		if got := snap.NextSession.Rounds; len(got) == 0 || got[0] != "alpha/r1" {
			t.Errorf("forecast puts %v first, want alpha/r1 (qa) so the brief's order is a different one", got)
		}
		b := run(t, newStateFrom(t, "desk"), briefClock, "brief")
		if !strings.Contains(b.stdout, "Next question forecast: 2026-11-20T23:00:00Z\n") || !strings.Contains(b.stdout, "(alpha / r1)") {
			t.Errorf("brief does not carry the forecast:\n%s", b.stdout)
		}
	})

	t.Run("round_form", func(t *testing.T) {
		dir := newStateFrom(t, "desk")
		r := run(t, dir, briefClock, "brief", "beta", "b1")
		if r.code != 0 {
			t.Fatalf("brief beta b1 exit %d: %s", r.code, r.stderr)
		}
		if !strings.Contains(r.stdout, "No design recorded yet.") || !strings.Contains(r.stdout, "question qb") || strings.Contains(r.stdout, "question qa") || strings.Contains(r.stdout, "merge-b4-1") {
			t.Errorf("round brief:\n%s", r.stdout)
		}
		for _, args := range [][]string{{"brief", "gamma", "b1"}, {"brief", "beta", "b9"}} {
			if r := run(t, dir, briefClock, args...); r.code != 1 || !strings.Contains(r.stderr, args[len(args)-1]+`"`) && !strings.Contains(r.stderr, args[1]) {
				t.Errorf("%v: exit %d, stderr %q, want exit 1 naming the missing part", args, r.code, r.stderr)
			}
		}
		if r := run(t, dir, briefClock, "brief", "beta"); r.code == 0 {
			t.Errorf("brief with one argument exited 0")
		}
	})

	t.Run("writes_nothing", func(t *testing.T) {
		dir := newStateFrom(t, "desk")
		before := listFiles(t, dir)
		run(t, dir, briefClock, "brief")
		after := listFiles(t, dir)
		if strings.Join(before, ",") != strings.Join(after, ",") {
			t.Errorf("brief changed the state folder:\nbefore %v\nafter  %v", before, after)
		}
	})

	t.Run("broken_file", func(t *testing.T) {
		dir := newStateFrom(t, "desk")
		bad := filepath.Join(dir, "releases", "alpha.json")
		if err := os.WriteFile(bad, []byte(`{"format":1,"surprise":true}`), 0o644); err != nil {
			t.Fatal(err)
		}
		before := listFiles(t, dir)
		r := run(t, dir, briefClock, "brief")
		if r.code != 1 || !strings.Contains(r.stderr, "alpha.json") {
			t.Errorf("exit %d, stderr %q, want exit 1 naming alpha.json", r.code, r.stderr)
		}
		if after := listFiles(t, dir); strings.Join(before, ",") != strings.Join(after, ",") {
			t.Errorf("a broken file made brief write: before %v, after %v", before, after)
		}
	})
}

// listFiles returns the paths under dir, sorted, with each file's size, so a
// write of any kind shows.
func listFiles(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		if rel == "lock" {
			return nil
		}
		out = append(out, rel+":"+info.ModTime().String()+":"+string(rune('0'+info.Size()%10)))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}
