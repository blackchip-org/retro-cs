package monitor

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/chzyer/readline"

	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/blackchip-org/retro-cs/rcs/m6502"
	"github.com/blackchip-org/retro-cs/rcs/z80"
)

type modCPU struct {
	name   string
	mon    *Monitor
	out    *log.Logger
	cpu    rcs.CPU
	mem    *rcs.Memory
	dasm   *rcs.Disassembler
	brkpts map[int]struct{}
}

func newModCPU(mon *Monitor, comp rcs.Component) module {
	c := comp.C.(rcs.CPU)
	var dasm *rcs.Disassembler
	cpud, ok := c.(rcs.CPUDisassembler)
	if ok {
		dasm = cpud.NewDisassembler()
	}
	mod := &modCPU{
		name:   comp.Name,
		mon:    mon,
		out:    mon.out,
		cpu:    c,
		mem:    c.Memory(),
		dasm:   dasm,
		brkpts: mon.mach.Breakpoints[comp.Name],
	}
	return mod
}

func (m *modCPU) Command(args []string) error {
	if len(args) == 0 {
		return m.cmdInfo(args[0:])
	}
	switch args[0] {
	case "breakpoint", "bp":
		return m.cmdBreakpoint(args[1:])
	case "disassemble", "d":
		return m.cmdDisassemble(args[1:])
	case "info", "i":
		return m.cmdInfo(args[1:])
	case "next", "n":
		return m.cmdNext(args[1:])
	case "step", "s":
		return m.cmdStep(args[1:])
	case "select":
		return m.cmdSelect(args[1:])
	case "trace", "t":
		return m.cmdTrace(args[1:])
	}
	return fmt.Errorf("no such command: %v", args[0])
}

func (m *modCPU) cmdBreakpoint(args []string) error {
	if len(args) == 0 {
		return m.cmdBreakpointList(args[0:])
	}
	switch args[0] {
	case "list":
		return m.cmdBreakpointList(args[1:])
	case "none":
		return m.cmdBreakpointNone(args[1:])
	case "address":
		return m.cmdBreakpointSwitch(args[1:])
	}
	return m.cmdBreakpointSwitch(args[0:])
}

func (m *modCPU) cmdBreakpointClear(args []string) error {
	if err := checkLen(args, 1, 1); err != nil {
		return err
	}
	addr, err := parseAddress(m.mem, args[0])
	if err != nil {
		return err
	}
	delete(m.brkpts, addr)
	return nil
}

func (m *modCPU) cmdBreakpointList(args []string) error {
	if err := checkLen(args, 0, 0); err != nil {
		return err
	}
	if len(m.brkpts) == 0 {
		return nil
	}
	addrs := make([]string, 0, 0)
	for k := range m.brkpts {
		// FIXME: Hard-coded format
		addrs = append(addrs, fmt.Sprintf("%v$%04x", m.prefix(), k))
	}
	sort.Strings(addrs)
	m.mon.out.Printf(strings.Join(addrs, "\n"))
	return nil
}

func (m *modCPU) cmdBreakpointNone(args []string) error {
	if err := checkLen(args, 0, 0); err != nil {
		return err
	}
	for k := range m.brkpts {
		delete(m.brkpts, k)
	}
	return nil
}

func (m *modCPU) cmdBreakpointSwitch(args []string) error {
	if err := checkLen(args, 1, 2); err != nil {
		return err
	}
	addr, err := parseAddress(m.mem, args[0])
	if err != nil {
		return fmt.Errorf("invalid address: %v", args[0])
	}
	if len(args) == 1 {
		if _, ok := m.brkpts[addr]; ok {
			m.out.Println("on")
		} else {
			m.out.Println("off")
		}
		return nil
	}
	switch args[1] {
	case "on":
		m.brkpts[addr] = struct{}{}
		return nil
	case "off":
		delete(m.brkpts, addr)
		return nil
	}
	return fmt.Errorf("invalid argument: %v", args[0])
}

func (m *modCPU) cmdDisassemble(args []string) error {
	if err := checkLen(args, 0, 2); err != nil {
		return err
	}
	if m.dasm == nil {
		return fmt.Errorf("cannot disassemble this processor")
	}
	if len(args) > 0 {
		addr, err := parseAddress(m.mem, args[0])
		if err != nil {
			return err
		}
		m.dasm.SetPC(addr)
	}
	if len(args) > 1 {
		// list until at ending address
		addrEnd, err := parseAddress(m.mem, args[1])
		if err != nil {
			return err
		}
		for m.dasm.PC() <= addrEnd {
			m.mon.out.Printf("%v%v\n", m.prefix(), m.dasm.Next())
		}
	} else {
		// list number of lines
		lines := m.mon.dasmLines
		if lines == 0 {
			_, h, err := readline.GetSize(0)
			if err != nil {
				return err
			}
			lines = h - 1
			if lines <= 0 {
				lines = 1
			}
		}
		for i := 0; i < lines; i++ {
			m.mon.out.Printf("%v%v\n", m.prefix(), m.dasm.Next())
		}
	}
	// m.lastCmd = m.cmdDasmList
	return nil
}

