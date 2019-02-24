package monitor

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/chzyer/readline"
)

type modC64 struct {
	mon *Monitor
}

func newModC64(mon *Monitor, comp rcs.Component) module {
	return &modC64{mon: mon}
}

func (m *modC64) Command(args []string) error {
	if err := checkLen(args, 1, maxArgs); err != nil {
		return err
	}
	switch args[0] {
	case "load-prg":
		return m.cmdLoadPrg(args[1:])
	}
	return fmt.Errorf("no such command: %v", args[0])
}

func (m *modC64) cmdLoadPrg(args []string) error {
	if err := checkLen(args, 1, 2); err != nil {
		return err
	}
	basic := true
	if len(args) == 2 {
		if strings.TrimSpace(args[1]) != "1" {
			return fmt.Errorf("invalid argument: %v", args[1])
		}
		basic = false
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
	mem := m.mon.mach.CPU["cpu"].Memory()
	addr := int(data[0]) | (int(data[1]) << 8)
	end := addr + len(data) - 2 // minus the address bytes
	for i, d := range data[2:] {
		mem.Write(addr+i, d)
	}
	if basic {
		// new start of variables is after the basic program
		vstart := end + 1
		mem.WriteLE(0x002d, vstart) // basic variable storage
		mem.WriteLE(0x002f, vstart) // basic array storage
	}
	return nil
}

func (m *modC64) AutoComplete() []readline.PrefixCompleterInterface {
	return []readline.PrefixCompleterInterface{
		readline.PcItem("load-prg",
			readline.PcItemDynamic(acDataFiles(m.mon, ".prg")),
		),
	}
}