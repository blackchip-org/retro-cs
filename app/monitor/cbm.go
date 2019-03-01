package monitor

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/blackchip-org/retro-cs/system/c128"
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

type modC128MMU struct {
	mon *Monitor
	out *log.Logger
	mmu *c128.MMU
}

func newModC128MMU(mon *Monitor, comp rcs.Component) module {
	return &modC128MMU{
		mon: mon,
		out: mon.out,
		mmu: comp.C.(*c128.MMU),
	}
}

func (m *modC128MMU) Command(args []string) error {
	if err := checkLen(args, 1, maxArgs); err != nil {
		return err
	}
	switch args[0] {
	case "mode":
		return valueUint8(m.out, &m.mmu.Mode, args[1:])
	case "watch-cr-write":
		return valueBool(m.out, &m.mmu.WatchCR.Write, args[1:])
	case "watch-cr-read":
		return valueBool(m.out, &m.mmu.WatchCR.Read, args[1:])
	case "watch-lcr-write":
		return valueBool(m.out, &m.mmu.WatchLCR.Write, args[1:])
	case "watch-lcr-read":
		return valueBool(m.out, &m.mmu.WatchLCR.Read, args[1:])
	case "watch-pcr-write":
		return valueBool(m.out, &m.mmu.WatchPCR.Write, args[1:])
	case "watch-pcr-read":
		return valueBool(m.out, &m.mmu.WatchPCR.Read, args[1:])
	case "watch-all":
		return terminal(args[1:], func() error {
			m.mmu.WatchCR.Write = true
			m.mmu.WatchCR.Read = true
			m.mmu.WatchLCR.Write = true
			m.mmu.WatchLCR.Read = true
			m.mmu.WatchPCR.Write = true
			m.mmu.WatchPCR.Read = true
			return nil
		})
	case "watch-none":
		return terminal(args[1:], func() error {
			m.mmu.WatchCR.Write = false
			m.mmu.WatchCR.Read = false
			m.mmu.WatchLCR.Write = false
			m.mmu.WatchLCR.Read = false
			m.mmu.WatchPCR.Write = false
			m.mmu.WatchPCR.Read = false
			return nil
		})
	}
	return fmt.Errorf("no such command: %v", args[0])
}

func (m *modC128MMU) AutoComplete() []readline.PrefixCompleterInterface {
	return []readline.PrefixCompleterInterface{
		readline.PcItem("watch-cr-write"),
		readline.PcItem("watch-cr-read"),
		readline.PcItem("watch-lcr-write"),
		readline.PcItem("watch-lcr-read"),
		readline.PcItem("watch-pcr-write"),
		readline.PcItem("watch-pcr-read"),
		readline.PcItem("watch-all"),
		readline.PcItem("watch-none"),
	}
}

type modC128VDC struct {
	mon *Monitor
	out *log.Logger
	vdc *c128.VDC
}

func newModC128VDC(mon *Monitor, comp rcs.Component) module {
	return &modC128VDC{
		mon: mon,
		out: mon.out,
		vdc: comp.C.(*c128.VDC),
	}
}

func (m *modC128VDC) Command(args []string) error {
	if err := checkLen(args, 1, maxArgs); err != nil {
		return err
	}
	switch args[0] {
	case "info":
		return m.info(args[1:])
	case "watch-address-write":
		return valueBool(m.out, &m.vdc.WatchAddr.Write, args[1:])
	case "watch-status-read":
		return valueBool(m.out, &m.vdc.WatchStatus.Read, args[1:])
	case "watch-data-write":
		return valueBool(m.out, &m.vdc.WatchData.Write, args[1:])
	case "watch-data-read":
		return valueBool(m.out, &m.vdc.WatchData.Read, args[1:])
	case "watch-all":
		return terminal(args[1:], func() error {
			m.vdc.WatchAddr.Write = true
			m.vdc.WatchStatus.Read = true
			m.vdc.WatchData.Write = true
			m.vdc.WatchData.Read = true
			return nil
		})
	case "watch-none":
		return terminal(args[1:], func() error {
			m.vdc.WatchAddr.Write = false
			m.vdc.WatchStatus.Read = false
			m.vdc.WatchData.Write = false
			m.vdc.WatchData.Read = false
			return nil
		})
	}
	return fmt.Errorf("no such command: %v", args[0])
}

func (m *modC128VDC) info(args []string) error {
	if err := checkLen(args, 0, 0); err != nil {
		return err
	}
	format := strings.TrimSpace(`
addr  : %02x
status: %v
data  : %02x
	`)
	m.out.Println(format, m.vdc.Addr, formatValue(int(m.vdc.Status)), m.vdc.Data)
	return nil
}

func (m *modC128VDC) AutoComplete() []readline.PrefixCompleterInterface {
	return []readline.PrefixCompleterInterface{
		readline.PcItem("info"),
		readline.PcItem("watch-address-write"),
		readline.PcItem("watch-status-read"),
		readline.PcItem("watch-data-write"),
		readline.PcItem("watch-data-read"),
		readline.PcItem("watch-all"),
		readline.PcItem("watch-none"),
	}
}
