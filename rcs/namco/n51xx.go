package namco

import "log"

type N51XX struct {
	WatchR bool
	WatchW bool
}

func NewN51XX() *N51XX {
	return &N51XX{}
}

func (n *N51XX) Write(v uint8) {
	if n.WatchW {
		log.Printf("write input controller: %02v", v)
	}
}

func (n *N51XX) Read() uint8 {
	if n.WatchR {
		log.Printf("read input controller")
	}
	return 0
}
