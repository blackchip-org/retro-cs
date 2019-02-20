package monitor

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/blackchip-org/retro-cs/config"
	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/chzyer/readline"
)

func (m *Monitor) cmd(args []string) error {
	switch args[0] {
	case "b", "break":
		return m.cmdBreak(args[1:])
	case "cpu":
		return m.cmdCPU(args[1:])
	case "d":
		return m.cmdDasmList(args[1:])
	case "dasm":
		return m.cmdDasm(args[1:])
	case "export":
		return m.cmdExport(args[1:])
	case "g", "go":
		return m.cmdGo(args[1:])
	case "import":
		return m.cmdImport(args[1:])
	case "m":
		return m.cmdMemoryDump(args[1:])
	case "mem":
		return m.cmdMemory(args[1:])
	case "n", "next":
		return m.cmdNext(args[1:])
	case "p", "pause":
		return m.cmdPause(args[1:])
	case "poke":
		return m.cmdPoke(args[1:])
	case "peek":
		return m.cmdPeek(args[1:])
	case "q", "quit":
		return m.cmdQuit(args[1:])
	case "r":
		return m.cmdCPU([]string{})
	case "s", "step":
		return m.cmdStep(args[1:])
	case "t", "trace":
		return m.cmdTrace(args[1:])
	case "w", "watch":
		return m.cmdWatch(args[1:])
	case "x":
		return m.cmdX()
	case "_yield":
		return m.cmdYield()
	}

	if config.System == "c64" {
		switch args[0] {
		case "load-prg":
			return m.cmdLoadPrg(args[1:])
		}
	}

	val, err := m.parseValue(args[0])
	if err == nil {
		m.out.Print(m.formatValue(val))
		return nil
	}

	return fmt.Errorf("no such command: %v", args[0])
}

func (m *Monitor) cmdBreak(args []string) error {
	if err := checkLen(args, 0, maxArgs); err != nil {
		return err
	}
	if len(args) == 0 {
		return m.cmdBreakList()
	}
	switch args[0] {
	case "clear":
		return m.cmdBreakClear(args[1:])
	case "clear-all":
		return m.cmdBreakClearAll()
	case "list":
		return m.cmdBreakList()
	case "set":
		return m.cmdBreakSet(args[1:])
	}
	return fmt.Errorf("no such command: %v", args[0])
}

func (m *Monitor) cmdBreakClear(args []string) error {
	if err := checkLen(args, 1, 1); err != nil {
		return err
	}
	addr, err := m.parseAddress(args[0])
	if err != nil {
		return err
	}
	delete(m.sc.brkpts, addr)
	return nil
}

func (m *Monitor) cmdBreakClearAll() error {
	for k := range m.sc.brkpts {
		delete(m.sc.brkpts, k)
	}
	return nil
}

func (m *Monitor) cmdBreakList() error {
	if len(m.sc.brkpts) == 0 {
		return nil
	}
	addrs := make([]string, 0, 0)
	for k := range m.sc.brkpts {
		// FIXME: Hard-coded format
		addrs = append(addrs, fmt.Sprintf("$%04x", k))
	}
	sort.Strings(addrs)
	m.out.Printf(strings.Join(addrs, "\n"))
	return nil
}

func (m *Monitor) cmdBreakSet(args []string) error {
	if err := checkLen(args, 1, 1); err != nil {
		return err
	}
	addr, err := m.parseAddress(args[0])
	if err != nil {
		return err
	}
	m.sc.brkpts[addr] = struct{}{}
	return nil
}

func (m *Monitor) cmdCPU(args []string) error {
	if len(args) == 0 {
		m.out.Printf("[%v]\n", m.mach.Status)
		m.out.Printf("%v\n", m.sc.cpu)
		return nil
	}
	switch args[0] {
	case "flag":
		return m.cmdCPUFlag(args[1:])
	case "reg":
		return m.cmdCPUReg(args[1:])
	case "select":
		return m.cmdCPUSelect(args[1:])
	}
	return fmt.Errorf("unknown command: %v", args[0])
}

func (m *Monitor) cmdCPUReg(args []string) error {
	if err := checkLen(args, 0, 2); err != nil {
		return err
	}
	editor, ok := m.sc.cpu.(rcs.CPUEditor)
	if !ok {
		m.out.Printf("no registers")
	}
	if len(args) == 0 {
		return m.cmdCPURegList(editor)
	}
	if len(args) == 1 {
		return m.cmdCPURegGet(editor, args[0])
	}
	return m.cmdCPURegPut(editor, args[0], args[1])
}

