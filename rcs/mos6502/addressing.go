package mos6502

func (c *CPU) loadAbsolute() uint8 {
	c.addrLoad = c.fetch2()
	return c.mem.Read(c.addrLoad)
}

func (c *CPU) loadAbsoluteX() uint8 {
	arg := c.fetch2()
	c.addrLoad = arg + int(c.X)
	if c.addrLoad&0xff00 != arg&0xff00 {
		c.pageCross = true
	}
	return c.mem.Read(c.addrLoad)
}

func (c *CPU) loadAbsoluteY() uint8 {
	arg := c.fetch2()
	c.addrLoad = arg + int(c.Y)
	if c.addrLoad&0xff00 != arg&0xff00 {
		c.pageCross = true
	}
	return c.mem.Read(c.addrLoad)
}

func (c *CPU) loadA() uint8 {
	return c.A
}

func (c *CPU) loadX() uint8 {
	return c.X
}

func (c *CPU) loadY() uint8 {
	return c.Y
}

func (c *CPU) loadSP() uint8 {
	return c.SP
}

func (c *CPU) loadImmediate() uint8 {
	return c.fetch()
}

func (c *CPU) loadIndirectX() uint8 {
	zpaddr := int(c.fetch() + c.X)
	c.addrLoad = c.mem.ReadLE(zpaddr)
	return c.mem.Read(c.addrLoad)
}

func (c *CPU) loadIndirectY() uint8 {
	c.addrLoad = c.mem.ReadLE(int(c.fetch()))
	indexed := c.addrLoad + int(c.Y)
	if c.addrLoad&0xff00 != indexed&0xff00 {
		c.pageCross = true
	}
	return c.mem.Read(indexed)
}

func (c *CPU) loadZeroPage() uint8 {
	c.addrLoad = int(c.fetch())
	return c.mem.Read(c.addrLoad)
}

func (c *CPU) loadZeroPageX() uint8 {
	c.addrLoad = int(c.fetch() + c.X)
	return c.mem.Read(c.addrLoad)
}

func (c *CPU) loadZeroPageY() uint8 {
	c.addrLoad = int(c.fetch() + c.Y)
	return c.mem.Read(c.addrLoad)
}

func (c *CPU) storeAbsolute(value uint8) {
	c.mem.Write(c.fetch2(), value)
}

func (c *CPU) storeAbsoluteX(value uint8) {
	c.mem.Write(c.fetch2()+int(c.X), value)
}

func (c *CPU) storeAbsoluteY(v uint8) {
	c.mem.Write(c.fetch2()+int(c.Y), v)
}

func (c *CPU) storeIndirectX(v uint8) {
	zpaddr := int(c.fetch() + c.X)
	c.mem.Write(c.mem.ReadLE(zpaddr), v)
}

func (c *CPU) storeIndirectY(v uint8) {
	addr := c.mem.ReadLE(int(c.fetch())) + int(c.Y)
	c.mem.Write(addr, v)
}

func (c *CPU) storeZeroPage(v uint8) {
	c.mem.Write(int(c.fetch()), v)
}

func (c *CPU) storeZeroPageX(v uint8) {
	c.mem.Write(int(c.fetch()+c.X), v)
}

func (c *CPU) storeZeroPageY(v uint8) {
	c.mem.Write(int(c.fetch()+c.Y), v)
}

func (c *CPU) storeA(v uint8) {
	c.A = v
}

func (c *CPU) storeX(v uint8) {
	c.X = v
}

func (c *CPU) storeY(v uint8) {
	c.Y = v
}

func (c *CPU) storeSP(v uint8) {
	c.SP = v
}

func (c *CPU) storeBack(v uint8) {
	c.mem.Write(c.addrLoad, v)
}
