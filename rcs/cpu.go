package rcs

import (
	"fmt"
	"strings"
)

// CPU is a central processing unit.
//
// The program counter is an integer to accomodate address busses of at
// least 32-bit. The program counter stored within the actual struct
// should proablby be the actual size of the bus.
//
// Offset is for CPUs, like the 6502, that increment the program counter
// before fetching the instruction opcode. In this case, one should be
// returned. If the program counter is incremented after the fetch, zero
// should be returned.
type CPU interface {
	Next() (pc int, halt bool) // Execute the next instruction
	PC() int                   // Address of the program counter
	SetPC(int)                 // Set the address of the program counter
	Offset() int               // The next instruction is at PC() + Offset()
}

// CPUEditor allow editing of register and flag values for CPUs that
// support these methods.
type CPUEditor interface {
	Registers() map[string]Value
	Flags() map[string]Value
}

// CPUDisassembler provides a disassembler instance for CPUs that
// support this method.
type CPUDisassembler interface {
	NewDisassembler() *Disassembler
}

// Stmt represents a single statement in a disassembly.
type Stmt struct {
	Addr    int     // Address of the instruction
	Label   string  // Label for this address, "CHROUT"
	Op      string  // Formated operation, "lda #$40"
	Bytes   []uint8 // Bytes that represent this instruction
	Comment string  // Any notes from the source code
}

// CodeReader reads the next instruction using provided pointer and
// fills in the fields found in the Stmt.
type CodeReader func(StmtEval)

// CodeFormat formats a statement in a string suitible for display to the
// end user.
type CodeFormatter func(Stmt) string

type Disassembler struct {
	mem    *Memory
	ptr    *Pointer
	read   CodeReader
	format CodeFormatter
}

type StmtEval struct {
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
	eval := StmtEval{
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
	d.ptr.SetAddr(addr)
}

func (d *Disassembler) PC() int {
	return d.ptr.Addr()
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