func (m *modCPU) cmdInfo(args []string) error {
	if err := checkLen(args, 0, 0); err != nil {
		return err
	}
	name := ""
	if m.name != "cpu" {
		name = m.name + ":"
	}
	m.mon.out.Printf("[%v%v]\n", name, m.mon.mach.Status)
	m.mon.out.Printf("%v", m.cpu)
	return nil
}

func (m *modCPU) cmdNext(args []string) error {
	if err := checkLen(args, 0, 0); err != nil {
		return err
	}
	ppc := m.dasm.PC()
	m.dasm.SetPC(m.cpu.PC() + m.cpu.Offset())
	m.mon.out.Println(m.dasm.Next())
	m.dasm.SetPC(ppc)
	return nil
}

func (m *modCPU) cmdSelect(args []string) error {
	if err := checkLen(args, 0, 0); err != nil {
		return err
	}
	m.mon.sc = m.name
	m.mon.rl.SetPrompt(m.mon.getPrompt())
	return nil
}

func (m *modCPU) cmdStep(args []string) error {
	if err := checkLen(args, 0, 0); err != nil {
		return err
	}
	m.cpu.Next()
	ppc := m.dasm.PC()
	m.dasm.SetPC(m.cpu.PC() + m.cpu.Offset())
	m.mon.out.Println(m.dasm.Next())
	m.dasm.SetPC(ppc)
	m.mon.defaultCmd = fmt.Sprintf("%v step", m.name)
	return nil
}

func (m *modCPU) cmdTrace(args []string) error {
	if err := checkLen(args, 0, 1); err != nil {
		return err
	}
	if len(args) == 0 {
		m.mon.mach.Command(rcs.MachTrace, m.name)
		return nil
	}
	v, err := parseBool(args[0])
	if err != nil {
		return err
	}
	m.mon.mach.Command(rcs.MachTrace, m.name, v)
	return nil
}

func (m *modCPU) AutoComplete() []readline.PrefixCompleterInterface {
	return []readline.PrefixCompleterInterface{
		readline.PcItem("breakpoint",
			readline.PcItem("list"),
			readline.PcItem("none"),
			readline.PcItem("address"),
		),
		readline.PcItem("disassemble"),
		readline.PcItem("info"),
		readline.PcItem("next"),
		readline.PcItem("step"),
		readline.PcItem("select"),
		readline.PcItem("trace"),
	}
}

func (m *modCPU) Silence() error {
	m.cmdTrace([]string{"off"})
	return nil
}

func (m *modCPU) prefix() string {
	if m.name == "cpu" {
		return ""
	}
	return m.name + "  "
}

type modM6502 struct {
	parent module
	mon    *Monitor
	out    *log.Logger
	cpu    *m6502.CPU
}

func newModM6502(mon *Monitor, comp rcs.Component) module {
	cpu := comp.C.(*m6502.CPU)
	return &modM6502{
		parent: newModCPU(mon, comp),
		mon:    mon,
		out:    mon.out,
		cpu:    cpu,
	}
}

func (m *modM6502) Command(args []string) error {
	if len(args) == 0 {
		return m.parent.Command(args)
	}

	switch args[0] {
	case "r.pc":
		return valueIntF(m.out, m.cpu.PC, m.cpu.SetPC, args[1:])
	case "r.a":
		return valueUint8(m.out, &m.cpu.A, args[1:])
	case "r.x":
		return valueUint8(m.out, &m.cpu.X, args[1:])
	case "r.y":
		return valueUint8(m.out, &m.cpu.Y, args[1:])
	case "r.sp":
		return valueUint8(m.out, &m.cpu.SP, args[1:])
	case "r.sr":
		return valueUint8(m.out, &m.cpu.SR, args[1:])
	case "f.c":
		return valueBit(m.out, &m.cpu.SR, m6502.FlagC, args[1:])
	case "f.z":
		return valueBit(m.out, &m.cpu.SR, m6502.FlagZ, args[1:])
	case "f.i":
		return valueBit(m.out, &m.cpu.SR, m6502.FlagI, args[1:])
	case "f.d":
		return valueBit(m.out, &m.cpu.SR, m6502.FlagD, args[1:])
	case "f.b":
		return valueBit(m.out, &m.cpu.SR, m6502.FlagB, args[1:])
	case "f.v":
		return valueBit(m.out, &m.cpu.SR, m6502.FlagV, args[1:])
	case "f.n":
		return valueBit(m.out, &m.cpu.SR, m6502.FlagN, args[1:])

	case "watch-brk":
		return valueBool(m.out, &m.cpu.WatchBRK, args[1:])
	case "watch-irq":
		return valueBool(m.out, &m.cpu.WatchIRQ, args[1:])
	case "watch-stack":
		return valueBool(m.out, &m.cpu.WatchStack, args[1:])
	}
	return m.parent.Command(args)
}

