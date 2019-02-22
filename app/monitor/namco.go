package monitor

import (
	"fmt"

	"github.com/chzyer/readline"

	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/blackchip-org/retro-cs/rcs/namco"
)

type n06XX struct {
	mon   *Monitor
	n06xx *namco.N06XX
}

func newN06XX(mon *Monitor, comp *rcs.Component) module {
	return &n06XX{n06xx: comp.C.(*namco.N06XX)}
}

func (n *n06XX) Command(args []string) error {
	if err := checkLen(args, 1, maxArgs); err != nil {
		return err
	}
	switch args[0] {
	case "debug-data-write":
		return n.mon.valueBool(&n.n06xx.DebugDataW, args[1:])
	case "debug-data-read":
		return n.mon.valueBool(&n.n06xx.DebugDataR, args[1:])
	case "debug-control-write":
		return n.mon.valueBool(&n.n06xx.DebugCtrlW, args[1:])
	case "debug-control-read":
		return n.mon.valueBool(&n.n06xx.DebugCtrlR, args[1:])
	case "debug-nmi":
		return n.mon.valueBool(&n.n06xx.DebugNMI, args[1:])
	case "debug-all":
		return n.mon.terminal(args[1:], func() error {
			n.n06xx.DebugDataW = true
			n.n06xx.DebugDataR = true
			n.n06xx.DebugCtrlW = true
			n.n06xx.DebugCtrlR = true
			n.n06xx.DebugNMI = true
			return nil
		})
	case "debug-none":
		return n.mon.terminal(args[1:], func() error {
			n.n06xx.DebugDataW = false
			n.n06xx.DebugDataR = false
			n.n06xx.DebugCtrlW = false
			n.n06xx.DebugCtrlR = false
			n.n06xx.DebugNMI = false
			return nil
		})
	}
	return fmt.Errorf("no such command: %v", args[0])
}

func (n *n06XX) AutoComplete() []readline.PrefixCompleterInterface {
	return []readline.PrefixCompleterInterface{
		readline.PcItem("debug-data-write"),
		readline.PcItem("debug-data-read"),
		readline.PcItem("debug-control-write"),
		readline.PcItem("debug-control-read"),
		readline.PcItem("debug-nmi"),
		readline.PcItem("debug-all"),
		readline.PcItem("debug-none"),
	}
}
