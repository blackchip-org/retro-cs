package cbm

import (
	"fmt"
	"log"

	"github.com/blackchip-org/retro-cs/rcs"
	"github.com/veandco/go-sdl2/sdl"
)

type Keyboard struct {
	a      *uint8   // data port A
	b      *uint8   // data port B
	DirA   uint8    // direction of data port A
	DirB   uint8    // direction of data port B
	Matrix [][]bool // keyboard matrix, indexes are col then row, true = pressed
}

type pos struct {
	col int
	row int
}

func NewKeyboard(a *uint8, b *uint8) *Keyboard {
	k := &Keyboard{
		a:      a,
		b:      b,
		Matrix: make([][]bool, 8, 8),
	}
	for i := 0; i < 8; i++ {
		k.Matrix[i] = make([]bool, 8, 8)
	}
	return k
}

func (k *Keyboard) Handle(e *sdl.KeyboardEvent) error {
	val := e.Type == sdl.KEYDOWN
	p, ok := standard[e.Keysym.Sym]
	if !ok {
		return nil
	}
	fmt.Printf("setting: %+v\n", p)
	k.Matrix[p.row][p.col] = val
	return nil
}

func (k *Keyboard) Scan() {
	if k.DirA != 0xff || k.DirB != 0x00 {
		log.Printf("keyboard: unsupported configuration: %02x %02x", k.DirA, k.DirB)
		return
	}
	// bit set determines which column to read
	sel := *k.A
	row := 0
	for i := 0; i < 8; i++ {
		if sel&1 == 0 {
			row = i
			break
		}
		sel >>= 1
	}
	// set each bit in data port B, one if not pressed, zero if pressed
	keys := uint8(0)
	for col := 0; col < 8; col++ {
		keys <<= 1
		if !k.Matrix[row][col] {
			keys |= 1
		}
	}
	*k.B = keys
	fmt.Printf("row: %v, keys: %v, A: %v\n", row, rcs.B8(*k.B), rcs.X8(*k.A))
}

var standard = map[sdl.Keycode]pos{
	sdl.K_ESCAPE:      pos{7, 7},
	sdl.K_q:           pos{7, 6},
	sdl.K_APPLICATION: pos{7, 5}, // Commodore key
	sdl.K_SPACE:       pos{7, 4},
	sdl.K_2:           pos{7, 3},
	sdl.K_LCTRL:       pos{7, 2},
	sdl.K_BACKQUOTE:   pos{7, 1}, // <- key
	sdl.K_1:           pos{7, 0},
	sdl.K_SLASH:       pos{6, 7},
	sdl.K_EQUALS:      pos{6, 5},
	sdl.K_RSHIFT:      pos{6, 4},
	sdl.K_HOME:        pos{6, 3},
	sdl.K_SEMICOLON:   pos{6, 2},
	sdl.K_ASTERISK:    pos{6, 1},
	sdl.K_BACKSLASH:   pos{6, 0}, // British pound
	sdl.K_COMMA:       pos{5, 7},
	sdl.K_AT:          pos{5, 6},
	sdl.K_COLON:       pos{5, 5},
	sdl.K_PERIOD:      pos{5, 4},
	sdl.K_MINUS:       pos{5, 3},
	sdl.K_l:           pos{5, 2},
	sdl.K_p:           pos{5, 1},
	sdl.K_PLUS:        pos{5, 0},
	sdl.K_n:           pos{4, 7},
	sdl.K_o:           pos{4, 6},
	sdl.K_k:           pos{4, 5},
	sdl.K_m:           pos{4, 4},
	sdl.K_0:           pos{4, 3},
	sdl.K_j:           pos{4, 2},
	sdl.K_i:           pos{4, 1},
	sdl.K_9:           pos{4, 0},
	sdl.K_v:           pos{3, 7},
	sdl.K_u:           pos{3, 6},
	sdl.K_h:           pos{3, 5},
	sdl.K_b:           pos{3, 4},
	sdl.K_8:           pos{3, 3},
	sdl.K_g:           pos{3, 2},
	sdl.K_y:           pos{3, 1},
	sdl.K_7:           pos{3, 0},
	sdl.K_x:           pos{2, 7},
	sdl.K_t:           pos{2, 6},
	sdl.K_f:           pos{2, 5},
	sdl.K_c:           pos{2, 4},
	sdl.K_6:           pos{2, 3},
	sdl.K_d:           pos{2, 2},
	sdl.K_r:           pos{2, 1},
	sdl.K_5:           pos{2, 0},
	sdl.K_LSHIFT:      pos{1, 7},
	sdl.K_e:           pos{1, 6},
	sdl.K_s:           pos{1, 5},
	sdl.K_z:           pos{1, 4},
	sdl.K_4:           pos{1, 3},
	sdl.K_a:           pos{1, 2},
	sdl.K_w:           pos{1, 1},
	sdl.K_3:           pos{1, 0},
	sdl.K_DOWN:        pos{0, 7},
	sdl.K_F5:          pos{0, 6},
	sdl.K_F3:          pos{0, 5},
	sdl.K_F1:          pos{0, 4},
	sdl.K_F7:          pos{0, 3},
	sdl.K_RIGHT:       pos{0, 2},
	sdl.K_RETURN:      pos{0, 1},
	sdl.K_DELETE:      pos{0, 0},
}
