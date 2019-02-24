package c128

import "github.com/blackchip-org/retro-cs/rcs"

var SystemROM = []rcs.ROM{
	rcs.NewROM("basichi", "basichi", "c4fb4a714e48a7bf6c28659de0302183a0e0d6c0"),
	rcs.NewROM("basiclo", "basiclo", "d53a7884404f7d18ebd60dd3080c8f8d71067441"),
	rcs.NewROM("chargen", "chargen", "29ed066d513f2d5c09ff26d9166ba23c2afb2b3f"),
	rcs.NewROM("kernal ", "kernal ", "ceb6e1a1bf7e08eb9cbc651afa29e26adccf38ab"),
}
