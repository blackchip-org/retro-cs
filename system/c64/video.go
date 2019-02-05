package c64

import (
	"image/color"

	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	width      = 320
	height     = 200
	screenW    = 404 // actually 403?
	screenH    = 284
	borderW    = (screenW - width) / 2
	borderH    = (screenH - height) / 2
	charSheetW = 32
	charSheetH = 16
)

type Video struct {
	borderColor uint8
	bgColor     uint8
	texture     *sdl.Texture
	charSheet   rcs.TileSheet
	mem         *rcs.Memory
}

func NewVideo(r *sdl.Renderer, mem *rcs.Memory, charData []uint8) (*Video, error) {
	t, err := r.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET,
		screenW, screenH)
	if err != nil {
		return nil, err
	}
	charSheet, err := CharGen(r, charData)
	if err != nil {
		return nil, err
	}
	return &Video{
		texture:   t,
		charSheet: charSheet,
		mem:       mem,
	}, nil
}

func (v *Video) draw(r *sdl.Renderer) error {
	v.mem.Write(0xd012, 00) // HACK: set raster line to zero
	r.SetRenderTarget(v.texture)
	r.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
	v.drawBorder(r)
	v.drawBackground(r)
	v.drawCharacters(r)
	r.SetRenderTarget(nil)
	return nil
}

func (v *Video) drawBorder(r *sdl.Renderer) {
	c := Palette[v.borderColor&0x0f]
	r.SetDrawColor(c.R, c.G, c.B, c.A)
	topBorder := sdl.Rect{
		X: 0,
		Y: 0,
		W: screenW,
		H: borderH,
	}
	r.FillRect(&topBorder)
	bottomBorder := sdl.Rect{
		X: 0,
		Y: borderH + height,
		W: screenW,
		H: borderH,
	}
	r.FillRect(&bottomBorder)
	leftBorder := sdl.Rect{
		X: 0,
		Y: borderH,
		W: borderW,
		H: height,
	}
	r.FillRect(&leftBorder)
	rightBorder := sdl.Rect{
		X: borderW + width,
		Y: borderH,
		W: borderW,
		H: height,
	}
	r.FillRect(&rightBorder)
}

func (v *Video) drawBackground(r *sdl.Renderer) {
	c := Palette[v.bgColor&0x0f]
	r.SetDrawColor(c.R, c.G, c.B, c.A)
	background := sdl.Rect{
		X: borderW,
		Y: borderH,
		W: width,
		H: height,
	}
	r.FillRect(&background)
}

func (v *Video) drawCharacters(r *sdl.Renderer) {
	addrScreenMem := 0x0400
	addrColorMem := 0xd800
	baseX := 0
	baseY := 0
	for baseY < height {
		ch := v.mem.Read(addrScreenMem)
		clr := Palette[v.mem.Read(addrColorMem)&0x0f]
		v.charSheet.Texture.SetColorMod(clr.R, clr.G, clr.B)
		chx := int32(ch) % charSheetW * 8
		chy := int32(ch) / charSheetW * 8
		src := sdl.Rect{X: chx, Y: chy, W: 8, H: 8}
		dest := sdl.Rect{
			X: int32(baseX + borderW),
			Y: int32(baseY + borderH),
			W: 8,
			H: 8,
		}
		r.Copy(v.charSheet.Texture, &src, &dest)
		addrScreenMem++
		addrColorMem++
		baseX += 8
		if baseX >= width {
			baseX = 0
			baseY += 8
		}
	}
}

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
	r.SetDrawColor(0, 0, 0, 0)
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
