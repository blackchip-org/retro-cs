package rcs

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

/*
Memory represents an address space used to access RAM, ROM, IO ports,
and external devices.

To create a memory space with a 16 line address bus and a single bank, use:

	mem := rcs.NewMemory(1, 0x10000)

This struct is just a container for the address space and has no actual
memory mapped to it yet. Reads or writes to an unmapped address emit a
warning through the standard logger.

Single values are mapped for read/write access using the MapRW method. This is
useful for mapping ports or registers of a device into the address space. In
this example from the Commodore 64, the X coordinate for sprite #0 is mapped to
0xd000 and the Y coordinate is mapped to 0xd001:

	mem.MapRW(0xd000, &sprites[0].X)
	mem.MapRW(0xd001, &sprites[0].Y)

Use MapRO for a read-only mapping and MapWO for a write-only mapping. In
this example from Pac-Man, the value of port IN0 is read through address
0x5000 but interrupts can be enabled or disabled by writing to address
0x5000:

	mem.MapRO(0x5000, &portIN0)
	mem.MapWO(0x5000, &irqEnable)

A range of addresses can map to the same value with:

	for i := 0x50c0; i <= 0x50ff; i++ {
		mem.MapWO(i, &watchdogReset)
	}

Large blocks can be mapped by passing in a uint8 slice using MapRAM
for read/write access and MapROM for read-only access. The following example
maps a 16KB block of ROM to 0x0000 - 0x3fff and a 48KB block of RAM
to 0x4000 - 0xffff:

	rom := []uint8{ ... 16KB data ... }
	ram := make([]uint8, 48*1024, 48*1024)
	mem.MapROM(0x0000, rom)
	mem.MapRAM(0x4000, ram)

To "overlay" ROM on top of RAM, first use MapRAM and then MapROM which will
replace the read mappings while leaving the write mappings untouched. In
the following example, reads in the first 16KB come from the ROM but
writes go to the RAM:

	rom := []uint8{ ... 16KB data ... }
	ram := make([]uint8, 64*1024, 64*1024)
	mem.MapRAM(0x0000, ram)
	mem.MapROM(0x0000, rom)

To use banked memory, create memory with more than one bank:

	mem := rcs.NewMemory(2, 0x10000)

All map, read, and write operations are performed on the selected bank
which by default is zero. The following example has a ROM overlay
in bank 0 and full access to the RAM in bank 1:

	rom := []uint8{ ... 16KB data ... }
	ram := make([]uint8, 64*1024, 64*1024)

	mem.SetBank(0)
	mem.MapRAM(0x0000, ram)
	mem.MapROM(0x0000, rom)

	mem.SetBank(1)
	mem.MapRAM(0x0000, ram)
*/
type Memory struct {
	MaxAddr int // maximum valid address

	// read and write functions for each bank
	reads  [][]Load8
	writes [][]Store8

	// selected bank index
	bank int

	// read and write functions for the selected bank
	read  []Load8
	write []Store8
}

// NewMemory creates a memory space of uint8 values that are addressable
// from 0 to size - 1. This function only creates the address space;
// values must be mapped using the Map methods. To create banked memory, use a
// value greater than one for the banks argument.
func NewMemory(banks int, size int) *Memory {
	if banks < 1 {
		banks = 1
	}
	mem := &Memory{
		MaxAddr: size - 1,
		reads:   make([][]Load8, banks, banks),
		writes:  make([][]Store8, banks, banks),
	}
	for b := 0; b < banks; b++ {
		mem.reads[b] = make([]Load8, size, size)
		mem.writes[b] = make([]Store8, size, size)
		for addr := 0; addr < size; addr++ {
			mem.reads[b][addr] = warnUnmappedRead(b, addr)
			mem.writes[b][addr] = warnUnmappedWrite(b, addr)
		}
	}
	mem.read = mem.reads[0]
	mem.write = mem.writes[0]
	return mem
}

// Read returns the 8-bit value at the given address.
func (m *Memory) Read(addr int) uint8 {
	return m.read[addr]()
}

// Write sets the 8-bit value at the given address.
func (m *Memory) Write(addr int, val uint8) {
	m.write[addr](val)
}

