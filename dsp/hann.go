package dsp

import "math"

func Hann(n int) []float64 {
	w := make([]float64, n)
	n--
	for i := range w {
		v := math.Sin(math.Pi * float64(i) / float64(n))
		w[i] = v * v
	}
	return w
}
