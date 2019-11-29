// Package pacman is the hardware cabinet for Pac-Man and Ms. Pac-Man.
package pacman

import (
	"github.com/blackchip-org/retro-cs/config"
	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/blackchip-org/retro-cs/rcs/namco"
	"github.com/blackchip-org/retro-cs/rcs/z80"
)

type system struct {
	cpu   *z80.CPU
	mem   *rcs.Memory
	ram   []uint8
	video *namco.Video

	intSelect       uint8 // value sent during interrupt to select vector (port 0)
	in0             uint8 // joystick #1, rack advance, coin slot, service button
	interruptEnable uint8
	soundEnable     uint8
	unknown0        uint8
	flipScreen      uint8
	lampPlayer1     uint8
	lampPlayer2     uint8
	coinLockout     uint8
	coinCounter     uint8
	in1             uint8 // joystick #2, board test, start buttons, cabinet mode
	dipSwitches     uint8
	watchdogReset   uint8
}

func new(ctx rcs.SDLContext, set []rcs.ROM) (*rcs.Mach, error) {
	s := &system{}
	roms, err := rcs.LoadROMs(config.DataDir, set)
	if err != nil {
		return nil, err
	}

	s.mem = rcs.NewMemory(1, 0x10000)
	ram := make([]uint8, 0x1000, 0x1000)

	s.mem.MapROM(0x0000, roms["code"])
	s.mem.MapRAM(0x4000, ram)

	// Register range. Nil mappings first then add real mappings
	for i := 0x5000; i < 0x6000; i++ {
		s.mem.MapNil(i)
	}
	s.mem.MapWO(0x5000, &s.interruptEnable)
	for i := 0x5000; i < 0x503f; i++ {
		s.mem.MapRO(i, &s.in0)
	}
	s.mem.MapWO(0x5001, &s.soundEnable)
	s.mem.MapWO(0x5002, &s.unknown0)
	s.mem.MapRW(0x5003, &s.flipScreen)
	s.mem.MapRW(0x5004, &s.lampPlayer1)
	s.mem.MapRW(0x5005, &s.lampPlayer2)
	s.mem.MapRW(0x5006, &s.coinLockout)
	s.mem.MapRW(0x5007, &s.coinCounter)
	for i := 0x5040; i <= 0x507f; i++ {
		s.mem.MapRO(i, &s.in1)
	}
	for i := 0x5080; i <= 0x50bf; i++ {
		s.mem.MapRO(i, &s.dipSwitches)
	}
	for i := 0x50c0; i <= 0x50ff; i++ {
		s.mem.MapWO(i, &s.watchdogReset)
	}

	if code2, ok := roms["code2"]; ok {
		s.mem.MapROM(0x8000, code2)
	}

	// The first interrupt is executed without the stack pointer being set.
	// The machine attempts to write the return address to 0xffff and 0xfffe
	// but no memory is mapped there. Ms. Pac-Man also writes to 0xfffd
	// and 0xfffc. Remove this warning.
	s.mem.MapNil(0xfffc)
	s.mem.MapNil(0xfffd)
	s.mem.MapNil(0xfffe)
	s.mem.MapNil(0xffff)

	cpu := z80.New(s.mem)
	cpu.Ports.MapRW(0x00, &s.intSelect)

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
		s.mem.MapRAM(0x4000, video.TileMemory)
		s.mem.MapRAM(0x4400, video.ColorMemory)

		// Pacman is missing address line A15 so an access to $c000 is the
		// same as accessing $4000. Ms. Pacman has additional ROMs in high
		// memory so it has an A15 line but it appears to have the RAM mapped at
		// $c000 as well. Text for HIGH SCORE and CREDIT accesses this high
		// memory when writing to video memory. Copy protection?
		s.mem.MapRAM(0xc000, video.TileMemory)
		s.mem.MapRAM(0xc400, video.ColorMemory)

		for i := 0; i < 8; i++ {
			s.mem.MapRW(0x5060+(i*2), &video.SpriteCoords[i].X)
			s.mem.MapRW(0x5061+(i*2), &video.SpriteCoords[i].Y)
			s.mem.MapRW(0x4ff0+(i*2), &video.SpriteInfo[i])
			s.mem.MapRW(0x4ff1+(i*2), &video.SpritePalettes[i])
		}
		screen = rcs.Screen{
			W:         namco.W,
			H:         namco.H,
			Texture:   video.Texture,
			ScanLineV: true,
			Draw:      video.Draw,
		}
	}

	var synth *audio
	if ctx.AudioSpec.Channels > 0 {
		data := audioData{
			waveforms: roms["waveforms"],
		}
		synth, err = newAudio(ctx.AudioSpec, data)
		if err != nil {
			return nil, err
		}
		s.mem.MapWO(0x5040, &synth.voices[0].acc[0])
		s.mem.MapWO(0x5041, &synth.voices[0].acc[1])
		s.mem.MapWO(0x5042, &synth.voices[0].acc[2])
		s.mem.MapWO(0x5043, &synth.voices[0].acc[3])
		s.mem.MapWO(0x5044, &synth.voices[0].acc[4])
		s.mem.MapWO(0x5045, &synth.voices[0].waveform)
		s.mem.MapWO(0x5046, &synth.voices[1].acc[0])
		s.mem.MapWO(0x5047, &synth.voices[1].acc[1])
		s.mem.MapWO(0x5048, &synth.voices[1].acc[2])
		s.mem.MapWO(0x5049, &synth.voices[1].acc[3])
		s.mem.MapWO(0x504a, &synth.voices[1].waveform)
		s.mem.MapWO(0x504b, &synth.voices[2].acc[0])
		s.mem.MapWO(0x504c, &synth.voices[2].acc[1])
		s.mem.MapWO(0x504d, &synth.voices[2].acc[2])
		s.mem.MapWO(0x504e, &synth.voices[2].acc[3])
		s.mem.MapRW(0x504f, &synth.voices[2].waveform)

		s.mem.MapWO(0x5050, &synth.voices[0].freq[0])
		s.mem.MapWO(0x5051, &synth.voices[0].freq[1])
		s.mem.MapWO(0x5052, &synth.voices[0].freq[2])
		s.mem.MapWO(0x5053, &synth.voices[0].freq[3])
		s.mem.MapWO(0x5054, &synth.voices[0].freq[4])
		s.mem.MapWO(0x5055, &synth.voices[0].vol)
		s.mem.MapWO(0x5056, &synth.voices[1].freq[0])
		s.mem.MapWO(0x5057, &synth.voices[1].freq[1])
		s.mem.MapWO(0x5058, &synth.voices[1].freq[2])
		s.mem.MapWO(0x5059, &synth.voices[1].freq[3])
		s.mem.MapWO(0x505a, &synth.voices[1].vol)
		s.mem.MapWO(0x505b, &synth.voices[2].freq[0])
		s.mem.MapWO(0x505c, &synth.voices[2].freq[1])
		s.mem.MapWO(0x505d, &synth.voices[2].freq[2])
		s.mem.MapWO(0x505e, &synth.voices[2].freq[3])
		s.mem.MapRW(0x505f, &synth.voices[2].vol)
	}

	keyboard := newKeyboard(s)
	joystick := newJoystick(s)

	// Note: If in0 and in1 are not initialized to valid values, the
	// game will crash during the game demo in attract mode.

	// Joystick #1 in neutral position
	// Rack advance not pressed
	// Coin slots clear
	// Service button released
	s.in0 = 0xbf

	// Joystick #2 in neutral position
	// Board test off
	// Player start buttons released
	// Upright cabinet
	s.in1 = 0xff

	s.dipSwitches |= (1 << 0)  // 1 coin per game
	s.dipSwitches &^= (1 << 1) // ...
	s.dipSwitches |= (1 << 3)  // 3 lives
	s.dipSwitches |= (1 << 7)  // Normal ghost names

	vblank := func() {
		if s.interruptEnable != 0 {
			cpu.IRQ = true
			cpu.IRQData = s.intSelect
		}
	}

	s.cpu = cpu
	s.ram = ram
	s.video = video

	mach := &rcs.Mach{
		Sys: s,
		Comps: []rcs.Component{
			rcs.NewComponent("mem", "mem", "", s.mem),
			rcs.NewComponent("cpu", "z80", "mem", s.cpu),
		},
		CharDecoders: map[string]rcs.CharDecoder{
			"pacman": PacmanDecoder,
		},
		Ctx:           ctx,
		Screen:        screen,
		VBlankFunc:    vblank,
		QueueAudio:    synth.queue,
		Keyboard:      keyboard.handle,
		ButtonHandler: joystick.buttonHandler,
	}

	return mach, nil
}

