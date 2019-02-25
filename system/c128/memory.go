package c128

import "github.com/blackchip-org/retro-cs/rcs"

func newMemory(s *System) *rcs.Memory {
	mem := rcs.NewMemory(256, 0x10000)

	for i := 0; i < 256; i++ {
		mem.SetBank(i)
		cr := uint8(i)
		blockRAM := rcs.SliceBits(cr, 6, 7)
		blockC000 := rcs.SliceBits(cr, 4, 5)
		block8000 := rcs.SliceBits(cr, 2, 3)
		block4000 := rcs.SliceBits(cr, 1, 1)
		blockIO := rcs.SliceBits(cr, 0, 0)

		switch blockRAM {
		case 0:
			mem.MapRAM(0x0000, s.RAM0)
		case 1:
			mem.MapRAM(0x0000, s.RAM0)
			mem.MapRAM(0x0400, s.RAM1)
		// no block RAM 2 or 3. If set, accesses 0 and 1 instead.
		case 2:
			mem.MapRAM(0x0000, s.RAM0)
		case 3:
			mem.MapRAM(0x0000, s.RAM0)
			mem.MapRAM(0x0400, s.RAM1)
		}

		switch blockC000 {
		case 0:
			mem.MapROM(0xc000, s.Kernal)
			mem.MapROM(0xd000, s.CharGen)
		case 1:
			// internal function ROM
		case 2:
			// external function ROM
		case 3:
			// RAM
		}

		switch block8000 {
		case 0:
			mem.MapROM(0x8000, s.BasicHi)
		case 1:
			// internal function ROM
		case 2:
			// external function ROM
		case 3:
			// RAM
		}

		switch block4000 {
		case 0:
			mem.MapROM(0x8000, s.BasicLo)
		case 1:
			// RAM
		}

		switch blockIO {
		case 0:
			mem.MapROM(0xd000, s.IO)
		case 1:
			// RAM or ROM as selected by bits 4 and 5
		}
	}
	return mem
}
