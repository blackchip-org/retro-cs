// Package c64 is the Commodore 64.
package c64

import (
	"github.com/blackchip-org/retro-cs/config"
	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/blackchip-org/retro-cs/rcs/cbm"
	"github.com/blackchip-org/retro-cs/rcs/m6502"
)

type system struct {
	cpu  *m6502.CPU
	ram  []uint8
	io   []uint8
	bank uint8
}

func New(ctx rcs.SDLContext) (*rcs.Mach, error) {
	sys := &system{}
	roms, err := rcs.LoadROMs(config.DataDir, SystemROM)
	if err != nil {
		return nil, err
	}
	ram := make([]uint8, 0x10000, 0x10000)
	io := make([]uint8, 0x1000, 0x1000)

	mem := newMemory(ram, io, roms)
	// setup IO port on the 6510, map address 1 to "PLA" in all banks
	for b := 0; b < 32; b++ {
		mem.SetBank(b)
		mem.MapLoad(1, sys.ioPortLoad(mem))
		mem.MapStore(1, sys.ioPortStore(mem))
	}
	// GAME and EXROM on to start
	sys.bank = 0x18
	// HIMEM, LOMEM, CHAREN on to start
	mem.Write(1, 0x7)

	cpu := m6502.New(mem)

	var screen rcs.Screen
	if ctx.Renderer != nil {
		video, err := newVideo(ctx.Renderer, mem, roms["chargen"])
		if err != nil {
			return nil, err
		}
		mem.MapRW(0xd020, &video.borderColor)
		mem.MapRW(0xd021, &video.bgColor)
		screen = rcs.Screen{
			W:         screenW,
			H:         screenH,
			Texture:   video.texture,
			ScanLineH: true,
			Draw:      video.draw,
		}
	}

	kb := newKeyboard()
	mem.MapRW(0x0091, &kb.stkey) // stop key
	mem.MapRW(0x00c6, &kb.ndx)   // buffer index
	mem.MapRAM(0x277, kb.buf)

	sys.cpu = cpu
	sys.ram = ram
	sys.io = io

	mach := &rcs.Mach{
		Sys: sys,
		Mem: []*rcs.Memory{mem},
		CPU: []rcs.CPU{cpu},
		CharDecoders: map[string]rcs.CharDecoder{
			"petscii":         cbm.PetsciiDecoder,
			"petscii-shifted": cbm.PetsciiShiftedDecoder,
			"screen":          cbm.ScreenDecoder,
			"screen-shifted":  cbm.ScreenShiftedDecoder,
		},
		DefaultEncoding: "petscii",
		Ctx:             ctx,
		VBlankFunc: func() {
			cpu.IRQ = true
		},
		Screen:   screen,
		Keyboard: kb.handle,
	}

	return mach, nil
}

func (s *system) ioPortStore(mem *rcs.Memory) rcs.Store8 {
	return func(v uint8) {
		// PLA information is in the bottom 3 bits
		s.bank &^= 0x7
		s.bank |= v & 0x7
		mem.SetBank(int(s.bank))
	}
}

func (s *system) ioPortLoad(mem *rcs.Memory) rcs.Load8 {
	return func() uint8 {
		// Only return the bottom 3 bits for now
		return s.bank & 0x7
	}
}

func (s *system) Save(enc *rcs.Encoder) {
	s.cpu.Save(enc)
	enc.Encode(s.ram)
	enc.Encode(s.io)
}

func (s *system) Load(dec *rcs.Decoder) {
	s.cpu.Load(dec)
	dec.Decode(&s.ram)
	dec.Decode(&s.io)
}
