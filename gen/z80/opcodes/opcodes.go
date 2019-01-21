package main

//go:generate go run .
//go:generate go fmt ../../../rcs/z80/opcodes.go

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/blackchip-org/retro-cs/rcs"
)

var (
	root      = filepath.Join("..", "..", "..")
	targetDir = filepath.Join(root, "rcs", "z80")
)

type regtab struct {
	name string
	r    map[int]string
	rp   map[int]string
	rp2  map[int]string
}

// unprefixed registers
var un = &regtab{
	name: "un",
	r: map[int]string{
		0: "B",
		1: "C",
		2: "D",
		3: "E",
		4: "H",
		5: "L",
		6: "IndHL",
		7: "A",
	},
	rp: map[int]string{
		0: "BC",
		1: "DE",
		2: "HL",
		3: "SP",
	},
	rp2: map[int]string{
		0: "BC",
		1: "DE",
		2: "HL",
		3: "AF",
	},
}

// dd-prefixed registers
var dd = &regtab{
	name: "dd",
	r: map[int]string{
		0: "B",
		1: "C",
		2: "D",
		3: "E",
		4: "IXH",
		5: "IXL",
		6: "IndIX",
		7: "A",
	},
	rp: map[int]string{
		0: "BC",
		1: "DE",
		2: "IX",
		3: "SP",
	},
	rp2: map[int]string{
		0: "BC",
		1: "DE",
		2: "IX",
		3: "AF",
	},
}

// fd-prefixed registers
var fd = &regtab{
	name: "fd",
	r: map[int]string{
		0: "B",
		1: "C",
		2: "D",
		3: "E",
		4: "IYH",
		5: "IYL",
		6: "IndIY",
		7: "A",
	},
	rp: map[int]string{
		0: "BC",
		1: "DE",
		2: "IY",
		3: "SP",
	},
	rp2: map[int]string{
		0: "BC",
		1: "DE",
		2: "IY",
		3: "AF",
	},
}

// ddcb-prefixed registers
var ddcb = &regtab{
	name: "ddcb",
	r: map[int]string{
		0: "B",
		1: "C",
		2: "D",
		3: "E",
		4: "IXH",
		5: "IXL",
		6: "IndIX",
		7: "A",
	},
	rp:  nil,
	rp2: nil,
}

// fdcb-prefixed registers
var fdcb = &regtab{
	name: "fdcb",
	r: map[int]string{
		0: "B",
		1: "C",
		2: "D",
		3: "E",
		4: "IYH",
		5: "IYL",
		6: "IndIY",
		7: "A",
	},
	rp:  nil,
	rp2: nil,
}

var cc = map[int]string{
	0: "FlagZ, false",
	1: "FlagZ, true",
	2: "FlagC, false",
	3: "FlagC, true",
	4: "FlagV, false",
	5: "FlagV, true",
	6: "FlagS, false",
	7: "FlagS, true",
}

