package z80

import (
	"testing"

	"github.com/blackchip-org/retro-cs/mock"
	"github.com/blackchip-org/retro-cs/rcs"
)

// Set a test name here to test a single test
var testSingle = ""

// TODO: Write single tests for:
// ADC/SBC: Check that both bytes are zero for zero flag when doing 16-bits

func TestOps(t *testing.T) {
	for _, test := range fuseIn {
		if testSingle != "" && test.name != testSingle {
			continue
		}
		t.Run(test.name, func(t *testing.T) {
			cpu := load(test)
			i := 0
			setupPorts(cpu, fuseExpected[test.name])
			for {
				if ok := cpu.Next(); !ok {
					t.Skip("unimplemented")
				}
				if test.name == "dd00" {
					if cpu.PC() == 0x0003 {
						break
					}
				} else if test.name == "ddfd00" {
					if cpu.PC() == 0x0004 {
						break
					}
				} else {
					if cpu.mem.Read(cpu.PC()) == 0 && cpu.PC() != 0 {
						break
					}
					if test.tstates == 1 {
						break
					}
				}
				if i > 100 {
					t.Fatalf("exceeded execution limit")
				}
				i++
			}
			expected := load(fuseExpected[test.name])

			if cpu.String() != expected.String() {
				t.Errorf("\n have: \n%v \n want: \n%v", cpu.String(), expected.String())
			}
			testMemory(t, cpu.mem, fuseExpected[test.name].memory)
			testMemory(t, cpu.Ports, fuseExpected[test.name].portWrites)
			testHalt(t, cpu, fuseExpected[test.name])
		})
	}
}

func testMemory(t *testing.T, mem *rcs.Memory, expected [][]int) {
	for _, av := range expected {
		addr := av[0]
		value := uint8(av[1])
		have := mem.Read(addr)
		if have != value {
			t.Errorf("\n addr %04x have: %02x \n addr %04x want: %02x", addr, have, addr, value)
		}
	}
}

func testHalt(t *testing.T, cpu *CPU, expected fuseTest) {
	if cpu.Halt != (expected.halt != 0) {
		t.Errorf("\n want: halt(%v) \n have: halt(%v)", cpu.Halt, expected.halt)
	}
}

func setupPorts(cpu *CPU, expected fuseTest) {
	for _, avs := range expected.portReads {
		addr := avs[0]
		values := avs[1:]
		cpu.Ports.MapLoad(addr, mock.MockRead(values))
	}
}

func load(test fuseTest) *CPU {
	mock.ResetMemory()
	cpu := New(mock.TestMemory)

	cpu.A, cpu.F = uint8(test.af>>8), uint8(test.af)
	cpu.B, cpu.C = uint8(test.bc>>8), uint8(test.bc)
	cpu.D, cpu.E = uint8(test.de>>8), uint8(test.de)
	cpu.H, cpu.L = uint8(test.hl>>8), uint8(test.hl)

	cpu.A1, cpu.F1 = uint8(test.af1>>8), uint8(test.af1)
	cpu.B1, cpu.C1 = uint8(test.bc1>>8), uint8(test.bc1)
	cpu.D1, cpu.E1 = uint8(test.de1>>8), uint8(test.de1)
	cpu.H1, cpu.L1 = uint8(test.hl1>>8), uint8(test.hl1)

	cpu.IXH, cpu.IXL = uint8(test.ix>>8), uint8(test.ix)
	cpu.IYH, cpu.IYL = uint8(test.iy>>8), uint8(test.iy)

	cpu.SP = test.sp
	cpu.SetPC(int(test.pc))
	cpu.I = test.i
	cpu.R = test.r
	cpu.IFF1 = test.iff1 != 0
	cpu.IFF2 = test.iff2 != 0
	cpu.IM = uint8(test.im)

	for _, av := range test.memory {
		addr := av[0]
		value := uint8(av[1])
		mock.TestMemory.Write(addr, value)
	}

	return cpu
}

type fuseTest struct {
	name    string
	af      uint16
	bc      uint16
	de      uint16
	hl      uint16
	af1     uint16
	bc1     uint16
	de1     uint16
	hl1     uint16
	ix      uint16
	iy      uint16
	sp      uint16
	pc      uint16
	i       uint8
	r       uint8
	iff1    int
	iff2    int
	im      int
	halt    int
	tstates int

	memory     [][]int
	portReads  [][]int
	portWrites [][]int
}
