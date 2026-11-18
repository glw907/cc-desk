package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// deskPrelude is what a desk shell runs before the test's commands: it
// records its PID, then blocks until the test writes to the FIFO.
const deskPrelude = `echo $$ > "$DESK_PID_FILE"
read release < "$DESK_FIFO"
`

// startDesk starts a long-lived shell that stands in for a desk session's
// process. With terminal true the shell runs under script(1) and so has a
// controlling terminal; otherwise it runs under setsid and has none. Its
// standard input is /dev/null, as go test gives it, and its output goes to a
// file, so a failed start prints that file.
//
// The shell blocks until release is called, then runs script and exits.
// wait blocks until it has exited. The pseudo-terminal turns line ends into
// CRLF, so script reports through files only.
func startDesk(t *testing.T, terminal bool, env []string, script string) (pid int, release, wait func()) {
	t.Helper()
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "pid")
	fifo := filepath.Join(dir, "fifo")
	outFile := filepath.Join(dir, "desk.out")
	deskSh := filepath.Join(dir, "desk.sh")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(deskSh, []byte(deskPrelude+script+"\nwait\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := os.Create(outFile)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { out.Close() })

	var cmd *exec.Cmd
	if terminal {
		cmd = exec.Command("script", "-qec", "sh "+deskSh, "/dev/null")
	} else {
		cmd = exec.Command("setsid", "sh", deskSh)
	}
	cmd.Env = append(os.Environ(), env...)
	cmd.Env = append(cmd.Env, "DESK_PID_FILE="+pidFile, "DESK_FIFO="+fifo)
	cmd.Stdin = nil
	cmd.Stdout, cmd.Stderr = out, out
	cmd.WaitDelay = 5 * time.Second
	if err := cmd.Start(); err != nil {
		t.Fatalf("start desk: %v", err)
	}
	done := make(chan struct{})
	go func() {
		cmd.Wait()
		close(done)
	}()

	var once sync.Once
	release = func() {
		once.Do(func() {
			// O_RDWR never blocks on a FIFO, even if the shell is gone.
			f, err := os.OpenFile(fifo, os.O_RDWR, 0)
			if err != nil {
				t.Errorf("release desk: %v", err)
				return
			}
			f.WriteString("go\n")
			f.Close()
		})
	}
	wait = func() {
		t.Helper()
		select {
		case <-done:
		case <-time.After(60 * time.Second):
			t.Fatalf("desk did not exit; its output:\n%s", readFile(t, outFile))
		}
	}
	t.Cleanup(func() {
		release()
		cmd.Process.Kill()
		<-done
	})

	deadline := time.Now().Add(5 * time.Second)
	for {
		b, err := os.ReadFile(pidFile)
		if err == nil && strings.HasSuffix(string(b), "\n") {
			pid, err = strconv.Atoi(strings.TrimSpace(string(b)))
			if err != nil {
				t.Fatalf("desk pid file %q: %v", b, err)
			}
			return pid, release, wait
		}
		if time.Now().After(deadline) {
			t.Fatalf("desk wrote no PID in 5 seconds; its output:\n%s", readFile(t, outFile))
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("(%v)", err)
	}
	return string(b)
}

// procStat is the part of /proc/<pid>/stat the guard reads.
type procStat struct {
	ppid, tty int
}

// parseProcStat reads the fields after the last ")" of a stat line, so a
// command name holding ")" or spaces doesn't shift them.
func parseProcStat(t *testing.T, line string) procStat {
	t.Helper()
	i := strings.LastIndex(line, ")")
	if i < 0 {
		t.Fatalf("no ) in stat line %q", line)
	}
	f := strings.Fields(line[i+1:])
	if len(f) < 5 {
		t.Fatalf("short stat line %q", line)
	}
	ppid, err1 := strconv.Atoi(f[1])
	tty, err2 := strconv.Atoi(f[4])
	if err1 != nil || err2 != nil {
		t.Fatalf("bad stat line %q", line)
	}
	return procStat{ppid, tty}
}

func readProcStat(t *testing.T, pid int) procStat {
	t.Helper()
	return parseProcStat(t, readFile(t, fmt.Sprintf("/proc/%d/stat", pid)))
}

// chainScript walks from its own process up through its parents and prints
// "pid ppid tty_nr" for each, using only sh and awk.
const chainScript = `p=$$
while [ "$p" -gt 1 ]; do
  line=$(cat /proc/$p/stat) || break
  echo "$p $(echo "$line" | awk '{ sub(/.*\) /, ""); print $2, $5 }')"
  p=$(echo "$line" | awk '{ sub(/.*\) /, ""); print $2 }')
done
`

func TestDeskHarness(t *testing.T) {
	out := t.TempDir()
	if err := os.WriteFile(filepath.Join(out, "chain.sh"), []byte(chainScript), 0o644); err != nil {
		t.Fatal(err)
	}
	script := `( cat /proc/self/stat > "$OUT/job.stat" ) &
( setsid sh "$OUT/chain.sh" > "$OUT/sjob.chain" ) &
`
	env := []string{"OUT=" + out}

	termPID, release, wait := startDesk(t, true, env, script)
	if tty := readProcStat(t, termPID).tty; tty == 0 {
		t.Error("terminal desk shell has tty_nr 0")
	}
	setsidPID, releaseS, _ := startDesk(t, false, nil, "")
	if tty := readProcStat(t, setsidPID).tty; tty != 0 {
		t.Errorf("setsid desk shell has tty_nr %d, want 0", tty)
	}
	releaseS()

	release()
	wait()

	job := parseProcStat(t, readFile(t, filepath.Join(out, "job.stat")))
	if job.ppid != termPID {
		t.Errorf("background job's parent is %d, want the desk shell %d", job.ppid, termPID)
	}

	lines := strings.Split(strings.TrimSpace(readFile(t, filepath.Join(out, "sjob.chain"))), "\n")
	if len(lines) < 2 {
		t.Fatalf("setsid job's chain too short:\n%s", strings.Join(lines, "\n"))
	}
	first := strings.Fields(lines[0])
	if len(first) != 3 || first[2] != "0" {
		t.Errorf("setsid job reads tty_nr %q, want 0", lines[0])
	}
	found := false
	for _, l := range lines {
		if f := strings.Fields(l); len(f) == 3 && f[0] == strconv.Itoa(termPID) {
			found = true
		}
	}
	if !found {
		t.Errorf("desk shell %d is not an ancestor of the setsid job:\n%s", termPID, strings.Join(lines, "\n"))
	}
}
