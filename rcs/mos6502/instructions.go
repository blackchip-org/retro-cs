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

func brk(c *CPU) {
	c.SR |= FlagB
	c.fetch()
}
