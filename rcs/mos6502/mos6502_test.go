package mos6502

import (
	"testing"

	"github.com/blackchip-org/retro/rcs"
)

func newTestCPU() *CPU {
	mem := rcs.NewMemory(1, 0x10000)
	mem.MapRAM(0, make([]uint8, 0x1000, 0x1000)) // Only 4096 bytes
	cpu := New(mem)
	cpu.SP = 0xff
	cpu.SetPC(0x1ff)
	return cpu
}

func testRunCPU(t *testing.T, cpu *CPU) error {
	cycles := 0
	for cpu.SR&FlagB == 0 {
		cycles++
		if cycles > 100 {
			t.Error("max cycles exceeded")
		}
		cpu.Next()
	}
	return nil
}

func TestCPUString(t *testing.T) {

	var tests = []struct {
		setup func(*CPU)
		want  string
	}{
		{func(cpu *CPU) { cpu.SetPC(0x1234) },
			"" +
				" pc  sr ac xr yr sp  n v - b d i z c\n" +
				"1234 20 00 00 00 ff  . . * . . . . ."},
		{func(cpu *CPU) { cpu.A = 0x56 },
			"" +
				" pc  sr ac xr yr sp  n v - b d i z c\n" +
				"01ff 20 56 00 00 ff  . . * . . . . ."},
		{func(cpu *CPU) { cpu.X = 0x78 },
			"" +
				" pc  sr ac xr yr sp  n v - b d i z c\n" +
				"01ff 20 00 78 00 ff  . . * . . . . ."},
		{func(cpu *CPU) { cpu.Y = 0x9a },
			"" +
				" pc  sr ac xr yr sp  n v - b d i z c\n" +
				"01ff 20 00 00 9a ff  . . * . . . . ."},
		{func(cpu *CPU) { cpu.SP = 0xbc },
			"" +
				" pc  sr ac xr yr sp  n v - b d i z c\n" +
				"01ff 20 00 00 00 bc  . . * . . . . ."},
		{func(cpu *CPU) { cpu.SR = FlagC },
			"" +
				" pc  sr ac xr yr sp  n v - b d i z c\n" +
				"01ff 21 00 00 00 ff  . . * . . . . *"},
		{func(cpu *CPU) { cpu.SR = FlagZ },
			"" +
				" pc  sr ac xr yr sp  n v - b d i z c\n" +
				"01ff 22 00 00 00 ff  . . * . . . * ."},
		{func(cpu *CPU) { cpu.SR = FlagI },
			"" +
				" pc  sr ac xr yr sp  n v - b d i z c\n" +
				"01ff 24 00 00 00 ff  . . * . . * . ."},
		{func(cpu *CPU) { cpu.SR = FlagD },
			"" +
				" pc  sr ac xr yr sp  n v - b d i z c\n" +
				"01ff 28 00 00 00 ff  . . * . * . . ."},
		{func(cpu *CPU) { cpu.SR = FlagB },
			"" +
				" pc  sr ac xr yr sp  n v - b d i z c\n" +
				"01ff 30 00 00 00 ff  . . * * . . . ."},
		{func(cpu *CPU) { cpu.SR = FlagV },
			"" +
				" pc  sr ac xr yr sp  n v - b d i z c\n" +
				"01ff 60 00 00 00 ff  . * * . . . . ."},
		{func(cpu *CPU) { cpu.SR = FlagN },
			"" +
				" pc  sr ac xr yr sp  n v - b d i z c\n" +
				"01ff a0 00 00 00 ff  * . * . . . . ."},
	}

	for _, test := range tests {
		cpu := newTestCPU()
		test.setup(cpu)
		have := cpu.String()
		if test.want != have {
			t.Errorf("\n want: \n%v \n have: \n%v\n", test.want, have)
		}
	}
}
