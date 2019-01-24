package z80

import (
	"github.com/blackchip-org/retro-cs/rcs"
)

// add
func add(cpu *CPU, load0 rcs.Load8, load1 rcs.Load8) {
	out, fc, fh, fv := rcs.Add(load0(), load1(), false)

	cpu.F = 0
	if out&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if out == 0 {
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
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	cpu.A = out
}

// add with carry
func adc(cpu *CPU, load0 rcs.Load8, load1 rcs.Load8) {
	out, fc, fh, fv := rcs.Add(load0(), load1(), cpu.F&FlagC != 0)

	cpu.F = 0
	if out&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if out == 0 {
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
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	cpu.A = out
}

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

// bitwise logical and
func and(cpu *CPU, load rcs.Load8) {
	out := cpu.A & load()

	cpu.F = 0
	if out&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if out == 0 {
		cpu.F |= FlagZ
	}
	cpu.F |= FlagH
	if rcs.Parity8(out) {
		cpu.F |= FlagP
	}
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	cpu.A = out
}

// test bit
func bit(cpu *CPU, n int, load rcs.Load8) {
	bit := uint(n)
	out := load()
	cpu.F &^= FlagS | FlagZ | FlagV | FlagN | Flag5 | Flag3
	if n == 7 && out&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if out&(1<<bit) == 0 {
		cpu.F |= FlagZ
	}
	cpu.F |= FlagH
	if out&(1<<bit) == 0 {
		cpu.F |= FlagV
	}
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
}

// call, conditional
func call(cpu *CPU, flag uint8, condition bool, load rcs.Load) {
	addr := load()
	if (cpu.F&flag != 0) == condition {
		cpu.SP -= 2
		cpu.mem.WriteLE(int(cpu.SP), cpu.PC())
		cpu.SetPC(addr)
	}
}

// call, always
func calla(cpu *CPU, load rcs.Load) {
	addr := load()
	cpu.SP -= 2
	cpu.mem.WriteLE(int(cpu.SP), cpu.PC())
	cpu.SetPC(addr)
}

// invert carry flag
func ccf(cpu *CPU) {
	// The H flag was tricky. Correct definition in the Z80 User Manual
	cpu.F &^= FlagH | FlagN | FlagC | Flag5 | Flag3
	if cpu.F|FlagC != 0 {
		cpu.F |= FlagH
	}
	if cpu.F|FlagC == 0 {
		cpu.F |= FlagC
	}
	if cpu.A&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if cpu.A&(1<<3) != 0 {
		cpu.F |= Flag3
	}
}

// CP is a subtraction from A that doesn't update A, only the flags it would
// have set/reset if it really was subtracted.
//
// F5 and F3 are copied from the operand, not the result
func cp(cpu *CPU, load rcs.Load8) {
	a := cpu.A
	in := load()
	sub(cpu, cpu.loadA, func() uint8 { return in })
	cpu.F &^= Flag5 | Flag3
	if in&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if in&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	cpu.A = a
}

// invert accumulator, one's complement
func cpl(cpu *CPU) {
	out := ^cpu.A
	cpu.F &^= Flag5 | Flag3
	cpu.F |= FlagH | FlagN
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	cpu.A = out
}

// decimal adjust in a
// http://datasheets.chipdb.org/Zilog/Z80/z80-documented-0.90.pdf
//
// Note: some documentation omits that the adjustment is negative when the
// N flag is set.
//
// Eventually ported directly from the MAME source code.
func daa(cpu *CPU) {
	out := cpu.A
	half := false
	if cpu.F&FlagN != 0 {
		if cpu.F&FlagH != 0 || cpu.A&0xf > 9 {
			out -= 6
		}
		if cpu.F&FlagC != 0 || cpu.A > 0x99 {
			out -= 0x60
		}
		if cpu.F&FlagH != 0 && cpu.A&0xf <= 0x5 {
			half = true
		}
	} else {
		if cpu.F&FlagH != 0 || cpu.A&0xf > 9 {
			out += 6
		}
		if cpu.F&FlagC != 0 || cpu.A > 0x99 {
			out += 0x60
		}
		if cpu.A&0xf > 0x9 {
			half = true
		}
	}

	cpu.F &^= FlagS | FlagZ | FlagH | FlagP | FlagC | Flag5 | Flag3
	if out&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if out == 0 {
		cpu.F |= FlagZ
	}
	if half {
		cpu.F |= FlagH
	}
	if rcs.Parity8(out) {
		cpu.F |= FlagP
	}
	if cpu.A > 0x99 {
		cpu.F |= FlagC
	}
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	cpu.A = out
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

// disable interrupts
func di(cpu *CPU) {
	cpu.IFF1 = false
	cpu.IFF2 = false
}

// decrement B and jump if not zero
func djnz(cpu *CPU, load rcs.Load8) {
	delta := load()
	cpu.B--
	if cpu.B != 0 {
		cpu.SetPC(cpu.PC() + int(int8(delta)))
	}
}

// enable interrupts
func ei(cpu *CPU) {
	cpu.IFF1 = true
	cpu.IFF2 = true
}

// exchange
func ex(cpu *CPU, load0 rcs.Load, store0 rcs.Store, load1 rcs.Load, store1 rcs.Store) {
	in0 := load0()
	in1 := load1()
	store0(in1)
	store1(in0)
}

// EXX exchanges BC, DE, and HL with shadow registers with BC', DE', and HL'.
func exx(cpu *CPU) {
	ex(cpu, cpu.loadBC, cpu.storeBC, cpu.loadBC1, cpu.storeBC1)
	ex(cpu, cpu.loadDE, cpu.storeDE, cpu.loadDE1, cpu.storeDE1)
	ex(cpu, cpu.loadHL, cpu.storeHL, cpu.loadHL1, cpu.storeHL1)
}

// halt
func halt(cpu *CPU) {
	cpu.Halt = true
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

// jump absolute, conditional
func jp(cpu *CPU, flag uint8, condition bool, load rcs.Load) {
	addr := load()
	if (cpu.F&flag != 0) == condition {
		cpu.SetPC(addr)
	}
}

// jump absolute, always
func jpa(cpu *CPU, load rcs.Load) {
	cpu.SetPC(load())
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

// bitwise logical or
func or(cpu *CPU, load rcs.Load8) {
	out := cpu.A | load()

	cpu.F = 0
	if out&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if out == 0 {
		cpu.F |= FlagZ
	}
	if rcs.Parity8(out) {
		cpu.F |= FlagP
	}
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	cpu.A = out
}

// Copies the two bytes from (SP) into the operand, then increases SP by 2.
func pop(cpu *CPU, store rcs.Store) {
	store(cpu.mem.ReadLE(int(cpu.SP)))
	cpu.SP += 2
}

// Decrements the SP by 2 then copies the operand into (SP)
func push(cpu *CPU, load rcs.Load) {
	cpu.SP -= 2
	cpu.mem.WriteLE(int(cpu.SP), load())
}

// reset bit
func res(cpu *CPU, n int, store rcs.Store8, load rcs.Load8) {
	bit := uint(n)
	store(load() &^ (1 << bit))
}

// return, conditional
func ret(cpu *CPU, flag uint8, value bool) {
	if (cpu.F&flag != 0) == value {
		reta(cpu)
	}
}

// return, always
func reta(cpu *CPU) {
	cpu.SetPC(cpu.mem.ReadLE(int(cpu.SP)))
	cpu.SP += 2
}

// rotate left
func rl(cpu *CPU, store rcs.Store8, load rcs.Load8) {
	in := load()
	carryOut := in&(1<<7) != 0
	out := in << 1
	if cpu.F&FlagC != 0 {
		out++
	}

	cpu.F = 0
	if out&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if out == 0 {
		cpu.F |= FlagZ
	}
	if rcs.Parity8(out) {
		cpu.F |= FlagP
	}
	if carryOut {
		cpu.F |= FlagC
	}
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	store(out)
}

// rotate accumulator left
func rla(cpu *CPU) {
	carryOut := cpu.A&(1<<7) != 0
	out := cpu.A << 1
	if cpu.F|FlagC != 0 {
		out++
	}

	cpu.F &^= FlagH | FlagN | FlagC | Flag5 | Flag3
	if carryOut {
		cpu.F |= FlagC
	}
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	cpu.A = out
}

// rotate accumulator left with carry
func rlca(cpu *CPU) {
	carryOut := cpu.A&(1<<7) != 0
	out := cpu.A << 1
	if carryOut {
		out++
	}

	cpu.F &^= FlagH | FlagN | FlagC | Flag5 | Flag3
	if carryOut {
		cpu.F |= FlagC
	}
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	cpu.A = out
}

// rotate left with carry
func rlc(cpu *CPU, store rcs.Store8, load rcs.Load8) {
	in := load()
	carryOut := in&(1<<7) != 0
	out := in << 1
	if carryOut {
		out++
	}

	cpu.F = 0
	if out&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if out == 0 {
		cpu.F |= FlagZ
	}
	if rcs.Parity8(out) {
		cpu.F |= FlagP
	}
	if carryOut {
		cpu.F |= FlagC
	}
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	store(out)
}

// rotate right
func rr(cpu *CPU, store rcs.Store8, load rcs.Load8) {
	in := load()
	carryOut := in&1 != 0
	out := in >> 1
	if cpu.F&FlagC != 0 {
		out |= (1 << 7)
	}

	cpu.F = 0
	if out&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if out == 0 {
		cpu.F |= FlagZ
	}
	if rcs.Parity8(out) {
		cpu.F |= FlagP
	}
	if carryOut {
		cpu.F |= FlagC
	}
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	store(out)
}

// rotate accumulator right
func rra(cpu *CPU) {
	carryOut := cpu.A&1 != 0
	out := cpu.A >> 1
	if cpu.F&FlagC != 0 {
		out |= (1 << 7)
	}

	cpu.F &^= FlagH | FlagN | FlagC | Flag5 | Flag3
	if carryOut {
		cpu.F |= FlagC
	}
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	cpu.A = out
}

// rotate right
func rrc(cpu *CPU, store rcs.Store8, load rcs.Load8) {
	in := load()
	carryOut := in&1 != 0
	out := in >> 1
	if carryOut {
		out |= (1 << 7)
	}

	cpu.F = 0
	if out&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if out == 0 {
		cpu.F |= FlagZ
	}
	if rcs.Parity8(out) {
		cpu.F |= FlagP
	}
	if carryOut {
		cpu.F |= FlagC
	}
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}

	store(out)
}

// rotate accumulator right with carry
func rrca(cpu *CPU) {
	carryOut := cpu.A&1 != 0
	out := cpu.A >> 1
	if carryOut {
		out |= (1 << 7)
	}

	cpu.F &^= FlagH | FlagN | FlagC | Flag5 | Flag3
	if carryOut {
		cpu.F |= FlagC
	}
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	cpu.A = out
}

// reset
func rst(cpu *CPU, y int) {
	cpu.SP -= 2
	cpu.mem.WriteLE(int(cpu.SP), cpu.PC())
	cpu.SetPC(y * 8)
}

// set carry flag
func scf(cpu *CPU) {
	cpu.F &^= FlagH | FlagN | Flag5 | Flag3
	cpu.F |= FlagC
	if cpu.A&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if cpu.A&(1<<3) != 0 {
		cpu.F |= Flag3
	}
}

// set bit
func set(cpu *CPU, n int, store rcs.Store8, load rcs.Load8) {
	bit := uint(n)
	store(load() | (1 << bit))
}

// shift left, arithemtic
func sla(cpu *CPU, store rcs.Store8, load rcs.Load8) {
	in := load()
	carryOut := in&(1<<7) != 0
	out := in << 1

	cpu.F = 0
	if out&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if out == 0 {
		cpu.F |= FlagZ
	}
	if rcs.Parity8(out) {
		cpu.F |= FlagV
	}
	if carryOut {
		cpu.F |= FlagC
	}
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	store(out)
}

// shift right, arithemtic
func sra(cpu *CPU, store rcs.Store8, load rcs.Load8) {
	in := load()
	carryOut := in&1 != 0
	bit7 := in&(1<<7) != 0
	out := in >> 1
	if bit7 {
		out |= 1 << 7
	}

	cpu.F = 0
	if out&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if out == 0 {
		cpu.F |= FlagZ
	}
	if rcs.Parity8(out) {
		cpu.F |= FlagV
	}
	if carryOut {
		cpu.F |= FlagC
	}
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	store(out)
}

// undocumented: shift left, logical
func sll(cpu *CPU, store rcs.Store8, load rcs.Load8) {
	in := load()
	carryOut := in&(1<<7) != 0
	out := in<<1 + 1

	cpu.F = 0
	if out&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if out == 0 {
		cpu.F |= FlagZ
	}
	if rcs.Parity8(out) {
		cpu.F |= FlagV
	}
	if carryOut {
		cpu.F |= FlagC
	}
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	store(out)
}

// shift right, logical
func srl(cpu *CPU, store rcs.Store8, load rcs.Load8) {
	in := load()
	carryOut := in&1 != 0
	out := in >> 1

	cpu.F = 0
	if out&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if out == 0 {
		cpu.F |= FlagZ
	}
	if rcs.Parity8(out) {
		cpu.F |= FlagV
	}
	if carryOut {
		cpu.F |= FlagC
	}
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	store(out)
}

// subtract
func sub(cpu *CPU, load0 rcs.Load8, load1 rcs.Load8) {
	out, fc, fh, fv := rcs.Sub(load0(), load1(), false)

	cpu.F = 0
	if out&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if out == 0 {
		cpu.F |= FlagZ
	}
	if fh {
		cpu.F |= FlagH
	}
	if fv {
		cpu.F |= FlagV
	}
	cpu.F |= FlagN
	if fc {
		cpu.F |= FlagC
	}
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	cpu.A = out
}

// subtract with carry
func sbc(cpu *CPU, load0 rcs.Load8, load1 rcs.Load8) {
	out, fc, fh, fv := rcs.Sub(load0(), load1(), cpu.F&FlagC != 0)

	cpu.F = 0
	if out&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if out == 0 {
		cpu.F |= FlagZ
	}
	if fh {
		cpu.F |= FlagH
	}
	if fv {
		cpu.F |= FlagV
	}
	cpu.F |= FlagN
	if fc {
		cpu.F |= FlagC
	}
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	cpu.A = out
}

// bitwise logical exclusive or
func xor(cpu *CPU, load rcs.Load8) {
	out := cpu.A ^ load()

	cpu.F = 0
	if out&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if out == 0 {
		cpu.F |= FlagZ
	}
	if rcs.Parity8(out) {
		cpu.F |= FlagP
	}
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	cpu.A = out
}
