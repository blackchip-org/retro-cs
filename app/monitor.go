package app

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/chzyer/readline"
)

var cmds = map[string]func(*Monitor, []string) error{}

const (
	maxArgs = 0x100
)

type Monitor struct {
	mach        *rcs.Mach
	cpu         rcs.CPU
	mem         *rcs.Memory
	dasms       []*rcs.Disassembler // for each core
	dasm        *rcs.Disassembler   // for selected core
	breakpoints map[int]struct{}
	in          io.ReadCloser
	out         *log.Logger
	rl          *readline.Instance
	encoding    string
	lastCmd     func([]string) error
	memPtr      *rcs.Pointer
	coreSel     int // selected core
	memLines    int
	dasmLines   int
}

func NewMonitor(mach *rcs.Mach) *Monitor {
	m := &Monitor{
		mach:     mach,
		in:       readline.NewCancelableStdin(os.Stdin),
		dasms:    make([]*rcs.Disassembler, len(mach.CPU), len(mach.CPU)),
		out:      log.New(os.Stdout, "", 0),
		memPtr:   rcs.NewPointer(nil), // will be set on core command
		memLines: 16,                  // show a full page on "m" command
	}
	for i, cpu := range mach.CPU {
		lister, ok := cpu.(rcs.CodeLister)
		if ok {
			m.dasms[i] = lister.NewDisassembler()
		}
	}
	mach.Reply = m.handleReply
	if mach.DefaultEncoding != "" {
		m.encoding = mach.DefaultEncoding
	} else {
		for name := range mach.CharDecoders {
			m.encoding = name
			break
		}
	}
	return m
}

func (m *Monitor) Run() error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	rl, err := readline.NewEx(&readline.Config{
		Prompt:       m.getPrompt(),
		HistoryFile:  filepath.Join(usr.HomeDir, ".retro-cs-history"),
		Stdin:        m.in,
		AutoComplete: newCompleter(m),
	})
	if err != nil {
		return err
	}
	m.rl = rl
	m.core([]string{"1"})
	for {
		line, err := rl.Readline()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		m.parse(line)
	}
}

func (m *Monitor) parse(line string) {
	line = strings.TrimSpace(line)
	if line == "" && m.lastCmd != nil {
		m.lastCmd([]string{})
		return
	}
	if line == "" {
		return
	}
	m.lastCmd = nil
	args := strings.Split(line, " ")
	err := m.cmd(args)
	if err != nil {
		m.out.Printf("%v", err)
		return
	}
}

//============================================================================
// commands

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
	case "g", "go":
		return m.cmdGo(args[1:])
	case "m":
		return m.cmdMemoryDump(args[1:])
	case "mem":
		return m.cmdMemory(args[1:])
	case "p", "pause":
		return m.cmdPause(args[1:])
	case "poke":
		return m.cmdPoke(args[1:])
	case "r":
		return m.cmdCPU([]string{})
	case "q", "quit":
		return m.cmdQuit(args[1:])
	case "_yield":
		return m.cmdYield()
	}
	return fmt.Errorf("no such command: %v", args[0])
}

func (m *Monitor) core(args []string) error {
	if err := checkLen(args, 1, 1); err != nil {
		return err
	}
	n, err := parseValue(args[0])
	if err != nil {
		return err
	}
	if n < 1 || int(n) > len(m.mach.CPU) {
		return fmt.Errorf("invalid core")
	}
	n = n - 1
	m.coreSel = int(n)
	m.cpu = m.mach.CPU[n]
	m.breakpoints = m.mach.Breakpoints[n]
	m.mem = m.mach.Mem[n]
	m.dasm = m.dasms[n]
	m.memPtr.Mem = m.mem
	m.rl.SetPrompt(m.getPrompt())
	return nil
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
	delete(m.breakpoints, addr)
	return nil
}

func (m *Monitor) cmdBreakClearAll() error {
	for k := range m.breakpoints {
		delete(m.breakpoints, k)
	}
	return nil
}