// WriteN sets multiple 8-bit values starting with the given address.
func (m *Memory) WriteN(addr int, values ...uint8) {
	for i, val := range values {
		m.write[addr+i](val)
	}
}

// ReadLE returns the 16-bit value at addr and addr+1 stored in little endian
// byte order.
func (m *Memory) ReadLE(addr int) int {
	lo := int(m.Read(addr))
	hi := int(m.Read(addr + 1))
	return hi<<8 + lo
}

// WriteLE puts a 16-bit value at addr and addr+1 stored in little endian
// byte order.
func (m *Memory) WriteLE(addr int, val int) {
	hi := uint8(val >> 8)
	lo := uint8(val)
	m.Write(addr, lo)
	m.Write(addr+1, hi)
}

// MapRAM adds read/write maps to all of the 8-bit values in ram starting at
// addr. Any existing read or write maps are replaced.
func (m *Memory) MapRAM(addr int, ram []uint8) {
	for i := 0; i < len(ram); i++ {
		j := i
		m.read[addr+i] = func() uint8 { return ram[j] }
		m.write[addr+i] = func(v uint8) { ram[j] = v }
	}
}

// MapROM adds read maps to all of the 8-bit values in rom starting at
// addr. Any existing read maps are replaced but write maps are not altered.
// If the rom passed in is nil, no mappings are changed.
func (m *Memory) MapROM(addr int, rom []uint8) {
	if rom == nil {
		return
	}
	for i := 0; i < len(rom); i++ {
		j := i
		m.read[addr+i] = func() uint8 { return rom[j] }
	}
}

// MapRW adds a read and write to the given 8-bit value at addr. Any existing
// mappings are replaced.
func (m *Memory) MapRW(addr int, b *uint8) {
	m.MapRO(addr, b)
	m.MapWO(addr, b)
}

// MapRO adds a read mapping to the given 8-bit value at addr. If there is
// already a read mapping, it is replaced. Write mappings are not altered.
func (m *Memory) MapRO(addr int, b *uint8) {
	m.read[addr] = func() uint8 { return *b }
}

// MapWO adds a write mapping to the given 8-bit value at addr. If there is
// already a write mapping, it is replaced. Read mappings are not altered.
func (m *Memory) MapWO(addr int, b *uint8) {
	m.write[addr] = func(v uint8) { *b = v }
}

// MapLoad adds a read mapping to the given function. When this address is
// read from, the function is invoked to get the value. If there is already a
// read mapping for this address, it is replaced. Write mappings are not
// altered.
func (m *Memory) MapLoad(addr int, load Load8) {
	m.read[addr] = load
}

// MapStore adds a write mapping to the given function. When this address is
// written to, the function is invoked with the value to write. If there
// is already a write mapping for this address, it is replaced. Read mappings
// are not altered.
func (m *Memory) MapStore(addr int, store Store8) {
	m.write[addr] = store
}

func (m *Memory) Map(startAddr int, m1 *Memory) {
	endAddr := startAddr + len(m1.read)
	for i, addr := 0, startAddr; addr < endAddr; i, addr = i+1, addr+1 {
		m.read[addr] = m1.read[i]
		m.write[addr] = m1.write[i]
	}
}

func (m *Memory) Unmap(addr int) {
	m.read[addr] = warnUnmappedRead(m.bank, addr)
	m.write[addr] = warnUnmappedWrite(m.bank, addr)
}

func (m *Memory) MapNil(addr int) {
	m.read[addr] = func() uint8 { return 0 }
	m.write[addr] = func(uint8) {}
}

// Bank returns the number of the selected bank. Banks are numbered starting
// with zero.
func (m *Memory) Bank() int {
	return m.bank
}

// SetBank changes the selected bank. Banks are numbered starting with zero.
func (m *Memory) SetBank(bank int) {
	m.bank = bank
	m.read = m.reads[bank]
	m.write = m.writes[bank]
}

func warnUnmappedRead(bank int, addr int) Load8 {
	return func() uint8 {
		log.Printf("unmapped memory read, bank %v, addr 0x%x", bank, addr)
		return 0
	}
}

