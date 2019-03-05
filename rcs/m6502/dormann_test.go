// +build ext

package m6502

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"testing"

	"github.com/blackchip-org/retro-cs/config"
	"github.com/blackchip-org/retro-cs/rcs"
)

var (
	sourceDir = filepath.Join(config.ResourceDir(), "ext", "m6502")
)

func TestDormann(t *testing.T) {
	mem := rcs.NewMemory(1, 0x10000)
	ram := make([]uint8, 0x10000, 0x10000)
	code, err := ioutil.ReadFile(filepath.Join(sourceDir, "6502_functional_test.bin"))
	if err != nil {
		t.Fatalf("unable to load test runner: %v", err)
	}

	mem.MapRAM(0x0, ram)
	mem.MapRAM(0x0, code)

	cpu := New(mem)
	cpu.SetPC(0x03ff)
	log.SetFlags(0)
	cpu.WatchBRK = true
	dasm := cpu.NewDisassembler()
	for {
		here := cpu.PC() + cpu.Offset()
		// Success points
		if here == 0x346c || here == 0x3469 {
			break
		}
		dasm.SetPC(here)
		ppc := cpu.PC()
		cpu.Next()
		// If the PC hasn't moved, its a trap
		if ppc == cpu.PC() {
			t.Fatalf("\n[trap]\n%v", cpu)
		}
	}
}
