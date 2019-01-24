package rcs

import (
	"math/bits"
	"strconv"
)

// Load8 is a function which loads an unsigned 8-bit value
type Load8 func() uint8

// Store8 is a function which stores an unsiged 8-bit value
type Store8 func(uint8)

// Load is a function which loads an integer value
type Load func() int

// Store is a function which stores an integer value
type Store func(int)

// FromBCD converts a binary-coded decimal to an integer value.
func FromBCD(v uint8) uint8 {
	low := v & 0x0f
	high := v >> 4
	return high*10 + low
}

// ToBCD converts an integer value to a binary-coded decimal.
func ToBCD(v uint8) uint8 {
	low := v % 10
	high := (v / 10) % 10
	return high<<4 | low
}

// ParseBits parses the base-2 string value s to a uint8. Panics if s is not
// a valid number. Use strconv.ParseUint for input which may be malformed.
func ParseBits(s string) uint8 {
	value, err := strconv.ParseUint(s, 2, 8)
	if err != nil {
		panic(err)
	}
	return uint8(value)
}

// SliceBits extracts a sequence of bits in value from bit lo to bit hi,
// inclusive.
func SliceBits(value uint8, lo int, hi int) uint8 {
	value = value >> uint(lo)
	bits := uint(hi - lo + 1)
	mask := uint8(1)<<bits - 1
	return value & mask
}

func Add(in0, in1 uint8, carry bool) (out uint8, c, h, v bool) {
	// https://stackoverflow.com/questions/8034566/overflow-and-carry-flags-on-z80/8037485#8037485
	var carryOut uint8

	if carry {
		if in0 >= 0xff-in1 {
			carryOut = 1
		}
		out = in0 + in1 + 1
	} else {
		if in0 > 0xff-in1 {
			carryOut = 1
		}
		out = in0 + in1
	}
	carryIns := out ^ in0 ^ in1

	c = carryOut != 0
	h = carryIns&(1<<4) != 0
	v = (carryIns>>7)^carryOut != 0
	return
}

func Sub(in0, in1 uint8, borrow bool) (out uint8, fc, fh, fv bool) {
	fc = !borrow
	out, fc, fh, fv = Add(in0, ^in1, fc)
	fc = !fc
	fh = !fh
	return
}

func Parity8(v uint8) bool {
	p := bits.OnesCount8(v)
	return p == 0 || p == 2 || p == 4 || p == 6 || p == 8
}
