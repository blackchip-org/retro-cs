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

	// If false, a borrow is used during subraction when the carry is set.
	// If true, a borrow is used during subraction when the carry is clear.
	ClearBorrow bool

	// Ignore is a mask of flags to ignore when performing operations.
	Ignore uint8
}

// Add performs addition of in0 and in1 and returns the results. If the carry
// flag in is set, the result is incremented by one. The Z and S flags are
// updated.
func (a ALU) Add(flags *uint8, in0 uint8, in1 uint8) uint8 {
	// https://stackoverflow.com/questions/8034566/overflow-and-carry-flags-on-z80/8037485#8037485
	var out uint8
	var carryOut uint8

	if *flags&a.C != 0 {
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
	halfCarry := carryIns & (1 << 4)
	overflow := (carryIns >> 7) ^ carryOut
	p := bits.OnesCount8(out)

	a.setFlag(flags, a.C, carryOut != 0)
	a.setFlag(flags, a.H, halfCarry != 0)
	a.setFlag(flags, a.V, overflow != 0)
	a.setFlag(flags, a.P, p == 0 || p == 2 || p == 4 || p == 6)
	a.setFlag(flags, a.Z, out == 0)
	a.setFlag(flags, a.S, out&(1<<7) != 0)

	return out
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
	if !a.ClearBorrow {
		*flags ^= a.C
	}
	out := a.Add(flags, in0, ^in1)
	if !a.ClearBorrow {
		*flags ^= a.C
	}
	return out
}

/*
func (a ALU) Subtract2(flags *uint8, in0 uint8, in1 uint8) uint8 {
	borrow := 0
	if *flags&a.C != 0 { // carry same as borrow!
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
	// a.parity(flags, ur)
	a.overflow(flags, sr)
	// a.zero(flags, ur)
	// a.sign(flags, ur)
	return ur
}
*/

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

func (a ALU) setFlag(flags *uint8, flag uint8, set bool) {
	if flag&a.Ignore != 0 {
		return
	}
	if set {
		*flags |= flag
	} else {
		*flags &^= flag
	}
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
