package main

import (
	"image/color"

	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/blackchip-org/retro-cs/system/c64"
	"github.com/veandco/go-sdl2/sdl"
)

var views = map[string]view{
	"c64:colors": view{
		system: "c64",
		render: func(r *sdl.Renderer, m map[string]*rcs.Memory) (rcs.TileSheet, error) {
			palettes := [][]color.RGBA{c64.Palette}
			return rcs.NewColorSheet(r, palettes)
		},
	},
}
