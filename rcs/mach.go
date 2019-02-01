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

type Mach struct {
	Mem []*Memory
	CPU []CPU

	Status        Status
	EventCallback func(EventType, interface{})
	Breakpoints   []map[int]struct{}

	quit bool
	cmd  chan interface{}
}

type MachStart struct{}
type MachQuit struct{}

type EventType int

const (
	StatusEvent EventType = iota
	TraceEvent
	ErrorEvent
)

func (m *Mach) init() {
	m.quit = false
	if m.EventCallback == nil {
		m.EventCallback = func(EventType, interface{}) {}
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

func (m *Mach) command(mc interface{}) {
	switch mc.(type) {
	case MachStart:
		m.setStatus(Run)
	case MachQuit:
		m.quit = true
	default:
		m.EventCallback(ErrorEvent, fmt.Errorf("invalid command: %v", mc))
	}
}

func (m *Mach) setStatus(s Status) {
	m.Status = s
	m.EventCallback(StatusEvent, s)
}
