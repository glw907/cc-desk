package desk

import (
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"
)

func TestParseStat(t *testing.T) {
	tests := []struct {
		name string
		line string
		want stat
	}{
		{
			"plain name",
			"4242 (bash) S 4100 4242 4242 34820 4242 4194560 1500 0 0 0 3 1 0 0 20 0 1 0 987654 12345678 900 18446744073709551615 0 0 0 0 0 0 0 0 0 0 0 0 17 2 0 0",
			stat{4100, 34820, 987654},
		},
		{
			"name with a paren and spaces",
			"77 (a) b (c) S 9 77 77 0 -1 4194560 1 0 0 0 0 0 0 0 20 0 1 0 55 1 1 18446744073709551615 0 0 0 0 0 0 0 0 0 0 0 0 17 0 0 0",
			stat{9, 0, 55},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseStat(tt.line)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("parseStat = %+v, want %+v", got, tt.want)
			}
		})
	}
	for _, bad := range []string{"", "1 (x", "1 (x) S 2 3"} {
		if _, err := parseStat(bad); err == nil {
			t.Errorf("parseStat(%q) succeeded, want an error", bad)
		}
	}
}

func TestChain(t *testing.T) {
	chain := Chain()
	if len(chain) < 2 {
		t.Fatalf("chain has %d entries, want at least 2", len(chain))
	}
	if chain[0].PID != os.Getpid() {
		t.Errorf("chain starts at %d, want own PID %d", chain[0].PID, os.Getpid())
	}
	last := chain[len(chain)-1]
	st, err := readStat(last.PID)
	if err != nil {
		t.Fatal(err)
	}
	if last.PID != 1 && st.ppid > 0 {
		t.Errorf("chain ends at %d, whose parent %d is still readable", last.PID, st.ppid)
	}
}

func TestEntry(t *testing.T) {
	before := time.Now()
	cmd := exec.Command("sleep", "30")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	after := time.Now()
	t.Cleanup(func() {
		cmd.Process.Kill()
		cmd.Wait()
	})
	d, err := Entry("session-1", cmd.Process.Pid)
	if err != nil {
		t.Fatal(err)
	}
	if d.SessionID != "session-1" || d.PID != cmd.Process.Pid {
		t.Errorf("entry = %+v", d)
	}
	lo, hi := before.Add(-2*time.Second), after.Add(2*time.Second)
	if d.Started.Before(lo) || d.Started.After(hi) {
		t.Errorf("start %s is outside %s to %s", d.Started, lo, hi)
	}
	p, err := Read(cmd.Process.Pid)
	if err != nil {
		t.Fatal(err)
	}
	if !p.Started.Equal(d.Started) {
		t.Errorf("Read start %s differs from Entry start %s", p.Started, d.Started)
	}
}

func TestEntryDeadProcess(t *testing.T) {
	if _, err := Entry("s", 1<<30); err == nil {
		t.Error("Entry for a missing process succeeded")
	}
}

func TestCallerSessionID(t *testing.T) {
	want := "sess-" + strconv.Itoa(os.Getpid())
	t.Setenv("CLAUDE_CODE_SESSION_ID", want)
	c := Caller()
	if c.SessionID != want || len(c.Chain) == 0 {
		t.Errorf("caller = %+v", c)
	}
}
