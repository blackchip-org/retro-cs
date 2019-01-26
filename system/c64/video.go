package c64

import (
	"image/color"

	"github.com/blackchip-org/retro-cs/rcs"

	"github.com/veandco/go-sdl2/sdl"
)

var (
	Palette = []color.RGBA{
		color.RGBA{0x00, 0x00, 0x00, 0xff}, // black
		color.RGBA{0xff, 0xff, 0xff, 0xff}, // white
		color.RGBA{0x88, 0x00, 0x00, 0xff}, // red
		color.RGBA{0xaa, 0xff, 0xee, 0xff}, // cyan
		color.RGBA{0xcc, 0x44, 0xcc, 0xff}, // purple
		color.RGBA{0x00, 0xcc, 0x55, 0xff}, // green
		color.RGBA{0x00, 0x00, 0xaa, 0xff}, // blue
		color.RGBA{0xee, 0xee, 0x77, 0xff}, // yellow
		color.RGBA{0xdd, 0x88, 0x55, 0xff}, // orange
		color.RGBA{0x66, 0x44, 0x00, 0xff}, // brown
		color.RGBA{0xff, 0x77, 0x77, 0xff}, // light red
		color.RGBA{0x33, 0x33, 0x33, 0xff}, // dark gray
		color.RGBA{0x77, 0x77, 0x77, 0xff}, // gray
		color.RGBA{0xaa, 0xff, 0x66, 0xff}, // light green
		color.RGBA{0x00, 0x88, 0xff, 0xff}, // light blue
		color.RGBA{0xbb, 0xbb, 0xbb, 0xff}, // light gray
	}
)

func CharGen(r *sdl.Renderer, data []uint8) (rcs.TileSheet, error) {
	tileW, tileH := int32(8), int32(8)
	texW := tileW * 32
	texH := tileH * 16
	t, err := r.CreateTexture(sdl.PIXELFORMAT_RGBA8888,
		sdl.TEXTUREACCESS_TARGET, texW, texH)
	if err != nil {
		return rcs.TileSheet{}, err
	}
	r.SetRenderTarget(t)
	r.SetDrawColor(0, 0, 0, 0xff)
	r.FillRect(nil)
	r.SetDrawColor(0xff, 0xff, 0xff, 0xff)
	baseX := int32(0)
	baseY := int32(0)
	addr := 0
	for baseY < texH {
		for y := baseY; y < baseY+8; y++ {
			line := data[addr]
			addr++
			for x := baseX; x < baseX+8; x++ {
				bit := line & 0x80
				line = line << 1
				if bit != 0 {
					r.DrawPoint(x, y)
				}
			}
		}
		baseX += 8
		if baseX >= texW {
			baseX = 0
			baseY += 8
		}
	}
	t.SetBlendMode(sdl.BLENDMODE_BLEND)
	r.SetRenderTarget(nil)
	return rcs.TileSheet{
		TextureW: texW,
		TextureH: texH,
		TileW:    tileW,
		TileH:    tileH,
		Texture:  t,
	}, nil
}
