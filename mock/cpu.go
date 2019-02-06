package mock

import (
	"fmt"

	"github.com/blackchip-org/retro-cs/rcs"
)

type CPU struct {
	mem *rcs.Memory
	pc  uint16
	A   uint8 // sample register
	B   uint8 // sample register
	Q   bool  // sample flag
	Z   bool  // sample flag

	// Offset to be added to the program counter to get the address of the
	// next instruction.
	OffsetPC int
}

func NewCPU(mem *rcs.Memory) *CPU {
	return &CPU{mem: mem}
}

func (c *CPU) PC() int {
	return int(c.pc)
}

func (c *CPU) SetPC(addr int) {
	c.pc = uint16(addr)
}

func (c *CPU) Offset() int {
	return c.OffsetPC
}

// Next reads the next byte at the program counter as the "opcode". The high
// nibble is the number of "arguments" it will fetch (max two).
func (c *CPU) Next() {
	if c.OffsetPC == 1 {
		c.pc++
	}
	opcode := c.mem.Read(int(c.pc))
	if c.OffsetPC == 0 {
		c.pc++
	}
	narg := int(opcode) >> 4
	if narg > 2 {
		narg = 2
	}
	c.pc += uint16(narg)
}

func (c *CPU) String() string {
	return fmt.Sprintf("pc:%04x a:%02x b:%02x q:%v z:%v", c.pc, c.A, c.B, c.Q, c.Z)
}

func (c *CPU) Registers() map[string]rcs.Value {
	return map[string]rcs.Value{
		"pc": rcs.Value{
			Get: func() uint16 { return c.pc },
			Put: func(addr uint16) { c.pc = addr },
		},
		"a": rcs.Value{
			Get: func() uint8 { return c.A },
			Put: func(v uint8) { c.A = v },
		},
		"b": rcs.Value{
			Get: func() uint8 { return c.B },
			Put: func(v uint8) { c.B = v },
		},
	}
}

func (c *CPU) Flags() map[string]rcs.Value {
	return map[string]rcs.Value{
		"q": rcs.Value{
			Get: func() bool { return c.Q },
			Put: func(v bool) { c.Q = v },
		},
		"z": rcs.Value{
			Get: func() bool { return c.Z },
			Put: func(v bool) { c.Z = v },
		},
	}
}

func reader(e rcs.StmtEval) {
	e.Stmt.Addr = e.Ptr.Addr()
	opcode := e.Ptr.Fetch()
	e.Stmt.Bytes = append(e.Stmt.Bytes, opcode)
	argN := opcode >> 4
	switch argN {
	case 0:
		e.Stmt.Op = fmt.Sprintf("i%02x", opcode)
	case 1:
		value := e.Ptr.Fetch()
		e.Stmt.Bytes = append(e.Stmt.Bytes, value)
		e.Stmt.Op = fmt.Sprintf("i%02x $%02x", opcode, value)
	case 2:
		value := e.Ptr.FetchLE()
		e.Stmt.Bytes = append(e.Stmt.Bytes, uint8(value&0xff))
		e.Stmt.Bytes = append(e.Stmt.Bytes, uint8(value>>8))
		e.Stmt.Op = fmt.Sprintf("i%02x $%04x", opcode, value)
	default:
		e.Stmt.Op = fmt.Sprintf("?%02x", opcode)
	}
}

func formatter() rcs.CodeFormatter {
	options := rcs.FormatOptions{
		BytesFormat: "%-8s",
	}
	return func(s rcs.Stmt) string {
		return rcs.FormatStmt(s, options)
	}
}

func (c *CPU) NewDisassembler() *rcs.Disassembler {
	return rcs.NewDisassembler(c.mem, reader, formatter())
}
