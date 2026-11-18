// Package desk reads processes from /proc for the ruling guard: a process's
// start time and terminal, the chain of parents above the calling process,
// and the desk entry a registered session needs.
package desk

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/glw907/cc-desk/internal/state"
)

// tick is the length of one clock tick in /proc/<pid>/stat. Linux fixes it
// at 100 a second for that file (USER_HZ), whatever the kernel's own rate.
const tick = time.Second / 100

// maxChain bounds the walk up the parents; a real chain is a few dozen deep.
const maxChain = 256

// stat is the part of /proc/<pid>/stat the guard reads.
type stat struct {
	ppid, tty int
	startTick int64
}

// parseStat reads the fields after the last ")" of a stat line, so a command
// name that holds ")" or spaces can't shift them. After the name, state is
// field 1, parent PID 2, tty_nr 5, and starttime 20.
func parseStat(line string) (stat, error) {
	i := strings.LastIndex(line, ")")
	if i < 0 {
		return stat{}, errors.New("stat line has no command name")
	}
	f := strings.Fields(line[i+1:])
	if len(f) < 20 {
		return stat{}, fmt.Errorf("stat line has %d fields after the name, want at least 20", len(f))
	}
	ppid, err := strconv.Atoi(f[1])
	if err != nil {
		return stat{}, fmt.Errorf("parent PID %q: %w", f[1], err)
	}
	tty, err := strconv.Atoi(f[4])
	if err != nil {
		return stat{}, fmt.Errorf("tty_nr %q: %w", f[4], err)
	}
	ticks, err := strconv.ParseInt(f[19], 10, 64)
	if err != nil {
		return stat{}, fmt.Errorf("starttime %q: %w", f[19], err)
	}
	return stat{ppid, tty, ticks}, nil
}

// bootTime reads btime from /proc/stat. It is whole seconds, so two reads of
// it can differ by one, which is why state.SameStart allows two seconds.
func bootTime() (time.Time, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return time.Time{}, fmt.Errorf("read boot time: %w", err)
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		if v, ok := strings.CutPrefix(line, "btime "); ok {
			secs, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
			if err != nil {
				return time.Time{}, fmt.Errorf("boot time %q: %w", v, err)
			}
			return time.Unix(secs, 0).UTC(), nil
		}
	}
	return time.Time{}, errors.New("boot time: no btime line in /proc/stat")
}

// startTime is the one conversion from clock ticks since boot to a time.
func startTime(boot time.Time, ticks int64) time.Time {
	return boot.Add(time.Duration(ticks) * tick).UTC()
}

func readStat(pid int) (stat, error) {
	data, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return stat{}, err
	}
	st, err := parseStat(string(data))
	if err != nil {
		return stat{}, fmt.Errorf("process %d: %w", pid, err)
	}
	return st, nil
}

// Read reads one process from /proc.
func Read(pid int) (state.Proc, error) {
	st, err := readStat(pid)
	if err != nil {
		return state.Proc{}, err
	}
	boot, err := bootTime()
	if err != nil {
		return state.Proc{}, err
	}
	return state.Proc{PID: pid, Started: startTime(boot, st.startTick), TTY: st.tty}, nil
}

// Chain walks from the calling process up through its parents. It stops
// after PID 1 or at a parent PID of 0, which is how the top of a PID
// namespace shows. A process that exits or can't be read mid-walk ends the
// chain there, so Chain never fails.
func Chain() []state.Proc {
	boot, err := bootTime()
	if err != nil {
		return nil
	}
	var chain []state.Proc
	pid := os.Getpid()
	for range maxChain {
		st, err := readStat(pid)
		if err != nil {
			break
		}
		chain = append(chain, state.Proc{PID: pid, Started: startTime(boot, st.startTick), TTY: st.tty})
		if pid == 1 || st.ppid <= 0 {
			break
		}
		pid = st.ppid
	}
	return chain
}

// Caller is the calling process's session id and chain of parents.
func Caller() state.Caller {
	return state.Caller{SessionID: os.Getenv("CLAUDE_CODE_SESSION_ID"), Chain: Chain()}
}

// Entry builds a desk entry for the session whose process is pid.
func Entry(sessionID string, pid int) (state.Desk, error) {
	p, err := Read(pid)
	if err != nil {
		return state.Desk{}, fmt.Errorf("read desk process: %w", err)
	}
	return state.Desk{SessionID: sessionID, PID: pid, Started: p.Started}, nil
}
