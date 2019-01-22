package z80

import (
	"testing"
)

func TestString(t *testing.T) {
	cpu := New(nil)
	cpu.A = 0x0a
	cpu.F = 0xff
	cpu.B = 0x0b
	cpu.C = 0x0c
	cpu.D = 0x0d
	cpu.E = 0x0e
	cpu.H = 0xf0
	cpu.L = 0x0f
	cpu.IXH = 0x12
	cpu.IXL = 0x34
	cpu.IYH = 0x56
	cpu.IYL = 0x78
	cpu.SP = 0xabcd
	cpu.I = 0xee
	cpu.R = 0xff

	cpu.A1 = 0xa0
	cpu.F1 = 0x88
	cpu.B1 = 0xb0
	cpu.C1 = 0xc0
	cpu.D1 = 0xd0
	cpu.E1 = 0xe0
	cpu.H1 = 0x0f
	cpu.L1 = 0xf0

	cpu.IFF1 = true
	cpu.IFF2 = true

	have := cpu.String()
	want := "" +
		" pc   af   bc   de   hl   ix   iy   sp   i  r\n" +
		"0000 0aff 0b0c 0d0e f00f 1234 5678 abcd  ee ff iff1\n" +
		"im 0 a088 b0c0 d0e0 0ff0      S Z 5 H 3 V N C  iff2\n"
	if have != want {
		t.Errorf("\n have: \n%v \n want: \n%v", have, want)
	}
}
