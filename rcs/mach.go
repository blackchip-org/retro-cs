package rcs

import (
	"fmt"
	"time"
)

type Status int

const (
	Pause Status = iota
	Run
	Break
)

const (
	vblank   = time.Duration(16670 * time.Microsecond)
	perJiffy = 20000 // instructions per jiffy
)

func (s Status) String() string {
	switch s {
	case Pause:
		return "pause"
	case Run:
		return "run"
	case Break:
		return "break"
	}
	return "???"
}

type MachCmd int

const (
	MachPause MachCmd = iota
	MachStart
	MachTrace
	MachQuit
)

type message struct {
	Cmd  MachCmd
	Args []interface{}
}

type MachEvent int

const (
	StatusEvent MachEvent = iota
	TraceEvent
	ErrorEvent
)

type Mach struct {
	Mem             []*Memory
	CPU             []CPU
	CharDecoders    map[string]CharDecoder
	DefaultEncoding string

	Status      Status
	Callback    func(MachEvent, ...interface{})
	Breakpoints []map[int]struct{}

	init    bool
	tracing bool
	quit    bool
	cmd     chan message
}

type StatusReply struct {
	Status Status
}

type ErrorReply struct {
	Err error
}

func (m *Mach) Init() {
	if m.init {
		return
	}
	m.quit = false
	if m.CharDecoders == nil {
		m.CharDecoders = map[string]CharDecoder{
			"ascii": AsciiDecoder,
		}
		m.DefaultEncoding = "ascii"
	}
	m.cmd = make(chan message, 1)
	cores := len(m.CPU)
	m.Breakpoints = make([]map[int]struct{}, cores, cores)
	for i := 0; i < cores; i++ {
		m.Breakpoints[i] = make(map[int]struct{})
	}
	m.init = true
}

func (m *Mach) Run() {
	m.Init()
	ticker := time.NewTicker(vblank)
	for {
		select {
		case c := <-m.cmd:
			m.handleCommand(c)
		case <-ticker.C:
			m.jiffy()
		}
		if m.quit {
			return
		}
	}
}

func (m *Mach) Command(cmd MachCmd, args ...interface{}) {
	m.cmd <- message{Cmd: cmd, Args: args}
}

func (m *Mach) jiffy() {
	if m.Status == Run {
		m.execute()
	}
	time.Sleep(100 * time.Millisecond) // until vsync gets in
}

func (m *Mach) execute() {
	for core, cpu := range m.CPU {
		for t := 0; t < perJiffy; t++ {
			ppc := cpu.PC()
			cpu.Next()
			// if the program counter didn't change, it is either stuck
			// in an infinite loop or not advancing due to a halt-like
			// instruction
			stuck := ppc == cpu.PC()
			if m.tracing && !stuck {
				m.event(TraceEvent, core, ppc)
			}
			// at a breakpoint? only honor it if the processor is not stuck.
			// when at a halt-like instruction, this causes a break once
			// instead of each time.
			if _, yes := m.Breakpoints[core][cpu.PC()]; yes && !stuck {
				m.setStatus(Break)
				break // allow other CPUs to be serviced
			}
		}
	}
}

func (m *Mach) handleCommand(msg message) {
	switch msg.Cmd {
	case MachPause:
		m.setStatus(Pause)
	case MachStart:
		m.setStatus(Run)
	case MachTrace:
		m.tracing = !m.tracing
	case MachQuit:
		m.quit = true
	default:
		m.event(ErrorEvent, fmt.Errorf("unknown command: %v", msg.Cmd))
	}
}

func (m *Mach) event(evt MachEvent, args ...interface{}) {
	if m.Callback == nil {
		return
	}
	m.Callback(evt, args...)
}

func (m *Mach) setStatus(s Status) {
	m.Status = s
	m.event(StatusEvent, s)
}

var AsciiDecoder = func(code uint8) (rune, bool) {
	printable := code >= 32 && code < 128
	return rune(code), printable
}
