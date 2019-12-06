// Package c64 is the Commodore 64.
package c64

import (
	"github.com/blackchip-org/retro-cs/config"
	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/blackchip-org/retro-cs/rcs/cbm"
	"github.com/blackchip-org/retro-cs/rcs/cbm/petscii"
	"github.com/blackchip-org/retro-cs/rcs/m6502"
)

type system struct {
	cpu    *m6502.CPU
	mem    *rcs.Memory
	screen rcs.Screen
	vic    *cbm.VIC
	ram    []uint8
	io     []uint8
	bank   uint8
}

func New(ctx rcs.SDLContext) (*rcs.Mach, error) {
	s := &system{}
	roms, err := rcs.LoadROMs(config.DataDir, SystemROM)
	if err != nil {
		return nil, err
	}
	s.ram = make([]uint8, 0x10000, 0x10000)
	s.io = make([]uint8, 0x1000, 0x1000)

	s.mem = newMemory(s.ram, s.io, roms)
	kb := newKeyboard()

	v, err := cbm.NewVIC(ctx.Renderer, s.mem, roms["chargen"])
	if err != nil {
		return nil, err
	}
	s.vic = v
	s.screen = rcs.Screen{
		W:         v.W,
		H:         v.H,
		Texture:   v.Texture,
		ScanLineH: true,
		Draw:      v.Draw,
	}

	for b := 0; b < 32; b++ {
		s.mem.SetBank(b)
		// setup IO port on the 6510, map address 1 to "PLA"s
		s.mem.MapLoad(0x01, s.ioPortLoad)
		s.mem.MapStore(0x01, s.ioPortStore)

		s.mem.MapRW(0xd020, &s.vic.BorderColor)
		s.mem.MapRW(0xd021, &s.vic.BgColor)

		s.mem.MapRW(0x0091, &kb.stkey) // stop key
		s.mem.MapRW(0x00c6, &kb.ndx)   // buffer index
		s.mem.MapRAM(0x0277, kb.buf)

		s.mem.MapRW(0xdc00, &kb.joy2)
	}
	// Initialize to bank 31
	s.mem.SetBank(31)
	// GAME and EXROM on to start
	s.bank = 0x18
	// HIMEM, LOMEM, CHAREN on to start
	s.mem.Write(1, 0x7)

	// CPU should be created after memory is completely setup to obtain
	// the correct reset vector
	s.cpu = m6502.New(s.mem)

	mach := &rcs.Mach{
		Sys: s,
		Comps: []rcs.Component{
			rcs.NewComponent("c64", "c64", "", s),
			rcs.NewComponent("cpu", "m6502", "mem", s.cpu),
			rcs.NewComponent("mem", "mem", "", s.mem),
		},
		CharDecoders: map[string]rcs.CharDecoder{
			"petscii":         petscii.Decoder,
			"petscii-shifted": petscii.ShiftedDecoder,
			"screen":          cbm.ScreenDecoder,
			"screen-shifted":  cbm.ScreenShiftedDecoder,
		},
		DefaultEncoding: "petscii",
		Ctx:             ctx,
		VBlankFunc: func() {
			s.cpu.IRQ = true
		},
		Screen:   s.screen,
		Keyboard: kb.handle,
	}

	return mach, nil
}

func (s *system) ioPortStore(v uint8) {
	// PLA information is in the bottom 3 bits
	s.bank &^= 0x7
	s.bank |= v & 0x7
	s.mem.SetBank(int(s.bank))
}

func (s *system) ioPortLoad() uint8 {
	// Only return the bottom 3 bits for now
	return s.bank & 0x7
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
