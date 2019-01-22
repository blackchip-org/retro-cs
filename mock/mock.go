// Package mock contains testing mocks and other various utilities
// for testing code.
package mock

import "bytes"

type PanicWriter struct {
	buf bytes.Buffer
}

func (p *PanicWriter) Write(out []byte) (int, error) {
	p.buf.Write(out)
	for _, b := range out {
		if b == '\n' {
			panic(p.buf.String())
		}
	}
	return len(out), nil
}
