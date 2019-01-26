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

func TestSliceBits(t *testing.T) {
	b := ParseBits
	tests := []struct {
		lo   int
		hi   int
		in   uint8
		out  uint8
		name string
	}{
		{6, 7, b("11000000"), b("011"), "high one"},
		{6, 7, b("00111111"), b("000"), "high zero"},
		{3, 5, b("00111000"), b("111"), "middle one"},
		{3, 5, b("11000111"), b("000"), "middle zero"},
		{0, 2, b("00000111"), b("111"), "low one"},
		{0, 2, b("11111000"), b("000"), "low zero"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			slice := SliceBits(test.in, test.lo, test.hi)
			if slice != test.out {
				t.Errorf("\n have: %08b \n want: %08b", slice, test.out)
			}
		})
	}
}

func ExampleSliceBits() {
	value := ParseBits("00111000")
	fmt.Printf("%03b", SliceBits(value, 3, 5))
	// Output: 111
}

func TestBitPlane(t *testing.T) {
	b := ParseBits
	p := []int{0, 4}
	tests := []struct {
		offset int
		in     uint8
		out    uint8
	}{
		{0, b("00010001"), b("11")},
		{1, b("00100010"), b("11")},
		{2, b("01000100"), b("11")},
		{3, b("10001000"), b("11")},
	}
	for _, test := range tests {
		out := BitPlane(test.in, p, test.offset)
		if out != test.out {
			t.Errorf("\n have: %08b \n want: %08b", out, test.out)
		}
	}
}
