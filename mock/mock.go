// Package mock contains testing mocks and other various utilities
// for testing code.
package mock

import "bytes"

// PanicWriter is a write that panics after the first newline has been
// written. Useful for tracking down where warnings are being emitted
// in the logger.
type PanicWriter struct {
	buf bytes.Buffer
}

// Write appends the output to an internal buffer and then panics with
// the written text once a newline is found.
func (p *PanicWriter) Write(out []byte) (int, error) {
	p.buf.Write(out)
	for _, b := range out {
		if b == '\n' {
			panic(p.buf.String())
		}
	}
	return len(out), nil
}
