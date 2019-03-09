package monitor

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/chzyer/readline"

	"github.com/blackchip-org/retro-cs/rcs"
)

type modMemory struct {
	name    string
	mon     *Monitor
	out     *log.Logger
	mem     *rcs.Memory
	ptr     *rcs.Pointer
	watches map[int]string
}

func newModMemory(mon *Monitor, comp rcs.Component) module {
	mem := comp.C.(*rcs.Memory)
	mod := &modMemory{
		name:    comp.Name,
		mon:     mon,
		out:     mon.out,
		mem:     mem,
		ptr:     rcs.NewPointer(mem),
		watches: make(map[int]string),
	}
	mem.Callback = mod.watchCallback
	return mod
}

func (m *modMemory) Command(args []string) error {
	if err := checkLen(args, 0, maxArgs); err != nil {
		return err
	}
	if len(args) == 0 {
		return m.cmdDump(args[0:])
	}
	switch args[0] {
	case "dump":
		return m.cmdDump(args[1:])
	case "fill":
		return m.cmdFill(args[1:])
	case "peek":
		return m.cmdPeek(args[1:])
	case "poke":
		return m.cmdPoke(args[1:])
	case "watch", "w":
		return m.cmdWatch(args[1:])
	}
	return m.cmdDump(args[0:])
}

func (m *modMemory) cmdDump(args []string) error {
	if err := checkLen(args, 0, 2); err != nil {
		return err
	}
	addrStart := 0
	if len(args) == 0 {
		addrStart = m.ptr.Addr()
	}
	if len(args) > 0 {
		addr, err := parseAddress(m.mem, args[0])
		if err != nil {
			return err
		}
		addrStart = addr
	}
	addrEnd := addrStart + (m.mon.memLines * 16)
	if len(args) > 1 {
		addr, err := parseAddress(m.mem, args[1])
		if err != nil {
			return err
		}
		addrEnd = addr
	}
	decoder, ok := m.mon.mach.CharDecoders[m.mon.encoding]
	if !ok {
		return fmt.Errorf("invalid encoding: %v", m.mon.encoding)
	}
	m.mon.out.Println(dump(m.mem, addrStart, addrEnd, decoder, m.prefix()))
	m.ptr.SetAddr(addrEnd)
	m.mon.defaultCmd = fmt.Sprintf("%v dump", m.name)
	return nil
}

func (m *modMemory) cmdFill(args []string) error {
	if err := checkLen(args, 3, 3); err != nil {
		return err
	}
	startAddr, err := parseAddress(m.mem, args[0])
	if err != nil {
		return err
	}
	endAddr, err := parseAddress(m.mem, args[1])
	if err != nil {
		return err
	}
	value, err := parseValue8(args[2])
	if err != nil {
		return err
	}
	if startAddr > endAddr {
		return nil
	}
	for addr := startAddr; addr <= endAddr; addr++ {
		m.mem.Write(addr, value)
	}
	return nil
}

func (m *modMemory) cmdPeek(args []string) error {
	if err := checkLen(args, 1, 1); err != nil {
		return err
	}
	addr, err := parseAddress(m.mem, args[0])
	if err != nil {
		return err
	}
	v := m.mem.Read(addr)
	m.mon.out.Print(formatValue(int(v)))
	return nil
}

func (m *modMemory) cmdPoke(args []string) error {
	if err := checkLen(args, 2, maxArgs); err != nil {
		return err
	}
	addr, err := parseAddress(m.mem, args[0])
	if err != nil {
		return err
	}
	values := []uint8{}
	for _, str := range args[1:] {
		v, err := parseValue8(str)
		if err != nil {
			return err
		}
		values = append(values, v)
	}
	m.mem.WriteN(addr, values...)
	return nil
}

func (m *modMemory) cmdWatch(args []string) error {
	if len(args) == 0 {
		return m.cmdWatchList(args[0:])
	}
	switch args[0] {
	case "list":
		return m.cmdWatchList(args[1:])
	case "none":
		return m.cmdWatchNone(args[1:])
	}
	return m.cmdWatchSwitch(args[0:])
}

func (m *modMemory) cmdWatchClear(args []string) error {
	if err := checkLen(args, 1, 1); err != nil {
		return err
	}
	addr, err := parseAddress(m.mem, args[0])
	if err != nil {
		return err
	}
	if _, ok := m.watches[addr]; ok {
		delete(m.watches, addr)
		m.mem.Unwatch(addr)
	}
	return nil
}

