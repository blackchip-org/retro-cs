package z80

import (
	"fmt"
	"testing"

	"github.com/blackchip-org/retro-cs/mock"
	"github.com/blackchip-org/retro-cs/rcs"
)

type harstonTest struct {
	Name  string
	Op    string
	Bytes []uint8
}

func TestDasm(t *testing.T) {
	for _, test := range harstonTests {
		t.Run(test.Name, func(t *testing.T) {
			mock.ResetMemory()
			ptr := rcs.NewPointer(mock.TestMemory)
			dasm := NewDisassembler(mock.TestMemory)
			dasm.SetPC(0x10)
			ptr.SetAddr(0x10)
			ptr.PutN(test.Bytes...)
			ptr.SetAddr(0x10)
			s := dasm.NextStmt()
			if s.Op != test.Op {
				t.Errorf("\n have: %v \n want: %v", s.Op, test.Op)
			}
		})
	}
}

func TestInvalid(t *testing.T) {
	var tests = []struct {
		name   string
		prefix []uint8
	}{
		{"dd", []uint8{0xdd}},
		{"ed", []uint8{0xed}},
		{"fd", []uint8{0xfd}},
		{"ddcb", []uint8{0xdd, 0xcb}},
		{"fdcb", []uint8{0xfd, 0xcb}},
	}

	for _, test := range tests {
		for opcode := 0; opcode < 0x100; opcode++ {
			if test.name == "dd" || test.name == "fd" {
				switch opcode {
				case 0xdd, 0xed, 0xfd, 0xcb:
					continue
				}
			}
			mock.ResetMemory()
			ptr := rcs.NewPointer(mock.TestMemory)
			ptr.Put(test.prefix[0])
			if len(test.prefix) > 1 {
				ptr.Put(test.prefix[1])
				ptr.Put(0) // displacement byte
			}
			ptr.Put(uint8(opcode))
			dasm := NewDisassembler(mock.TestMemory)
			s := dasm.NextStmt()
			if s.Op[0] == '?' {
				name := fmt.Sprintf("%v%02x", test.name, opcode)
				t.Run(name, func(t *testing.T) {
					want := fmt.Sprintf("?%s%02x", test.name, opcode)
					if s.Op != want {
						t.Errorf("\n have: %v \n want: %v", s.Op, want)
					}
				})
			}
		}
	}
}

func TestInvalidDD(t *testing.T) {
	for i := 0; i <= 0xff; i++ {
		name := fmt.Sprintf("%02x", i)
		t.Run(name, func(t *testing.T) {
			mock.ResetMemory()
			mem := mock.TestMemory
			mem.Write(0, 0xdd)
			mem.Write(1, uint8(i))
			dasm := NewDisassembler(mem)
			s := dasm.NextStmt()
			if s.Op[0] == '?' && i != 0xdd && i != 0xed && i != 0xfd && i != 0xcb {
				want := fmt.Sprintf("?dd%02x", i)
				if s.Op != want {
					t.Errorf("\n have: %v \n want: %v", s.Op, want)
				}
			}
		})
	}
}

func TestInvalidFD(t *testing.T) {
	for i := 0; i <= 0xff; i++ {
		name := fmt.Sprintf("%02x", i)
		t.Run(name, func(t *testing.T) {
			mock.ResetMemory()
			mem := mock.TestMemory
			mem.Write(0, 0xfd)
			mem.Write(1, uint8(i))
			dasm := NewDisassembler(mem)
			s := dasm.NextStmt()
			if s.Op[0] == '?' && i != 0xdd && i != 0xed && i != 0xfd && i != 0xcb {
				want := fmt.Sprintf("?fd%02x", i)
				if s.Op != want {
					t.Errorf("\n have: %v \n want: %v", s.Op, want)
				}
			}
		})
	}
}

func TestInvalidFDCB(t *testing.T) {
	for i := 0; i <= 0xff; i++ {
		name := fmt.Sprintf("%02x", i)
		t.Run(name, func(t *testing.T) {
			mock.ResetMemory()
			mem := mock.TestMemory
			mem.Write(0, 0xfd)
			mem.Write(1, 0xcb)
			mem.Write(2, 0)
			mem.Write(3, uint8(i))
			dasm := NewDisassembler(mem)
			s := dasm.NextStmt()
			if s.Op[0] == '?' {
				want := fmt.Sprintf("?fdcb%02x", i)
				if s.Op != want {
					t.Errorf("\n have: %v \n want: %v", s.Op, want)
				}
			}
		})
	}
}
