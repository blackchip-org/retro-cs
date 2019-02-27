package rcs

import (
	"fmt"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"
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
	MachExport MachCmd = iota
	MachImport
	MachPause
	MachStart
	MachTrace
	MachTraceAll
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
	Sys             interface{}
	Comps           []Component
	CharDecoders    map[string]CharDecoder
	DefaultEncoding string
	Ctx             SDLContext
	Screen          Screen
	VBlankFunc      func()
	QueueAudio      func() error
	Keyboard        func(*sdl.KeyboardEvent) error

	CPU         map[string]CPU
	Proc        map[string]Proc
	Status      Status
	Callback    func(MachEvent, ...interface{})
	Breakpoints map[string]map[int]struct{}

	scanLines *sdl.Texture
	init      bool
	tracing   map[string]bool
	quit      bool
	cmd       chan message
}

func (m *Mach) Init() error {
	if m.init {
		return nil
	}
	m.CPU = make(map[string]CPU)
	m.Proc = make(map[string]Proc)
	m.tracing = make(map[string]bool)
	for _, comp := range m.Comps {
		switch v := comp.C.(type) {
		case CPU:
			m.CPU[comp.Name] = v
			m.tracing[comp.Name] = false
		case Proc:
			m.Proc[comp.Name] = v
		}
	}

	m.quit = false
	if m.CharDecoders == nil {
		m.CharDecoders = map[string]CharDecoder{
			"ascii": ASCIIDecoder,
		}
		m.DefaultEncoding = "ascii"
	}
	m.cmd = make(chan message, 10)

	m.Breakpoints = make(map[string]map[int]struct{})
	for name := range m.CPU {
		m.Breakpoints[name] = make(map[int]struct{})
	}
	if m.VBlankFunc == nil {
		m.VBlankFunc = func() {}
	}
	if m.Keyboard == nil {
		m.Keyboard = func(*sdl.KeyboardEvent) error { return nil }
	}

	if m.Ctx.Window != nil && m.Screen.W > 0 {
		r := m.Ctx.Renderer
		winx, winy := m.Ctx.Window.GetSize()
		FitInWindow(winx, winy, &m.Screen)
		drawW := m.Screen.W * m.Screen.Scale
		drawH := m.Screen.H * m.Screen.Scale
		if m.Screen.ScanLineH {
			scanLines, err := NewScanLinesH(r, drawW, drawH, 2)
			if err != nil {
				return err
			}
			m.scanLines = scanLines
		} else if m.Screen.ScanLineV {
			scanLines, err := NewScanLinesV(r, drawW, drawH, 2)
			if err != nil {
				return err
			}
			m.scanLines = scanLines
		}
	}
	m.init = true
	return nil
}

func (m *Mach) Run() error {
	if err := m.Init(); err != nil {
		return err
	}
	ticker := time.NewTicker(vblank)
	panicked := true
	defer func() {
		if panicked {
			m.reportCrash()
		}
	}()
	for {
		select {
		case c := <-m.cmd:
			m.handleCommand(c)
		case <-ticker.C:
			m.jiffy()
		}
		if m.quit {
			break
		}
	}
	panicked = false
	return nil
}

func (m *Mach) Command(cmd MachCmd, args ...interface{}) {
	m.cmd <- message{Cmd: cmd, Args: args}
}

func (m *Mach) jiffy() {
	if m.Status == Run {
		m.execute()
	}
	if m.QueueAudio != nil && m.Ctx.AudioSpec.Freq != 0 {
		if err := m.QueueAudio(); err != nil {
			m.event(ErrorEvent, err)
		}
	}
	if m.Ctx.Renderer != nil && m.Screen.W > 0 {
		m.render()
	} else {
		time.Sleep(10 * time.Millisecond)
	}
	m.sdl()
	if m.Status == Run {
		m.VBlankFunc()
	}
}

