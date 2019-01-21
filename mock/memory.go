package mock

import (
	"fmt"
	"strings"

	"github.com/blackchip-org/retro-cs/rcs"
)

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
