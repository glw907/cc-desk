package sessions

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// ReadTimeout bounds `claude logs`, and ProbeTimeout bounds the network probe.
const (
	ReadTimeout  = 10 * time.Second
	ProbeTimeout = 10 * time.Second
)

// SignInLimit is how long the sign-in read may run.
const SignInLimit = 30 * time.Second

// SignedIn runs `<claude> auth status --json` and gives when the stored
// sign-in was last written: the modification time of .credentials.json in
// the configDirectory the command names. "loggedIn": true alone is no
// evidence of a new sign-in, since it reads the stored credentials and can
// stay true after they expire. The zero time means the read isn't good: a
// failed exit, output it can't decode, "loggedIn" false, a loginExpiresAt
// already past, no credentials file, or a run past the limit. At the limit
// the command's process group is killed, so a child holding the output pipe
// dies with it.
func SignedIn(ctx context.Context, claude string, limit time.Duration) time.Time {
	if claude == "" {
		return time.Time{}
	}
	ctx, cancel := context.WithTimeout(ctx, limit)
	defer cancel()
	cmd := exec.CommandContext(ctx, claude, "auth", "status", "--json")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = time.Second
	var out bytes.Buffer
	cmd.Stdout = &out
	if cmd.Run() != nil {
		return time.Time{}
	}
	var status struct {
		LoggedIn        bool      `json:"loggedIn"`
		ConfigDirectory string    `json:"configDirectory"`
		LoginExpiresAt  time.Time `json:"loginExpiresAt"`
	}
	if json.Unmarshal(out.Bytes(), &status) != nil || !status.LoggedIn || status.ConfigDirectory == "" {
		return time.Time{}
	}
	if !status.LoginExpiresAt.IsZero() && !status.LoginExpiresAt.After(time.Now()) {
		return time.Time{}
	}
	fi, err := os.Stat(filepath.Join(status.ConfigDirectory, ".credentials.json"))
	if err != nil {
		return time.Time{}
	}
	return fi.ModTime().UTC()
}

// List runs `<claude> agents --json --all` and decodes its rows.
func List(ctx context.Context, claude string) ([]Row, error) {
	if claude == "" {
		return nil, errNoCommand
	}
	ctx, cancel := context.WithTimeout(ctx, ReadTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, claude, "agents", "--json", "--all").Output()
	if err != nil {
		return nil, err
	}
	return ParseList(out)
}

// Screen runs `<claude> logs <short id>` and gives its text. A failure or a
// timeout reads as no text, never an error: the screen is evidence, and its
// absence is none.
func Screen(ctx context.Context, claude, shortID string) string {
	if claude == "" || shortID == "" {
		return ""
	}
	ctx, cancel := context.WithTimeout(ctx, ReadTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, claude, "logs", shortID).Output()
	if err != nil {
		return ""
	}
	return string(out)
}

// ReadJob reads detail and needs from <jobsDir>/<short id>/state.json. Claude
// Code's docs tell scripts not to read that file, so a missing file, a parse
// failure, or a field of another type reads as empty.
func ReadJob(jobsDir, shortID string) Job {
	if jobsDir == "" || shortID == "" {
		return Job{}
	}
	data, err := os.ReadFile(filepath.Join(jobsDir, shortID, "state.json"))
	if err != nil {
		return Job{}
	}
	var raw map[string]any
	if json.Unmarshal(data, &raw) != nil {
		return Job{}
	}
	detail, _ := raw["detail"].(string)
	needs, _ := raw["needs"].(string)
	return Job{Detail: detail, Needs: needs}
}

// Probe sends one request to url. Any HTTP answer, whatever its status, says
// the link is up. An error or a timeout says it is down.
func Probe(ctx context.Context, url string) bool {
	ctx, cancel := context.WithTimeout(ctx, ProbeTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}
