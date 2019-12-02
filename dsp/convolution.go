package dsp

func Convolve(y, f, g []float64) {
	if len(y) != len(f) || len(y) != len(g) {
		panic("invalid input")
	}

	n := len(y)

	fi := make([]float64, n)
	gi := make([]float64, n)
	yi := make([]float64, n)

	FFT(f, fi)
	FFT(g, gi)

	ComplexMult(y, yi, f, fi, g, gi)

	IFFT(y, yi)
	IFFT(f, fi)
	IFFT(g, gi)
}
