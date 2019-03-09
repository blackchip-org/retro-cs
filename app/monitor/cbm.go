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
	// Is this a BASIC program?
	basic := true
	if len(args) == 2 {
		// TODO: "1" because of load "*", 8, 1 which is to use the address
		// found in the PRG header. If so, this probably isn't a BASIC
		// program. Is this actually correct?
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
	// Header is two bytes so there must be at least 2 bytes.
	if len(data) < 2 {
		return fmt.Errorf("invalid prg file: %v", filename)
	}
	mem := m.mon.mach.CPU["cpu"].Memory()
	// First two bytes are the memory location where the data should
	// be stored
	addr := int(data[0]) | (int(data[1]) << 8)
	end := addr + len(data) - 2 // minus the address bytes
	for i, d := range data[2:] {
		mem.Write(addr+i, d)
	}
	// Update pointers if this is a BASIC program
	if basic {
		// new start of variables is after the program
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

func (m *modC64) Silence() error {
	return nil
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
	if len(args) == 0 {
		return m.cmdInfo(args[0:])
	}
	switch args[0] {
	case "cr": // configuration register
		return valueFunc8(m.out, m.mmu.ReadCR, m.mmu.WriteCR, args[1:])
	case "info":
		return m.cmdInfo(args[1:])
	// case "mode":
	// 	return valueUint8(m.out, &m.mmu.Mode, args[1:])
	case "watch", "w":
		return m.cmdWatch(args[1:])
	}
	return fmt.Errorf("no such command: %v", args[0])
}

func (m *modC128MMU) cmdWatch(args []string) error {
	if err := checkLen(args, 1, 2); err != nil {
		return err
	}
	switch args[0] {
	case "cr": // configuration register
		return valueRW(m.out, &m.mmu.WatchCR, args[1:])
	case "lcr": // load configuration register
		return valueRW(m.out, &m.mmu.WatchLCR, args[1:])
	case "pcr": // pre-configuration register
		return valueRW(m.out, &m.mmu.WatchPCR, args[1:])
	case "all":
		return terminal(args[1:], func() error {
			m.mmu.WatchCR.W = true
			m.mmu.WatchCR.R = true
			m.mmu.WatchLCR.W = true
			m.mmu.WatchLCR.R = true
			m.mmu.WatchPCR.W = true
			m.mmu.WatchPCR.R = true
			return nil
		})
	case "none":
		return terminal(args[1:], m.Silence)
	}
	return fmt.Errorf("invalid argument: %v", args[0])
}

func (m *modC128MMU) cmdInfo(args []string) error {
	if err := checkLen(args, 0, 0); err != nil {
		return err
	}
	info := fmt.Sprintf(`
cr : %v
lcr: %v %v %v %v
pcr: %v %v %v %v
`,
		rcs.X8(uint8(m.mmu.Mem.Bank())),
		rcs.X8(m.mmu.LCR[0]),
		rcs.X8(m.mmu.LCR[1]),
		rcs.X8(m.mmu.LCR[2]),
		rcs.X8(m.mmu.LCR[3]),
		rcs.X8(m.mmu.PCR[0]),
		rcs.X8(m.mmu.PCR[1]),
		rcs.X8(m.mmu.PCR[2]),
		rcs.X8(m.mmu.PCR[3]),
	)
	m.out.Println(strings.TrimSpace(info))
	return nil
}

func (m *modC128MMU) AutoComplete() []readline.PrefixCompleterInterface {
	return []readline.PrefixCompleterInterface{
		readline.PcItem("cr"),
		// readline.PcItem("mode"),
		readline.PcItem("watch",
			readline.PcItem("cr", acRW...),
			readline.PcItem("lcr", acRW...),
			readline.PcItem("pcr", acRW...),
			readline.PcItem("all"),
			readline.PcItem("none"),
		),
	}
}

func (m *modC128MMU) Silence() error {
	m.mmu.WatchCR.W = false
	m.mmu.WatchCR.R = false
	m.mmu.WatchLCR.W = false
	m.mmu.WatchLCR.R = false
	m.mmu.WatchPCR.W = false
	m.mmu.WatchPCR.R = false
	return nil
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
	if len(args) == 0 {
		return m.info(args[0:])
	}
	switch args[0] {
	case "info":
		return m.info(args[1:])
	case "watch", "w":
		return m.cmdWatch(args[1:])
	}
	return fmt.Errorf("no such command: %v", args[0])
}

func (m *modC128VDC) cmdWatch(args []string) error {
	if err := checkLen(args, 1, 2); err != nil {
		return err
	}
	switch args[0] {
	case "address":
		return valueRW(m.out, &m.vdc.WatchAddr, args[1:])
	case "data":
		return valueRW(m.out, &m.vdc.WatchData, args[1:])
	case "all":
		return terminal(args[1:], func() error {
			m.vdc.WatchAddr.W = true
			m.vdc.WatchStatus.R = true
			m.vdc.WatchData.W = true
			m.vdc.WatchData.R = true
			return nil
		})
	case "none":
		return terminal(args[1:], m.Silence)
	}
	return fmt.Errorf("invalid argument: %v", args[0])
}

func (m *modC128VDC) info(args []string) error {
	if err := checkLen(args, 0, 0); err != nil {
		return err
	}
	format := strings.TrimSpace(`
addr  : %v
status: %v %v %v
mempos: %v
vss   : %v %v %v
			`)
	m.out.Printf(format,
		rcs.X8(m.vdc.Addr),
		m.vdc.Status, rcs.X8(m.vdc.Status), rcs.B8(m.vdc.Status),
		rcs.X16(m.vdc.MemPos),
		m.vdc.VSS, rcs.X8(m.vdc.VSS), rcs.B8(m.vdc.VSS),
	)
	return nil
}

func (m *modC128VDC) AutoComplete() []readline.PrefixCompleterInterface {
	return []readline.PrefixCompleterInterface{
		readline.PcItem("info"),
		readline.PcItem("watch",
			readline.PcItem("address", acRW...),
			readline.PcItem("data", acRW...),
			readline.PcItem("all"),
			readline.PcItem("none"),
		),
	}
}

func (m *modC128VDC) Silence() error {
	m.vdc.WatchAddr.W = false
	m.vdc.WatchStatus.R = false
	m.vdc.WatchData.W = false
	m.vdc.WatchData.R = false
	return nil
}
