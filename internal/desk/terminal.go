package desk

import (
	"os"
	"syscall"
	"unsafe"
)

func isTerminal(f *os.File) bool {
	var t syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), syscall.TCGETS, uintptr(unsafe.Pointer(&t)))
	return errno == 0
}

// HasTerminal reports whether standard input and output are both terminals
// and this process has a controlling terminal. The ioctl is the test, since
// /dev/null is a character device too. A desk under setsid has terminals on
// both streams and no controlling one, and it couldn't pass the ruling
// guard, so it doesn't open.
func HasTerminal() bool {
	if !isTerminal(os.Stdin) || !isTerminal(os.Stdout) {
		return false
	}
	p, err := Read(os.Getpid())
	return err == nil && p.TTY != 0
}
