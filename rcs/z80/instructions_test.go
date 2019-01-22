package z80

import (
	"testing"

	"github.com/blackchip-org/retro-cs/mock"
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
			//setupPorts(cpu, fuseExpected[test.name])
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
			testMemory(t, cpu, fuseExpected[test.name])
			/*
				testHalt(t, cpu, fuseExpected[test.name])
				testPorts(t, cpu, fuseExpected[test.name])
			*/
		})
	}

}

func testMemory(t *testing.T, cpu *CPU, expected fuseTest) {
	diff, equal := mock.Verify(cpu.mem, expected.memory)
	if !equal {
		t.Fatalf("\nmemory mismatch (have, want): \n%v", diff.String())
	}
}

/*
func testHalt(t *testing.T, cpu *CPU, expected fuseTest) {
	WithFormat(t, "halt(%v)").Expect(cpu.Halt).ToBe(expected.Halt != 0)
}

func setupPorts(cpu *CPU, expected fuseTest) {
	cpu.Ports = newMockIO(expected.PortReads)
}

func testPorts(t *testing.T, cpu *CPU, expected fuseTest) {
	diff, ok := memory.Verify(cpu.Ports, expected.PortWrites)
	if !ok {
		t.Fatalf("\n write ports mismatch: \n%v", diff.String())
	}
}
*/

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

	for _, slice := range test.memory {
		mock.Import(mock.TestMemory, slice)
	}
	return cpu
}

/*
type mockIO struct {
	data map[uint8][]uint8
}

func newMockIO(snapshots []memory.Snapshot) memory.IO {
	mio := &mockIO{
		data: make(map[uint8][]uint8),
	}
	for _, snapshot := range snapshots {
		addr := uint8(snapshot.Address)
		stack, exists := mio.data[addr]
		if !exists {
			stack = make([]uint8, 0, 0)
		}
		stack = append(stack, snapshot.Values[0])
		mio.data[addr] = stack
	}
	return mio
}

func (m *mockIO) Load(addr uint16) uint8 {
	stack, exists := m.data[uint8(addr)]
	if !exists {
		return 0
	}
	if len(stack) == 0 {
		return 0
	}
	v := stack[0]
	stack = stack[1:]
	m.data[uint8(addr)] = stack
	return v
}

func (m *mockIO) Store(addr uint16, value uint8) {
	stack, exists := m.data[uint8(addr)]
	if !exists {
		stack = make([]uint8, 0, 0)
	}
	stack = append(stack, value)
	m.data[uint8(addr)] = stack
}

func (m *mockIO) Length() int {
	return 0
}

func (m *mockIO) Port(n int) *memory.Port {
	return &memory.Port{}
}

func (m *mockIO) Save(_ *state.Encoder) {}

func (m *mockIO) Restore(_ *state.Decoder) {}
*/

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

	memory     []mock.Slice
	portReads  []mock.Slice
	portWrites []mock.Slice
}
