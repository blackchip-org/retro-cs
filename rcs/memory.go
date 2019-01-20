package rcs

import "log"

// Memory represents an address space used to access RAM, ROM, IO ports,
// and external devices.
type Memory struct {
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
		reads:  make([][]Load8, banks, banks),
		writes: make([][]Store8, banks, banks),
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
func (m *Memory) MapROM(addr int, rom []uint8) {
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