func (m *modMemory) cmdWatchList(args []string) error {
	if err := checkLen(args, 0, 0); err != nil {
		return err
	}
	if len(m.watches) == 0 {
		return nil
	}
	list := make([]string, 0, len(m.watches))
	for addr, mode := range m.watches {
		list = append(list, fmt.Sprintf("%v$%04x %v", m.prefix(), addr, mode))
	}
	sort.Strings(list)
	m.mon.out.Print(strings.Join(list, "\n"))
	return nil
}

func (m *modMemory) cmdWatchNone(args []string) error {
	if err := checkLen(args, 0, 0); err != nil {
		return err
	}
	for addr := range m.watches {
		delete(m.watches, addr)
		m.mem.Unwatch(addr)
	}
	return nil
}

func (m *modMemory) cmdWatchSwitch(args []string) error {
	if err := checkLen(args, 1, 2); err != nil {
		return nil
	}
	addr, err := parseAddress(m.mem, args[0])
	if err != nil {
		return fmt.Errorf("invalid address: %v", args[0])
	}
	if len(args) == 1 {
		if state, ok := m.watches[addr]; ok {
			m.out.Println(state)
		} else {
			m.out.Println("off")
		}
		return nil
	}
	mode := args[1]
	switch mode {
	case "r", "ro":
		m.mem.WatchRO(addr)
		m.watches[addr] = "r"
	case "w", "wo":
		m.mem.WatchWO(addr)
		m.watches[addr] = "w"
	case "rw":
		m.mem.WatchRW(addr)
		m.watches[addr] = "rw"
	case "off":
		m.mem.Unwatch(addr)
		delete(m.watches, addr)
	default:
		return fmt.Errorf("invalid argument: %v", mode)
	}
	return nil
}

func (m *modMemory) watchCallback(evt rcs.MemoryEvent) {
	// FIXME: hard coded address format
	a := fmt.Sprintf("$%04x", evt.Addr)
	if m.mem.NBank > 1 {
		a = fmt.Sprintf("%v:$%04x", evt.Bank, evt.Addr)
	}
	if evt.Read {
		m.mon.out.Printf("%v$%02x <= read(%v)", m.prefix(), evt.Value, a)
	} else {
		// FIXME: change the arrow to <=
		m.mon.out.Printf("%vwrite(%v) => $%02x", m.prefix(), a, evt.Value)
	}
}

func (m *modMemory) prefix() string {
	if m.name == "mem" {
		return ""
	}
	return m.name + "  "
}

func (m *modMemory) AutoComplete() []readline.PrefixCompleterInterface {
	return []readline.PrefixCompleterInterface{
		readline.PcItem("dump"),
		readline.PcItem("fill"),
		readline.PcItem("peek"),
		readline.PcItem("poke"),
		readline.PcItem("watch-clear"),
		readline.PcItem("watch-list"),
		readline.PcItem("watch-none"),
		readline.PcItem("watch-set"),
	}
}

func (m *modMemory) Silence() error {
	return m.cmdWatchNone([]string{})
}

func dump(m *rcs.Memory, start int, end int, decode rcs.CharDecoder, prefix string) string {
	var buf bytes.Buffer
	var chars bytes.Buffer

	a0 := start / 0x10 * 0x10
	a1 := end / 0x10 * 0x10
	if a1 != end {
		a1 += 0x10
	}
	for addr := a0; addr < a1; addr++ {
		if addr%0x10 == 0 {
			buf.WriteString(fmt.Sprintf("%v$%04x ", prefix, addr))
			chars.Reset()
		}
		if addr < start || addr > end {
			buf.WriteString("   ")
			chars.WriteString(" ")
		} else {
			value := m.Read(addr)
			buf.WriteString(fmt.Sprintf(" %02x", value))
			ch, printable := decode(value)
			if printable {
				chars.WriteString(fmt.Sprintf("%c", ch))
			} else {
				chars.WriteString(".")
			}
		}
		if addr%0x10 == 7 {
			buf.WriteString(" ")
		}
		if addr%0x10 == 0x0f {
			buf.WriteString("  " + chars.String())
			if addr < end-1 {
				buf.WriteString("\n")
			}
		}
	}
	return buf.String()
}
