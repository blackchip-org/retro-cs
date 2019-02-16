package c64

import (
	"github.com/veandco/go-sdl2/sdl"
)

const kbBufLen = 10 // length of keyboard buffer

type keyboard struct {
	buf   []uint8
	ndx   uint8 // Number of characters in keyboard buffer
	stkey uint8 // Was STOP Key Pressed?
	joy2  uint8 // HACK: joystick 2, move elsewhere
}

func newKeyboard() *keyboard {
	return &keyboard{
		buf:  make([]uint8, kbBufLen, kbBufLen),
		joy2: 0xff,
	}
}

func (k *keyboard) handle(e *sdl.KeyboardEvent) error {
	ch, ok := k.lookup(e)
	if !ok {
		return nil
	}
	if k.ndx >= kbBufLen {
		return nil
	}
	k.buf[k.ndx] = ch
	k.ndx++
	return nil
}

const (
	keyCursorDown  = uint8(0x11)
	keyCursorLeft  = uint8(0x9d)
	keyCursorRight = uint8(0x1d)
	keyCursorUp    = uint8(0x91)
)

// https://wiki.libsdl.org/SDLKeycodeLookup
type keymap map[sdl.Keycode]uint8

var keys = keymap{
	sdl.K_BACKSPACE:    0x14,
	sdl.K_RETURN:       0x0d,
	sdl.K_SPACE:        0x20,
	sdl.K_QUOTE:        0x27,
	sdl.K_PERIOD:       0x2e,
	sdl.K_COMMA:        0x2c,
	sdl.K_SLASH:        0x2f,
	sdl.K_0:            0x30,
	sdl.K_1:            0x31,
	sdl.K_2:            0x32,
	sdl.K_3:            0x33,
	sdl.K_4:            0x34,
	sdl.K_5:            0x35,
	sdl.K_6:            0x36,
	sdl.K_7:            0x37,
	sdl.K_8:            0x38,
	sdl.K_9:            0x39,
	sdl.K_SEMICOLON:    0x3b,
	sdl.K_EQUALS:       0x3d,
	sdl.K_LEFTBRACKET:  0x5b,
	sdl.K_BACKSLASH:    0x5c, // british pound
	sdl.K_RIGHTBRACKET: 0x5d,
	sdl.K_a:            0x41,
	sdl.K_b:            0x42,
	sdl.K_c:            0x43,
	sdl.K_d:            0x44,
	sdl.K_e:            0x45,
	sdl.K_f:            0x46,
	sdl.K_g:            0x47,
	sdl.K_h:            0x48,
	sdl.K_i:            0x49,
	sdl.K_j:            0x4a,
	sdl.K_k:            0x4b,
	sdl.K_l:            0x4c,
	sdl.K_m:            0x4d,
	sdl.K_n:            0x4e,
	sdl.K_o:            0x4f,
	sdl.K_p:            0x50,
	sdl.K_q:            0x51,
	sdl.K_r:            0x52,
	sdl.K_s:            0x53,
	sdl.K_t:            0x54,
	sdl.K_u:            0x55,
	sdl.K_v:            0x56,
	sdl.K_w:            0x57,
	sdl.K_x:            0x58,
	sdl.K_y:            0x59,
	sdl.K_z:            0x5a,
	sdl.K_DOWN:         keyCursorDown,
	sdl.K_LEFT:         keyCursorLeft,
	sdl.K_RIGHT:        keyCursorRight,
	sdl.K_UP:           keyCursorUp,
}

