package app

import (
	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/blackchip-org/retro-cs/system/c64"
	"github.com/blackchip-org/retro-cs/system/pacman"
)

var Systems = map[string]func(rcs.SDLContext) (*rcs.Mach, error){
	"c64":    c64.New,
	"pacman": pacman.New,
}
