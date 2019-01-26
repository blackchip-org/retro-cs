package rcs

import (
	"fmt"
	"image/color"

	"github.com/veandco/go-sdl2/sdl"
)

type TileSheet struct {
	TextureW int32
	TextureH int32
	TileW    int32
	TileH    int32
	Texture  *sdl.Texture
}

func NewColorSheet(r *sdl.Renderer, palettes [][]color.RGBA) (TileSheet, error) {
	tileW := int32(32)
	tileH := int32(32)
	texW := 16 * tileW
	texH := 16 * tileH

	t, err := r.CreateTexture(sdl.PIXELFORMAT_RGBA8888,
		sdl.TEXTUREACCESS_TARGET, texW, texH)
	if err != nil {
		return TileSheet{}, fmt.Errorf("unable to create sheet: %v", err)
	}
	r.SetRenderTarget(t)

	x := int32(0)
	y := int32(0)
	for _, pal := range palettes {
		for _, c := range pal {
			r.SetDrawColor(c.R, c.G, c.B, c.A)
			r.FillRect(&sdl.Rect{
				X: x,
				Y: y,
				W: tileW,
				H: tileH,
			})
			x += tileW
			if x >= texW {
				x = 0
				y += tileH
			}
		}
	}
	r.SetRenderTarget(nil)
	return TileSheet{
		TextureW: texW,
		TextureH: texH,
		TileW:    tileW,
		TileH:    tileH,
		Texture:  t,
	}, nil
}
