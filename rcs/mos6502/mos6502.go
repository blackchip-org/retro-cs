package mos6502

import (
	"fmt"
	"log"

	"github.com/blackchip-org/retro/rcs"
)

const (
	addrStack = 0x0100
)

// CPU is the MOS Technology 6502 series processor.
type CPU struct {
	pc uint16 // program counter
	A  uint8  // accumulator
	X  uint8  // x register
	Y  uint8  // y register
	SP uint8  // stack pointer
	SR uint8  // status register

	IRQ bool // interrupt request

	mem       *rcs.Memory
	ops       map[uint8]func(*CPU)
	alu       rcs.ALU
	addrLoad  int // memory address where the last value was loaded from
	pageCross bool
}

const (
	// FlagC is the carry flag
	FlagC = uint8(1 << 0)

	// FlagZ is the zero flag
	FlagZ = uint8(1 << 1)

	// FlagI is the interrupt disable flag
	FlagI = uint8(1 << 2)

	// FlagD is the decimal mode flag
	FlagD = uint8(1 << 3)

	// FlagB is the break flag
	FlagB = uint8(1 << 4)

	// FlagV is the overflow flag
	FlagV = uint8(1 << 6)

	// FlagN is the signed/negative flag
	FlagN = uint8(1 << 7)
)

func New(mem *rcs.Memory) *CPU {
	c := &CPU{mem: mem, ops: opcodes}
	c.alu = rcs.ALU{
		C: FlagC,
		Z: FlagZ,
		V: FlagV,
		S: FlagN,
	}
	return c
}

func (c *CPU) Next() {
	c.pageCross = false
	opcode := c.fetch()
	execute, ok := c.ops[opcode]
	if !ok {
		log.Printf("illegal instruction: 0x%02x", opcode)
		return
	}
	execute(c)
}

func (c *CPU) PC() int {
	return int(c.pc)
}

func (c *CPU) SetPC(addr int) {
	c.pc = uint16(addr)
}

func (c *CPU) String() string {
	b := func(v bool) string {
		if v {
			return "*"
		}
		return "."
	}
	return fmt.Sprintf(""+
		" pc  sr ac xr yr sp  n v - b d i z c\n"+
		"%04x %02x %02x %02x %02x %02x  %s %s %s %s %s %s %s %s",
		c.pc,
		c.SR|(1<<5), // bit 5 hard wired on
		c.A,
		c.X,
		c.Y,
		c.SP,
		b(c.SR&FlagN != 0),
		b(c.SR&FlagV != 0),
		b(true),
		b(c.SR&FlagB != 0),
		b(c.SR&FlagD != 0),
		b(c.SR&FlagI != 0),
		b(c.SR&FlagZ != 0),
		b(c.SR&FlagC != 0),
	)
}

func (c *CPU) fetch() uint8 {
	c.pc++
	if c.pc > 0xffff {
		c.pc = 0
	}
	return c.mem.Read(int(c.pc))
}

func (c *CPU) fetch2() int {
	return int(c.fetch()) + (int(c.fetch()) << 8)
}

func (c *CPU) push(v uint8) {
	c.mem.Write(addrStack+int(c.SP), v)
}

func (c *CPU) push2(v int) {
	c.push(uint8(v >> 8))
	c.push(uint8(v))
}

func (c *CPU) pull() uint8 {
	c.SP++
	return c.mem.Read(addrStack + int(c.SP))
}

func (c *CPU) pull2() uint16 {
	return uint16(c.pull()) | uint16(c.pull())<<8
}
