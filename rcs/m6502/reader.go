package m6502

import (
	"fmt"
	"strings"

	"github.com/blackchip-org/retro-cs/rcs"
)

func Reader(e rcs.Eval) rcs.Statement {
	e.Stmt.Addr = e.Ptr.Addr
	opcode := e.Ptr.Fetch()
	e.Stmt.Bytes = append(e.Stmt.Bytes, opcode)
	op, ok := dasmTable[opcode]
	if !ok {
		e.Stmt.Op = fmt.Sprintf("?%02x", opcode)
		return *e.Stmt
	}

	len := operandLengths[op.mode]
	operand := 0
	switch len {
	case 1:
		operand = int(e.Ptr.Fetch())
		e.Stmt.Bytes = append(e.Stmt.Bytes, uint8(operand))
	case 2:
		operand = e.Ptr.FetchLE()
		e.Stmt.Bytes = append(e.Stmt.Bytes, uint8(operand), uint8(operand>>8))
	}
	e.Stmt.Op = op.inst + formatOp(op, operand, e.Stmt.Addr)
	return *e.Stmt
}

func Formatter() rcs.CodeFormatter {
	options := rcs.FormatOptions{
		BytesFormat: "%-8s",
	}
	return func(s rcs.Statement) string {
		return rcs.Format(s, options)
	}
}

func formatOp(op op, operand int, addr int) string {
	format, ok := operandFormats[op.mode]
	result := ""
	if ok {
		// If this is a branch instruction, the value of the operand needs to be
		// added to the current addresss. Add two as it is relative after consuming
		// the instruction
		value := operand
		if op.mode == relative {
			value8 := int8(value)
			if value8 >= 0 {
				value = addr + int(value8) + 2
			} else {
				value = addr - int(value8*-1) + 2
			}
		}
		// If the format does not contain a formatting directive, just use as is.
		// For example: "asl a"
		if strings.Contains(format, "%") {
			result = " " + fmt.Sprintf(format, value)
		} else {
			result = " " + format
		}
	}
	return result
}
