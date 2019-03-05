package m6502

import (
	"github.com/blackchip-org/retro-cs/rcs"
)

// add with carry
func adc(cpu *CPU, load rcs.Load8) {
	if cpu.SR&FlagD != 0 {
		adcd(cpu, load)
		return
	}

	out, c, _, v := rcs.Add(cpu.A, load(), cpu.SR&FlagC != 0)
	cpu.SR &^= FlagN | FlagV | FlagZ | FlagC
	if out&(1<<7) != 0 {
		cpu.SR |= FlagN
	}
	if v {
		cpu.SR |= FlagV
	}
	if out == 0 {
		cpu.SR |= FlagZ
	}
	if c {
		cpu.SR |= FlagC
	}
	cpu.A = out
}

// add binary-coded decimal
func adcd(c *CPU, load rcs.Load8) {
	carry := 0
	if c.SR&FlagC != 0 {
		carry = 1
	}

	in0 := rcs.FromBCD(c.A)
	in1 := rcs.FromBCD(load())
	outn := uint16(in0) + uint16(in1) + uint16(carry)
	out := rcs.ToBCD(uint8(outn))

	c.SR &^= FlagC | FlagZ | FlagN
	if outn > 99 {
		c.SR |= FlagC
	}
	if out == 0 {
		c.SR |= FlagZ
	}
	if out&(1<<7) != 0 {
		c.SR |= FlagN
	}
	c.A = out
}

// logical and
func and(c *CPU, load rcs.Load8) {
	out := c.A & load()
	c.SR &^= FlagN | FlagZ
	if out&(1<<7) != 0 {
		c.SR |= FlagN
	}
	if out == 0 {
		c.SR |= FlagZ
	}
	c.A = out
}

// arithmetic shift left
func asl(c *CPU, store rcs.Store8, load rcs.Load8) {
	in := load()
	carryOut := in&(1<<7) != 0
	out := in << 1

	c.SR &^= FlagN | FlagZ | FlagC
	if carryOut {
		c.SR |= FlagC
	}
	if out&(1<<7) != 0 {
		c.SR |= FlagN
	}
	if out == 0 {
		c.SR |= FlagZ
	}
	store(out)
}

// test bits
func bit(c *CPU, load rcs.Load8) {
	in := load()

	if c.A&in == 0 {
		c.SR |= FlagZ
	} else {
		c.SR &^= FlagZ
	}

	if in&(1<<7) != 0 {
		c.SR |= FlagN
	} else {
		c.SR &^= FlagN
	}

	if in&(1<<6) != 0 {
		c.SR |= FlagV
	} else {
		c.SR &^= FlagV
	}
}

// branch instructions
func branch(c *CPU, do bool) {
	displacement := int8(c.fetch())
	if do {
		if displacement >= 0 {
			c.SetPC(c.PC() + int(displacement))
		} else {
			c.SetPC(c.PC() - int(displacement*-1))
		}
	}
}

// break
func brk(c *CPU) {
	if c.BreakFunc != nil {
		c.BreakFunc()
	} else {
		c.fetch()
		c.irqAck(true)
	}
}

// compare
func cmp(c *CPU, load0 rcs.Load8, load1 rcs.Load8) {
	// C set as if subtraction. Clear if 'borrow', otherwise set
	out := int16(load0()) - int16(load1())

	c.SR &^= FlagC | FlagN | FlagZ
	if out >= 0 {
		c.SR |= FlagC
	}
	if out&(1<<7) != 0 {
		c.SR |= FlagN
	}
	if out == 0 {
		c.SR |= FlagZ
	}
}

// decrement
func dec(c *CPU, store rcs.Store8, load rcs.Load8) {
	out := load() - 1
	c.SR &^= FlagN | FlagZ
	if out&(1<<7) != 0 {
		c.SR |= FlagN
	}
	if out == 0 {
		c.SR |= FlagZ
	}
	store(out)
}

// exclusive or
func eor(c *CPU, load rcs.Load8) {
	out := c.A ^ load()
	c.SR &^= FlagN | FlagZ
	if out&(1<<7) != 0 {
		c.SR |= FlagN
	}
	if out == 0 {
		c.SR |= FlagZ
	}
	c.A = out
}

// increment
func inc(c *CPU, store rcs.Store8, load rcs.Load8) {
	out := load() + 1
	c.SR &^= FlagN | FlagZ
	if out&(1<<7) != 0 {
		c.SR |= FlagN
	}
	if out == 0 {
		c.SR |= FlagZ
	}
	store(out)
}

// jump
func jmp(c *CPU) {
	c.pc = uint16(c.fetch2() - 1)
}

