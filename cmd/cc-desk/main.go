// Command cc-desk is the entry point for the cc-desk status page and desk.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/glw907/cc-desk/internal/state"
)

// exitCode maps an error to the process exit code: 2 for a refusal, 1 for
// any other error.
func exitCode(err error) int {
	if _, ok := errors.AsType[*state.RefusedError](err); ok {
		return 2
	}
	return 1
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitCode(err))
	}
}
