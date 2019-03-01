package c128

import (
	"log"

	"github.com/blackchip-org/retro-cs/rcs"
)

var mmuRegs = []string{"A", "B", "C", "D"}

// MMU is the memory management unit
type MMU struct {
	mem  *rcs.Memory // configuration register is the bank number
	lcr  [4]uint8    // load configuration register
	pcr  [4]uint8    // pre-configuration register
	Mode uint8

	WatchCR  rcs.FlagRW
	WatchLCR rcs.FlagRW
	WatchPCR rcs.FlagRW
}

func NewMMU(mem *rcs.Memory) *MMU {
	return &MMU{
		mem: mem,
	}
}

func (m *MMU) CR() uint8 {
	v := uint8(m.mem.Bank())
	if m.WatchCR.Read {
		log.Printf("$%02x <= mmu:cr", v)
	}
	return v
}

func (m *MMU) SetCR(v uint8) {
	if m.WatchCR.Write {
		log.Printf("mmu:cr <= $%02x", v)
	}
	m.mem.SetBank(int(v))
}

func (m *MMU) LCR(i int) uint8 {
	v := m.lcr[i]
	if m.WatchLCR.Read {
		log.Printf("$%02x <= mmu:lcr %v", v, mmuRegs[i])
	}
	return v
}

func (m *MMU) SetLCR(i int, v uint8) {
	if m.WatchLCR.Write {
		log.Printf("mmu:lcr %v <= %v", mmuRegs[i], v)
	}
	m.lcr[i] = v
	m.SetCR(m.lcr[i])
}

func (m *MMU) PCR(i int) uint8 {
	v := m.pcr[i]
	if m.WatchPCR.Read {
		log.Printf("$%02x <= mmu:pcr %v", v, mmuRegs[i])
	}
	return v
}

func (m *MMU) SetPCR(i int, v uint8) {
	if m.WatchPCR.Write {
		log.Printf("mmu:pcr %v <= $%02x", mmuRegs[i], v)
	}
	m.pcr[i] = v
}
