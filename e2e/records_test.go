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

	t.Run("report_add", func(t *testing.T) {
		dir := newState(t)
		in := input(t, "report.json")
		r := run(t, dir, workedClock, "report", "add", "alpha", in, "--round", "r1")
		if r.code != 0 || !strings.Contains(r.stderr, "recorded report rep-0001") {
			t.Fatalf("report add exit %d: %s", r.code, r.stderr)
		}
		files := folder(t, dir)
		var saved struct {
			Format  int `json:"format"`
			Reports []struct {
				ID, Project, Round, Title string
				Calls, Misses             []any
				Read                      *string
			} `json:"reports"`
		}
		if err := json.Unmarshal([]byte(files["reports.json"]), &saved); err != nil {
			t.Fatal(err)
		}
		if saved.Format != 1 || len(saved.Reports) != 1 || saved.Reports[0].Round != "r1" || saved.Reports[0].Read != nil ||
			len(saved.Reports[0].Calls) != 1 || len(saved.Reports[0].Misses) != 1 {
			t.Fatalf("reports.json = %s", files["reports.json"])
		}

		var snap struct {
			Reports []struct {
				ID     string  `json:"id"`
				Round  *string `json:"round"`
				Read   *string `json:"read"`
				Passes []struct {
					Name  string `json:"name"`
					State string `json:"state"`
				} `json:"passes"`
			} `json:"reports"`
		}
		if err := json.Unmarshal([]byte(files["snapshot.json"]), &snap); err != nil {
			t.Fatal(err)
		}
		if len(snap.Reports) != 1 || snap.Reports[0].Round == nil || *snap.Reports[0].Round != "r1" || snap.Reports[0].Read != nil {
			t.Fatalf("snapshot reports = %+v", snap.Reports)
		}
		for _, p := range snap.Reports[0].Passes {
			if p.State == "" {
				t.Errorf("a folded pass has no state: %+v", p)
			}
		}
		if len(snap.Reports[0].Passes) == 0 {
			t.Error("the report's passes were not folded from the release")
		}

		before := files
		if r := run(t, dir, workedClock, "report", "add", "alpha", in, "--round", "r1"); r.code != 2 {
			t.Errorf("a second report for the round exit %d, want 2: %s", r.code, r.stderr)
		}
		sameFolder(t, before, folder(t, dir))
		if r := run(t, dir, workedClock, "report", "add", "alpha", in, "--round", "r9"); r.code == 0 || !strings.Contains(r.stderr, "r9") {
			t.Errorf("a round that does not exist: exit %d, stderr %q", r.code, r.stderr)
		}
		if r := run(t, dir, workedClock, "report", "add", "nope", in); r.code == 0 || !strings.Contains(r.stderr, "nope") {
			t.Errorf("a project that does not exist: exit %d, stderr %q", r.code, r.stderr)
		}
		sameFolder(t, before, folder(t, dir))
		if r := run(t, dir, workedClock, "report", "add", "alpha", in); r.code != 0 {
			t.Errorf("the release's own report exit %d: %s", r.code, r.stderr)
		}
	})

	t.Run("report_empty_lists", func(t *testing.T) {
		dir := newState(t)
		if r := run(t, dir, workedClock, "report", "add", "alpha", input(t, "report-empty.json"), "--round", "r2"); r.code != 0 {
			t.Fatalf("report add exit %d: %s", r.code, r.stderr)
		}
		files := folder(t, dir)
		if !strings.Contains(files["reports.json"], `"calls": []`) || !strings.Contains(files["reports.json"], `"misses": []`) {
			t.Errorf("reports.json prints an empty list as null:\n%s", files["reports.json"])
		}
		if !strings.Contains(files["snapshot.json"], `"calls": []`) {
			t.Errorf("snapshot.json prints an empty list as null")
		}
	})

	t.Run("report_missing_parts", func(t *testing.T) {
		dir := newState(t)
		before := folder(t, dir)
		data, err := os.ReadFile(input(t, "report.json"))
		if err != nil {
			t.Fatal(err)
		}
		cases := []struct {
			name, want string
			edit       func(m map[string]any)
		}{
			{"no title", `"title"`, func(m map[string]any) { delete(m, "title") }},
			{"no summary", `"summary"`, func(m map[string]any) { delete(m, "summary") }},
			{"no calls", `"calls"`, func(m map[string]any) { delete(m, "calls") }},
			{"null calls", `"calls"`, func(m map[string]any) { m["calls"] = nil }},
			{"no misses", `"misses"`, func(m map[string]any) { delete(m, "misses") }},
			{"null misses", `"misses"`, func(m map[string]any) { m["misses"] = nil }},
			{"call without why", `calls[0]`, func(m map[string]any) {
				m["calls"] = []any{map[string]any{"call": "c", "chosen": "a", "other": "b"}}
			}},
			{"miss without cost", `misses[0]`, func(m map[string]any) {
				m["misses"] = []any{map[string]any{"miss": "m"}}
			}},
			{"unknown key", "extra", func(m map[string]any) { m["extra"] = 1 }},
		}
		for _, c := range cases {
			m := map[string]any{}
			if err := json.Unmarshal(data, &m); err != nil {
				t.Fatal(err)
			}
			c.edit(m)
			out, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			file := filepath.Join(t.TempDir(), "bad.json")
			if err := os.WriteFile(file, out, 0o644); err != nil {
				t.Fatal(err)
			}
			r := run(t, dir, workedClock, "report", "add", "alpha", file, "--round", "r1")
			if r.code == 0 || !strings.Contains(r.stderr, c.want) {
				t.Errorf("%s: exit %d, stderr %q, want a refusal naming %s", c.name, r.code, r.stderr, c.want)
			}
		}
		sameFolder(t, before, folder(t, dir))
	})

	t.Run("report_done_release", func(t *testing.T) {
		dir := newState(t)
		if err := os.MkdirAll(filepath.Join(dir, "releases", "done"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(filepath.Join(dir, "releases", "alpha.json"), filepath.Join(dir, "releases", "done", "alpha.json")); err != nil {
			t.Fatal(err)
		}
		if r := run(t, dir, workedClock, "report", "add", "alpha", input(t, "report.json"), "--round", "r1"); r.code != 0 {
			t.Fatalf("a done release's report exit %d: %s", r.code, r.stderr)
		}
		var snap struct {
			Reports []struct {
				Passes []struct{ Name string } `json:"passes"`
			} `json:"reports"`
		}
		if err := json.Unmarshal([]byte(folder(t, dir)["snapshot.json"]), &snap); err != nil {
			t.Fatal(err)
		}
		if len(snap.Reports) != 1 || len(snap.Reports[0].Passes) == 0 {
			t.Errorf("a done release's report has no passes: %+v", snap.Reports)
		}
	})

	t.Run("report_read", func(t *testing.T) {
		dir := newStateFrom(t, "desk")
		r := run(t, dir, briefClock, "report", "read", "rep-0001")
		if r.code != 0 {
			t.Fatalf("report read exit %d: %s", r.code, r.stderr)
		}
		files := folder(t, dir)
		if !strings.Contains(files["reports.json"], `"read": "2026-11-20T23:00:00Z"`) {
			t.Errorf("reports.json after a read:\n%s", files["reports.json"])
		}
		before := files
		if r := run(t, dir, "2026-11-21T08:00:00Z", "report", "read", "rep-0001"); r.code != 0 {
			t.Fatalf("a second read exit %d: %s", r.code, r.stderr)
		}
		if got := folder(t, dir)["reports.json"]; got != before["reports.json"] {
			t.Errorf("a second read changed reports.json:\n%s", got)
		}
		if r := run(t, dir, briefClock, "report", "read", "rep-0009"); r.code != 2 {
			t.Errorf("an unknown id exit %d, want 2: %s", r.code, r.stderr)
		}
	})

	t.Run("report_read_bad_id", func(t *testing.T) {
		dir := newStateFrom(t, "desk")
		before := folder(t, dir)
		for _, id := range []string{"", "a/b", "../reports"} {
			if r := run(t, dir, briefClock, "report", "read", id); r.code == 0 {
				t.Errorf("report read %q exit 0, want a refusal", id)
			}
		}
		if r := run(t, dir, briefClock, "report", "read", "--", "-x"); r.code == 0 {
			t.Error("report read -x exit 0, want a refusal")
		}
		sameFolder(t, before, folder(t, dir))
	})

	t.Run("report_read_no_usage", func(t *testing.T) {
		dir := newStateFrom(t, "desk")
		marker := filepath.Join(t.TempDir(), "ran")
		script := filepath.Join(t.TempDir(), "usage.sh")
		if err := os.WriteFile(script, []byte("#!/bin/sh\ntouch "+marker+"\nexit 1\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		before := folder(t, dir)
		if r := runWith(t, script, dir, briefClock, "report", "read", "rep-0001"); r.code != 0 {
			t.Fatalf("report read exit %d: %s", r.code, r.stderr)
		}
		if _, err := os.Stat(marker); err == nil {
			t.Error("report read ran the usage command")
		}
		after := folder(t, dir)
		if after["usage.json"] != before["usage.json"] {
			t.Error("report read changed usage.json")
		}
	})

	t.Run("report_broken_file", func(t *testing.T) {
		dir := newStateFrom(t, "desk")
		writeIn(t, dir, "reports.json", `{"format":1,"reports":[{"id":"rep-0001"}]}`)
		before := folder(t, dir)
		r := run(t, dir, briefClock, "report", "read", "rep-0001")
		if r.code == 0 || !strings.Contains(r.stderr, "reports.json") {
			t.Errorf("a broken reports.json: exit %d, stderr %q", r.code, r.stderr)
		}
		after := folder(t, dir)
		if _, ok := after["snapshot-error.json"]; !ok {
			t.Error("no snapshot-error.json was written for the broken file")
		}
		delete(after, "snapshot-error.json")
		sameFolder(t, before, after)
	})
}