func (m *Mach) execute() {
	for t := 0; t < perJiffy; t++ {
		for name, cpu := range m.CPU {
			ppc := cpu.PC()
			cpu.Next()
			// if the program counter didn't change, it is either stuck
			// in an infinite loop or not advancing due to a halt-like
			// instruction
			stuck := ppc == cpu.PC()
			if m.tracing[name] && !stuck {
				m.event(TraceEvent, name, ppc)
			}
			// at a breakpoint? only honor it if the processor is not stuck.
			// when at a halt-like instruction, this causes a break once
			// instead of each time.
			addr := cpu.PC() + cpu.Offset()
			if _, yes := m.Breakpoints[name][addr]; yes && !stuck {
				m.setStatus(Break)
				return
			}
		}
		for _, proc := range m.Proc {
			proc.Next()
		}
	}
}

func (m *Mach) render() error {
	r := m.Ctx.Renderer
	if err := m.Screen.Draw(r); err != nil {
		return err
	}
	dest := sdl.Rect{
		X: m.Screen.X,
		Y: m.Screen.Y,
		W: m.Screen.W * m.Screen.Scale,
		H: m.Screen.H * m.Screen.Scale,
	}
	r.Copy(m.Screen.Texture, nil, &dest)
	if m.scanLines != nil {
		r.Copy(m.scanLines, nil, &dest)
	}
	r.Present()
	return nil
}

func (m *Mach) sdl() {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		if _, ok := event.(*sdl.QuitEvent); ok {
			m.quit = true
		} else if e, ok := event.(*sdl.KeyboardEvent); ok {
			if e.Keysym.Sym == sdl.K_ESCAPE {
				m.quit = true
			} else {
				m.Keyboard(e)
			}
		}
	}
}

func (m *Mach) handleCommand(msg message) {
	switch msg.Cmd {
	case MachExport:
		m.cmdExport(msg.Args...)
	case MachImport:
		m.cmdImport(msg.Args...)
	case MachPause:
		m.setStatus(Pause)
	case MachStart:
		m.setStatus(Run)
	case MachTrace:
		m.cmdTrace(msg.Args...)
	case MachTraceAll:
		m.cmdTraceAll(msg.Args...)
	case MachQuit:
		m.quit = true
	default:
		m.event(ErrorEvent, fmt.Errorf("unknown command: %v", msg.Cmd))
	}
}

func (m *Mach) cmdExport(args ...interface{}) {
	sys, ok := m.Sys.(Saver)
	if !ok {
		m.event(ErrorEvent, "exporting is not supported")
		return
	}
	filename := args[0].(string)
	out, err := os.Create(filename)
	if err != nil {
		m.event(ErrorEvent, fmt.Sprintf("unable toe xport: %v", err))
		return
	}
	enc := NewEncoder(out)
	sys.Save(enc)
	if enc.Err != nil {
		m.event(ErrorEvent, fmt.Sprintf("unable to export: %v", enc.Err))
		return
	}
}

func (m *Mach) cmdImport(args ...interface{}) {
	sys, ok := m.Sys.(Loader)
	if !ok {
		m.event(ErrorEvent, "importing is not supported")
		return
	}
	filename := args[0].(string)
	in, err := os.Open(filename)
	if err != nil {
		m.event(ErrorEvent, fmt.Sprintf("unable to import: %v", err))
		return
	}
	dec := NewDecoder(in)
	sys.Load(dec)
	if dec.Err != nil {
		m.event(ErrorEvent, fmt.Sprintf("unable to import: %v", dec.Err))
		return
	}
}

func (m *Mach) cmdTrace(args ...interface{}) {
	name := args[0].(string)
	if len(args) == 1 {
		m.tracing[name] = !m.tracing[name]
		return
	}
	v, ok := args[1].(bool)
	if !ok {
		m.event(ErrorEvent, fmt.Sprintf("invalid trace mode: %v", args[0]))
		return
	}
	m.tracing[name] = v
}

func (m *Mach) cmdTraceAll(args ...interface{}) {
	v, ok := args[0].(bool)
	if !ok {
		m.event(ErrorEvent, fmt.Sprintf("invalid trace mode: %v", args[0]))
		return
	}
	for name := range m.tracing {
		m.tracing[name] = v
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

func (m *Mach) reportCrash() {
	for n, c := range m.CPU {
		fmt.Printf("[panic: %v]\n", n)
		fmt.Println(c)
	}
}
