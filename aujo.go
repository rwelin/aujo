package aujo

import (
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
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
	mutex sync.Mutex

	index int64 // index is the current time

	rem  []byte      // rem are the bytes that are ready to be read
	bufC chan []byte // bufC is used to send the bytes to be read

	seq     *Sequence // seq is the currently playing sequence
	event   int       // event is the index of the next event
	nextSeq *Sequence // nextSeq is played after the current sequence has finished

	SamplingInterval float64 // sampling interval in radians
	Level            float64 // master audio level
	Instruments      []Instrument
	Voices           []Voice
}

func ReadMixConfig(filename string) *Mix {
	m := &Mix{
		bufC: make(chan []byte),
		nextSeq: &Sequence{
			Events: []Event{
				{
					Time: 10000000,
				},
			},
		},
	}

	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	if err := json.NewDecoder(f).Decode(m); err != nil {
		panic(err)
	}

	return m
}

func (m *Mix) Lock() {
	m.mutex.Lock()
}

func (m *Mix) Unlock() {
	m.mutex.Unlock()
}

func (m *Mix) Read(buf []byte) (int, error) {
	if len(m.rem) == 0 {
		m.rem = <-m.bufC
	}
	n := copy(buf, m.rem)
	m.rem = m.rem[n:]
	return n, nil
}

func pitchToFreq(pitch float64) float64 {
	return 440.0 * math.Exp2((pitch-69)/12)
}

func (m *Mix) fill(buf []byte) int {
	m.Lock()
	defer m.Unlock()

	if m.index == 0 {
		m.seq = m.nextSeq
		m.event = 0
	}

	i := 0
	for ; i < len(buf); i += 2 {
		for {
			if len(m.seq.Events) == 0 {
				break
			}
			e := m.seq.Events[m.event]
			if e.Time > m.index {
				break
			}
			if e.Func != nil {
				e.Func(m)
			}
			m.event++
			if m.event >= len(m.seq.Events) {
				m.event = 0
				m.index = 0
			}
		}

		s := float64(m.index) * m.SamplingInterval
		var sum float64
		for _, v := range m.Voices {
			f := pitchToFreq(v.Pitch)
			sum += v.Level * m.Instruments[v.Instrument].Mix(s*f)
		}
		m.index++

		t := uint16(m.Level * sum)
		binary.LittleEndian.PutUint16(buf[i:i+2], t)
	}
	return i
}

func (m *Mix) Mix() {
	for {
		buf := make([]byte, 128)
		n := m.fill(buf)
		m.bufC <- buf[:n]
	}
}

func (m *Mix) SetNextSequence(s *Sequence) {
	m.Lock()
	defer m.Unlock()
	m.nextSeq = s
}

type Event struct {
	Time int64
	Func func(*Mix)
}

type Sequence struct {
	Events []Event
}
