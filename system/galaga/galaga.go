// Package galaga is the hardware cabinet for Galaga.
package galaga

import (
	"github.com/blackchip-org/retro-cs/config"
	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/blackchip-org/retro-cs/rcs/namco"
	"github.com/blackchip-org/retro-cs/rcs/z80"
)

type System struct {
	cpu   [3]*z80.CPU
	mem   [3]*rcs.Memory
	ram   []uint8
	n06xx *namco.N06XX
	n51xx *namco.N51XX
	n54xx *namco.N54XX

	video *namco.Video

	InterruptEnable0 uint8 // low bit
	InterruptEnable1 uint8 // low bit
	InterruptEnable2 uint8 // low bit
	reset            uint8
	dipSwitches      [8]uint8
}

func new(ctx rcs.SDLContext, set []rcs.ROM) (*rcs.Mach, error) {
	s := &System{}
	roms, err := rcs.LoadROMs(config.DataDir, set)
	if err != nil {
		return nil, err
	}

	// construct the common memory first
	mem := rcs.NewMemory(1, 0x10000)
	ram := make([]uint8, 0x2000, 0x2000)

	mem.MapRAM(0x6800, make([]uint8, 0x100, 0x100)) // temporary
	for i := 0; i < 8; i++ {
		mem.MapRW(0x6800+i, &s.dipSwitches[i])
	}
	mem.MapRW(0x6820, &s.InterruptEnable0)
	mem.MapRW(0x6821, &s.InterruptEnable1)
	mem.MapRW(0x6822, &s.InterruptEnable2)
	mem.MapRW(0x6823, &s.reset)

	mem.MapRAM(0x7000, make([]uint8, 0x1000, 0x1000))
	mem.MapRAM(0x8000, ram)
	mem.MapRAM(0xa000, make([]uint8, 0x1000, 0x1000))

	s.n51xx = namco.NewN51XX()
	s.n54xx = namco.NewN54XX()

	s.n06xx = namco.NewN06XX()
	s.n06xx.DeviceW[0] = s.n51xx.Write
	s.n06xx.DeviceR[0] = s.n51xx.Read
	s.n06xx.DeviceW[3] = s.n54xx.Write
	s.n06xx.DeviceR[3] = s.n54xx.Read
	for i, addr := 0, 0x7000; addr < 0x7100; addr, i = addr+1, i+1 {
		j := i
		mem.MapLoad(addr, s.n06xx.ReadData(j))
		mem.MapStore(addr, s.n06xx.WriteData(j))
	}
	for i, addr := 0, 0x7100; addr < 0x7200; addr, i = addr+1, i+1 {
		j := i
		mem.MapLoad(addr, s.n06xx.ReadCtrl(j))
		mem.MapStore(addr, s.n06xx.WriteCtrl(j))
	}

	var screen rcs.Screen
	var video *namco.Video
	if ctx.Renderer != nil {
		data := namco.Data{
			Palettes: roms["palettes"],
			Colors:   roms["colors"],
			Tiles:    roms["tiles"],
			Sprites:  roms["sprites"],
		}
		video, err = newVideo(ctx.Renderer, data)
		if err != nil {
			return nil, err
		}
		mem.MapRAM(0x8000, video.TileMemory)
		mem.MapRAM(0x8400, video.ColorMemory)

		screen = rcs.Screen{
			W:         namco.W,
			H:         namco.H,
			Texture:   video.Texture,
			ScanLineV: true,
			Draw:      video.Draw,
		}
	}

	s.dipSwitches[3] = 1
	s.dipSwitches[4] = 2
	s.dipSwitches[5] = 1
	s.dipSwitches[6] = 1

	// HACK
	mem.Write(0x9100, 0xff)
	mem.Write(0x9101, 0xff)

	// memory for each CPU
	s.mem[0] = rcs.NewMemory(1, 0x10000)
	s.mem[0].Map(0, mem)
	s.mem[0].MapROM(0x0000, roms["code1"])

	s.mem[1] = rcs.NewMemory(1, 0x10000)
	s.mem[1].Map(0, mem)
	s.mem[1].MapROM(0x0000, roms["code2"])

	s.mem[2] = rcs.NewMemory(1, 0x10000)
	s.mem[2].Map(0, mem)
	s.mem[2].MapROM(0x0000, roms["code3"])

	s.cpu[0] = z80.New(s.mem[0])
	s.cpu[0].Name = "cpu1"
	s.cpu[1] = z80.New(s.mem[1])
	s.cpu[1].Name = "cpu2"
	s.cpu[2] = z80.New(s.mem[2])
	s.cpu[2].Name = "cpu3"

	vblank := func() {
		if s.InterruptEnable0 != 0 {
			s.cpu[0].IRQ = true
		}
		if s.InterruptEnable1 != 0 {
			s.cpu[1].IRQ = true
		}
		s.cpu[2].IRQ = true
		if s.InterruptEnable2 != 0 {
			// FIXME: Is this correct??? Probably not
			s.cpu[2].NMI = true
		}
		if s.reset != 0 {
			s.reset = 0
			s.cpu[1].RESET = true
			s.cpu[2].RESET = true
		}
	}

	s.n06xx.NMI = func() {
		s.cpu[0].NMI = true
	}

	mach := &rcs.Mach{
		Sys: s,
		Comps: []rcs.Component{
			rcs.NewComponent("galaga", "galaga", "", s),
			rcs.NewComponent("mem1", "mem", "", s.mem[0]),
			rcs.NewComponent("mem2", "mem", "", s.mem[1]),
			rcs.NewComponent("mem3", "mem", "", s.mem[2]),
			rcs.NewComponent("cpu1", "z80", "mem1", s.cpu[0]),
			rcs.NewComponent("cpu2", "z80", "mem2", s.cpu[1]),
			rcs.NewComponent("cpu3", "z80", "mem3", s.cpu[2]),
			rcs.NewComponent("n06xx", "n06xx", "", s.n06xx),
			rcs.NewComponent("n51xx", "n51xx", "", s.n51xx),
			rcs.NewComponent("n54xx", "n54xx", "", s.n54xx),
		},
		CharDecoders: map[string]rcs.CharDecoder{
			"galaga": GalagaDecoder,
		},
		Ctx:        ctx,
		Screen:     screen,
		VBlankFunc: vblank,
	}
	return mach, nil
}

func New(ctx rcs.SDLContext) (*rcs.Mach, error) {
	return new(ctx, ROM["galaga"])
}