func processMain(tab *regtab, op uint8) string {
	//r := tab.r
	//rp := tab.rp
	//rp2 := tab.rp2

	x := int(rcs.SliceBits(op, 6, 7))
	y := int(rcs.SliceBits(op, 3, 5))
	z := int(rcs.SliceBits(op, 0, 2))
	//p := int(rcs.SliceBits(op, 4, 5))
	//q := int(rcs.SliceBits(op, 3, 3))

	// If the instruction has a ddcb or fdcb prefix, the instruction handler
	// will take care of it so this should never be called. Panic just
	// in case
	if tab.name == "dd" && op == 0xcb {
		return "panic(\"instruction prefix ddcb should be handled elsewhere\")"
	}
	if tab.name == "fd" && op == 0xcb {
		return "panic(\"instruction prefix fdcb should be handled elsewhere\")"
	}

	if x == 0 {
		if z == 0 {
			if y == 0 {
				return "nop()"
			}
			if y == 1 {
				return ""
				//return "ex(c, c.loadAF, c.storeAF, c.loadAF1, c.storeAF1)"
			}
			if y == 2 {
				return ""
				//return "djnz(c, c.loadImm)"
			}
			if y == 3 {
				return ""
				//return "jra(c, c.loadImm)"
			}
			if y >= 4 && y <= 7 {
				return ""
				//return fmt.Sprintf("jr(c, %v, c.loadImm)", cc[y-4])
			}
		}
		/*
			if z == 1 {
				if q == 0 {
					return fmt.Sprintf("ld16(c, c.store%v, c.loadImm16)", rp[p])
				}
				if q == 1 {
					return fmt.Sprintf("add16(c, c.store%v, c.load%v, c.load%v, false)", rp2[2], rp[2], rp[p])
				}
			}
			if z == 2 {
				if q == 0 {
					if p == 0 {
						return "ld(c, c.storeIndBC, c.loadA)"
					}
					if p == 1 {
						return "ld(c, c.storeIndDE, c.loadA)"
					}
					if p == 2 {
						return fmt.Sprintf("ld16(c, c.store16IndImm, c.load%v)", rp2[2])
					}
					if p == 3 {
						return "ld(c, c.storeIndImm, c.loadA)"
					}
				}
				if q == 1 {
					if p == 0 {
						return "ld(c, c.storeA, c.loadIndBC)"
					}
					if p == 1 {
						return "ld(c, c.storeA, c.loadIndDE)"
					}
					if p == 2 {
						return fmt.Sprintf("ld16(c, c.store%v, c.load16IndImm)", rp2[2])
					}
					if p == 3 {
						return "ld(c, c.storeA, c.loadIndImm)"
					}
				}
			}
			if z == 3 {
				if q == 0 {
					return fmt.Sprintf("inc16(c, c.store%v, c.load%v)", rp[p], rp[p])
				}
				if q == 1 {
					return fmt.Sprintf("dec16(c, c.store%v, c.load%v)", rp[p], rp[p])
				}
			}
			if z == 4 {
				return fmt.Sprintf("inc(c, c.store%v, c.load%v)", r[y], r[y])
			}
			if z == 5 {
				return fmt.Sprintf("dec(c, c.store%v, c.load%v)", r[y], r[y])
			}
			if z == 6 {
				return fmt.Sprintf("ld(c, c.store%v, c.loadImm)", r[y])
			}
			if z == 7 {
				if y == 0 {
					return "rlca(c)"
				}
				if y == 1 {
					return "rrca(c)"
				}
				if y == 2 {
					return "rla(c)"
				}
				if y == 3 {
					return "rra(c)"
				}
				if y == 4 {
					return "daa(c)"
				}
				if y == 5 {
					return "cpl(c)"
				}
				if y == 6 {
					return "scf(c)"
				}
				if y == 7 {
					return "ccf(c)"
				}
			}
		*/
	}
	/*
		if x == 1 {
			if z == 6 && y == 6 {
				return "halt(c)"
			}
			return fmt.Sprintf("ld(c, c.store%v, c.load%v)", r[y], r[z])
		}
		if x == 2 {
			if y == 0 {
				return fmt.Sprintf("add(c, c.loadA, c.load%v, false)", r[z])
			}
			if y == 1 {
				return fmt.Sprintf("add(c, c.loadA, c.load%v, true)", r[z])
			}
			if y == 2 {
				return fmt.Sprintf("sub(c, c.load%v, false)", r[z])
			}
			if y == 3 {
				return fmt.Sprintf("sub(c, c.load%v, true)", r[z])
			}
			if y == 4 {
				return fmt.Sprintf("and(c, c.load%v)", r[z])
			}
			if y == 5 {
				return fmt.Sprintf("xor(c, c.load%v)", r[z])
			}
			if y == 6 {
				return fmt.Sprintf("or(c, c.load%v)", r[z])
			}
			if y == 7 {
				return fmt.Sprintf("cp(c, c.load%v)", r[z])
			}
		}
		if x == 3 {
			if z == 0 {
				return fmt.Sprintf("ret(c, %v)", cc[y])
			}
			if z == 1 {
				if q == 0 {
					return fmt.Sprintf("pop(c, c.store%v)", rp2[p])
				}
				if q == 1 {
					if p == 0 {
						return "reta(c)"
					}
					if p == 1 {
						return "exx(c)"
					}
					if p == 2 {
						return fmt.Sprintf("jpa(c, c.load%v)", rp2[2])
					}
					if p == 3 {
						return fmt.Sprintf("ld16(c, c.storeSP, c.load%v)", rp2[2])
					}
				}
			}
			if z == 2 {
				return fmt.Sprintf("jp(c, %v, c.loadImm16)", cc[y])
			}
			if z == 3 {
				if y == 0 {
					return "jpa(c, c.loadImm16)"
				}
				if y == 1 && tab.name == "un" {
					return "cb(c)"
				}
				if y == 1 && tab.name != "un" {
					return ""
				}
				if y == 2 {
					// OUT (n), A
					return "ld(c, c.outIndImm, c.loadA)"
				}
				if y == 3 {
					// IN A, (n)
					return "ld(c, c.storeA, c.inIndImm)"
				}
				if y == 4 {
					return fmt.Sprintf("ex(c, c.load16IndSP, c.store16IndSP, c.load%v, c.store%v)", rp2[2], rp2[2])
				}
				if y == 5 {
					return "ex(c, c.loadDE, c.storeDE, c.loadHL, c.storeHL)"
				}
				if y == 6 {
					return "di(c)"
				}
				if y == 7 {
					return "ei(c)"
				}
			}
			if z == 4 {
				return fmt.Sprintf("call(c, %v, c.loadImm16)", cc[y])
			}
			if z == 5 {
				if q == 0 {
					return fmt.Sprintf("push(c, c.load%v)", rp2[p])
				}
				if q == 1 {
					if p == 0 {
						return "calla(c, c.loadImm16)"
					}
					if p == 1 && tab.name == "un" {
						return "ddfd(c, opsDD, opsDDCB)"
					}
					if p == 1 && tab.name != "un" {
						return "noni(c)"
					}
					if p == 2 {
						return "ed(c)"
					}
					if p == 3 && tab.name == "un" {
						return "ddfd(c, opsFD, opsFDCB)"
					}
					if p == 3 && tab.name != "un" {
						return "noni(c)"
					}
				}
			}
			if z == 6 {
				if y == 0 {
					return "add(c, c.loadImm, c.loadA, false)"
				}
				if y == 1 {
					return "add(c, c.loadImm, c.loadA, true)"
				}
				if y == 2 {
					return "sub(c, c.loadImm, false)"
				}
				if y == 3 {
					return "sub(c, c.loadImm, true)"
				}
				if y == 4 {
					return "and(c, c.loadImm)"
				}
				if y == 5 {
					return "xor(c, c.loadImm)"
				}
				if y == 6 {
					return "or(c, c.loadImm)"
				}
				if y == 7 {
					return "cp(c, c.loadImm)"
				}
			}
			if z == 7 {
				return fmt.Sprintf("rst(c, %v)", y)
			}
		}
	*/
	return ""
}

