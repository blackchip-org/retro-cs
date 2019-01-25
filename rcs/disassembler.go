package rcs

import (
	"fmt"
	"strings"
)

type Statement struct {
	Addr    int
	Label   string
	Op      string
	Bytes   []uint8
	Comment string
}

func NewStatement() *Statement {
	return &Statement{Bytes: make([]uint8, 0, 0)}
}

type CodeReader func(Eval) Statement
type CodeFormatter func(Statement) string

type Disassembler struct {
	mem    *Memory
	ptr    *Pointer
	read   CodeReader
	format CodeFormatter
}

type Eval struct {
	Ptr  *Pointer
	Stmt *Statement
}

func NewDisassembler(mem *Memory, r CodeReader, f CodeFormatter) *Disassembler {
	return &Disassembler{
		mem:    mem,
		ptr:    NewPointer(mem),
		read:   r,
		format: f,
	}
}

func (d *Disassembler) NextStatement() Statement {
	return d.read(Eval{
		Ptr:  d.ptr,
		Stmt: NewStatement(),
	})
}

func (d *Disassembler) Next() string {
	return d.format(d.NextStatement())
}

func (d *Disassembler) SetPC(addr int) {
	d.ptr.Addr = addr
}

func (d *Disassembler) PC() int {
	return d.ptr.Addr
}

type FormatOptions struct {
	BytesFormat string
}

func Format(s Statement, options FormatOptions) string {
	bytes := []string{}
	for _, b := range s.Bytes {
		bytes = append(bytes, fmt.Sprintf("%02x", b))
	}
	format := options.BytesFormat
	if format == "" {
		format = "%v"
	}
	sbytes := fmt.Sprintf(format, strings.Join(bytes, " "))
	return fmt.Sprintf("$%04x:  %s  %s", s.Addr, sbytes, s.Op)
}
