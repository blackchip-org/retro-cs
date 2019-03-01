package c128

import (
	"log"

	"github.com/blackchip-org/retro-cs/rcs"
)

const (
	VBlankFlag = (1 << 5)
	StatusFlag = (1 << 7)
)

const ()

type VDC struct {
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
		Status: VBlankFlag | StatusFlag,
	}
}

func (v *VDC) WriteAddr(val uint8) {
	if v.WatchAddr.Write {
		log.Printf("$%02x => vdc:addr", val)
	}
	v.Addr = val
}

func (v *VDC) ReadStatus() uint8 {
	if v.WatchAddr.Read {
		log.Printf("vdc:addr => $%02x", v.Status)
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
		log.Printf("vdc: read to unhandled addr $%02x", v.Addr)
	}
	if v.WatchData.Read {
		log.Printf("vdc[$%02x] => $%02x", v.Addr, val)
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
		log.Printf("vdc: write to unhandled addr $%02x, value $%02x", v.Addr, val)
	}
	if v.WatchData.Write {
		log.Printf("$%02x => vdc[$%02x]", val, v.Addr)
	}
}

func (v *VDC) blockOp(val uint8) {
	v.MemPos += uint16(val)
}
