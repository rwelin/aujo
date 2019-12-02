package dsp

import "math"

func fft(forward bool, x, y []float64) {
	n := len(x)
	m := 0
	for i := n; i > 1; i >>= 1 {
		m++
	}

	i2 := n >> 1
	j := 0
	for i := 0; i < n-1; i++ {
		if i < j {
			x[i], x[j] = x[j], x[i]
			y[i], y[j] = y[j], y[i]
		}
		k := i2
		for k <= j {
			j -= k
			k >>= 1
		}
		j += k
	}

	c1 := float64(-1.0)
	c2 := 0.0
	l2 := 1
	for l := 0; l < m; l++ {
		l1 := l2
		l2 <<= 1
		u1 := 1.0
		u2 := 0.0
		for j := 0; j < l1; j++ {
			for i := j; i < n; i += l2 {
				i1 := i + l1
				t1 := u1*x[i1] - u2*y[i1]
				t2 := u1*y[i1] + u2*x[i1]
				x[i1] = x[i] - t1
				y[i1] = y[i] - t2
				x[i] += t1
				y[i] += t2
			}
			u1, u2 = u1*c1-u2*c2, u1*c2+u2*c1
		}
		c2 = math.Sqrt((1 - c1) / 2)
		if forward {
			c2 = -c2
		}
		c1 = math.Sqrt((1 + c1) / 2)
	}

	if forward {
		for i := 0; i < n; i++ {
			x[i] /= float64(n)
			y[i] /= float64(n)
		}
	}
}

func FFT(x, y []float64) {
	fft(true, x, y)
}

func IFFT(x, y []float64) {
	fft(false, x, y)
}
