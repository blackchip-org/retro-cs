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
	Addr   uint8
	Status uint8
	Data   uint8

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
	if v.WatchData.Read {
		log.Printf("vdc:data => $%02x", v.Data)
	}
	return v.Data
}

func (v *VDC) WriteData(val uint8) {
	if v.WatchData.Write {
		log.Printf("$%02x => vdc:data", val)
	}
	v.Data = val
}
