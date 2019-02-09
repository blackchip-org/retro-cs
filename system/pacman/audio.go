package pacman

import (
	"github.com/blackchip-org/retro-cs/rcs"
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
	voices    []voice
	waveforms [16][]float64
	synth     *rcs.Synth
}

func newAudio(spec sdl.AudioSpec, data audioData) (*audio, error) {
	synth, err := rcs.NewSynth(spec, 3)
	if err != nil {
		return nil, err
	}
	a := &audio{
		voices: make([]voice, 3, 3),
		synth:  synth,
	}
	a.voices[0].acc = make([]uint8, 5, 5)
	a.voices[0].freq = make([]uint8, 5, 5)
	a.voices[1].acc = make([]uint8, 4, 4)
	a.voices[1].freq = make([]uint8, 4, 4)
	a.voices[2].acc = make([]uint8, 4, 4)
	a.voices[2].freq = make([]uint8, 4, 4)

	for i := 0; i < 16; i++ {
		addr := uint16(i * 32)
		a.waveforms[i] = rescale(data.waveforms, addr)
	}

	return a, nil
}

func (a *audio) queue() error {
	for i := 0; i < 3; i++ {
		v := a.voices[i]
		wf := rcs.SliceBits(v.waveform, 0, 2)

		// Voice 0 has 5 bytes but Voice 1 and 2 only have 4 bytes with
		// the missing lower byte being zero.
		nFreq := 4
		if i == 0 {
			nFreq = 5
		}
		a.synth.V[i].Freq = freq(v.freq, nFreq)
		a.synth.V[i].Vol = float64(v.vol&0xf) / 15
		a.synth.V[i].Waveform = a.waveforms[wf]
	}
	return a.synth.Queue()
}

func rescale(d []uint8, addr uint16) []float64 {
	out := make([]float64, 32, 32)
	for i := uint16(0); i < 32; i++ {
		v := d[addr+i]
		out[i] = (float64(v) - 7.5) / 8
	}
	return out
}

func freq(f []uint8, n int) int {
	val := uint32(0)
	shift := uint(0)
	if n == 4 {
		shift = 4
	}
	for i := 0; i < n; i++ {
		val += uint32(f[i]&0x0f) << shift
		shift += 4
	}
	freq := (375.0 / 4096.0) * float32(val)
	return int(freq)
}
