package rcs

import (
	"fmt"
	"testing"
)

func ExampleFromBCD() {
	v := FromBCD(0x42)
	fmt.Println(v)
	// Output: 42
}

func ExampleToBCD() {
	v := ToBCD(42)
	fmt.Printf("%02x", v)
	// Output: 42
}

func TestToBCDOverflow(t *testing.T) {
	want := uint8(0x12)
	have := ToBCD(112)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x\n", want, have)
	}
}
