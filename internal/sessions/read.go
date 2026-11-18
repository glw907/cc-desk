package sessions

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// ReadTimeout bounds `claude logs`, and ProbeTimeout bounds the network probe.
const (
	ReadTimeout  = 10 * time.Second
	ProbeTimeout = 10 * time.Second
)

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
