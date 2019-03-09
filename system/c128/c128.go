// Package c128 is the Commodore 128.
package c128

import (
	"github.com/blackchip-org/retro-cs/config"
	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/blackchip-org/retro-cs/rcs/cbm"
	"github.com/blackchip-org/retro-cs/rcs/m6502"
)

type System struct {
	cpu    *m6502.CPU
	mem    *rcs.Memory
	mmu    *MMU
	screen rcs.Screen
	vdc    *VDC
	vic    *cbm.VIC

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
	s.IO.MapRAM(0, s.IORAM)

	s.mmu = NewMMU(s.mem)
	s.vdc = NewVDC()
	v, err := cbm.NewVIC(ctx.Renderer, s.mem, roms["chargen"])
	if err != nil {
		return nil, err
	}
	s.vic = v
	s.screen = rcs.Screen{
		W:         v.W,
		H:         v.H,
		Texture:   v.Texture,
		ScanLineH: true,
		Draw:      v.Draw,
	}

	// IO mappings
	s.IO.MapRW(0x020, &s.vic.BorderColor)
	s.IO.MapRW(0x021, &s.vic.BgColor)
	s.IO.MapLoad(0x500, s.mmu.ReadCR)
	s.IO.MapStore(0x500, s.mmu.WriteCR)
	// PCR
	for i := 0; i < 4; i++ {
		i := i
		s.IO.MapLoad(0x501+i, func() uint8 { return s.mmu.ReadPCR(i) })
		s.IO.MapStore(0x501+i, func(v uint8) { s.mmu.WritePCR(i, v) })
	}
	// HACK
	s.IO.MapLoad(0x505, func() uint8 {
		return (1 << 7) | (1 << 4) | (1 << 5)
	})
	// HACK
	s.IO.MapLoad(0xd00, func() uint8 {
		return (1 << 7) | (1 << 6)
	})

	s.IO.MapLoad(0x600, s.vdc.ReadStatus)
	s.IO.MapStore(0x600, s.vdc.WriteAddr)
	s.IO.MapLoad(0x601, s.vdc.ReadData)
	s.IO.MapStore(0x601, s.vdc.WriteData)

	// map banks
	for _, i := range usedBanks {
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
			s.mem.MapROM(0xd000, s.CharGen)
			s.mem.MapROM(0xc000, s.Kernal)
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
			s.mem.MapROM(0x4000, s.BasicLo)
		case 1:
			// RAM
		}

		switch blockIO {
		case 0:
			s.mem.Map(0xd000, s.IO)
		case 1:
			// RAM or ROM as selected by bits 4 and 5
		}

		s.mem.MapLoad(0xff00, s.mmu.ReadCR)
		s.mem.MapStore(0xff00, s.mmu.WriteCR)
		for i := 0; i < 4; i++ {
			i := i
			s.mem.MapLoad(0xff01+i, func() uint8 { return s.mmu.ReadLCR(i) })
			s.mem.MapStore(0xff01+i, func(v uint8) { s.mmu.WriteLCR(i, v) })
		}
	}
	s.mem.SetBank(0) // bank 15

	s.mem.Write(0xd011, 0xff) // HACK
	s.mem.Write(0xd012, 0x09) // HACK: bcc on $8
	s.mem.Write(0xd600, 0xff) // HACK
	s.mem.Write(0xdc01, 0xff) // HACK: keyboard no press

	s.cpu = m6502.New(s.mem)
	mach := &rcs.Mach{
		Sys: s,
		Comps: []rcs.Component{
			rcs.NewComponent("cpu", "m6502", "mem", s.cpu),
			rcs.NewComponent("mem", "mem", "", s.mem),
			rcs.NewComponent("mmu", "c128/mmu", "", s.mmu),
			rcs.NewComponent("vdc", "c128/vdc", "", s.vdc),
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
		Screen: s.screen,
	}
	return mach, nil
}

var usedBanks = []int{
	0x3f, // bank 0
	0x7f, // bank 1
	0x01, // bank 14
	0x00, // bank 15

	0xc0, // on init
	0x80, // on init
	0x40, // on init
	0x2a, // on init
	0x16, // on init
}
