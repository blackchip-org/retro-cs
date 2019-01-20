package mos6502

import (
	"github.com/blackchip-org/retro/rcs"
)

func adc(c *CPU, load rcs.Load8) {
	if c.SR&FlagD != 0 {
		c.alu.AddBCD(&c.SR, &c.A, c.A, load())
	} else {
		c.alu.Add(&c.SR, &c.A, c.A, load())
	}
}

func and(c *CPU, load rcs.Load8) {
	c.alu.And(&c.SR, &c.A, c.A, load())
}

func asl(c *CPU, store rcs.Store8, load rcs.Load8) {
	var out uint8
	c.SR &^= FlagC // shift in zero
	c.alu.ShiftLeft(&c.SR, &out, load())
	store(out)
}

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

func brk(c *CPU) {
	c.SR |= FlagB
	c.fetch()
}

func ld(c *CPU, store rcs.Store8, load rcs.Load8) {
	value := load()
	store(value)
	c.alu.Pass(&c.SR, value)
}

func st(c *CPU, store rcs.Store8, load rcs.Load8) {
	store(load())
}
