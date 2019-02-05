package c64

import (
	"github.com/blackchip-org/retro-cs/config"
	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/blackchip-org/retro-cs/rcs/cbm"
	"github.com/blackchip-org/retro-cs/rcs/m6502"
)

func New(ctx rcs.SDLContext) (*rcs.Mach, error) {
	roms, err := rcs.LoadROMs(config.ROMDir, SystemROM)
	if err != nil {
		return nil, err
	}
	mem := newMemory(roms)
	mem.SetBank(31)
	cpu := m6502.New(mem)

	mach := &rcs.Mach{
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
	}

	if ctx.Renderer != nil {
		video, err := NewVideo(ctx.Renderer, mem, roms["chargen"])
		if err != nil {
			return nil, err
		}
		mem.MapRW(0xd020, &video.borderColor)
		mem.MapRW(0xd021, &video.bgColor)
		screen := rcs.Screen{
			W:         screenW,
			H:         screenH,
			Texture:   video.texture,
			ScanLineH: true,
			Draw:      video.draw,
		}
		mach.Screen = screen
	}

	return mach, nil
}
