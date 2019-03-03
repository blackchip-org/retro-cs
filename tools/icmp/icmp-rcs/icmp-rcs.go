package main

import (
	"fmt"

	"github.com/blackchip-org/retro-cs/mock"
	"github.com/blackchip-org/retro-cs/rcs/z80"
)

func main() {
	af()
}

func cpd() {
	for a := 0; a <= 0xff; a++ {
		for _hl := 0; _hl <= 0xff; _hl++ {
			for f := 0; f <= 1; f++ {
				for c := 2; c <= 2; c++ {
					mock.ResetMemory()
					mem := mock.TestMemory
					cpu := z80.New(mem)
					mem.Write(0x1234, uint8(_hl))
					cpu.H, cpu.L = 0x12, 0x34
					cpu.C = uint8(c)
					cpu.A = uint8(a)
					cpu.F = uint8(f)
					mem.WriteN(0x00, 0xed, 0xb9)
					cpu.Next()
					fmt.Printf("a:%02x _hl:%02x c:%02x f:%02x => b:%02x c:%02x f:%02x\n", a, _hl, c, f, cpu.B, cpu.C, cpu.F)
				}
			}
		}
	}
}

func af() {
	for f := 0; f <= 0xff; f++ {
		for a := 0; a <= 0xff; a++ {
			mock.ResetMemory()
			mem := mock.TestMemory
			cpu := z80.New(mem)
			mem.Write(0x00, 0x17)
			cpu.A = uint8(a)
			cpu.F = uint8(f)
			cpu.Next()
			fmt.Printf("a:%02x f:%02x => a:%02x f:%02x\n", a, f, cpu.A, cpu.F)
		}
	}
}
