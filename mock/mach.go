package mock

import "github.com/blackchip-org/retro-cs/rcs"

func NewMach() *rcs.Mach {
	ResetMemory()
	return &rcs.Mach{
		Mem: []*rcs.Memory{TestMemory},
		CPU: []rcs.CPU{NewCPU(TestMemory)},
	}
}
