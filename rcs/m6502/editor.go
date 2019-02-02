package m6502

import "github.com/blackchip-org/retro-cs/rcs"

func (c *CPU) Registers() map[string]rcs.Value {
	return map[string]rcs.Value{
		"pc": rcs.Value{
			Get: func() uint16 { return c.pc },
			Put: func(addr uint16) { c.pc = addr },
		},
		"sr": rcs.Value{Get: c.loadSR, Put: c.storeSR},
		"a":  rcs.Value{Get: c.loadA, Put: c.storeA},
		"x":  rcs.Value{Get: c.loadX, Put: c.storeX},
		"y":  rcs.Value{Get: c.loadY, Put: c.storeY},
		"sp": rcs.Value{Get: c.loadSP, Put: c.storeSP},
	}
}

func (c *CPU) Flags() map[string]rcs.Value {
	return map[string]rcs.Value{
		"c": rcs.Value{Get: c.getFlag(FlagC), Put: c.setFlag(FlagC)},
		"z": rcs.Value{Get: c.getFlag(FlagZ), Put: c.setFlag(FlagZ)},
		"i": rcs.Value{Get: c.getFlag(FlagI), Put: c.setFlag(FlagI)},
		"d": rcs.Value{Get: c.getFlag(FlagD), Put: c.setFlag(FlagD)},
		"b": rcs.Value{Get: c.getFlag(FlagB), Put: c.setFlag(FlagB)},
		"v": rcs.Value{Get: c.getFlag(FlagV), Put: c.setFlag(FlagV)},
		"n": rcs.Value{Get: c.getFlag(FlagN), Put: c.setFlag(FlagN)},
	}
}

func (c *CPU) getFlag(flag uint8) func() bool {
	return func() bool {
		return c.SR&flag != 0
	}
}

func (c *CPU) setFlag(flag uint8) func(bool) {
	return func(v bool) {
		c.SR &^= flag
		if v {
			c.SR |= flag
		}
	}
}
