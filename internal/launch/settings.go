package launch

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/glw907/cc-desk/internal/atomicfile"
)

// Deny holds the commands no launched session may run: the merge is the
// check's, never the session's.
var Deny = []string{"Bash(gh pr merge *)", "Bash(approval-queue land *)"}

// SessionReasons are the stop reasons a pass session reports itself. The
// check records the rest.
var SessionReasons = []string{"usage", "irreversible", "secret", "auth_data", "scope"}

type hookCommand struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout int    `json:"timeout"`
}

type hookGroup struct {
	Hooks []hookCommand `json:"hooks"`
}

type settings struct {
	Permissions struct {
		Allow []string `json:"allow"`
		Deny  []string `json:"deny"`
	} `json:"permissions"`
	Hooks struct {
		PermissionRequest []hookGroup `json:"PermissionRequest"`
	} `json:"hooks"`
}

// Verbs gives the report commands a session may run, as the exact command
// lines its prompt names. binary and stateDir must be absolute.
func Verbs(binary, stateDir string, kind, project, pass, round string) []string {
	base := binary + " --state-dir " + stateDir + " "
	switch kind {
	case "conductor":
		return []string{base + "session finish conductor " + project}
	case "design":
		return []string{base + "session finish design " + project + " " + round}
	}
	verbs := make([]string, len(SessionReasons))
	for i, r := range SessionReasons {
		verbs[i] = base + "pass stop " + project + " " + pass + " --reason " + r
	}
	return verbs
}

// WriteSettings writes <stateDir>/launch/<name>.json fresh and gives its
// path. The allow list is the release's rules plus one exact rule per report
// verb. binary is passed in, never read from os.Executable here, because a
// test's executable is not the command.
func WriteSettings(stateDir, binary, name string, allow, verbs []string) (string, error) {
	if !filepath.IsAbs(stateDir) || !filepath.IsAbs(binary) {
		return "", fmt.Errorf("settings need an absolute state folder and binary, got %q and %q", stateDir, binary)
	}
	var s settings
	s.Permissions.Allow = append([]string{}, allow...)
	for _, v := range verbs {
		s.Permissions.Allow = append(s.Permissions.Allow, "Bash("+v+")")
	}
	s.Permissions.Deny = Deny
	s.Hooks.PermissionRequest = []hookGroup{{Hooks: []hookCommand{{
		Type: "command", Command: binary + " --state-dir " + stateDir + " prompt hook", Timeout: 30,
	}}}}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", err
	}
	dir := filepath.Join(stateDir, "launch")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, name+".json")
	if err := atomicfile.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return "", fmt.Errorf("write settings %s: %w", path, err)
	}
	return path, nil
}
