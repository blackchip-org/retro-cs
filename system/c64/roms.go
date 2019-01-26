package c64

import "github.com/blackchip-org/retro-cs/rcs"

var SystemROM = []rcs.ROM{
	rcs.NewROM("basic  ", "basic  ", "79015323128650c742a3694c9429aa91f355905e"),
	rcs.NewROM("chargen", "chargen", "adc7c31e18c7c7413d54802ef2f4193da14711aa"),
	rcs.NewROM("kernal ", "kernal ", "1d503e56df85a62fee696e7618dc5b4e781df1bb"),
}