func processCB(tab *regtab, op uint8) string {
	r := tab.r
	x := int(rcs.SliceBits(op, 6, 7))
	y := int(rcs.SliceBits(op, 3, 5))
	z := int(rcs.SliceBits(op, 0, 2))

	if x == 0 {
		if y == 0 {
			return fmt.Sprintf("rlc(c, c.store%v, c.load%v)", r[z], r[z])
		}
		if y == 1 {
			// rrc
			return fmt.Sprintf("rotr(c, c.store%v, c.load%v)", r[z], r[z])
		}
		if y == 2 {
			// rl
			return fmt.Sprintf("shiftl(c, c.store%v, c.load%v, true)", r[z], r[z])
		}
		if y == 3 {
			// rr
			return fmt.Sprintf("shiftr(c, c.store%v, c.load%v, true)", r[z], r[z])
		}
		if y == 4 {
			// sla
			return fmt.Sprintf("shiftl(c, c.store%v, c.load%v, false)", r[z], r[z])
		}
		if y == 5 {
			return fmt.Sprintf("sra(c, c.store%v, c.load%v)", r[z], r[z])
		}
		if y == 6 {
			return fmt.Sprintf("sll(c, c.store%v, c.load%v)", r[z], r[z])
		}
		if y == 7 {
			// srl
			return fmt.Sprintf("shiftr(c, c.store%v, c.load%v, false)", r[z], r[z])
		}
	}
	if x == 1 {
		return fmt.Sprintf("bit(c, %v, c.load%v)", y, r[z])
	}
	if x == 2 {
		return fmt.Sprintf("res(c, %v, c.store%v, c.load%v)", y, r[z], r[z])
	}
	if x == 3 {
		return fmt.Sprintf("set(c, %v, c.store%v, c.load%v)", y, r[z], r[z])
	}
	return ""
}

