// Package sessions reads what Claude Code reports about background sessions
// and decides what each session the state files name needs.
package sessions

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Row is one entry of `claude agents --json --all`. Claude Code owns the
// shape and adds keys without notice, so rows decode tolerantly: unknown
// keys are ignored.
type Row struct {
	PID        int    `json:"pid"`
	ID         string `json:"id"`
	SessionID  string `json:"sessionId"`
	Name       string `json:"name"`
	Cwd        string `json:"cwd"`
	Kind       string `json:"kind"`
	StartedAt  int64  `json:"startedAt"`
	Status     string `json:"status"`
	State      string `json:"state"`
	WaitingFor string `json:"waitingFor"`
}

// Class is the one class the check sorts a row into.
type Class string

const (
	ClassWorking Class = "working"
	ClassBlocked Class = "blocked"
	ClassDone    Class = "done"
	ClassNone    Class = "none"
)

// Class sorts the row. A failed or stopped row, and any state the check does
// not know, is no row.
func (r Row) Class() Class {
	switch r.State {
	case "working":
		return ClassWorking
	case "blocked":
		return ClassBlocked
	case "done":
		return ClassDone
	}
	return ClassNone
}

// Live reports whether the row is working or blocked.
func (r Row) Live() bool {
	c := r.Class()
	return c == ClassWorking || c == ClassBlocked
}

// ParseList decodes the output of `claude agents --json --all`. Output that
// is not a JSON list of objects is an error. An empty list is not.
func ParseList(data []byte) ([]Row, error) {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("session list is not a JSON list: %w", err)
	}
	if raw == nil {
		return nil, errors.New("session list is null, not a JSON list")
	}
	rows := make([]Row, 0, len(raw))
	for i, m := range raw {
		var r Row
		if err := json.Unmarshal(m, &r); err != nil {
			return nil, fmt.Errorf("session list row %d: %w", i, err)
		}
		rows = append(rows, r)
	}
	return rows, nil
}

// Find gives the row for a short id. With two rows for one id, which
// `--all` can list, a live row wins, then the first.
func Find(rows []Row, shortID string) (Row, bool) {
	if shortID == "" {
		return Row{}, false
	}
	var found Row
	ok := false
	for _, r := range rows {
		if r.ID != shortID {
			continue
		}
		if r.Live() {
			return r, true
		}
		if !ok {
			found, ok = r, true
		}
	}
	return found, ok
}

var errNoCommand = errors.New("no claude command")
