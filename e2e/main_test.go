// Package e2e runs the built cc-desk binary against copies of the fixture
// state folder in testdata/state.
package e2e

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// workedClock is 2:00 pm Alaska time on 2026-11-20, the moment of the
// worked case in testdata/README.md.
const workedClock = "2026-11-20T23:00:00Z"

var binary string

func TestMain(m *testing.M) {
	os.Exit(runMain(m))
}

func runMain(m *testing.M) int {
	dir, err := os.MkdirTemp("", "cc-desk-e2e-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer os.RemoveAll(dir)
	binary = filepath.Join(dir, "cc-desk")
	build := exec.Command("go", "build", "-o", binary, "./cmd/cc-desk")
	build.Dir = ".."
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build cc-desk: %v\n%s", err, out)
		return 1
	}
	return m.Run()
}

type result struct {
	stdout, stderr string
	code           int
}

// run executes cc-desk on the state folder at the given clock.
func run(t *testing.T, dir, now string, args ...string) result {
	t.Helper()
	full := append([]string{"--state-dir", dir, "--now", now}, args...)
	cmd := exec.Command(binary, full...)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	code := 0
	if err != nil {
		ee, ok := errors.AsType[*exec.ExitError](err)
		if !ok {
			t.Fatalf("run %v: %v", args, err)
		}
		code = ee.ExitCode()
	}
	return result{out.String(), errOut.String(), code}
}

// newState copies the base fixture into a fresh folder.
func newState(t *testing.T) string {
	t.Helper()
	dst := filepath.Join(t.TempDir(), "state")
	err := filepath.WalkDir("testdata/state", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel("testdata/state", p)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return dst
}

// folder reads every file under dir, leaving out the lock file.
func folder(t *testing.T, dir string) map[string]string {
	t.Helper()
	files := map[string]string{}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		if rel == "lock" {
			return nil
		}
		data, err := os.ReadFile(p)
		files[rel] = string(data)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

// requireSnapshot fails unless snapshot.json exists and holds bytes, so an
// absent file can't pass as unchanged.
func requireSnapshot(t *testing.T, files map[string]string) {
	t.Helper()
	if files["snapshot.json"] == "" {
		t.Fatal("snapshot.json is absent or empty")
	}
}

func sameFolder(t *testing.T, before, after map[string]string) {
	t.Helper()
	for name, want := range before {
		got, ok := after[name]
		if !ok {
			t.Errorf("%s was removed", name)
		} else if got != want {
			t.Errorf("%s changed", name)
		}
	}
	for name := range after {
		if _, ok := before[name]; !ok {
			t.Errorf("%s was written", name)
		}
	}
}

func input(t *testing.T, name string) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join("testdata", "inputs", name))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func writeIn(t *testing.T, dir, rel, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, rel), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
