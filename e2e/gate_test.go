package e2e

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const noteSkill = "---\nname: note\ndescription: A skill the gate test reads.\n---\n\n" +
	"# Note\n\nThis skill holds a test.\n"

// scratchRepo builds a one-package module in a temporary folder, with a copy
// of this repo's gate, its vale config, one doc, and one skill. A files entry
// with an empty value removes that file. The scratch module holds
// no e2e package, so the gate can't recurse into this test.
func scratchRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	put := func(rel string, data []byte) {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gate, err := os.ReadFile("../scripts/gate.sh")
	if err != nil {
		t.Fatal(err)
	}
	put("scripts/gate.sh", gate)
	ini, err := os.ReadFile("../.vale.ini")
	if err != nil {
		t.Fatal(err)
	}
	put(".vale.ini", ini)
	err = filepath.WalkDir("../styles", func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel("..", p)
		data, err := os.ReadFile(p)
		put(rel, data)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	put("go.mod", []byte("module scratch\n\ngo 1.27\n"))
	put("docs/note.md", []byte("This folder holds a test.\n"))
	put("skills/note/SKILL.md", []byte(noteSkill))
	for rel, data := range files {
		if data == "" {
			if err := os.Remove(filepath.Join(dir, rel)); err != nil {
				t.Fatal(err)
			}
			continue
		}
		put(rel, []byte(data))
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "."}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return dir
}

func runGate(t *testing.T, dir string) (string, int) {
	t.Helper()
	cmd := exec.Command("bash", "scripts/gate.sh")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), 0
	}
	ee, ok := errors.AsType[*exec.ExitError](err)
	if !ok {
		t.Fatal(err)
	}
	return string(out), ee.ExitCode()
}

func TestGateFailsOnSkip(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		wantCode int
		wantOut  string
	}{
		{
			name: "a skipped test",
			files: map[string]string{"a/a_test.go": "package a\n\nimport \"testing\"\n\n" +
				"func TestSkipped(t *testing.T) {\n\tt.Skip(\"probe\")\n}\n"},
			wantCode: 1,
			wantOut:  "scratch/a TestSkipped",
		},
		{
			name: "a skipped subtest",
			files: map[string]string{"a/a_test.go": "package a\n\nimport \"testing\"\n\n" +
				"func TestOuter(t *testing.T) {\n\tt.Run(\"inner\", func(t *testing.T) { t.Skip(\"probe\") })\n}\n"},
			wantCode: 1,
			wantOut:  "TestOuter/inner",
		},
		{
			name: "no skip",
			files: map[string]string{"a/a_test.go": "package a\n\nimport \"testing\"\n\n" +
				"func TestRuns(t *testing.T) {}\n"},
			wantCode: 0,
			wantOut:  "gate: ok",
		},
		{
			name: "a package with no test files",
			files: map[string]string{
				"a/a_test.go": "package a\n\nimport \"testing\"\n\nfunc TestRuns(t *testing.T) {}\n",
				"b/b.go":      "package b\n",
			},
			wantCode: 0,
			wantOut:  "gate: ok",
		},
		{
			name: "a compile error",
			files: map[string]string{"a/a_test.go": "package a\n\nimport \"testing\"\n\n" +
				"func TestBroken(t *testing.T) { undefinedName() }\n"},
			wantCode: 1,
			wantOut:  "undefined: undefinedName",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := scratchRepo(t, tt.files)
			out, code := runGate(t, dir)
			if code != tt.wantCode {
				t.Errorf("gate exit %d, want %d\n%s", code, tt.wantCode, out)
			}
			if !strings.Contains(out, tt.wantOut) {
				t.Errorf("gate output lacks %q:\n%s", tt.wantOut, out)
			}
		})
	}
}

func TestGateSkills(t *testing.T) {
	const pkg = "package a\n\nimport \"testing\"\n\nfunc TestRuns(t *testing.T) {}\n"
	hinge := noteSkill + "\nThe check reads the record first, since the prompt names it. " +
		"The session stops on a refusal, which leaves the branch as it was. " +
		"The reader gets the file next, so the question stays plain.\n"
	tests := []struct {
		name     string
		skill    string
		wantCode int
		wantOut  string
	}{
		{name: "a clean skill", skill: noteSkill, wantCode: 0, wantOut: "gate: ok"},
		{name: "no skill", skill: "", wantCode: 1, wantOut: "no skills/*/SKILL.md is tracked"},
		{
			name:     "160 lines",
			skill:    noteSkill + strings.Repeat("\n", 160-strings.Count(noteSkill, "\n")),
			wantCode: 1,
			wantOut:  "has 160 lines",
		},
		{name: "a hinge run", skill: hinge, wantCode: 1, wantOut: "tellgrader failed on skills/note/SKILL.md (exit 2)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := scratchRepo(t, map[string]string{"a/a_test.go": pkg, "skills/note/SKILL.md": tt.skill})
			out, code := runGate(t, dir)
			if code != tt.wantCode {
				t.Errorf("gate exit %d, want %d\n%s", code, tt.wantCode, out)
			}
			if !strings.Contains(out, tt.wantOut) {
				t.Errorf("gate output lacks %q:\n%s", tt.wantOut, out)
			}
		})
	}
}
