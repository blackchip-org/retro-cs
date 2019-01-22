package z80

import (
	"github.com/blackchip-org/retro-cs/rcs"
)

func flag35(flags *uint8, v uint8) {
	*flags &^= Flag3 | Flag5
	if v&(1<<3) != 0 {
		*flags |= Flag3
	}
	if v&(1<<5) != 0 {
		*flags |= Flag5
	}
}

// 16-bit addition
func add16(cpu *CPU, store rcs.Store, load0 rcs.Load, load1 rcs.Load, withCarry bool) {
	in0 := load0()
	in1 := load1()

	defer func() {
		cpu.alu.Ignore = 0
	}()
	if !withCarry {
		cpu.alu.Ignore = FlagS | FlagZ | FlagV
	}
	lo := cpu.alu.Add(&cpu.F, uint8(in0), uint8(in1))
	hi := cpu.alu.Add(&cpu.F, uint8(in0>>8), uint8(in1>>8))

	if withCarry && lo == 0 && hi == 0 {
		cpu.F |= FlagZ
	} else if withCarry {
		cpu.F &^= FlagZ
	}
	cpu.F &^= FlagN
	flag35(&cpu.F, uint8(hi))
	store(int(hi)<<8 | int(lo))
}

// decrement B and jump if not zero
func djnz(cpu *CPU, load rcs.Load8) {
	delta := load()
	cpu.B--
	if cpu.B != 0 {
		cpu.SetPC(cpu.PC() + int(int8(delta)))
	}
}

// exchange
func ex(cpu *CPU, load0 rcs.Load, store0 rcs.Store, load1 rcs.Load, store1 rcs.Store) {
	in0 := load0()
	in1 := load1()
	store0(in1)
	store1(in0)
}

// jump relative, conditional
func jr(cpu *CPU, flag uint8, condition bool, load rcs.Load8) {
	delta := int(int8(load()))
	flagSet := cpu.F&flag != 0
	if flagSet == condition {
		cpu.SetPC(cpu.PC() + delta)
	}
}

// jump relative, always
func jra(cpu *CPU, load rcs.Load8) {
	delta := load()
	cpu.SetPC(cpu.PC() + int(int8(delta)))
}

// load
func ld(cpu *CPU, store rcs.Store8, load rcs.Load8) {
	store(load())
}

// load, 16-bit
func ld16(cpu *CPU, store rcs.Store, load rcs.Load) {
	store(load())
}

// no operation
func nop() {}
