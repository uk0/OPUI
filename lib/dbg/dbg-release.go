// Package dbg provided methods to write debug message to standard logger,
// which can be disabled at runtime and build time.

// +build release

package dbg

import "log"

// Enabled reports whether debug output is enabled.
const Enabled = false

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

// SetPrefix set the prefix for debug messages.
func SetPrefix(s string) {
}

// Log print debug message to the standard logger if enabled.
// Arguments are handled in the manner of fmt.Print.
func Log(v ...interface{}) {
	// do nothing
}

// Logf print debug message to the standard logger if enabled.
// Arguments are handled in the manner of fmt.Printf.
func Logf(format string, v ...interface{}) {
	// do nothing
}

// Logln print debug message to the standard logger if enabled.
// Arguments are handled in the manner of fmt.Println.
func Logln(v ...interface{}) {
	// do nothing
}
