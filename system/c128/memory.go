package c128

import "github.com/blackchip-org/retro-cs/rcs"

func newMemory(s *System) *rcs.Memory {
	// Simplified, only map the useful 4 of the 16 "defined" banks
	mem := rcs.NewMemory(16, 0x10000)

	mem.SetBank(0)
	mem.MapRAM(0x0000, s.RAM0)

	mem.SetBank(1)
	mem.MapRAM(0x0000, s.RAM0)
	mem.MapRAM(0x0400, s.RAM1)

	mem.SetBank(14)
	mem.MapRAM(0x0000, s.RAM0)
	mem.MapROM(0x4000, s.BasicLo)
	mem.MapROM(0x8000, s.BasicHi)
	mem.MapROM(0xc000, s.Kernal)
	mem.MapROM(0xd000, s.CharGen)

	mem.SetBank(15)
	mem.MapRAM(0x0000, s.RAM0)
	mem.MapROM(0x4000, s.BasicLo)
	mem.MapROM(0x8000, s.BasicHi)
	mem.MapROM(0xc000, s.Kernal)
	mem.MapROM(0xd000, s.IO)

	return mem
}