func (m *Monitor) cmdCPURegList(editor rcs.CPUEditor) error {
	names := []string{}
	for k := range editor.Registers() {
		names = append(names, k)
	}
	sort.Strings(names)
	m.out.Printf(strings.Join(names, "\n"))
	return nil
}

func (m *Monitor) cmdCPURegGet(editor rcs.CPUEditor, name string) error {
	reg, ok := editor.Registers()[name]
	if !ok {
		return fmt.Errorf("no such register: %v", name)
	}
	return formatGet(m, reg)
}

func (m *Monitor) cmdCPURegPut(editor rcs.CPUEditor, name string, val string) error {
	reg, ok := editor.Registers()[name]
	if !ok {
		return fmt.Errorf("no such register: %v", name)
	}
	return parsePut(m, val, reg)
}

func (m *Monitor) cmdCPUFlag(args []string) error {
	if err := checkLen(args, 0, 2); err != nil {
		return err
	}
	editor, ok := m.sc.cpu.(rcs.CPUEditor)
	if !ok {
		return fmt.Errorf("no registers")
	}
	switch len(args) {
	case 1:
		return m.cmdCPUFlagGet(editor, args[0])
	case 2:
		return m.cmdCPUFlagPut(editor, args[0], args[1])
	}
	return m.cmdCPUFlagList(editor)
}

func (m *Monitor) cmdCPUFlagList(editor rcs.CPUEditor) error {
	names := []string{}
	for k := range editor.Flags() {
		names = append(names, k)
	}
	sort.Strings(names)
	m.out.Printf(strings.Join(names, "\n"))
	return nil
}

func (m *Monitor) cmdCPUFlagGet(editor rcs.CPUEditor, name string) error {
	reg, ok := editor.Flags()[name]
	if !ok {
		return fmt.Errorf("no such flag: %v", name)
	}
	return formatGet(m, reg)
}

func (m *Monitor) cmdCPUFlagPut(editor rcs.CPUEditor, name string, val string) error {
	reg, ok := editor.Flags()[name]
	if !ok {
		return fmt.Errorf("no such flag: %v", name)
	}
	return parsePut(m, val, reg)
}

func (m *Monitor) cmdCPUSelect(args []string) error {
	if err := checkLen(args, 1, 1); err != nil {
		return err
	}
	value, err := m.parseValue(args[0])
	if err != nil {
		return err
	}
	if value < 0 || value >= len(m.mach.CPU) {
		return fmt.Errorf("invalid core: %v", value)
	}
	m.sc = &m.cores[value]
	m.rl.SetPrompt(m.getPrompt())
	return nil
}

func (m *Monitor) cmdDasm(args []string) error {
	if err := checkLen(args, 1, maxArgs); err != nil {
		return err
	}
	switch args[0] {
	case "list":
		return m.cmdDasmList(args[1:])
	case "lines":
		return m.cmdDasmLines(args[1:])
	}
	return fmt.Errorf("unknown command: %v", args[0])
}

func (m *Monitor) cmdDasmList(args []string) error {
	if err := checkLen(args, 0, 2); err != nil {
		return err
	}
	if m.sc.dasm == nil {
		return fmt.Errorf("cannot disassemble this processor")
	}
	if len(args) > 0 {
		addr, err := m.parseAddress(args[0])
		if err != nil {
			return err
		}
		m.sc.dasm.SetPC(addr)
	}
	if len(args) > 1 {
		// list until at ending address
		addrEnd, err := m.parseAddress(args[1])
		if err != nil {
			return err
		}
		for m.sc.dasm.PC() <= addrEnd {
			m.out.Println(m.sc.dasm.Next())
		}
	} else {
		// list number of lines
		lines := m.dasmLines
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
			m.out.Println(m.sc.dasm.Next())
		}
	}
	m.lastCmd = m.cmdDasmList
	return nil
}

func (m *Monitor) cmdDasmLines(args []string) error {
	if err := checkLen(args, 0, 1); err != nil {
		return err
	}
	if len(args) == 0 {
		m.out.Println(m.dasmLines)
		return nil
	}
	lines, err := m.parseValue(args[0])
	if err != nil {
		return err
	}
	if lines < 0 {
		m.out.Printf("invalid value: %v", args[0])
	}
	m.dasmLines = lines
	return nil
}

func (m *Monitor) cmdExport(args []string) error {
	if err := checkLen(args, 0, 1); err != nil {
		return err
	}
	filename := "state"
	if len(args) > 0 {
		filename = args[0]
	}
	file := filepath.Join(config.VarDir, filename)
	m.mach.Command(rcs.MachExport, file)
	return nil
}

