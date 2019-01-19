/*
Package rcs contains the common components used to create retro-computing
systems.

Memory

To create a memory space with a 16 line address bus and a single bank, use:

	mem := rcs.NewMemory(1, 0x10000)

This struct is just a container for the address space and has no actual
memory mapped to it yet. Reads or writes to an unmapped address emit a
warning through the standard logger.

Single values are mapped for read/write access using the MapRW method. This is
useful for mapping ports or registers of a device into the address space. In
this example from the Commodore 64, the X coordinate for sprite #0 is mapped to
0xd000 and the Y coordinate is mapped to 0xd001:

	mem.MapRW(0xd000, &sprites[0].X)
	mem.MapRW(0xd001, &sprites[0].Y)

Use MapRO for a read-only mapping and MapWO for a write-only mapping. In
this example from Pac-Man, the value of port IN0 is read through address
0x5000 but interrupts can be enabled or disabled by writing to address
0x5000:

	mem.MapRO(0x5000, &portIN0)
	mem.MapWO(0x5000, &irqEnable)

A range of addresses can map to the same value with:

	for i := 0x50c0; i <= 0x50ff; i++ {
		mem.MapWO(i, &watchdogReset)
	}

Large blocks can be mapped by passing in a uint8 slice using MapRAM
for read/write access and MapROM for read-only access. The following example
maps a 16KB block of ROM to 0x0000 - 0x3fff and a 48KB block of RAM
to 0x4000 - 0xffff:

	rom := []uint8{ ... 16KB data ... }
	ram := make([]uint8, 48*1024, 48*1024)
	mem.MapROM(0x0000, rom)
	mem.MapRAM(0x4000, ram)

To "overlay" ROM on top of RAM, first use MapRAM and then MapROM which will
replace the read mappings while leaving the write mappings untouched. In
the following example, reads in the first 16KB come from the ROM but
writes go to the RAM:

	rom := []uint8{ ... 16KB data ... }
	ram := make([]uint8, 64*1024, 64*1024)
	mem.MapRAM(0x0000, ram)
	mem.MapROM(0x0000, rom)

To use banked memory, create memory with more than one bank:

	mem := rcs.NewMemory(2, 0x10000)

All map, read, and write operations are performed on the selected bank
which by default is zero. The following example has a ROM overlay
in bank 0 and full access to the RAM in bank 1:

	rom := []uint8{ ... 16KB data ... }
	ram := make([]uint8, 64*1024, 64*1024)

	mem.SetBank(0)
	mem.MapRAM(0x0000, ram)
	mem.MapROM(0x0000, rom)

	mem.SetBank(1)
	mem.MapRAM(0x0000, ram)
*/
package rcs
