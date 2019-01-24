package mock

import (
	"github.com/blackchip-org/retro-cs/rcs"
)

// The FUSE tests want the full address space available. Creating memory
// is expensive and with the amount of tests to run it takes up to 12 seconds
// to complete. Instead, create memory once and zero out the backing
// slice each time.
var (
	TestMemory *rcs.Memory
	testRAM    []uint8
	testZero   []uint8
)

func init() {
	testRAM = make([]uint8, 0x10000, 0x10000)
	testZero = make([]uint8, 0x10000, 0x10000)
	TestMemory = rcs.NewMemory(1, 0x10000)
	TestMemory.MapRAM(0, testRAM)
}

// ResetMemory zeros out all memory values in TestMemory.
func ResetMemory() {
	copy(testRAM, testZero)
}

func MockRead(data []int) func() uint8 {
	pos := 0
	return func() uint8 {
		pos++
		return uint8(data[pos-1])
	}
}