func (m *Monitor) cmdGo(args []string) error {
	if err := checkLen(args, 0, 0); err != nil {
		return err
	}
	m.mach.Command(rcs.MachStart)
	return nil
}

func (m *Monitor) cmdImport(args []string) error {
	if err := checkLen(args, 0, 1); err != nil {
		return err
	}
	filename := "state"
	if len(args) > 0 {
		filename = args[0]
	}
	file := filepath.Join(config.VarDir, filename)
	m.mach.Command(rcs.MachImport, file)
	return nil
}

func (m *Monitor) cmdMemory(args []string) error {
	if err := checkLen(args, 1, maxArgs); err != nil {
		return err
	}
	switch args[0] {
	case "dump":
		return m.cmdMemoryDump(args[1:])
	case "fill":
		return m.cmdMemoryFill(args[1:])
	case "encoding":
		return m.cmdMemoryEncoding(args[1:])
	case "lines":
		return m.cmdMemoryLines(args[1:])
	}
	return fmt.Errorf("unknown command: %v", args[0])
}

func (m *Monitor) cmdMemoryDump(args []string) error {
	if err := checkLen(args, 0, 2); err != nil {
		return err
	}
	addrStart := m.sc.cpu.PC()
	if len(args) == 0 {
		addrStart = m.sc.ptr.Addr()
	}
	if len(args) > 0 {
		addr, err := m.parseAddress(args[0])
		if err != nil {
			return err
		}
		addrStart = addr
	}
	addrEnd := addrStart + (m.memLines * 16)
	if len(args) > 1 {
		addr, err := m.parseAddress(args[1])
		if err != nil {
			return err
		}
		addrEnd = addr
	}
	decoder, ok := m.mach.CharDecoders[m.encoding]
	if !ok {
		return fmt.Errorf("invalid encoding: %v", m.encoding)
	}
	m.out.Println(dump(m.sc.mem, addrStart, addrEnd, decoder))
	m.sc.ptr.SetAddr(addrEnd)
	m.lastCmd = m.cmdMemoryDump
	return nil
}

func (m *Monitor) cmdMemoryEncoding(args []string) error {
	if err := checkLen(args, 0, 1); err != nil {
		return err
	}
	if len(args) == 0 {
		return m.cmdMemoryEncodingList()
	}
	return m.cmdMemoryEncodingSet(args[0])
}

func (m *Monitor) cmdMemoryEncodingList() error {
	names := make([]string, 0, 0)
	for k := range m.mach.CharDecoders {
		names = append(names, k)
	}
	sort.Strings(names)
	for i := 0; i < len(names); i++ {
		if names[i] == m.encoding {
			names[i] = "* " + names[i]
		} else {
			names[i] = "  " + names[i]
		}
	}
	m.out.Println(strings.Join(names, "\n"))
	return nil
}

func (m *Monitor) cmdMemoryEncodingSet(enc string) error {
	_, ok := m.mach.CharDecoders[enc]
	if !ok {
		return fmt.Errorf("no such encoding: %v", enc)
	}
	m.encoding = enc
	return nil
}

func (m *Monitor) cmdMemoryFill(args []string) error {
	if err := checkLen(args, 3, 3); err != nil {
		return err
	}
	startAddr, err := m.parseAddress(args[0])
	if err != nil {
		return err
	}
	endAddr, err := m.parseAddress(args[1])
	if err != nil {
		return err
	}
	value, err := m.parseValue8(args[2])
	if err != nil {
		return err
	}
	if startAddr > endAddr {
		return nil
	}
	for addr := startAddr; addr <= endAddr; addr++ {
		m.sc.mem.Write(addr, value)
	}
	return nil
}

func (m *Monitor) cmdMemoryLines(args []string) error {
	if err := checkLen(args, 0, 1); err != nil {
		return err
	}
	if len(args) == 0 {
		m.out.Println(m.memLines)
		return nil
	}
	lines, err := m.parseValue(args[0])
	if err != nil {
		return err
	}
	if lines <= 0 {
		m.out.Printf("invalid value: %v", args[0])
	}
	m.memLines = lines
	return nil
}

func (m *Monitor) cmdNext(args []string) error {
	if err := checkLen(args, 0, 0); err != nil {
		return err
	}
	m.sc.dasm.SetPC(m.sc.cpu.PC() + m.sc.cpu.Offset())
	m.out.Println(m.sc.dasm.Next())
	return nil
}

