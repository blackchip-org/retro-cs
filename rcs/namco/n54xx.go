package namco

import (
	"github.com/blackchip-org/retro-cs/rcs"
)

type N54XX struct {
	debug rcs.Debugger
}

func NewN54XX() *N54XX {
	return &N54XX{
		debug: rcs.NewDebugger("NAMCO_54XX"),
	}
}

func (n *N54XX) Write(v uint8) {
	n.debug.Printf("write noise generator: %02v\n", v)
}

func (n *N54XX) Read() uint8 {
	n.debug.Printf("read noise generator\n")
	return 0
}
