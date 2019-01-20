package mos6502

import (
	"github.com/blackchip-org/retro/rcs"
)

func adc(c *CPU, load rcs.Load8) {
	if c.SR&FlagD != 0 {
		c.alu.AddBCD(load())
	} else {
		c.alu.Add(load())
	}
}

func and(c *CPU, load rcs.Load8) {
	c.alu.And(load())
}

func brk(c *CPU) {
	c.SR |= FlagB
	c.fetch()
}
