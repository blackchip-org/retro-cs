package app

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/chzyer/readline"
)

const (
	CmdBreakpoint  = "b"
	CmdCore        = "c"
	CmdDisassemble = "d"
	CmdFill        = "f"
	CmdGo          = "g"
	CmdHalt        = "h"
	CmdHelp        = "?"
	CmdMemory      = "m"
	CmdNext        = "n"
	CmdPokePeek    = "p"
	CmdRegisters   = "r"
	CmdStep        = "s"
	CmdRestore     = "si"
	CmdSave        = "so"
	CmdTrace       = "t"
	CmdQuit        = "q"
	CmdQuitLong    = "quit"
)

const (
	memPageLen  = 0x100
	dasmPageLen = 0x3f
	maxArgs     = 0x100
)

type Monitor struct {
	dasm        *rcs.Disassembler
	mach        *rcs.Mach
	cpu         rcs.CPU
	mem         *rcs.Memory
	breakpoints map[uint16]struct{}
	in          io.ReadCloser
	out         *log.Logger
	rl          *readline.Instance
	lastCmd     string
	memPtr      uint16
	dasmPtr     uint16
	coreSel     int // selected core
}

func NewMonitor(mach *rcs.Mach) *Monitor {
	m := &Monitor{
		mach: mach,
		in:   readline.NewCancelableStdin(os.Stdin),
		out:  log.New(os.Stdout, "", 0),
	}
	mach.EventCallback = m.handleEvent
	return m
}

func (m *Monitor) Run() error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	rl, err := readline.NewEx(&readline.Config{
		Prompt:      m.getPrompt(),
		HistoryFile: filepath.Join(usr.HomeDir, ".retro-cs-history"),
		Stdin:       m.in,
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
	if line == "" {
		if m.lastCmd != CmdStep && m.lastCmd != CmdGo && m.lastCmd != CmdMemory {
			return
		}
		line = m.lastCmd
	}
	fields := strings.Split(line, " ")

	if len(fields) == 0 {
		return
	}

	cmd := fields[0]
	args := fields[1:]
	var err error
	switch cmd {
	case CmdRegisters:
		err = m.registers(args)
	case CmdQuit, CmdQuitLong:
		m.rl.Close()
		m.mach.Command(rcs.MachQuit{})
		runtime.Goexit()
	default:
		err = fmt.Errorf("unknown command: %v", cmd)
	}

	if err != nil {
		m.out.Println(err)
	} else {
		m.lastCmd = cmd
	}
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
	m.mem = m.mach.Mem[n]
	m.rl.SetPrompt(m.getPrompt())
	return nil
}

func (m *Monitor) registers(args []string) error {
	if err := checkLen(args, 0, 2); err != nil {
		return err
	}

	// Print all registers
	if len(args) == 0 {
		m.out.Printf("[%v]\n", m.mach.Status)
		m.out.Printf("%v\n", m.cpu)
		return nil
	}
	/*
		name := strings.ToUpper(args[0])
		reg, ok := m.cpu.Info().Registers[name]
		if !ok {
			return errors.New("no such register")
		}

		// Get value of register
		if len(args) == 1 {
			switch get := reg.Get.(type) {
			case func() uint8:
				m.out.Println(formatValue(get()))
			case func() uint16:
				m.out.Println(formatValue16(get()))
			default:
				panic("unexpected type")
			}
			return nil
		}

		// Set value of register
			switch put := reg.Put.(type) {
			case func(uint8):
				v, err := parseValue(args[1])
				if err != nil {
					return nil
				}
				put(v)
			case func(uint16):
				v, err := parseValue16(args[1])
				if err != nil {
					return nil
				}
				put(v)
			}
	*/
	return nil
}

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

func (m *Monitor) handleEvent(ty rcs.EventType, event interface{}) {
	log.Printf("event: %v", event)
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

func parseAddress(str string) (uint16, error) {
	value, err := parseUint(str, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid address: %v", str)
	}
	return uint16(value), nil
}

func parseValue(str string) (uint8, error) {
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

func formatValue(v uint8) string {
	return fmt.Sprintf("$%02x +%d", v, v)
}

func formatValue16(v uint16) string {
	return fmt.Sprintf("$%04x +%d", v, v)
}

var AsciiDecoder = func(code uint8) (rune, bool) {
	printable := code >= 32 && code < 128
	return rune(code), printable
}
