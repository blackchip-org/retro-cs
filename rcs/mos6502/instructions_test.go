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

// ----------------------------------------------------------------------------
// asl
// ----------------------------------------------------------------------------
func TestAslAccumulator(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0200, 0x0a) // asl a
	c.A = 4
	testRunCPU(t, c)
	want := uint8(8)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestAslSigned(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0200, 0x0a) // asl a
	c.A = 1 << 6
	testRunCPU(t, c)
	want := FlagN | FlagB
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestAslCarry(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0200, 0x0a) // asl a
	c.A = 1 << 7
	testRunCPU(t, c)
	want := FlagC | FlagZ | FlagB
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestAslZeroPage(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x00ab, 4)           // .byte 4
	c.mem.WriteN(0x0200, 0x06, 0xab) // asl $ab
	testRunCPU(t, c)
	want := uint8(8)
	have := c.mem.Read(0x00ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestAslZeroPageX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x00ab, 4)           // .byte 4
	c.mem.WriteN(0x0200, 0x16, 0xa0) // asl $a0
	c.X = 0x0b
	testRunCPU(t, c)
	want := uint8(8)
	have := c.mem.Read(0x00ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestAslAbsolute(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 4)                 // .byte 4
	c.mem.WriteN(0x0200, 0x0e, 0xab, 0x02) // asl $02ab
	testRunCPU(t, c)
	want := uint8(8)
	have := c.mem.Read(0x02ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestAslAbsoluteX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 4)                 // .byte 4
	c.mem.WriteN(0x0200, 0x1e, 0xa0, 0x02) // asl $02a0,X
	c.X = 0x0b
	testRunCPU(t, c)
	want := uint8(8)
	have := c.mem.Read(0x02ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// bit
// ----------------------------------------------------------------------------
var bitTests = []struct {
	name          string
	a             uint8
	fetch         uint8
	expectedFlags uint8
}{
	{"zero", 0x00, 0x00, FlagZ | FlagB},
	{"non-zero", 0x01, 0x01, FlagB},
	{"and-zero", 0x01, 0x02, FlagZ | FlagB},
	{"bit6", 0x00, 1 << 6, FlagV | FlagZ | FlagB},
	{"bit7", 0x00, 1 << 7, FlagN | FlagZ | FlagB},
}

func TestBitAbsolute(t *testing.T) {
	for _, test := range bitTests {
		t.Run(test.name, func(t *testing.T) {
			c := newTestCPU()
			c.mem.Write(0x02ab, test.fetch)        // .byte test.fetch
			c.mem.WriteN(0x0200, 0x2c, 0xab, 0x02) // bit $02ab
			c.A = test.a
			testRunCPU(t, c)
			want := test.expectedFlags
			have := c.SR
			if want != have {
				flagError(t, want, have)
			}
		})
	}
}

func TestBitZeroPage(t *testing.T) {
	for _, test := range bitTests {
		t.Run(test.name, func(t *testing.T) {
			c := newTestCPU()
			c.mem.Write(0xab, test.fetch)    // .byte fetch
			c.mem.WriteN(0x0200, 0x24, 0xab) // bit $ab
			c.A = test.a
			testRunCPU(t, c)
			want := test.expectedFlags
			have := c.SR
			if want != have {
				flagError(t, want, have)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// branches
// ----------------------------------------------------------------------------
var branchTests = []struct {
	name      string
	op        uint8
	flags     uint8
	expectedA uint8
}{
	{"bpl yea", 0x10, 0, 0x02},
	{"bpl nay", 0x10, FlagN, 0x01},
	{"bmi yea", 0x30, FlagN, 0x02},
	{"bmi nay", 0x30, 0, 0x01},
	{"bvc yea", 0x50, 0, 0x02},
	{"bvc nay", 0x50, FlagV, 0x01},
	{"bvs yea", 0x70, FlagV, 0x02},
	{"bvs nay", 0x70, 0, 0x01},
	{"bcc yea", 0x90, 0, 0x02},
	{"bcc nay", 0x90, FlagC, 0x01},
	{"bcs yea", 0xb0, FlagC, 0x02},
	{"bcs nay", 0xb0, 0, 0x01},
	{"bne yea", 0xd0, 0, 0x02},
	{"bne nay", 0xd0, FlagZ, 0x01},
	{"beq yea", 0xf0, FlagZ, 0x02},
	{"beq nay", 0xf0, 0, 0x01},
}

func TestBranches(t *testing.T) {
	for _, test := range branchTests {
		t.Run(test.name, func(t *testing.T) {
			c := newTestCPU()
			c.mem.WriteN(0x0200, test.op, 0x03) // branch to $0205
			c.mem.WriteN(0x0202, 0xa9, 0x01)    // lda #$01
			c.mem.WriteN(0x0204, 0x00)          // brk
			c.mem.WriteN(0x0205, 0xa9, 0x02)    // lda #$02
			c.SR = test.flags
			testRunCPU(t, c)
			want := test.expectedA
			have := c.A
			if want != have {
				t.Errorf("\n want: %02x \n have: %02x \n", want, have)
			}
		})
	}
}

func TestBranchBackwards(t *testing.T) {
	c := newTestCPU()
	c.SetPC(0x0202)
	c.mem.WriteN(0x0200, 0xa9, 0x01) // lda #$01
	c.mem.WriteN(0x0202, 0x00)       // brk
	c.mem.WriteN(0x0203, 0xd0, 0xfb) // bne $0200
	testRunCPU(t, c)
	want := uint8(0x01)
	have := c.A
	if want != have {
		t.Errorf("\n want: %04x \n have: %04x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// lda
// ----------------------------------------------------------------------------
func TestLdaImmediate(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xa9, 0x12) // lda #$12
	testRunCPU(t, c)
	want := uint8(0x12)
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

func TestLdaZero(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xa9, 0x00) // lda #$00
	testRunCPU(t, c)
	want := FlagZ | FlagB 
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestLdaSigned(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xa9, 0xff) // lda #$ff
	testRunCPU(t, c)
	want := FlagN | FlagB 
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestLdaZeroPage(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0034, 0x12)        // .byte $12
	c.mem.WriteN(0x0200, 0xa5, 0x34) // lda $34
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestLdaZeroPageX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0034, 0x12)        // .byte $12
	c.mem.WriteN(0x0200, 0xb5, 0x30) // lda $30,X
	c.X = 0x4
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestLdaAbsolute(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x12)              // .byte $12
	c.mem.WriteN(0x0200, 0xad, 0xab, 0x02) // lda $02ab
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestLdaAbsoluteX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x12)              // .byte $12
	c.mem.WriteN(0x0200, 0xbd, 0xa0, 0x02) // lda $02a0,X
	c.X = 0x0b
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestLdaAbsoluteY(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x12)              // .byte $12
	c.mem.WriteN(0x0200, 0xb9, 0xa0, 0x02) // lda $02a0,Y
	c.Y = 0x0b
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestLdaIndirectX(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteLE(0x4a, 0x02ab)      // .word $02ab
	c.mem.Write(0x02ab, 0x12)        // .byte $12
	c.mem.WriteN(0x0200, 0xa1, 0x40) // lda ($40,X)
	c.X = 0x0a
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestLdaIndirectY(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteLE(0x4a, 0x02a0)      // .word $02a0
	c.mem.Write(0x02ab, 0x12)        // .byte $12
	c.mem.WriteN(0x0200, 0xb1, 0x4a) // lda ($4a),Y
	c.Y = 0x0b
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// ldx
// ----------------------------------------------------------------------------
func TestLdxImmediate(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xa2, 0x12) // ldx #$12
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.X
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagB 
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestLdxZero(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xa2, 0x00) // ldx #$00
	testRunCPU(t, c)
	want := FlagZ | FlagB 
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestLdxSigned(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xa2, 0xff) // ldx #$ff
	testRunCPU(t, c)
	want := FlagN | FlagB 
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestLdxZeroPage(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0034, 0x12)        // .byte $12
	c.mem.WriteN(0x0200, 0xa6, 0x34) // ldx $34
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.X
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestLdxZeroPageY(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0034, 0x12)        // .byte $12
	c.mem.WriteN(0x0200, 0xb6, 0x30) // ldx $30,Y
	c.Y = 0x04
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.X
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestLdxAbsolute(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x12)              // .byte $12
	c.mem.WriteN(0x0200, 0xae, 0xab, 0x02) // ldx $02ab
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.X
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestLdxAbsoluteY(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x12)              // .byte $12
	c.mem.WriteN(0x0200, 0xbe, 0xa0, 0x02) // ldx $02a0,Y
	c.Y = 0xb
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.X
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// ldy
// ----------------------------------------------------------------------------
func TestLdyImmediate(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xa0, 0x12) // ldy #$12
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.Y
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagB 
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestLdyZero(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xa0, 0x00) // ldy #$00
	testRunCPU(t, c)
	want := FlagZ | FlagB 
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestLdySigned(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xa0, 0xff) // ldy #$ff
	testRunCPU(t, c)
	want := FlagN | FlagB 
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestLdyZeroPage(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0034, 0x12)        // .byte $12
	c.mem.WriteN(0x0200, 0xa4, 0x34) // ldy $34
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.Y
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestLdyZeroPageX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0034, 0x12)        // .byte $12
	c.mem.WriteN(0x0200, 0xb4, 0x30) // ldy $30,X
	c.X = 0x4
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.Y
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestLdyAbsolute(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x12)              // .byte $12
	c.mem.WriteN(0x0200, 0xac, 0xab, 0x02) // ldy $02ab
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.Y
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestLdyAbsoluteX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x12)              // .byte $12
	c.mem.WriteN(0x0200, 0xbc, 0xa0, 0x02) // ldy $02a0,X
	c.X = 0xb
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.Y
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// sta
// ----------------------------------------------------------------------------
func TestStaZeroPage(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x85, 0x34) // sta $34
	c.A = 0x12
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.mem.Read(0x34)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestStaZeroPageX(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x95, 0x30) // sta $30,X
	c.A = 0x12
	c.X = 0x04
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.mem.Read(0x34)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestStaAbsolute(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x8d, 0xab, 0x02) // sta $02ab
	c.A = 0x12
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.mem.Read(0x02ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestStaAbsoluteX(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x9d, 0xa0, 0x02) // sta $02a0,X
	c.A = 0x12
	c.X = 0x0b
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.mem.Read(0x02ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestStaAbsoluteY(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x99, 0xa0, 0x02) // sta $02a0,Y
	c.A = 0x12
	c.Y = 0x0b
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.mem.Read(0x02ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestStaIndirectX(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteLE(0x4a, 0x02ab)      // .word $02ab
	c.mem.WriteN(0x0200, 0x81, 0x40) // sta ($40,X)
	c.A = 0x12
	c.X = 0x0a
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.mem.Read(0x02ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestStaIndirectY(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteLE(0x4a, 0x02a0)      // .word $02a0
	c.mem.WriteN(0x0200, 0x91, 0x4a) // sta ($4a),Y
	c.A = 0x12
	c.Y = 0x0b
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.mem.Read(0x02ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// stx
// ----------------------------------------------------------------------------

func TestStxZeroPage(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x86, 0x34) // stx $34
	c.X = 0x12
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.mem.Read(0x34)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestStxZeroPageY(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x96, 0x30) // stx $30,Y
	c.X = 0x12
	c.Y = 0x04
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.mem.Read(0x34)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestStxAbsolute(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x8e, 0xab, 0x02) // stx $02ab
	c.X = 0x12
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.mem.Read(0x02ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// sty
// ----------------------------------------------------------------------------

func TestStyZeroPage(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x84, 0x34) // sty $34
	c.Y = 0x12
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.mem.Read(0x34)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestStyZeroPageX(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x94, 0x30) // sty $30,X
	c.Y = 0x12
	c.X = 0x04
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.mem.Read(0x34)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestStyAbsolute(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x8c, 0xab, 0x02) // sty $02ab
	c.Y = 0x12
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.mem.Read(0x02ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}
