package namco

import (
	"fmt"
	"image/color"

	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/veandco/go-sdl2/sdl"
)

type Data struct {
	Palettes []uint8
	Colors   []uint8
	Tiles    []uint8
	Sprites  []uint8
}

type Config struct {
	TileLayout     SheetLayout
	SpriteLayout   SheetLayout
	PaletteEntries int
	PaletteColors  int
	Hack           bool
}

const (
	W = int32(224)
	H = int32(288)
)

type SpriteCoord struct {
	X uint8
	Y uint8
}

type Video struct {
	SpriteCoords   []SpriteCoord
	SpriteInfo     []uint8
	SpritePalettes []uint8
	TileMemory     []uint8
	ColorMemory    []uint8

	Texture  *sdl.Texture
	config   Config
	tiles    [64]rcs.TileSheet
	sprites  [64]rcs.TileSheet
	colors   []color.RGBA
	palettes [][]color.RGBA
}

func NewVideo(r *sdl.Renderer, config Config, data Data) (*Video, error) {
	tex, err := r.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, W, H)
	if err != nil {
		return nil, err
	}
	colors := ColorTable(config, data.Colors)
	palettes := PaletteTable(config, data.Palettes, colors)

	var tiles, sprites [64]rcs.TileSheet
	for pal := 0; pal < config.PaletteEntries; pal++ {
		t, err := NewTileSheet(r, data.Tiles, config.TileLayout, palettes[pal])
		if err != nil {
			return nil, err
		}
		tiles[pal] = t

		s, err := NewTileSheet(r, data.Sprites, config.SpriteLayout, palettes[pal])
		if err != nil {
			return nil, err
		}
		sprites[pal] = s
	}

	v := &Video{
		Texture:        tex,
		SpriteCoords:   make([]SpriteCoord, 8, 8),
		SpriteInfo:     make([]uint8, 8, 8),
		SpritePalettes: make([]uint8, 8, 8),
		TileMemory:     make([]uint8, 1024, 1024),
		ColorMemory:    make([]uint8, 1024, 1024),
		config:         config,
		tiles:          tiles,
		sprites:        sprites,
		colors:         colors,
		palettes:       palettes,
	}
	return v, nil
}

func (v *Video) Draw(r *sdl.Renderer) error {
	r.SetRenderTarget(v.Texture)
	r.SetDrawColorArray(0, 0, 0, 0xff)
	r.Clear()
	v.drawTiles(r)
	v.drawSprites(r)
	r.SetRenderTarget(nil)
	return nil
}

func (v *Video) drawTiles(r *sdl.Renderer) error {
	layout := v.config.TileLayout
	tileW := layout.TileW
	rowTiles := layout.TextureW / tileW

	// Render tiles
	for ty := 0; ty < 36; ty++ {
		for tx := 0; tx < 28; tx++ {
			var addr int
			if ty == 0 || ty == 1 {
				addr = 0x3dd + (ty * 0x20) - tx
			} else if ty == 34 || ty == 35 {
				addr = 0x01d + ((ty - 34) * 0x20) - tx
			} else {
				addr = 0x3a0 + (ty - 2) - (tx * 0x20)
			}

			tileN := int32(v.TileMemory[addr])
			sheetX := (tileN % rowTiles) * tileW
			sheetY := (tileN / rowTiles) * tileW
			src := sdl.Rect{
				X: int32(sheetX),
				Y: int32(sheetY),
				W: int32(layout.TileW),
				H: int32(layout.TileH),
			}
			screenX := int32(tx) * 8
			screenY := int32(ty) * 8
			dest := sdl.Rect{
				X: screenX,
				Y: screenY,
				W: layout.TileW,
				H: layout.TileH,
			}

			// Only 64 palettes, strip out the higher bits
			pal := v.ColorMemory[addr] & 0x3f
			r.Copy(v.tiles[pal].Texture, &src, &dest)
		}
	}
	return nil
}

func (v *Video) drawSprites(r *sdl.Renderer) error {
	// FIXME: Galaga testing
	if v.config.Hack {
		return nil
	}
	layout := v.config.SpriteLayout
	spriteW := layout.TileW
	spriteH := layout.TileH
	rowTiles := layout.TextureW / spriteW

	for s := 7; s >= 0; s-- {
		coordX := int32(v.SpriteCoords[s].X)
		coordY := int32(v.SpriteCoords[s].Y)
		info := v.SpriteInfo[s]
		spriteN := int32(info >> 2)
		flip := sdl.FLIP_NONE
		if info&0x02 > 0 {
			flip |= sdl.FLIP_HORIZONTAL
		}
		if info&0x01 > 0 {
			flip |= sdl.FLIP_VERTICAL
		}

		// do not render of off screen
		if coordX <= 30 || coordX >= 240 {
			continue
		}
		screenX := (W - coordX + spriteW)
		screenY := (H - coordY - spriteH)
		sheetX := (spriteN % rowTiles) * spriteW
		sheetY := (spriteN / rowTiles) * spriteH
		src := sdl.Rect{
			X: int32(sheetX),
			Y: int32(sheetY),
			W: spriteW,
			H: spriteH,
		}
		dest := sdl.Rect{
			X: screenX,
			Y: screenY,
			W: spriteW,
			H: spriteH,
		}
		// Only 64 palettes, strip out the higher bits
		pal := v.SpritePalettes[s] & 0x3f
		r.CopyEx(v.sprites[pal].Texture, &src, &dest, 0, nil, flip)
	}
	return nil
}

func (v *Video) Save(enc *rcs.Encoder) {
	enc.Encode(v.TileMemory)
	enc.Encode(v.ColorMemory)
}

func (v *Video) Load(enc *rcs.Decoder) {
	enc.Decode(&v.TileMemory)
	enc.Decode(&v.ColorMemory)
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
		r.SetDrawColor(c.R, c.G, c.B, c.A)
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

func ColorTable(config Config, data []uint8) []color.RGBA {
	// FIXME: Galaga testing
	if config.Hack {
		return []color.RGBA{}
	}
	colors := make([]color.RGBA, 16, 16)
	for addr := 0; addr < 16; addr++ {
		r, g, b := uint8(0), uint8(0), uint8(0)
		c := data[addr]
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

func PaletteTable(config Config, data []uint8, colors []color.RGBA) [][]color.RGBA {
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
			ref := data[addr+i] & 0x0f
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
