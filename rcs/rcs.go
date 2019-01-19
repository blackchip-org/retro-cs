package rcs

// Get8 is a function which loads an unsigned 8-bit value
type Get8 func() uint8

// Put8 is a function which stores an unsiged 8-bit value
type Put8 func(uint8)

// Get is a function which loads an integer value
type Get func() int

// Put is a function which stores an integer value
type Put func(int)
