package monitor

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/blackchip-org/retro-cs/config"
	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/chzyer/readline"
)

var cmds = map[string]func(*Monitor, []string) error{}

const (
	maxArgs = 0x100
)

type core struct {
	id      int
	cpu     rcs.CPU
	mem     *rcs.Memory
	ptr     *rcs.Pointer
	dasm    *rcs.Disassembler
	tracer  *rcs.Disassembler
	brkpts  map[int]struct{}
	watches map[int]string
}

type Monitor struct {
	mach      *rcs.Mach
	sc        *core // selected core
	cores     []core
	in        io.ReadCloser
	out       *log.Logger
	rl        *readline.Instance
	encoding  string
	lastCmd   func([]string) error
	memLines  int
	dasmLines int
}

func New(mach *rcs.Mach) *Monitor {
	mach.Init()
	m := &Monitor{
		mach:     mach,
		in:       readline.NewCancelableStdin(os.Stdin),
		cores:    make([]core, len(mach.CPU), len(mach.CPU)),
		memLines: 16, // show a full page on "m" command
	}
	for i, cpu := range mach.CPU {
		var dasm *rcs.Disassembler
		var tracer *rcs.Disassembler
		cpud, ok := cpu.(rcs.CPUDisassembler)
		if ok {
			dasm = cpud.NewDisassembler()
			tracer = cpud.NewDisassembler()
		}
		c := core{
			id:      i,
			cpu:     cpu,
			mem:     mach.Mem[i],
			ptr:     rcs.NewPointer(mach.Mem[i]),
			dasm:    dasm,
			tracer:  tracer,
			brkpts:  mach.Breakpoints[i],
			watches: make(map[int]string),
		}
		c.mem.Callback = m.memCallback
		m.cores[i] = c
	}
	m.sc = &m.cores[0]
	mach.Callback = m.cpuCallback
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
	historyFile := ""
	if config.UserDir != "" {
		historyFile = filepath.Join(config.UserDir, "history")
	}
	rl, err := readline.NewEx(&readline.Config{
		Prompt:       m.getPrompt(),
		Stdin:        m.in,
		HistoryFile:  historyFile,
		AutoComplete: newCompleter(m),
	})
	if err != nil {
		return err
	}
	m.rl = rl
	if m.out == nil {
		m.out = log.New(newConsoleWriter(rl), "", 0)
	}
	m.rl.SetPrompt(m.getPrompt())

	for {
		line, err := m.rl.Readline()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		m.parse(line)
	}
}

func (m *Monitor) Eval(str string) error {
	lines := strings.Split(str, "\n")
	for _, line := range lines {
		args := splitArgs(line)
		if len(args) > 0 {
			m.out.Printf("+ %v\n", line)
			err := m.cmd(args)
			if err != nil {
				m.out.Printf("%v", err)
				return err
			}
		}
	}
	return nil
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
	args := splitArgs(line)
	err := m.cmd(args)
	if err != nil {
		m.out.Printf("%v", err)
		return
	}
}

func (m *Monitor) Close() {
	m.in.Close()
	m.rl.Close()
}

func (m *Monitor) getPrompt() string {
	c := ""
	if len(m.mach.CPU) > 1 {
		c = fmt.Sprintf(":%v", m.sc.id)
	}
	return fmt.Sprintf("monitor%v> ", c)
}

func (m *Monitor) cpuCallback(evt rcs.MachEvent, args ...interface{}) {
	switch evt {
	case rcs.TraceEvent:
		core := args[0].(int)
		pc := args[1].(int)
		if core == m.sc.id {
			m.sc.dasm.SetPC(pc + m.sc.cpu.Offset())
			m.out.Printf("%v", m.sc.dasm.Next())
		}
	case rcs.ErrorEvent:
		m.out.Println(args[0])
	case rcs.StatusEvent:
		status := args[0].(rcs.Status)
		if status == rcs.Break {
			m.out.Println()
			m.cmdCPU([]string{})
			m.rl.Refresh()
		}
	}
}

func (m *Monitor) memCallback(evt rcs.MemoryEvent) {
	a := fmt.Sprintf("$%04x", evt.Addr)
	if m.sc.mem.NBank > 1 {
		a = fmt.Sprintf("%v:$%04x", evt.Bank, evt.Addr)
	}
	if evt.Read {
		m.out.Printf("$%02x <= read(%v)", evt.Value, a)
	} else {
		m.out.Printf("write(%v) => $%02x", a, evt.Value)
	}
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
	base := 10
	switch {
	case strings.HasPrefix(str, "$"):
		str = str[1:]
		base = 16
	case strings.HasPrefix(str, "0x"):
		str = str[2:]
		base = 16
	case strings.HasPrefix(str, "%"):
		str = str[1:]
		base = 2
	case strings.HasPrefix(str, "0b"):
		str = str[2:]
		base = 2
	}
	return strconv.ParseUint(str, base, bitSize)
}

func (m *Monitor) parseAddress(str string) (int, error) {
	value, err := parseUint(str, 64)
	if err != nil || int(value) > m.sc.mem.MaxAddr {
		return 0, fmt.Errorf("invalid address: %v", str)
	}
	return int(value), nil
}

func (m *Monitor) parseValue(str string) (int, error) {
	value, err := parseUint(str, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid value: %v", str)
	}
	return int(value), nil
}

func (m *Monitor) parseValue8(str string) (uint8, error) {
	value, err := parseUint(str, 8)
	if err != nil {
		return 0, fmt.Errorf("invalid value: %v", str)
	}
	return uint8(value), nil
}

func (m *Monitor) parseValue16(str string) (uint16, error) {
	value, err := parseUint(str, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid value: %v", str)
	}
	return uint16(value), nil
}

func (m *Monitor) parseBool(str string) (bool, error) {
	switch str {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	}
	return false, fmt.Errorf("invalid value: %v", str)
}

func (m *Monitor) formatValue(v int) string {
	return fmt.Sprintf("%d $%x %%%b", v, v, v)
}

func formatGet(m *Monitor, val rcs.Value) error {
	switch get := val.Get.(type) {
	case func() uint8:
		m.out.Print(m.formatValue(int(get())))
	case func() uint16:
		m.out.Print(m.formatValue(int(get())))
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
		v, err := m.parseValue8(in)
		if err != nil {
			return err
		}
		put(v)
	case func(uint16):
		v, err := m.parseValue16(in)
		if err != nil {
			return err
		}
		put(v)
	case func(bool):
		v, err := m.parseBool(in)
		if err != nil {
			return err
		}
		put(v)
	default:
		return fmt.Errorf("unknown type: %v", reflect.TypeOf(val.Put))
	}
	return nil
}

func loadPath(name string) string {
	if filepath.IsAbs(name) {
		return name
	}
	return filepath.Join(config.DataDir, name)
}

var whitespaceRegex = regexp.MustCompile("\\s+")

func splitArgs(line string) []string {
	line = strings.TrimSpace(line)
	if line == "" || line[0] == '#' {
		return []string{}
	}
	return whitespaceRegex.Split(line, -1)
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
