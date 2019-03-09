package monitor

import (
	"fmt"
	"log"

	"github.com/chzyer/readline"

	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/blackchip-org/retro-cs/system/galaga"
)

type modGalaga struct {
	mon    *Monitor
	out    *log.Logger
	galaga *galaga.System
}

func newModGalaga(mon *Monitor, comp rcs.Component) module {
	return &modGalaga{
		mon:    mon,
		out:    mon.out,
		galaga: comp.C.(*galaga.System),
	}
}

func (m *modGalaga) Command(args []string) error {
	if err := checkLen(args, 1, maxArgs); err != nil {
		return err
	}
	switch args[0] {
	case "interrupt-enable1":
		return valueBit(m.out, &m.galaga.InterruptEnable0, (1 << 0), args[1:])
	case "interrupt-enable2":
		return valueBit(m.out, &m.galaga.InterruptEnable1, (1 << 0), args[1:])
	case "interrupt-enable3":
		return valueBit(m.out, &m.galaga.InterruptEnable2, (1 << 0), args[1:])
	}
	return fmt.Errorf("no such command: %v", args[0])
}

func (m *modGalaga) AutoComplete() []readline.PrefixCompleterInterface {
	return []readline.PrefixCompleterInterface{
		readline.PcItem("interrupt-enable1"),
		readline.PcItem("interrupt-enable2"),
		readline.PcItem("interrupt-enable3"),
	}
}

func (m *modGalaga) Silence() error {
	return nil
}