func processED(tab *regtab, op uint8) string {
	r := tab.r
	rp := tab.rp
	x := int(rcs.SliceBits(op, 6, 7))
	y := int(rcs.SliceBits(op, 3, 5))
	z := int(rcs.SliceBits(op, 0, 2))
	p := int(rcs.SliceBits(op, 4, 5))
	q := int(rcs.SliceBits(op, 3, 3))

	if x == 0 || x == 3 {
		return "invalid()"
	}
	if x == 1 {
		if z == 0 {
			if y != 6 {
				// IN r[y], (C)
				return fmt.Sprintf("in(c, c.store%v, c.loadIndC)", r[y])
			}
			if y == 6 {
				// IN (C) ; undocumented
				return "in(c, c.storeNil, c.loadIndC)"
			}
		}
		if z == 1 {
			if y != 6 {
				// OUT (C), r[y]
				return fmt.Sprintf("ld(c, c.outIndC, c.load%v)", r[y])
			}
			if y == 6 {
				// OUT (C), 0 ; undocumented
				return "ld(c, c.outIndC, c.loadZero)"
			}
		}
		if z == 2 {
			if q == 0 {
				// SBC HL, rp[p]
				return fmt.Sprintf("sub16(c, c.storeHL, c.loadHL, c.load%v, true)", rp[p])
			}
			if q == 1 {
				// ADC HL, rp[p]
				return fmt.Sprintf("add16(c, c.storeHL, c.loadHL, c.load%v, true)", rp[p])
			}
		}
		if z == 3 {
			if q == 0 {
				return fmt.Sprintf("ld16(c, c.store16IndImm, c.load%v)", rp[p])
			}
			if q == 1 {
				return fmt.Sprintf("ld16(c, c.store%v, c.load16IndImm)", rp[p])
			}
		}
		if z == 4 {
			return "neg(c)"
		}
		if z == 5 {
			if y != 1 {
				return "retn(c)"
			}
			if y == 1 {
				return "reti(c)"
			}
		}
		if z == 6 {
			if y == 0 || y == 4 {
				return "im(c, 0)"
			}
			if y == 1 || y == 5 {
				return "im(c, 0)"
			}
			if y == 2 || y == 6 {
				return "im(c, 1)"
			}
			if y == 3 || y == 7 {
				return "im(c, 2)"
			}
		}
		if z == 7 {
			if y == 0 {
				return "ld(c, c.storeI, c.loadA)"
			}
			if y == 1 {
				return "ld(c, c.storeR, c.loadA)"
			}
			if y == 2 {
				// ld a, i
				return "ldair(c, c.loadI)"
			}
			if y == 3 {
				// ld a, r
				return "ldair(c, c.loadR)"
			}
			if y == 4 {
				return "rrd(c)"
			}
			if y == 5 {
				return "rld(c)"
			}
			return "nop()"
		}
	}
	if x == 2 {
		if z == 0 { // b == 0
			if y == 4 {
				// ldi
				return "blockl(c, 1)"
			}
			if y == 5 {
				// ldd
				return "blockl(c, -1)"
			}
			if y == 6 {
				// ldir
				return "blocklr(c, 1)"
			}
			if y == 7 {
				// lddr
				return "blocklr(c, -1)"
			}
		}
		if z == 1 { // b == 1
			if y == 4 {
				// cpi
				return "blockc(c, 1)"
			}
			if y == 5 {
				// cpd
				return "blockc(c, -1)"
			}
			if y == 6 {
				// cpir
				return "blockcr(c, 1)"
			}
			if y == 7 {
				// cpdr
				return "blockcr(c, -1)"
			}
		}
		if z == 2 {
			if y == 4 { // b == 2
				// ini
				return "blockin(c, 1)"
			}
			if y == 5 {
				// ind
				return "blockin(c, -1)"
			}
			if y == 6 {
				// inir
				return "blockinr(c, 1)"
			}
			if y == 7 {
				// indr
				return "blockinr(c, -1)"
			}
		}
		if z == 3 {
			if y == 4 {
				// outi
				return "blockout(c, 1)"
			}
			if y == 5 {
				// outd
				return "blockout(c, -1)"
			}
			if y == 6 {
				// otir
				return "blockoutr(c, 1)"
			}
			if y == 7 {
				// otdr
				return "blockoutr(c, -1)"
			}
		}
		return "invalid()"
	}
	return ""
}

