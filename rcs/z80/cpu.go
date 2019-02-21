package z80

import (
	"fmt"
	"log"

	"github.com/blackchip-org/retro-cs/rcs"
)

// CPU is the Zilog Z80 processor.
type CPU struct {
	pc uint16 // Program counter
	A  uint8  // Accumulator
	F  uint8  // Flags
	B  uint8
	C  uint8
	D  uint8
	E  uint8
	H  uint8
	L  uint8

	A1 uint8 // Shadow registers
	F1 uint8
	B1 uint8
	C1 uint8
	D1 uint8
	E1 uint8
	H1 uint8
	L1 uint8

	I   uint8 // Interrupt vector base
	R   uint8 // DRAM refresh counter
	IXH uint8
	IXL uint8
	IYH uint8
	IYL uint8
	SP  uint16 // Stack pointer

	IFF1 bool // Interrupt flip flops
	IFF2 bool
	IM   uint8 // Interrupt mode
	Halt bool  // Halted by instruction

	Ports   *rcs.Memory
	IRQ     bool
	IRQData uint8
	NMI     bool
	RESET   bool

	opcodes     map[uint8]func(*CPU)
	opcodesCB   map[uint8]func(*CPU)
	opcodesED   map[uint8]func(*CPU)
	opcodesDD   map[uint8]func(*CPU)
	opcodesFD   map[uint8]func(*CPU)
	opcodesDDCB map[uint8]func(*CPU)
	opcodesFDCB map[uint8]func(*CPU)

	mem   *rcs.Memory
	delta uint8
	// address used to load on the last (IX+d) or (IY+d) instruction
	iaddr int
}

const (
	// FlagC is the carry flag
	FlagC = uint8(1 << 0)

	// FlagN is set after subtraction
	FlagN = uint8(1 << 1)

	// FlagV is the overflow flag (also parity)
	FlagV = uint8(1 << 2)

	// FlagP is the parity flag (also overflow)
	FlagP = uint8(1 << 2)

	// Flag3 is undefined
	Flag3 = uint8(1 << 3)

	// FlagH is the half-carry flag
	FlagH = uint8(1 << 4)

	// Flag5 is undefined
	Flag5 = uint8(1 << 5)

	// FlagZ is the zero flag
	FlagZ = uint8(1 << 6)

	// FlagS is the sign flag
	FlagS = uint8(1 << 7)
)

func New(mem *rcs.Memory) *CPU {
	c := &CPU{
		mem:         mem,
		Ports:       rcs.NewMemory(1, 0x100),
		opcodes:     opcodes,
		opcodesCB:   opcodesCB,
		opcodesED:   opcodesED,
		opcodesDD:   opcodesDD,
		opcodesFD:   opcodesFD,
		opcodesDDCB: opcodesDDCB,
		opcodesFDCB: opcodesFDCB,
	}
	c.Ports.MapRAM(0, make([]uint8, 0x100, 0x100))
	return c
}

func (c *CPU) Next() {
	if !c.Halt {
		c.execute()
	}
	if c.IRQ {
		c.IRQ = false
		if c.IFF1 {
			c.irqAck()
		}
	}
	if c.NMI {
		c.NMI = false
		c.nmiAck()
	}
	if c.RESET {
		c.RESET = false
		c.resetAck()
	}
}

func (c *CPU) execute() {
	here := c.PC()
	opcode := c.fetch()
	c.refreshR()

	prefix := ""
	var table map[uint8]func(*CPU)
	switch opcode {
	case 0xcb:
		table = c.opcodesCB
		opcode = c.fetch()
		c.refreshR()
		prefix = "cb"
	case 0xed:
		table = c.opcodesED
		opcode = c.fetch()
		c.refreshR()
		prefix = "ed"
	case 0xdd:
		table = c.opcodesDD
		opcode = c.fetch()
		c.refreshR()
		prefix = "dd"
		if opcode == 0xcb {
			table = c.opcodesDDCB
			c.fetchd()
			opcode = c.fetch()
			prefix = "ddcb"
		}
	case 0xfd:
		table = c.opcodesFD
		opcode = c.fetch()
		c.refreshR()
		prefix = "fd"
		if opcode == 0xcb {
			table = c.opcodesFDCB
			c.fetchd()
			opcode = c.fetch()
			prefix = "fdcb"
		}
	default:
		table = c.opcodes
	}

	opFunc, ok := table[opcode]
	if !ok {
		log.Printf("%04x: illegal instruction: %v%02x", here, prefix, opcode)
		return
	}
	opFunc(c)
}

func (c *CPU) irqAck() {
	if c.IM == 0 {
		log.Printf("unsupported interrupt mode 0")
		return
	}
	c.Halt = false
	c.IFF1 = false
	c.IFF2 = false
	c.SP -= 2
	c.mem.WriteLE(int(c.SP), c.PC())
	if c.IM == 2 {
		vector := int(c.I)<<8 | int(c.IRQData)
		c.SetPC(c.mem.ReadLE(vector))
	} else {
		c.pc = 0x0038
	}
}

