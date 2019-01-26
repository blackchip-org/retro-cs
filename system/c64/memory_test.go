package c64

import (
	"fmt"
	"testing"
)

func TestMemoryLoad(t *testing.T) {
	const (
		page00 = 0x0000
		page10 = 0x1000
		page80 = 0x8000
		pagea0 = 0xa000
		pagec0 = 0xc000
		paged0 = 0xd000
		pagee0 = 0xe000
	)

	const (
		vram    = 0xf0
		vbasic  = 0xba
		vkernal = 0xea
		vchar   = 0xca
		vcartlo = 0x0c
		vcarthi = 0xc0
		vio     = 0x10
	)

	cart := make([]byte, 16*1024, 16*1024)
	cart[0x0000] = vcartlo
	cart[0x2000] = vcarthi

	roms := map[string][]byte{
		"basic":   []byte{vbasic},
		"kernal":  []byte{vkernal},
		"chargen": []byte{vchar},
		"cart":    cart,
	}

	mem := newMemory(roms)
	mem.SetBank(31)
	mem.Write(paged0, vio)

	mem.SetBank(0)
	mem.Write(page00, vram)
	mem.Write(page10, vram)
	mem.Write(page80, vram)
	mem.Write(pagea0, vram)
	mem.Write(pagec0, vram)
	mem.Write(paged0, vram)
	mem.Write(pagee0, vram)

	var tests = []struct {
		mode int
		addr int
		want uint8
	}{
		{0, page00, vram},
		{0, page10, vram},
		{0, page80, vram},
		{0, pagea0, vram},
		{0, pagec0, vram},
		{0, paged0, vram},
		{0, pagee0, vram},

		{1, page00, vram},
		{1, page10, vram},
		{1, page80, vram},
		{1, pagea0, vram},
		{1, pagec0, vram},
		{1, paged0, vram},
		{1, pagee0, vram},

		{2, page00, vram},
		{2, page10, vram},
		{2, page80, vram},
		{2, pagea0, vcarthi},
		{2, pagec0, vram},
		{2, paged0, vchar},
		{2, pagee0, vkernal},

		{3, page00, vram},
		{3, page10, vram},
		{3, page80, vcartlo},
		{3, pagea0, vcarthi},
		{3, pagec0, vram},
		{3, paged0, vchar},
		{3, pagee0, vkernal},

		{4, page00, vram},
		{4, page10, vram},
		{4, page80, vram},
		{4, pagea0, vram},
		{4, pagec0, vram},
		{4, paged0, vram},
		{4, pagee0, vram},

		{5, page00, vram},
		{5, page10, vram},
		{5, page80, vram},
		{5, pagea0, vram},
		{5, pagec0, vram},
		{5, paged0, vio},
		{5, pagee0, vram},

		{6, page00, vram},
		{6, page10, vram},
		{6, page80, vram},
		{6, pagea0, vcarthi},
		{6, pagec0, vram},
		{6, paged0, vio},
		{6, pagee0, vkernal},

		{7, page00, vram},
		{7, page10, vram},
		{7, page80, vcartlo},
		{7, pagea0, vcarthi},
		{7, pagec0, vram},
		{7, paged0, vio},
		{7, pagee0, vkernal},

		{8, page00, vram},
		{8, page10, vram},
		{8, page80, vram},
		{8, pagea0, vram},
		{8, pagec0, vram},
		{8, paged0, vram},
		{8, pagee0, vram},

		{9, page00, vram},
		{9, page10, vram},
		{9, page80, vram},
		{9, pagea0, vram},
		{9, pagec0, vram},
		{9, paged0, vchar},
		{9, pagee0, vram},

		{10, page00, vram},
		{10, page10, vram},
		{10, page80, vram},
		{10, pagea0, vram},
		{10, pagec0, vram},
		{10, paged0, vchar},
		{10, pagee0, vkernal},

		{11, page00, vram},
		{11, page10, vram},
		{11, page80, vcartlo},
		{11, pagea0, vbasic},
		{11, pagec0, vram},
		{11, paged0, vchar},
		{11, pagee0, vkernal},

		{12, page00, vram},
		{12, page10, vram},
		{12, page80, vram},
		{12, pagea0, vram},
		{12, pagec0, vram},
		{12, paged0, vram},
		{12, pagee0, vram},

		{13, page00, vram},
		{13, page10, vram},
		{13, page80, vram},
		{13, pagea0, vram},
		{13, pagec0, vram},
		{13, paged0, vio},
		{13, pagee0, vram},

		{14, page00, vram},
		{14, page10, vram},
		{14, page80, vram},
		{14, pagea0, vram},
		{14, pagec0, vram},
		{14, paged0, vio},
		{14, pagee0, vkernal},

		{15, page00, vram},
		{15, page10, vram},
		{15, page80, vcartlo},
		{15, pagea0, vbasic},
		{15, pagec0, vram},
		{15, paged0, vio},
		{15, pagee0, vkernal},

		{16, page00, vram},
		{16, page10, 0},
		{16, page80, vcartlo},
		{16, pagea0, 0},
		{16, pagec0, 0},
		{16, paged0, vio},
		{16, pagee0, vcarthi},

		{17, page00, vram},
		{17, page10, 0},
		{17, page80, vcartlo},
		{17, pagea0, 0},
		{17, pagec0, 0},
		{17, paged0, vio},
		{17, pagee0, vcarthi},

		{18, page00, vram},
		{18, page10, 0},
		{18, page80, vcartlo},
		{18, pagea0, 0},
		{18, pagec0, 0},
		{18, paged0, vio},
		{18, pagee0, vcarthi},

		{19, page00, vram},
		{19, page10, 0},
		{19, page80, vcartlo},
		{19, pagea0, 0},
		{19, pagec0, 0},
		{19, paged0, vio},
		{19, pagee0, vcarthi},

		{20, page00, vram},
		{20, page10, 0},
		{20, page80, vcartlo},
		{20, pagea0, 0},
		{20, pagec0, 0},
		{20, paged0, vio},
		{20, pagee0, vcarthi},

		{21, page00, vram},
		{21, page10, 0},
		{21, page80, vcartlo},
		{21, pagea0, 0},
		{21, pagec0, 0},
		{21, paged0, vio},
		{21, pagee0, vcarthi},

		{22, page00, vram},
		{22, page10, 0},
		{22, page80, vcartlo},
		{22, pagea0, 0},
		{22, pagec0, 0},
		{22, paged0, vio},
		{22, pagee0, vcarthi},

		{23, page00, vram},
		{23, page10, 0},
		{23, page80, vcartlo},
		{23, pagea0, 0},
		{23, pagec0, 0},
		{23, paged0, vio},
		{23, pagee0, vcarthi},

		{24, page00, vram},
		{24, page10, vram},
		{24, page80, vram},
		{24, pagea0, vram},
		{24, pagec0, vram},
		{24, paged0, vram},
		{24, pagee0, vram},

		{25, page00, vram},
		{25, page10, vram},
		{25, page80, vram},
		{25, pagea0, vram},
		{25, pagec0, vram},
		{25, paged0, vchar},
		{25, pagee0, vram},

		{26, page00, vram},
		{26, page10, vram},
		{26, page80, vram},
		{26, pagea0, vram},
		{26, pagec0, vram},
		{26, paged0, vchar},
		{26, pagee0, vkernal},

		{27, page00, vram},
		{27, page10, vram},
		{27, page80, vram},
		{27, pagea0, vbasic},
		{27, pagec0, vram},
		{27, paged0, vchar},
		{27, pagee0, vkernal},

		{28, page00, vram},
		{28, page10, vram},
		{28, page80, vram},
		{28, pagea0, vram},
		{28, pagec0, vram},
		{28, paged0, vram},
		{28, pagee0, vram},

		{29, page00, vram},
		{29, page10, vram},
		{29, page80, vram},
		{29, pagea0, vram},
		{29, pagec0, vram},
		{29, paged0, vio},
		{29, pagee0, vram},

		{30, page00, vram},
		{30, page10, vram},
		{30, page80, vram},
		{30, pagea0, vram},
		{30, pagec0, vram},
		{30, paged0, vio},
		{30, pagee0, vkernal},

		{31, page00, vram},
		{31, page10, vram},
		{31, page80, vram},
		{31, pagea0, vbasic},
		{31, pagec0, vram},
		{31, paged0, vio},
		{31, pagee0, vkernal},
	}

	for _, test := range tests {
		label := fmt.Sprintf("mode %02d addr %04x", test.mode, test.addr)
		t.Run(label, func(t *testing.T) {
			mem.SetBank(test.mode)
			have := mem.Read(test.addr)
			if test.want != have {
				t.Errorf("\n want: %02x \n have: %02x", test.want, have)
			}
		})
	}
}
