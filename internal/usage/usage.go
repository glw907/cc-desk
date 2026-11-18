// Package usage takes the weekly usage reading from claude -p "/usage" and
// decides when the reading kept in usage.json is still fresh.
package usage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/glw907/cc-desk/internal/state"
)

const (
	// MaxAge is how long a reading stays fresh.
	MaxAge = 5 * time.Minute

	// Timeout bounds one run of the command.
	Timeout = 60 * time.Second

	// waitDelay cuts off a killed command's output pipe. A shell stand-in's
	// child holds stdout open after the kill, so without it the timeout
	// would not return until the child exits.
	waitDelay = time.Second
)

var weekLine = regexp.MustCompile(`Current week \(all models\): (\d+)% used · resets (.+)`)

// Parse reads the weekly figure from the raw JSON result of /usage: the
// percent and reset text on the first "Current week (all models)" line.
// Other lines that carry a percent are ignored.
func Parse(raw []byte) (percent int, resetText string, err error) {
	var doc struct {
		IsError bool   `json:"is_error"`
		Result  string `json:"result"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return 0, "", fmt.Errorf("the result is not JSON: %w", err)
	}
	if doc.IsError {
		return 0, "", errors.New("the result reports an error")
	}
	m := weekLine.FindStringSubmatch(doc.Result)
	if m == nil {
		return 0, "", errors.New(`the result has no "Current week (all models)" line`)
	}
	percent, err = strconv.Atoi(m[1])
	if err != nil {
		return 0, "", fmt.Errorf("the percent %q: %w", m[1], err)
	}
	return percent, strings.TrimSpace(m[2]), nil
}

// Read runs command with no arguments and returns what it prints. An empty
// command runs claude -p "/usage" on the haiku model with JSON output. The
// run stops after Timeout, or sooner when ctx ends. Standard input is the
// null device.
func Read(ctx context.Context, command string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()
	var cmd *exec.Cmd
	if command == "" {
		cmd = exec.CommandContext(ctx, "claude", "-p", "/usage", "--model", "haiku", "--output-format", "json")
	} else {
		cmd = exec.CommandContext(ctx, command)
	}
	cmd.WaitDelay = waitDelay
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() != nil {
			return out, fmt.Errorf("the command timed out: %w", ctx.Err())
		}
		return out, fmt.Errorf("the command failed: %w", err)
	}
	return out, nil
}

// Cached reads usage.json in dir leniently and reports whether the reading
// is fresh at now. A file that is absent, does not parse, has no read time,
// is dated after now, or is MaxAge or more old is not fresh.
func Cached(dir string, now time.Time) (u state.Usage, fresh bool) {
	b, err := os.ReadFile(filepath.Join(dir, state.UsageFile))
	if err != nil || json.Unmarshal(b, &u) != nil || u.ReadAt.IsZero() {
		return state.Usage{}, false
	}
	return u, !u.ReadAt.After(now) && now.Sub(u.ReadAt) < MaxAge
}

// Take returns the reading to fold into the state at now. It is the cached
// reading when that is fresh and force is false. Otherwise it runs command
// and returns the new reading with took true, and the raw bytes the command
// printed, if any. The new reading holds an error when the run fails or its
// result can't be read.
func Take(ctx context.Context, dir, command string, now time.Time, force bool) (u state.Usage, raw []byte, took bool) {
	if cached, ok := Cached(dir, now); ok && !force {
		return cached, nil, false
	}
	raw, err := Read(ctx, command)
	u = state.Usage{Format: state.Format, ReadAt: now.UTC()}
	if err == nil {
		var percent int
		var reset string
		if percent, reset, err = Parse(raw); err == nil {
			u.Percent, u.ResetText = percent, reset
		}
	}
	if err != nil {
		u.Error = err.Error()
	}
	return u, raw, true
}
