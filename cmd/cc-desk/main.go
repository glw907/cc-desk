// Command cc-desk is the entry point for the cc-desk status page and desk.
package main

import (
	"errors"
	"fmt"
	"os"
	// The reset parser loads a zone by name, with no system database.
	_ "time/tzdata"

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
	// ~/.local/secrets exports ANTHROPIC_API_KEY for scripts that call the API.
	// Every claude that cc-desk starts, the desk and the usage reading included,
	// would bill that key ahead of the claude.ai login, so none of them gets it.
	os.Unsetenv("ANTHROPIC_API_KEY")
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitCode(err))
	}
}
