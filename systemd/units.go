// Package systemd holds the user units that run the cc-desk check. The
// files in this folder are the ones `cc-desk timer install` writes.
package systemd

import _ "embed"

// Service is cc-desk-check.service.
//
//go:embed cc-desk-check.service
var Service []byte

// Timer is cc-desk-check.timer.
//
//go:embed cc-desk-check.timer
var Timer []byte
