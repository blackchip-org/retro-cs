package main

import (
	"fmt"

	"github.com/blackchip-org/retro-cs/mock"
	"github.com/blackchip-org/retro-cs/rcs/z80"
)

func main() {
	// Inputs for the DAA instruction are the accumulator and the status
	// flags. Iterate through all values.
	for a := 0; a <= 0xff; a++ {
		for f := 0; f <= 0xff; f++ {
			// Zero out the testing memory and create a CPU
			mock.ResetMemory()
			mem := mock.TestMemory
			cpu := z80.New(mem)
			// Write the one instruction program to memory.
			mem.Write(0x00, 0x27) // daa
			// Set the flags and execute
			cpu.A = uint8(a)
			cpu.F = uint8(f)
			cpu.Next()
			// Print the contents of A and F before the operation on one line
			// and after the operation on the next line
			fmt.Printf("%02x %02x\n%02x %02x\n\n", a, f, cpu.A, cpu.F)
		}
	}
}
