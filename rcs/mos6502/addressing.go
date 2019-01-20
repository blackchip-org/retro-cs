package mos6502

func (cpu *CPU) loadAbsolute() uint8 {
	return cpu.mem.Read(cpu.fetch2())
}

func (cpu *CPU) loadAbsoluteX() uint8 {
	addr := cpu.fetch2()
	indexed := addr + int(cpu.X)
	if addr&0xff00 != indexed&0xff00 {
		cpu.pageCross = true
	}
	return cpu.mem.Read(indexed)
}

func (cpu *CPU) loadAbsoluteY() uint8 {
	addr := cpu.fetch2()
	indexed := addr + int(cpu.Y)
	if addr&0xff00 != indexed&0xff00 {
		cpu.pageCross = true
	}
	return cpu.mem.Read(indexed)
}

func (cpu *CPU) loadAccumulator() uint8 {
	return cpu.A
}

func (cpu *CPU) loadImmediate() uint8 {
	return cpu.fetch()
}

func (cpu *CPU) loadIndirectX() uint8 {
	zpaddr := int(cpu.fetch() + cpu.X)
	return cpu.mem.Read(cpu.mem.ReadLE(zpaddr))
}

func (cpu *CPU) loadIndirectY() uint8 {
	addr := cpu.mem.ReadLE(int(cpu.fetch()))
	indexed := addr + int(cpu.Y)
	if addr&0xff00 != indexed&0xff00 {
		cpu.pageCross = true
	}
	return cpu.mem.Read(indexed)
}

func (cpu *CPU) loadZeroPage() uint8 {
	return cpu.mem.Read(int(cpu.fetch()))
}

func (cpu *CPU) loadZeroPageX() uint8 {
	return cpu.mem.Read(int(cpu.fetch() + cpu.X))
}

func (cpu *CPU) loadZeroPageY() uint8 {
	return cpu.mem.Read(int(cpu.fetch() + cpu.Y))
}

func (cpu *CPU) storeAbsolute(value uint8) {
	cpu.mem.Write(cpu.fetch2(), value)
}

func (cpu *CPU) storeAbsoluteX(value uint8) {
	cpu.mem.Write(cpu.fetch2()+int(cpu.X), value)
}

func (cpu *CPU) storeAbsoluteY(v uint8) {
	cpu.mem.Write(cpu.fetch2()+int(cpu.Y), v)
}

func (cpu *CPU) storeIndirectX(v uint8) {
	zpaddr := int(cpu.fetch() + cpu.X)
	cpu.mem.Write(cpu.mem.ReadLE(zpaddr), v)
}

func (cpu *CPU) storeIndirectY(v uint8) {
	addr := cpu.mem.ReadLE(int(cpu.fetch())) + int(cpu.Y)
	cpu.mem.Write(addr, v)
}

func (cpu *CPU) storeZeroPage(v uint8) {
	cpu.mem.Write(int(cpu.fetch()), v)
}

func (cpu *CPU) storeZeroPageX(v uint8) {
	cpu.mem.Write(int(cpu.fetch()+cpu.X), v)
}

func (cpu *CPU) storeZeroPageY(v uint8) {
	cpu.mem.Write(int(cpu.fetch()+cpu.Y), v)
}
