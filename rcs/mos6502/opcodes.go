package mos6502

var opcodes = map[uint8]func(*CPU){
	0x00: func(c *CPU) { brk(c) },

	0x61: func(c *CPU) { adc(c, c.loadIndirectX) },
	0x65: func(c *CPU) { adc(c, c.loadZeroPage) },
	0x69: func(c *CPU) { adc(c, c.loadImmediate) },
	0x6d: func(c *CPU) { adc(c, c.loadAbsolute) },

	0x71: func(c *CPU) { adc(c, c.loadIndirectY) },
	0x75: func(c *CPU) { adc(c, c.loadZeroPageX) },
	0x79: func(c *CPU) { adc(c, c.loadAbsoluteY) },
	0x7d: func(c *CPU) { adc(c, c.loadAbsoluteX) },
}
