package rcs

import (
	"fmt"
	"strings"
)

type Proc interface {
	Next()
}

type CPU interface {
	Next()
	PC() int
	SetPC(int)
}

// Stmt represents a single statement in a disassembly.
type Stmt struct {
	Addr    int     // Address of the instruction
	Label   string  // Jump label
	Op      string  // Formated operation, "lda #$40"
	Bytes   []uint8 // Bytes that represent this instruction
	Comment string  // Any notes from the source code
}

type CodeReader func(Eval)
type CodeFormatter func(Stmt) string

type Disassembler struct {
	mem    *Memory
	ptr    *Pointer
	read   CodeReader
	format CodeFormatter
}

type Eval struct {
	Ptr  *Pointer
	Stmt *Stmt
}

func NewDisassembler(mem *Memory, r CodeReader, f CodeFormatter) *Disassembler {
	return &Disassembler{
		mem:    mem,
		ptr:    NewPointer(mem),
		read:   r,
		format: f,
	}
}

func (d *Disassembler) NextStmt() Stmt {
	eval := Eval{
		Ptr: d.ptr,
		Stmt: &Stmt{
			Bytes: make([]byte, 0, 0),
		},
	}
	d.read(eval)
	return *eval.Stmt
}

func (d *Disassembler) Next() string {
	return d.format(d.NextStmt())
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

func FormatStmt(s Stmt, options FormatOptions) string {
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
