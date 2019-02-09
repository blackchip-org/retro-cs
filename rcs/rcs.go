package rcs

import (
	"encoding/gob"
	"io"
	"math/bits"
	"strconv"

	"github.com/veandco/go-sdl2/sdl"
)

// Load8 is a function which loads an unsigned 8-bit value
type Load8 func() uint8

// Store8 is a function which stores an unsiged 8-bit value
type Store8 func(uint8)

// Load is a function which loads an integer value
type Load func() int

// Store is a function which stores an integer value
type Store func(int)

// FromBCD converts a binary-coded decimal to an integer value.
func FromBCD(v uint8) uint8 {
	low := v & 0x0f
	high := v >> 4
	return high*10 + low
}

// ToBCD converts an integer value to a binary-coded decimal.
func ToBCD(v uint8) uint8 {
	low := v % 10
	high := (v / 10) % 10
	return high<<4 | low
}

// ParseBits parses the base-2 string value s to a uint8. Panics if s is not
// a valid number. Use strconv.ParseUint for input which may be malformed.
func ParseBits(s string) uint8 {
	value, err := strconv.ParseUint(s, 2, 8)
	if err != nil {
		panic(err)
	}
	return uint8(value)
}

// SliceBits extracts a sequence of bits in value from bit lo to bit hi,
// inclusive.
func SliceBits(value uint8, lo int, hi int) uint8 {
	value = value >> uint(lo)
	bits := uint(hi - lo + 1)
	mask := uint8(1)<<bits - 1
	return value & mask
}

// Add performs addition on in0 and in1 with a carry and returns the result
// along with the new values for the carry, half-carry, and overflow
// flags.
func Add(in0, in1 uint8, carry bool) (out uint8, c, h, v bool) {
	// https://stackoverflow.com/questions/8034566/overflow-and-carry-flags-on-z80/8037485#8037485
	var carryOut uint8

	if carry {
		if in0 >= 0xff-in1 {
			carryOut = 1
		}
		out = in0 + in1 + 1
	} else {
		if in0 > 0xff-in1 {
			carryOut = 1
		}
		out = in0 + in1
	}
	carryIns := out ^ in0 ^ in1

	c = carryOut != 0
	h = carryIns&(1<<4) != 0
	v = (carryIns>>7)^carryOut != 0
	return
}

// Sub performs subtraction of in1 from in0 with a borrow and returns the result
// along with the new values for the borrow, half-borrow, and overflow
// flags.
func Sub(in0, in1 uint8, borrow bool) (out uint8, fc, fh, fv bool) {
	fc = !borrow
	out, fc, fh, fv = Add(in0, ^in1, fc)
	fc = !fc
	fh = !fh
	return
}

// Parity returns true if there are an even number of bits set in the
// given value.
func Parity(v uint8) bool {
	p := bits.OnesCount8(v)
	return p == 0 || p == 2 || p == 4 || p == 6 || p == 8
}

// BitPlane4 returns the nth 2-bit value stored in 4-bit planes found in v.
// If n is 0, returns bits 0 and 4. If n is 1, returns bits 1 and 5, etc.
func BitPlane4(v uint8, n int) uint8 {
	result := 0
	for i, start := range []int{0, 4} {
		checkBit := uint8(1) << uint(start+n)
		if v&checkBit != 0 {
			result += 1 << uint(i)
		}
	}
	return uint8(result)
}

// CharDecoder converts a byte value to a unicode character and indicates
// if this character is considered to be printable.
type CharDecoder func(uint8) (ch rune, printable bool)

// ASCIIDecoder is a pass through of byte values to unicode characters.
// Values 32 to 128 are considered printable.
var ASCIIDecoder = func(code uint8) (rune, bool) {
	printable := code >= 32 && code < 128
	return rune(code), printable
}

// Value is a function wrapper around an arbitrary value for changing
// its contents.
type Value struct {
	Get interface{}
	Put interface{}
}

// SDLContext contains the window for rendering and the audio specs
// available for use.
type SDLContext struct {
	Window    *sdl.Window
	Renderer  *sdl.Renderer
	AudioSpec sdl.AudioSpec
}

type Encoder struct {
	Err error
	enc *gob.Encoder
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		enc: gob.NewEncoder(w),
	}
}

func (e *Encoder) Encode(v interface{}) {
	if e.Err != nil {
		return
	}
	e.Err = e.enc.Encode(v)
}

type Decoder struct {
	Err error
	dec *gob.Decoder
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		dec: gob.NewDecoder(r),
	}
}

func (d *Decoder) Decode(v interface{}) {
	if d.Err != nil {
		return
	}
	d.Err = d.dec.Decode(v)
}

type Saver interface {
	Save(*Encoder)
}

type Loader interface {
	Load(*Decoder)
}
