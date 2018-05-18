// Package dbg provided methods to write debug message to standard logger,
// which can be disabled at runtime and build time.

// +build !release

package dbg

import (
	"fmt"
	"log"
)

// Enabled reports whether debug output is enabled.
const Enabled = true

var prefix = `(D) `

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

// SetPrefix set the prefix for debug messages.
func SetPrefix(s string) {
	prefix = s
}

// Log print debug message to the standard logger if enabled.
// Arguments are handled in the manner of fmt.Print.
func Log(v ...interface{}) {
	log.Output(2, prefix+fmt.Sprint(v...))
}

// Logf print debug message to the standard logger if enabled.
// Arguments are handled in the manner of fmt.Printf.
func Logf(format string, v ...interface{}) {
	log.Output(2, prefix+fmt.Sprintf(format, v...))
}

// Logln print debug message to the standard logger if enabled.
// Arguments are handled in the manner of fmt.Println.
func Logln(v ...interface{}) {
	log.Output(2, prefix+fmt.Sprintln(v...))
}
