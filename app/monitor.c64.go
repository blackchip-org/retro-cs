package app

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func (m *Monitor) cmdLoadPrg(args []string) error {
	if err := checkLen(args, 1, 1); err != nil {
		return err
	}
	filename := loadPath(args[0])
	if !strings.HasSuffix(filename, ".prg") {
		filename += ".prg"
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	if len(data) < 2 {
		return fmt.Errorf("invalid prg file: %v", filename)
	}
	mem := m.mach.Mem[0]
	addr := int(data[0]) | (int(data[1]) << 8)
	end := addr + len(data) - 2 // minus the address bytes
	for i, d := range data[2:] {
		mem.Write(addr+i, d)
	}
	// new start of variables is after the basic program
	vstart := end + 1
	mem.WriteLE(0x002d, vstart) // basic variable storage
	mem.WriteLE(0x002f, vstart) // basic array storage
	return nil
}
