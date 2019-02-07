package z80

import "github.com/blackchip-org/retro-cs/rcs"

func (c *CPU) Registers() map[string]rcs.Value {
	return map[string]rcs.Value{
		"pc": rcs.Value{
			Get: func() uint16 { return c.pc },
			Put: func(addr uint16) { c.pc = addr },
		},
		"a": rcs.Value{Get: c.loadA, Put: c.storeA},
		"f": rcs.Value{Get: c.loadF, Put: c.storeF},
		"b": rcs.Value{Get: c.loadB, Put: c.storeB},
		"c": rcs.Value{Get: c.loadC, Put: c.storeC},
		"d": rcs.Value{Get: c.loadD, Put: c.storeD},
		"e": rcs.Value{Get: c.loadE, Put: c.storeE},
		"h": rcs.Value{Get: c.loadH, Put: c.storeH},
		"l": rcs.Value{Get: c.loadL, Put: c.storeL},

		"af": rcs.Value{Get: c.loadAF, Put: c.storeAF},
		"bc": rcs.Value{Get: c.loadBC, Put: c.storeBC},
		"de": rcs.Value{Get: c.loadDE, Put: c.storeDE},
		"hl": rcs.Value{Get: c.loadHL, Put: c.storeHL},

		"a1": rcs.Value{Get: c.loadA, Put: c.storeA},
		"f1": rcs.Value{Get: c.loadF, Put: c.storeF},
		"b1": rcs.Value{Get: c.loadB, Put: c.storeB},
		"c1": rcs.Value{Get: c.loadC, Put: c.storeC},
		"d1": rcs.Value{Get: c.loadD, Put: c.storeD},
		"e1": rcs.Value{Get: c.loadE, Put: c.storeE},
		"h1": rcs.Value{Get: c.loadH, Put: c.storeH},
		"l1": rcs.Value{Get: c.loadL, Put: c.storeL},

		"af1": rcs.Value{Get: c.loadAF, Put: c.storeAF},
		"bc1": rcs.Value{Get: c.loadBC, Put: c.storeBC},
		"de1": rcs.Value{Get: c.loadDE, Put: c.storeDE},
		"hl1": rcs.Value{Get: c.loadHL, Put: c.storeHL},

		"i":  rcs.Value{Get: c.loadI, Put: c.storeI},
		"r":  rcs.Value{Get: c.loadR, Put: c.storeR},
		"ix": rcs.Value{Get: c.loadIX, Put: c.storeIX},
		"iy": rcs.Value{Get: c.loadIY, Put: c.storeIY},
		"sp": rcs.Value{Get: c.loadSP, Put: c.storeSP},

		"iff1": rcs.Value{
			Get: func() bool { return c.IFF1 },
			Put: func(v bool) { c.IFF1 = v },
		},
		"iff2": rcs.Value{
			Get: func() bool { return c.IFF2 },
			Put: func(v bool) { c.IFF2 = v },
		},
		"im": rcs.Value{
			Get: func() uint8 { return c.IM },
			Put: func(v uint8) { c.IM = v },
		},
		"halt": rcs.Value{
			Get: func() bool { return c.Halt },
			Put: func(v bool) { c.Halt = v },
		},
	}
}

func (c *CPU) Flags() map[string]rcs.Value {
	return map[string]rcs.Value{
		"c": rcs.Value{Get: c.getFlag(FlagC), Put: c.setFlag(FlagC)},
		"n": rcs.Value{Get: c.getFlag(FlagN), Put: c.setFlag(FlagN)},
		"v": rcs.Value{Get: c.getFlag(FlagV), Put: c.setFlag(FlagV)},
		"p": rcs.Value{Get: c.getFlag(FlagP), Put: c.setFlag(FlagP)},
		"3": rcs.Value{Get: c.getFlag(Flag3), Put: c.setFlag(Flag3)},
		"h": rcs.Value{Get: c.getFlag(FlagH), Put: c.setFlag(FlagH)},
		"5": rcs.Value{Get: c.getFlag(Flag5), Put: c.setFlag(Flag5)},
		"z": rcs.Value{Get: c.getFlag(FlagZ), Put: c.setFlag(FlagZ)},
		"s": rcs.Value{Get: c.getFlag(FlagS), Put: c.setFlag(FlagS)},
	}
}

func (c *CPU) getFlag(flag uint8) func() bool {
	return func() bool {
		return c.F&flag != 0
	}
}

func (c *CPU) setFlag(flag uint8) func(bool) {
	return func(v bool) {
		c.F &^= flag
		if v {
			c.F |= flag
		}
	}
}
