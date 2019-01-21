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

// Add performs addition of in0 and in1 and returns the results. If the carry
// flag in is set, the result is incremented by one. The Z and S flags are
// updated.
func (a ALU) Add(flags *uint8, in0 uint8, in1 uint8) uint8 {
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
	return ur
}

// AddBCD performs a binary-coded decimal addition of in0 and in1 and
// returns the result. Results are undefined if either value is not a
// valid BCD number. If the carry flag is set, the result is incremented
// by one. The Z and S flags are updated.
func (a ALU) AddBCD(flags *uint8, in0 uint8, in1 uint8) uint8 {
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
	return rb
}

// And performs a logical "and" between in0 and in1 and returns the result.
// Flags P, Z, and S are updated.
func (a ALU) And(flags *uint8, in0 uint8, in1 uint8) uint8 {
	r := in0 & in1
	a.parity(flags, r)
	a.zero(flags, r)
	a.sign(flags, r)
	return r
}

// Decrement subracts one from in and returns the results. Only the P, Z,
// and S flags are updated.
func (a ALU) Decrement(flags *uint8, in uint8) uint8 {
	r := in - 1
	a.parity(flags, r)
	a.zero(flags, r)
	a.sign(flags, r)
	return r
}

// ExclusiveOr performs a logical exclusive-or between in0 and in1 and returns
// the results. The P, Z, and S flags are updated.
func (a ALU) ExclusiveOr(flags *uint8, in0 uint8, in1 uint8) uint8 {
	r := in0 ^ in1
	a.parity(flags, r)
	a.zero(flags, r)
	a.sign(flags, r)
	return r
}

// Increment add one from in and returns the results. Only the P, Z, and S
// flags are updated.
func (a ALU) Increment(flags *uint8, in uint8) uint8 {
	r := in + 1
	a.parity(flags, r)
	a.zero(flags, r)
	a.sign(flags, r)
	return r
}

// Pass performs a pass-through of the value in and adjusts the P, Z, and
// S flags.
func (a ALU) Pass(flags *uint8, in uint8) {
	a.parity(flags, in)
	a.zero(flags, in)
	a.sign(flags, in)
}

// ShiftLeft performs a left bit-shift of in and returns the result.
// Bit 0 becomes the value of the carry. Bit 7, that is shifted out, becomes
// the new value of carry. The  C, P, Z, and S flags are updated.
func (a ALU) ShiftLeft(flags *uint8, in uint8) uint8 {
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
	return r
}

// ShiftRight performs a right bit-shift of in and returns the result. Bit 7
// becomes the value of the carry. Bit 0, that is shifted out, becomes the
// new value of carry. The C, P, Z, and S flags are updated.
func (a ALU) ShiftRight(flags *uint8, in uint8) uint8 {
	carryOut := in&0x01 != 0
	r := in >> 1
	if *flags&a.C != 0 {
		r |= (1 << 7)
	}

	a.parity(flags, r)
	a.zero(flags, r)
	a.sign(flags, r)
	if carryOut {
		*flags |= a.C
	} else {
		*flags &^= a.C
	}
	return r
}

// Subtract performs subtraction of in1 from in0 and returns the
// results. If the carry flag in is set, the result is incremented by one.
// The Z and S flags are updated.
func (a ALU) Subtract(flags *uint8, in0 uint8, in1 uint8) uint8 {
	borrow := 0
	if *flags&a.C == 0 { // borrow if carry clear
		borrow = 1
	}

	// result of 8 bit addition into 16 bits
	r := int16(in0) - int16(in1) - int16(borrow)
	// signed result, 16-bit
	sr := int16(int8(in0)) - int16(int8(in1)) - int16(borrow)
	// unsigned result, 8-bit
	ur := uint8(r)
	// FIXME: result of half subtraction
	// hr := in0&0xf - in1&0xf - uint8(borrow)

	a.borrow(flags, r)
	// a.carry4(flags, hr)
	a.overflow(flags, sr)
	a.parity(flags, ur)
	a.zero(flags, ur)
	a.sign(flags, ur)
	return ur
}

// SubtractBCD performs a binary-coded decimal subtraction of in1 from in0 and
// returns the result. Results are undefined if either value is not a
// valid BCD number. If the carry flag is set, the result is incremented
// by one. The Z and S flags are updated.
func (a ALU) SubtractBCD(flags *uint8, in0 uint8, in1 uint8) uint8 {
	borrow := 0
	if *flags&a.C == 0 {
		borrow = 1 // borrow on carry clear
	}

	in0b := FromBCD(in0)
	in1b := FromBCD(in1)
	r := int16(in0b) - int16(in1b) - int16(borrow)
	if r < 0 {
		r += 100
	}
	rb := ToBCD(uint8(r))

	a.borrow(flags, r)
	a.zero(flags, rb)
	a.sign(flags, rb)
	return rb
}

// Or performs a logical or between in0 and in1 and returns the results.
// The P, Z, and S flags are updated.
func (a ALU) Or(flags *uint8, in0 uint8, in1 uint8) uint8 {
	r := in0 | in1
	a.parity(flags, r)
	a.zero(flags, r)
	a.sign(flags, r)
	return r
}

func (a ALU) borrow(f *uint8, v int16) {
	if v >= 0 {
		*f |= a.C
	} else {
		*f &^= a.C // carry clear on borrow
	}
}

func (a ALU) carry(f *uint8, v uint16) {
	if v > 0xff {
		*f |= a.C
	} else {
		*f &^= a.C
	}
}

func (a ALU) carryBCD(f *uint8, v uint16) {
	if v > 99 {
		*f |= a.C
	} else {
		*f &^= a.C
	}
}

func (a ALU) carry4(f *uint8, v uint8) {
	if v > 0xf {
		*f |= a.H
	} else {
		*f &^= a.H
	}
}

func (a ALU) overflow(f *uint8, v int16) {
	if v < -128 || v > 127 {
		*f |= a.V
	} else {
		*f &^= a.V
	}
}

func (a ALU) parity(f *uint8, v uint8) {
	if bits.OnesCount8(v)%2 == 0 {
		*f |= a.P
	} else {
		*f &^= a.P
	}
}

func (a ALU) zero(f *uint8, v uint8) {
	if v == 0 {
		*f |= a.Z
	} else {
		*f &^= a.Z
	}
}

func (a ALU) sign(f *uint8, v uint8) {
	if v&0x80 != 0 {
		*f |= a.S
	} else {
		*f &^= a.S
	}
}
