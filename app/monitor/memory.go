package monitor

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/chzyer/readline"

	"github.com/blackchip-org/retro-cs/rcs"
)

type modMemory struct {
	name    string
	mon     *Monitor
	mem     *rcs.Memory
	ptr     *rcs.Pointer
	watches map[int]string
}

func newModMemory(mon *Monitor, comp rcs.Component) module {
	mem := comp.C.(*rcs.Memory)
	mod := &modMemory{
		name:    comp.Name,
		mon:     mon,
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
		return m.dump(args[0:])
	}
	switch args[0] {
	case "dump":
		return m.dump(args[1:])
	case "fill":
		return m.fill(args[1:])
	case "peek":
		return m.peek(args[1:])
	case "poke":
		return m.poke(args[1:])
	case "watch-clear":
		return m.watchClear(args[1:])
	case "watch-list", "watch":
		return m.watchList(args[1:])
	case "watch-none":
		return m.watchNone(args[1:])
	case "watch-set":
		return m.watchSet(args[1:])
	}
	return m.dump(args[0:])
}

func (m *modMemory) dump(args []string) error {
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
	m.mon.out.Println(dump(m.mem, addrStart, addrEnd, decoder, m.name))
	m.ptr.SetAddr(addrEnd)
	//mon.lastCmd = m.cmdMemoryDump
	return nil
}

func (m *modMemory) fill(args []string) error {
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

func (m *modMemory) peek(args []string) error {
	if err := checkLen(args, 1, 1); err != nil {
		return err
	}
	addr, err := parseAddress(m.mem, args[0])
	if err != nil {
		return err
	}
	v := m.mem.Read(addr)
	m.mon.out.Print(m.mon.formatValue(int(v)))
	return nil
}

func (m *modMemory) poke(args []string) error {
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

func (m *modMemory) watchClear(args []string) error {
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

func (m *modMemory) watchList(args []string) error {
	if err := checkLen(args, 0, 0); err != nil {
		return err
	}
	list := make([]string, 0, len(m.watches))
	for addr, mode := range m.watches {
		list = append(list, fmt.Sprintf("$%04x %v", addr, mode))
	}
	sort.Strings(list)
	m.mon.out.Printf(strings.Join(list, "\n"))
	return nil
}

func (m *modMemory) watchNone(args []string) error {
	if err := checkLen(args, 0, 0); err != nil {
		return err
	}
	for addr := range m.watches {
		delete(m.watches, addr)
		m.mem.Unwatch(addr)
	}
	return nil
}

func (m *modMemory) watchSet(args []string) error {
	if err := checkLen(args, 2, 2); err != nil {
		return err
	}
	addr, err := parseAddress(m.mem, args[0])
	if err != nil {
		return err
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
	default:
		return fmt.Errorf("unknown watch mode: %v", mode)
	}
	return nil

}

func (m *modMemory) watchCallback(evt rcs.MemoryEvent) {
	// FIXME: hard coded address format
	a := fmt.Sprintf("%v  $%04x", m.name, evt.Addr)
	if m.mem.NBank > 1 {
		a = fmt.Sprintf("%v  %v:$%04x", m.name, evt.Bank, evt.Addr)
	}
	if evt.Read {
		m.mon.out.Printf("$%02x <= read(%v)", evt.Value, a)
	} else {
		m.mon.out.Printf("write(%v) => $%02x", a, evt.Value)
	}
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