func (m *Monitor) cmdPause(args []string) error {
	if err := checkLen(args, 0, 0); err != nil {
		return err
	}
	m.mach.Command(rcs.MachPause)
	return nil
}

func (m *Monitor) cmdPoke(args []string) error {
	if err := checkLen(args, 1, maxArgs); err != nil {
		return err
	}
	addr, err := m.parseAddress(args[0])
	if err != nil {
		return err
	}
	values := []uint8{}
	for _, str := range args[1:] {
		v, err := m.parseValue8(str)
		if err != nil {
			return err
		}
		values = append(values, v)
	}
	m.sc.mem.WriteN(addr, values...)
	return nil
}

func (m *Monitor) cmdPeek(args []string) error {
	if err := checkLen(args, 1, 1); err != nil {
		return err
	}
	addr, err := m.parseAddress(args[0])
	if err != nil {
		return err
	}
	v := m.sc.mem.Read(addr)
	m.out.Print(m.formatValue(int(v)))
	return nil
}

func (m *Monitor) cmdQuit(args []string) error {
	m.rl.Close()
	m.mach.Command(rcs.MachQuit)
	runtime.Goexit()
	return nil
}

func (m *Monitor) cmdStep(args []string) error {
	if err := checkLen(args, 0, 0); err != nil {
		return err
	}
	m.sc.cpu.Next()
	m.sc.dasm.SetPC(m.sc.cpu.PC() + m.sc.cpu.Offset())
	m.out.Println(m.sc.dasm.Next())
	m.lastCmd = m.cmdStep
	return nil
}

func (m *Monitor) cmdTrace(args []string) error {
	if err := checkLen(args, 0, 0); err != nil {
		return err
	}
	m.mach.Command(rcs.MachTrace)
	return nil
}

func (m *Monitor) cmdWatch(args []string) error {
	if err := checkLen(args, 0, maxArgs); err != nil {
		return err
	}
	if len(args) == 0 {
		return m.cmdWatchList([]string{})
	}
	switch args[0] {
	case "clear":
		return m.cmdWatchClear(args[1:])
	case "clear-all":
		return m.cmdWatchClearAll(args[1:])
	case "list":
		return m.cmdWatchList(args[1:])
	case "set":
		return m.cmdWatchSet(args[1:])
	}
	return fmt.Errorf("no such command: %v", args[0])
}

func (m *Monitor) cmdWatchClear(args []string) error {
	if err := checkLen(args, 1, 1); err != nil {
		return err
	}
	addr, err := m.parseAddress(args[0])
	if err != nil {
		return err
	}
	if _, ok := m.sc.watches[addr]; ok {
		delete(m.sc.watches, addr)
		m.sc.mem.Unwatch(addr)
	}
	return nil
}

func (m *Monitor) cmdWatchClearAll(args []string) error {
	if err := checkLen(args, 0, 0); err != nil {
		return err
	}
	for addr := range m.sc.watches {
		delete(m.sc.watches, addr)
		m.sc.mem.Unwatch(addr)
	}
	return nil
}

func (m *Monitor) cmdWatchList(args []string) error {
	if err := checkLen(args, 0, 0); err != nil {
		return err
	}
	list := make([]string, 0, len(m.sc.watches))
	for addr, mode := range m.sc.watches {
		list = append(list, fmt.Sprintf("$%04x %v", addr, mode))
	}
	sort.Strings(list)
	m.out.Printf(strings.Join(list, "\n"))
	return nil
}

func (m *Monitor) cmdWatchSet(args []string) error {
	if err := checkLen(args, 2, 2); err != nil {
		return err
	}
	addr, err := m.parseAddress(args[0])
	if err != nil {
		return err
	}
	mode := args[1]
	switch mode {
	case "r", "ro":
		m.sc.mem.WatchRO(addr)
		m.sc.watches[addr] = "r"
	case "w", "wo":
		m.sc.mem.WatchWO(addr)
		m.sc.watches[addr] = "w"
	case "rw":
		m.sc.mem.WatchRW(addr)
		m.sc.watches[addr] = "rw"
	default:
		return fmt.Errorf("unknown watch mode: %v", mode)
	}
	return nil
}

func (m *Monitor) cmdX() error {
	m.mach.Command(rcs.MachTrace, false)
	m.cmdWatchClearAll([]string{})
	return nil
}

func (m *Monitor) cmdYield() error {
	runtime.Gosched()
	time.Sleep(500 * time.Millisecond)
	return nil
}
