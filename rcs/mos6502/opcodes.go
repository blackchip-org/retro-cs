package mos6502

var opcodes = map[uint8]func(*CPU){
	0x00: func(c *CPU) { brk(c) },
	0x06: func(c *CPU) { asl(c, c.storeBack, c.loadZeroPage) },
	0x0a: func(c *CPU) { asl(c, c.storeA, c.loadA) },
	0x0e: func(c *CPU) { asl(c, c.storeBack, c.loadAbsolute) },

	0x10: func(c *CPU) { branch(c, c.SR&FlagN == 0) }, // bpl
	0x16: func(c *CPU) { asl(c, c.storeBack, c.loadZeroPageX) },
	0x1e: func(c *CPU) { asl(c, c.storeBack, c.loadAbsoluteX) },

	0x21: func(c *CPU) { and(c, c.loadIndirectX) },
	0x24: func(c *CPU) { bit(c, c.loadZeroPage) },
	0x25: func(c *CPU) { and(c, c.loadZeroPage) },
	0x29: func(c *CPU) { and(c, c.loadImmediate) },
	0x2c: func(c *CPU) { bit(c, c.loadAbsolute) },
	0x2d: func(c *CPU) { and(c, c.loadAbsolute) },

	0x30: func(c *CPU) { branch(c, c.SR&FlagN != 0) }, // bmi
	0x31: func(c *CPU) { and(c, c.loadIndirectY) },
	0x35: func(c *CPU) { and(c, c.loadZeroPageX) },
	0x39: func(c *CPU) { and(c, c.loadAbsoluteY) },
	0x3d: func(c *CPU) { and(c, c.loadAbsoluteX) },

	0x50: func(c *CPU) { branch(c, c.SR&FlagV == 0) }, // bvc

	0x61: func(c *CPU) { adc(c, c.loadIndirectX) },
	0x65: func(c *CPU) { adc(c, c.loadZeroPage) },
	0x69: func(c *CPU) { adc(c, c.loadImmediate) },
	0x6d: func(c *CPU) { adc(c, c.loadAbsolute) },

	0x70: func(c *CPU) { branch(c, c.SR&FlagV != 0) }, // bvs
	0x71: func(c *CPU) { adc(c, c.loadIndirectY) },
	0x75: func(c *CPU) { adc(c, c.loadZeroPageX) },
	0x79: func(c *CPU) { adc(c, c.loadAbsoluteY) },
	0x7d: func(c *CPU) { adc(c, c.loadAbsoluteX) },

	0x81: func(c *CPU) { st(c, c.storeIndirectX, c.loadA) },
	0x84: func(c *CPU) { st(c, c.storeZeroPage, c.loadY) },
	0x85: func(c *CPU) { st(c, c.storeZeroPage, c.loadA) },
	0x86: func(c *CPU) { st(c, c.storeZeroPage, c.loadX) },
	0x8c: func(c *CPU) { st(c, c.storeAbsolute, c.loadY) },
	0x8d: func(c *CPU) { st(c, c.storeAbsolute, c.loadA) },
	0x8e: func(c *CPU) { st(c, c.storeAbsolute, c.loadX) },

	0x90: func(c *CPU) { branch(c, c.SR&FlagC == 0) }, // bcc
	0x91: func(c *CPU) { st(c, c.storeIndirectY, c.loadA) },
	0x94: func(c *CPU) { st(c, c.storeZeroPageX, c.loadY) },
	0x95: func(c *CPU) { st(c, c.storeZeroPageX, c.loadA) },
	0x96: func(c *CPU) { st(c, c.storeZeroPageY, c.loadX) },
	0x99: func(c *CPU) { st(c, c.storeAbsoluteY, c.loadA) },
	0x9d: func(c *CPU) { st(c, c.storeAbsoluteX, c.loadA) },

	0xa0: func(c *CPU) { ld(c, c.storeY, c.loadImmediate) },
	0xa1: func(c *CPU) { ld(c, c.storeA, c.loadIndirectX) },
	0xa2: func(c *CPU) { ld(c, c.storeX, c.loadImmediate) },
	0xa4: func(c *CPU) { ld(c, c.storeY, c.loadZeroPage) },
	0xa5: func(c *CPU) { ld(c, c.storeA, c.loadZeroPage) },
	0xa6: func(c *CPU) { ld(c, c.storeX, c.loadZeroPage) },
	0xa9: func(c *CPU) { ld(c, c.storeA, c.loadImmediate) },
	0xac: func(c *CPU) { ld(c, c.storeY, c.loadAbsolute) },
	0xad: func(c *CPU) { ld(c, c.storeA, c.loadAbsolute) },
	0xae: func(c *CPU) { ld(c, c.storeX, c.loadAbsolute) },

	0xb0: func(c *CPU) { branch(c, c.SR&FlagC != 0) }, // bcs
	0xb1: func(c *CPU) { ld(c, c.storeA, c.loadIndirectY) },
	0xb4: func(c *CPU) { ld(c, c.storeY, c.loadZeroPageX) },
	0xb5: func(c *CPU) { ld(c, c.storeA, c.loadZeroPageX) },
	0xb6: func(c *CPU) { ld(c, c.storeX, c.loadZeroPageY) },
	0xb9: func(c *CPU) { ld(c, c.storeA, c.loadAbsoluteY) },
	0xbc: func(c *CPU) { ld(c, c.storeY, c.loadAbsoluteX) },
	0xbd: func(c *CPU) { ld(c, c.storeA, c.loadAbsoluteX) },
	0xbe: func(c *CPU) { ld(c, c.storeX, c.loadAbsoluteY) },

	0xd0: func(c *CPU) { branch(c, c.SR&FlagZ == 0) }, // bne

	0xf0: func(c *CPU) { branch(c, c.SR&FlagZ != 0) }, // beq
}
