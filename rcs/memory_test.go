package rcs

import (
	"bytes"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestMemoryUnmapped(t *testing.T) {
	var buf bytes.Buffer
	log.SetFlags(0)
	log.SetOutput(&buf)
	defer func() {
		log.SetFlags(log.LstdFlags)
		log.SetOutput(os.Stderr)
	}()

	mem := NewMemory(1, 0x10000)
	mem.Write(0x1234, 0xaa)
	mem.Read(0x5678)

	msg := []string{
		"unmapped memory write, bank 0, addr 0x1234, value 0xaa",
		"unmapped memory read, bank 0, addr 0x5678",
		"",
	}
	have := buf.String()
	want := strings.Join(msg, "\n")
	if have != want {
		t.Errorf("\n have: \n%v \n want: \n%v", have, want)
	}
}

func TestMemoryRAM(t *testing.T) {
	in := []uint8{10, 11, 12, 13, 14}
	ram := make([]uint8, 5, 5)
	out := make([]uint8, 5, 5)
	mem := NewMemory(1, 15)
	mem.MapRAM(10, ram)

	for i := 0; i < 5; i++ {
		mem.Write(i+10, in[i])
	}
	for i := 0; i < 5; i++ {
		out[i] = mem.Read(i + 10)
	}
	if !reflect.DeepEqual(out, in) {
		t.Errorf("\n have: %v \n want: %v", out, in)
	}
}

func TestMemoryROM(t *testing.T) {
	var buf bytes.Buffer
	log.SetFlags(0)
	log.SetOutput(&buf)
	defer func() {
		log.SetFlags(log.LstdFlags)
		log.SetOutput(os.Stderr)
	}()

	rom := []uint8{10, 11, 12, 13, 14}
	out := make([]uint8, 5, 5)
	mem := NewMemory(1, 15)
	mem.MapROM(10, rom)

	for i := 0; i < 5; i++ {
		mem.Write(i+10, 0xff)
		out[i] = mem.Read(i + 10)
	}
	if !reflect.DeepEqual(out, rom) {
		t.Errorf("\n have: %v \n want: %v", out, rom)
	}

	msg := []string{
		"unmapped memory write, bank 0, addr 0xa, value 0xff",
		"unmapped memory write, bank 0, addr 0xb, value 0xff",
		"unmapped memory write, bank 0, addr 0xc, value 0xff",
		"unmapped memory write, bank 0, addr 0xd, value 0xff",
		"unmapped memory write, bank 0, addr 0xe, value 0xff",
		"",
	}
	have := buf.String()
	want := strings.Join(msg, "\n")
	if have != want {
		t.Errorf("\n have: \n%v \n want: \n%v", have, want)
	}
}

func TestMemoryMapValue(t *testing.T) {
	in := []uint8{10, 11, 12, 13, 14}
	ram := make([]uint8, 5, 5)
	out := make([]uint8, 5, 5)

	mem := NewMemory(1, 15)
	mem.MapRW(10, &in[0])
	mem.MapRW(11, &in[1])
	mem.MapRW(12, &in[2])
	mem.MapRW(13, &in[3])
	mem.MapRW(14, &in[4])

	for i := 0; i < 5; i++ {
		mem.Write(i+10, ram[i])
	}
	for i := 0; i < 5; i++ {
		out[i] = mem.Read(i + 10)
	}
	if !reflect.DeepEqual(out, in) {
		t.Errorf("\n have: %v \n want: %v", out, in)
	}
}

func TestMemoryMapFunc(t *testing.T) {
	reads := 0
	writes := 0
	out := make([]uint8, 5, 5)

	mem := NewMemory(1, 15)
	for i := 0; i < 5; i++ {
		j := i
		mem.MapGet(i+10, func() uint8 { reads++; return 40 + uint8(j) })
		mem.MapPut(i+10, func(uint8) { writes++ })
	}
	for i := 0; i < 5; i++ {
		out[i] = mem.Read(i + 10)
		mem.Write(i+10, 99)
	}
	want := []uint8{40, 41, 42, 43, 44}
	if !reflect.DeepEqual(out, want) {
		t.Errorf("\n have: %v \n want: %v", out, want)
	}
	if reads != 5 {
		t.Errorf("expected 5 reads, got %v", reads)
	}
	if writes != 5 {
		t.Errorf("expected 5 writes, got %v", writes)
	}
}

func TestMemoryReadLE(t *testing.T) {
	mem := NewMemory(1, 2)
	mem.MapROM(0, []uint8{0xcd, 0xab})

	have := mem.ReadLE(0)
	want := 0xabcd
	if want != have {
		t.Errorf("\n have: %04x \n want: %04x", have, want)
	}
}

func TestMemoryWriteLE(t *testing.T) {
	mem := NewMemory(1, 2)
	ram := make([]uint8, 2, 2)
	mem.MapRAM(0, ram)

	mem.WriteLE(0, 0xabcd)
	want := []uint8{0xcd, 0xab}
	if !reflect.DeepEqual(ram, want) {
		t.Errorf("\n have: %v \n want: %v", ram, want)
	}
}

func TestMemoryBank(t *testing.T) {
	mem := NewMemory(2, 2)
	ram0 := []uint8{10, 0}
	ram1 := []uint8{30, 0}
	mem.MapRAM(0, ram0)
	mem.SetBank(1)
	mem.MapRAM(0, ram1)

	out := make([]uint8, 4, 4)
	mem.SetBank(0)
	mem.Write(1, 20)
	out[0] = mem.Read(0)
	out[1] = mem.Read(1)

	mem.SetBank(1)
	mem.Write(1, 40)
	out[2] = mem.Read(0)
	out[3] = mem.Read(1)

	want := []uint8{10, 20, 30, 40}
	if !reflect.DeepEqual(out, want) {
		t.Errorf("\n have: %v \n want: %v", out, want)
	}
}

func benchmarkMemoryW(count int, b *testing.B) {
	mem := NewMemory(1, count)
	mem.MapRAM(0, make([]uint8, count, count))
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i := 0; i < count; i++ {
			mem.Write(i, 0xff)
		}
	}
}

func benchmarkMemoryR(count int, b *testing.B) {
	mem := NewMemory(1, count)
	mem.MapRAM(0, make([]uint8, count, count))
	out := make([]uint8, count, count)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i := 0; i < count; i++ {
			out[i] = mem.Read(i)
		}
	}
}

func BenchmarkMemoryW(b *testing.B)      { benchmarkMemoryW(1, b) }
func BenchmarkMemoryPageW(b *testing.B)  { benchmarkMemoryW(0x100, b) }
func BenchmarkMemorySpaceW(b *testing.B) { benchmarkMemoryW(0x10000, b) }

func BenchmarkMemoryR(b *testing.B)      { benchmarkMemoryR(1, b) }
func BenchmarkMemoryPageR(b *testing.B)  { benchmarkMemoryR(0x100, b) }
func BenchmarkMemorySpaceR(b *testing.B) { benchmarkMemoryR(0x10000, b) }
