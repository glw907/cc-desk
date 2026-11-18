package sessions

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// CheckState is the bucket `gh pr checks` puts a check in.
type CheckState string

// The buckets, as cli/cli's aggregateChecks names them.
const (
	Pass     CheckState = "pass"
	Skipping CheckState = "skipping"
	Fail     CheckState = "fail"
	Cancel   CheckState = "cancel"
	Pending  CheckState = "pending"
)

// Check is one entry of a pull request's status rollup, with its bucket.
type Check struct {
	Name   string
	Bucket CheckState
}

// Pull is one pull request as the read saw it. State is OPEN, MERGED, or
// CLOSED. Merged and MergeCommit are empty until the pull request merges.
type Pull struct {
	Number      int
	State       string
	Head        string
	Merged      *time.Time
	MergeCommit string
	Checks      []Check
}

// PullRead is the result of the read for one pass. Unreachable means GitHub
// did not answer; a reachable read with no Pulls means no pull request.
type PullRead struct {
	Unreachable bool
	Pulls       []Pull
}

// Choose returns the pull request the ready report numbers, else the first.
func (r PullRead) Choose(number int) (Pull, bool) {
	for _, p := range r.Pulls {
		if p.Number == number {
			return p, true
		}
	}
	if len(r.Pulls) == 0 {
		return Pull{}, false
	}
	return r.Pulls[0], true
}

// ReadPulls runs `<gh> pr list` for the branch's pull requests. Any exec
// error or non-zero exit reads as Unreachable, because `pr list` exits 0
// whenever GitHub answered, with `[]` for no pull request.
func ReadPulls(ctx context.Context, gh, repo, branch string) PullRead {
	ctx, cancel := context.WithTimeout(ctx, ReadTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, gh, "pr", "list", "-R", repo, "--head", branch,
		"--state", "all", "--json", "number,state,headRefOid,mergedAt,mergeCommit,statusCheckRollup").Output()
	if err != nil {
		return PullRead{Unreachable: true}
	}
	pulls, err := ParsePulls(out)
	if err != nil {
		return PullRead{Unreachable: true}
	}
	return PullRead{Pulls: pulls}
}

type rawPull struct {
	Number      int        `json:"number"`
	State       string     `json:"state"`
	Head        string     `json:"headRefOid"`
	MergedAt    *time.Time `json:"mergedAt"`
	MergeCommit *struct {
		Oid string `json:"oid"`
	} `json:"mergeCommit"`
	Rollup []struct {
		Name       string `json:"name"`
		Context    string `json:"context"`
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
		State      string `json:"state"`
	} `json:"statusCheckRollup"`
}

// ParsePulls decodes the output of the read. Output that is not a JSON list
// is an error.
func ParsePulls(data []byte) ([]Pull, error) {
	var raw []rawPull
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("decode pull request list: %w", err)
	}
	var out []Pull
	for _, r := range raw {
		p := Pull{Number: r.Number, State: r.State, Head: r.Head, Merged: r.MergedAt}
		if r.MergeCommit != nil {
			p.MergeCommit = r.MergeCommit.Oid
		}
		for _, c := range r.Rollup {
			name := c.Name
			if name == "" {
				name = c.Context
			}
			p.Checks = append(p.Checks, Check{Name: name, Bucket: bucket(c.State, c.Status, c.Conclusion)})
		}
		out = append(out, p)
	}
	return out, nil
}

// bucket copies cli/cli's aggregateChecks: a StatusContext's state, else a
// CheckRun's conclusion once completed, else its status.
func bucket(state, status, conclusion string) CheckState {
	if state == "" {
		if status == "COMPLETED" {
			state = conclusion
		} else {
			state = status
		}
	}
	switch state {
	case "SUCCESS":
		return Pass
	case "SKIPPED", "NEUTRAL":
		return Skipping
	case "ERROR", "FAILURE", "TIMED_OUT", "ACTION_REQUIRED":
		return Fail
	case "CANCELLED":
		return Cancel
	}
	return Pending
}
