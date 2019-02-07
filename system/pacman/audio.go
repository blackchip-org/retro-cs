package pacman

import (
	"github.com/veandco/go-sdl2/sdl"
)

type audioData struct {
	waveforms []uint8
}

type voice struct {
	acc      []uint8
	waveform uint8
	freq     []uint8
	vol      uint8
}

type audio struct {
	voices []voice
}

func newAudio(spec sdl.AudioSpec, data audioData) (*audio, error) {
	a := &audio{
		voices: make([]voice, 3, 3),
	}
	a.voices[0].acc = make([]uint8, 5, 5)
	a.voices[0].freq = make([]uint8, 5, 5)
	a.voices[1].acc = make([]uint8, 4, 4)
	a.voices[1].freq = make([]uint8, 4, 4)
	a.voices[2].acc = make([]uint8, 4, 4)
	a.voices[2].freq = make([]uint8, 4, 4)
	return a, nil
}
