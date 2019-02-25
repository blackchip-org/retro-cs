package rcs

import (
	"bytes"
	"fmt"
	"io/ioutil"
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
		mem.MapLoad(i+10, func() uint8 { reads++; return 40 + uint8(j) })
		mem.MapStore(i+10, func(uint8) { writes++ })
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

func TestMemoryMap(t *testing.T) {
	main := NewMemory(1, 15)
	mem := NewMemory(1, 5)
	mem.MapRAM(0, make([]uint8, 5, 5))
	main.Map(0, mem)
	main.Map(5, mem)

	main.Write(1, 22)
	have := main.Read(6)
	want := uint8(22)
	if have != want {
		t.Errorf("\n have: %04x \n want: %04x", have, want)
	}
}

func TestMemoryUnmap(t *testing.T) {
	var buf bytes.Buffer
	log.SetFlags(0)
	log.SetOutput(&buf)
	defer func() {
		log.SetFlags(log.LstdFlags)
		log.SetOutput(os.Stderr)
	}()

	mem := NewMemory(1, 10)
	mem.MapRAM(0, make([]uint8, 10, 10))
	mem.Write(7, 44)
	mem.Unmap(7)
	mem.Read(7)

	msg := []string{
		"unmapped memory read, bank 0, addr 0x7",
		"",
	}
	have := buf.String()
	want := strings.Join(msg, "\n")
	if have != want {
		t.Errorf("\n have: \n%v \n want: \n%v", have, want)
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

func TestWriteN(t *testing.T) {
	mem := NewMemory(1, 4)
	ram := make([]uint8, 4, 4)
	mem.MapRAM(0, ram)

	mem.WriteN(1, 10, 11, 12)
	want := []uint8{0, 10, 11, 12}
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

func TestMemoryMirror(t *testing.T) {
	mem := NewMemory(1, 20)
	ram := make([]uint8, 10, 10)
	mem.MapRAM(0, ram)
	mem.MapRAM(10, ram)
	mem.Write(4, 99)
	have := mem.Read(14)
	want := uint8(99)
	if have != want {
		t.Errorf("\n have: %v \n want: %v", have, want)
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

func TestPointerFetch(t *testing.T) {
	mem := NewMemory(1, 10)
	mem.MapRAM(0, make([]uint8, 10, 10))

	mem.Write(4, 44)
	p := NewPointer(mem)
	p.SetAddr(4)

	have := p.Fetch()
	want := uint8(44)
	if have != want {
		fmt.Printf("\n have: %v \n want: %v", have, want)
	}
}

func TestPointerFetch2(t *testing.T) {
	mem := NewMemory(1, 10)
	mem.MapRAM(0, make([]uint8, 10, 10))

	mem.Write(4, 44)
	mem.Write(5, 55)
	p := NewPointer(mem)
	p.SetAddr(4)

	p.Fetch()
	have := p.Fetch()
	want := uint8(55)
	if have != want {
		fmt.Printf("\n have: %v \n want: %v", have, want)
	}
}

func TestPeek(t *testing.T) {
	mem := NewMemory(1, 10)
	mem.MapRAM(0, make([]uint8, 10, 10))

	mem.Write(4, 44)
	p := NewPointer(mem)
	p.SetAddr(4)

	p.Peek()
	have := p.Peek()
	want := uint8(44)
	if have != want {
		fmt.Printf("\n have: %v \n want: %v", have, want)
	}
}

func TestFetchLE(t *testing.T) {
	mem := NewMemory(1, 10)
	mem.MapRAM(0, make([]uint8, 10, 10))

	mem.Write(4, 0x44)
	mem.Write(5, 0x55)
	p := NewPointer(mem)
	p.SetAddr(4)

	have := p.FetchLE()
	want := 0x5544
	if have != want {
		fmt.Printf("\n have: %04x \n want: %04x", have, want)
	}
}

func TestPutN(t *testing.T) {
	mem := NewMemory(1, 5)
	ram := make([]uint8, 5, 5)
	mem.MapRAM(0, ram)

	p := NewPointer(mem)
	p.PutN(1, 2, 3)
	p.PutN(4, 5)

	want := []uint8{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(ram, want) {
		fmt.Printf("\n have: %04x \n want: %04x", ram, want)
	}
}

func TestLoadROMs(t *testing.T) {
	data0 := []byte{1, 2}
	data1 := []byte{3, 4}
	readFile = func(filename string) ([]byte, error) {
		switch filename {
		case "data0":
			return data0, nil
		case "data1":
			return data1, nil
		}
		return nil, fmt.Errorf("invalid file: %v", filename)
	}
	defer func() { readFile = ioutil.ReadFile }()

	rom0 := NewROM(" data0 ", " data0 ", "0ca623e2855f2c75c842ad302fe820e41b4d197d")
	rom1 := NewROM(" data1 ", " data1 ", "c512123626a98914cb55a769db20808db3df3af7")
	chunks, err := LoadROMs("", []ROM{rom0, rom1})
	if err != nil {
		t.Error(err)
	}
	want := map[string][]byte{
		"data0": data0,
		"data1": data1,
	}
	if !reflect.DeepEqual(chunks, want) {
		t.Errorf("\n have: %+v \n want: %+v", chunks, want)
	}
}

func TestLoadROMsCombine(t *testing.T) {
	data0 := []byte{1, 2}
	data1 := []byte{3, 4}
	readFile = func(filename string) ([]byte, error) {
		switch filename {
		case "data0":
			return data0, nil
		case "data1":
			return data1, nil
		}
		return nil, fmt.Errorf("invalid file")
	}
	defer func() { readFile = ioutil.ReadFile }()

	rom0 := NewROM(" data ", " data0 ", "0ca623e2855f2c75c842ad302fe820e41b4d197d")
	rom1 := NewROM(" data ", " data1 ", "c512123626a98914cb55a769db20808db3df3af7")
	chunks, err := LoadROMs("", []ROM{rom0, rom1})
	if err != nil {
		t.Error(err)
	}
	want := map[string][]byte{
		"data": []byte{1, 2, 3, 4},
	}
	if !reflect.DeepEqual(chunks, want) {
		t.Errorf("\n have: %+v \n want: %+v", chunks, want)
	}
}

func TestLoadROMsChecksumError(t *testing.T) {
	data0 := []byte{1, 2}
	readFile = func(filename string) ([]byte, error) {
		switch filename {
		case "/data0":
			return data0, nil
		}
		return nil, fmt.Errorf("invalid file")
	}
	defer func() { readFile = ioutil.ReadFile }()

	rom0 := NewROM("data0", "data0", "xx")
	_, err := LoadROMs("/", []ROM{rom0})
	if err == nil {
		t.Errorf("expected error")
	}
}
