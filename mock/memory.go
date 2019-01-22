package mock

import (
	"fmt"
	"strings"

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

// Slice is a representation of a series of memory values at a given address.
type Slice struct {
	Addr   int
	Values []uint8
}

// Diff is the difference between two memory values (A and B) at Address.
type Diff struct {
	Addr int
	A    uint8
	B    uint8
}

func (d *Diff) String() string {
	return fmt.Sprintf("%04x: %02x %02x", d.Addr, d.A, d.B)
}

// DiffReport is a list of differences.
type DiffReport []Diff

func (d DiffReport) String() string {
	reports := make([]string, 0, 0)
	for _, diff := range d {
		reports = append(reports, diff.String())
	}
	return strings.Join(reports, "\n")
}

// Verify checks that the values in the slices match up with the values in
// memory. Returns true if all snapshot values match.
func Verify(a *rcs.Memory, b []Slice) (DiffReport, bool) {
	diff := make([]Diff, 0, 0)
	ptr := rcs.NewPointer(a)
	for _, slice := range b {
		ptr.Addr = slice.Addr
		for i, bval := range slice.Values {
			aval := ptr.Fetch()
			if aval != bval {
				diff = append(diff, Diff{
					Addr: slice.Addr + i,
					A:    aval,
					B:    bval,
				})
			}
		}
	}
	return diff, len(diff) == 0
}

// Import loads memory with the values in the slice.
func Import(m *rcs.Memory, slice Slice) {
	for i, value := range slice.Values {
		m.Write(slice.Addr+i, value)
	}
}
