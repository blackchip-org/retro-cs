package main

import (
	"image/color"

	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/blackchip-org/retro-cs/system/c64"
	"github.com/veandco/go-sdl2/sdl"
)

var views = map[string]view{
	"c64:chars": view{
		system: "c64",
		roms:   c64.ROMS,
		render: func(r *sdl.Renderer, d map[string][]byte) (rcs.TileSheet, error) {
			return c64.CharGen(r, d["chargen"])
		},
	},
	"c64:colors": view{
		system: "c64",
		render: func(r *sdl.Renderer, m map[string][]byte) (rcs.TileSheet, error) {
			palettes := [][]color.RGBA{c64.Palette}
			return rcs.NewColorSheet(r, palettes)
		},
	},
}
