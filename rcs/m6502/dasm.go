package m6502

type mode int

const (
	absolute mode = iota
	absoluteX
	absoluteY
	accumulator
	immediate
	implied
	indirect
	indirectX
	indirectY
	relative
	zeroPage
	zeroPageX
	zeroPageY
)

type op struct {
	inst string
	mode mode
}

var operandLengths = map[mode]int{
	absolute:    2,
	absoluteX:   2,
	absoluteY:   2,
	accumulator: 0,
	immediate:   1,
	implied:     0,
	indirect:    2,
	indirectX:   1,
	indirectY:   1,
	relative:    1,
	zeroPage:    1,
	zeroPageX:   1,
	zeroPageY:   1,
}

var operandFormats = map[mode]string{
	absolute:    "$%04x",
	absoluteX:   "$%04x,x",
	absoluteY:   "$%04x,y",
	accumulator: "a",
	immediate:   "#$%02x",
	indirect:    "($%04x)",
	indirectX:   "($%02x,x)",
	indirectY:   "($%02x),y",
	relative:    "$%04x",
	zeroPage:    "$%02x",
	zeroPageX:   "$%02x,x",
	zeroPageY:   "$%02x,y",
}

var dasmTable = map[uint8]op{
	0x00: op{"brk", implied},
	0x01: op{"ora", indirectX},
	0x05: op{"ora", zeroPage},
	0x06: op{"asl", zeroPage},
	0x08: op{"php", implied},
	0x09: op{"ora", immediate},
	0x0a: op{"asl", accumulator},
	0x0d: op{"ora", absolute},
	0x0e: op{"asl", absolute},

	0x10: op{"bpl", relative},
	0x11: op{"ora", indirectY},
	0x15: op{"ora", zeroPageX},
	0x16: op{"asl", zeroPageX},
	0x18: op{"clc", implied},
	0x19: op{"ora", absoluteY},
	0x1d: op{"ora", absoluteX},
	0x1e: op{"asl", absoluteX},

	0x20: op{"jsr", absolute},
	0x21: op{"and", indirectX},
	0x24: op{"bit", zeroPage},
	0x25: op{"and", zeroPage},
	0x26: op{"rol", zeroPage},
	0x28: op{"plp", implied},
	0x29: op{"and", immediate},
	0x2a: op{"rol", accumulator},
	0x2c: op{"bit", absolute},
	0x2d: op{"and", absolute},
	0x2e: op{"rol", absolute},

	0x30: op{"bmi", relative},
	0x31: op{"and", indirectY},
	0x35: op{"and", zeroPageX},
	0x36: op{"rol", zeroPageX},
	0x38: op{"sec", implied},
	0x39: op{"and", absoluteY},
	0x3d: op{"and", absoluteX},
	0x3e: op{"rol", absoluteX},

	0x40: op{"rti", implied},
	0x41: op{"eor", indirectX},
	0x45: op{"eor", zeroPage},
	0x46: op{"lsr", zeroPage},
	0x48: op{"pha", implied},
	0x49: op{"eor", immediate},
	0x4a: op{"lsr", accumulator},
	0x4c: op{"jmp", absolute},
	0x4d: op{"eor", absolute},
	0x4e: op{"lsr", absolute},

	0x50: op{"bvc", relative},
	0x51: op{"eor", indirectY},
	0x55: op{"eor", zeroPageX},
	0x56: op{"lsr", zeroPageX},
	0x58: op{"cli", implied},
	0x59: op{"eor", absoluteY},
	0x5d: op{"eor", absoluteX},
	0x5e: op{"lsr", absoluteX},

	0x60: op{"rts", implied},
	0x61: op{"adc", indirectX},
	0x65: op{"adc", zeroPage},
	0x66: op{"ror", zeroPage},
	0x68: op{"pla", implied},
	0x69: op{"adc", immediate},
	0x6a: op{"ror", accumulator},
	0x6c: op{"jmp", indirect},
	0x6d: op{"adc", absolute},
	0x6e: op{"ror", absolute},

	0x70: op{"bvs", relative},
	0x71: op{"adc", indirectY},
	0x75: op{"adc", zeroPageX},
	0x76: op{"ror", zeroPageX},
	0x78: op{"sei", implied},
	0x79: op{"adc", absoluteY},
	0x7d: op{"adc", absoluteX},
	0x7e: op{"ror", absoluteX},

	0x81: op{"sta", indirectX},
	0x84: op{"sty", zeroPage},
	0x85: op{"sta", zeroPage},
	0x86: op{"stx", zeroPage},
	0x88: op{"dey", implied},
	0x8a: op{"txa", implied},
	0x8c: op{"sty", absolute},
	0x8d: op{"sta", absolute},
	0x8e: op{"stx", absolute},

	0x90: op{"bcc", relative},
	0x91: op{"sta", indirectY},
	0x94: op{"sty", zeroPageX},
	0x95: op{"sta", zeroPageX},
	0x96: op{"stx", zeroPageY},
	0x98: op{"tya", implied},
	0x99: op{"sta", absoluteY},
	0x9a: op{"txs", implied},
	0x9d: op{"sta", absoluteX},

	0xa0: op{"ldy", immediate},
	0xa1: op{"lda", indirectX},
	0xa2: op{"ldx", immediate},
	0xa4: op{"ldy", zeroPage},
	0xa5: op{"lda", zeroPage},
	0xa6: op{"ldx", zeroPage},
	0xa8: op{"tay", implied},
	0xa9: op{"lda", immediate},
	0xaa: op{"tax", implied},
	0xac: op{"ldy", absolute},
	0xad: op{"lda", absolute},
	0xae: op{"ldx", absolute},

	0xb0: op{"bcs", relative},
	0xb1: op{"lda", indirectY},
	0xb4: op{"ldy", zeroPageX},
	0xb5: op{"lda", zeroPageX},
	0xb6: op{"ldx", zeroPageY},
	0xb8: op{"clv", implied},
	0xb9: op{"lda", absoluteY},
	0xba: op{"tsx", implied},
	0xbd: op{"lda", absoluteX},
	0xbc: op{"ldy", absoluteX},
	0xbe: op{"ldx", absoluteY},

	0xc0: op{"cpy", immediate},
	0xc1: op{"cmp", indirectX},
	0xc4: op{"cpy", zeroPage},
	0xc5: op{"cmp", zeroPage},
	0xc6: op{"dec", zeroPage},
	0xc8: op{"iny", implied},
	0xc9: op{"cmp", immediate},
	0xca: op{"dex", implied},
	0xcc: op{"cpy", absolute},
	0xcd: op{"cmp", absolute},
	0xce: op{"dec", absolute},

	0xd0: op{"bne", relative},
	0xd1: op{"cmp", indirectY},
	0xd5: op{"cmp", zeroPageX},
	0xd6: op{"dec", zeroPageX},
	0xd8: op{"cld", implied},
	0xd9: op{"cmp", absoluteY},
	0xdd: op{"cmp", absoluteX},
	0xde: op{"dec", absoluteX},

	0xe0: op{"cpx", immediate},
	0xe1: op{"sbc", indirectX},
	0xe4: op{"cpx", zeroPage},
	0xe5: op{"sbc", zeroPage},
	0xe6: op{"inc", zeroPage},
	0xe8: op{"inx", implied},
	0xe9: op{"sbc", immediate},
	0xea: op{"nop", implied},
	0xec: op{"cpx", absolute},
	0xed: op{"sbc", absolute},
	0xee: op{"inc", absolute},

	0xf0: op{"beq", relative},
	0xf1: op{"sbc", indirectY},
	0xf5: op{"sbc", zeroPageX},
	0xf6: op{"inc", zeroPageX},
	0xf8: op{"sed", implied},
	0xf9: op{"sbc", absoluteY},
	0xfd: op{"sbc", absoluteX},
	0xfe: op{"inc", absoluteX},
}
