package main

import (
	"bytes"
	"testing"
)

func TestRunPrintsVersion(t *testing.T) {
	var buf bytes.Buffer
	if err := run(&buf); err != nil {
		t.Fatal(err)
	}
	if got, want := buf.String(), "cc-desk "+version+"\n"; got != want {
		t.Errorf("run() wrote %q, want %q", got, want)
	}
}
