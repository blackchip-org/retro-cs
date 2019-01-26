package namco

import (
	"fmt"
	"image/color"

	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/veandco/go-sdl2/sdl"
)

type Config struct {
	TileLayout     SheetLayout
	SpriteLayout   SheetLayout
	PaletteEntries int
	PaletteColors  int
	Hack           bool
	/*
		VideoAddr      uint16
	*/
}

var ViewerPalette = []color.RGBA{
	color.RGBA{0, 0, 0, 0},
	color.RGBA{128, 128, 128, 255},
	color.RGBA{192, 192, 192, 255},
	color.RGBA{255, 255, 255, 255},
}

type SheetLayout struct {
	TextureW     int32
	TextureH     int32
	TileW        int32
	TileH        int32
	BytesPerCell int32
	PixelLayout  [][]int
	PixelReader  func([]byte, int, int) uint8
}

func NewTileSheet(r *sdl.Renderer, d []byte, l SheetLayout, pal []color.RGBA) (rcs.TileSheet, error) {
	t, err := r.CreateTexture(sdl.PIXELFORMAT_RGBA8888,
		sdl.TEXTUREACCESS_TARGET, l.TextureW, l.TextureH)
	if err != nil {
		return rcs.TileSheet{}, fmt.Errorf("unable to create sheet: %v", err)
	}
	r.SetRenderTarget(t)

	rowTiles := l.TextureW / l.TileW
	for i := int32(0); i < l.TextureW*l.TextureH; i++ {
		targetX := i % l.TextureW
		targetY := i / l.TextureW
		tileX := targetX / l.TileW
		offsetX := targetX % l.TileW
		tileY := targetY / l.TileH
		offsetY := targetY % l.TileH
		tileN := tileX + tileY*rowTiles
		baseAddr := int(tileN * l.BytesPerCell)
		pixelN := int(l.PixelLayout[offsetY][offsetX])
		value := l.PixelReader(d, baseAddr, pixelN)

		c := pal[value]
		r.SetDrawColor(c.R, c.G, c.G, c.A)
		r.DrawPoint(targetX, targetY)
	}
	t.SetBlendMode(sdl.BLENDMODE_BLEND)
	r.SetRenderTarget(nil)

	return rcs.TileSheet{
		TextureW: l.TextureW,
		TextureH: l.TextureH,
		TileW:    l.TileW,
		TileH:    l.TileH,
		Texture:  t,
	}, nil
}

func ColorTable(d []byte, config Config) []color.RGBA {
	// FIXME: Galaga testing
	if config.Hack {
		return []color.RGBA{}
	}
	colors := make([]color.RGBA, 16, 16)
	for addr := 0; addr < 16; addr++ {
		r, g, b := uint8(0), uint8(0), uint8(0)
		c := d[addr]
		for bit := uint8(0); bit < 8; bit++ {
			if c&(1<<bit) != 0 {
				r += colorWeights[bit][0]
				g += colorWeights[bit][1]
				b += colorWeights[bit][2]
			}
		}
		alpha := uint8(0xff)
		// Color 0 is actually transparent
		if addr == 0 {
			alpha = 0x00
		}
		colors[addr] = color.RGBA{r, g, b, alpha}
	}
	return colors
}

func PaletteTable(d []byte, config Config, colors []color.RGBA) [][]color.RGBA {
	palettes := make([][]color.RGBA, config.PaletteEntries, config.PaletteEntries)
	for pal := 0; pal < config.PaletteEntries; pal++ {
		// FIXME: Galaga testing
		if config.Hack {
			palettes[pal] = ViewerPalette
			continue
		}
		addr := pal * config.PaletteColors
		entry := make([]color.RGBA, config.PaletteColors, config.PaletteColors)
		for i := 0; i < config.PaletteColors; i++ {
			ref := d[addr+i] & 0x0f
			entry[i] = colors[ref]
		}
		palettes[pal] = entry
	}
	return palettes
}

var colorWeights = [][]uint8{
	[]uint8{0x21, 0x00, 0x00},
	[]uint8{0x47, 0x00, 0x00},
	[]uint8{0x97, 0x00, 0x00},
	[]uint8{0x00, 0x21, 0x00},
	[]uint8{0x00, 0x47, 0x00},
	[]uint8{0x00, 0x97, 0x00},
	[]uint8{0x00, 0x00, 0x51},
	[]uint8{0x00, 0x00, 0xae},
}
