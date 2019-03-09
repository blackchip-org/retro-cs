package c128

import (
	"log"

	"github.com/blackchip-org/retro-cs/rcs"
)

var mmuRegs = []string{"A", "B", "C", "D"}

// MMU is the memory management unit
type MMU struct {
	Mem *rcs.Memory // configuration register is the bank number
	LCR [4]uint8    // load configuration register
	PCR [4]uint8    // pre-configuration register
	// FIXME: What is this for?
	// Mode uint8

	WatchCR  rcs.FlagRW
	WatchLCR rcs.FlagRW
	WatchPCR rcs.FlagRW
}

func NewMMU(mem *rcs.Memory) *MMU {
	return &MMU{
		Mem: mem,
	}
}

func (m *MMU) ReadCR() uint8 {
	v := uint8(m.Mem.Bank())
	if m.WatchCR.R {
		log.Printf("0x%02x <= mmu:cr", v)
	}
	return v
}

func (m *MMU) WriteCR(v uint8) {
	if m.WatchCR.W {
		log.Printf("mmu:cr <= 0x%02x", v)
	}
	m.Mem.SetBank(int(v))
}

func (m *MMU) ReadLCR(i int) uint8 {
	v := m.LCR[i]
	if m.WatchLCR.R {
		log.Printf("0x%02x <= mmu:lcr-%v", v, mmuRegs[i])
	}
	return v
}

func (m *MMU) WriteLCR(i int, v uint8) {
	if m.WatchLCR.W {
		log.Printf("mmu:lcr-%v <= 0x%02x", mmuRegs[i], v)
	}
	m.LCR[i] = v
	// FIXME: something is wrong here
	//m.StoreCR(m.LCR[i])
}

func (m *MMU) ReadPCR(i int) uint8 {
	v := m.PCR[i]
	if m.WatchPCR.R {
		log.Printf("0x%02x <= mmu:pcr-%v", v, mmuRegs[i])
	}
	return v
}

func (m *MMU) WritePCR(i int, v uint8) {
	if m.WatchPCR.W {
		log.Printf("mmu:pcr-%v <= 0x%02x", mmuRegs[i], v)
	}
	m.PCR[i] = v
}
