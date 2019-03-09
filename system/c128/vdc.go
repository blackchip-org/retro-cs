package c128

import (
	"log"

	"github.com/blackchip-org/retro-cs/rcs"
)

const (
	VBlankFlag = (1 << 5)
	StatusFlag = (1 << 7)
)

type VDC struct {
	Name   string
	Addr   uint8
	Status uint8
	MemPos uint16
	VSS    uint8 // vertical smooth scrolling and control register

	WatchAddr   rcs.FlagRW
	WatchStatus rcs.FlagRW
	WatchData   rcs.FlagRW
}

func NewVDC() *VDC {
	return &VDC{
		Name:   "vdc",
		Status: VBlankFlag | StatusFlag,
	}
}

func (v *VDC) WriteAddr(val uint8) {
	if v.WatchAddr.W {
		log.Printf("%v:addr <= %v", v.Name, rcs.X8(val))
	}
	v.Addr = val
}

func (v *VDC) ReadStatus() uint8 {
	if v.WatchAddr.R {
		log.Printf("%v <= %v:addr", rcs.X8(v.Status), v.Name)
	}
	return v.Status
}

func (v *VDC) ReadData() uint8 {
	val := uint8(0)
	switch v.Addr {
	case 0x12: // current memory address (high byte)
		val = uint8(v.MemPos >> 8)
	case 0x13: // current memory address (low byte)
		val = uint8(v.MemPos)
	case 0x18: // vertical smooth scrolling and control register
		val = v.VSS
	case 0x1f: // memory read/write register
		v.MemPos++
	default:
		log.Printf("(!) %v: read to unhandled addr %v", v.Name, rcs.X8(v.Addr))
	}
	if v.WatchData.R {
		log.Printf("%v <= %v[%v]", rcs.X8(val), v.Name, rcs.X8(v.Addr))
	}
	return val
}

func (v *VDC) WriteData(val uint8) {
	switch v.Addr {
	case 0x12: // current memory address (high byte)
		v.MemPos = uint16(val)<<8 | v.MemPos&0xff
	case 0x13: // current memory address (low byte)
		v.MemPos = v.MemPos&0xff00 | uint16(val)
	case 0x18: // vertical smooth scrolling and control register
		v.VSS = val
	case 0x1e: // number of bytes to copy or fill
		v.blockOp(val)
	case 0x1f: // memory read/write register
		v.MemPos++
	default:
		log.Printf("(!) %v: write to unhandled addr %v, value %v", v.Name, rcs.X8(v.Addr), rcs.X8(val))
	}
	if v.WatchData.W {
		log.Printf("%v[%v] <= %v", v.Name, rcs.X8(v.Addr), rcs.X8(val))
	}
}

func (v *VDC) blockOp(val uint8) {
	v.MemPos += uint16(val)
}
