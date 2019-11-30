package aujo

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"math"
	"os"
	"sync"
)

const SamplingInterval = 2.0 * math.Pi / 44100.0

type Envelope struct {
	Value float64
	Time  int64
}

type Instrument struct {
	Harmonics []float64
	Attack    Envelope
	Decay     Envelope
	Sustain   Envelope
	Release   Envelope
}

func interpolate(index int64, e0, e1 Envelope) float64 {
	return (e1.Value-e0.Value)/float64(e1.Time-e0.Time)*float64(index-e0.Time) + e0.Value
}

func (inst *Instrument) Mix(index int64, step float64) float64 {
	if index < 0 {
		return 0
	}

	var lev float64
	if index < inst.Attack.Time {
		lev = interpolate(index, Envelope{}, inst.Attack)
	} else if index < inst.Decay.Time {
		lev = interpolate(index, inst.Attack, inst.Decay)
	} else if index < inst.Sustain.Time {
		lev = interpolate(index, inst.Decay, inst.Sustain)
	} else if index < inst.Release.Time {
		lev = interpolate(index, inst.Sustain, inst.Release)
	}

	var sum float64
	for i, v := range inst.Harmonics {
		sum += v * math.Sin(step*float64(i+1))
	}

	return lev * sum
}

type Voice struct {
	Level      float64
	Pitch      float64
	PitchTime  int64
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

	Level       float64 // master audio level
	Instruments []Instrument
	Voices      []Voice
}

var wavHeader = []byte{
	'R', 'I', 'F', 'F', // ChunkID
	0, 0, 0, 0, // ChunkSize
	'W', 'A', 'V', 'E', // Format
	'f', 'm', 't', ' ', // Subchunk1ID
	0x10, 0x0, 0x0, 0x0, // Subchunk1Size PCM
	0x1, 0x0, // AudioFormat PCM
	0x1, 0x0, // NumChannels Mono
	0x44, 0xAC, 0x0, 0x0, // SampleRate 44100Hz
	0x88, 0x58, 0x1, 0x0, // ByteRate 44100 * 1 * 16/8
	0x2, 0x0, // BlockAlign
	0x10, 0x0, // BitsPerSample
	'd', 'a', 't', 'a', // Subchunk2ID
	0xFF, 0xFF, 0xFF, 0xFF, // Subchunk2Size
}

func NewMix() *Mix {
	return &Mix{
		bufC: make(chan []byte),
		rem:  wavHeader,
	}
}

func ReadMixConfig(filename string) *Mix {
	m := NewMix()

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

func (m *Mix) Play(out io.Writer) {
	go m.Mix()

	for {
		io.CopyN(out, m, 1024)
	}
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

		s := float64(m.index) * SamplingInterval
		var sum float64
		for _, v := range m.Voices {
			f := pitchToFreq(v.Pitch)
			pitchTime := m.index - v.PitchTime
			sum += v.Level * m.Instruments[v.Instrument].Mix(pitchTime, s*f)
		}
		m.index++

		t := uint16(m.Level * sum)
		binary.LittleEndian.PutUint16(buf[i:i+2], t)
	}
	return i
}

func (m *Mix) Mix() {
	for {
		buf := make([]byte, 1024)
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
