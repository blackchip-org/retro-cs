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
	if hi&(1<<7) != 0 {
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
	if rcs.Parity(out) {
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

// bit n,(ix+d)
func biti(cpu *CPU, n int, load rcs.Load8) {
	bit(cpu, n, load)

	// http://www.z80.info/zip/z80-documented.pdf
	// "This is where things start to get strange"
	cpu.F &^= Flag5 | Flag3
	iaddrh := cpu.iaddr >> 8
	if iaddrh&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if iaddrh&(1<<3) != 0 {
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
	carryIn := cpu.F&FlagC != 0
	cpu.F &^= FlagH | FlagN | FlagC | Flag5 | Flag3
	if carryIn {
		cpu.F |= FlagH
	}
	if !carryIn {
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

func cpx(cpu *CPU, increment int) {
	carry := cpu.F&FlagC != 0
	out, _, fh, _ := rcs.Sub(cpu.A, cpu.loadIndHL(), false)

	cpu.storeHL(cpu.loadHL() + int(increment))
	cpu.storeBC(cpu.loadBC() - int(1))

	dresult := out
	if fh {
		//dresult--
	}
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
	if cpu.loadBC() != 0 {
		cpu.F |= FlagV
	}
	cpu.F |= FlagN
	if carry {
		cpu.F |= FlagC
	}
	if dresult&(1<<1) != 0 { // yes, one
		cpu.F |= Flag5
	}
	if dresult&(1<<3) != 0 {
		cpu.F |= Flag3
	}
}

func cpxr(cpu *CPU, increment int) {
	cpx(cpu, increment)
	if cpu.B == 0 && cpu.C == 0 {
		return
	}
	if cpu.F&FlagZ != 0 {
		return
	}
	cpu.SetPC(cpu.PC() - 2)
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
	carry := false
	half := false
	if cpu.F&FlagN != 0 {
		if cpu.F&FlagH != 0 || cpu.A&0xf > 9 {
			out -= 6
		}
		if cpu.F&FlagC != 0 || cpu.A > 0x99 {
			out -= 0x60
			carry = true
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
			carry = true
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
	if rcs.Parity(out) {
		cpu.F |= FlagP
	}
	/*
		if cpu.A > 0x99 {
			cpu.F |= FlagC
		}
	*/
	if carry {
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

// interrupt mode set
func im(cpu *CPU, mode int) {
	cpu.IM = uint8(mode)
}

// port in
func in(cpu *CPU, store rcs.Store8, load rcs.Load8) {
	out := load()

	cpu.F &^= FlagS | FlagZ | FlagH | FlagV | FlagN | Flag5 | Flag3
	if out&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if out == 0 {
		cpu.F |= FlagZ
	}
	if rcs.Parity(out) {
		cpu.F |= FlagP
	}
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	store(out)
}

// increment
func inc(cpu *CPU, store rcs.Store8, load rcs.Load8) {
	in := load()
	out, _, fh, fv := rcs.Add(in, 1, false)

	cpu.F &^= FlagS | FlagZ | FlagH | FlagV | FlagN | Flag5 | Flag3
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

// port in, blocked
func inx(cpu *CPU, increment int) {
	in := cpu.inIndC()
	cpu.B--

	// https://github.com/mamedev/mame/blob/master/src/devices/device/proc/z80/z80.cpp
	// I was unable to figure this out by reading all the conflicting
	// documentation for these "undefined" flags
	t := uint16(cpu.C+uint8(increment)) + uint16(in)
	p := uint8(t&0x07) ^ cpu.B // parity check
	halfAndCarry := t&0x100 != 0

	cpu.F = 0
	if cpu.B&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if cpu.B == 0 {
		cpu.F |= FlagZ
	}
	if halfAndCarry {
		cpu.F |= FlagH
	}
	if rcs.Parity(p) {
		cpu.F |= FlagP
	}
	if in&(1<<7) != 0 {
		cpu.F |= FlagN
	}
	if halfAndCarry {
		cpu.F |= FlagC
	}
	if cpu.B&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if cpu.B&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	cpu.storeIndHL(in)
	ihl := (int(cpu.H)<<8 | int(cpu.L)) + increment
	cpu.H, cpu.L = uint8(ihl>>8), uint8(ihl)
}

// port in, blocked, repeat
func inxr(cpu *CPU, increment int) {
	inx(cpu, increment)
	for cpu.B != 0 {
		cpu.refreshR()
		cpu.refreshR()
		inx(cpu, increment)
	}
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

// ld a, i or ld a, r
func ldair(cpu *CPU, load rcs.Load8) {
	out := load()

	cpu.F &^= FlagS | FlagZ | FlagH | FlagV | FlagN | Flag5 | Flag3
	if out&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if out == 0 {
		cpu.F |= FlagZ
	}
	if cpu.IFF2 {
		cpu.F |= FlagV
	}
	if out&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if out&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	cpu.A = out
}

// load, 16-bit
func ld16(cpu *CPU, store rcs.Store, load rcs.Load) {
	store(load())
}

func ldx(cpu *CPU, increment int) {
	source := int(cpu.H)<<8 | int(cpu.L)
	target := int(cpu.D)<<8 | int(cpu.E)
	v := cpu.mem.Read(source)
	cpu.mem.Write(target, v)

	isource := source + increment
	itarget := target + increment

	cpu.H, cpu.L = uint8(isource>>8), uint8(isource)
	cpu.D, cpu.E = uint8(itarget>>8), uint8(itarget)

	counter := int(cpu.B)<<8 | int(cpu.C)
	counter--
	cpu.B, cpu.C = uint8(counter>>8), uint8(counter)

	cpu.F &^= FlagH | FlagV | FlagN | Flag5 | Flag3
	if counter != 0 {
		cpu.F |= FlagV
	}
	if (v+cpu.A)&(1<<1) != 0 { // yes, bit one
		cpu.F |= Flag5
	}
	if (v+cpu.A)&(1<<3) != 0 {
		cpu.F |= Flag3
	}
}

func ldxr(cpu *CPU, increment int) {
	ldx(cpu, increment)
	for cpu.B != 0 || cpu.C != 0 {
		cpu.refreshR()
		cpu.refreshR()
		ldx(cpu, increment)
	}
}

// no operation
func nop() {}

// negate accumulator
func neg(cpu *CPU) {
	sub(cpu, func() uint8 { return 0 }, cpu.loadA)
}

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
	if rcs.Parity(out) {
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

func outx(cpu *CPU, increment int) {
	in := cpu.mem.Read(int(cpu.H)<<8 | int(cpu.L))
	cpu.B--
	ihl := (int(cpu.H)<<8 | int(cpu.L)) + increment
	cpu.H, cpu.L = uint8(ihl>>8), uint8(ihl)

	// https://github.com/mamedev/mame/blob/master/src/devices/device/proc/z80/z80.cpp
	// I was unable to figure this out by reading all the conflicting
	// documentation for these "undefined" flags
	t := uint16(cpu.L) + uint16(in)
	p := uint8(t&0x07) ^ cpu.B // parity check
	halfAndCarry := t&0x100 != 0

	cpu.F = 0
	if cpu.B&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if cpu.B == 0 {
		cpu.F |= FlagZ
	}
	if halfAndCarry {
		cpu.F |= FlagH
	}
	if rcs.Parity(p) {
		cpu.F |= FlagP
	}
	if in&(1<<7) != 0 {
		cpu.F |= FlagN
	}
	if halfAndCarry {
		cpu.F |= FlagC
	}
	if cpu.B&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if cpu.B&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	cpu.Ports.Write(int(cpu.C), in)
}

// port out, blocked, repeat
func outxr(cpu *CPU, increment int) {
	outx(cpu, increment)
	for cpu.B != 0 {
		cpu.refreshR()
		cpu.refreshR()
		outx(cpu, increment)
	}
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

// return from interrupt
func reti(cpu *CPU) {
	cpu.SetPC(cpu.mem.ReadLE(int(cpu.SP)))
	cpu.SP += 2
}

// return from non-maskable interrupt
func retn(cpu *CPU) {
	cpu.IFF1 = cpu.IFF2
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
	if rcs.Parity(out) {
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
	if cpu.F&FlagC != 0 {
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
	if rcs.Parity(out) {
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

func rld(cpu *CPU) {
	addr := int(cpu.H)<<8 | int(cpu.L)
	ahi, alo := cpu.A>>4, cpu.A&0x0f
	readv := cpu.mem.Read(addr)
	memhi, memlo := readv>>4, readv&0x0f

	cpu.A = ahi<<4 | memhi
	memval := memlo<<4 | alo
	cpu.mem.Write(addr, memval)

	out := cpu.A
	cpu.F &^= FlagS | FlagZ | FlagH | FlagV | FlagN | Flag5 | Flag3
	if out&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if out == 0 {
		cpu.F |= FlagZ
	}
	if rcs.Parity(out) {
		cpu.F |= FlagP
	}
	if cpu.A&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if cpu.A&(1<<3) != 0 {
		cpu.F |= Flag3
	}
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
	if rcs.Parity(out) {
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
	if rcs.Parity(out) {
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

func rrd(cpu *CPU) {
	addr := int(cpu.H)<<8 | int(cpu.L)
	ahi, alo := cpu.A>>4, cpu.A&0x0f
	readv := cpu.mem.Read(addr)
	memhi, memlo := readv>>4, readv&0x0f

	cpu.A = ahi<<4 | memlo
	memval := alo<<4 | memhi
	cpu.mem.Write(addr, memval)

	out := cpu.A
	cpu.F &^= FlagS | FlagZ | FlagH | FlagV | FlagN | Flag5 | Flag3
	if out&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if out == 0 {
		cpu.F |= FlagZ
	}
	if rcs.Parity(out) {
		cpu.F |= FlagP
	}
	if cpu.A&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if cpu.A&(1<<3) != 0 {
		cpu.F |= Flag3
	}
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
	if rcs.Parity(out) {
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
	if rcs.Parity(out) {
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
	if rcs.Parity(out) {
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
	if rcs.Parity(out) {
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

// subtract 16-bit with carry
func sbc16(cpu *CPU, store rcs.Store, load0 rcs.Load, load1 rcs.Load) {
	in0 := load0()
	in1 := load1()

	lo, fc, _, _ := rcs.Sub(uint8(in0), uint8(in1), cpu.F&FlagC != 0)
	hi, fc, fh, fv := rcs.Sub(uint8(in0>>8), uint8(in1>>8), fc)

	cpu.F = 0
	if hi&(1<<7) != 0 {
		cpu.F |= FlagS
	}
	if hi == 0 && lo == 0 {
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
	if hi&(1<<5) != 0 {
		cpu.F |= Flag5
	}
	if hi&(1<<3) != 0 {
		cpu.F |= Flag3
	}
	store(int(hi)<<8 | int(lo))
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
	if rcs.Parity(out) {
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