func (s *system) Components() []*rcs.Component {
	return []*rcs.Component{}
}

func (s *system) Save(enc *rcs.Encoder) {
	s.cpu.Save(enc)
	if s.video != nil {
		s.video.Save(enc)
	}
	enc.Encode(s.ram)
	enc.Encode(s.intSelect)
	enc.Encode(s.in0)
	enc.Encode(s.interruptEnable)
	enc.Encode(s.soundEnable)
	enc.Encode(s.unknown0)
	enc.Encode(s.flipScreen)
	enc.Encode(s.lampPlayer1)
	enc.Encode(s.lampPlayer2)
	enc.Encode(s.coinLockout)
	enc.Encode(s.coinCounter)
	enc.Encode(s.in1)
	enc.Encode(s.dipSwitches)
	enc.Encode(s.watchdogReset)
}

func (s *system) Load(dec *rcs.Decoder) {
	s.cpu.Load(dec)
	if s.video != nil {
		s.video.Load(dec)
	}
	dec.Decode(&s.ram)
	dec.Decode(&s.intSelect)
	dec.Decode(&s.in0)
	dec.Decode(&s.interruptEnable)
	dec.Decode(&s.soundEnable)
	dec.Decode(&s.unknown0)
	dec.Decode(&s.flipScreen)
	dec.Decode(&s.lampPlayer1)
	dec.Decode(&s.lampPlayer2)
	dec.Decode(&s.coinLockout)
	dec.Decode(&s.coinCounter)
	dec.Decode(&s.in1)
	dec.Decode(&s.dipSwitches)
	dec.Decode(&s.watchdogReset)
}

func New(ctx rcs.SDLContext) (*rcs.Mach, error) {
	return new(ctx, ROM["pacman"])
}

func NewMs(ctx rcs.SDLContext) (*rcs.Mach, error) {
	return new(ctx, ROM["mspacman"])
}
