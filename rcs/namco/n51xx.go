package namco

import (
	"github.com/blackchip-org/retro-cs/rcs"
)

type N51XX struct {
	debug rcs.Debugger
}

func NewN51XX() *N51XX {
	return &N51XX{
		debug: rcs.NewDebugger("NAMCO_51XX"),
	}
}

func (n *N51XX) Write(v uint8) {
	n.debug.Printf("write input controller: %02v\n", v)
}

func (n *N51XX) Read() uint8 {
	n.debug.Printf("read input controller\n")
	return 0
}
