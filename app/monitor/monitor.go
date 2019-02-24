package monitor

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/blackchip-org/retro-cs/config"
	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/chzyer/readline"
)

var cmds = map[string]func(*Monitor, []string) error{}

const (
	maxArgs = 0x100
)

type module interface {
	Command([]string) error
	AutoComplete() []readline.PrefixCompleterInterface
}

var modules = map[string]func(m *Monitor, comp rcs.Component) module{
	"c64":   newModC64,
	"cpu":   newModCPU,
	"m6502": newModM6502,
	"mem":   newModMemory,
	"n06xx": newModN06XX,
	"n51xx": newModN51XX,
	"n54xx": newModN54XX,
	"z80":   newModZ80,
}

type Monitor struct {
	mach      *rcs.Mach
	comps     map[string]rcs.Component
	mods      map[string]module
	cpu       map[string]rcs.CPU
	tracers   map[string]*rcs.Disassembler
	sc        string // selected CPU core
	in        io.ReadCloser
	out       *log.Logger
	rl        *readline.Instance
	encoding  string
	lastCmd   func([]string) error
	memLines  int
	dasmLines int
}

func New(mach *rcs.Mach) (*Monitor, error) {
	mach.Init()
	cw := newConsoleWriter(os.Stdout)
	rw := newRepeatWriter(cw)
	log.SetOutput(rw)

	m := &Monitor{
		mach:     mach,
		comps:    make(map[string]rcs.Component),
		mods:     make(map[string]module),
		cpu:      make(map[string]rcs.CPU),
		tracers:  make(map[string]*rcs.Disassembler),
		in:       readline.NewCancelableStdin(os.Stdin),
		out:      log.New(cw, "", 0),
		memLines: 16, // show a full page on "m" command
	}

	mach.Callback = m.cpuCallback
	if mach.DefaultEncoding != "" {
		m.encoding = mach.DefaultEncoding
	} else {
		for name := range mach.CharDecoders {
			m.encoding = name
			break
		}
	}

	for _, comp := range mach.Comps {
		m.comps[comp.Name] = comp
		m.mods[comp.Name] = modules[comp.Module](m, comp)
		if cpu, ok := comp.C.(rcs.CPU); ok {
			m.cpu[comp.Name] = cpu
			// select the first CPU seen
			if m.sc == "" {
				m.sc = comp.Name
			}

			var tracer *rcs.Disassembler
			cpud, ok := cpu.(rcs.CPUDisassembler)
			if ok {
				tracer = cpud.NewDisassembler()
			}
			m.tracers[comp.Name] = tracer
		}
	}

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
		return nil, err
	}
	m.rl = rl
	cw.RefreshFunc = func() { m.rl.Refresh() }
	m.rl.SetPrompt(m.getPrompt())
	return m, nil
}

func (m *Monitor) Run() error {
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
			err := m.dispatch(args)
			if err != nil {
				m.out.Printf("%v", err)
			}
		}
	}
	return nil
}

func (m *Monitor) Close() {
	m.in.Close()
	m.rl.Close()
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
	err := m.dispatch(args)
	if err != nil {
		m.out.Printf("%v", err)
		return
	}
}

func (m *Monitor) dispatch(args []string) error {
	switch args[0] {
	case
		"breakpoint-clear", "bpc",
		"breakpoint-list", "bpl", "bp",
		"breakpoint-none", "bpn",
		"breakpoint-set", "bps",
		"disassemble", "d",
		"info", "i",
		"next", "n",
		"step", "s",
		"trace", "t":
		return m.mods[m.sc].Command(args)
	case "m":
		parent := m.comps[m.sc].Parent
		return m.mods[parent].Command(args[1:])
	case
		"peek",
		"poke",
		"watch-clear", "wc",
		"watch-list", "w", "wl",
		"watch-none", "wn",
		"watch-set", "ws":
		parent := m.comps[m.sc].Parent
		return m.mods[parent].Command(args)
	case "config":
		return m.cmdConfig(args[1:])
	case "encoding", "e":
		return m.cmdEncoding(args[1:])
	case "export":
		return m.cmdExport(args[1:])
	case "go", "g":
		return m.cmdGo(args[1:])
	case "import":
		return m.cmdImport(args[1:])
	case "pause", "p":
		return m.cmdPause(args[1:])
	case "sleep":
		return m.cmdSleep(args[1:])
	case "q", "quit":
		return m.cmdQuit(args[1:])
	}

	if mod, ok := m.mods[args[0]]; ok {
		return mod.Command(args[1:])
	}

	val, err := parseValue(args[0])
	if err == nil {
		m.out.Print(formatValue(val))
		return nil
	}

	return fmt.Errorf("no such command: %v", args[0])
}

