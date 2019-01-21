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

	mem   *rcs.Memory
	ops   map[uint8]func(*CPU)
	alu   rcs.ALU
	delta uint8
	// address used to load on the last (IX+d) or (IY+d) instruction
	iaddr uint16
}

const (
	FlagS = 7
	FlagZ = 6
	Flag5 = 5
	FlagH = 4
	Flag3 = 3
	FlagV = 2
	FlagN = 1
	FlagC = 0
)

func New(mem *rcs.Memory) *CPU {
	c := &CPU{mem: mem}
	c.alu = rcs.ALU{
		C: FlagC,
		V: FlagV,
		H: FlagH,
		Z: FlagZ,
		S: FlagS,
	}
	return c
}

func (c *CPU) Next() {
	here := c.PC()
	opcode := c.fetch()
	execute, ok := c.ops[opcode]
	c.refreshR()
	if !ok {
		log.Printf("%04x: illegal instruction: 0x%02x", here, opcode)
		return
	}
	execute(c)
}

func (c *CPU) PC() uint16 {
	return c.pc
}

func (c *CPU) SetPC(pc uint16) {
	c.pc = pc
}

func (c *CPU) fetch() uint8 {
	c.pc++
	return c.mem.Read(int(c.pc - 1))
}

func (c *CPU) fetch2() int {
	return int(c.fetch()) + (int(c.fetch()) << 8)
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