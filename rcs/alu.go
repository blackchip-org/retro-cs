package rcs

import (
	"math/bits"
)

type FlagMap struct {
	C uint8 // carry flag
	V uint8 // overflow flag
	P uint8 // parity flag
	H uint8 // half carry flag
	Z uint8 // zero flag
	S uint8 // sign flag
}

// ALU is an 8-bit arithmetic logic unit.
type ALU struct {
	// Accumulator
	Acc *uint8
	// Status register holding flag conditions
	Status *uint8

	Flags FlagMap
}

func NewALU(acc *uint8, status *uint8, flags FlagMap) *ALU {
	return &ALU{
		Acc:    acc,
		Status: status,
		Flags:  flags,
	}
}

// Add adds the value of v to alu.A. If the carry is set, increments the
// result by one.
func (a *ALU) Add(v uint8) {
	carry := 0
	if *a.Status&a.Flags.C != 0 {
		carry = 1
	}

	// result of 8 bit addition into 16 bits
	r := uint16(*a.Acc) + uint16(v) + uint16(carry)
	// signed result, 16-bit
	sr := int16(int8(*a.Acc)) + int16(int8(v)) + int16(carry)
	// unsigned result, 8-bit
	ur := uint8(r)
	// result of half add
	hr := *a.Acc&0xf + v&0xf + uint8(carry)

	a.carry(r)
	a.carry4(hr)
	a.overflow(sr)
	a.parity(ur)
	a.zero(ur)
	a.sign(ur)
	*a.Acc = ur
}

// AddBCD adds the value of v to alu.A using binary-coded decimal. If the carry
// is set, increments the result by one. Results are undefined if either
// value is not a valid BCD number.
func (a *ALU) AddBCD(v uint8) {
	carry := 0
	if *a.Status&a.Flags.C != 0 {
		carry = 1
	}

	ba := FromBCD(*a.Acc)
	bv := FromBCD(v)
	r := uint16(ba) + uint16(bv) + uint16(carry)
	bcdr := ToBCD(uint8(r))

	a.carryBCD(r)
	a.zero(bcdr)
	a.sign(bcdr)
	*a.Acc = bcdr
}

func (a *ALU) carry(v uint16) {
	if v > 0xff {
		*a.Status |= a.Flags.C
	} else {
		*a.Status &^= a.Flags.C
	}
}

func (a *ALU) carryBCD(v uint16) {
	if v > 99 {
		*a.Status |= a.Flags.C
	} else {
		*a.Status &^= a.Flags.C
	}
}

func (a *ALU) carry4(v uint8) {
	if v > 0xf {
		*a.Status |= a.Flags.H
	} else {
		*a.Status &^= a.Flags.H
	}
}

func (a *ALU) overflow(v int16) {
	if v < -128 || v > 127 {
		*a.Status |= a.Flags.V
	} else {
		*a.Status &^= a.Flags.V
	}
}

func (a *ALU) parity(v uint8) {
	if bits.OnesCount8(v)%2 == 0 {
		*a.Status |= a.Flags.P
	} else {
		*a.Status &^= a.Flags.P
	}
}

func (a *ALU) zero(v uint8) {
	if v == 0 {
		*a.Status |= a.Flags.Z
	} else {
		*a.Status &^= a.Flags.Z
	}
}

func (a *ALU) sign(v uint8) {
	if v&0x80 != 0 {
		*a.Status |= a.Flags.S
	} else {
		*a.Status &^= a.Flags.S
	}
}