var shifted = keymap{
	sdl.K_QUOTE:     0x22,
	sdl.K_PERIOD:    0x3e,
	sdl.K_COMMA:     0x3c,
	sdl.K_SLASH:     0x3f,
	sdl.K_0:         0x29,
	sdl.K_1:         0x21,
	sdl.K_2:         0x40,
	sdl.K_3:         0x23,
	sdl.K_4:         0x24,
	sdl.K_5:         0x25,
	sdl.K_6:         0x5e,
	sdl.K_7:         0x26,
	sdl.K_8:         0x2a,
	sdl.K_9:         0x28,
	sdl.K_SEMICOLON: 0x3a,
	sdl.K_EQUALS:    0x2b,
	sdl.K_a:         0xc1,
	sdl.K_b:         0xc2,
	sdl.K_c:         0xc3,
	sdl.K_d:         0xc4,
	sdl.K_e:         0xc5,
	sdl.K_f:         0xc6,
	sdl.K_g:         0xc7,
	sdl.K_h:         0xc8,
	sdl.K_i:         0xc9,
	sdl.K_j:         0xca,
	sdl.K_k:         0xcb,
	sdl.K_l:         0xcc,
	sdl.K_m:         0xcd,
	sdl.K_n:         0xce,
	sdl.K_o:         0xcf,
	sdl.K_p:         0xd0,
	sdl.K_q:         0xd1,
	sdl.K_r:         0xd2,
	sdl.K_s:         0xd3,
	sdl.K_t:         0xd4,
	sdl.K_u:         0xd5,
	sdl.K_v:         0xd6,
	sdl.K_w:         0xd7,
	sdl.K_x:         0xd8,
	sdl.K_y:         0xd9,
	sdl.K_z:         0xda,
}

var keymaps = map[sdl.Keymod]keymap{
	sdl.KMOD_NONE:   keys,
	sdl.KMOD_LSHIFT: shifted,
	sdl.KMOD_RSHIFT: shifted,
}

func (k *keyboard) lookup(e *sdl.KeyboardEvent) (uint8, bool) {
	keysym := e.Keysym
	switch {
	/*
		case keysym.Mod&sdl.KMOD_CTRL > 0 && keysym.Sym == sdl.K_ESCAPE:
				if e.Type == sdl.KEYUP {
					k.mach.Reset()
					return 0, false
				}
	*/
	case keysym.Mod&sdl.KMOD_CTRL > 0 && keysym.Sym == sdl.K_c:
		if e.Type == sdl.KEYDOWN {
			k.stkey = 0x7f
		}
	case keysym.Sym == sdl.K_c && e.Type == sdl.KEYUP:
		k.stkey = 0xff

	case keysym.Sym == sdl.K_UP && e.Type == sdl.KEYDOWN:
		k.joy2 &^= (1 << 0)
	case keysym.Sym == sdl.K_DOWN && e.Type == sdl.KEYDOWN:
		k.joy2 &^= (1 << 1)
	case keysym.Sym == sdl.K_LEFT && e.Type == sdl.KEYDOWN:
		k.joy2 &^= (1 << 2)
	case keysym.Sym == sdl.K_RIGHT && e.Type == sdl.KEYDOWN:
		k.joy2 &^= (1 << 3)
	case keysym.Sym == sdl.K_SPACE && e.Type == sdl.KEYDOWN:
		k.joy2 &^= (1 << 4)

	case keysym.Sym == sdl.K_UP && e.Type == sdl.KEYUP:
		k.joy2 |= (1 << 0)
	case keysym.Sym == sdl.K_DOWN && e.Type == sdl.KEYUP:
		k.joy2 |= (1 << 1)
	case keysym.Sym == sdl.K_LEFT && e.Type == sdl.KEYUP:
		k.joy2 |= (1 << 2)
	case keysym.Sym == sdl.K_RIGHT && e.Type == sdl.KEYUP:
		k.joy2 |= (1 << 3)
	case keysym.Sym == sdl.K_SPACE && e.Type == sdl.KEYUP:
		k.joy2 |= (1 << 4)
	}

	if e.Type != sdl.KEYDOWN {
		return 0, false
	}
	keymap0 := keymaps[sdl.KMOD_NONE]
	mod := sdl.Keymod(keysym.Mod & 0x03) // Just take the lower two bits
	keymap, ok := keymaps[mod]
	if !ok {
		keymap = keymap0
	}
	ch, ok := keymap[keysym.Sym]
	if !ok {
		ch, ok = keymap0[keysym.Sym]
	}
	if !ok {
		return 0, false
	}
	return ch, true
}

func (k *keyboard) special(keysym sdl.Keysym) bool {
	return false
}
