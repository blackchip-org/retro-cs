package z80

func (c *CPU) storeNil(v uint8) {}

func (c *CPU) storeIndImm(v uint8) { c.mem.Write(c.fetch2(), v) }
func (c *CPU) store16IndImm(v int) { c.mem.WriteLE(c.fetch2(), v) }

func (c *CPU) storeA(v uint8)   { c.A = v }
func (c *CPU) storeF(v uint8)   { c.F = v }
func (c *CPU) storeB(v uint8)   { c.B = v }
func (c *CPU) storeC(v uint8)   { c.C = v }
func (c *CPU) storeD(v uint8)   { c.D = v }
func (c *CPU) storeE(v uint8)   { c.E = v }
func (c *CPU) storeH(v uint8)   { c.H = v }
func (c *CPU) storeL(v uint8)   { c.L = v }
func (c *CPU) storeI(v uint8)   { c.I = v }
func (c *CPU) storeR(v uint8)   { c.R = v }
func (c *CPU) storeIXH(v uint8) { c.IXH = v }
func (c *CPU) storeIXL(v uint8) { c.IXL = v }
func (c *CPU) storeIYH(v uint8) { c.IYH = v }
func (c *CPU) storeIYL(v uint8) { c.IYL = v }

func (c *CPU) storeA1(v uint8) { c.A1 = v }
func (c *CPU) storeF1(v uint8) { c.F1 = v }
func (c *CPU) storeB1(v uint8) { c.B1 = v }
func (c *CPU) storeC1(v uint8) { c.C1 = v }
func (c *CPU) storeD1(v uint8) { c.D1 = v }
func (c *CPU) storeE1(v uint8) { c.E1 = v }
func (c *CPU) storeH1(v uint8) { c.H1 = v }
func (c *CPU) storeL1(v uint8) { c.L1 = v }

func (c *CPU) storeAF(v int) { c.A, c.F = uint8(v>>8), uint8(v) }
func (c *CPU) storeBC(v int) { c.B, c.C = uint8(v>>8), uint8(v) }
func (c *CPU) storeDE(v int) { c.D, c.E = uint8(v>>8), uint8(v) }
func (c *CPU) storeHL(v int) { c.H, c.L = uint8(v>>8), uint8(v) }
func (c *CPU) storeSP(v int) { c.SP = uint16(v) }
func (c *CPU) storeIX(v int) { c.IXH, c.IXL = uint8(v>>8), uint8(v) }
func (c *CPU) storeIY(v int) { c.IYH, c.IYL = uint8(v>>8), uint8(v) }

func (c *CPU) store16IndSP(v int) { c.mem.WriteLE(int(c.SP), v) }

func (c *CPU) storeAF1(v int) { c.A1, c.F1 = uint8(v>>8), uint8(v) }
func (c *CPU) storeBC1(v int) { c.B1, c.C1 = uint8(v>>8), uint8(v) }
func (c *CPU) storeDE1(v int) { c.D1, c.E1 = uint8(v>>8), uint8(v) }
func (c *CPU) storeHL1(v int) { c.H1, c.L1 = uint8(v>>8), uint8(v) }

func (c *CPU) storeIndHL(v uint8) { c.mem.Write(int(c.H)<<8|int(c.L), v) }

func (c *CPU) storeIndBC(v uint8) { c.mem.Write(int(c.B)<<8|int(c.C), v) }
func (c *CPU) storeIndDE(v uint8) { c.mem.Write(int(c.D)<<8|int(c.E), v) }

func (c *CPU) loadZero() uint8   { return 0 }
func (c *CPU) loadImm() uint8    { return c.fetch() }
func (c *CPU) loadImm16() int    { return c.fetch2() }
func (c *CPU) loadIndImm() uint8 { return c.mem.Read(c.fetch2()) }
func (c *CPU) load16IndImm() int { return c.mem.ReadLE(c.fetch2()) }

