package rcs

import (
	"math/bits"
)

// ALU is an 8-bit arithmetic logic unit.
type ALU struct {
	C uint8 // carry flag
	V uint8 // overflow flag
	P uint8 // parity flag
	H uint8 // half carry flag
	Z uint8 // zero flag
	S uint8 // sign flag
}

// Add performs a binary-coded decimal addition of in0 and in1 and
// places the result in out. Results are undefined if either value is not a
// valid BCD number. If the carry flag in is set, the result is incremented
// by one. The Z and S flags are updated.
func (a *ALU) Add(flags *uint8, out *uint8, in0 uint8, in1 uint8) {
	carry := 0
	if *flags&a.C != 0 {
		carry = 1
	}

	// result of 8 bit addition into 16 bits
	r := uint16(in0) + uint16(in1) + uint16(carry)
	// signed result, 16-bit
	sr := int16(int8(in0)) + int16(int8(in1)) + int16(carry)
	// unsigned result, 8-bit
	ur := uint8(r)
	// result of half add
	hr := in0&0xf + in1&0xf + uint8(carry)

	a.carry(flags, r)
	a.carry4(flags, hr)
	a.overflow(flags, sr)
	a.parity(flags, ur)
	a.zero(flags, ur)
	a.sign(flags, ur)
	*out = ur
}

// AddBCD performs a binary-coded decimal addition of in0 and in1 and
// places the result in out. Results are undefined if either value is not a
// valid BCD number. If the carry flag is set, the result is incremented
// by one. The Z and S flags are updated.
func (a *ALU) AddBCD(flags *uint8, out *uint8, in0 uint8, in1 uint8) {
	carry := 0
	if *flags&a.C != 0 {
		carry = 1
	}

	in0b := FromBCD(in0)
	in1b := FromBCD(in1)
	r := uint16(in0b) + uint16(in1b) + uint16(carry)
	rb := ToBCD(uint8(r))

	a.carryBCD(flags, r)
	a.zero(flags, rb)
	a.sign(flags, rb)
	*out = rb
}

// And performs a logical "and" between in0 and in1 and places the result
// in out. Flags P, Z, and S are updated.
func (a *ALU) And(flags *uint8, out *uint8, in0 uint8, in1 uint8) {
	r := in0 & in1
	a.parity(flags, r)
	a.zero(flags, r)
	a.sign(flags, r)
	*out = r
}

// ShiftLeft performs a left bit-shift of in and places the result in out.
// Bit 0 becomes the value of the carry. Bit 7, that is shifted out, becomes
// the new value of carry. Flags C, P, Z, and S are updated.
func (a *ALU) ShiftLeft(flags *uint8, out *uint8, in uint8) {
	carryOut := in&0x80 != 0
	r := in << 1
	if *flags&a.C != 0 {
		r++
	}

	a.parity(flags, r)
	a.zero(flags, r)
	a.sign(flags, r)
	if carryOut {
		*flags |= a.C
	} else {
		*flags &^= a.C
	}
	*out = r
}

func (a *ALU) carry(f *uint8, v uint16) {
	if v > 0xff {
		*f |= a.C
	} else {
		*f &^= a.C
	}
}

func (a *ALU) carryBCD(f *uint8, v uint16) {
	if v > 99 {
		*f |= a.C
	} else {
		*f &^= a.C
	}
}

func (a *ALU) carry4(f *uint8, v uint8) {
	if v > 0xf {
		*f |= a.H
	} else {
		*f &^= a.H
	}
}

func (a *ALU) overflow(f *uint8, v int16) {
	if v < -128 || v > 127 {
		*f |= a.V
	} else {
		*f &^= a.V
	}
}

func (a *ALU) parity(f *uint8, v uint8) {
	if bits.OnesCount8(v)%2 == 0 {
		*f |= a.P
	} else {
		*f &^= a.P
	}
}

func (a *ALU) zero(f *uint8, v uint8) {
	if v == 0 {
		*f |= a.Z
	} else {
		*f &^= a.Z
	}
}

func (a *ALU) sign(f *uint8, v uint8) {
	if v&0x80 != 0 {
		*f |= a.S
	} else {
		*f &^= a.S
	}
}
