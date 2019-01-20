package mos6502

import (
	"github.com/blackchip-org/retro/rcs"
)

func adc(cpu *CPU, load rcs.Load8) {
	if cpu.SR&FlagD != 0 {
		cpu.alu.AddBCD(load())
	} else {
		cpu.alu.Add(load())
	}
}

func brk(cpu *CPU) {
	cpu.SR |= FlagB
	cpu.fetch()
}
