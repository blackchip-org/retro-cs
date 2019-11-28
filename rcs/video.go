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

type Screen struct {
	W         int32
	H         int32
	X         int32
	Y         int32
	Scale     int32
	Texture   *sdl.Texture
	ScanLineH bool
	ScanLineV bool
	Draw      func(*sdl.Renderer) error
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

	pixels := make([]uint32, w*h, w*h)
	for y := int32(0); y < h; y++ {
		for x := int32(0); x < w; x += 2 * size {
			ptr := (y * w) + x
			for i := int32(0); i < size; i++ {
				pixels[ptr] = 0x00000000
				ptr++
			}
			for i := int32(size); i < size*2; i++ {
				pixels[ptr] = 0x00000020
				ptr++
			}
		}
	}
	tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	tex.UpdateRGBA(nil, pixels, int(w))
	return tex, nil
}

func NewScanLinesH(r *sdl.Renderer, w int32, h int32, size int32) (*sdl.Texture, error) {
	tex, err := r.CreateTexture(sdl.PIXELFORMAT_RGBA8888,
		sdl.TEXTUREACCESS_TARGET, w, h)
	if err != nil {
		return nil, err
	}

	pixels := make([]uint32, w*h, w*h)
	for x := int32(0); x < w; x++ {
		for y := int32(0); y < h; y += 2 * size {
			ptr := (y * w) + x
			for i := int32(0); i < size; i++ {
				pixels[ptr] = 0x00000000
				ptr += w
			}
			for i := int32(size); i < size*2; i++ {
				pixels[ptr] = 0x00000020
				ptr += w
			}
		}
	}
	tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	tex.UpdateRGBA(nil, pixels, int(w))
	return tex, nil
}

func FitInWindow(winW int32, winH int32, screen *Screen) {
	deltaW, deltaH := winW-screen.W, winH-screen.H
	scale := int32(1)
	if deltaW < deltaH {
		scale = winW / screen.W
	} else {
		scale = winH / screen.H
	}
	scaledW, scaledH := screen.W*scale, screen.H*scale
	screen.X = (winW - scaledW) / 2
	screen.Y = (winH - scaledH) / 2
	screen.Scale = scale
}
