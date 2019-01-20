package rcs

import (
	"testing"
)

func testALU() ALU {
	return ALU{
		C: 1 << 0,
		V: 1 << 1,
		P: 1 << 2,
		H: 1 << 3,
		Z: 1 << 4,
		S: 1 << 5,
	}
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
			alu.P,
			"add with carry",
		},
		{
			255, 1, false, 0,
			alu.C | alu.P | alu.H | alu.Z,
			"add results in carry",
		},
		{
			15, 1, false, 16,
			alu.H,
			"add results in half carry",
		},
		{
			127, 10, false, 137,
			alu.V | alu.H | alu.S,
			"add results in overflow",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var sr uint8
			if test.carry {
				sr |= alu.C
			}
			out := alu.Add(&sr, test.a, test.b)
			if out != test.result {
				t.Errorf("\n have: %v \n want: %v", out, test.result)
			}
			if sr != test.status {
				t.Errorf("\n have: %08b \n want: %08b", sr, test.status)
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
			alu.Z | alu.C,
			"add bcd results in carry",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var sr uint8
			if test.carry {
				sr |= alu.C
			}
			out := alu.AddBCD(&sr, test.a, test.b)
			if out != test.result {
				t.Errorf("\n have: 0x%02x \n want: 0x%02x", out, test.result)
			}
			if sr != test.status {
				t.Errorf("\n have: %08b \n want: %08b", sr, test.status)
			}
		})
	}
}

var benchU8 uint8

func BenchmarkALUAdd(b *testing.B) {
	alu := testALU()
	var out uint8
	var sr uint8
	for n := 0; n < b.N; n++ {
		out = alu.Add(&sr, 2, 2)
	}
	benchU8 = out
}