func processXCB(tab *regtab, op uint8) string {
	r := tab.r
	x := int(rcs.SliceBits(op, 6, 7))
	y := int(rcs.SliceBits(op, 3, 5))
	z := int(rcs.SliceBits(op, 0, 2))
	// p := int(rcs.SliceBits(op, 4, 5))
	// q := int(rcs.SliceBits(op, 3, 3))

	if x == 0 {
		if z != 6 {
			if y == 0 {
				return fmt.Sprintf("rlc(c, c.store%v, c.load%v); ld(c, c.storeLastInd, c.load%v)", un.r[z], r[6], un.r[z])
			}
			if y == 1 {
				// rrc
				return fmt.Sprintf("rotr(c, c.store%v, c.load%v); ld(c, c.storeLastInd, c.load%v)", un.r[z], r[6], un.r[z])
			}
			if y == 2 {
				// rl
				return fmt.Sprintf("shiftl(c, c.store%v, c.load%v, true); ld(c, c.storeLastInd, c.load%v)", un.r[z], r[6], un.r[z])
			}
			if y == 3 {
				// rr
				return fmt.Sprintf("shiftr(c, c.store%v, c.load%v, true); ld(c, c.storeLastInd, c.load%v)", un.r[z], r[6], un.r[z])
			}
			if y == 4 {
				// sla
				return fmt.Sprintf("shiftl(c, c.store%v, c.load%v, false); ld(c, c.storeLastInd, c.load%v)", un.r[z], r[6], un.r[z])
			}
			if y == 5 {
				return fmt.Sprintf("sra(c, c.store%v, c.load%v); ld(c, c.storeLastInd, c.load%v)", un.r[z], r[6], un.r[z])
			}
			if y == 6 {
				return fmt.Sprintf("sll(c, c.store%v, c.load%v); ld(c, c.storeLastInd, c.load%v)", un.r[z], r[6], un.r[z])
			}
			if y == 7 {
				// srl
				return fmt.Sprintf("shiftr(c, c.store%v, c.load%v, false); ld(c, c.storeLastInd, c.load%v)", un.r[z], r[6], un.r[z])
			}
		}
		if z == 6 {
			if y == 0 {
				return fmt.Sprintf("rlc(c, c.storeLastInd, c.load%v)", r[6])
			}
			if y == 1 {
				// rrc
				return fmt.Sprintf("rotr(c, c.storeLastInd, c.load%v)", r[6])
			}
			if y == 2 {
				// rl
				return fmt.Sprintf("shiftl(c, c.storeLastInd, c.load%v, true)", r[6])
			}
			if y == 3 {
				// rr
				return fmt.Sprintf("shiftr(c, c.storeLastInd, c.load%v, true)", r[6])
			}
			if y == 4 {
				// sla
				return fmt.Sprintf("shiftl(c, c.storeLastInd, c.load%v, false)", r[6])
			}
			if y == 5 {
				return fmt.Sprintf("sra(c, c.storeLastInd, c.load%v)", r[6])
			}
			if y == 6 {
				return fmt.Sprintf("sll(c, c.storeLastInd, c.load%v)", r[6])
			}
			if y == 7 {
				// srl
				return fmt.Sprintf("shiftr(c, c.storeLastInd, c.load%v, false)", r[6])
			}
		}
	}
	if x == 1 {
		return fmt.Sprintf("biti(c, %v, c.load%v)", y, r[6])
	}
	if x == 2 {
		if z != 6 {
			return fmt.Sprintf("res(c, %v, c.store%v, c.load%v); ld(c, c.storeLastInd, c.load%v)", y, un.r[z], r[6], un.r[z])
		}
		if z == 6 {
			return fmt.Sprintf("res(c, %v, c.storeLastInd, c.load%v)", y, r[6])
		}
	}
	if x == 3 {
		if z != 6 {
			return fmt.Sprintf("set(c, %v, c.store%v, c.load%v); ld(c, c.storeLastInd, c.load%v)", y, un.r[z], r[6], un.r[z])
		}
		if z == 6 {
			return fmt.Sprintf("set(c, %v, c.storeLastInd, c.load%v)", y, r[6])
		}
	}
	return ""
}

