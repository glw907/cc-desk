package desk

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestWriteSettings(t *testing.T) {
	state := t.TempDir()
	words := filepath.Join(state, "desk", "words")
	prefix := "/bin/cc-desk --state-dir " + state
	path, err := WriteSettings(state, "abc", prefix, words, "")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(state, "launch", "desk-abc.json"); path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]struct{ Allow []string }
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("keys = %v, want permissions only", got)
	}
	allow := got["permissions"].Allow
	for _, want := range []string{
		"Bash(" + prefix + " ruling record *)",
		"Bash(" + prefix + " brief*)",
		"Bash(" + prefix + " prompt *)",
		"Bash(" + prefix + " pass start *)",
		"Read(/" + state + "/**)",
		"Edit(/" + words + "/**)",
	} {
		if !slices.Contains(allow, want) {
			t.Errorf("allow lacks %q:\n%s", want, strings.Join(allow, "\n"))
		}
	}
	if slices.ContainsFunc(allow, func(r string) bool { return strings.HasPrefix(r, "Write(") }) {
		t.Error("allow holds a Write rule, which matches nothing")
	}
}

func TestRulesSkillFile(t *testing.T) {
	rules := Rules("p", "/s", "/s/desk/words", "/r/SKILL.md")
	if !slices.Contains(rules, "Read(//r/SKILL.md)") {
		t.Errorf("rules = %q, want a Read rule for the skill file", rules)
	}
}

func TestSettingsNeedAbsolutePaths(t *testing.T) {
	if _, err := WriteSettings("rel", "abc", "p", "/w", ""); err == nil {
		t.Error("a relative state folder was accepted")
	}
}

func TestNewID(t *testing.T) {
	a, err := NewID()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := NewID()
	if len(a) != 36 || a[14] != '4' || a == b {
		t.Errorf("ids %q and %q: want distinct v4 UUIDs", a, b)
	}
}
