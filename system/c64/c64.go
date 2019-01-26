package c64

import (
	"github.com/blackchip-org/retro-cs/config"
	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/blackchip-org/retro-cs/rcs/m6502"
)

func New() (*rcs.Mach, error) {
	roms, err := rcs.LoadROMs(config.DataDir, SystemROM)
	if err != nil {
		return nil, err
	}
	mem := newMemory(roms)
	mem.SetBank(31)
	cpu := m6502.New(mem)

	return &rcs.Mach{
		Mem: []*rcs.Memory{mem},
		CPU: []rcs.CPU{cpu},
	}, nil
}
