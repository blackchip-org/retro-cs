// +build ext

package z80

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"testing"

	"github.com/blackchip-org/retro-cs/mock"
	"github.com/blackchip-org/retro-cs/rcs"
)

var (
	root      = filepath.Join("..", "..")
	sourceDir = filepath.Join(root, "ext", "zex")
)

var zexdocTests = []string{
	"adc16",
	"add16",
	"add16x",
	"add16y",
	"alu8i",
	"alu8r",
	"alu8rx",
	"alu8x",
	"bitx",
	"bitz80",
	"cpd1",
	"cpi1",
	"daa",
	"inca",
	"incb",
	"incbc",
	"incc",
	"incd",
	"incde",
	"ince",
	"inch",
	"inchl",
	"incix",
	"inciy",
	"incl",
	"incm",
	"incsp",
	"incx",
	"incxh",
	"incxl",
	"incyh",
	"incyl",
	"ld161",
	"ld162",
	"ld163",
	"ld164",
	"ld165",
	"ld166",
	"ld167",
	"ld168",
	"ld16im",
	"ld16ix",
	"ld8bd",
	"ld8im",
	"ld8imx",
	"ld8ix1",
	"ld8ix2",
	"ld8ix3",
	"ld8ixy",
	"ld8rr",
	"ld8rrx",
	"lda",
	"ldd1",
	"ldd2",
	"ldi1",
	"ldi2",
	"neg",
	"rld",
	"rot8080",
	"rotxy",
	"rotz80",
	"srz80",
	"srzx",
	"st8ix1",
	"st8ix2",
	"st8ix3",
	"stabd",
}

var zexdoc []byte

func init() {
	var err error
	zexdocFile := filepath.Join(sourceDir, "zexdoc.com")
	zexdoc, err = ioutil.ReadFile(zexdocFile)
	if err != nil {
		log.Panicf("unable to read %v: %v", zexdocFile, err)
	}
}

// Running the full zexdoc can take more than 10 minutes. This test instead
// breaks up each test into an individual run. The HL register is loaded with
// the address of the test and the program counter is set to the beginning
// of the normal test loop. Execution is stopped when the program counter
// returns to the top of the loop. Output is then checked for "ERROR" to
// determine if the test passes or fails.
const loopStart = 0x0122

func TestZexdoc(t *testing.T) {
	testBaseAddr := uint16(0x013a)
	for i, test := range zexdocTests {
		addr := testBaseAddr + (uint16(i) * 2)
		t.Run(test, func(t *testing.T) {
			runner := newRunner(zexdoc, addr)
			passed := runner.Run()
			if !passed {
				t.Fail()
			}
		})
	}
}

func BenchmarkZexdoc(b *testing.B) {
	testBaseAddr := uint16(0x013a)
	for i, test := range zexdocTests {
		addr := testBaseAddr + (uint16(i) * 2)
		runner := newRunner(zexdoc, addr)
		b.Run(test, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				runner.Next()
				if runner.Done() {
					runner.Reset()
				}
			}
		})
	}
}

type zexRunner struct {
	mem      *rcs.Memory
	cpu      *CPU
	out      bytes.Buffer
	testAddr uint16
}

func newRunner(code []byte, addr uint16) *zexRunner {
	// Follow the notes at:
	// https://floooh.github.io/2016/07/12/z80-rust-ms1.html
	mock.ResetMemory()
	// Import the code
	for i, b := range code {
		mock.TestMemory.Write(0x100+i, b)
	}
	c := New(mock.TestMemory)
	zr := &zexRunner{mem: mock.TestMemory, cpu: c, testAddr: addr}
	zr.Reset()
	return zr
}

func (z *zexRunner) Next() {
	z.cpu.Next()
}

func (z *zexRunner) Syscall() {
	// System call that outputs to the screen
	if z.cpu.PC() == 0x0005 {
		// Single character out
		if z.cpu.C == 0x02 {
			msg := fmt.Sprintf("%c", rune(z.cpu.C))
			z.out.WriteString(msg)
			fmt.Print(msg)
		}
		// String out, terminated by $
		if z.cpu.C == 0x09 {
			addr := int(z.cpu.D)<<8 | int(z.cpu.E)
			for {
				ch := rune(z.mem.Read(addr))
				if ch == '$' {
					break
				}
				msg := fmt.Sprintf("%c", ch)
				z.out.WriteString(msg)
				fmt.Print(msg)
				addr++
			}
		}
		// Return from subroutine
		z.cpu.SetPC(z.mem.ReadLE(int(z.cpu.SP)))
		z.cpu.SP += 2
	}
}

func (z *zexRunner) Done() bool {
	return z.cpu.PC() == loopStart
}

func (z *zexRunner) Passed() bool {
	return !strings.Contains(z.out.String(), "ERROR")
}

func (z *zexRunner) Reset() {
	z.cpu.SP = 0xf000
	// HL register is loaded with the address of the test to run
	z.cpu.H, z.cpu.L = uint8(z.testAddr>>8), uint8(z.testAddr)
	z.cpu.SetPC(loopStart)
}

func (z *zexRunner) Run() bool {
	for {
		z.Next()
		z.Syscall()
		if z.Done() {
			break
		}
	}
	return z.Passed()
}
