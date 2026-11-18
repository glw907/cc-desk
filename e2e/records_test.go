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