func (m *modM6502) AutoComplete() []readline.PrefixCompleterInterface {
	cmd := m.parent.AutoComplete()
	cmd = append(cmd, []readline.PrefixCompleterInterface{
		readline.PcItem("r.a"),
		readline.PcItem("r.x"),
		readline.PcItem("r.y"),
		readline.PcItem("r.sp"),
		readline.PcItem("r.sr"),
		readline.PcItem("f.c"),
		readline.PcItem("f.z"),
		readline.PcItem("f.i"),
		readline.PcItem("f.d"),
		readline.PcItem("f.b"),
		readline.PcItem("f.v"),
		readline.PcItem("f.n"),

		readline.PcItem("watch-brk"),
		readline.PcItem("watch-irq"),
	}...)
	sort.Sort(byName(cmd))
	return cmd
}

func (m *modM6502) Silence() error {
	m.parent.Silence()
	m.cpu.WatchBRK = false
	m.cpu.WatchIRQ = false
	m.cpu.WatchStack = false
	return nil
}

type modZ80 struct {
	parent module
	mon    *Monitor
	out    *log.Logger
	cpu    *z80.CPU
}

func newModZ80(mon *Monitor, comp rcs.Component) module {
	cpu := comp.C.(*z80.CPU)
	return &modZ80{
		parent: newModCPU(mon, comp),
		mon:    mon,
		out:    mon.out,
		cpu:    cpu,
	}
}

func (m *modZ80) Command(args []string) error {
	if len(args) == 0 {
		return m.parent.Command(args)
	}

	switch args[0] {
	case "f.c":
		return valueBit(m.out, &m.cpu.F, z80.FlagC, args[1:])
	case "f.n":
		return valueBit(m.out, &m.cpu.F, z80.FlagN, args[1:])
	case "f.v":
		return valueBit(m.out, &m.cpu.F, z80.FlagV, args[1:])
	case "f.p":
		return valueBit(m.out, &m.cpu.F, z80.FlagP, args[1:])
	case "f.3":
		return valueBit(m.out, &m.cpu.F, z80.Flag3, args[1:])
	case "f.h":
		return valueBit(m.out, &m.cpu.F, z80.FlagH, args[1:])
	case "f.5":
		return valueBit(m.out, &m.cpu.F, z80.Flag5, args[1:])
	case "f.z":
		return valueBit(m.out, &m.cpu.F, z80.FlagZ, args[1:])
	case "f.s":
		return valueBit(m.out, &m.cpu.F, z80.FlagS, args[1:])

	case "r.pc":
		return valueIntF(m.out, m.cpu.PC, m.cpu.SetPC, args[1:])
	case "r.a":
		return valueUint8(m.out, &m.cpu.A, args[1:])
	case "r.f":
		return valueUint8(m.out, &m.cpu.F, args[1:])
	case "r.b":
		return valueUint8(m.out, &m.cpu.B, args[1:])
	case "r.c":
		return valueUint8(m.out, &m.cpu.C, args[1:])
	case "r.d":
		return valueUint8(m.out, &m.cpu.D, args[1:])
	case "r.e":
		return valueUint8(m.out, &m.cpu.E, args[1:])
	case "r.h":
		return valueUint8(m.out, &m.cpu.H, args[1:])
	case "r.l":
		return valueUint8(m.out, &m.cpu.L, args[1:])

	case "r.a1":
		return valueUint8(m.out, &m.cpu.A1, args[1:])
	case "r.f1":
		return valueUint8(m.out, &m.cpu.F1, args[1:])
	case "r.b1":
		return valueUint8(m.out, &m.cpu.B1, args[1:])
	case "r.c1":
		return valueUint8(m.out, &m.cpu.C1, args[1:])
	case "r.d1":
		return valueUint8(m.out, &m.cpu.D1, args[1:])
	case "r.e1":
		return valueUint8(m.out, &m.cpu.E1, args[1:])
	case "r.h1":
		return valueUint8(m.out, &m.cpu.H1, args[1:])
	case "r.l1":
		return valueUint8(m.out, &m.cpu.L1, args[1:])

	case "r.af":
		return valueUint16HL(m.out, &m.cpu.A, &m.cpu.F, args[1:])
	case "r.bc":
		return valueUint16HL(m.out, &m.cpu.B, &m.cpu.C, args[1:])
	case "r.de":
		return valueUint16HL(m.out, &m.cpu.D, &m.cpu.E, args[1:])
	case "r.hl":
		return valueUint16HL(m.out, &m.cpu.H, &m.cpu.L, args[1:])

	case "r.af1":
		return valueUint16HL(m.out, &m.cpu.A1, &m.cpu.F1, args[1:])
	case "r.bc1":
		return valueUint16HL(m.out, &m.cpu.B1, &m.cpu.C1, args[1:])
	case "r.de1":
		return valueUint16HL(m.out, &m.cpu.D1, &m.cpu.E1, args[1:])
	case "r.hl1":
		return valueUint16HL(m.out, &m.cpu.H1, &m.cpu.L1, args[1:])

	case "r.i":
		return valueUint8(m.out, &m.cpu.I, args[1:])
	case "r.r":
		return valueUint8(m.out, &m.cpu.R, args[1:])
	case "r.ixh":
		return valueUint8(m.out, &m.cpu.IXH, args[1:])
	case "r.ixl":
		return valueUint8(m.out, &m.cpu.IXL, args[1:])
	case "r.iyh":
		return valueUint8(m.out, &m.cpu.IYH, args[1:])
	case "r.iyl":
		return valueUint8(m.out, &m.cpu.IYL, args[1:])
	case "r.sp":
		return valueUint16(m.out, &m.cpu.SP, args[1:])

	case "r.ix":
		return valueUint16HL(m.out, &m.cpu.IXH, &m.cpu.IXL, args[1:])
	case "r.iy":
		return valueUint16HL(m.out, &m.cpu.IYH, &m.cpu.IYL, args[1:])

	case "r.iff1":
		return valueBool(m.out, &m.cpu.IFF1, args[1:])
	case "r.iff2":
		return valueBool(m.out, &m.cpu.IFF2, args[1:])
	case "r.im":
		return valueUint8(m.out, &m.cpu.IM, args[1:])

	case "watch-irq":
		return valueBool(m.out, &m.cpu.WatchIRQ, args[1:])

	}
	return m.parent.Command(args)
}

