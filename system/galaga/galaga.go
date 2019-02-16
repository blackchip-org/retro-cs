// Package galaga is the hardware cabinet for Galaga.
package galaga

import (
	"github.com/blackchip-org/retro-cs/config"
	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/blackchip-org/retro-cs/rcs/namco"
	"github.com/blackchip-org/retro-cs/rcs/z80"
)

type system struct {
	cpu   [3]*z80.CPU
	mem   [3]*rcs.Memory
	ram   []uint8
	n06xx *namco.N06XX
	n51xx *namco.N51XX
	n54xx *namco.N54XX

	video *namco.Video

	interruptEnable1 uint8 // low bit
	interruptEnable2 uint8 // low bit
	interruptEnable3 uint8 // low bit
	dipSwitches      [8]uint8
}

func new(ctx rcs.SDLContext, set []rcs.ROM) (*rcs.Mach, error) {
	s := &system{}
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
	mem.MapRW(0x6820, &s.interruptEnable1)
	mem.MapRW(0x6821, &s.interruptEnable2)
	mem.MapRW(0x6822, &s.interruptEnable3)

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
	mem1 := rcs.NewMemory(1, 0x10000)
	mem1.Map(0, mem)
	mem1.MapROM(0x0000, roms["code1"])

	mem2 := rcs.NewMemory(1, 0x10000)
	mem2.Map(0, mem)
	mem2.MapROM(0x0000, roms["code2"])

	mem3 := rcs.NewMemory(1, 0x10000)
	mem3.Map(0, mem)
	mem3.MapROM(0x0000, roms["code3"])

	cpu1 := z80.New(mem1)
	cpu2 := z80.New(mem2)
	cpu3 := z80.New(mem3)

	vblank := func() {
		if s.interruptEnable1 != 0 {
			cpu1.IRQ = true
		}
		if s.interruptEnable2 != 0 {
			cpu2.IRQ = true
		}
	}

	s.n06xx.NMI = func() {
		cpu1.NMI = true
	}

	mach := &rcs.Mach{
		Sys:  s,
		Mem:  []*rcs.Memory{mem1, mem2, mem3},
		CPU:  []rcs.CPU{cpu1, cpu2, cpu3},
		Proc: []rcs.Proc{s.n06xx},
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
