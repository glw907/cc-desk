package launch

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"syscall"
	"time"
)

// DefaultLimit is how long a merge call, a fetch, or a pull may run.
const DefaultLimit = 60 * time.Second

// killDelay is how long Run waits for the pipes after the kill before it
// gives up on them.
const killDelay = 5 * time.Second

// Run runs one command under its own limit and gives its stdout and stderr.
// A limit of zero or less means no limit. At the limit the command's whole
// process group is killed, so a grandchild holding the output pipe dies with
// it; WaitDelay is only the backstop for a pipe still held after that. The
// error then reads "stopped at the <limit> limit".
func Run(ctx context.Context, limit time.Duration, name string, args ...string) (stdout, stderr string, err error) {
	tctx := ctx
	if limit > 0 {
		var cancel context.CancelFunc
		tctx, cancel = context.WithTimeout(ctx, limit)
		defer cancel()
	}
	cmd := exec.CommandContext(tctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = killDelay
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err = cmd.Run()
	if err != nil && errors.Is(tctx.Err(), context.DeadlineExceeded) && ctx.Err() == nil {
		err = LimitError(limit)
	}
	return out.String(), errOut.String(), err
}

// LimitError is the error of a call stopped at its limit.
func LimitError(limit time.Duration) error {
	if limit%time.Second == 0 {
		return fmt.Errorf("stopped at the %ds limit", limit/time.Second)
	}
	return fmt.Errorf("stopped at the %s limit", limit)
}