// ============================================================================
// base commands

func (m *Monitor) cmdConfig(args []string) error {
	if err := checkLen(args, 1, maxArgs); err != nil {
		return err
	}
	switch args[0] {
	case "lines-memory":
		return valueInt(m.out, &m.memLines, args[1:])
	case "lines-disassembly":
		return valueInt(m.out, &m.dasmLines, args[1:])
	}
	return fmt.Errorf("no such configuration: %v", args[0])
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

func (m *Monitor) cmdEncoding(args []string) error {
	list := make([]string, 0, 0)
	for k := range m.mach.CharDecoders {
		list = append(list, k)
	}
	return valueList(m.out, &m.encoding, list, args)
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

func (m *Monitor) cmdGo(args []string) error {
	if err := checkLen(args, 0, 0); err != nil {
		return err
	}
	m.mach.Command(rcs.MachStart)
	return nil
}

func (m *Monitor) cmdPause(args []string) error {
	if err := checkLen(args, 0, 0); err != nil {
		return err
	}
	m.mach.Command(rcs.MachPause)
	return nil
}

func (m *Monitor) cmdQuit(args []string) error {
	m.rl.Close()
	m.mach.Command(rcs.MachQuit)
	runtime.Goexit()
	return nil
}

func (m *Monitor) cmdSleep(args []string) error {
	if err := checkLen(args, 0, 1); err != nil {
		return err
	}
	duration := 1 * time.Second
	if len(args) > 0 {
		v, err := parseValue(args[0])
		if err != nil {
			return err
		}
		duration = time.Duration(v) * time.Millisecond
	}
	runtime.Gosched()
	time.Sleep(duration)
	return nil
}

// ============================================================================
// autocomplete

type byName []readline.PrefixCompleterInterface

func (n byName) Len() int      { return len(n) }
func (n byName) Swap(i, j int) { n[i], n[j] = n[j], n[i] }
func (n byName) Less(i, j int) bool {
	return string(n[i].GetName()) < string(n[j].GetName())
}

func newCompleter(m *Monitor) *readline.PrefixCompleter {
	cmds := []readline.PrefixCompleterInterface{
		readline.PcItem("breakpoint-clear"),
		readline.PcItem("breakpoint-list"),
		readline.PcItem("breakpoint-none"),
		readline.PcItem("breakpoint-set"),
		readline.PcItem("config",
			readline.PcItem("lines-memory"),
			readline.PcItem("lines-disassembly"),
		),
		readline.PcItem("encoding",
			readline.PcItemDynamic(acEncodings(m)),
		),
		readline.PcItem("export"),
		readline.PcItem("disassemble"),
		readline.PcItem("import"),
		readline.PcItem("info"),
		readline.PcItem("next"),
		readline.PcItem("quit"),
		readline.PcItem("step"),
		readline.PcItem("sleep"),
		readline.PcItem("watch-clear"),
		readline.PcItem("watch-list"),
		readline.PcItem("watch-none"),
		readline.PcItem("watch-set"),
	}
	for key, mod := range m.mods {
		cmds = append(cmds, []readline.PrefixCompleterInterface{
			readline.PcItem(key, mod.AutoComplete()...),
		}...)
	}
	sort.Sort(byName(cmds))
	return readline.NewPrefixCompleter(cmds...)
}

func acDataFiles(m *Monitor, suffix string) func(string) []string {
	return func(line string) []string {
		results := make([]string, 0, 0)
		arg := ""
		i := strings.LastIndex(line, " ")
		if i >= 0 {
			arg = line[i:]
		}
		var dir, prefix string
		if filepath.IsAbs(arg) {
			dir = filepath.Dir(arg)
			prefix = filepath.Base(arg)
		} else {
			dir = config.DataDir
			prefix = arg
		}
		prefix = strings.TrimSpace(prefix)
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			m.out.Println(err)
			return []string{}
		}
		for _, f := range files {
			name := f.Name()
			if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, suffix) {
				results = append(results, name)
			}
		}
		return results
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

// ============================================================================
// aux

func (m *Monitor) cpuCallback(evt rcs.MachEvent, args ...interface{}) {
	switch evt {
	case rcs.TraceEvent:
		name := args[0].(string)
		pc := args[1].(int)
		prefix := ""
		if name != "cpu" {
			prefix = name + "  "
		}
		m.tracers[name].SetPC(pc + m.cpu[m.sc].Offset())
		m.out.Printf("%v%v", prefix, m.tracers[name].Next())
	case rcs.ErrorEvent:
		m.out.Println(args[0])
	case rcs.StatusEvent:
		status := args[0].(rcs.Status)
		if status == rcs.Break {
			m.out.Println()
			m.dispatch([]string{"i"})
			m.rl.Refresh()
		}
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

func formatValue(v int) string {
	return fmt.Sprintf("%d $%x %%%b", v, v, v)
}

func valueBool(out *log.Logger, val *bool, args []string) error {
	if err := checkLen(args, 0, 1); err != nil {
		return err
	}
	if len(args) == 0 {
		out.Println(*val)
		return nil
	}
	sval := strings.ToLower(args[0])
	switch sval {
	case "true", "yes", "on", "t", "1":
		*val = true
	case "false", "no", "off", "f", "0":
		*val = false
	default:
		return fmt.Errorf("invalid value: %v", args[0])
	}
	return nil
}

func valueInt(out *log.Logger, val *int, args []string) error {
	if err := checkLen(args, 0, 1); err != nil {
		return err
	}
	if len(args) == 0 {
		out.Println(formatValue(int(*val)))
		return nil
	}
	v, err := parseValue(args[0])
	if err != nil {
		return err
	}
	*val = v
	return nil
}

func valueUint8(out *log.Logger, val *uint8, args []string) error {
	if err := checkLen(args, 0, 1); err != nil {
		return err
	}
	if len(args) == 0 {
		out.Println(formatValue(int(*val)))
		return nil
	}
	v, err := parseValue8(args[0])
	if err != nil {
		return err
	}
	*val = v
	return nil
}

func valueUint16(out *log.Logger, val *uint16, args []string) error {
	if err := checkLen(args, 0, 1); err != nil {
		return err
	}
	if len(args) == 0 {
		out.Println(formatValue(int(*val)))
		return nil
	}
	v, err := parseValue16(args[0])
	if err != nil {
		return err
	}
	*val = v
	return nil
}

func valueUint16HL(out *log.Logger, hi *uint8, lo *uint8, args []string) error {
	if err := checkLen(args, 0, 1); err != nil {
		return err
	}
	if len(args) == 0 {
		val := int(*hi)<<8 | int(*lo)
		out.Println(formatValue(val))
		return nil
	}
	v, err := parseValue16(args[0])
	if err != nil {
		return err
	}
	*hi = uint8(v >> 8)
	*lo = uint8(v)
	return nil
}

func valueIntF(out *log.Logger, load rcs.Load, store rcs.Store, args []string) error {
	if err := checkLen(args, 0, 1); err != nil {
		return err
	}
	if len(args) == 0 {
		out.Println(formatValue(load()))
		return nil
	}
	v, err := parseValue(args[0])
	if err != nil {
		return err
	}
	store(v)
	return nil
}

func valueList(out *log.Logger, val *string, list []string, args []string) error {
	if err := checkLen(args, 0, 1); err != nil {
		return err
	}
	if len(args) == 0 {
		out.Println(*val)
		return nil
	}
	for _, i := range list {
		if i == args[0] {
			*val = i
			return nil
		}
	}
	return fmt.Errorf("invalid value: %v", args[0])
}

func valueBit(out *log.Logger, val *uint8, mask uint8, args []string) error {
	if err := checkLen(args, 0, 1); err != nil {
		return err
	}
	if len(args) == 0 {
		out.Println(*val&mask != 0)
		return nil
	}
	switch args[0] {
	case "true", "yes", "on", "t", "1":
		*val |= mask
	case "false", "no", "off", "f", "0":
		*val &^= mask
	default:
		return fmt.Errorf("invalid value: %v", args[0])
	}
	return nil
}

func terminal(args []string, fn func() error) error {
	if err := checkLen(args, 0, 0); err != nil {
		return err
	}
	return fn()
}

func parseAddress(mem *rcs.Memory, str string) (int, error) {
	value, err := parseUint(str, 64)
	if err != nil || int(value) > mem.MaxAddr {
		return 0, fmt.Errorf("invalid address: %v", str)
	}
	return int(value), nil
}

// =========================================================================
// console

// carriage return to go to the beginning of the line
// then ansi escape sequence to clear the line
const (
	ansiClearLine  = "\r\033[2K"
	ansiReset      = "\033[0m"
	ansiLightBlue  = "\033[1;34m"
	ansiLightGreen = "\033[1;32m"
)

func (m *Monitor) getPrompt() string {
	c := ""
	if len(m.mach.CPU) > 1 {
		c = fmt.Sprintf(":%v%v%v", ansiLightBlue, m.sc, ansiReset)
	}
	return fmt.Sprintf("%vmonitor%v%v> ", ansiLightGreen, ansiReset, c)
}

type consoleWriter struct {
	RefreshFunc     func()
	rl              *readline.Instance
	w               io.Writer
	line            bytes.Buffer
	backlog         bytes.Buffer
	timer           *time.Timer
	mutex           sync.Mutex
	firstInterval   time.Duration // wait this long before emitting the first line
	backlogInterval time.Duration // wait this long between backlog processing
	maxUpdate       int           // maximum number of charaters per update
}

func newConsoleWriter(w io.Writer) *consoleWriter {
	cw := &consoleWriter{
		w:               w,
		firstInterval:   time.Millisecond * 10,
		backlogInterval: time.Millisecond * 100,
		maxUpdate:       2000,
	}
	return cw
}

func (c *consoleWriter) Write(p []byte) (int, error) {
	for _, b := range p {
		c.line.WriteByte(b)
		if b == '\n' {
			c.backlog.Write(c.line.Bytes())
			if c.timer == nil {
				c.timer = time.AfterFunc(c.firstInterval, c.emit)
			}
			c.line.Reset()
		}
	}
	return len(p), nil
}

func (c *consoleWriter) emit() {
	c.mutex.Lock()
	defer func() {
		c.mutex.Unlock()
	}()

	if c.backlog.Len() == 0 {
		c.timer = nil
		return
	}
	update := c.backlog.String()
	lines := strings.Count(update, "\n")
	omission := false
	if lines > c.maxUpdate {
		// Count backwards to find start of the first line in the
		// maximum lines allowed per update
		omission = true
		seen := 0
		for i := len(update) - 1; i >= 0; i-- {
			if update[i] == '\n' {
				seen++
				if seen == c.maxUpdate {
					update = update[i+1 : len(update)-1]
					break
				}
			}
		}
	}
	// carriage return to go to the begnning of the line
	// then ansi escape sequence to clear the line
	io.WriteString(c.w, ansiClearLine)
	if omission {
		text := fmt.Sprintf("... omitted %v lines\n", lines-c.maxUpdate)
		io.WriteString(c.w, text)
	}
	io.WriteString(c.w, update)
	c.RefreshFunc()
	c.backlog.Reset()
	c.timer = time.AfterFunc(c.backlogInterval, c.emit)
}

type repeatWriter struct {
	w       io.Writer
	buf     strings.Builder
	prev    string
	repeats int
	ansi    bool
}

func newRepeatWriter(w io.Writer) *repeatWriter {
	return &repeatWriter{w: w, ansi: true}
}

func (r *repeatWriter) Write(p []byte) (int, error) {
	for _, b := range p {
		r.buf.WriteByte(b)
		if b == '\n' {
			r.eoln()
		}
	}
	return len(p), nil
}

func (r *repeatWriter) Close() error {
	if !strings.HasSuffix(r.buf.String(), "\n") {
		r.buf.WriteString("\n")
	}
	r.eoln()
	return nil
}

func (r *repeatWriter) eoln() {
	str := r.buf.String()
	if str != r.prev {
		if r.repeats > 0 {
			if r.ansi {
				io.WriteString(r.w, ansiClearLine)
				io.WriteString(r.w, "... repeats ")
			}
			if r.repeats == 1 {
				io.WriteString(r.w, "1 time\n")
			} else if r.repeats > 1 {
				io.WriteString(r.w, fmt.Sprintf("%d times\n", r.repeats))
			}
		}
		r.repeats = 0
		io.WriteString(r.w, str)
	} else {
		if r.repeats == 0 {
			io.WriteString(r.w, "... repeats ")
		}
		r.repeats++
	}
	r.buf.Reset()
	r.prev = str
}
