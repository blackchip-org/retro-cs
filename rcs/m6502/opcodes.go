package m6502

var opcodes = map[uint8]func(*CPU){
	0x00: func(c *CPU) { brk(c) },
	0x01: func(c *CPU) { ora(c, c.loadIndirectX) },
	0x05: func(c *CPU) { ora(c, c.loadZeroPage) },
	0x06: func(c *CPU) { asl(c, c.storeBack, c.loadZeroPage) },
	0x08: func(c *CPU) { php(c) }, // php
	0x09: func(c *CPU) { ora(c, c.loadImmediate) },
	0x0a: func(c *CPU) { asl(c, c.storeA, c.loadA) },
	0x0d: func(c *CPU) { ora(c, c.loadAbsolute) },
	0x0e: func(c *CPU) { asl(c, c.storeBack, c.loadAbsolute) },

	0x10: func(c *CPU) { branch(c, c.SR&FlagN == 0) }, // bpl
	0x11: func(c *CPU) { ora(c, c.loadIndirectY) },
	0x15: func(c *CPU) { ora(c, c.loadZeroPageX) },
	0x16: func(c *CPU) { asl(c, c.storeBack, c.loadZeroPageX) },
	0x18: func(c *CPU) { c.SR &^= FlagC }, // clc
	0x19: func(c *CPU) { ora(c, c.loadAbsoluteY) },
	0x1d: func(c *CPU) { ora(c, c.loadAbsoluteX) },
	0x1e: func(c *CPU) { asl(c, c.storeBack, c.loadAbsoluteX) },

	0x20: func(c *CPU) { jsr(c) },
	0x21: func(c *CPU) { and(c, c.loadIndirectX) },
	0x24: func(c *CPU) { bit(c, c.loadZeroPage) },
	0x25: func(c *CPU) { and(c, c.loadZeroPage) },
	0x26: func(c *CPU) { rol(c, c.storeBack, c.loadZeroPage) },
	0x28: func(c *CPU) { c.SR = c.pull() }, // plp
	0x29: func(c *CPU) { and(c, c.loadImmediate) },
	0x2a: func(c *CPU) { rol(c, c.storeA, c.loadA) },
	0x2c: func(c *CPU) { bit(c, c.loadAbsolute) },
	0x2d: func(c *CPU) { and(c, c.loadAbsolute) },
	0x2e: func(c *CPU) { rol(c, c.storeBack, c.loadAbsolute) },

	0x30: func(c *CPU) { branch(c, c.SR&FlagN != 0) }, // bmi
	0x31: func(c *CPU) { and(c, c.loadIndirectY) },
	0x35: func(c *CPU) { and(c, c.loadZeroPageX) },
	0x36: func(c *CPU) { rol(c, c.storeBack, c.loadZeroPageX) },
	0x38: func(c *CPU) { c.SR |= FlagC }, // sec
	0x39: func(c *CPU) { and(c, c.loadAbsoluteY) },
	0x3d: func(c *CPU) { and(c, c.loadAbsoluteX) },
	0x3e: func(c *CPU) { rol(c, c.storeBack, c.loadAbsoluteX) },

	0x40: func(c *CPU) { rti(c) },
	0x41: func(c *CPU) { eor(c, c.loadIndirectX) },
	0x45: func(c *CPU) { eor(c, c.loadZeroPage) },
	0x46: func(c *CPU) { lsr(c, c.storeBack, c.loadZeroPage) },
	0x48: func(c *CPU) { c.push(c.A) }, // pha
	0x49: func(c *CPU) { eor(c, c.loadImmediate) },
	0x4a: func(c *CPU) { lsr(c, c.storeA, c.loadA) },
	0x4c: func(c *CPU) { jmp(c) },
	0x4d: func(c *CPU) { eor(c, c.loadAbsolute) },
	0x4e: func(c *CPU) { lsr(c, c.storeBack, c.loadAbsolute) },

	0x50: func(c *CPU) { branch(c, c.SR&FlagV == 0) }, // bvc
	0x51: func(c *CPU) { eor(c, c.loadIndirectY) },
	0x55: func(c *CPU) { eor(c, c.loadZeroPageX) },
	0x56: func(c *CPU) { lsr(c, c.storeBack, c.loadZeroPageX) },
	0x58: func(c *CPU) { c.SR &^= FlagI }, // cli
	0x59: func(c *CPU) { eor(c, c.loadAbsoluteY) },
	0x5d: func(c *CPU) { eor(c, c.loadAbsoluteX) },
	0x5e: func(c *CPU) { lsr(c, c.storeBack, c.loadAbsoluteX) },

	0x60: func(c *CPU) { c.pc = c.pull2() }, // rts
	0x61: func(c *CPU) { adc(c, c.loadIndirectX) },
	0x65: func(c *CPU) { adc(c, c.loadZeroPage) },
	0x66: func(c *CPU) { ror(c, c.storeBack, c.loadZeroPage) },
	0x68: func(c *CPU) { pla(c) },
	0x69: func(c *CPU) { adc(c, c.loadImmediate) },
	0x6a: func(c *CPU) { ror(c, c.storeA, c.loadA) },
	0x6c: func(c *CPU) { jmpIndirect(c) },
	0x6d: func(c *CPU) { adc(c, c.loadAbsolute) },
	0x6e: func(c *CPU) { ror(c, c.storeBack, c.loadAbsolute) },

	0x70: func(c *CPU) { branch(c, c.SR&FlagV != 0) }, // bvs
	0x71: func(c *CPU) { adc(c, c.loadIndirectY) },
	0x75: func(c *CPU) { adc(c, c.loadZeroPageX) },
	0x76: func(c *CPU) { ror(c, c.storeBack, c.loadZeroPageX) },
	0x78: func(c *CPU) { c.SR |= FlagI }, // sei
	0x79: func(c *CPU) { adc(c, c.loadAbsoluteY) },
	0x7d: func(c *CPU) { adc(c, c.loadAbsoluteX) },
	0x7e: func(c *CPU) { ror(c, c.storeBack, c.loadAbsoluteX) },

	0x81: func(c *CPU) { st(c, c.storeIndirectX, c.loadA) },
	0x84: func(c *CPU) { st(c, c.storeZeroPage, c.loadY) },
	0x85: func(c *CPU) { st(c, c.storeZeroPage, c.loadA) },
	0x86: func(c *CPU) { st(c, c.storeZeroPage, c.loadX) },
	0x88: func(c *CPU) { dec(c, c.storeY, c.loadY) },
	0x8a: func(c *CPU) { ld(c, c.storeA, c.loadX) }, // txa
	0x8c: func(c *CPU) { st(c, c.storeAbsolute, c.loadY) },
	0x8d: func(c *CPU) { st(c, c.storeAbsolute, c.loadA) },
	0x8e: func(c *CPU) { st(c, c.storeAbsolute, c.loadX) },

	0x90: func(c *CPU) { branch(c, c.SR&FlagC == 0) }, // bcc
	0x91: func(c *CPU) { st(c, c.storeIndirectY, c.loadA) },
	0x94: func(c *CPU) { st(c, c.storeZeroPageX, c.loadY) },
	0x95: func(c *CPU) { st(c, c.storeZeroPageX, c.loadA) },
	0x96: func(c *CPU) { st(c, c.storeZeroPageY, c.loadX) },
	0x98: func(c *CPU) { ld(c, c.storeA, c.loadY) }, // tya
	0x99: func(c *CPU) { st(c, c.storeAbsoluteY, c.loadA) },
	0x9a: func(c *CPU) { ld(c, c.storeSP, c.loadX) }, // txs
	0x9d: func(c *CPU) { st(c, c.storeAbsoluteX, c.loadA) },

	0xa0: func(c *CPU) { ld(c, c.storeY, c.loadImmediate) },
	0xa1: func(c *CPU) { ld(c, c.storeA, c.loadIndirectX) },
	0xa2: func(c *CPU) { ld(c, c.storeX, c.loadImmediate) },
	0xa4: func(c *CPU) { ld(c, c.storeY, c.loadZeroPage) },
	0xa5: func(c *CPU) { ld(c, c.storeA, c.loadZeroPage) },
	0xa6: func(c *CPU) { ld(c, c.storeX, c.loadZeroPage) },
	0xa8: func(c *CPU) { ld(c, c.storeY, c.loadA) }, // tay
	0xa9: func(c *CPU) { ld(c, c.storeA, c.loadImmediate) },
	0xaa: func(c *CPU) { ld(c, c.storeX, c.loadA) }, // tax
	0xac: func(c *CPU) { ld(c, c.storeY, c.loadAbsolute) },
	0xad: func(c *CPU) { ld(c, c.storeA, c.loadAbsolute) },
	0xae: func(c *CPU) { ld(c, c.storeX, c.loadAbsolute) },

	0xb0: func(c *CPU) { branch(c, c.SR&FlagC != 0) }, // bcs
	0xb1: func(c *CPU) { ld(c, c.storeA, c.loadIndirectY) },
	0xb4: func(c *CPU) { ld(c, c.storeY, c.loadZeroPageX) },
	0xb5: func(c *CPU) { ld(c, c.storeA, c.loadZeroPageX) },
	0xb6: func(c *CPU) { ld(c, c.storeX, c.loadZeroPageY) },
	0xb8: func(c *CPU) { c.SR &^= FlagV }, // clv
	0xb9: func(c *CPU) { ld(c, c.storeA, c.loadAbsoluteY) },
	0xba: func(c *CPU) { ld(c, c.storeX, c.loadSP) }, // tsx
	0xbc: func(c *CPU) { ld(c, c.storeY, c.loadAbsoluteX) },
	0xbd: func(c *CPU) { ld(c, c.storeA, c.loadAbsoluteX) },
	0xbe: func(c *CPU) { ld(c, c.storeX, c.loadAbsoluteY) },

	0xc0: func(c *CPU) { cmp(c, c.loadY, c.loadImmediate) },
	0xc1: func(c *CPU) { cmp(c, c.loadA, c.loadIndirectX) },
	0xc4: func(c *CPU) { cmp(c, c.loadY, c.loadZeroPage) },
	0xc5: func(c *CPU) { cmp(c, c.loadA, c.loadZeroPage) },
	0xc6: func(c *CPU) { dec(c, c.storeBack, c.loadZeroPage) },
	0xc8: func(c *CPU) { inc(c, c.storeY, c.loadY) },
	0xc9: func(c *CPU) { cmp(c, c.loadA, c.loadImmediate) },
	0xca: func(c *CPU) { dec(c, c.storeX, c.loadX) },
	0xcc: func(c *CPU) { cmp(c, c.loadY, c.loadAbsolute) },
	0xcd: func(c *CPU) { cmp(c, c.loadA, c.loadAbsolute) },
	0xce: func(c *CPU) { dec(c, c.storeBack, c.loadAbsolute) },

	0xd0: func(c *CPU) { branch(c, c.SR&FlagZ == 0) }, // bne
	0xd1: func(c *CPU) { cmp(c, c.loadA, c.loadIndirectY) },
	0xd5: func(c *CPU) { cmp(c, c.loadA, c.loadZeroPageX) },
	0xd6: func(c *CPU) { dec(c, c.storeBack, c.loadZeroPageX) },
	0xd8: func(c *CPU) { c.SR &^= FlagD }, // cld
	0xd9: func(c *CPU) { cmp(c, c.loadA, c.loadAbsoluteY) },
	0xdd: func(c *CPU) { cmp(c, c.loadA, c.loadAbsoluteX) },
	0xde: func(c *CPU) { dec(c, c.storeBack, c.loadAbsoluteX) },

	0xe0: func(c *CPU) { cmp(c, c.loadX, c.loadImmediate) },
	0xe1: func(c *CPU) { sbc(c, c.loadIndirectX) },
	0xe4: func(c *CPU) { cmp(c, c.loadX, c.loadZeroPage) },
	0xe5: func(c *CPU) { sbc(c, c.loadZeroPage) },
	0xe6: func(c *CPU) { inc(c, c.storeBack, c.loadZeroPage) },
	0xe8: func(c *CPU) { inc(c, c.storeX, c.loadX) },
	0xe9: func(c *CPU) { sbc(c, c.loadImmediate) },
	0xea: func(c *CPU) {}, // nop
	0xec: func(c *CPU) { cmp(c, c.loadX, c.loadAbsolute) },
	0xed: func(c *CPU) { sbc(c, c.loadAbsolute) },
	0xee: func(c *CPU) { inc(c, c.storeBack, c.loadAbsolute) },

	0xf0: func(c *CPU) { branch(c, c.SR&FlagZ != 0) }, // beq
	0xf1: func(c *CPU) { sbc(c, c.loadIndirectY) },
	0xf5: func(c *CPU) { sbc(c, c.loadZeroPageX) },
	0xf6: func(c *CPU) { inc(c, c.storeBack, c.loadZeroPageX) },
	0xf8: func(c *CPU) { c.SR |= FlagD }, // sed
	0xf9: func(c *CPU) { sbc(c, c.loadAbsoluteY) },
	0xfd: func(c *CPU) { sbc(c, c.loadAbsoluteX) },
	0xfe: func(c *CPU) { inc(c, c.storeBack, c.loadAbsoluteX) },
}
