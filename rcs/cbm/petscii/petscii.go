package petscii

const (
	White      = 0x05
	Red        = 0x1c
	Green      = 0x1e
	Blue       = 0x1f
	Orange     = 0x81
	Black      = 0x90
	Brown      = 0x95
	LightRed   = 0x96
	DarkGray   = 0x97
	MediumGray = 0x98
	LightGreen = 0x99
	LightBlue  = 0x9a
	LightGray  = 0x9b
	Purple     = 0x9c
	Yellow     = 0x9e
	Cyan       = 0x9f
)

// Decoder converts byte values to PETSCII equivilents in Unicode.
var Decoder = func(code uint8) (rune, bool) {
	ch, printable := tableUnshifted[code]
	return ch, printable
}

// ShiftedDecoder converts byte values to PETSCII equivilents in Unicode.
var ShiftedDecoder = func(code uint8) (rune, bool) {
	ch, printable := tableShifted[code]
	return ch, printable
}
