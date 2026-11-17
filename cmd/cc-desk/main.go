// Command cc-desk is the entry point for the cc-desk status page and desk.
package main

import (
	"fmt"
	"io"
	"os"
)

// version is the release this build reports.
const version = "0.0.0-prove"

// run writes the version line to w.
func run(w io.Writer) error {
	_, err := fmt.Fprintf(w, "cc-desk %s\n", version)
	return err
}

func main() {
	if err := run(os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "cc-desk:", err)
		os.Exit(1)
	}
}
