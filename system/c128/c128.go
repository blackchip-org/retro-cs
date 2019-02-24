// Package c128 is the Commodore 128.
package c128

import (
	"github.com/blackchip-org/retro-cs/config"
	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/blackchip-org/retro-cs/rcs/cbm"
	"github.com/blackchip-org/retro-cs/rcs/m6502"
)

type System struct {
	cpu *m6502.CPU
	mem *rcs.Memory

	BasicLo []uint8
	BasicHi []uint8
	CharGen []uint8
	Kernal  []uint8
	RAM0    []uint8
	RAM1    []uint8
	IO      []uint8
}

func New(ctx rcs.SDLContext) (*rcs.Mach, error) {
	s := &System{}
	roms, err := rcs.LoadROMs(config.DataDir, SystemROM)
	if err != nil {
		return nil, err
	}
	s.BasicLo = roms["basiclo"]
	s.BasicHi = roms["basichi"]
	s.CharGen = roms["chargen"]
	s.Kernal = roms["kernal"]
	s.RAM0 = make([]uint8, 0x10000, 0x10000)
	s.RAM1 = make([]uint8, 0xc000, 0xc000)
	s.IO = make([]uint8, 0x1000, 0x1000)
	s.mem = newMemory(s)

	s.mem.SetBank(15)

	s.cpu = m6502.New(s.mem)

	mach := &rcs.Mach{
		Sys: s,
		Comps: []rcs.Component{
			rcs.NewComponent("cpu", "m6502", "mem", s.cpu),
			rcs.NewComponent("mem", "mem", "", s.mem),
		},
		CharDecoders: map[string]rcs.CharDecoder{
			"petscii":         cbm.PetsciiDecoder,
			"petscii-shifted": cbm.PetsciiShiftedDecoder,
			"screen":          cbm.ScreenDecoder,
			"screen-shifted":  cbm.ScreenShiftedDecoder,
		},
		DefaultEncoding: "petscii",
		Ctx:             ctx,
		VBlankFunc: func() {
			s.cpu.IRQ = true
		},
	}
	return mach, nil
}
