package mock

import "github.com/blackchip-org/retro-cs/rcs"

func NewMach() *rcs.Mach {
	ResetMemory()
	return &rcs.Mach{
		Mem: []*rcs.Memory{TestMemory},
		CPU: []rcs.CPU{NewCPU(TestMemory)},
		CharDecoders: map[string]rcs.CharDecoder{
			"ascii": rcs.ASCIIDecoder,
			"az26":  AZ26Decoder,
		},
		DefaultEncoding: "ascii",
	}
}

var AZ26Decoder = func(code uint8) (rune, bool) {
	if code < 1 || code > 26 {
		return 0, false
	}
	return rune(64 + code), true
}
