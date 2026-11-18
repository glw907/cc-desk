package e2e

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// keyPaths returns every key path in one JSON document, joined with dots, with
// an array element shown by its index. A leaf is a scalar, an empty object, or
// an empty array.
func keyPaths(t *testing.T, name string, data []byte) map[string]bool {
	t.Helper()
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	out := map[string]bool{}
	var walk func(prefix string, v any)
	walk = func(prefix string, v any) {
		switch x := v.(type) {
		case map[string]any:
			if len(x) == 0 {
				out[prefix] = true
			}
			for k, c := range x {
				walk(prefix+"."+k, c)
			}
		case []any:
			if len(x) == 0 {
				out[prefix] = true
			}
			for i, c := range x {
				walk(prefix+"."+strconv.Itoa(i), c)
			}
		default:
			out[prefix] = true
		}
	}
	walk(name, v)
	return out
}

// folderKeyPaths walks every JSON and JSONL file under dir, leaving out the
// files named in skip, and returns each file's key paths. A JSONL line is
// shown as an index after the file name.
func folderKeyPaths(t *testing.T, dir string, skip ...string) map[string]bool {
	t.Helper()
	out := map[string]bool{}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		if slices.Contains(skip, rel) {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		switch filepath.Ext(rel) {
		case ".json":
			for k := range keyPaths(t, rel, data) {
				out[k] = true
			}
		case ".jsonl":
			for i, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
				if line == "" {
					continue
				}
				for k := range keyPaths(t, fmt.Sprintf("%s#%d", rel, i), []byte(line)) {
					out[k] = true
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// TestRealCopy reads a copy of the real state folder taken on 2026-11-18,
// before any record of this round existed, and checks that one verb that
// rewrites its release keeps every field the file held.
func TestRealCopy(t *testing.T) {
	dir := newStateFrom(t, "real-2026-11-18")
	var check struct {
		LastRun string `json:"last_run"`
	}
	data, err := os.ReadFile(filepath.Join(dir, "check.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &check); err != nil || check.LastRun == "" {
		t.Fatalf("check.json last_run: %q, %v", check.LastRun, err)
	}
	now := check.LastRun
	if r := run(t, dir, now, "snapshot"); r.code != 0 {
		t.Fatalf("snapshot exit %d: %s", r.code, r.stderr)
	}

	before := folderKeyPaths(t, dir, "snapshot.json")
	rel := filepath.Join(dir, "releases", "cc-desk-scratch.json")
	bytesBefore, err := os.ReadFile(rel)
	if err != nil {
		t.Fatal(err)
	}
	if r := run(t, dir, now, "release", "checkout", "cc-desk-scratch", t.TempDir(), "--default-branch", "main-copytest"); r.code != 0 {
		t.Fatalf("release checkout exit %d: %s", r.code, r.stderr)
	}
	bytesAfter, err := os.ReadFile(rel)
	if err != nil {
		t.Fatal(err)
	}
	if string(bytesBefore) == string(bytesAfter) {
		t.Fatal("release file bytes did not change, so the verb rewrote nothing")
	}
	after := folderKeyPaths(t, dir, "snapshot.json")
	if len(before) < 50 {
		t.Fatalf("only %d key paths before; the copy looks empty", len(before))
	}
	for k := range before {
		if strings.HasPrefix(k, "check.json.") || strings.HasPrefix(k, "usage.json.") {
			continue
		}
		if !after[k] {
			t.Errorf("key path %s is gone after the verb", k)
		}
	}
}

// TestRecords holds one subtest for each record the desk round adds. Each
// later pass of the round adds its own.
func TestRecords(t *testing.T) {
	t.Run("question_fields", func(t *testing.T) {
		dir := newState(t)
		in := input(t, "question-q9.json")
		if r := run(t, dir, workedClock, "question", "add", "alpha", "r2", in); r.code != 0 {
			t.Fatalf("question add exit %d: %s", r.code, r.stderr)
		}
		files := folder(t, dir)
		var snap any
		if err := json.Unmarshal([]byte(files["snapshot.json"]), &snap); err != nil {
			t.Fatal(err)
		}
		// A question is printed wherever a pass or a need lists it, so find it by its id.
		var found []map[string]any
		var walk func(v any)
		walk = func(v any) {
			switch x := v.(type) {
			case map[string]any:
				if x["id"] == "alpha-q9" {
					found = append(found, x)
				}
				for _, c := range x {
					walk(c)
				}
			case []any:
				for _, c := range x {
					walk(c)
				}
			}
		}
		walk(snap)
		if len(found) == 0 {
			t.Fatal("alpha-q9 is not in the snapshot")
		}
		for _, q := range found {
			bg, _ := q["background"].([]any)
			causes, _ := q["causes"].(map[string]any)
			if len(bg) != 2 || len(causes) != 2 || q["reason"] != "Why the recommendation." || q["never_guess"] != true {
				t.Errorf("question fields in the snapshot = %v", q)
			}
		}

		if r := run(t, dir, workedClock, "question", "add", "alpha", "r2", in); r.code != 2 {
			t.Errorf("a second add exit %d, want 2: %s", r.code, r.stderr)
		}
		before := folder(t, dir)

		bad := filepath.Join(t.TempDir(), "bad.json")
		if err := os.WriteFile(bad, []byte(`{"id":"alpha-q8","text":"T?","options":["a","b"],"recommended":"a","passes":["p14"]}`), 0o644); err != nil {
			t.Fatal(err)
		}
		r := run(t, dir, workedClock, "question", "add", "alpha", "r2", bad)
		if r.code == 0 || !strings.Contains(r.stderr, `"background"`) {
			t.Errorf("a file with no background: exit %d, stderr %q", r.code, r.stderr)
		}
		sameFolder(t, before, folder(t, dir))
	})

	t.Run("question_never_guess", func(t *testing.T) {
		dir := newState(t)
		file := filepath.Join(t.TempDir(), "q.json")
		q := `{"id":"alpha-q9","text":"T?","options":["yes","no"],"recommended":"yes","background":["B."],"causes":{"yes":"Y.","no":"N."},"reason":"R.","never_guess":true,"passes":["p13"]}`
		if err := os.WriteFile(file, []byte(q), 0o644); err != nil {
			t.Fatal(err)
		}
		if r := run(t, dir, workedClock, "question", "add", "alpha", "r2", file); r.code != 0 {
			t.Fatalf("question add exit %d: %s", r.code, r.stderr)
		}
		before := folder(t, dir)
		r := run(t, dir, workedClock, "pass", "start", "alpha", "p13", "--guess", "alpha-q9")
		if r.code != 2 || !strings.Contains(r.stderr, "alpha-q9") || !strings.Contains(r.stderr, "never guessed") {
			t.Errorf("guess start: exit %d, stderr %q", r.code, r.stderr)
		}
		sameFolder(t, before, folder(t, dir))
	})

	t.Run("needs_you_now", func(t *testing.T) {
		dir := newStateFrom(t, "desk")
		r := run(t, dir, workedClock, "snapshot")
		if r.code != 0 {
			t.Fatalf("snapshot exit %d: %s", r.code, r.stderr)
		}
		var snap struct {
			NeedsYouNow []struct {
				Project string  `json:"project"`
				Round   *string `json:"round"`
				Reason  string  `json:"reason"`
			} `json:"needs_you_now"`
			Sessions []struct {
				Name      string  `json:"name"`
				StoppedAt *string `json:"stopped_at"`
				Refusals  []struct {
					Tool   string `json:"tool"`
					Target string `json:"target"`
				} `json:"refusals"`
			} `json:"sessions"`
			Releases []struct {
				Project string `json:"project"`
				Rounds  []struct {
					Passes []struct {
						Name        string  `json:"name"`
						StoppedAt   *string `json:"stopped_at"`
						NeedsKind   string  `json:"needs_kind"`
						NeedsReason *string `json:"needs_reason"`
					} `json:"passes"`
				} `json:"rounds"`
			} `json:"releases"`
		}
		if err := json.Unmarshal([]byte(r.stdout), &snap); err != nil {
			t.Fatal(err)
		}

		got := map[string]string{}
		for _, n := range snap.NeedsYouNow {
			got[n.Project] = n.Reason
		}
		// alpha's conductor ended on an expired sign-in, so only its stopped
		// pass asks; beta's design writer ended on a failed launch, which stays.
		if len(snap.NeedsYouNow) != 2 || got["alpha"] != "scope" || got["beta"] != "launch_failed" {
			t.Errorf("needs_you_now = %+v, want alpha scope and beta launch_failed", snap.NeedsYouNow)
		}

		for _, s := range snap.Sessions {
			if s.Name == "alpha-conductor" {
				if s.StoppedAt == nil || len(s.Refusals) != 1 || s.Refusals[0].Tool != "Bash" || s.Refusals[0].Target != "rm -r build" {
					t.Errorf("alpha-conductor entry = %+v", s)
				}
			}
			if s.Refusals == nil {
				t.Errorf("session %s prints refusals as null, want a list", s.Name)
			}
		}

		for _, rel := range snap.Releases {
			for _, rd := range rel.Rounds {
				for _, p := range rd.Passes {
					switch p.Name {
					case "a-stopped":
						if p.StoppedAt == nil || *p.StoppedAt != "2026-11-20T22:40:00Z" || p.NeedsKind != "nothing" ||
							p.NeedsReason == nil || *p.NeedsReason != "It only renames files." {
							t.Errorf("a-stopped = %+v", p)
						}
					case "a-free":
						if p.StoppedAt != nil || p.NeedsKind != "questions" || p.NeedsReason != nil {
							t.Errorf("a-free = %+v", p)
						}
					}
				}
			}
		}
	})
}
