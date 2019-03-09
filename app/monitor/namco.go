package monitor

import (
	"fmt"
	"log"

	"github.com/chzyer/readline"

	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/blackchip-org/retro-cs/rcs/namco"
)

type modN06XX struct {
	mon   *Monitor
	out   *log.Logger
	n06xx *namco.N06XX
}

func newModN06XX(mon *Monitor, comp rcs.Component) module {
	return &modN06XX{
		mon:   mon,
		out:   mon.out,
		n06xx: comp.C.(*namco.N06XX),
	}
}

func (m *modN06XX) Command(args []string) error {
	if err := checkLen(args, 1, maxArgs); err != nil {
		return err
	}
	switch args[0] {
	case "watch-data-write":
		return valueBool(m.out, &m.n06xx.WatchDataW, args[1:])
	case "watch-data-read":
		return valueBool(m.out, &m.n06xx.WatchDataR, args[1:])
	case "watch-control-write":
		return valueBool(m.out, &m.n06xx.WatchCtrlW, args[1:])
	case "watch-control-read":
		return valueBool(m.out, &m.n06xx.WatchCtrlR, args[1:])
	case "watch-nmi":
		return valueBool(m.out, &m.n06xx.WatchNMI, args[1:])
	case "watch-all":
		return terminal(args[1:], func() error {
			m.n06xx.WatchDataW = true
			m.n06xx.WatchDataR = true
			m.n06xx.WatchCtrlW = true
			m.n06xx.WatchCtrlR = true
			m.n06xx.WatchNMI = true
			return nil
		})
	case "watch-none":
		return terminal(args[1:], m.Silence)
	}
	return fmt.Errorf("no such command: %v", args[0])
}

func (m *modN06XX) AutoComplete() []readline.PrefixCompleterInterface {
	return []readline.PrefixCompleterInterface{
		readline.PcItem("watch-data-write"),
		readline.PcItem("watch-data-read"),
		readline.PcItem("watch-control-write"),
		readline.PcItem("watch-control-read"),
		readline.PcItem("watch-nmi"),
		readline.PcItem("watch-all"),
		readline.PcItem("watch-none"),
	}
}

func (m *modN06XX) Silence() error {
	m.n06xx.WatchDataW = false
	m.n06xx.WatchDataR = false
	m.n06xx.WatchCtrlW = false
	m.n06xx.WatchCtrlR = false
	m.n06xx.WatchNMI = false
	return nil
}

type modN51XX struct {
	mon   *Monitor
	out   *log.Logger
	n51xx *namco.N51XX
}

func newModN51XX(mon *Monitor, comp rcs.Component) module {
	return &modN51XX{
		mon:   mon,
		out:   mon.out,
		n51xx: comp.C.(*namco.N51XX),
	}
}

func (m *modN51XX) Command(args []string) error {
	if err := checkLen(args, 1, maxArgs); err != nil {
		return err
	}
	switch args[0] {
	case "watch-write":
		return valueBool(m.out, &m.n51xx.WatchW, args[1:])
	case "watch-read":
		return valueBool(m.out, &m.n51xx.WatchR, args[1:])
	case "watch-all":
		return terminal(args[1:], func() error {
			m.n51xx.WatchW = true
			m.n51xx.WatchR = true
			return nil
		})
	case "watch-none":
		return terminal(args[1:], m.Silence)
	}
	return fmt.Errorf("no such command: %v", args[0])
}

func (m *modN51XX) AutoComplete() []readline.PrefixCompleterInterface {
	return []readline.PrefixCompleterInterface{
		readline.PcItem("watch-write"),
		readline.PcItem("watch-read"),
		readline.PcItem("watch-all"),
		readline.PcItem("watch-none"),
	}
}

func (m *modN51XX) Silence() error {
	m.n51xx.WatchW = false
	m.n51xx.WatchR = false
	return nil
}

type modN54XX struct {
	mon   *Monitor
	out   *log.Logger
	n54xx *namco.N54XX
}

func newModN54XX(mon *Monitor, comp rcs.Component) module {
	return &modN54XX{
		mon:   mon,
		out:   mon.out,
		n54xx: comp.C.(*namco.N54XX),
	}
}

func (m *modN54XX) Command(args []string) error {
	if err := checkLen(args, 1, maxArgs); err != nil {
		return err
	}
	switch args[0] {
	case "watch-write":
		return valueBool(m.out, &m.n54xx.WatchW, args[1:])
	case "watch-read":
		return valueBool(m.out, &m.n54xx.WatchR, args[1:])
	case "watch-all":
		return terminal(args[1:], func() error {
			m.n54xx.WatchW = true
			m.n54xx.WatchR = true
			return nil
		})
	case "watch-none":
		return terminal(args[1:], m.Silence)
	}
	return fmt.Errorf("no such command: %v", args[0])
}

func (m *modN54XX) AutoComplete() []readline.PrefixCompleterInterface {
	return []readline.PrefixCompleterInterface{
		readline.PcItem("watch-write"),
		readline.PcItem("watch-read"),
		readline.PcItem("watch-all"),
		readline.PcItem("watch-none"),
	}
}

func (m *modN54XX) Silence() error {
	m.n54xx.WatchW = false
	m.n54xx.WatchR = false
	return nil
}
