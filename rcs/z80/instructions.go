package z80

import (
	"github.com/blackchip-org/retro-cs/rcs"
)

// 16-bit addition, without carry
func add16(cpu *CPU, store rcs.Store, load0 rcs.Load, load1 rcs.Load) {
	in0 := load0()
	in1 := load1()

	lo, fc, _, _ := rcs.Add(uint8(in0), uint8(in1), false)
	hi, fc, fh, _ := rcs.Add(uint8(in0>>8), uint8(in1>>8), fc)

	cpu.F &^= FlagH | FlagN | FlagC | Flag5 | Flag3

	if fh {
		cpu.F |= FlagH
	}
	if fc {
		cpu.F |= FlagC
	}
	if hi&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if hi&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	store(int(hi)<<8 | int(lo))
}

// 16-bit addition, with carry
func adc16(cpu *CPU, store rcs.Store, load0 rcs.Load, load1 rcs.Load) {
	in0 := load0()
	in1 := load1()

	lo, fc, _, _ := rcs.Add(uint8(in0), uint8(in1), cpu.F&FlagC != 0)
	hi, fc, fh, fv := rcs.Add(uint8(in0>>8), uint8(in1>>8), fc)

	cpu.F = 0
	if hi&(1<<7) == 0 {
		cpu.F |= FlagS
	}
	if lo == 0 && hi == 0 {
		cpu.F |= FlagZ
	}
	if fh {
		cpu.F |= FlagH
	}
	if fv {
		cpu.F |= FlagV
	}
	if fc {
		cpu.F |= FlagC
	}
	if hi&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if hi&(1<<3) != 0 {
		cpu.F |= Flag3
	}

	store(int(hi)<<8 | int(lo))
}

// decrement
func dec(cpu *CPU, store rcs.Store8, load rcs.Load8) {
	in := load()
	out, _, fh, fv := rcs.Sub(in, 1, false)

	cpu.F &^= FlagS | FlagZ | FlagH | FlagV | Flag5 | Flag3
	if out&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if out == 0 {
		cpu.F |= FlagZ
	}
	if fh {
		cpu.F |= FlagH
	}
	// The Z80 user manual says that the overflow flag is set if the value
	// "was 0x80 before operation". This conflicts with the FUSE tests which
	// assume the overflow flag is set as if subtraction was done.
	if fv {
		cpu.F |= FlagV
	}
	cpu.F |= FlagN
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	store(out)
}

// decrement 16-bit, no flags altered
func dec16(cpu *CPU, store rcs.Store, load rcs.Load) {
	in0 := uint16(load())
	store(int(in0 - 1))
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

// increment
func inc(cpu *CPU, store rcs.Store8, load rcs.Load8) {
	in := load()
	out, _, fh, fv := rcs.Add(in, 1, false)

	cpu.F &^= FlagS | FlagZ | FlagH | FlagV | Flag5 | Flag3
	if out&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if out == 0 {
		cpu.F |= FlagZ
	}
	if fh {
		cpu.F |= FlagH
	}
	// The Z80 user manual says that the overflow flag is set if the value
	// "was 0x80 before operation". This conflicts with the FUSE tests which
	// assume the overflow flag is set as if subtraction was done.
	if fv {
		cpu.F |= FlagV
	}
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	store(out)
}

// increment 16-bit, no flags altered
func inc16(cpu *CPU, store rcs.Store, load rcs.Load) {
	in0 := uint16(load())
	store(int(in0 + 1))
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
