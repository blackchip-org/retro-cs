package mos6502

import "testing"

func flagError(t *testing.T, want uint8, have uint8) {
	t.Errorf("\n       nv-bdizc\n want: %08b \n have: %08b \n", want, have)
}

// ----------------------------------------------------------------------------
// adc
// ----------------------------------------------------------------------------
func TestAdcImmediate(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x69, 0x02) // lda #$02
	c.A = 0x08
	testRunCPU(t, c)
	want := uint8(0x0a)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagB
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestAdcWithCarry(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x69, 0x02) // lda #$02
	c.A = 0x08
	c.SR = FlagC
	testRunCPU(t, c)
	want := uint8(0x0b)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagB
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestAdcCarryResult(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x69, 0x02) // lda #$02
	c.A = 0xff
	testRunCPU(t, c)
	want := uint8(0x01)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagC | FlagB
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestAdcZero(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x69, 0x02) // lda #$02
	c.A = 0xfe
	testRunCPU(t, c)
	want := uint8(0x00)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagZ | FlagC | FlagB
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestAdcZeroSigned(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x69, 0x02) // lda #$02
	c.A = 0xf0
	testRunCPU(t, c)
	want := uint8(0xf2)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagN | FlagB
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestAdcOverflowSet(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x69, 0x02) // lda #$02
	c.A = 0x7f
	testRunCPU(t, c)
	want := uint8(0x81)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagV | FlagN | FlagB
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestAdcOverflowClear(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x69, 0xff) // lda #$ff (-1)
	c.SR |= FlagV
	c.A = 0x81
	testRunCPU(t, c)
	want := uint8(0x80)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagC | FlagN | FlagB
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestAdcBcd(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x69, 0x02) // lda $#02
	c.SR |= FlagD
	c.A = 0x08
	testRunCPU(t, c)
	want := uint8(0x10)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestAdcBcdWithCarry(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x69, 0x02) // lda $#02
	c.SR |= FlagD | FlagC
	c.A = 0x08
	testRunCPU(t, c)
	want := uint8(0x11)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestAdcBcdCarryResult(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x69, 0x02) // lda #$02
	c.SR |= FlagD
	c.A = 0x99
	testRunCPU(t, c)
	want := uint8(0x01)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestAdcZeroPage(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0034, 0x08)        // .byte $08
	c.mem.WriteN(0x0200, 0x65, 0x34) // adc $34
	c.A = 0x02
	testRunCPU(t, c)
	want := uint8(0x0a)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestAdcZeroPageX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0034, 0x08)        // .byte $08
	c.mem.WriteN(0x0200, 0x75, 0x30) // adc $30,X
	c.A = 0x02
	c.X = 0x04
	testRunCPU(t, c)
	want := uint8(0x0a)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestAdcAbsolute(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x08)              // .byte $08
	c.mem.WriteN(0x0200, 0x6d, 0xab, 0x02) // adc $02ab
	c.A = 0x02
	testRunCPU(t, c)
	want := uint8(0x0a)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestAdcAbsoluteX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x08)              // .byte $08
	c.mem.WriteN(0x0200, 0x7d, 0xa0, 0x02) // adc $02a0,X
	c.A = 0x02
	c.X = 0x0b
	testRunCPU(t, c)
	want := uint8(0x0a)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestAdcAbsoluteY(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x08)              // .byte $08
	c.mem.WriteN(0x0200, 0x79, 0xa0, 0x02) // adc $02a0,Y
	c.A = 0x02
	c.Y = 0x0b
	testRunCPU(t, c)
	want := uint8(0x0a)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestAdcIndirectX(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteLE(0x4a, 0x02ab)      // .word $02ab
	c.mem.Write(0x02ab, 0x08)        // .byte $08
	c.mem.WriteN(0x0200, 0x61, 0x40) // adc ($40,X)
	c.A = 0x02
	c.X = 0x0a
	testRunCPU(t, c)
	want := uint8(0x0a)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestAdcIndirectY(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteLE(0x4a, 0x02a0)      // .word $02a0
	c.mem.Write(0x02ab, 0x08)        // .byte $08
	c.mem.WriteN(0x0200, 0x71, 0x4a) // adc ($4a),Y
	c.A = 0x02
	c.Y = 0x0b
	testRunCPU(t, c)
	want := uint8(0x0a)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// and
// ----------------------------------------------------------------------------
func TestAndImmediate(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x29, 0x0f) // and #$0f
	c.A = 0xcd
	testRunCPU(t, c)
	want := uint8(0x0d)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagB
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestAndZero(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x29, 0xf0) // and #$f0
	c.A = 0x0f
	testRunCPU(t, c)
	want := FlagZ | FlagB
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestAndSigned(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x29, 0xf0) // and #$f0
	c.A = 0xff
	testRunCPU(t, c)

	want0 := uint8(0xf0)
	have0 := c.A
	if want0 != have0 {
		t.Errorf("\n want: %02x \n have: %02x \n", want0, have0)
	}

	want1 := FlagN | FlagB
	have1 := c.SR
	if want1 != have1 {
		flagError(t, want1, have1)
	}
}

func TestAndZeroPage(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0034, 0x0f)        // .byte $0f
	c.mem.WriteN(0x0200, 0x25, 0x34) // and $34
	c.A = 0xcd
	testRunCPU(t, c)
	want := uint8(0x0d)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestAndZeroPageX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0034, 0x0f)        // .byte $0f
	c.mem.WriteN(0x0200, 0x35, 0x30) // and $30,X
	c.A = 0xcd
	c.X = 0x04
	testRunCPU(t, c)
	want := uint8(0x0d)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestAndAbsolute(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x0f)              // .byte $0f
	c.mem.WriteN(0x0200, 0x2d, 0xab, 0x02) // and $02ab
	c.A = 0xcd
	testRunCPU(t, c)
	want := uint8(0x0d)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestAndAbsoluteX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x0f)              // .byte $0f
	c.mem.WriteN(0x0200, 0x3d, 0xa0, 0x02) // and $02a0,X
	c.A = 0xcd
	c.X = 0x0b
	testRunCPU(t, c)
	want := uint8(0x0d)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestAndAbsoluteY(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x0f)              // .byte $0f
	c.mem.WriteN(0x0200, 0x39, 0xa0, 0x02) // and $02a0,Y
	c.A = 0xcd
	c.Y = 0x0b
	testRunCPU(t, c)
	want := uint8(0x0d)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestAndIndirectX(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteLE(0x4a, 0x02ab)      // .word $02ab
	c.mem.Write(0x02ab, 0x0f)        // .byte $0f
	c.mem.WriteN(0x0200, 0x21, 0x40) // and ($40,X)
	c.A = 0xcd
	c.X = 0x0a
	testRunCPU(t, c)
	want := uint8(0x0d)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestAndIndirectY(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteLE(0x4a, 0x02a0)      // .word $02a0
	c.mem.Write(0x02ab, 0x0f)        // .byte $0f
	c.mem.WriteN(0x0200, 0x31, 0x4a) // and ($4a),Y
	c.A = 0xcd
	c.Y = 0x0b
	testRunCPU(t, c)
	want := uint8(0x0d)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}
