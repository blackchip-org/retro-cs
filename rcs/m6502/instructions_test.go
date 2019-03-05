package m6502

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
	want = Flag5
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
	want = Flag5
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
	want = FlagC | Flag5
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
	want = FlagZ | FlagC | Flag5
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
	want = FlagN | Flag5
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
	want = FlagV | FlagN | Flag5
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
	want = FlagC | FlagN | Flag5
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
	want = Flag5
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
	want := FlagZ | Flag5
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

	want1 := FlagN | Flag5
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
	want := FlagN | Flag5
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
	want := FlagC | FlagZ | Flag5
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
	{"zero", 0x00, 0x00, FlagZ | Flag5},
	{"non-zero", 0x01, 0x01, Flag5},
	{"and-zero", 0x01, 0x02, FlagZ | Flag5},
	{"bit6", 0x00, 1 << 6, FlagV | FlagZ | Flag5},
	{"bit7", 0x00, 1 << 7, FlagN | FlagZ | Flag5},
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
// cmp
// ----------------------------------------------------------------------------
func TestCmpImmediateEqual(t *testing.T) {
	c := newTestCPU()
	c.A = 0x12
	c.mem.WriteN(0x0200, 0xc9, 0x12) // cmp #$12
	testRunCPU(t, c)
	want := FlagZ | FlagC | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestCmpImmediateLessThan(t *testing.T) {
	c := newTestCPU()
	c.A = 0x02
	c.mem.WriteN(0x0200, 0xc9, 0x12) // cmp #$12
	testRunCPU(t, c)
	want := FlagN | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestCmpImmediateGreaterThan(t *testing.T) {
	c := newTestCPU()
	c.A = 0x22
	c.mem.WriteN(0x0200, 0xc9, 0x12) // cmp #$12
	testRunCPU(t, c)
	want := FlagC | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestCmpZeroPage(t *testing.T) {
	c := newTestCPU()
	c.A = 0x12
	c.mem.Write(0x0034, 0x12)        // .byte $12
	c.mem.WriteN(0x0200, 0xc5, 0x34) // cmp $34
	testRunCPU(t, c)
	want := FlagZ | FlagC | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestCmpZeroPageX(t *testing.T) {
	c := newTestCPU()
	c.A = 0x12
	c.mem.Write(0x0034, 0x12)        // .byte $12
	c.mem.WriteN(0x0200, 0xd5, 0x30) // cmp $30,X
	c.X = 0x04
	testRunCPU(t, c)
	want := FlagZ | FlagC | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestCmpAbsolute(t *testing.T) {
	c := newTestCPU()
	c.A = 0x12
	c.mem.Write(0x02ab, 0x12)              // .byte $12
	c.mem.WriteN(0x0200, 0xcd, 0xab, 0x02) // cmp $02ab
	testRunCPU(t, c)
	want := FlagZ | FlagC | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestCmpAbsoluteX(t *testing.T) {
	c := newTestCPU()
	c.A = 0x12
	c.mem.Write(0x02ab, 0x12)              // .byte $12
	c.mem.WriteN(0x0200, 0xdd, 0xa0, 0x02) // cmp $02a0,X
	c.X = 0x0b
	testRunCPU(t, c)
	want := FlagZ | FlagC | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestCmpAbsoluteY(t *testing.T) {
	c := newTestCPU()
	c.A = 0x12
	c.mem.Write(0x02ab, 0x12)              // .byte $12
	c.mem.WriteN(0x0200, 0xd9, 0xa0, 0x02) // cmp $02a0,Y
	c.Y = 0x0b
	testRunCPU(t, c)
	want := FlagZ | FlagC | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestCmpIndirectX(t *testing.T) {
	c := newTestCPU()
	c.A = 0x12
	c.mem.WriteLE(0x4a, 0x02ab)      // .word $02ab
	c.mem.Write(0x02ab, 0x12)        // .byte $12
	c.mem.WriteN(0x0200, 0xc1, 0x40) // cmp ($40,X)
	c.X = 0x0a
	testRunCPU(t, c)
	want := FlagZ | FlagC | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestCmpIndirectY(t *testing.T) {
	c := newTestCPU()
	c.A = 0x12
	c.mem.WriteLE(0x4a, 0x02a0)      // .word $02a0
	c.mem.Write(0x02ab, 0x12)        // .byte $12
	c.mem.WriteN(0x0200, 0xd1, 0x4a) // cmp ($4a),Y
	c.Y = 0x0b
	testRunCPU(t, c)
	want := FlagZ | FlagC | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

// ----------------------------------------------------------------------------
// cpx
// ----------------------------------------------------------------------------
func TestCpxImmediateEqual(t *testing.T) {
	c := newTestCPU()
	c.X = 0x12
	c.mem.WriteN(0x0200, 0xe0, 0x12) // cpx #$12
	testRunCPU(t, c)
	want := FlagZ | FlagC | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestCpxImmediateLessThan(t *testing.T) {
	c := newTestCPU()
	c.X = 0x02
	c.mem.WriteN(0x0200, 0xe0, 0x12) // cpx #$12
	testRunCPU(t, c)
	want := FlagN | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestCpxImmediateGreaterThan(t *testing.T) {
	c := newTestCPU()
	c.X = 0x22
	c.mem.WriteN(0x0200, 0xe0, 0x12) // cpx #$12
	testRunCPU(t, c)
	want := FlagC | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestCpxZeroPage(t *testing.T) {
	c := newTestCPU()
	c.X = 0x12
	c.mem.Write(0x0034, 0x12)        // .byte $12
	c.mem.WriteN(0x0200, 0xe4, 0x34) // cpx $34
	testRunCPU(t, c)
	want := FlagZ | FlagC | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestCpxAbsolute(t *testing.T) {
	c := newTestCPU()
	c.X = 0x12
	c.mem.Write(0x02ab, 0x12)              // .byte $12
	c.mem.WriteN(0x0200, 0xec, 0xab, 0x02) // cpx $02ab
	testRunCPU(t, c)
	want := FlagZ | FlagC | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

// ----------------------------------------------------------------------------
// cpy
// ----------------------------------------------------------------------------
func TestCpyImmediateEqual(t *testing.T) {
	c := newTestCPU()
	c.Y = 0x12
	c.mem.WriteN(0x0200, 0xc0, 0x12) // cpy #$12
	testRunCPU(t, c)
	want := FlagZ | FlagC | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestCpyImmediateLessThan(t *testing.T) {
	c := newTestCPU()
	c.Y = 0x02
	c.mem.WriteN(0x0200, 0xc0, 0x12) // cpy #$12
	testRunCPU(t, c)
	want := FlagN | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestCpyImmediateGreaterThan(t *testing.T) {
	c := newTestCPU()
	c.Y = 0x22
	c.mem.WriteN(0x0200, 0xc0, 0x12) // cpy #$12
	testRunCPU(t, c)
	want := FlagC | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestCpyZeroPage(t *testing.T) {
	c := newTestCPU()
	c.Y = 0x12
	c.mem.Write(0x0034, 0x12)        // .byte $12
	c.mem.WriteN(0x0200, 0xc4, 0x34) // cpy $34
	testRunCPU(t, c)
	want := FlagZ | FlagC | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestCpyAbsolute(t *testing.T) {
	c := newTestCPU()
	c.Y = 0x12
	c.mem.Write(0x02ab, 0x12)              // .byte $12
	c.mem.WriteN(0x0200, 0xcc, 0xab, 0x02) // cpy $02ab
	testRunCPU(t, c)
	want := FlagZ | FlagC | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

// ----------------------------------------------------------------------------
// dec
// ----------------------------------------------------------------------------
func TestDecZeroPage(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0xab, 0x12)          // .byte $12
	c.mem.WriteN(0x0200, 0xc6, 0xab) // dec $ab
	testRunCPU(t, c)
	want := uint8(0x11)
	have := c.mem.Read(0xab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestDecZeroPageZero(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0xab, 0x01)          // .byte $01
	c.mem.WriteN(0x0200, 0xc6, 0xab) // dec $ab
	testRunCPU(t, c)
	want := uint8(0x00)
	have := c.mem.Read(0xab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagZ | Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestDecZeroPageSigned(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0xab, 0x00)          // .byte $00
	c.mem.WriteN(0x0200, 0xc6, 0xab) // dec $ab
	testRunCPU(t, c)
	want := uint8(0xff)
	have := c.mem.Read(0xab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagN | Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestDecZeroPageX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0xab, 0x12)          // .byte $12
	c.mem.WriteN(0x0200, 0xd6, 0xa0) // dec $a0,X
	c.X = 0x0b
	testRunCPU(t, c)
	want := uint8(0x11)
	have := c.mem.Read(0xab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestDecAbsolute(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x12)              // .byte $12
	c.mem.WriteN(0x0200, 0xce, 0xab, 0x02) // dec $02ab
	testRunCPU(t, c)
	want := uint8(0x11)
	have := c.mem.Read(0x02ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestDecAbsoluteX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x12)              // .byte $12
	c.mem.WriteN(0x0200, 0xde, 0xa0, 0x02) // dec $02a0,X
	c.X = 0x0b
	testRunCPU(t, c)
	want := uint8(0x11)
	have := c.mem.Read(0x02ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

// ----------------------------------------------------------------------------
// dex
// ----------------------------------------------------------------------------
func TestDex(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xca)
	c.X = 0x12
	testRunCPU(t, c)
	want := uint8(0x11)
	have := c.X
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestDexZero(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xca)
	c.X = 0x01
	testRunCPU(t, c)
	want := FlagZ | Flag5
	have := c.SR
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestDexSigned(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xca)
	c.X = 0x00
	testRunCPU(t, c)
	want := FlagN | Flag5
	have := c.SR
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// dey
// ----------------------------------------------------------------------------
func TestDey(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x88)
	c.Y = 0x12
	testRunCPU(t, c)
	want := uint8(0x11)
	have := c.Y
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestDeyZero(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x88)
	c.Y = 0x01
	testRunCPU(t, c)
	want := FlagZ | Flag5
	have := c.SR
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestDeySigned(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x88)
	c.Y = 0x00
	testRunCPU(t, c)
	want := FlagN | Flag5
	have := c.SR
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// eor
// ----------------------------------------------------------------------------
func TestEorImmdediate(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x200, 0x49, 0x01) // eor #$01
	c.A = 0x02
	testRunCPU(t, c)
	want := uint8(0x03)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestEorImmediateZero(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x200, 0x49, 0x01) // eor #$01
	c.A = 0x01
	testRunCPU(t, c)
	want := uint8(0x00)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagZ | Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestEorImmediateSigned(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x200, 0x49, 0x0f) // eor #$0f
	c.A = 0xf0
	testRunCPU(t, c)
	want := uint8(0xff)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagN | Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestEorZeroPage(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0034, 0x01)        // .byte $01
	c.mem.WriteN(0x0200, 0x45, 0x34) // eor $34
	c.A = 0x02
	testRunCPU(t, c)
	want := uint8(0x03)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestEorZeroPageX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0034, 0x01)        // .byte $01
	c.mem.WriteN(0x0200, 0x55, 0x30) // eor $30,X
	c.A = 0x02
	c.X = 0x04
	testRunCPU(t, c)
	want := uint8(0x03)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestEorAbsolute(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x01)              // .byte $01
	c.mem.WriteN(0x0200, 0x4d, 0xab, 0x02) // eor $02ab
	c.A = 0x02
	testRunCPU(t, c)
	want := uint8(0x03)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestEorAbsoluteX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x01)              // .byte $01
	c.mem.WriteN(0x0200, 0x5d, 0xa0, 0x02) // eor $02a0,X
	c.A = 0x02
	c.X = 0x0b
	testRunCPU(t, c)
	want := uint8(0x03)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestEorAbsoluteY(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x01)              // .byte $01
	c.mem.WriteN(0x0200, 0x59, 0xa0, 0x02) // eor $02a0,Y
	c.A = 0x02
	c.Y = 0x0b
	testRunCPU(t, c)
	want := uint8(0x03)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestEorIndirectX(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteLE(0x4a, 0x02ab)      // .word $02ab
	c.mem.Write(0x02ab, 0x01)        // .byte $01
	c.mem.WriteN(0x0200, 0x41, 0x40) // eor ($40,X)
	c.A = 0x02
	c.X = 0x0a
	testRunCPU(t, c)
	want := uint8(0x03)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestEorIndirectY(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteLE(0x4a, 0x02a0)      // .word $02a0
	c.mem.Write(0x02ab, 0x01)        // .byte $01
	c.mem.WriteN(0x0200, 0x51, 0x4a) // lda ($4a),Y
	c.A = 0x02
	c.Y = 0x0b
	testRunCPU(t, c)
	want := uint8(0x03)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// flags
// ----------------------------------------------------------------------------
func TestClc(t *testing.T) {
	c := newTestCPU()
	c.SP |= FlagC
	c.mem.WriteN(0x0200, 0x18) // clc
	testRunCPU(t, c)
	want := Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestSec(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x38) // sec
	testRunCPU(t, c)
	want := FlagC | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestCli(t *testing.T) {
	c := newTestCPU()
	c.SR |= FlagI
	c.mem.WriteN(0x0200, 0x58) // cli
	testRunCPU(t, c)
	want := Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestSei(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x78) // sei
	testRunCPU(t, c)
	want := FlagI | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestClv(t *testing.T) {
	c := newTestCPU()
	c.SR |= FlagV
	c.mem.WriteN(0x0200, 0xb8) // clv
	testRunCPU(t, c)
	want := Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestCld(t *testing.T) {
	c := newTestCPU()
	c.SR |= FlagD
	c.mem.WriteN(0x0200, 0xd8) // cld
	testRunCPU(t, c)
	want := Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestSed(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xf8) // sed
	testRunCPU(t, c)
	want := FlagD | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

// ----------------------------------------------------------------------------
// inc
// ----------------------------------------------------------------------------
func TestIncZeroPage(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0xab, 0x12)          // .byte $12
	c.mem.WriteN(0x0200, 0xe6, 0xab) // inc $ab
	testRunCPU(t, c)
	want := uint8(0x13)
	have := c.mem.Read(0xab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestIncZeroPageZero(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0xab, 0xff)          // .byte $ff
	c.mem.WriteN(0x0200, 0xe6, 0xab) // inc $ab
	testRunCPU(t, c)
	want := uint8(0x00)
	have := c.mem.Read(0xab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagZ | Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestIncZeroPageSigned(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0xab, 0x7f)          // .byte $7f
	c.mem.WriteN(0x0200, 0xe6, 0xab) // inc $ab
	testRunCPU(t, c)
	want := uint8(0x80)
	have := c.mem.Read(0xab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagN | Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestIncZeroPageX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0xab, 0x12)          // .byte $12
	c.mem.WriteN(0x0200, 0xf6, 0xa0) // inc $a0,X
	c.X = 0x0b
	testRunCPU(t, c)
	want := uint8(0x13)
	have := c.mem.Read(0xab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestIncAbsolute(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x12)              // .byte $12
	c.mem.WriteN(0x0200, 0xee, 0xab, 0x02) // inc $02ab
	testRunCPU(t, c)
	want := uint8(0x13)
	have := c.mem.Read(0x02ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestIncAbsoluteX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x12)              // .byte $12
	c.mem.WriteN(0x0200, 0xfe, 0xa0, 0x02) // inc $02a0,X
	c.X = 0x0b
	testRunCPU(t, c)
	want := uint8(0x13)
	have := c.mem.Read(0x02ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

// ----------------------------------------------------------------------------
// inx
// ----------------------------------------------------------------------------
func TestInx(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xe8)
	c.X = 0x12
	testRunCPU(t, c)
	want := uint8(0x13)
	have := c.X
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestInxZero(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xe8)
	c.X = 0xff
	testRunCPU(t, c)
	want := FlagZ | Flag5
	have := c.SR
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestInxSigned(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xe8)
	c.X = 0x7f
	testRunCPU(t, c)
	want := FlagN | Flag5
	have := c.SR
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// iny
// ----------------------------------------------------------------------------
func TestIny(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xc8)
	c.Y = 0x12
	testRunCPU(t, c)
	want := uint8(0x13)
	have := c.Y
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestInyZero(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xc8)
	c.Y = 0xff
	testRunCPU(t, c)
	want := FlagZ | Flag5
	have := c.SR
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestInySigned(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xc8)
	c.Y = 0x7f
	testRunCPU(t, c)
	want := FlagN | Flag5
	have := c.SR
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// jmp
// ----------------------------------------------------------------------------
func TestJmpAbsolute(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x4c, 0x30, 0x02) // jmp $0230
	c.Next()
	want := uint16(0x022f)
	have := c.pc
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestJmpIndirect(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteLE(0x0230, 0x0240)
	c.mem.WriteN(0x0200, 0x6c, 0x30, 0x02) // jmp ($0230)
	c.Next()
	want := uint16(0x023f)
	have := c.pc
	if want != have {
		t.Errorf("\n want: %04x \n have: %04x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// jsr
// ----------------------------------------------------------------------------
func TestJsr(t *testing.T) {
	c := newTestCPU()
	c.pc = 0x02ff
	c.mem.WriteN(0x0300, 0x20, 0x30, 0x04) // jsr $0430
	c.Next()
	want0 := uint16(0x042f)
	have0 := c.pc
	if want0 != have0 {
		t.Errorf("\n want: %04x \n have: %04x \n", want0, have0)
	}
	want1 := 0x0302
	have1 := c.mem.ReadLE(addrStack + 0x100 - 2)
	if want1 != have1 {
		t.Errorf("\n want: %04x \n have: %04x \n", want1, have1)
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
	want = Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestLdaZero(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xa9, 0x00) // lda #$00
	testRunCPU(t, c)
	want := FlagZ | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestLdaSigned(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xa9, 0xff) // lda #$ff
	testRunCPU(t, c)
	want := FlagN | Flag5
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
	want = Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestLdxZero(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xa2, 0x00) // ldx #$00
	testRunCPU(t, c)
	want := FlagZ | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestLdxSigned(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xa2, 0xff) // ldx #$ff
	testRunCPU(t, c)
	want := FlagN | Flag5
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
	want = Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestLdyZero(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xa0, 0x00) // ldy #$00
	testRunCPU(t, c)
	want := FlagZ | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestLdySigned(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xa0, 0xff) // ldy #$ff
	testRunCPU(t, c)
	want := FlagN | Flag5
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
// lsr
// ----------------------------------------------------------------------------
func TestLsrAccumulator(t *testing.T) {
	c := newTestCPU()
	c.A = 0x04
	c.mem.WriteN(0x0200, 0x04a) // lsr a
	testRunCPU(t, c)
	want := uint8(0x02)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestLsrShiftOut(t *testing.T) {
	c := newTestCPU()
	c.A = 0x01
	c.mem.WriteN(0x0200, 0x04a) // lsr a
	testRunCPU(t, c)
	want := uint8(0x00)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagZ | FlagC | Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestLsrZeroPage(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x00ab, 4)           // .byte 4
	c.mem.WriteN(0x0200, 0x46, 0xab) // lsr $ab
	testRunCPU(t, c)
	want := uint8(2)
	have := c.mem.Read(0x00ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestLsrZeroPageX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x00ab, 4)           // .byte 4
	c.mem.WriteN(0x0200, 0x56, 0xa0) // lsr $a0
	c.X = 0x0b
	testRunCPU(t, c)
	want := uint8(2)
	have := c.mem.Read(0x00ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestLsrAbsolute(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 4)                 // .byte 4
	c.mem.WriteN(0x0200, 0x4e, 0xab, 0x02) // lsr $02ab
	testRunCPU(t, c)
	want := uint8(2)
	have := c.mem.Read(0x02ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestLsrAbsoluteX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 4)                 // .byte 4
	c.mem.WriteN(0x0200, 0x5e, 0xa0, 0x02) // lsr $02a0,X
	c.X = 0x0b
	testRunCPU(t, c)
	want := uint8(2)
	have := c.mem.Read(0x02ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// nop
// ----------------------------------------------------------------------------
func TestNop(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0200, 0xea)
	c.Next()
	want := uint16(0x0200)
	have := c.pc
	if want != have {
		t.Errorf("\n want: %04x \n have: %04x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// ora
// ----------------------------------------------------------------------------
func TestOraImmediate(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x09, 0x01) // ora #$01
	c.A = 0x02
	testRunCPU(t, c)
	want := uint8(0x03)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestOraZero(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x09, 0x00) // ora #$00
	c.A = 0x00
	testRunCPU(t, c)
	want := FlagZ | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestOraSigned(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xa9, 0xf0) // and #$f0
	c.A = 0x0f
	testRunCPU(t, c)
	want := FlagN | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestOraZeroPage(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0034, 0x01)        // .byte $01
	c.mem.WriteN(0x0200, 0x05, 0x34) // ora $34
	c.A = 0x02
	testRunCPU(t, c)
	want := uint8(0x03)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestOraZeroPageX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0034, 0x01)        // .byte $01
	c.mem.WriteN(0x0200, 0x15, 0x30) // ora $30,X
	c.A = 0x02
	c.X = 0x04
	testRunCPU(t, c)
	want := uint8(0x03)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestOraAbsolute(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x01)              // .byte $01
	c.mem.WriteN(0x0200, 0x0d, 0xab, 0x02) // ora $02ab
	c.A = 0x02
	testRunCPU(t, c)
	want := uint8(0x03)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestOraAbsoluteX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x01)              // .byte $01
	c.mem.WriteN(0x0200, 0x1d, 0xa0, 0x02) // ora $02a0,X
	c.A = 0x02
	c.X = 0x0b
	testRunCPU(t, c)
	want := uint8(0x03)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestOraAbsoluteY(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x01)              // .byte $01
	c.mem.WriteN(0x0200, 0x19, 0xa0, 0x02) // ora $02a0,Y
	c.A = 0x02
	c.Y = 0x0b
	testRunCPU(t, c)
	want := uint8(0x03)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestOraIndirectX(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteLE(0x4a, 0x02ab)      // .word $02ab
	c.mem.Write(0x02ab, 0x01)        // .byte $01
	c.mem.WriteN(0x0200, 0x01, 0x40) // ora ($40,X)
	c.A = 0x02
	c.X = 0x0a
	testRunCPU(t, c)
	want := uint8(0x03)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestOraIndirectY(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteLE(0x4a, 0x02a0)      // .word $02a0
	c.mem.Write(0x02ab, 0x01)        // .byte $01
	c.mem.WriteN(0x0200, 0x11, 0x4a) // ora ($4a),Y
	c.A = 0x02
	c.Y = 0x0b
	testRunCPU(t, c)
	want := uint8(0x03)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// pha
// ----------------------------------------------------------------------------
func TestPha(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x48)
	c.A = 0x12
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.mem.Read(addrStack + 0xff)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// php
// ----------------------------------------------------------------------------
func TestPhp(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x08)
	c.SR |= FlagC
	testRunCPU(t, c)
	want := FlagC | Flag5 | FlagB
	have := c.mem.Read(addrStack + 0xff)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// pla
// ----------------------------------------------------------------------------
func TestPla(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x68)
	c.SP = 0xfe
	c.mem.Write(addrStack+0xff, 0x12)
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestPlaZero(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x68)
	c.SP = 0xfe
	c.mem.Write(addrStack+0xff, 0x00)
	testRunCPU(t, c)
	want := FlagZ | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestPlaSigned(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x68)
	c.SP = 0xfe
	c.mem.Write(addrStack+0xff, 0xff)
	testRunCPU(t, c)
	want := FlagN | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

// ----------------------------------------------------------------------------
// plp
// ----------------------------------------------------------------------------
func TestPlp(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x28)
	c.SP = 0xfe
	c.mem.Write(addrStack+0xff, FlagC|FlagN)
	testRunCPU(t, c)
	want := FlagC | FlagN | Flag5
	have := c.SR
	if want != have {
		flagError(t, want, have)
	}
}

// ----------------------------------------------------------------------------
// rol
// ----------------------------------------------------------------------------
func TestRolAccumulator(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x2a) // rol a
	c.A = 4
	testRunCPU(t, c)
	want := uint8(0x08)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestRolRotateOut(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x2a) // rol a
	c.A = 1 << 7
	testRunCPU(t, c)
	want := uint8(0)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagZ | FlagC | Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestRolRotateIn(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x2a) // rol a
	c.SR |= FlagC
	c.A = 0
	testRunCPU(t, c)
	want := uint8(0x01)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestRolSigned(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x2a) // rol a
	c.A = 1 << 6
	testRunCPU(t, c)
	want := uint8(1 << 7)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagN | Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestRolZeroPage(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x00ab, 0x04)        // .byte 0x04
	c.mem.WriteN(0x0200, 0x26, 0xab) // rol $ab
	testRunCPU(t, c)
	want := uint8(0x08)
	have := c.mem.Read(0x00ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}
func TestRolZeroPageX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x00ab, 0x04)        // .byte 0x04
	c.mem.WriteN(0x0200, 0x36, 0xa0) // rol $a0
	c.X = 0x0b
	testRunCPU(t, c)
	want := uint8(0x08)
	have := c.mem.Read(0x00ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestRolAbsolute(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x04)              // .byte 0x04
	c.mem.WriteN(0x0200, 0x2e, 0xab, 0x02) // rol $02ab
	testRunCPU(t, c)
	want := uint8(0x08)
	have := c.mem.Read(0x02ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestRolAbsoluteX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x04)              // .byte 0x04
	c.mem.WriteN(0x0200, 0x3e, 0xa0, 0x02) // rol $02a0,X
	c.X = 0x0b
	testRunCPU(t, c)
	want := uint8(0x08)
	have := c.mem.Read(0x02ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// ror
// ----------------------------------------------------------------------------
func TestRorAccumulator(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x6a) // ror a
	c.A = 4
	testRunCPU(t, c)
	want := uint8(0x02)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}
func TestRorRotateOut(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x6a) // ror a
	c.A = 1
	testRunCPU(t, c)
	want := uint8(0)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagZ | FlagC | Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestRorRotateIn(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0x6a) // ror a
	c.SR |= FlagC
	c.A = 0
	testRunCPU(t, c)
	want := uint8(0x80)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagN | Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestRorZeroPage(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x00ab, 0x04)        // .byte 0x04
	c.mem.WriteN(0x0200, 0x66, 0xab) // ror $ab
	testRunCPU(t, c)
	want := uint8(0x02)
	have := c.mem.Read(0x00ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestRorZeroPageX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x00ab, 0x04)        // .byte 0x04
	c.mem.WriteN(0x0200, 0x76, 0xa0) // ror $a0
	c.X = 0x0b
	testRunCPU(t, c)
	want := uint8(0x02)
	have := c.mem.Read(0x00ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestRorAbsolute(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x04)              // .byte 0x04
	c.mem.WriteN(0x0200, 0x6e, 0xab, 0x02) // ror $02ab
	testRunCPU(t, c)
	want := uint8(0x02)
	have := c.mem.Read(0x02ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestRorAbsoluteX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x04)              // .byte 0x04
	c.mem.WriteN(0x0200, 0x7e, 0xa0, 0x02) // ror $02a0,X
	c.X = 0x0b
	testRunCPU(t, c)
	want := uint8(0x02)
	have := c.mem.Read(0x02ab)
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// rti
// ----------------------------------------------------------------------------
func TestRti(t *testing.T) {
	c := newTestCPU()
	c.push2(0x1234)
	c.push(FlagC)
	c.mem.Write(0x0200, 0x40) //rti
	testRunCPU(t, c)
	wantSR := FlagC | Flag5
	haveSR := c.SR
	if wantSR != haveSR {
		t.Errorf("\n want: %02x \n have: %02x \n", wantSR, haveSR)
	}
	wantPC := uint16(0x1234)
	havePC := c.pc
	if wantPC != havePC {
		t.Errorf("\n want: %04x \n have: %04x \n", wantPC, havePC)
	}
}

// ----------------------------------------------------------------------------
// rts
// ----------------------------------------------------------------------------
func TestRts(t *testing.T) {
	c := newTestCPU()
	c.push2(0x1234)
	c.mem.Write(0x0200, 0x60) // rts
	c.Next()
	want := uint16(0x1234)
	have := c.pc
	if want != have {
		t.Errorf("\n want: %04x \n have: %04x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// sbc
// ----------------------------------------------------------------------------
func TestSbcImmediate(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xe9, 0x02) // sbc #$02
	c.A = 0x08
	c.SR |= FlagC
	testRunCPU(t, c)
	want := uint8(0x06)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagC | Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestAdcWithBorrow(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xe9, 0x02) // sbc #$02
	c.A = 0x08
	testRunCPU(t, c)
	want := uint8(0x05)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagC | Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestSbcCarryResult(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xe9, 0x02) // sbc #$02
	c.A = 0x01
	c.SR |= FlagC
	testRunCPU(t, c)
	want := uint8(0xff)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagN | Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestSbcZero(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xe9, 0x02) // sbc #$02
	c.A = 0x02
	c.SR |= FlagC
	testRunCPU(t, c)
	want := uint8(0x00)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagZ | FlagC | Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestSbcOverflowSet(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xe9, 0x02) // sbc #$02
	c.A = 0x81
	c.SR |= FlagC
	testRunCPU(t, c)
	want := uint8(0x7f)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagV | Flag5 | FlagC
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestSbcOverflowClear(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xe9, 0xff) // sbc #$ff (-1)
	c.A = 0x82
	c.SR |= FlagC | FlagV
	testRunCPU(t, c)
	want := uint8(0x83)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
	want = FlagN | Flag5
	have = c.SR
	if want != have {
		flagError(t, want, have)
	}
}

func TestSbcBcd(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xe9, 0x02) // sbc $#02
	c.SR |= FlagD | FlagC
	c.A = 0x11
	testRunCPU(t, c)
	want := uint8(0x09)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestSbcBcdWithBorrow(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xe9, 0x02) // lda $#02
	c.SR |= FlagD
	c.A = 0x11
	testRunCPU(t, c)
	want := uint8(0x08)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestSbcBcdBorrowResult(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteN(0x0200, 0xe9, 0x02) // sbc #$02
	c.SR |= FlagD | FlagC
	c.A = 0x01
	testRunCPU(t, c)
	want := uint8(0x99)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestSbcZeroPage(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0034, 0x08)        // .byte $08
	c.mem.WriteN(0x0200, 0xe5, 0x34) // sbc $34
	c.A = 0x0a
	c.SR |= FlagC
	testRunCPU(t, c)
	want := uint8(0x02)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestSbcZeroPageX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0034, 0x08)        // .byte $08
	c.mem.WriteN(0x0200, 0xf5, 0x30) // sbc $30,X
	c.A = 0x0a
	c.X = 0x04
	c.SR |= FlagC
	testRunCPU(t, c)
	want := uint8(0x02)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestSbcAbsolute(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x08)              // .byte $08
	c.mem.WriteN(0x0200, 0xed, 0xab, 0x02) // sbc $02ab
	c.A = 0x0a
	c.SR |= FlagC
	testRunCPU(t, c)
	want := uint8(0x02)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestSbcAbsoluteX(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x08)              // .byte $08
	c.mem.WriteN(0x0200, 0xfd, 0xa0, 0x02) // sbc $02a0,X
	c.A = 0x0a
	c.X = 0x0b
	c.SR |= FlagC
	testRunCPU(t, c)
	want := uint8(0x02)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestSbcAbsoluteY(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x02ab, 0x08)              // .byte $08
	c.mem.WriteN(0x0200, 0xf9, 0xa0, 0x02) // sbc $02a0,Y
	c.A = 0x0a
	c.Y = 0x0b
	c.SR |= FlagC
	testRunCPU(t, c)
	want := uint8(0x02)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestSbcIndirectX(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteLE(0x4a, 0x02ab)      // .word $02ab
	c.mem.Write(0x02ab, 0x08)        // .byte $08
	c.mem.WriteN(0x0200, 0xe1, 0x40) // sbc ($40,X)
	c.A = 0x0a
	c.X = 0x0a
	c.SR |= FlagC
	testRunCPU(t, c)
	want := uint8(0x02)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestSbcIndirectY(t *testing.T) {
	c := newTestCPU()
	c.mem.WriteLE(0x4a, 0x02a0)      // .word $02a0
	c.mem.Write(0x02ab, 0x08)        // .byte $08
	c.mem.WriteN(0x0200, 0xf1, 0x4a) // sbc ($4a),Y
	c.A = 0x0a
	c.Y = 0x0b
	c.SR |= FlagC
	testRunCPU(t, c)
	want := uint8(0x02)
	have := c.A
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

// ----------------------------------------------------------------------------
// tax
// ----------------------------------------------------------------------------
func TestTax(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0200, 0xaa)
	c.A = 0x12
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.X
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestTaxZero(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0200, 0xaa)
	c.A = 0x00
	testRunCPU(t, c)
	want := FlagZ | Flag5
	have := c.SR
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestTaxSigned(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0200, 0xaa)
	c.A = 0xff
	testRunCPU(t, c)
	want := FlagN | Flag5
	have := c.SR
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// tay
// ----------------------------------------------------------------------------
func TestTay(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0200, 0xa8)
	c.A = 0x12
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.Y
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestTayZero(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0200, 0xa8)
	c.A = 0x00
	testRunCPU(t, c)
	want := FlagZ | Flag5
	have := c.SR
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestTaySigned(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0200, 0xa8)
	c.A = 0xff
	testRunCPU(t, c)
	want := FlagN | Flag5
	have := c.SR
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// tsx
// ----------------------------------------------------------------------------
func TestTsx(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0200, 0xba)
	c.SP = 0x12
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.X
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestTsxZero(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0200, 0xba)
	c.SP = 0x00
	testRunCPU(t, c)
	want := FlagZ | Flag5
	have := c.SR
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestTsxSigned(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0200, 0xba)
	c.SP = 0xff
	testRunCPU(t, c)
	want := FlagN | Flag5
	have := c.SR
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// txa
// ----------------------------------------------------------------------------
func TestTxa(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0200, 0x8a)
	c.X = 0x12
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestTxaZero(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0200, 0x8a)
	c.X = 0x00
	testRunCPU(t, c)
	want := FlagZ | Flag5
	have := c.SR
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestTxaSigned(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0200, 0x8a)
	c.X = 0xff
	testRunCPU(t, c)
	want := FlagN | Flag5
	have := c.SR
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// txs
// ----------------------------------------------------------------------------
func TestTxs(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0200, 0x9a)
	c.X = 0x12
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.SP
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

// ----------------------------------------------------------------------------
// tya
// ----------------------------------------------------------------------------
func TestTya(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0200, 0x98)
	c.Y = 0x12
	testRunCPU(t, c)
	want := uint8(0x12)
	have := c.A
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestTyaZero(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0200, 0x98)
	c.Y = 0x00
	testRunCPU(t, c)
	want := FlagZ | Flag5
	have := c.SR
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}

func TestTyaSigned(t *testing.T) {
	c := newTestCPU()
	c.mem.Write(0x0200, 0x98)
	c.Y = 0xff
	testRunCPU(t, c)
	want := FlagN | Flag5
	have := c.SR
	if want != have {
		t.Errorf("\n want: %02x \n have: %02x \n", want, have)
	}
}
