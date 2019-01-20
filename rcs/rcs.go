package rcs

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