// jump indirect
func jmpIndirect(c *CPU) {
	c.pc = uint16(c.mem.ReadLE(c.fetch2()) - 1)
}

// jump to subroutine
func jsr(c *CPU) {
	addr := uint16(c.fetch2())
	c.push2(c.pc)
	c.pc = addr - 1
}

// load
func ld(c *CPU, store rcs.Store8, load rcs.Load8) {
	out := load()
	store(out)
	c.SR &^= FlagN | FlagZ
	if out&(1<<7) != 0 {
		c.SR |= FlagN
	}
	if out == 0 {
		c.SR |= FlagZ
	}
}

// logical shift right
func lsr(c *CPU, store rcs.Store8, load rcs.Load8) {
	in := load()
	carryOut := in&(1<<0) != 0
	out := in >> 1

	c.SR &^= FlagN | FlagZ | FlagC
	if carryOut {
		c.SR |= FlagC
	}
	if out&(1<<7) != 0 {
		c.SR |= FlagN
	}
	if out == 0 {
		c.SR |= FlagZ
	}
	store(out)
}

// logical or
func ora(c *CPU, load rcs.Load8) {
	out := c.A | load()
	c.SR &^= FlagN | FlagZ
	if out&(1<<7) != 0 {
		c.SR |= FlagN
	}
	if out == 0 {
		c.SR |= FlagZ
	}
	c.A = out
}

// push processor status
func php(c *CPU) {
	// https://wiki.nesdev.com/w/index.php/Status_flags
	c.push(c.SR | FlagB | Flag5)
}

// pull accumulator
func pla(c *CPU) {
	out := c.pull()
	c.SR &^= FlagN | FlagZ
	if out&(1<<7) != 0 {
		c.SR |= FlagN
	}
	if out == 0 {
		c.SR |= FlagZ
	}
	c.A = out
}

// rotate left
func rol(c *CPU, store rcs.Store8, load rcs.Load8) {
	in := load()
	carryOut := in&(1<<7) != 0
	out := in << 1
	if c.SR&FlagC != 0 {
		out++
	}

	c.SR &^= FlagN | FlagZ | FlagC
	if carryOut {
		c.SR |= FlagC
	}
	if out&(1<<7) != 0 {
		c.SR |= FlagN
	}
	if out == 0 {
		c.SR |= FlagZ
	}
	store(out)
}

// rotate right
func ror(c *CPU, store rcs.Store8, load rcs.Load8) {
	in := load()
	carryOut := in&(1<<0) != 0
	out := in >> 1
	if c.SR&FlagC != 0 {
		out |= (1 << 7)
	}

	c.SR &^= FlagN | FlagZ | FlagC
	if carryOut {
		c.SR |= FlagC
	}
	if out&(1<<7) != 0 {
		c.SR |= FlagN
	}
	if out == 0 {
		c.SR |= FlagZ
	}
	store(out)
}

// return from interrupt
func rti(c *CPU) {
	// http://www.6502.org/tutorials/6502opcodes.html#RTI
	// Note that unlike RTS, the return address on the stack is the
	// actual address rather than the address-1.
	c.SR = c.pull()
	c.pc = c.pull2() - 1
}

// subtract with carry
func sbc(c *CPU, load rcs.Load8) {
	if c.SR&FlagD != 0 {
		sbcd(c, load)
		return
	}

	out, fc, _, fv := rcs.Sub(c.A, load(), c.SR&FlagC == 0) // carry clear on borrow
	c.SR &^= FlagN | FlagV | FlagZ | FlagC
	if out&(1<<7) != 0 {
		c.SR |= FlagN
	}
	if fv {
		c.SR |= FlagV
	}
	if out == 0 {
		c.SR |= FlagZ
	}
	if !fc { // carry clear on borrow
		c.SR |= FlagC
	}
	c.A = out
}

// subtract, binary-coded decimal
func sbcd(c *CPU, load rcs.Load8) {
	borrow := 0
	if c.SR&FlagC == 0 {
		borrow = 1 // borrow on carry clear
	}

	in0 := rcs.FromBCD(c.A)
	in1 := rcs.FromBCD(load())
	outn := int16(in0) - int16(in1) - int16(borrow)
	carry := outn >= 0
	if outn < 0 {
		outn += 100
	}
	out := rcs.ToBCD(uint8(outn))

	c.SR &^= FlagC | FlagZ | FlagN
	if carry {
		c.SR |= FlagC
	}
	if out == 0 {
		c.SR |= FlagZ
	}
	if out&(1<<7) != 0 {
		c.SR |= FlagN
	}
	c.A = out
}

// store
func st(c *CPU, store rcs.Store8, load rcs.Load8) {
	store(load())
}
