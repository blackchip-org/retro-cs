package mock

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/blackchip-org/retro-cs/rcs"
)

func ExampleVerify() {
	mem := rcs.NewMemory(1, 10)
	rom := []uint8{0x12, 0x34, 0x45, 0x67}
	mem.MapROM(0, rom)

	b := []Slice{
		Slice{Addr: 0, Values: []uint8{0x12, 0xff}},
		Slice{Addr: 2, Values: []uint8{0x45, 0xff}},
	}
	diff, equal := Verify(mem, b)
	if !equal {
		fmt.Println(diff.String())
	}
	// Output:
	// 0001: 34 ff
	// 0003: 67 ff
}

func TestImport(t *testing.T) {
	mem := rcs.NewMemory(1, 5)
	ram := make([]uint8, 5, 5)
	mem.MapRAM(0, ram)

	Import(mem, Slice{Addr: 1, Values: []uint8{11, 22, 33}})
	want := []uint8{0, 11, 22, 33, 0}

	if !reflect.DeepEqual(ram, want) {
		t.Errorf("\n have: %v \n want: %v", ram, want)
	}
}
