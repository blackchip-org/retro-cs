package rcs

import (
	"fmt"
	"image/color"
	"math"

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

	colorN := len(palettes) * len(palettes[0])
	per := int32(math.Sqrt(float64(colorN)))

	texW := per * tileW
	texH := per * tileH

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

func NewScanLinesV(r *sdl.Renderer, w int32, h int32, size int32) (*sdl.Texture, error) {
	tex, err := r.CreateTexture(sdl.PIXELFORMAT_RGBA8888,
		sdl.TEXTUREACCESS_TARGET, w, h)
	if err != nil {
		return nil, err
	}

	r.SetRenderTarget(tex)
	for y := int32(0); y < h; y++ {
		for x := int32(0); x < w; x += 2 * size {
			r.SetDrawColorArray(0, 0, 0, 0)
			for i := int32(0); i < size; i++ {
				r.DrawPoint(x+i, y)
			}
			r.SetDrawColorArray(0, 0, 0, 0x20)
			for i := int32(size); i < size*2; i++ {
				r.DrawPoint(x+i, y)
			}
		}
	}
	tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	r.SetRenderTarget(nil)
	return tex, nil
}

func NewScanLinesH(r *sdl.Renderer, w int32, h int32, size int32) (*sdl.Texture, error) {
	tex, err := r.CreateTexture(sdl.PIXELFORMAT_RGBA8888,
		sdl.TEXTUREACCESS_TARGET, w, h)
	if err != nil {
		return nil, err
	}

	r.SetRenderTarget(tex)
	for x := int32(0); x < w; x++ {
		for y := int32(0); y < h; y += 2 * size {
			r.SetDrawColorArray(0, 0, 0, 0)
			for i := int32(0); i < size; i++ {
				r.DrawPoint(x, y+i)
			}
			r.SetDrawColorArray(0, 0, 0, 0x20)
			for i := int32(size); i < size*2; i++ {
				r.DrawPoint(x, y+i)
			}
		}
	}
	tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	r.SetRenderTarget(nil)
	return tex, nil
}
