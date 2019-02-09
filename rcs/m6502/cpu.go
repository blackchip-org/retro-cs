package m6502

import (
	"fmt"
	"log"

	"github.com/blackchip-org/retro-cs/rcs"
)

// http://www.6502.org/tutorials/6502opcodes.html

const (
	addrStack = 0x0100 // starting address of the stack
	addrReset = 0xfffc // reset vector
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

	mem       *rcs.Memory          // CPU's view into memory
	ops       map[uint8]func(*CPU) // opcode table
	addrLoad  int                  // memory address where the last value was loaded from
	pageCross bool                 // if set, add a one cycle penalty for crossing a page boundary
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

// New creates a new CPU with a view of the provided memory.
func New(mem *rcs.Memory) *CPU {
	return &CPU{
		mem: mem,
		pc:  uint16(mem.ReadLE(addrReset) - 1), // reset vector
		ops: opcodes,
	}
}

// Next executes the next instruction.
func (c *CPU) Next() (here int, halt bool) {
	halt = false
	if c.SR&FlagB != 0 {
		c.SR &^= FlagB
		c.SR |= FlagI
		c.irqAck()
	}
	if c.IRQ {
		c.IRQ = false
		if c.SR&FlagI == 0 {
			c.irqAck()
		}
	}

	here = c.PC()
	c.pageCross = false
	opcode := c.fetch()
	execute, ok := c.ops[opcode]
	if !ok {
		log.Printf("%04x: illegal instruction: 0x%02x", here, opcode)
		return
	}
	execute(c)
	return
}

// interrupt handler
func (c *CPU) irqAck() {
	// http://www.6502.org/tutorials/6502opcodes.html#RTI
	// Note that unlike RTS, the return address on the stack is the
	// actual address rather than the address-1.
	c.push2(c.pc + 1)
	c.push(c.SR)
	c.pc = uint16(c.mem.ReadLE(0xfffe) - 1) // irq vector
}

// PC returns the value of the program counter.
func (c *CPU) PC() int {
	return int(c.pc)
}

// SetPC sets the value of the program counter.
func (c *CPU) SetPC(addr int) {
	c.pc = uint16(addr)
}

// Offset is the value to be added to the program counter to get the
// address of the next instruction. The value is 1 for this CPU since
// the program counter is incremented before fetching the opcode.
func (c *CPU) Offset() int {
	return 1
}

// NewDisassembler creates a disassembler that can handle 6502 machine
// code.
func (c *CPU) NewDisassembler() *rcs.Disassembler {
	dasm := rcs.NewDisassembler(c.mem, Reader, Formatter())
	return dasm
}

// String returns the status of the CPU in the form of:
// 		 pc  sr ac xr yr sp  n v - b d i z c
// 		1234 20 00 00 00 ff  . . * . . . . .
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

// Increment the program counter and return the 8-bit value at the
// program counter.
func (c *CPU) fetch() uint8 {
	c.pc++
	return c.mem.Read(int(c.pc))
}

// Like fetch, but return the next 16-bit value.
func (c *CPU) fetch2() int {
	return int(c.fetch()) + (int(c.fetch()) << 8)
}

// Push a 8-bit value to the stack.
func (c *CPU) push(v uint8) {
	c.mem.Write(addrStack+int(c.SP), v)
	c.SP--
}

// Push a 16-bit value to the stack.
func (c *CPU) push2(v uint16) {
	c.push(uint8(v >> 8))
	c.push(uint8(v))
}

// Pull a 8-bit value from the stack.
func (c *CPU) pull() uint8 {
	c.SP++
	return c.mem.Read(addrStack + int(c.SP))
}

// Pull a 16-bit value from the stack.
func (c *CPU) pull2() uint16 {
	return uint16(c.pull()) | uint16(c.pull())<<8
}

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
