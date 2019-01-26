package c64

import (
	"github.com/blackchip-org/retro-cs/rcs"
)

func newMemory(roms map[string][]byte) *rcs.Memory {
	ram := make([]uint8, 0x10000, 0x10000)
	basic := roms["basic"]
	kernal := roms["kernal"]
	chargen := roms["chargen"]

	io := rcs.NewMemory(1, 0x1000)
	io.MapRAM(0, make([]uint8, 0x1000, 0x1000))

	var cartlo, carthi []uint8
	cart, ok := roms["cart"]
	if ok {
		cartlo = cart[0x0000:0x2000]
		carthi = cart[0x2000:0x4000]
	}

	mem := rcs.NewMemory(32, 0x10000)

	// https://www.c64-wiki.com/wiki/Bank_Switching
	mem.SetBank(31)
	mem.MapRAM(0x0000, ram)
	mem.MapROM(0xa000, basic)
	mem.Map(0xd000, io)
	mem.MapROM(0xe000, kernal)

	for _, bank := range []int{30, 14} {
		mem.SetBank(bank)
		mem.MapRAM(0x0000, ram)
		mem.Map(0xd000, io)
		mem.MapROM(0xe000, kernal)
	}

	for _, bank := range []int{29, 13} {
		mem.SetBank(bank)
		mem.MapRAM(0x0000, ram)
		mem.Map(0xd000, io)
	}

	for _, bank := range []int{28, 24} {
		mem.SetBank(bank)
		mem.MapRAM(0x0000, ram)
	}

	mem.SetBank(27)
	mem.MapRAM(0x0000, ram)
	mem.MapROM(0xa000, basic)
	mem.MapROM(0xd000, chargen)
	mem.MapROM(0xe000, kernal)

	for _, bank := range []int{26, 10} {
		mem.SetBank(bank)
		mem.MapRAM(0x0000, ram)
		mem.MapROM(0xd000, chargen)
		mem.MapROM(0xe000, kernal)
	}

	for _, bank := range []int{25, 9} {
		mem.SetBank(bank)
		mem.MapRAM(0x0000, ram)
		mem.MapROM(0xd000, chargen)
	}

	for bank := 23; bank >= 16; bank-- {
		mem.SetBank(bank)
		mem.MapRAM(0x0000, ram)
		mem.Unmap(0x1000, 0x7fff)
		mem.MapROM(0x8000, cartlo)
		mem.Unmap(0xa000, 0xcfff)
		mem.Map(0xd000, io)
		mem.MapROM(0xe000, carthi)
	}

	mem.SetBank(15)
	mem.MapRAM(0x0000, ram)
	mem.MapROM(0x8000, cartlo)
	mem.MapROM(0xa000, basic)
	mem.Map(0xd000, io)
	mem.MapROM(0xe000, kernal)

	for _, bank := range []int{12, 8, 4, 0} {
		mem.SetBank(bank)
		mem.MapRAM(0x0000, ram)
	}

	mem.SetBank(11)
	mem.MapRAM(0x0000, ram)
	mem.MapROM(0x8000, cartlo)
	mem.MapROM(0xa000, basic)
	mem.MapROM(0xd000, chargen)
	mem.MapROM(0xe000, kernal)

	mem.SetBank(7)
	mem.MapRAM(0x0000, ram)
	mem.MapROM(0x8000, cartlo)
	mem.MapROM(0xa000, carthi)
	mem.Map(0xd000, io)
	mem.MapROM(0xe000, kernal)

	mem.SetBank(6)
	mem.MapRAM(0x0000, ram)
	mem.MapROM(0xa000, carthi)
	mem.Map(0xd000, io)
	mem.MapROM(0xe000, kernal)

	mem.SetBank(5)
	mem.MapRAM(0x0000, ram)
	mem.Map(0xd000, io)

	mem.SetBank(3)
	mem.MapRAM(0x0000, ram)
	mem.MapROM(0x8000, cartlo)
	mem.MapROM(0xa000, carthi)
	mem.MapROM(0xd000, chargen)
	mem.MapROM(0xe000, kernal)

	mem.SetBank(2)
	mem.MapRAM(0x0000, ram)
	mem.MapROM(0xa000, carthi)
	mem.MapROM(0xd000, chargen)
	mem.MapROM(0xe000, kernal)

	mem.SetBank(1)
	mem.MapRAM(0x0000, ram)

	return mem
}
