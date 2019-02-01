package mock

import "github.com/blackchip-org/retro-cs/rcs"

func NewMach() *rcs.Mach {
	ResetMemory()
	return &rcs.Mach{
		Mem: []*rcs.Memory{TestMemory},
		CPU: []rcs.CPU{NewCPU(TestMemory)},
	}
}

type CPU struct {
	mem *rcs.Memory
	pc  int
}

func NewCPU(mem *rcs.Memory) *CPU {
	return &CPU{mem: mem}
}

func (c *CPU) PC() int {
	return c.pc
}

func (c *CPU) SetPC(addr int) {
	c.pc = addr
}

// Next reads the next byte at the program counter as the "opcode". The high
// nibble is the number of "arguments" it will fetch.
func (c *CPU) Next() {
	opcode := c.mem.Read(c.pc)
	c.pc++
	narg := int(opcode) >> 4
	c.pc += int(narg)
}

func (c *CPU) String() string {
	return "cpu status registers"
}