func (m *modZ80) AutoComplete() []readline.PrefixCompleterInterface {
	cmd := m.parent.AutoComplete()
	cmd = append(cmd, []readline.PrefixCompleterInterface{
		readline.PcItem("r.pc"),
		readline.PcItem("r.a"),
		readline.PcItem("r.f"),
		readline.PcItem("r.b"),
		readline.PcItem("r.c"),
		readline.PcItem("r.d"),
		readline.PcItem("r.e"),
		readline.PcItem("r.h"),
		readline.PcItem("r.l"),

		readline.PcItem("r.a1"),
		readline.PcItem("r.f1"),
		readline.PcItem("r.b1"),
		readline.PcItem("r.c1"),
		readline.PcItem("r.d1"),
		readline.PcItem("r.e1"),
		readline.PcItem("r.h1"),
		readline.PcItem("r.l1"),

		readline.PcItem("r.af"),
		readline.PcItem("r.bc"),
		readline.PcItem("r.de"),
		readline.PcItem("r.hl"),

		readline.PcItem("r.af1"),
		readline.PcItem("r.bc1"),
		readline.PcItem("r.de1"),
		readline.PcItem("r.hl1"),

		readline.PcItem("r.i"),
		readline.PcItem("r.r"),
		readline.PcItem("r.ixh"),
		readline.PcItem("r.ixl"),
		readline.PcItem("r.iyh"),
		readline.PcItem("r.iyl"),
		readline.PcItem("r.sp"),

		readline.PcItem("r.ix"),
		readline.PcItem("r.iy"),

		readline.PcItem("r.iff1"),
		readline.PcItem("r.iff2"),
		readline.PcItem("r.im"),

		readline.PcItem("f.c"),
		readline.PcItem("f.n"),
		readline.PcItem("f.v"),
		readline.PcItem("f.p"),
		readline.PcItem("f.3"),
		readline.PcItem("f.h"),
		readline.PcItem("f.5"),
		readline.PcItem("f.z"),
		readline.PcItem("f.s"),

		readline.PcItem("watch-irq"),
	}...)
	sort.Sort(byName(cmd))
	return cmd
}

func (m *modZ80) Silence() error {
	m.parent.Silence()
	m.cpu.WatchIRQ = false
	return nil
}
