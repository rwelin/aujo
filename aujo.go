package aujo

import (
	"encoding/binary"
	"math"
	"sync"
)

type Instrument struct {
	Harmonics []float64
}

func (inst *Instrument) Mix(step float64) float64 {
	var sum float64
	for i, v := range inst.Harmonics {
		sum += v * math.Sin(step*float64(i+1))
	}
	return sum
}

type Voice struct {
	Level      float64
	Pitch      float64
	Instrument int
}

type Mix struct {
	mutex            sync.Mutex
	SamplingInterval float64
	Level            float64
	Instruments      []Instrument
	Voices           []Voice
}

func (m *Mix) Lock() {
	m.mutex.Lock()
}

func (m *Mix) Unlock() {
	m.mutex.Unlock()
}

func (m *Mix) Mix(buf []byte, step int) int {
	m.Lock()
	defer m.Unlock()

	i := 0
	for ; i < len(buf)/2; i++ {
		s := float64(step) * m.SamplingInterval
		var sum float64
		for _, v := range m.Voices {
			f := 440.0 * math.Exp2((v.Pitch-69)/12)
			sum += v.Level * m.Instruments[v.Instrument].Mix(f*s)
		}
		step++

		t := uint16(m.Level * sum)
		binary.LittleEndian.PutUint16(buf[2*i:2*(i+1)], t)
	}

	return i
}
