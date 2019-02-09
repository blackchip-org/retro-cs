package rcs

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func TestFill(t *testing.T) {
	v := NewVoice(8)
	v.Freq = 2
	v.Vol = 1.0
	v.Waveform = []float64{-1, 0, 1, 0}

	have := make([]float64, 8, 8)
	want := []float64{-1, 0, 1, 0, -1, 0, 1, 0}
	v.Fill(have, len(have))

	if !reflect.DeepEqual(have, want) {
		t.Errorf("\n have: %v \n want: %v", have, want)
	}
}

func TestFillHalfVol(t *testing.T) {
	v := NewVoice(8)
	v.Freq = 2
	v.Vol = 0.5
	v.Waveform = []float64{-1, 0, 1, 0}

	have := make([]float64, 8, 8)
	want := []float64{-0.5, 0, 0.5, 0, -0.5, 0, 0.5, 0}
	v.Fill(have, len(have))

	if !reflect.DeepEqual(have, want) {
		t.Errorf("\n have: %v \n want: %v", have, want)
	}
}

func TestFillStretch(t *testing.T) {
	v := NewVoice(8)
	v.Freq = 1
	v.Vol = 1.0
	v.Waveform = []float64{-1, 0, 1, 0}

	have := make([]float64, 8, 8)
	want := []float64{-1, -1, 0, 0, 1, 1, 0, 0}
	v.Fill(have, len(have))

	if !reflect.DeepEqual(have, want) {
		t.Errorf("\n have: %v \n want: %v", have, want)
	}
}

func TestConvertSample(t *testing.T) {
	tests := []struct {
		from float64
		to   int16
	}{
		{-1, -math.MaxInt16},
		{0, 0},
		{1, math.MaxInt16},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%v to %v", test.from, test.to), func(t *testing.T) {
			have := convert(test.from)
			want := test.to
			if have != want {
				t.Errorf("\n have: %v \n want: %v", have, want)
			}
		})
	}
}
