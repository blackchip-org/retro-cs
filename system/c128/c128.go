// Package c128 is the Commodore 128.
package c128

import (
	"github.com/blackchip-org/retro-cs/config"
	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/blackchip-org/retro-cs/rcs/cbm"
	"github.com/blackchip-org/retro-cs/rcs/m6502"
)

type System struct {
	cpu *m6502.CPU
	mem *rcs.Memory
	mmu *MMU

	BasicLo []uint8
	BasicHi []uint8
	CharGen []uint8
	Kernal  []uint8
	RAM0    []uint8
	RAM1    []uint8
	IORAM   []uint8
	IO      *rcs.Memory
}

func New(ctx rcs.SDLContext) (*rcs.Mach, error) {
	s := &System{}
	roms, err := rcs.LoadROMs(config.DataDir, SystemROM)
	if err != nil {
		return nil, err
	}
	s.mem = rcs.NewMemory(256, 0x10000)
	s.BasicLo = roms["basiclo"]
	s.BasicHi = roms["basichi"]
	s.CharGen = roms["chargen"]
	s.Kernal = roms["kernal"]
	s.RAM0 = make([]uint8, 0x10000, 0x10000)
	s.RAM1 = make([]uint8, 0xc000, 0xc000)
	s.IORAM = make([]uint8, 0x1000, 0x1000)
	s.IO = rcs.NewMemory(1, 0x1000)

	s.mmu = NewMMU(s.mem)
	s.IO.MapLoad(0x500, s.mmu.CR)
	s.IO.MapStore(0x500, s.mmu.SetCR)
	for i := 0; i < 4; i++ {
		i := i
		s.IO.MapLoad(0x501+i, func() uint8 { return s.mmu.PCR(i) })
		s.IO.MapStore(0x501+i, func(v uint8) { s.mmu.SetPCR(i, v) })
	}

	// map banks
	for i := 0; i < 256; i++ {
		s.mem.SetBank(i)
		cr := uint8(i)
		blockRAM := rcs.SliceBits(cr, 6, 7)
		blockC000 := rcs.SliceBits(cr, 4, 5)
		block8000 := rcs.SliceBits(cr, 2, 3)
		block4000 := rcs.SliceBits(cr, 1, 1)
		blockIO := rcs.SliceBits(cr, 0, 0)

		switch blockRAM {
		case 0:
			s.mem.MapRAM(0x0000, s.RAM0)
		case 1:
			s.mem.MapRAM(0x0000, s.RAM0)
			s.mem.MapRAM(0x0400, s.RAM1)
		// no block RAM 2 or 3. If set, accesses 0 and 1 instead.
		case 2:
			s.mem.MapRAM(0x0000, s.RAM0)
		case 3:
			s.mem.MapRAM(0x0000, s.RAM0)
			s.mem.MapRAM(0x0400, s.RAM1)
		}

		switch blockC000 {
		case 0:
			s.mem.MapROM(0xc000, s.Kernal)
			s.mem.MapROM(0xd000, s.CharGen)
		case 1:
			// internal function ROM
		case 2:
			// external function ROM
		case 3:
			// RAM
		}

		switch block8000 {
		case 0:
			s.mem.MapROM(0x8000, s.BasicHi)
		case 1:
			// internal function ROM
		case 2:
			// external function ROM
		case 3:
			// RAM
		}

		switch block4000 {
		case 0:
			s.mem.MapROM(0x8000, s.BasicLo)
		case 1:
			// RAM
		}

		switch blockIO {
		case 0:
			s.mem.Map(0xd000, s.IO)
		case 1:
			// RAM or ROM as selected by bits 4 and 5
		}

		s.mem.MapLoad(0xff00, s.mmu.CR)
		s.mem.MapStore(0xff00, s.mmu.SetCR)
		for i := 0; i < 4; i++ {
			i := i
			s.mem.MapLoad(0xff01+i, func() uint8 { return s.mmu.LCR(i) })
			s.mem.MapStore(0xff01+i, func(v uint8) { s.mmu.SetLCR(i, v) })
		}
	}
	s.mem.SetBank(0) // bank 15
	s.cpu = m6502.New(s.mem)

	mach := &rcs.Mach{
		Sys: s,
		Comps: []rcs.Component{
			rcs.NewComponent("cpu", "m6502", "mem", s.cpu),
			rcs.NewComponent("mem", "mem", "", s.mem),
			rcs.NewComponent("mmu", "c128/mmu", "", s.mmu),
		},
		CharDecoders: map[string]rcs.CharDecoder{
			"petscii":         cbm.PetsciiDecoder,
			"petscii-shifted": cbm.PetsciiShiftedDecoder,
			"screen":          cbm.ScreenDecoder,
			"screen-shifted":  cbm.ScreenShiftedDecoder,
		},
		DefaultEncoding: "petscii",
		Ctx:             ctx,
		VBlankFunc: func() {
			s.cpu.IRQ = true
		},
	}
	return mach, nil
}

func mapBanks(s *System) {
}
