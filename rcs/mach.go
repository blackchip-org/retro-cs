package rcs

import (
	"fmt"
	"reflect"
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

type Mach struct {
	Mem             []*Memory
	CPU             []CPU
	CharDecoders    map[string]CharDecoder
	DefaultEncoding string

	Status      Status
	Reply       func(interface{})
	Breakpoints []map[int]struct{}

	quit bool
	cmd  chan interface{}
}

type MachStart struct {
	At   bool
	Core int
	Addr int
}
type MachPause struct{}
type MachQuit struct{}

type StatusReply struct {
	Status Status
}

type ErrorReply struct {
	Err error
}

func (m *Mach) init() {
	m.quit = false
	if m.CharDecoders == nil {
		m.CharDecoders = map[string]CharDecoder{
			"ascii": AsciiDecoder,
		}
		m.DefaultEncoding = "ascii"
	}
	if m.Reply == nil {
		m.Reply = func(interface{}) {}
	}
	m.cmd = make(chan interface{}, 1)
	cores := len(m.CPU)
	m.Breakpoints = make([]map[int]struct{}, cores, cores)
	for i := 0; i < cores; i++ {
		m.Breakpoints[i] = make(map[int]struct{})
	}
}

func (m *Mach) Run() {
	m.init()
	ticker := time.NewTicker(vblank)
	for {
		select {
		case c := <-m.cmd:
			m.command(c)
		case <-ticker.C:
			m.jiffy()
		}
		if m.quit {
			return
		}
	}
}

func (m *Mach) Command(c interface{}) {
	m.cmd <- c
}

func (m *Mach) jiffy() {
	if m.Status == Run {
		m.execute()
	}
	time.Sleep(100 * time.Millisecond) // until vsync gets in
}

func (m *Mach) execute() {
	for i, cpu := range m.CPU {
		for t := 0; t < perJiffy; t++ {
			ppc := cpu.PC()
			cpu.Next()
			// if the program counter didn't change, it is either stuck
			// in an infinite loop or not advancing due to a halt-like
			// instruction
			stuck := ppc == cpu.PC()
			// at a breakpoint? only honor it if the processor is not stuck.
			// when at a halt-like instruction, this causes a break once
			// instead of each time.
			if _, yes := m.Breakpoints[i][cpu.PC()]; yes && !stuck {
				m.setStatus(Break)
				continue // allow other CPUs to be serviced
			}
		}
	}
}

func (m *Mach) command(c interface{}) {
	switch cmd := c.(type) {
	case MachPause:
		m.setStatus(Pause)
	case MachStart:
		if cmd.At {
			m.CPU[cmd.Core].SetPC(cmd.Addr)
		}
		m.setStatus(Run)
	case MachQuit:
		m.quit = true
	default:
		m.Reply(ErrorReply{
			Err: fmt.Errorf("unknown command: %v", reflect.TypeOf(c)),
		})
	}
}

func (m *Mach) setStatus(s Status) {
	m.Status = s
	m.Reply(StatusReply{Status: s})
}

var AsciiDecoder = func(code uint8) (rune, bool) {
	printable := code >= 32 && code < 128
	return rune(code), printable
}