func (m *Monitor) cmdBreakList() error {
	if len(m.breakpoints) == 0 {
		return nil
	}
	addrs := make([]string, 0, 0)
	for k := range m.breakpoints {
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
	m.breakpoints[addr] = struct{}{}
	return nil
}

func (m *Monitor) cmdCPU(args []string) error {
	if len(args) == 0 {
		m.out.Printf("[%v]\n", m.mach.Status)
		m.out.Printf("%v\n", m.cpu)
		return nil
	}
	switch args[0] {
	case "reg":
		return m.cmdCPUReg(args[1:])
	case "flag":
		return m.cmdCPUFlag(args[1:])
	}
	return fmt.Errorf("unknown command: %v", args[0])
}

func (m *Monitor) cmdCPUReg(args []string) error {
	if err := checkLen(args, 0, 2); err != nil {
		return err
	}
	editor, ok := m.cpu.(rcs.CPUEditor)
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
	editor, ok := m.cpu.(rcs.CPUEditor)
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
	if m.dasm == nil {
		return fmt.Errorf("cannot disassemble this processor")
	}
	if len(args) > 0 {
		addr, err := m.parseAddress(args[0])
		if err != nil {
			return err
		}
		m.dasm.SetPC(addr)
	}
	if len(args) > 1 {
		// list until at ending address
		addrEnd, err := m.parseAddress(args[1])
		if err != nil {
			return err
		}
		for m.dasm.PC() <= addrEnd {
			m.out.Println(m.dasm.Next())
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
			m.out.Println(m.dasm.Next())
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
	lines, err := parseValue(args[0])
	if err != nil {
		return err
	}
	if lines < 0 {
		m.out.Printf("invalid value: %v", args[0])
	}
	m.dasmLines = lines
	return nil
}

func (m *Monitor) cmdGo(args []string) error {
	if err := checkLen(args, 0, 1); err != nil {
		return err
	}
	cmd := rcs.MachStart{Core: m.coreSel}
	if len(args) == 1 {
		addr, err := m.parseAddress(args[0])
		if err != nil {
			return err
		}
		cmd.At = true
		cmd.Addr = addr
	}
	m.mach.Command(cmd)
	return nil
}

func (m *Monitor) cmdMemory(args []string) error {
	if err := checkLen(args, 1, maxArgs); err != nil {
		return err
	}
	switch args[0] {
	case "dump":
		return m.cmdMemoryDump(args[1:])
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
	addrStart := m.cpu.PC()
	if len(args) == 0 {
		addrStart = m.memPtr.Addr()
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
	m.out.Println(dump(m.mem, addrStart, addrEnd, decoder))
	m.memPtr.SetAddr(addrEnd)
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

func (m *Monitor) cmdMemoryLines(args []string) error {
	if err := checkLen(args, 0, 1); err != nil {
		return err
	}
	if len(args) == 0 {
		m.out.Println(m.memLines)
		return nil
	}
	lines, err := parseValue(args[0])
	if err != nil {
		return err
	}
	if lines <= 0 {
		m.out.Printf("invalid value: %v", args[0])
	}
	m.memLines = lines
	return nil
}

func (m *Monitor) cmdPause(args []string) error {
	if err := checkLen(args, 0, 0); err != nil {
		return err
	}
	m.mach.Command(rcs.MachPause{})
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
		v, err := parseValue8(str)
		if err != nil {
			return err
		}
		values = append(values, v)
	}
	m.mem.WriteN(addr, values...)
	return nil
}

func (m *Monitor) cmdQuit(args []string) error {
	m.rl.Close()
	m.mach.Command(rcs.MachQuit{})
	runtime.Goexit()
	return nil
}

func (m *Monitor) cmdYield() error {
	runtime.Gosched()
	time.Sleep(100 * time.Millisecond)
	return nil
}

//============================================================================
// autocomplete

func newCompleter(m *Monitor) *readline.PrefixCompleter {
	return readline.NewPrefixCompleter(
		readline.PcItem("break",
			readline.PcItem("clear"),
			readline.PcItem("clear-all"),
			readline.PcItem("list"),
			readline.PcItem("set"),
		),
		readline.PcItem("cpu",
			readline.PcItem("reg",
				readline.PcItemDynamic(acRegisters(m)),
			),
			readline.PcItem("flag",
				readline.PcItemDynamic(acFlags(m)),
			),
		),
		readline.PcItem("d"),
		readline.PcItem("dasm",
			readline.PcItem("lines"),
			readline.PcItem("list"),
		),
		readline.PcItem("g"),
		readline.PcItem("go"),
		readline.PcItem("m"),
		readline.PcItem("mem",
			readline.PcItem("dump"),
			readline.PcItem("encoding",
				readline.PcItemDynamic(acEncodings(m)),
			),
			readline.PcItem("lines"),
		),
		readline.PcItem("p"),
		readline.PcItem("pause"),
		readline.PcItem("poke"),
		readline.PcItem("r"),
		readline.PcItem("q"),
		readline.PcItem("quit"),
	)
}

func acRegisters(m *Monitor) func(string) []string {
	return func(line string) []string {
		cpu, ok := m.cpu.(rcs.CPUEditor)
		if !ok {
			return []string{}
		}
		names := make([]string, 0)
		for k := range cpu.Registers() {
			names = append(names, k)
		}
		sort.Strings(names)
		return names
	}
}

func acFlags(m *Monitor) func(string) []string {
	return func(line string) []string {
		cpu, ok := m.cpu.(rcs.CPUEditor)
		if !ok {
			return []string{}
		}
		names := make([]string, 0)
		for k := range cpu.Flags() {
			names = append(names, k)
		}
		sort.Strings(names)
		return names
	}
}

func acEncodings(m *Monitor) func(string) []string {
	return func(line string) []string {
		names := make([]string, 0)
		for k := range m.mach.CharDecoders {
			names = append(names, k)
		}
		sort.Strings(names)
		return names
	}
}

//=============================================================================
// aux

func (m *Monitor) Close() {
	m.in.Close()
}

func (m *Monitor) getPrompt() string {
	c := ""
	if len(m.mach.CPU) > 1 {
		c = fmt.Sprintf(":%v", m.coreSel+1)
	}
	return fmt.Sprintf("monitor%v> ", c)
}

func (m *Monitor) handleReply(reply interface{}) {
	log.Printf("reply: %v", reply)
}

func checkLen(args []string, min int, max int) error {
	if len(args) < min {
		return errors.New("not enough arguments")
	}
	if len(args) > max {
		return errors.New("too many arguments")
	}
	return nil
}

func parseUint(str string, bitSize int) (uint64, error) {
	base := 16
	switch {
	case strings.HasPrefix(str, "$"):
		str = str[1:]
	case strings.HasPrefix(str, "0x"):
		str = str[2:]
	case strings.HasPrefix(str, "+"):
		str = str[1:]
		base = 10
	}
	return strconv.ParseUint(str, base, bitSize)
}

func (m *Monitor) parseAddress(str string) (int, error) {
	value, err := parseUint(str, 64)
	if err != nil || int(value) > m.mem.MaxAddr {
		return 0, fmt.Errorf("invalid address: %v", str)
	}
	return int(value), nil
}

func parseValue(str string) (int, error) {
	value, err := parseUint(str, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid value: %v", str)
	}
	return int(value), nil
}

func parseValue8(str string) (uint8, error) {
	value, err := parseUint(str, 8)
	if err != nil {
		return 0, fmt.Errorf("invalid value: %v", str)
	}
	return uint8(value), nil
}

func parseValue16(str string) (uint16, error) {
	value, err := parseUint(str, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid value: %v", str)
	}
	return uint16(value), nil
}

func parseBool(str string) (bool, error) {
	switch str {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	}
	return false, fmt.Errorf("invalid value: %v", str)
}

func formatValue8(v uint8) string {
	return fmt.Sprintf("$%02x +%d %%%08b", v, v, v)
}

func formatValue16(v uint16) string {
	return fmt.Sprintf("$%04x +%d", v, v)
}

func formatGet(m *Monitor, val rcs.Value) error {
	switch get := val.Get.(type) {
	case func() uint8:
		m.out.Print(formatValue8(get()))
	case func() uint16:
		m.out.Print(formatValue16(get()))
	case func() bool:
		m.out.Printf("%v", get())
	default:
		return fmt.Errorf("unknown type: %v", reflect.TypeOf(val.Get))
	}
	return nil
}

func parsePut(m *Monitor, in string, val rcs.Value) error {
	switch put := val.Put.(type) {
	case func(uint8):
		v, err := parseValue8(in)
		if err != nil {
			return err
		}
		put(v)
	case func(uint16):
		v, err := parseValue16(in)
		if err != nil {
			return err
		}
		put(v)
	case func(bool):
		v, err := parseBool(in)
		if err != nil {
			return err
		}
		put(v)
	default:
		return fmt.Errorf("unknown type: %v", reflect.TypeOf(val.Put))
	}
	return nil
}

func dump(m *rcs.Memory, start int, end int, decode rcs.CharDecoder) string {
	var buf bytes.Buffer
	var chars bytes.Buffer

	a0 := start / 0x10 * 0x10
	a1 := end / 0x10 * 0x10
	if a1 != end {
		a1 += 0x10
	}
	for addr := a0; addr < a1; addr++ {
		if addr%0x10 == 0 {
			buf.WriteString(fmt.Sprintf("$%04x ", addr))
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
