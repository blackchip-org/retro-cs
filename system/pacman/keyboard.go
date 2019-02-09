package pacman

import (
	"github.com/veandco/go-sdl2/sdl"
)

type keyboard struct {
	s *system
}

func newKeyboard(s *system) *keyboard {
	return &keyboard{s: s}
}

func (k *keyboard) handle(e *sdl.KeyboardEvent) error {
	s := k.s
	if e.Type == sdl.KEYDOWN {
		switch e.Keysym.Sym {
		case sdl.K_1:
			s.in1 |= 1 << 5
		case sdl.K_2:
			s.in1 |= 1 << 6
		case sdl.K_c:
			s.in0 |= 1 << 5
		case sdl.K_r:
			s.in0 |= 1 << 4
		case sdl.K_UP:
			s.in0 &^= 1 << 0
		case sdl.K_LEFT:
			s.in0 &^= 1 << 1
		case sdl.K_RIGHT:
			s.in0 &^= 1 << 2
		case sdl.K_DOWN:
			s.in0 &^= 1 << 3
		}
	} else if e.Type == sdl.KEYUP {
		switch e.Keysym.Sym {
		case sdl.K_1:
			s.in1 &^= 1 << 5
		case sdl.K_2:
			s.in1 &^= 1 << 6
		case sdl.K_c:
			s.in0 &^= 1 << 5
		case sdl.K_r:
			s.in0 &^= 1 << 4
		case sdl.K_UP:
			s.in0 |= 1 << 0
		case sdl.K_LEFT:
			s.in0 |= 1 << 1
		case sdl.K_RIGHT:
			s.in0 |= 1 << 2
		case sdl.K_DOWN:
			s.in0 |= 1 << 3
		}
	}
	return nil
}
