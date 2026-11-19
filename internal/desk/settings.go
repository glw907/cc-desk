package desk

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/glw907/cc-desk/internal/atomicfile"
)

// CheckCommandPath refuses a path that cannot sit in an allow rule or in the
// command the session types: one with whitespace, a glob character, a quote,
// or another character the shell or the rule matcher reads.
func CheckCommandPath(path string) error {
	if strings.ContainsAny(path, " \t\n*?[]{}()<>|&;$`'\"\\!#~") {
		return fmt.Errorf("the path %q has a space or a character the desk's permission rules read; run cc-desk from a path without one", path)
	}
	return nil
}

// Rules gives the allow rules of a desk session. prefix is the command the
// session types before each verb, so the rules match its text exactly. A
// Write rule matches nothing in Claude Code, so the words folder takes an
// Edit rule, which covers the Write tool too. skillFile may be empty.
func Rules(prefix, stateDir, wordsDir, skillFile string) []string {
	rules := []string{
		"Bash(" + prefix + " ruling record *)",
		"Bash(" + prefix + " brief*)",
		"Bash(" + prefix + " prompt *)",
		"Bash(" + prefix + " pass start *)",
		"Read(/" + stateDir + "/**)",
	}
	if skillFile != "" {
		rules = append(rules, "Read(/"+skillFile+")")
	}
	return append(rules, "Edit(/"+wordsDir+"/**)")
}

// WriteSettings writes <stateDir>/launch/desk-<id>.json and gives its path.
// The file holds permissions.allow only: no hook, and no deny change.
// stateDir, wordsDir, and skillFile are absolute.
func WriteSettings(stateDir, id, prefix, wordsDir, skillFile string) (string, error) {
	if !filepath.IsAbs(stateDir) || !filepath.IsAbs(wordsDir) {
		return "", fmt.Errorf("desk settings need absolute folders, got %q and %q", stateDir, wordsDir)
	}
	var s struct {
		Permissions struct {
			Allow []string `json:"allow"`
		} `json:"permissions"`
	}
	s.Permissions.Allow = Rules(prefix, stateDir, wordsDir, skillFile)
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", err
	}
	dir := filepath.Join(stateDir, "launch")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "desk-"+id+".json")
	if err := atomicfile.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return "", fmt.Errorf("write settings %s: %w", path, err)
	}
	return path, nil
}
