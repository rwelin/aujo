package dsp

import "math"

func Sinc(n int, a float64, f float64) []float64 {
	w := make([]float64, n)
	for i := 0; i < n; i++ {
		x := f * math.Pi * float64(i+1) / float64(n+1)
		w[i] = a * math.Sin(x) / x
	}
	return w
}