func process(out *bytes.Buffer, getFn func(*regtab, uint8) string, tab *regtab) {
	for i := 0; i < 0x100; i++ {
		fn := getFn(tab, uint8(i))
		if fn == "" {
			continue
		}
		/*
			if fn == "" {
				fn = "panic(\"unhandled instruction\")"
			}
		*/
		emit := true
		if tab.name == "dd" || tab.name == "fd" {
			// If there is an indirect call, the next byte needs to be
			// fetched for the displacement
			if strings.Contains(fn, "IndIX") || strings.Contains(fn, "IndIY") {
				fn = "c.fetchd(); " + fn
				// "any other instances of H and L will be unaffected".
				// If (HL) was transformed, use H and L instead.
				fn = strings.Replace(fn, "IXH", "H", -1)
				fn = strings.Replace(fn, "IXL", "L", -1)
				fn = strings.Replace(fn, "IYH", "H", -1)
				fn = strings.Replace(fn, "IYL", "L", -1)
			}
			// If the unprefixed version of the op is the same as the
			// prefixed one, just leave the function blank
			unfn := getFn(un, uint8(i))
			if unfn == fn {
				emit = false
			}
		}

		line := fmt.Sprintf("0x%02x: func(c *CPU){%v},\n", i, fn)
		if !emit {
			line = fmt.Sprintf("0x%02x: nil,\n", i)
		}

		out.WriteString(line)
	}
}

func main() {
	var out bytes.Buffer

	out.WriteString(`
// Code generated by gen/z80/opcodes/opcodes.go. DO NOT EDIT.

package z80

`)
	out.WriteString("var opcodes = map[uint8]func(c *CPU){\n")
	process(&out, processMain, un)
	out.WriteString("}\n")

	/*
		out.WriteString("var opsCB = map[uint8]func(c *CPU){\n")
		process(&out, processCB, un)
		out.WriteString("}\n")

		out.WriteString("var opsED = map[uint8]func(c *CPU){\n")
		process(&out, processED, un)
		out.WriteString("}\n")

		out.WriteString("var opsDD = map[uint8]func(c *CPU){\n")
		process(&out, processMain, dd)
		out.WriteString("}\n")

		out.WriteString("var opsFD = map[uint8]func(c *CPU){\n")
		process(&out, processMain, fd)
		out.WriteString("}\n")

		out.WriteString("var opsDDCB = map[uint8]func(c *CPU){\n")
		process(&out, processXCB, ddcb)
		out.WriteString("}\n")

		out.WriteString("var opsFDCB = map[uint8]func(c *CPU){\n")
		process(&out, processXCB, fdcb)
		out.WriteString("}\n")
	*/

	filename := filepath.Join(targetDir, "opcodes.go")
	err := ioutil.WriteFile(filename, out.Bytes(), 0644)
	if err != nil {
		fmt.Printf("unable to write file: %v", err)
		os.Exit(1)
	}
}
