package namco

import "log"

type N54XX struct {
	WatchR bool
	WatchW bool
}

func NewN54XX() *N54XX {
	return &N54XX{}
}

func (n *N54XX) Write(v uint8) {
	if n.WatchW {
		log.Printf("write noise generator: %02v\n", v)
	}
}

func (n *N54XX) Read() uint8 {
	if n.WatchR {
		log.Printf("read noise generator\n")
	}
	return 0
}
