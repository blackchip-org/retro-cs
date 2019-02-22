package namco

import (
	"log"

	"github.com/blackchip-org/retro-cs/rcs"
)

type N06XX struct {
	DeviceR [4]rcs.Load8
	DeviceW [4]rcs.Store8

	ctrl    uint8
	elapsed int
	timing  bool
	NMI     func()

	WatchDataW bool
	WatchDataR bool
	WatchCtrlW bool
	WatchCtrlR bool
	WatchNMI   bool
}

func NewN06XX() *N06XX {
	n := &N06XX{}
	for i := 0; i < 4; i++ {
		j := i
		n.DeviceR[i] = func() uint8 {
			log.Printf("n06xx device %v not mapped for read", j)
			return 0
		}
		n.DeviceW[i] = func(uint8) {
			log.Printf("n06xx device %v not mapped for write", j)
		}
	}
	return n
}

func (n *N06XX) WriteData(addr int) rcs.Store8 {
	return func(v uint8) {
		if n.ctrl&0x10 != 0 {
			return
		}
		if n.WatchDataW {
			log.Printf("n06xx data write($%04x) => $%02x\n", addr, v)
		}
		dev := n.ctrl & 0x03
		switch dev {
		case 1 << 0:
			n.DeviceW[0](v)
		case 1 << 1:
			n.DeviceW[1](v)
		case 1 << 2:
			n.DeviceW[2](v)
		case 1 << 3:
			n.DeviceW[3](v)
		}
	}
}

func (n *N06XX) ReadData(addr int) rcs.Load8 {
	return func() uint8 {
		if n.ctrl&0x10 != 0 {
			return 0
		}
		dev := n.ctrl & 0x03
		v := uint8(0xff)
		switch dev {
		case 1 << 0:
			v = n.DeviceR[0]()
		case 1 << 1:
			v = n.DeviceR[1]()
		case 1 << 2:
			v = n.DeviceR[2]()
		case 1 << 3:
			v = n.DeviceR[3]()
		}
		if n.WatchDataR {
			log.Printf("n06xx data $%02x <= read($%04x)\n", v, addr)
		}
		return v
	}
}

func (n *N06XX) WriteCtrl(addr int) rcs.Store8 {
	return func(v uint8) {
		if n.WatchCtrlW {
			log.Printf("n06xx ctrl write($%04x) => $%02x\n", addr, v)
		}
		n.ctrl = v
		if v&0x0f == 0 {
			n.timing = false
		} else {
			n.elapsed = 0
			n.timing = true
		}
	}
}

func (n *N06XX) ReadCtrl(addr int) rcs.Load8 {
	return func() uint8 {
		if n.WatchCtrlR {
			log.Printf("n06xx ctrl $%02x <= read(addr $%04x)\n", n.ctrl, addr)
		}
		return n.ctrl
	}
}

func (n *N06XX) Next() {
	if n.timing {
		n.elapsed++
		if n.elapsed > 2000 {
			if n.WatchNMI {
				log.Println("n06xx NMI")
			}
			n.NMI()
			n.elapsed = 0
		}
	}
}
