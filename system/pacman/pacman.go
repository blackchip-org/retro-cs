// Package pacman is the hardware cabinet for Pac-Man and Ms. Pac-Man.
package pacman

import (
	"github.com/blackchip-org/retro-cs/config"
	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/blackchip-org/retro-cs/rcs/namco"
	"github.com/blackchip-org/retro-cs/rcs/z80"
)

type system struct {
	ram []uint8
	io  []uint8

	intSelect       uint8 // value sent during interrupt to select vector (port 0)
	in0             int8  // joystick and coin slot
	interruptEnable uint8
	coinCounter     uint8
	in1             uint8
	dipSwitches     uint8
	watchdogReset   uint8
}

func New(ctx rcs.SDLContext) (*rcs.Mach, error) {
	roms, err := rcs.LoadROMs(config.ROMDir, ROM["pacman"])
	if err != nil {
		return nil, err
	}
	mem := rcs.NewMemory(1, 0x10000)
	ram := make([]uint8, 0x1000, 0x1000)
	mem.MapROM(0x0000, roms["code"])
	mem.MapRAM(0x4000, ram)
	// Pacman is missing address line A15 so an access to $c000 is the
	// same as accessing $4000. Ms. Pacman has additional ROMs in high
	// memory so it has an A15 line but it appears to have the RAM mapped at
	// $c000 as well. Text for HIGH SCORE and CREDIT accesses this high memory
	// when writing to video memory. Copy protection?
	mem.MapRAM(0xc000, ram)

	cpu := z80.New(mem)

	sys := &system{}
	mach := &rcs.Mach{
		Mem: []*rcs.Memory{mem},
		CPU: []rcs.CPU{cpu},
		CharDecoders: map[string]rcs.CharDecoder{
			"pacman": PacmanDecoder,
		},
		Ctx: ctx,
	}

	mach.VBlankFunc = func() {
		if sys.interruptEnable != 0 {
			//cpu.INT(sys.intSelect)
		}
	}

	mem.MapWO(0x5000, &sys.interruptEnable)
	mem.MapRW(0x5007, &sys.coinCounter)
	for i := 0x5040; i <= 0x507f; i++ {
		mem.MapRO(i, &sys.in1)
	}
	for i := 0x5080; i <= 0x50bf; i++ {
		mem.MapRO(i, &sys.dipSwitches)
	}
	for i := 0x50c0; i <= 0x50ff; i++ {
		mem.MapWO(i, &sys.watchdogReset)
	}

	// FIXME: this turns the joystick "off", etc.
	// Game does not work unless this is set!
	sys.in0 = 0x3f
	sys.in1 = 0x7f

	if ctx.Renderer != nil {
		data := namco.Data{
			Palettes: roms["palette"],
			Colors:   roms["color"],
			Tiles:    roms["tile"],
			Sprites:  roms["sprite"],
		}
		video, err := newVideo(ctx.Renderer, data)
		if err != nil {
			return nil, err
		}
		mem.MapRAM(0x4000, video.TileMemory)
		mem.MapRAM(0x4400, video.ColorMemory)
		for i := 0; i < 8; i++ {
			mem.MapRW(0x5060+(i*2), &video.SpriteCoords[i].X)
			mem.MapRW(0x5061+(i*2), &video.SpriteCoords[i].Y)
			mem.MapRW(0x4ff0+(i*2), &video.SpriteInfo[i])
			mem.MapRW(0x4ff1+(i*2), &video.SpritePalettes[i])
		}
		screen := rcs.Screen{
			W:         namco.W,
			H:         namco.H,
			Texture:   video.Texture,
			ScanLineV: true,
			Draw:      video.Draw,
		}
		mach.Screen = screen
	}

	return mach, nil
}
