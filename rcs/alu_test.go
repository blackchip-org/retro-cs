package rcs

import (
	"testing"
)

func testALU() *ALU {
	var acc, status uint8
	return NewALU(&acc, &status, FlagMap{
		C: 1 << 0,
		V: 1 << 1,
		P: 1 << 2,
		H: 1 << 3,
		Z: 1 << 4,
		S: 1 << 5,
	})
}

func TestAdd(t *testing.T) {
	alu := testALU()
	var tests = []struct {
		a      uint8
		b      uint8
		carry  bool
		result uint8
		status uint8
		name   string
	}{
		{
			1, 1, false, 2,
			0,
			"add",
		},
		{
			1, 1, true, 3,
			alu.Flags.P,
			"add with carry",
		},
		{
			255, 1, false, 0,
			alu.Flags.C | alu.Flags.P | alu.Flags.H | alu.Flags.Z,
			"add results in carry",
		},
		{
			15, 1, false, 16,
			alu.Flags.H,
			"add results in half carry",
		},
		{
			127, 10, false, 137,
			alu.Flags.V | alu.Flags.H | alu.Flags.S,
			"add results in overflow",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			*alu.Acc = test.a
			*alu.Status = 0
			if test.carry {
				*alu.Status |= alu.Flags.C
			}
			alu.Add(test.b)
			if *alu.Acc != test.result {
				t.Errorf("\n have: %v \n want: %v", *alu.Acc, test.result)
			}
			if *alu.Status != test.status {
				t.Errorf("\n have: %08b \n want: %08b", *alu.Status, test.status)
			}
		})
	}
}

func TestAddBCD(t *testing.T) {
	alu := testALU()
	var tests = []struct {
		a      uint8
		b      uint8
		carry  bool
		result uint8
		status uint8
		name   string
	}{
		{
			0x09, 0x01, false, 0x10,
			0,
			"add bcd",
		},
		{
			0x09, 0x01, true, 0x11,
			0,
			"add bcd with carry",
		},
		{
			0x99, 0x01, false, 0x00,
			alu.Flags.Z | alu.Flags.C,
			"add bcd results in carry",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			*alu.Acc = test.a
			*alu.Status = 0
			if test.carry {
				*alu.Status |= alu.Flags.C
			}
			alu.AddBCD(test.b)
			if *alu.Acc != test.result {
				t.Errorf("\n have: 0x%02x \n want: 0x%02x", *alu.Acc, test.result)
			}
			if *alu.Status != test.status {
				t.Errorf("\n have: %08b \n want: %08b", *alu.Status, test.status)
			}
		})
	}
}

func BenchmarkALUAdd(b *testing.B) {
	alu := testALU()
	for n := 0; n < b.N; n++ {
		alu.Add(2)
	}
}
