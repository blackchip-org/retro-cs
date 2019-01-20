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
