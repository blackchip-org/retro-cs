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

	intSelect       uint8 // value sent during interrupt to select vector (port 0)
	in0             uint8 // joystick and coin slot
	interruptEnable uint8
	soundEnable     uint8
	unknown0        uint8
	flipScreen      uint8
	lampPlayer1     uint8
	lampPlayer2     uint8
	coinLockout     uint8
	coinCounter     uint8
	in1             uint8
	dipSwitches     uint8
	watchdogReset   uint8
}

func New(ctx rcs.SDLContext) (*rcs.Mach, error) {
	sys := &system{}
	roms, err := rcs.LoadROMs(config.ROMDir, ROM["pacman"])
	if err != nil {
		return nil, err
	}

	mem := rcs.NewMemory(1, 0x10000)
	ram := make([]uint8, 0x1000, 0x1000)

	mem.MapROM(0x0000, roms["code"])
	mem.MapRAM(0x4000, ram)
	mem.MapWO(0x5000, &sys.interruptEnable)
	mem.MapWO(0x5001, &sys.soundEnable)
	mem.MapWO(0x5002, &sys.unknown0)
	mem.MapRW(0x5003, &sys.flipScreen)
	mem.MapRW(0x5004, &sys.lampPlayer1)
	mem.MapRW(0x5005, &sys.lampPlayer2)
	mem.MapRW(0x5006, &sys.coinLockout)
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
	// Pacman is missing address line A15 so an access to $c000 is the
	// same as accessing $4000. Ms. Pacman has additional ROMs in high
	// memory so it has an A15 line but it appears to have the RAM mapped at
	// $c000 as well. Text for HIGH SCORE and CREDIT accesses this high memory
	// when writing to video memory. Copy protection?
	mem.MapRAM(0xc000, ram)

	cpu := z80.New(mem)
	cpu.Ports.MapRW(0x00, &sys.intSelect)

	var screen rcs.Screen
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
		screen = rcs.Screen{
			W:         namco.W,
			H:         namco.H,
			Texture:   video.Texture,
			ScanLineV: true,
			Draw:      video.Draw,
		}
	}

	if ctx.AudioSpec.Channels > 0 {
		data := audioData{
			waveforms: roms["waveform"],
		}
		audio, err := newAudio(ctx.AudioSpec, data)
		if err != nil {
			return nil, err
		}
		mem.MapRAM(0x5040, audio.voices[0].acc)
		mem.MapRW(0x5045, &audio.voices[0].waveform)
		mem.MapRAM(0x5046, audio.voices[1].acc)
		mem.MapRW(0x504a, &audio.voices[1].waveform)
		mem.MapRAM(0x504b, audio.voices[2].acc)
		mem.MapRW(0x504f, &audio.voices[2].waveform)

		mem.MapRAM(0x5050, audio.voices[0].freq)
		mem.MapRW(0x5055, &audio.voices[0].vol)
		mem.MapRAM(0x5056, audio.voices[1].freq)
		mem.MapRW(0x505a, &audio.voices[1].vol)
		mem.MapRAM(0x505b, audio.voices[2].freq)
		mem.MapRW(0x505f, &audio.voices[2].vol)
	}

	// FIXME: this turns the joystick "off", etc.
	// Game does not work unless this is set!
	sys.in0 = 0x3f
	sys.in1 = 0x7f

	sys.in0 |= (1 << 7)          // Service button released
	sys.in1 |= (1 << 4)          // Board test switch disabled
	sys.in1 |= (1 << 7)          // // Upright cabinet
	sys.dipSwitches |= (1 << 0)  // 1 coin per game
	sys.dipSwitches &^= (1 << 1) // ...
	sys.dipSwitches |= (1 << 3)  // 3 lives
	sys.dipSwitches |= (1 << 7)  // Normal ghost names

	vblank := func() {
		if sys.interruptEnable != 0 {
			cpu.IRQ = true
			cpu.IRQData = sys.intSelect
		}
	}

	mach := &rcs.Mach{
		Mem: []*rcs.Memory{mem},
		CPU: []rcs.CPU{cpu},
		CharDecoders: map[string]rcs.CharDecoder{
			"pacman": PacmanDecoder,
		},
		Ctx:        ctx,
		Screen:     screen,
		VBlankFunc: vblank,
	}

	return mach, nil
}