func (c *CPU) nmiAck() {
	c.SP -= 2
	c.mem.WriteLE(int(c.SP), c.PC())
	c.pc = 0x0066
}

func (c *CPU) resetAck() {
	c.IFF1 = false
	c.IFF2 = false
	c.pc = 0
	c.I = 0
	c.R = 0
	c.IM = 0
}

// PC returns the value of the program counter.
func (c *CPU) PC() int {
	return int(c.pc)
}

// SetPC sets the value of the program counter.
func (c *CPU) SetPC(pc int) {
	c.pc = uint16(pc)
}

// Offset is the value to be added to the program counter to get the
// address of the next instruction. The value is 0 for this CPU since
// the program counter is incremented after fetching the opcode.
func (c *CPU) Offset() int {
	return 0
}

// NewDisassembler creates a disassembler that can handle Z80 machine
// code.
func (c *CPU) NewDisassembler() *rcs.Disassembler {
	dasm := rcs.NewDisassembler(c.mem, Reader, Formatter())
	return dasm
}

func (c *CPU) fetch() uint8 {
	c.pc++
	return c.mem.Read(int(c.pc - 1))
}

func (c *CPU) fetch2() int {
	return int(c.fetch()) + (int(c.fetch()) << 8)
}

func (cpu *CPU) fetchd() {
	cpu.delta = cpu.fetch()
}

func (c *CPU) refreshR() {
	// Lower 7 bits of the refresh register are incremented on an instruction
	// fetch
	bit7 := c.R & 0x80
	c.R = (c.R+1)&0x7f | bit7
}

func (c *CPU) String() string {
	b := func(v uint8, ch string) string {
		if v != 0 {
			return ch
		}
		return "."
	}

	iff1 := ""
	if c.IFF1 {
		iff1 = "iff1"
	}
	iff2 := ""
	if c.IFF2 {
		iff2 = "iff2"
	}

	return fmt.Sprintf(""+
		" pc   af   bc   de   hl   ix   iy   sp   i  r\n"+
		"%04x %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x %04x  %02x %02x %v\n"+
		"im %v %02x%02x %02x%02x %02x%02x %02x%02x      %v %v %v %v %v %v %v %v  %v\n",
		// line 1
		c.pc,
		c.A, c.F,
		c.B, c.C,
		c.D, c.E,
		c.H, c.L,
		c.IXH, c.IXL,
		c.IYH, c.IYL,
		c.SP,
		c.I,
		c.R,
		iff1,
		// line 2
		c.IM,
		c.A1, c.F1,
		c.B1, c.C1,
		c.D1, c.E1,
		c.H1, c.L1,
		// flags
		b(c.F&FlagS, "S"),
		b(c.F&FlagZ, "Z"),
		b(c.F&Flag5, "5"),
		b(c.F&FlagH, "H"),
		b(c.F&Flag3, "3"),
		b(c.F&FlagV, "V"),
		b(c.F&FlagN, "N"),
		b(c.F&FlagC, "C"),
		iff2,
	)
}

func (c *CPU) Save(enc *rcs.Encoder) {
	enc.Encode(c.A)
	enc.Encode(c.F)
	enc.Encode(c.B)
	enc.Encode(c.C)
	enc.Encode(c.D)
	enc.Encode(c.H)
	enc.Encode(c.L)

	enc.Encode(c.A1)
	enc.Encode(c.F1)
	enc.Encode(c.B1)
	enc.Encode(c.C1)
	enc.Encode(c.D1)
	enc.Encode(c.H1)
	enc.Encode(c.L1)

	enc.Encode(c.I)
	enc.Encode(c.R)
	enc.Encode(c.IXH)
	enc.Encode(c.IXL)
	enc.Encode(c.IYH)
	enc.Encode(c.IYL)
	enc.Encode(c.SP)
	enc.Encode(c.pc)

	enc.Encode(c.IFF1)
	enc.Encode(c.IFF2)
	enc.Encode(c.IM)
	enc.Encode(c.Halt)
}

func (c *CPU) Load(dec *rcs.Decoder) {
	dec.Decode(&c.A)
	dec.Decode(&c.F)
	dec.Decode(&c.B)
	dec.Decode(&c.C)
	dec.Decode(&c.D)
	dec.Decode(&c.H)
	dec.Decode(&c.L)

	dec.Decode(&c.A1)
	dec.Decode(&c.F1)
	dec.Decode(&c.B1)
	dec.Decode(&c.C1)
	dec.Decode(&c.D1)
	dec.Decode(&c.H1)
	dec.Decode(&c.L1)

	dec.Decode(&c.I)
	dec.Decode(&c.R)
	dec.Decode(&c.IXH)
	dec.Decode(&c.IXL)
	dec.Decode(&c.IYH)
	dec.Decode(&c.IYL)
	dec.Decode(&c.SP)
	dec.Decode(&c.pc)

	dec.Decode(&c.IFF1)
	dec.Decode(&c.IFF2)
	dec.Decode(&c.IM)
	dec.Decode(&c.Halt)
}

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
