package dsp

func ComplexMult(y, yi, f, fi, g, gi []float64) {
	for i := 0; i < len(y); i++ {
		y[i] = f[i]*g[i] - fi[i]*gi[i]
		yi[i] = f[i]*gi[i] + fi[i]*g[i]
	}
}