func (c *CPU) loadA() uint8    { return c.A }
func (c *CPU) loadF() uint8    { return c.F }
func (c *CPU) loadB() uint8    { return c.B }
func (c *CPU) loadC() uint8    { return c.C }
func (c *CPU) loadD() uint8    { return c.D }
func (c *CPU) loadE() uint8    { return c.E }
func (c *CPU) loadH() uint8    { return c.H }
func (c *CPU) loadL() uint8    { return c.L }
func (c *CPU) loadI() uint8    { return c.I }
func (c *CPU) loadR() uint8    { return c.R }
func (c *CPU) loadIXL() uint8  { return c.IXL }
func (c *CPU) loadIXH() uint8  { return c.IXH }
func (c *CPU) loadIYL() uint8  { return c.IYL }
func (c *CPU) loadIYH() uint8  { return c.IYH }
func (c *CPU) loadIndC() uint8 { panic("not done") /*return c.Ports.Load(uint16(c.C)) */ }

func (c *CPU) loadA1() uint8 { return c.A1 }
func (c *CPU) loadF1() uint8 { return c.F1 }
func (c *CPU) loadB1() uint8 { return c.B1 }
func (c *CPU) loadC1() uint8 { return c.C1 }
func (c *CPU) loadD1() uint8 { return c.D1 }
func (c *CPU) loadE1() uint8 { return c.E1 }
func (c *CPU) loadH1() uint8 { return c.H1 }
func (c *CPU) loadL1() uint8 { return c.L1 }

func (c *CPU) loadAF() int      { return int(c.A)<<8 | int(c.F) }
func (c *CPU) loadBC() int      { return int(c.B)<<8 | int(c.C) }
func (c *CPU) loadDE() int      { return int(c.D)<<8 | int(c.E) }
func (c *CPU) loadHL() int      { return int(c.H)<<8 | int(c.L) }
func (c *CPU) loadSP() int      { return int(c.SP) }
func (c *CPU) loadIX() int      { return int(c.IXH)<<8 | int(c.IXL) }
func (c *CPU) loadIY() int      { return int(c.IYH)<<8 | int(c.IYL) }
func (c *CPU) load16IndSP() int { return c.mem.ReadLE(int(c.SP)) }

func (c *CPU) loadAF1() int { return int(c.A1)<<8 | int(c.F1) }
func (c *CPU) loadBC1() int { return int(c.B1)<<8 | int(c.C1) }
func (c *CPU) loadDE1() int { return int(c.D1)<<8 | int(c.E1) }
func (c *CPU) loadHL1() int { return int(c.H1)<<8 | int(c.L1) }

func (c *CPU) loadIndHL() uint8 { return c.mem.Read(int(c.H)<<8 | int(c.L)) }

func (c *CPU) loadIndBC() uint8 { return c.mem.Read(int(c.B)<<8 | int(c.C)) }
func (c *CPU) loadIndDE() uint8 { return c.mem.Read(int(c.D)<<8 | int(c.E)) }

func (c *CPU) loadIndIX() uint8 {
	ix := int(c.IXH)<<8 | int(c.IXL)
	c.iaddr = ix + int(int8(c.delta))
	return c.mem.Read(c.iaddr)
}

func (c *CPU) loadIndIY() uint8 {
	iy := int(c.IYH)<<8 | int(c.IYL)
	c.iaddr = iy + int(int8(c.delta))
	return c.mem.Read(c.iaddr)
}

func (c *CPU) storeIndIX(v uint8) {
	ix := int(c.IXH)<<8 | int(c.IXL)
	addr := ix + int(int8(c.delta))
	c.mem.Write(addr, v)
}

func (c *CPU) storeIndIY(v uint8) {
	iy := int(c.IYH)<<8 | int(c.IYL)
	addr := iy + int(int8(c.delta))
	c.mem.Write(addr, v)
}

func (c *CPU) storeLastInd(v uint8) {
	c.mem.Write(c.iaddr, v)
}

func (c *CPU) outIndImm(v uint8) {
	/*
		addr := uint16(c.fetch())
		c.Ports.Store(addr, v)
	*/
	panic("not done")
}

func (c *CPU) inIndImm() uint8 {
	/*
		addr := uint16(c.fetch())
		return c.Ports.Load(addr)
	*/
	panic("not done")
}

func (c *CPU) outIndC(v uint8) {
	// c.Ports.Store(uint16(c.C), v)
	panic("not done")
}

func (c *CPU) inIndC() uint8 {
	// return c.Ports.Load(uint16(c.C))
	panic("not done")
}