func warnUnmappedWrite(bank int, addr int) Store8 {
	return func(v uint8) {
		log.Printf("unmapped memory write, bank %v, addr 0x%x, value 0x%x", bank, addr, v)
	}
}

// Pointer points to a location in memory.
type Pointer struct {
	addr int     // Current position.
	Mask int     // Address mask
	Mem  *Memory // Memory view
}

// NewPointer creates pointer at address zero on the provided memory.
func NewPointer(mem *Memory) *Pointer {
	return &Pointer{Mem: mem, Mask: 0xffff}
}

func (p *Pointer) Addr() int {
	return p.addr
}

func (p *Pointer) SetAddr(addr int) {
	p.addr = addr & p.Mask
}

// Fetch returns the byte at current position as an 8-bit value and advances
// the pointer by one.
func (p *Pointer) Fetch() uint8 {
	value := p.Mem.Read(p.addr)
	p.addr = (p.addr + 1) & p.Mask
	return value
}

// Peek returns the byte at the current position as an 8-bit value. The
// pointer is not moved.
func (p *Pointer) Peek() uint8 {
	return p.Mem.Read(p.addr)
}

// FetchLE returns the next two bytes as a 16-bit value stored in little
// endian format and advances the pointer by two.
func (p *Pointer) FetchLE() int {
	lo := int(p.Fetch())
	hi := int(p.Fetch())
	return hi<<8 + lo
}

// Put sets the value at the current address and advances the pointer by one.
func (p *Pointer) Put(value uint8) {
	p.Mem.Write(p.addr, value)
	p.addr = (p.addr + 1) & p.Mask
}

// PutN calls Put for each value.
func (p *Pointer) PutN(values ...uint8) {
	for _, value := range values {
		p.Put(value)
	}
}

/*
ROM represents a dump of read-only memory data found on disk.

To load in a set of ROMs, define a slice of ROM definitions each with
the "name" of the ROM, the filename found on disk, and its SHA1
checksum as a string of hexadecimal values.

	roms := []rcs.ROM{
		NewROM("code"  ", "chip1.rom  ", "da39a3ee5e6b4b0d3255bfef95601890afd80709"),
		NewROM("code   ", "chip2.rom  ", "342d21fb707e89a6d407117c810795abcc481c52"),
		NewROM("sprites", "chip3az.rom", "38dc6d5fe1a085f2e885748a00fbe5b7b3b8b1a6"),
	}

Then call LoadROMs to return a map of names to byte slices:

	data, err := LoadROMs("/path/to/roms", roms)

*/
type ROM struct {
	Name     string
	File     string
	Checksum string
}

// NewROM creates a new ROM definition.
func NewROM(name string, file string, checksum string) ROM {
	return ROM{
		Name:     strings.TrimSpace(name),
		File:     strings.TrimSpace(file),
		Checksum: checksum,
	}
}

var readFile = ioutil.ReadFile

/*
LoadROMs loads all ROMs in the given slice of definitions. An attempt will be
made to load each ROM. Any errors encountered during the attempts are be
combined into a single error.

The returned map will contain a byte slice for each ROM identified by its name.
ROMS that are given the same name are concatenated together. Extra whitespace
found at the beginning or ending of the name or filename are removed and
is useful for aligning the ROM definitions in the source code.
*/
func LoadROMs(dir string, roms []ROM) (map[string][]byte, error) {
	buffers := make(map[string]bytes.Buffer)
	e := make([]string, 0, 0)
	for _, rom := range roms {
		buf := buffers[rom.Name]
		path := filepath.Join(dir, rom.File)
		data, err := readFile(path)
		if err != nil {
			e = append(e, err.Error())
			continue
		}
		checksum := fmt.Sprintf("%040x", sha1.Sum(data))
		if checksum != rom.Checksum {
			e = append(e, fmt.Sprintf("%v: invalid checksum", path))
			continue
		}
		buf.Write(data)
		buffers[rom.Name] = buf
	}
	if len(e) > 0 {
		return nil, errors.New(strings.Join(e, "\n"))
	}
	chunks := make(map[string][]byte)
	for name, buf := range buffers {
		chunks[name] = buf.Bytes()
	}
	return chunks, nil
}
