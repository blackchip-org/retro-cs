package z80

import (
	"fmt"
	"strings"

	"github.com/blackchip-org/retro-cs/rcs"
)

func Reader(e rcs.StmtEval) {
	e.Stmt.Addr = e.Ptr.Addr()
	opcode := e.Ptr.Fetch()
	e.Stmt.Bytes = []uint8{opcode}
	dasmTable[opcode](e)
}

func Formatter() rcs.CodeFormatter {
	options := rcs.FormatOptions{
		BytesFormat: "%-11s",
	}
	return func(s rcs.Stmt) string {
		return rcs.FormatStmt(s, options)
	}
}

func NewDisassembler(mem *rcs.Memory) *rcs.Disassembler {
	return rcs.NewDisassembler(mem, Reader, Formatter())
}

func op1(e rcs.StmtEval, parts ...string) {
	var out strings.Builder
	for i, part := range parts {
		v := part
		switch {
		case i == 0:
			v = fmt.Sprintf("%-4s", part)
		case parts[0] == "rst" && i == 1:
			// Reset statements have the argment encoded in the opcode. Change
			// the hex notation from & to $ in the second part
			v = "$" + v[1:]
		case part == "&4546":
			// This is an address that is a 8-bit displacement from the
			// current program counter
			delta := e.Ptr.Fetch()
			e.Stmt.Bytes = append(e.Stmt.Bytes, delta)
			addr := displace(e.Stmt.Addr+2, delta)
			v = fmt.Sprintf("$%04x", addr)
		case part == "&0000":
			lo := e.Ptr.Fetch()
			e.Stmt.Bytes = append(e.Stmt.Bytes, lo)
			hi := e.Ptr.Fetch()
			e.Stmt.Bytes = append(e.Stmt.Bytes, hi)
			addr := int(hi)<<8 | int(lo)
			v = fmt.Sprintf("$%04x", addr)
		case part == "(&0000)":
			lo := e.Ptr.Fetch()
			e.Stmt.Bytes = append(e.Stmt.Bytes, lo)
			hi := e.Ptr.Fetch()
			e.Stmt.Bytes = append(e.Stmt.Bytes, hi)
			addr := int(hi)<<8 | int(lo)
			v = fmt.Sprintf("($%04x)", addr)
		case part == "&00":
			arg := e.Ptr.Fetch()
			e.Stmt.Bytes = append(e.Stmt.Bytes, arg)
			v = fmt.Sprintf("$%02x", arg)
		case part == "(&00)":
			arg := e.Ptr.Fetch()
			e.Stmt.Bytes = append(e.Stmt.Bytes, arg)
			v = fmt.Sprintf("($%02x)", arg)
		case part == "(ix+0)":
			delta := e.Ptr.Fetch()
			e.Stmt.Bytes = append(e.Stmt.Bytes, delta)
			v = fmt.Sprintf("(ix+$%02x)", delta)
		case part == "(iy+0)":
			delta := e.Ptr.Fetch()
			e.Stmt.Bytes = append(e.Stmt.Bytes, delta)
			v = fmt.Sprintf("(iy+$%02x)", delta)
		}

		if i == 1 {
			out.WriteString(" ")
		}
		if i == 2 {
			out.WriteString(",")
		}
		out.WriteString(v)
	}
	e.Stmt.Op = strings.TrimSpace(out.String())
}

func op2(e rcs.StmtEval, parts ...string) {
	var out strings.Builder
	for i, part := range parts {
		v := part
		switch {
		case i == 0:
			v = fmt.Sprintf("%-4s", part)
		case part == "(ix+0)":
			delta := e.Stmt.Bytes[len(e.Stmt.Bytes)-2]
			v = fmt.Sprintf("(ix+$%02x)", delta)
		case part == "(iy+0)":
			delta := e.Stmt.Bytes[len(e.Stmt.Bytes)-2]
			v = fmt.Sprintf("(iy+$%02x)", delta)
		}

		if i == 1 {
			out.WriteString(" ")
		}
		if i == 2 {
			out.WriteString(",")
		}
		out.WriteString(v)
	}
	e.Stmt.Op = strings.TrimSpace(out.String())
}

func opDD(e rcs.StmtEval) {
	next := e.Ptr.Peek()
	if next == 0xdd || next == 0xed || next == 0xfd {
		e.Stmt.Op = "?dd"
		return
	}
	opcode := e.Ptr.Fetch()
	e.Stmt.Bytes = append(e.Stmt.Bytes, opcode)
	if opcode == 0xcb {
		opDDCB(e)
		return
	}
	dasmTableDD[opcode](e)
}

func opFD(e rcs.StmtEval) {
	next := e.Ptr.Peek()
	if next == 0xdd || next == 0xed || next == 0xfd {
		e.Stmt.Op = "?fd"
		return
	}
	opcode := e.Ptr.Fetch()
	e.Stmt.Bytes = append(e.Stmt.Bytes, opcode)
	if opcode == 0xcb {
		opFDCB(e)
		return
	}
	dasmTableFD[opcode](e)
}

func opCB(e rcs.StmtEval) {
	opcode := e.Ptr.Fetch()
	e.Stmt.Bytes = append(e.Stmt.Bytes, opcode)
	dasmTableCB[opcode](e)
}

func opED(e rcs.StmtEval) {
	opcode := e.Ptr.Fetch()
	e.Stmt.Bytes = append(e.Stmt.Bytes, opcode)
	dasmTableED[opcode](e)
}

func opFDCB(e rcs.StmtEval) {
	delta := e.Ptr.Fetch()
	e.Stmt.Bytes = append(e.Stmt.Bytes, delta)
	opcode := e.Ptr.Fetch()
	e.Stmt.Bytes = append(e.Stmt.Bytes, opcode)
	dasmTableFDCB[opcode](e)
}

func opDDCB(e rcs.StmtEval) {
	delta := e.Ptr.Fetch()
	e.Stmt.Bytes = append(e.Stmt.Bytes, delta)
	opcode := e.Ptr.Fetch()
	e.Stmt.Bytes = append(e.Stmt.Bytes, opcode)
	dasmTableDDCB[opcode](e)
}

func displace(value int, delta uint8) uint16 {
	sdelta := int8(delta)
	v := int(value) + int(sdelta)
	return uint16(v)
}
